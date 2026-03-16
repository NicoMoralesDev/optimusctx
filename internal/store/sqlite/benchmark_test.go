package sqlite

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
)

func TestApplyMigrationsCreatesBenchmarkTables(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	assertTablesExist(t, store.DB(), "benchmark_runs", "benchmark_lane_samples", "benchmark_lane_metrics")
	assertIndexColumns(t, store.DB(), "benchmark_runs", []string{"repository_id", "suite_id", "started_at"})
	assertIndexColumns(t, store.DB(), "benchmark_runs", []string{"repository_id", "arm_kind", "started_at"})
	assertIndexColumns(t, store.DB(), "benchmark_lane_samples", []string{"benchmark_run_id", "lane"})
	assertIndexColumns(t, store.DB(), "benchmark_lane_metrics", []string{"benchmark_lane_sample_id", "metric_name"})

	if _, _, err := store.SaveBenchmarkRun(ctx, BenchmarkRunRecord{
		RepositoryID:  repoID,
		SuiteID:       "go-benchmark-discovery-v1",
		SuiteVersion:  "v1",
		FixtureID:     "go-benchmark",
		FixturePath:   "go-benchmark/v1/repository",
		ArmKind:       repository.BenchmarkArmKindBaseline,
		ArmName:       "Baseline",
		Attempt:       1,
		WorkspacePath: layout.RepoRoot,
		StartedAt:     time.Date(2026, 3, 16, 15, 0, 0, 0, time.UTC),
		CompletedAt:   time.Date(2026, 3, 16, 15, 0, 1, 0, time.UTC),
	}, nil); err != nil {
		t.Fatalf("SaveBenchmarkRun() error = %v", err)
	}
}

func TestBenchmarkLaneMetricsPersist(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	runStarted := time.Date(2026, 3, 16, 16, 0, 0, 0, time.UTC)
	runCompleted := runStarted.Add(2 * time.Second)
	run, samples, err := store.SaveBenchmarkRun(ctx, BenchmarkRunRecord{
		RepositoryID:  repoID,
		SuiteID:       "go-benchmark-discovery-v1",
		SuiteVersion:  "v1",
		FixtureID:     "go-benchmark",
		FixturePath:   "go-benchmark/v1/repository",
		ArmKind:       repository.BenchmarkArmKindOptimusCtx,
		ArmName:       "OptimusCtx MCP workflow",
		Attempt:       2,
		WorkspacePath: layout.RepoRoot,
		StartedAt:     runStarted,
		CompletedAt:   runCompleted,
		MetadataJSON:  `{"source":"runner"}`,
	}, []BenchmarkLaneSampleBundle{
		{
			Sample: BenchmarkLaneSampleRecord{
				Lane:          repository.BenchmarkLaneDiscovery,
				StartMarker:   "discovery_started",
				SuccessMarker: "target_identified",
				StopMarker:    "target_identified",
				StartedAt:     runStarted,
				FinishedAt:    runStarted.Add(500 * time.Millisecond),
				ElapsedMS:     500,
				Success:       true,
			},
			Metrics: []BenchmarkLaneMetricRecord{
				{MetricName: benchmarkMetricActionCount, ValueInt: 2},
				{MetricName: string(repository.BenchmarkMetricBroadSearchActions), ValueInt: 1},
				{MetricName: string(repository.BenchmarkMetricTargetedLookupActions), ValueInt: 1},
				{MetricName: string(repository.BenchmarkMetricConsultedArtifacts), Ordinal: 0, ValueInt: 1, ValueText: "internal/http/handler/rollout.go"},
			},
		},
		{
			Sample: BenchmarkLaneSampleRecord{
				Lane:          repository.BenchmarkLaneContextAssembly,
				StartMarker:   "context_started",
				SuccessMarker: "context_ready",
				StopMarker:    "context_ready",
				StartedAt:     runStarted.Add(750 * time.Millisecond),
				FinishedAt:    runStarted.Add(1400 * time.Millisecond),
				ElapsedMS:     650,
				Success:       true,
			},
			Metrics: []BenchmarkLaneMetricRecord{
				{MetricName: benchmarkMetricActionCount, ValueInt: 1},
				{MetricName: string(repository.BenchmarkMetricFileReadActions), ValueInt: 1},
				{MetricName: string(repository.BenchmarkMetricBytesRead), ValueInt: 144},
				{MetricName: string(repository.BenchmarkMetricConsultedArtifacts), Ordinal: 0, ValueInt: 1, ValueText: "internal/config/loader.go"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveBenchmarkRun() error = %v", err)
	}

	gotRun, gotSamples, err := store.LoadBenchmarkRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("LoadBenchmarkRun() error = %v", err)
	}

	run.CreatedAt = gotRun.CreatedAt
	run.UpdatedAt = gotRun.UpdatedAt
	if !reflect.DeepEqual(gotRun, run) {
		t.Fatalf("run mismatch\n got=%+v\nwant=%+v", gotRun, run)
	}
	if len(gotSamples) != 2 {
		t.Fatalf("len(gotSamples) = %d, want 2", len(gotSamples))
	}

	for idx := range samples {
		samples[idx].Sample.ID = gotSamples[idx].Sample.ID
		samples[idx].Sample.BenchmarkRunID = run.ID
		samples[idx].Sample.StartedAt = gotSamples[idx].Sample.StartedAt
		samples[idx].Sample.FinishedAt = gotSamples[idx].Sample.FinishedAt
		samples[idx].Metrics = reorderBenchmarkMetricsForLoad(samples[idx].Metrics)
		for metricIdx := range samples[idx].Metrics {
			samples[idx].Metrics[metricIdx].ID = gotSamples[idx].Metrics[metricIdx].ID
			samples[idx].Metrics[metricIdx].BenchmarkLaneSampleID = gotSamples[idx].Sample.ID
		}
	}
	if !reflect.DeepEqual(gotSamples, samples) {
		t.Fatalf("samples mismatch\n got=%+v\nwant=%+v", gotSamples, samples)
	}
}

func reorderBenchmarkMetricsForLoad(metrics []BenchmarkLaneMetricRecord) []BenchmarkLaneMetricRecord {
	ordered := append([]BenchmarkLaneMetricRecord(nil), metrics...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].MetricName == ordered[j].MetricName {
			if ordered[i].Ordinal == ordered[j].Ordinal {
				return ordered[i].ID < ordered[j].ID
			}
			return ordered[i].Ordinal < ordered[j].Ordinal
		}
		return ordered[i].MetricName < ordered[j].MetricName
	})
	return ordered
}

func TestBenchmarkMutationLanesPersistEvidence(t *testing.T) {
	t.Parallel()

	result := repository.BenchmarkRunResult{
		SchemaVersion: repository.BenchmarkSuiteSchemaV1,
		SuiteID:       "go-benchmark-refresh-v1",
		SuiteVersion:  "v1",
		FixtureID:     "go-worktree",
		FixturePath:   "go-worktree/v1/repository",
		WorkspacePath: "/tmp/benchmark",
		Arms: []repository.BenchmarkArmRunResult{
			{
				Kind:       repository.BenchmarkArmKindOptimusCtx,
				Name:       "OptimusCtx CLI and MCP workflow",
				Workspace:  "/tmp/benchmark/optimusctx",
				StartedAt:  time.Date(2026, 3, 16, 17, 0, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, 3, 16, 17, 0, 4, 0, time.UTC),
				LaneResults: []repository.BenchmarkLaneRunResult{
					{
						Lane:           repository.BenchmarkLaneRefreshReady,
						StartMarker:    "refresh_after_change_started",
						SuccessMarker:  "refresh_ready",
						StopMarker:     "refresh_ready",
						SetupAppliedAt: time.Date(2026, 3, 16, 17, 0, 0, 0, time.UTC),
						StartedAt:      time.Date(2026, 3, 16, 17, 0, 1, 0, time.UTC),
						FinishedAt:     time.Date(2026, 3, 16, 17, 0, 2, 0, time.UTC),
						Elapsed:        time.Second,
						Success:        true,
						Setup: []repository.EvalSetupAction{{
							Kind:    repository.EvalSetupActionOverwriteFile,
							Path:    "docs/notes.txt",
							Content: "mutated benchmark note\n",
						}},
						Assertions: []repository.BenchmarkAssertion{{
							File:     "docs/notes.txt",
							Kind:     repository.EvalAssertionKindContains,
							Contains: "mutated benchmark note",
						}},
						EvidencePaths: []string{"docs/notes.txt", ".optimusctx/state.json"},
						Effort: repository.BenchmarkLaneEffort{
							ActionCount:           2,
							TargetedLookupActions: 2,
							ConsultedArtifacts:    []string{"docs/notes.txt", ".optimusctx/state.json"},
						},
					},
					{
						Lane:          repository.BenchmarkLaneTaskCompletion,
						StartMarker:   "task_completion_started",
						SuccessMarker: "task_complete",
						StopMarker:    "task_complete",
						StartedAt:     time.Date(2026, 3, 16, 17, 0, 2, 0, time.UTC),
						FinishedAt:    time.Date(2026, 3, 16, 17, 0, 4, 0, time.UTC),
						Elapsed:       2 * time.Second,
						Success:       true,
						Assertions: []repository.BenchmarkAssertion{{
							File:     "docs/notes.txt",
							Kind:     repository.EvalAssertionKindContains,
							Contains: "mutated benchmark note",
						}},
						EvidencePaths: []string{"docs/notes.txt", "artifacts/pack.json"},
						Effort: repository.BenchmarkLaneEffort{
							ActionCount:        2,
							FileReadActions:    1,
							BytesRead:          128,
							ConsultedArtifacts: []string{"docs/notes.txt", "artifacts/pack.json"},
						},
					},
				},
			},
		},
	}

	persisted := BenchmarkPersistedArmsFromResult(42, 3, result)
	if got, want := len(persisted), 1; got != want {
		t.Fatalf("len(persisted) = %d, want %d", got, want)
	}

	var runMetadata map[string]any
	if err := json.Unmarshal([]byte(persisted[0].Run.MetadataJSON), &runMetadata); err != nil {
		t.Fatalf("run metadata json: %v", err)
	}
	if got, want := runMetadata["workspacePath"], "/tmp/benchmark/optimusctx"; got != want {
		t.Fatalf("run metadata workspacePath = %#v, want %#v", got, want)
	}

	var laneMetadata map[string]any
	if err := json.Unmarshal([]byte(persisted[0].Samples[0].Sample.MetadataJSON), &laneMetadata); err != nil {
		t.Fatalf("lane metadata json: %v", err)
	}
	if got := laneMetadata["setupAppliedAt"]; got == "" {
		t.Fatalf("setupAppliedAt missing from lane metadata: %+v", laneMetadata)
	}
	evidence, ok := laneMetadata["evidencePaths"].([]any)
	if !ok || len(evidence) != 2 {
		t.Fatalf("lane metadata evidencePaths = %#v", laneMetadata["evidencePaths"])
	}
	assertions, ok := laneMetadata["assertions"].([]any)
	if !ok || len(assertions) != 1 {
		t.Fatalf("lane metadata assertions = %#v", laneMetadata["assertions"])
	}
}

func TestBenchmarkRecomputedAttributionMatchesExportUsesLatestAttempts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	for attempt := 1; attempt <= 4; attempt++ {
		result := benchmarkSQLiteEvidenceRunResult(layout.RepoRoot, attempt)
		for _, persisted := range BenchmarkPersistedArmsFromResult(repoID, attempt, result) {
			if _, _, err := store.SaveBenchmarkRun(ctx, persisted.Run, persisted.Samples); err != nil {
				t.Fatalf("SaveBenchmarkRun(attempt=%d) error = %v", attempt, err)
			}
		}
	}

	latest, err := store.ListLatestBenchmarkRuns(ctx, repoID, "go-benchmark-refresh-v1", "v1", 4)
	if err != nil {
		t.Fatalf("ListLatestBenchmarkRuns() error = %v", err)
	}
	if got, want := len(latest), 4; got != want {
		t.Fatalf("len(latest) = %d, want %d", got, want)
	}
	attempts := []int{latest[0].Run.Attempt, latest[1].Run.Attempt, latest[2].Run.Attempt, latest[3].Run.Attempt}
	if !reflect.DeepEqual(attempts, []int{3, 3, 4, 4}) {
		t.Fatalf("latest attempts = %+v", attempts)
	}
	for _, run := range latest {
		for _, sample := range run.Samples {
			var metadata map[string]any
			if err := json.Unmarshal([]byte(sample.Sample.MetadataJSON), &metadata); err != nil {
				t.Fatalf("json.Unmarshal(metadata) error = %v", err)
			}
			if _, ok := metadata["attribution"]; !ok {
				t.Fatalf("metadata missing attribution: %+v", metadata)
			}
		}
	}
}

func benchmarkSQLiteEvidenceRunResult(repositoryRoot string, attempt int) repository.BenchmarkRunResult {
	return repository.BenchmarkRunResult{
		SchemaVersion: repository.BenchmarkSuiteSchemaV1,
		SuiteID:       "go-benchmark-refresh-v1",
		SuiteVersion:  "v1",
		FixtureID:     "go-worktree",
		FixturePath:   "go-worktree/v1/repository",
		WorkspacePath: repositoryRoot,
		Arms: []repository.BenchmarkArmRunResult{
			{
				Kind:       repository.BenchmarkArmKindBaseline,
				Name:       "Baseline",
				Workspace:  repositoryRoot,
				StartedAt:  time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, 3, 16, 20, 0, 2, 0, time.UTC),
				LaneResults: []repository.BenchmarkLaneRunResult{{
					Lane:          repository.BenchmarkLaneRefreshReady,
					StartMarker:   "refresh_after_change_started",
					SuccessMarker: "refresh_ready",
					StopMarker:    "refresh_ready",
					StartedAt:     time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC),
					FinishedAt:    time.Date(2026, 3, 16, 20, 0, 1, 0, time.UTC),
					Elapsed:       time.Second,
					Success:       true,
					EvidencePaths: []string{"docs/notes.txt"},
					Effort: repository.BenchmarkLaneEffort{
						ActionCount:     1,
						FileReadActions: 1,
						BytesRead:       256,
					},
				}},
			},
			{
				Kind:       repository.BenchmarkArmKindOptimusCtx,
				Name:       "OptimusCtx CLI and MCP workflow",
				Workspace:  repositoryRoot,
				StartedAt:  time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, 3, 16, 20, 0, 2, 0, time.UTC),
				LaneResults: []repository.BenchmarkLaneRunResult{{
					Lane:          repository.BenchmarkLaneRefreshReady,
					StartMarker:   "refresh_after_change_started",
					SuccessMarker: "refresh_ready",
					StopMarker:    "refresh_ready",
					StartedAt:     time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC),
					FinishedAt:    time.Date(2026, 3, 16, 20, 0, 1, 0, time.UTC),
					Elapsed:       time.Second,
					Success:       true,
					EvidencePaths: []string{"artifacts/pack.json"},
					Effort: repository.BenchmarkLaneEffort{
						ActionCount:        2,
						FileReadActions:    1,
						BytesRead:          128,
						ConsultedArtifacts: []string{"artifacts/pack.json"},
					},
					Attribution: []repository.BenchmarkArtifactConsumption{{
						StepID:          "pack-export",
						Lane:            repository.BenchmarkLaneRefreshReady,
						ArtifactType:    repository.BenchmarkArtifactTypePackExport,
						ReportLabel:     repository.BenchmarkReportArtifactLabelPackExport,
						SourceKind:      repository.BenchmarkTokenEstimateSourcePackExportSection,
						ArtifactPath:    "artifacts/pack.json",
						EstimatedBytes:  128,
						EstimatedTokens: 32 + int64(attempt),
					}},
				}},
			},
		},
	}
}

func TestAttributionPersistenceInputs(t *testing.T) {
	t.Parallel()

	result := repository.BenchmarkRunResult{
		SchemaVersion: repository.BenchmarkSuiteSchemaV1,
		SuiteID:       "go-benchmark-discovery-v1",
		SuiteVersion:  "v1",
		FixtureID:     "go-benchmark",
		FixturePath:   "go-benchmark/v1/repository",
		WorkspacePath: "/tmp/benchmark",
		Arms: []repository.BenchmarkArmRunResult{
			{
				Kind:       repository.BenchmarkArmKindOptimusCtx,
				Name:       "OptimusCtx",
				Workspace:  "/tmp/benchmark/optimusctx",
				StartedAt:  time.Date(2026, 3, 16, 17, 30, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, 3, 16, 17, 30, 2, 0, time.UTC),
				LaneResults: []repository.BenchmarkLaneRunResult{{
					Lane:          repository.BenchmarkLaneContextAssembly,
					StartMarker:   "context_started",
					SuccessMarker: "context_ready",
					StopMarker:    "context_ready",
					StartedAt:     time.Date(2026, 3, 16, 17, 30, 0, 0, time.UTC),
					FinishedAt:    time.Date(2026, 3, 16, 17, 30, 2, 0, time.UTC),
					Elapsed:       2 * time.Second,
					Success:       true,
					EvidencePaths: []string{"artifacts/pack.json"},
					Attribution: []repository.BenchmarkArtifactConsumption{
						{
							StepID:          "pack",
							StepName:        "Export pack",
							Lane:            repository.BenchmarkLaneContextAssembly,
							Surface:         repository.BenchmarkTreatmentSurfaceCLI,
							Command:         repository.EvalCommandPackExport,
							ArtifactType:    repository.BenchmarkArtifactTypePackExport,
							ReportLabel:     repository.BenchmarkReportArtifactLabelPackExport,
							SourceKind:      repository.BenchmarkTokenEstimateSourcePackExportSection,
							ArtifactPath:    "artifacts/pack.json",
							EstimatedTokens: 41,
						},
					},
				}},
			},
		},
	}

	persisted := BenchmarkPersistedArmsFromResult(42, 1, result)
	if len(persisted) != 1 {
		t.Fatalf("len(persisted) = %d, want 1", len(persisted))
	}

	var runMetadata struct {
		WorkspacePath         string                                    `json:"workspacePath"`
		TokenEstimateContract repository.BenchmarkTokenEstimateContract `json:"tokenEstimateContract"`
	}
	if err := json.Unmarshal([]byte(persisted[0].Run.MetadataJSON), &runMetadata); err != nil {
		t.Fatalf("run metadata json: %v", err)
	}
	if runMetadata.WorkspacePath != "/tmp/benchmark/optimusctx" {
		t.Fatalf("workspacePath = %q", runMetadata.WorkspacePath)
	}
	if runMetadata.TokenEstimateContract.Policy.Name != repository.BenchmarkTokenEstimatorPolicyName {
		t.Fatalf("estimate policy = %q", runMetadata.TokenEstimateContract.Policy.Name)
	}

	var laneMetadata struct {
		Attribution []repository.BenchmarkArtifactConsumption `json:"attribution"`
	}
	if err := json.Unmarshal([]byte(persisted[0].Samples[0].Sample.MetadataJSON), &laneMetadata); err != nil {
		t.Fatalf("lane metadata json: %v", err)
	}
	if len(laneMetadata.Attribution) != 1 {
		t.Fatalf("lane attribution = %+v, want 1 record", laneMetadata.Attribution)
	}
	if laneMetadata.Attribution[0].ReportLabel != repository.BenchmarkReportArtifactLabelPackExport {
		t.Fatalf("report label = %q", laneMetadata.Attribution[0].ReportLabel)
	}
	if laneMetadata.Attribution[0].SourceKind != repository.BenchmarkTokenEstimateSourcePackExportSection {
		t.Fatalf("source kind = %q", laneMetadata.Attribution[0].SourceKind)
	}
}

func TestBenchmarkComparisonSummary(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}
	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	result := repository.BenchmarkRunResult{
		SchemaVersion: repository.BenchmarkSuiteSchemaV1,
		SuiteID:       "go-benchmark-refresh-v1",
		SuiteVersion:  "v1",
		FixtureID:     "go-worktree",
		FixturePath:   "go-worktree/v1/repository",
		WorkspacePath: "/tmp/benchmark",
		Arms: []repository.BenchmarkArmRunResult{
			{
				Kind:       repository.BenchmarkArmKindBaseline,
				Name:       "Baseline",
				Workspace:  "/tmp/benchmark/baseline",
				StartedAt:  time.Date(2026, 3, 16, 18, 0, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, 3, 16, 18, 0, 3, 0, time.UTC),
				LaneResults: []repository.BenchmarkLaneRunResult{{
					Lane:          repository.BenchmarkLaneRefreshReady,
					StartMarker:   "refresh_after_change_started",
					SuccessMarker: "refresh_ready",
					StopMarker:    "refresh_ready",
					StartedAt:     time.Date(2026, 3, 16, 18, 0, 0, 0, time.UTC),
					FinishedAt:    time.Date(2026, 3, 16, 18, 0, 1, 0, time.UTC),
					Elapsed:       time.Second,
					Success:       true,
					Effort: repository.BenchmarkLaneEffort{
						ActionCount:        2,
						ConsultedArtifacts: []string{"docs/notes.txt"},
					},
				}},
			},
			{
				Kind:       repository.BenchmarkArmKindOptimusCtx,
				Name:       "OptimusCtx",
				Workspace:  "/tmp/benchmark/optimusctx",
				StartedAt:  time.Date(2026, 3, 16, 18, 0, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, 3, 16, 18, 0, 2, 0, time.UTC),
				LaneResults: []repository.BenchmarkLaneRunResult{{
					Lane:          repository.BenchmarkLaneRefreshReady,
					StartMarker:   "refresh_after_change_started",
					SuccessMarker: "refresh_ready",
					StopMarker:    "refresh_ready",
					StartedAt:     time.Date(2026, 3, 16, 18, 0, 0, 0, time.UTC),
					FinishedAt:    time.Date(2026, 3, 16, 18, 0, 1, 0, time.UTC),
					Elapsed:       time.Second,
					Success:       true,
					Effort: repository.BenchmarkLaneEffort{
						ActionCount:        1,
						ConsultedArtifacts: []string{"docs/notes.txt", ".optimusctx/state.json"},
					},
				}},
			},
		},
	}

	for attempt := 1; attempt <= 2; attempt++ {
		for _, arm := range BenchmarkPersistedArmsFromResult(repoID, attempt, result) {
			if _, _, err := store.SaveBenchmarkRun(ctx, arm.Run, arm.Samples); err != nil {
				t.Fatalf("SaveBenchmarkRun(attempt=%d) error = %v", attempt, err)
			}
		}
	}

	if got, want := mustNextBenchmarkAttempt(t, store, ctx, repoID, result.SuiteID, result.SuiteVersion), 3; got != want {
		t.Fatalf("NextBenchmarkAttempt() = %d, want %d", got, want)
	}

	runs, err := store.ListBenchmarkRuns(ctx, repoID, result.SuiteID, result.SuiteVersion)
	if err != nil {
		t.Fatalf("ListBenchmarkRuns() error = %v", err)
	}
	if got, want := len(runs), 4; got != want {
		t.Fatalf("len(runs) = %d, want %d", got, want)
	}
	if runs[0].Run.Attempt != 1 || runs[2].Run.Attempt != 2 {
		t.Fatalf("attempt ordering = [%d %d %d %d], want grouped attempts", runs[0].Run.Attempt, runs[1].Run.Attempt, runs[2].Run.Attempt, runs[3].Run.Attempt)
	}
}

func TestBenchmarkEvidenceBundleSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	assertTablesExist(t, store.DB(), "benchmark_evidence_bundles", "benchmark_evidence_lane_summaries", "benchmark_evidence_attributions")
	assertIndexColumns(t, store.DB(), "benchmark_evidence_bundles", []string{"repository_id", "suite_id", "suite_version", "generated_at"})
	assertIndexColumns(t, store.DB(), "benchmark_evidence_lane_summaries", []string{"benchmark_evidence_bundle_id", "arm_kind", "lane"})
	assertIndexColumns(t, store.DB(), "benchmark_evidence_attributions", []string{"benchmark_evidence_bundle_id", "attempt", "arm_kind", "lane"})

	bundle := testBenchmarkEvidenceBundle(layout.RepoRoot)
	saved, err := store.SaveBenchmarkEvidenceBundle(ctx, repoID, bundle)
	if err != nil {
		t.Fatalf("SaveBenchmarkEvidenceBundle() error = %v", err)
	}
	if saved.SchemaVersion != repository.BenchmarkEvidenceBundleSchemaV1 {
		t.Fatalf("saved schema version = %q", saved.SchemaVersion)
	}

	var bundleCount int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM benchmark_evidence_bundles`).Scan(&bundleCount); err != nil {
		t.Fatalf("count benchmark_evidence_bundles: %v", err)
	}
	if bundleCount != 1 {
		t.Fatalf("bundle count = %d, want 1", bundleCount)
	}

	var attributionCount int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM benchmark_evidence_attributions`).Scan(&attributionCount); err != nil {
		t.Fatalf("count benchmark_evidence_attributions: %v", err)
	}
	if attributionCount != 2 {
		t.Fatalf("attribution count = %d, want 2", attributionCount)
	}
}

func TestBenchmarkExportDeterminism(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	bundle := testBenchmarkEvidenceBundle(layout.RepoRoot)
	if _, err := store.SaveBenchmarkEvidenceBundle(ctx, repoID, bundle); err != nil {
		t.Fatalf("SaveBenchmarkEvidenceBundle() error = %v", err)
	}

	got, err := store.LoadLatestBenchmarkEvidenceBundle(ctx, repoID, bundle.SuiteID, bundle.SuiteVersion)
	if err != nil {
		t.Fatalf("LoadLatestBenchmarkEvidenceBundle() error = %v", err)
	}
	if !reflect.DeepEqual(got, repository.NormalizeBenchmarkEvidenceBundle(bundle)) {
		t.Fatalf("bundle mismatch\n got=%+v\nwant=%+v", got, repository.NormalizeBenchmarkEvidenceBundle(bundle))
	}

	firstJSON, err := repository.MarshalBenchmarkEvidenceBundle(bundle)
	if err != nil {
		t.Fatalf("MarshalBenchmarkEvidenceBundle(bundle) error = %v", err)
	}
	secondJSON, err := repository.MarshalBenchmarkEvidenceBundle(got)
	if err != nil {
		t.Fatalf("MarshalBenchmarkEvidenceBundle(loaded) error = %v", err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("bundle JSON drifted after reload\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}
	if !strings.Contains(string(firstJSON), `"reportLabel": "pack_export"`) {
		t.Fatalf("bundle JSON missing report labels: %s", firstJSON)
	}
}

func mustNextBenchmarkAttempt(t *testing.T, store *Store, ctx context.Context, repositoryID int64, suiteID string, suiteVersion string) int {
	t.Helper()

	attempt, err := store.NextBenchmarkAttempt(ctx, repositoryID, suiteID, suiteVersion)
	if err != nil {
		t.Fatalf("NextBenchmarkAttempt() error = %v", err)
	}
	return attempt
}

func testBenchmarkEvidenceBundle(repositoryRoot string) repository.BenchmarkEvidenceBundle {
	return repository.NormalizeBenchmarkEvidenceBundle(repository.BenchmarkEvidenceBundle{
		SchemaVersion:          repository.BenchmarkEvidenceBundleSchemaV1,
		GeneratedAt:            time.Date(2026, 3, 16, 19, 0, 0, 0, time.UTC),
		RepositoryRoot:         repositoryRoot,
		SuiteID:                "go-benchmark-refresh-v1",
		SuiteVersion:           "v1",
		FixtureID:              "go-worktree",
		FixturePath:            "go-worktree/v1/repository",
		TokenEstimateContract:  repository.DefaultBenchmarkTokenEstimateContract(),
		MethodologyFingerprint: "go-benchmark-refresh-v1|baseline|optimusctx",
		RerunCommand:           "go run ./cmd/optimusctx eval benchmark export --suite go-benchmark-refresh-v1 --attempts 2",
		Verification: repository.BenchmarkEvidenceVerification{
			Passed: true,
		},
		Comparison: []repository.BenchmarkEvidenceArmSummary{
			{
				ArmKind: repository.BenchmarkArmKindOptimusCtx,
				ArmName: "OptimusCtx",
				Lanes: []repository.BenchmarkEvidenceLaneSummary{{
					Lane:                  repository.BenchmarkLaneContextAssembly,
					AttemptCount:          1,
					SuccessCount:          1,
					ElapsedMS:             repository.BenchmarkEvidenceInt64Stats{Min: 1000, Max: 1000, Median: 1000, Mean: 1000},
					ActionCount:           repository.BenchmarkEvidenceInt64Stats{Min: 2, Max: 2, Median: 2, Mean: 2},
					TargetedLookupActions: repository.BenchmarkEvidenceInt64Stats{Min: 1, Max: 1, Median: 1, Mean: 1},
					BytesRead:             repository.BenchmarkEvidenceInt64Stats{Min: 512, Max: 512, Median: 512, Mean: 512},
					ConsultedArtifacts:    []string{"artifacts/pack.json", "docs/notes.txt"},
				}},
			},
		},
		Attempts: []repository.BenchmarkEvidenceAttempt{{
			Attempt: 1,
			Arms: []repository.BenchmarkEvidenceArmAttempt{{
				Kind:          repository.BenchmarkArmKindOptimusCtx,
				Name:          "OptimusCtx",
				WorkspacePath: repositoryRoot,
				StartedAt:     time.Date(2026, 3, 16, 19, 0, 0, 0, time.UTC),
				FinishedAt:    time.Date(2026, 3, 16, 19, 0, 3, 0, time.UTC),
				Lanes: []repository.BenchmarkEvidenceLane{{
					Lane:          repository.BenchmarkLaneContextAssembly,
					StartMarker:   "context_started",
					SuccessMarker: "context_ready",
					StopMarker:    "context_ready",
					StartedAt:     time.Date(2026, 3, 16, 19, 0, 1, 0, time.UTC),
					FinishedAt:    time.Date(2026, 3, 16, 19, 0, 2, 0, time.UTC),
					ElapsedMS:     1000,
					Success:       true,
					EvidencePaths: []string{"docs/notes.txt", "artifacts/pack.json"},
					Effort: repository.BenchmarkLaneEffort{
						ActionCount:           2,
						TargetedLookupActions: 1,
						BytesRead:             512,
						ConsultedArtifacts:    []string{"artifacts/pack.json", "docs/notes.txt"},
					},
					Attribution: []repository.BenchmarkArtifactConsumption{
						{
							StepID:          "context-pack",
							StepName:        "Export context pack",
							Lane:            repository.BenchmarkLaneContextAssembly,
							Surface:         repository.BenchmarkTreatmentSurfaceCLI,
							Command:         repository.EvalCommandPackExport,
							ArtifactType:    repository.BenchmarkArtifactTypePackExport,
							ReportLabel:     repository.BenchmarkReportArtifactLabelPackExport,
							SourceKind:      repository.BenchmarkTokenEstimateSourcePackExportSection,
							ArtifactPath:    "artifacts/pack.json",
							EstimatedBytes:  512,
							EstimatedTokens: 128,
						},
						{
							StepID:          "health-check",
							StepName:        "Inspect health",
							Lane:            repository.BenchmarkLaneContextAssembly,
							Surface:         repository.BenchmarkTreatmentSurfaceMCP,
							Tool:            "optimusctx.health",
							ArtifactType:    repository.BenchmarkArtifactTypeHealth,
							ReportLabel:     repository.BenchmarkReportArtifactLabelOperational,
							SourceKind:      repository.BenchmarkTokenEstimateSourceDirectPayload,
							ArtifactPath:    ".optimusctx/state.json",
							EstimatedBytes:  64,
							EstimatedTokens: 16,
						},
					},
				}},
			}},
		}},
	})
}
