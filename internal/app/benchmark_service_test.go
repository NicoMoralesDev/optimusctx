package app

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

func TestBenchmarkExportPersistence(t *testing.T) {
	t.Parallel()

	repoRoot, store, repoID := initBenchmarkEvidenceStore(t)
	defer store.Close()

	service := NewBenchmarkService()
	bundle, err := service.ExportEvidenceBundle(context.Background(), BenchmarkEvidenceBundleRequest{
		StartPath:    repoRoot,
		SuiteID:      "go-benchmark-refresh-v1",
		SuitesDir:    filepath.Join("..", "..", "testdata", "eval", "benchmarks"),
		FixturesRoot: filepath.Join("..", "..", "testdata", "eval", "fixtures"),
	})
	if err != nil {
		t.Fatalf("ExportEvidenceBundle() error = %v", err)
	}
	if bundle.SchemaVersion != repository.BenchmarkEvidenceBundleSchemaV1 {
		t.Fatalf("SchemaVersion = %q", bundle.SchemaVersion)
	}
	if bundle.MethodologyFingerprint == "" {
		t.Fatal("MethodologyFingerprint should not be empty")
	}
	if !strings.Contains(bundle.RerunCommand, "optimusctx eval benchmark export --suite go-benchmark-refresh-v1 --attempts 2") {
		t.Fatalf("RerunCommand = %q", bundle.RerunCommand)
	}

	loaded, err := store.LoadLatestBenchmarkEvidenceBundle(context.Background(), repoID, bundle.SuiteID, bundle.SuiteVersion)
	if err != nil {
		t.Fatalf("LoadLatestBenchmarkEvidenceBundle() error = %v", err)
	}
	if loaded.MethodologyFingerprint != bundle.MethodologyFingerprint {
		t.Fatalf("loaded fingerprint = %q, want %q", loaded.MethodologyFingerprint, bundle.MethodologyFingerprint)
	}
}

func TestBenchmarkComparisonExport(t *testing.T) {
	t.Parallel()

	repoRoot, store, _ := initBenchmarkEvidenceStore(t)
	defer store.Close()

	service := NewBenchmarkService()
	bundle, err := service.ExportEvidenceBundle(context.Background(), BenchmarkEvidenceBundleRequest{
		StartPath:    repoRoot,
		SuiteID:      "go-benchmark-refresh-v1",
		SuitesDir:    filepath.Join("..", "..", "testdata", "eval", "benchmarks"),
		FixturesRoot: filepath.Join("..", "..", "testdata", "eval", "fixtures"),
	})
	if err != nil {
		t.Fatalf("ExportEvidenceBundle() error = %v", err)
	}
	if got, want := len(bundle.Comparison), 2; got != want {
		t.Fatalf("len(bundle.Comparison) = %d, want %d", got, want)
	}
	if bundle.TokenEstimateContract.BillingDisambiguator != repository.BenchmarkTokenEstimateBillingDisambiguator {
		t.Fatalf("BillingDisambiguator = %q", bundle.TokenEstimateContract.BillingDisambiguator)
	}
	if len(bundle.Attempts) != 2 {
		t.Fatalf("len(bundle.Attempts) = %d, want 2", len(bundle.Attempts))
	}
	foundPackLabel := false
	for _, attempt := range bundle.Attempts {
		for _, arm := range attempt.Arms {
			for _, lane := range arm.Lanes {
				for _, attribution := range lane.Attribution {
					if attribution.ReportLabel == repository.BenchmarkReportArtifactLabelPackExport {
						foundPackLabel = true
					}
				}
			}
		}
	}
	if !foundPackLabel {
		t.Fatalf("bundle missing pack export attribution labels: %+v", bundle.Attempts)
	}
}

func TestBenchmarkRerunCommandContract(t *testing.T) {
	t.Parallel()

	if got := benchmarkEvidenceRerunCommand("go-benchmark-refresh-v1", "", 2); got != "go run ./cmd/optimusctx eval benchmark export --suite go-benchmark-refresh-v1 --attempts 2" {
		t.Fatalf("benchmarkEvidenceRerunCommand() = %q", got)
	}
	if got := benchmarkEvidenceRerunCommand("", "/tmp/suite.json", 3); got != "go run ./cmd/optimusctx eval benchmark export --suite-file /tmp/suite.json --attempts 3" {
		t.Fatalf("benchmarkEvidenceRerunCommand() suite-file = %q", got)
	}
}

func TestBenchmarkHumanSummaryInputs(t *testing.T) {
	t.Parallel()

	bundle := buildBenchmarkEvidenceBundle("/tmp/repo", repository.DefaultBenchmarkTokenEstimateContract(), summarizeBenchmarkAttempts([]BenchmarkAttemptResult{
		{Attempt: 1, Result: benchmarkEvidenceRunResult("/tmp/repo")},
		{Attempt: 2, Result: benchmarkEvidenceRunResult("/tmp/repo")},
	}, "go-benchmark-refresh-v1", "v1"), []BenchmarkAttemptResult{
		{Attempt: 1, Result: benchmarkEvidenceRunResult("/tmp/repo")},
		{Attempt: 2, Result: benchmarkEvidenceRunResult("/tmp/repo")},
	}, "go run ./cmd/optimusctx eval benchmark report --suite go-benchmark-refresh-v1 --attempts 2")

	summary := BuildBenchmarkHumanSummary(bundle)
	if summary.AttemptCount != 2 {
		t.Fatalf("AttemptCount = %d, want 2", summary.AttemptCount)
	}
	if got, want := len(summary.LaneComparisons), 2; got != want {
		t.Fatalf("len(LaneComparisons) = %d, want %d", got, want)
	}
	var foundTaskCompletion bool
	for _, lane := range summary.LaneComparisons {
		if lane.Lane != repository.BenchmarkLaneTaskCompletion {
			continue
		}
		foundTaskCompletion = true
		if lane.TreatmentEstimatedTokens.Median != 128 {
			t.Fatalf("treatment median tokens = %d, want 128", lane.TreatmentEstimatedTokens.Median)
		}
	}
	if !foundTaskCompletion {
		t.Fatalf("missing task completion lane: %+v", summary.LaneComparisons)
	}
	if got, want := len(summary.AttributionRows), 1; got != want {
		t.Fatalf("len(AttributionRows) = %d, want %d", got, want)
	}
	if summary.AttributionRows[0].DisplayLabel != "Pack Export" {
		t.Fatalf("DisplayLabel = %q", summary.AttributionRows[0].DisplayLabel)
	}
}

func TestBenchmarkComparisonReportRendering(t *testing.T) {
	t.Parallel()

	bundle := buildBenchmarkEvidenceBundle("/tmp/repo", repository.DefaultBenchmarkTokenEstimateContract(), summarizeBenchmarkAttempts([]BenchmarkAttemptResult{
		{Attempt: 1, Result: benchmarkEvidenceRunResult("/tmp/repo")},
		{Attempt: 2, Result: benchmarkEvidenceRunResult("/tmp/repo")},
	}, "go-benchmark-refresh-v1", "v1"), []BenchmarkAttemptResult{
		{Attempt: 1, Result: benchmarkEvidenceRunResult("/tmp/repo")},
		{Attempt: 2, Result: benchmarkEvidenceRunResult("/tmp/repo")},
	}, "go run ./cmd/optimusctx eval benchmark report --suite go-benchmark-refresh-v1 --attempts 2")

	report := RenderBenchmarkComparisonReport(BuildBenchmarkHumanSummary(bundle))
	for _, fragment := range []string{
		"benchmark report",
		"lane comparison",
		"Refresh After Change",
		"Task Completion",
		"Pack Export",
		"estimated tokens use bytes_div_4_ceiling",
		"not provider-billed token invoices",
	} {
		if !strings.Contains(report, fragment) {
			t.Fatalf("report missing %q:\n%s", fragment, report)
		}
	}
}

func TestBenchmarkReportReusesPersistedEvidence(t *testing.T) {
	t.Parallel()

	repoRoot, store, repoID := initBenchmarkEvidenceStore(t)
	defer store.Close()

	service := NewBenchmarkService()
	bundle, err := service.ExportEvidenceBundle(context.Background(), BenchmarkEvidenceBundleRequest{
		StartPath:    repoRoot,
		SuiteID:      "go-benchmark-refresh-v1",
		SuitesDir:    filepath.Join("..", "..", "testdata", "eval", "benchmarks"),
		FixturesRoot: filepath.Join("..", "..", "testdata", "eval", "fixtures"),
	})
	if err != nil {
		t.Fatalf("ExportEvidenceBundle() error = %v", err)
	}
	fromService := RenderBenchmarkComparisonReport(BuildBenchmarkHumanSummary(bundle))

	loaded, err := store.LoadLatestBenchmarkEvidenceBundle(context.Background(), repoID, bundle.SuiteID, bundle.SuiteVersion)
	if err != nil {
		t.Fatalf("LoadLatestBenchmarkEvidenceBundle() error = %v", err)
	}
	fromStore := RenderBenchmarkComparisonReport(BuildBenchmarkHumanSummary(loaded))
	if fromStore != fromService {
		t.Fatalf("report should reuse persisted evidence bundle\nservice:\n%s\nstore:\n%s", fromService, fromStore)
	}
}

func TestBenchmarkRerunReproducibility(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	attempts := []BenchmarkAttemptResult{
		{Attempt: 1, Result: benchmarkEvidenceRunResult(repoRoot)},
		{Attempt: 2, Result: benchmarkEvidenceRunResult(repoRoot)},
	}
	summary := summarizeBenchmarkAttempts(attempts, "go-benchmark-refresh-v1", "v1")
	bundle := buildBenchmarkEvidenceBundle(repoRoot, repository.DefaultBenchmarkTokenEstimateContract(), summary, attempts, "go run ./cmd/optimusctx eval benchmark verify --suite go-benchmark-refresh-v1 --attempts 2")
	recomputed := buildBenchmarkEvidenceBundle(repoRoot, repository.DefaultBenchmarkTokenEstimateContract(), summary, attempts, "go run ./cmd/optimusctx eval benchmark verify --suite go-benchmark-refresh-v1 --attempts 2")

	reasons := compareBenchmarkEvidenceBundles(bundle, recomputed)
	if len(reasons) != 0 {
		t.Fatalf("compareBenchmarkEvidenceBundles() = %+v, want no drift", reasons)
	}
	if wording := verifyBenchmarkReportWording(RenderBenchmarkComparisonReport(BuildBenchmarkHumanSummary(recomputed))); len(wording) != 0 {
		t.Fatalf("verifyBenchmarkReportWording() = %+v", wording)
	}
}

func TestBenchmarkMethodologyFingerprint(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	attempts := []BenchmarkAttemptResult{
		{Attempt: 1, Result: benchmarkEvidenceRunResult(repoRoot)},
		{Attempt: 2, Result: benchmarkEvidenceRunResult(repoRoot)},
	}
	summary := summarizeBenchmarkAttempts(attempts, "go-benchmark-refresh-v1", "v1")
	bundle := buildBenchmarkEvidenceBundle(repoRoot, repository.DefaultBenchmarkTokenEstimateContract(), summary, attempts, "go run ./cmd/optimusctx eval benchmark verify --suite go-benchmark-refresh-v1 --attempts 2")

	driftedResult := benchmarkEvidenceRunResult(repoRoot)
	driftedResult.Arms[0].LaneResults[0].StopMarker = "refresh_ready_drifted"
	driftedAttempts := []BenchmarkAttemptResult{
		{Attempt: 1, Result: benchmarkEvidenceRunResult(repoRoot)},
		{Attempt: 2, Result: driftedResult},
	}
	driftedSummary := summarizeBenchmarkAttempts(driftedAttempts, "go-benchmark-refresh-v1", "v1")
	driftedBundle := buildBenchmarkEvidenceBundle(repoRoot, repository.DefaultBenchmarkTokenEstimateContract(), driftedSummary, driftedAttempts, "go run ./cmd/optimusctx eval benchmark verify --suite go-benchmark-refresh-v1 --attempts 2")

	reasons := compareBenchmarkEvidenceBundles(bundle, driftedBundle)
	if len(reasons) == 0 {
		t.Fatal("compareBenchmarkEvidenceBundles() should report methodology drift")
	}
	if !strings.Contains(strings.Join(reasons, " "), "methodology fingerprint") {
		t.Fatalf("reasons = %+v", reasons)
	}
}

func TestBuildBenchmarkEvidenceBundleFromPersistedRuns(t *testing.T) {
	t.Parallel()

	repoRoot, store, repoID := initBenchmarkEvidenceStore(t)
	defer store.Close()

	persisted, err := store.ListBenchmarkRuns(context.Background(), repoID, "go-benchmark-refresh-v1", "v1")
	if err != nil {
		t.Fatalf("ListBenchmarkRuns() error = %v", err)
	}
	bundle, err := buildBenchmarkEvidenceBundleFromPersistedRuns(repoRoot, persisted, repository.BenchmarkSuiteDefinition{
		ID:      "go-benchmark-refresh-v1",
		Version: "v1",
	}, "go run ./cmd/optimusctx eval benchmark verify --suite go-benchmark-refresh-v1 --attempts 2")
	if err != nil {
		t.Fatalf("buildBenchmarkEvidenceBundleFromPersistedRuns() error = %v", err)
	}
	if bundle.MethodologyFingerprint == "" {
		t.Fatal("MethodologyFingerprint should not be empty")
	}
	if len(bundle.Attempts) != 2 {
		t.Fatalf("len(bundle.Attempts) = %d, want 2", len(bundle.Attempts))
	}
}

func initBenchmarkEvidenceStore(t *testing.T) (string, *sqlite.Store, int64) {
	t.Helper()

	repoRoot := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, output)
	}

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
		for _, persisted := range sqlite.BenchmarkPersistedArmsFromResult(record.ID, attempt, benchmarkEvidenceRunResult(repoRoot)) {
			if _, _, err := store.SaveBenchmarkRun(context.Background(), persisted.Run, persisted.Samples); err != nil {
				t.Fatalf("SaveBenchmarkRun(attempt=%d) error = %v", attempt, err)
			}
		}
	}
	return repoRoot, store, record.ID
}

func benchmarkEvidenceRunResult(repositoryRoot string) repository.BenchmarkRunResult {
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
						ActionCount:        2,
						BytesRead:          128,
						ConsultedArtifacts: []string{"docs/notes.txt"},
					},
				}},
			},
			{
				Kind:       repository.BenchmarkArmKindOptimusCtx,
				Name:       "OptimusCtx",
				Workspace:  repositoryRoot,
				StartedAt:  time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, 3, 16, 20, 0, 2, 0, time.UTC),
				LaneResults: []repository.BenchmarkLaneRunResult{{
					Lane:          repository.BenchmarkLaneTaskCompletion,
					StartMarker:   "task_completion_started",
					SuccessMarker: "task_complete",
					StopMarker:    "task_complete",
					StartedAt:     time.Date(2026, 3, 16, 20, 0, 1, 0, time.UTC),
					FinishedAt:    time.Date(2026, 3, 16, 20, 0, 2, 0, time.UTC),
					Elapsed:       time.Second,
					Success:       true,
					EvidencePaths: []string{"artifacts/pack.json", "docs/notes.txt"},
					Effort: repository.BenchmarkLaneEffort{
						ActionCount:           2,
						TargetedLookupActions: 1,
						BytesRead:             512,
						ConsultedArtifacts:    []string{"artifacts/pack.json", "docs/notes.txt"},
					},
					Attribution: []repository.BenchmarkArtifactConsumption{
						{
							StepID:          "pack-export",
							StepName:        "Export pack",
							Lane:            repository.BenchmarkLaneTaskCompletion,
							Surface:         repository.BenchmarkTreatmentSurfaceCLI,
							Command:         repository.EvalCommandPackExport,
							ArtifactType:    repository.BenchmarkArtifactTypePackExport,
							ReportLabel:     repository.BenchmarkReportArtifactLabelPackExport,
							SourceKind:      repository.BenchmarkTokenEstimateSourcePackExportSection,
							ArtifactPath:    "artifacts/pack.json",
							EstimatedBytes:  512,
							EstimatedTokens: 128,
						},
					},
				}},
			},
		},
	}
}
