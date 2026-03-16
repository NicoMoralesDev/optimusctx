package app

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

func TestBenchmarkExportPersistence(t *testing.T) {
	t.Parallel()

	repoRoot, suitePath, store, repoID := initBenchmarkEvidenceStore(t, true)
	defer store.Close()

	service := NewBenchmarkService()
	bundle, err := service.ExportEvidenceBundle(context.Background(), BenchmarkEvidenceBundleRequest{
		StartPath:    repoRoot,
		SuitePath:    suitePath,
		FixturesRoot: filepath.Join(repoRoot, "fixtures"),
	})
	if err != nil {
		t.Fatalf("ExportEvidenceBundle() error = %v", err)
	}
	if bundle.SchemaVersion != repository.BenchmarkEvidenceBundleSchemaV2 {
		t.Fatalf("SchemaVersion = %q", bundle.SchemaVersion)
	}
	if bundle.Methodology.SuiteSchemaVersion != repository.BenchmarkSuiteSchemaV2 {
		t.Fatalf("Methodology.SuiteSchemaVersion = %q", bundle.Methodology.SuiteSchemaVersion)
	}
	if bundle.MethodologyFingerprint == "" {
		t.Fatal("MethodologyFingerprint should not be empty")
	}
	if !strings.Contains(bundle.RerunCommand, "optimusctx eval benchmark export --suite-file") {
		t.Fatalf("RerunCommand = %q", bundle.RerunCommand)
	}

	loaded, err := store.LoadLatestBenchmarkEvidenceBundle(context.Background(), repoID, bundle.SuiteID, bundle.SuiteVersion)
	if err != nil {
		t.Fatalf("LoadLatestBenchmarkEvidenceBundle() error = %v", err)
	}
	if loaded.MethodologyFingerprint != bundle.MethodologyFingerprint {
		t.Fatalf("loaded fingerprint = %q, want %q", loaded.MethodologyFingerprint, bundle.MethodologyFingerprint)
	}
	if err := loaded.Methodology.Boundary.Validate(); err != nil {
		t.Fatalf("loaded methodology boundary should validate: %v", err)
	}
}

func TestBenchmarkExportRejectsAmbiguousSemantics(t *testing.T) {
	t.Parallel()

	repoRoot, suitePath, store, _ := initBenchmarkEvidenceStore(t, false)
	defer store.Close()

	service := NewBenchmarkService()
	_, err := service.ExportEvidenceBundle(context.Background(), BenchmarkEvidenceBundleRequest{
		StartPath:    repoRoot,
		SuitePath:    suitePath,
		FixturesRoot: filepath.Join(repoRoot, "fixtures"),
	})
	if err == nil || !strings.Contains(err.Error(), "missing attribution boundary") {
		t.Fatalf("ExportEvidenceBundle() error = %v, want ambiguity rejection", err)
	}
}

func TestBenchmarkHumanSummaryInputs(t *testing.T) {
	t.Parallel()

	suite := benchmarkServiceSuite()
	attempts := []BenchmarkAttemptResult{
		{Attempt: 1, Result: benchmarkEvidenceRunResult("/tmp/repo", true)},
		{Attempt: 2, Result: benchmarkEvidenceRunResult("/tmp/repo", true)},
	}
	summary := summarizeBenchmarkAttempts(attempts, suite)
	bundle := buildBenchmarkEvidenceBundle("/tmp/repo", repository.DefaultBenchmarkTokenEstimateContract(), suite, summary, attempts, "go run ./cmd/optimusctx eval benchmark report --suite-file /tmp/suite.json --attempts 2")

	human := BuildBenchmarkHumanSummary(bundle)
	if human.AttemptCount != 2 {
		t.Fatalf("AttemptCount = %d, want 2", human.AttemptCount)
	}
	if got, want := len(human.LaneComparisons), 2; got != want {
		t.Fatalf("len(LaneComparisons) = %d, want %d", got, want)
	}
	var foundRefresh bool
	var foundTaskCompletion bool
	for _, lane := range human.LaneComparisons {
		switch lane.Lane {
		case repository.BenchmarkLaneRefreshReady:
			foundRefresh = true
			if lane.BaselineEstimatedTokens.Median != 12 {
				t.Fatalf("baseline refresh median tokens = %d, want 12", lane.BaselineEstimatedTokens.Median)
			}
			if lane.TreatmentEstimatedTokens.Median != 12 {
				t.Fatalf("treatment refresh median tokens = %d, want 12", lane.TreatmentEstimatedTokens.Median)
			}
		case repository.BenchmarkLaneTaskCompletion:
			foundTaskCompletion = true
			if lane.TreatmentEstimatedTokens.Median != 18 {
				t.Fatalf("treatment task-completion median tokens = %d, want 18", lane.TreatmentEstimatedTokens.Median)
			}
		}
	}
	if !foundRefresh || !foundTaskCompletion {
		t.Fatalf("missing expected lane comparisons: %+v", human.LaneComparisons)
	}
	if got, want := len(human.AttributionRows), 2; got != want {
		t.Fatalf("len(AttributionRows) = %d, want %d", got, want)
	}
	labels := []string{human.AttributionRows[0].DisplayLabel, human.AttributionRows[1].DisplayLabel}
	if !(slices.Contains(labels, "Operational") && slices.Contains(labels, "L2 Context")) {
		t.Fatalf("DisplayLabels = %+v, want Operational + L2 Context", labels)
	}
}

func TestBenchmarkComparisonReportRendering(t *testing.T) {
	t.Parallel()

	suite := benchmarkServiceSuite()
	attempts := []BenchmarkAttemptResult{
		{Attempt: 1, Result: benchmarkEvidenceRunResult("/tmp/repo", true)},
		{Attempt: 2, Result: benchmarkEvidenceRunResult("/tmp/repo", true)},
	}
	summary := summarizeBenchmarkAttempts(attempts, suite)
	bundle := buildBenchmarkEvidenceBundle("/tmp/repo", repository.DefaultBenchmarkTokenEstimateContract(), suite, summary, attempts, "go run ./cmd/optimusctx eval benchmark report --suite-file /tmp/suite.json --attempts 2")

	report := RenderBenchmarkComparisonReport(BuildBenchmarkHumanSummary(bundle))
	for _, fragment := range []string{
		"benchmark report",
		"lane comparison",
		"Refresh After Change",
		"Task Completion",
		"Operational",
		"L2 Context",
		"estimated tokens use bytes_div_4_ceiling",
		"not provider-billed token invoices",
		"fixes benchmark truthfulness",
	} {
		if !strings.Contains(report, fragment) {
			t.Fatalf("report missing %q:\n%s", fragment, report)
		}
	}
}

func TestBenchmarkReportWordingGuards(t *testing.T) {
	t.Parallel()

	suite := benchmarkServiceSuite()
	attempts := []BenchmarkAttemptResult{
		{Attempt: 1, Result: benchmarkEvidenceRunResult("/tmp/repo", true)},
		{Attempt: 2, Result: benchmarkEvidenceRunResult("/tmp/repo", true)},
	}
	summary := summarizeBenchmarkAttempts(attempts, suite)
	bundle := buildBenchmarkEvidenceBundle("/tmp/repo", repository.DefaultBenchmarkTokenEstimateContract(), suite, summary, attempts, "go run ./cmd/optimusctx eval benchmark verify --suite-file /tmp/suite.json --attempts 2")
	if reasons := verifyBenchmarkReportWording(RenderBenchmarkComparisonReport(BuildBenchmarkHumanSummary(bundle))); len(reasons) != 0 {
		t.Fatalf("verifyBenchmarkReportWording() = %+v", reasons)
	}
}

func TestBenchmarkMethodologyFingerprint(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	suite := benchmarkServiceSuite()
	attempts := []BenchmarkAttemptResult{
		{Attempt: 1, Result: benchmarkEvidenceRunResult(repoRoot, true)},
		{Attempt: 2, Result: benchmarkEvidenceRunResult(repoRoot, true)},
	}
	summary := summarizeBenchmarkAttempts(attempts, suite)
	bundle := buildBenchmarkEvidenceBundle(repoRoot, repository.DefaultBenchmarkTokenEstimateContract(), suite, summary, attempts, "go run ./cmd/optimusctx eval benchmark verify --suite-file /tmp/suite.json --attempts 2")

	driftedSuite := benchmarkServiceSuite()
	driftedSuite.CountedInputs[1].JSONPath = "refresh.status"
	driftedSummary := summarizeBenchmarkAttempts(attempts, driftedSuite)
	driftedBundle := buildBenchmarkEvidenceBundle(repoRoot, repository.DefaultBenchmarkTokenEstimateContract(), driftedSuite, driftedSummary, attempts, "go run ./cmd/optimusctx eval benchmark verify --suite-file /tmp/suite.json --attempts 2")

	reasons := compareBenchmarkEvidenceBundles(bundle, driftedBundle)
	if len(reasons) == 0 {
		t.Fatal("compareBenchmarkEvidenceBundles() should report methodology drift")
	}
	if !strings.Contains(strings.Join(reasons, " "), "methodology") {
		t.Fatalf("reasons = %+v", reasons)
	}
}

func TestBuildBenchmarkEvidenceBundleFromPersistedRuns(t *testing.T) {
	t.Parallel()

	repoRoot, _, store, repoID := initBenchmarkEvidenceStore(t, true)
	defer store.Close()

	persisted, err := store.ListBenchmarkRuns(context.Background(), repoID, "go-benchmark-refresh-v2", "v2")
	if err != nil {
		t.Fatalf("ListBenchmarkRuns() error = %v", err)
	}
	bundle, err := buildBenchmarkEvidenceBundleFromPersistedRuns(repoRoot, persisted, benchmarkServiceSuite(), "go run ./cmd/optimusctx eval benchmark verify --suite-file /tmp/suite.json --attempts 2")
	if err != nil {
		t.Fatalf("buildBenchmarkEvidenceBundleFromPersistedRuns() error = %v", err)
	}
	if bundle.MethodologyFingerprint == "" {
		t.Fatal("MethodologyFingerprint should not be empty")
	}
	if bundle.SchemaVersion != repository.BenchmarkEvidenceBundleSchemaV2 {
		t.Fatalf("SchemaVersion = %q", bundle.SchemaVersion)
	}
	if len(bundle.Attempts) != 2 {
		t.Fatalf("len(bundle.Attempts) = %d, want 2", len(bundle.Attempts))
	}
}

func initBenchmarkEvidenceStore(t *testing.T, withBoundaries bool) (string, string, *sqlite.Store, int64) {
	t.Helper()

	repoRoot := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, output)
	}

	fixturePath := filepath.Join(repoRoot, "fixtures", "go-worktree", "v2", "repository")
	if err := os.MkdirAll(fixturePath, 0o755); err != nil {
		t.Fatalf("MkdirAll(fixturePath) error = %v", err)
	}
	suitePath := filepath.Join(repoRoot, "fixtures", "benchmark-suite-v2.json")
	writeBenchmarkSuite(t, suitePath, benchmarkServiceSuite())

	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}
	store, err := sqlite.OpenOrCreateStore(context.Background(), layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("OpenOrCreateStore() error = %v", err)
	}
	record, err := store.UpsertRepository(context.Background(), repository.RepositoryRoot{
		RootPath:      repoRoot,
		DetectionMode: repository.DetectionModeGit,
		Fingerprint: repository.RepositoryFingerprint{
			RootPath:     repoRoot,
			GitCommonDir: filepath.Join(repoRoot, ".git"),
		},
	}, time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("UpsertRepository() error = %v", err)
	}

	for attempt := 1; attempt <= 2; attempt++ {
		for _, persisted := range sqlite.BenchmarkPersistedArmsFromResult(record.ID, attempt, benchmarkEvidenceRunResult(repoRoot, withBoundaries)) {
			if _, _, err := store.SaveBenchmarkRun(context.Background(), persisted.Run, persisted.Samples); err != nil {
				t.Fatalf("SaveBenchmarkRun(attempt=%d) error = %v", attempt, err)
			}
		}
	}
	return repoRoot, suitePath, store, record.ID
}

func benchmarkServiceSuite() repository.BenchmarkSuiteDefinition {
	return repository.BenchmarkSuiteDefinition{
		SchemaVersion: repository.BenchmarkSuiteSchemaV2,
		ID:            "go-benchmark-refresh-v2",
		Version:       "v2",
		Name:          "Go benchmark refresh and task completion",
		Boundary:      repository.DefaultBenchmarkBoundaryContract(),
		Fixture: repository.EvalFixtureRef{
			ID:          "go-worktree",
			Version:     "v2",
			Path:        "go-worktree/v2/repository",
			Materialize: repository.EvalFixtureModeCopyTree,
		},
		Task: repository.BenchmarkTaskDefinition{
			ID:         "docs-pack",
			Prompt:     "Refresh after mutation and export bounded context.",
			TargetPath: "docs/notes.txt",
		},
		CountedInputs: []repository.BenchmarkCountedInputDefinition{
			{
				ID:         "baseline-readiness-paths",
				ArmKind:    repository.BenchmarkArmKindBaseline,
				Lane:       repository.BenchmarkLaneRefreshReady,
				StepID:     "baseline-readiness",
				Name:       "Baseline readiness path hints",
				Kind:       repository.BenchmarkCountedInputKindPathList,
				SourceKind: repository.BenchmarkTokenEstimateSourcePathEstimate,
				Path:       "docs/notes.txt",
			},
			{
				ID:           "treatment-health-summary",
				ArmKind:      repository.BenchmarkArmKindOptimusCtx,
				Lane:         repository.BenchmarkLaneRefreshReady,
				StepID:       "health",
				Name:         "Projected health summary",
				Kind:         repository.BenchmarkCountedInputKindJSONFieldProjection,
				SourceKind:   repository.BenchmarkTokenEstimateSourceDirectPayload,
				ArtifactType: repository.BenchmarkArtifactTypeHealth,
				ReportLabel:  repository.BenchmarkReportArtifactLabelOperational,
				JSONPath:     "refresh.freshness",
			},
			{
				ID:         "baseline-task-slice",
				ArmKind:    repository.BenchmarkArmKindBaseline,
				Lane:       repository.BenchmarkLaneTaskCompletion,
				StepID:     "baseline-context",
				Name:       "Baseline updated notes slice",
				Kind:       repository.BenchmarkCountedInputKindFileSlice,
				SourceKind: repository.BenchmarkTokenEstimateSourceBoundedFileContent,
				Path:       "docs/notes.txt",
				StartLine:  1,
				EndLine:    20,
			},
			{
				ID:           "treatment-updated-context",
				ArmKind:      repository.BenchmarkArmKindOptimusCtx,
				Lane:         repository.BenchmarkLaneTaskCompletion,
				StepID:       "context",
				Name:         "Treatment updated notes context",
				Kind:         repository.BenchmarkCountedInputKindTextOutput,
				SourceKind:   repository.BenchmarkTokenEstimateSourceDirectPayload,
				ArtifactType: repository.BenchmarkArtifactTypeL2Context,
				ReportLabel:  repository.BenchmarkReportArtifactLabelL2Context,
				Path:         "artifacts/updated-notes.txt",
			},
		},
		Lanes: []repository.BenchmarkLaneDefinition{
			{
				Name: repository.BenchmarkLaneRefreshReady,
				Assertions: []repository.BenchmarkAssertion{{
					File:     "docs/notes.txt",
					Kind:     repository.EvalAssertionKindContains,
					Contains: "mutated benchmark note",
				}},
				FinalArtifact: &repository.BenchmarkFinalArtifactContract{
					ID:     "refresh-readiness",
					Name:   "Refresh readiness summary",
					Kind:   repository.BenchmarkFinalArtifactKindReadinessSummary,
					Path:   "artifacts/readiness.json",
					Format: repository.BenchmarkFinalArtifactFormatJSON,
					Normalization: repository.BenchmarkFinalArtifactNormalization{
						Mode:      repository.BenchmarkFinalArtifactNormalizationModeJSONFields,
						JSONPaths: []string{"freshness", "targetReady"},
					},
					Assertions: []repository.BenchmarkFinalArtifactAssertion{
						{Kind: repository.EvalAssertionKindJSONFieldPresent, Path: "freshness"},
					},
				},
				StopCondition: repository.BenchmarkStopCondition{
					Kind:   repository.BenchmarkStopConditionKindMarker,
					Marker: "refresh_ready",
				},
				Metrics: []repository.BenchmarkMetric{repository.BenchmarkMetricConsultedArtifacts},
			},
			{
				Name: repository.BenchmarkLaneTaskCompletion,
				FinalArtifact: &repository.BenchmarkFinalArtifactContract{
					ID:     "updated-notes-context",
					Name:   "Updated notes context",
					Kind:   repository.BenchmarkFinalArtifactKindTaskOutput,
					Path:   "artifacts/updated-notes.txt",
					Format: repository.BenchmarkFinalArtifactFormatText,
					Normalization: repository.BenchmarkFinalArtifactNormalization{
						Mode:           repository.BenchmarkFinalArtifactNormalizationModeTextTrimmed,
						TrimWhitespace: true,
					},
					Assertions: []repository.BenchmarkFinalArtifactAssertion{
						{Kind: repository.EvalAssertionKindContains, Contains: "mutated benchmark note"},
					},
				},
				Assertions: []repository.BenchmarkAssertion{{
					File:     "docs/notes.txt",
					Kind:     repository.EvalAssertionKindContains,
					Contains: "mutated benchmark note",
				}},
				StopCondition: repository.BenchmarkStopCondition{
					Kind:   repository.BenchmarkStopConditionKindMarker,
					Marker: "task_complete",
				},
				Metrics: []repository.BenchmarkMetric{repository.BenchmarkMetricFileReadActions},
			},
		},
		Arms: []repository.BenchmarkArmDefinition{
			{
				Kind: repository.BenchmarkArmKindBaseline,
				Name: "Baseline",
				Steps: []repository.BenchmarkStep{
					{
						ID:   "baseline-readiness",
						Name: "Search mutated note",
						Lane: repository.BenchmarkLaneRefreshReady,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:  repository.BenchmarkBaselineActionGitGrep,
							Query: "mutated benchmark note",
						},
					},
					{
						ID:   "baseline-ready",
						Name: "Mark refresh ready",
						Lane: repository.BenchmarkLaneRefreshReady,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:   repository.BenchmarkBaselineActionMarkLaneComplete,
							Marker: "refresh_ready",
						},
					},
					{
						ID:   "baseline-context",
						Name: "Read docs slice",
						Lane: repository.BenchmarkLaneTaskCompletion,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:      repository.BenchmarkBaselineActionReadFileSlice,
							Path:      "docs/notes.txt",
							StartLine: 1,
							EndLine:   20,
						},
					},
					{
						ID:   "baseline-done",
						Name: "Mark task complete",
						Lane: repository.BenchmarkLaneTaskCompletion,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:   repository.BenchmarkBaselineActionMarkLaneComplete,
							Marker: "task_complete",
						},
					},
				},
			},
			{
				Kind: repository.BenchmarkArmKindOptimusCtx,
				Name: "OptimusCtx",
				Steps: []repository.BenchmarkStep{
					{
						ID:   "refresh",
						Name: "Refresh repository",
						Lane: repository.BenchmarkLaneRefreshReady,
						Treatment: &repository.BenchmarkTreatmentAction{
							Surface: repository.BenchmarkTreatmentSurfaceCLI,
							Command: repository.EvalCommandRefresh,
						},
					},
					{
						ID:   "health",
						Name: "Check repository health",
						Lane: repository.BenchmarkLaneRefreshReady,
						Treatment: &repository.BenchmarkTreatmentAction{
							Surface: repository.BenchmarkTreatmentSurfaceMCP,
							Tool:    "optimusctx.health",
						},
					},
					{
						ID:   "context",
						Name: "Fetch bounded updated notes context",
						Lane: repository.BenchmarkLaneTaskCompletion,
						Treatment: &repository.BenchmarkTreatmentAction{
							Surface: repository.BenchmarkTreatmentSurfaceMCP,
							Tool:    "optimusctx.targeted_context",
						},
					},
				},
			},
		},
	}
}

func benchmarkEvidenceRunResult(repositoryRoot string, withBoundaries bool) repository.BenchmarkRunResult {
	boundary := func(value repository.BenchmarkEvidenceBoundary) repository.BenchmarkEvidenceBoundary {
		if withBoundaries {
			return value
		}
		return ""
	}

	return repository.BenchmarkRunResult{
		SchemaVersion: repository.BenchmarkSuiteSchemaV2,
		SuiteID:       "go-benchmark-refresh-v2",
		SuiteVersion:  "v2",
		FixtureID:     "go-worktree",
		FixturePath:   "go-worktree/v2/repository",
		WorkspacePath: repositoryRoot,
		Arms: []repository.BenchmarkArmRunResult{
			{
				Kind:       repository.BenchmarkArmKindBaseline,
				Name:       "Baseline",
				Workspace:  repositoryRoot,
				StartedAt:  time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, 3, 16, 20, 0, 2, 0, time.UTC),
				LaneResults: []repository.BenchmarkLaneRunResult{
					{
						Lane:          repository.BenchmarkLaneRefreshReady,
						StartMarker:   "refresh_after_change_started",
						SuccessMarker: "refresh_ready",
						StopMarker:    "refresh_ready",
						StartedAt:     time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC),
						FinishedAt:    time.Date(2026, 3, 16, 20, 0, 1, 0, time.UTC),
						Elapsed:       time.Second,
						Success:       true,
						EvidencePaths: []string{"docs/notes.txt"},
						FinalArtifact: &repository.BenchmarkLaneFinalArtifactVerification{
							ContractID: "refresh-readiness",
							Path:       "artifacts/readiness.json",
							Passed:     true,
						},
						Effort: repository.BenchmarkLaneEffort{
							ActionCount:        2,
							BytesRead:          128,
							ConsultedArtifacts: []string{"docs/notes.txt"},
						},
						Attribution: []repository.BenchmarkArtifactConsumption{
							{
								StepID:          "baseline-readiness",
								StepName:        "Search docs note",
								Lane:            repository.BenchmarkLaneRefreshReady,
								Boundary:        boundary(repository.BenchmarkEvidenceBoundaryAgentInput),
								SourceKind:      repository.BenchmarkTokenEstimateSourcePathEstimate,
								ArtifactPath:    "docs/notes.txt",
								EstimatedBytes:  48,
								EstimatedTokens: 12,
							},
						},
					},
					{
						Lane:          repository.BenchmarkLaneTaskCompletion,
						StartMarker:   "task_completion_started",
						SuccessMarker: "task_complete",
						StopMarker:    "task_complete",
						StartedAt:     time.Date(2026, 3, 16, 20, 0, 1, 0, time.UTC),
						FinishedAt:    time.Date(2026, 3, 16, 20, 0, 2, 0, time.UTC),
						Elapsed:       time.Second,
						Success:       true,
						EvidencePaths: []string{"docs/notes.txt"},
						FinalArtifact: &repository.BenchmarkLaneFinalArtifactVerification{
							ContractID: "updated-notes-context",
							Path:       "artifacts/updated-notes.txt",
							Passed:     true,
						},
						Effort: repository.BenchmarkLaneEffort{
							ActionCount:        1,
							FileReadActions:    1,
							BytesRead:          512,
							ConsultedArtifacts: []string{"docs/notes.txt"},
						},
						Attribution: []repository.BenchmarkArtifactConsumption{
							{
								StepID:          "baseline-context",
								StepName:        "Read docs slice",
								Lane:            repository.BenchmarkLaneTaskCompletion,
								Boundary:        boundary(repository.BenchmarkEvidenceBoundaryAgentInput),
								SourceKind:      repository.BenchmarkTokenEstimateSourceBoundedFileContent,
								ArtifactPath:    "docs/notes.txt",
								EstimatedBytes:  88,
								EstimatedTokens: 22,
							},
						},
					},
				},
			},
			{
				Kind:       repository.BenchmarkArmKindOptimusCtx,
				Name:       "OptimusCtx",
				Workspace:  repositoryRoot,
				StartedAt:  time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, 3, 16, 20, 0, 2, 0, time.UTC),
				LaneResults: []repository.BenchmarkLaneRunResult{
					{
						Lane:          repository.BenchmarkLaneRefreshReady,
						StartMarker:   "refresh_after_change_started",
						SuccessMarker: "refresh_ready",
						StopMarker:    "refresh_ready",
						StartedAt:     time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC),
						FinishedAt:    time.Date(2026, 3, 16, 20, 0, 1, 0, time.UTC),
						Elapsed:       time.Second,
						Success:       true,
						EvidencePaths: []string{"docs/notes.txt", "stdout"},
						FinalArtifact: &repository.BenchmarkLaneFinalArtifactVerification{
							ContractID: "refresh-readiness",
							Path:       "artifacts/readiness.json",
							Passed:     true,
						},
						Effort: repository.BenchmarkLaneEffort{
							ActionCount:           2,
							TargetedLookupActions: 2,
							ConsultedArtifacts:    []string{"docs/notes.txt", "stdout"},
						},
						Attribution: []repository.BenchmarkArtifactConsumption{
							{
								StepID:          "refresh",
								StepName:        "Refresh repository",
								Lane:            repository.BenchmarkLaneRefreshReady,
								Boundary:        boundary(repository.BenchmarkEvidenceBoundarySystemProvenance),
								Surface:         repository.BenchmarkTreatmentSurfaceCLI,
								Command:         repository.EvalCommandRefresh,
								ArtifactType:    repository.BenchmarkArtifactTypeRefresh,
								ReportLabel:     repository.BenchmarkReportArtifactLabelOperational,
								SourceKind:      repository.BenchmarkTokenEstimateSourceDirectPayload,
								ArtifactPath:    "stdout",
								EstimatedBytes:  32,
								EstimatedTokens: 8,
							},
							{
								StepID:          "health",
								StepName:        "Check repository health",
								Lane:            repository.BenchmarkLaneRefreshReady,
								Boundary:        boundary(repository.BenchmarkEvidenceBoundaryAgentInput),
								Surface:         repository.BenchmarkTreatmentSurfaceMCP,
								Tool:            "optimusctx.health",
								ArtifactType:    repository.BenchmarkArtifactTypeHealth,
								ReportLabel:     repository.BenchmarkReportArtifactLabelOperational,
								SourceKind:      repository.BenchmarkTokenEstimateSourceDirectPayload,
								ArtifactPath:    "artifacts/readiness.json",
								EstimatedBytes:  48,
								EstimatedTokens: 12,
							},
						},
					},
					{
						Lane:          repository.BenchmarkLaneTaskCompletion,
						StartMarker:   "task_completion_started",
						SuccessMarker: "task_complete",
						StopMarker:    "task_complete",
						StartedAt:     time.Date(2026, 3, 16, 20, 0, 1, 0, time.UTC),
						FinishedAt:    time.Date(2026, 3, 16, 20, 0, 2, 0, time.UTC),
						Elapsed:       time.Second,
						Success:       true,
						EvidencePaths: []string{"docs/notes.txt"},
						FinalArtifact: &repository.BenchmarkLaneFinalArtifactVerification{
							ContractID: "updated-notes-context",
							Path:       "artifacts/updated-notes.txt",
							Passed:     true,
						},
						Effort: repository.BenchmarkLaneEffort{
							ActionCount:           1,
							TargetedLookupActions: 1,
							BytesRead:             72,
							ConsultedArtifacts:    []string{"docs/notes.txt"},
						},
						Attribution: []repository.BenchmarkArtifactConsumption{
							{
								StepID:          "context",
								StepName:        "Fetch bounded updated notes context",
								Lane:            repository.BenchmarkLaneTaskCompletion,
								Boundary:        boundary(repository.BenchmarkEvidenceBoundaryAgentInput),
								Surface:         repository.BenchmarkTreatmentSurfaceMCP,
								Tool:            "optimusctx.targeted_context",
								ArtifactType:    repository.BenchmarkArtifactTypeL2Context,
								ReportLabel:     repository.BenchmarkReportArtifactLabelL2Context,
								SourceKind:      repository.BenchmarkTokenEstimateSourceDirectPayload,
								ArtifactPath:    "artifacts/updated-notes.txt",
								EstimatedBytes:  72,
								EstimatedTokens: 18,
							},
						},
					},
				},
			},
		},
	}
}

func writeBenchmarkSuite(t *testing.T, path string, suite repository.BenchmarkSuiteDefinition) {
	t.Helper()

	data, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}
