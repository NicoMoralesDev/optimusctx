package app

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestBenchmarkBaselineRules(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	fixturesRoot := filepath.Join(root, "fixtures")
	suitesDir := filepath.Join(root, "benchmarks")

	writeEvalFixtureFile(t, filepath.Join(fixturesRoot, "go-benchmark", "v1", "repository", "go.mod"), "module fixture/benchmark\n\ngo 1.23.0\n")
	writeBenchmarkSuiteFile(t, filepath.Join(suitesDir, "invalid.json"), repository.BenchmarkSuiteDefinition{
		SchemaVersion: repository.BenchmarkSuiteSchemaV2,
		ID:            "invalid-baseline-v2",
		Version:       "v2",
		Name:          "Invalid baseline",
		Boundary:      repository.DefaultBenchmarkBoundaryContract(),
		Fixture: repository.EvalFixtureRef{
			ID:          "go-benchmark",
			Version:     "v1",
			Path:        "go-benchmark/v1/repository",
			Materialize: repository.EvalFixtureModeCopyTree,
		},
		Task: repository.BenchmarkTaskDefinition{
			ID:         "target",
			Prompt:     "Find the rollout handler",
			TargetPath: "internal/http/handler/rollout.go",
			FinalArtifact: &repository.BenchmarkFinalArtifactContract{
				ID:     "target-locator",
				Name:   "Target locator",
				Kind:   repository.BenchmarkFinalArtifactKindTargetLocator,
				Path:   "artifacts/target.json",
				Format: repository.BenchmarkFinalArtifactFormatJSON,
				Normalization: repository.BenchmarkFinalArtifactNormalization{
					Mode:      repository.BenchmarkFinalArtifactNormalizationModeJSONFields,
					JSONPaths: []string{"path"},
				},
				Assertions: []repository.BenchmarkFinalArtifactAssertion{
					{Kind: repository.EvalAssertionKindJSONFieldPresent, Path: "path"},
				},
			},
		},
		CountedInputs: []repository.BenchmarkCountedInputDefinition{{
			ID:         "baseline-search",
			ArmKind:    repository.BenchmarkArmKindBaseline,
			Lane:       repository.BenchmarkLaneDiscovery,
			StepID:     "bad",
			Name:       "Baseline discovery paths",
			Kind:       repository.BenchmarkCountedInputKindPathList,
			SourceKind: repository.BenchmarkTokenEstimateSourcePathEstimate,
			Path:       "internal",
		}},
		Lanes: []repository.BenchmarkLaneDefinition{
			{
				Name: repository.BenchmarkLaneDiscovery,
				StopCondition: repository.BenchmarkStopCondition{
					Kind:   repository.BenchmarkStopConditionKindMarker,
					Marker: "target_identified",
				},
				Metrics: []repository.BenchmarkMetric{repository.BenchmarkMetricBroadSearchActions},
			},
		},
		Arms: []repository.BenchmarkArmDefinition{
			{
				Kind: repository.BenchmarkArmKindBaseline,
				Name: "Baseline",
				Steps: []repository.BenchmarkStep{{
					ID:   "bad",
					Name: "Uses MCP",
					Lane: repository.BenchmarkLaneDiscovery,
					Treatment: &repository.BenchmarkTreatmentAction{
						Surface: repository.BenchmarkTreatmentSurfaceMCP,
						Tool:    "optimusctx.repository_map",
					},
				}},
			},
			{
				Kind: repository.BenchmarkArmKindOptimusCtx,
				Name: "Treatment",
				Steps: []repository.BenchmarkStep{{
					ID:   "map",
					Name: "Use repository map",
					Lane: repository.BenchmarkLaneDiscovery,
					Treatment: &repository.BenchmarkTreatmentAction{
						Surface: repository.BenchmarkTreatmentSurfaceMCP,
						Tool:    "optimusctx.repository_map",
					},
				}},
			},
		},
	})

	_, err := NewBenchmarkRunner().LoadSuites(BenchmarkSuiteRequest{
		SuitesDir:    suitesDir,
		FixturesRoot: fixturesRoot,
	})
	if err == nil || !strings.Contains(err.Error(), "baseline arm must use only baseline actions") {
		t.Fatalf("LoadSuites() error = %v, want baseline rule violation", err)
	}
}

func TestBenchmarkSuitePersistence(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	fixturesRoot := filepath.Join(root, "fixtures")
	suitesDir := filepath.Join(root, "benchmarks")

	writeEvalFixtureFile(t, filepath.Join(fixturesRoot, "go-benchmark", "v1", "repository", "go.mod"), "module fixture/benchmark\n\ngo 1.23.0\n")
	writeBenchmarkSuiteFile(t, filepath.Join(suitesDir, "suite.json"), validBenchmarkSuite())

	runner := NewBenchmarkRunner()
	suite, err := runner.LoadSuite(BenchmarkSuiteRequest{
		SuiteID:      "go-benchmark-discovery-v1",
		SuitesDir:    suitesDir,
		FixturesRoot: fixturesRoot,
	})
	if err != nil {
		t.Fatalf("LoadSuite() error = %v", err)
	}
	if suite.Task.TargetSymbol != "LoadRolloutConfig" {
		t.Fatalf("TargetSymbol = %q, want LoadRolloutConfig", suite.Task.TargetSymbol)
	}

	suites, err := runner.LoadSuites(BenchmarkSuiteRequest{
		SuitesDir:    suitesDir,
		FixturesRoot: fixturesRoot,
	})
	if err != nil {
		t.Fatalf("LoadSuites() error = %v", err)
	}
	if got, want := len(suites), 1; got != want {
		t.Fatalf("len(LoadSuites()) = %d, want %d", got, want)
	}
}

func TestBenchmarkFixtureSelection(t *testing.T) {
	t.Parallel()

	runner := NewBenchmarkRunner()
	suite, err := runner.LoadSuite(BenchmarkSuiteRequest{
		SuiteID:      "go-benchmark-discovery-v1",
		SuitesDir:    filepath.Join("..", "..", "testdata", "eval", "benchmarks"),
		FixturesRoot: filepath.Join("..", "..", "testdata", "eval", "fixtures"),
	})
	if err != nil {
		t.Fatalf("LoadSuite() error = %v", err)
	}
	if suite.Fixture.ID != "go-benchmark" {
		t.Fatalf("Fixture.ID = %q, want go-benchmark", suite.Fixture.ID)
	}
	if suite.Arms[0].Kind != repository.BenchmarkArmKindBaseline || suite.Arms[1].Kind != repository.BenchmarkArmKindOptimusCtx {
		t.Fatalf("suite arms = %+v, want baseline+optimusctx", suite.Arms)
	}
}

func TestBenchmarkCorpusValidation(t *testing.T) {
	t.Parallel()

	suites, err := NewBenchmarkRunner().LoadSuites(BenchmarkSuiteRequest{
		SuitesDir:    filepath.Join("..", "..", "testdata", "eval", "benchmarks"),
		FixturesRoot: filepath.Join("..", "..", "testdata", "eval", "fixtures"),
	})
	if err != nil {
		t.Fatalf("LoadSuites() error = %v", err)
	}
	if got, want := len(suites), 2; got != want {
		t.Fatalf("len(suites) = %d, want %d", got, want)
	}
	for _, suite := range suites {
		if suite.Task.Prompt == "" {
			t.Fatalf("suite %q missing prompt", suite.ID)
		}
		if len(suite.Lanes) < 2 {
			t.Fatalf("suite %q should define at least two lanes", suite.ID)
		}
	}
}

func TestBenchmarkSuiteV2FrozenCorpus(t *testing.T) {
	t.Parallel()

	suites, err := NewBenchmarkRunner().LoadSuites(BenchmarkSuiteRequest{
		SuitesDir:    filepath.Join("..", "..", "testdata", "eval", "benchmarks"),
		FixturesRoot: filepath.Join("..", "..", "testdata", "eval", "fixtures"),
	})
	if err != nil {
		t.Fatalf("LoadSuites() error = %v", err)
	}

	byID := make(map[string]repository.BenchmarkSuiteDefinition, len(suites))
	for _, suite := range suites {
		byID[suite.ID] = suite
	}

	discovery, ok := byID["go-benchmark-discovery-v1"]
	if !ok {
		t.Fatal("missing go-benchmark-discovery-v1")
	}
	if discovery.SchemaVersion != repository.BenchmarkSuiteSchemaV2 {
		t.Fatalf("discovery SchemaVersion = %q, want %q", discovery.SchemaVersion, repository.BenchmarkSuiteSchemaV2)
	}
	if discovery.Version != "v1" {
		t.Fatalf("discovery Version = %q, want v1", discovery.Version)
	}
	if len(discovery.CountedInputs) < 6 {
		t.Fatalf("discovery counted inputs = %d, want explicit discovery and context projections", len(discovery.CountedInputs))
	}
	if discovery.Lanes[0].FinalArtifact == nil || discovery.Lanes[1].FinalArtifact == nil {
		t.Fatalf("discovery lane final artifacts = %+v", discovery.Lanes)
	}

	refresh, ok := byID["go-benchmark-refresh-v1"]
	if !ok {
		t.Fatal("missing go-benchmark-refresh-v1")
	}
	if refresh.SchemaVersion != repository.BenchmarkSuiteSchemaV2 {
		t.Fatalf("refresh SchemaVersion = %q, want %q", refresh.SchemaVersion, repository.BenchmarkSuiteSchemaV2)
	}
	if refresh.Version != "v1" {
		t.Fatalf("refresh Version = %q, want v1", refresh.Version)
	}
	if len(refresh.CountedInputs) < 5 {
		t.Fatalf("refresh counted inputs = %d, want explicit readiness and task projections", len(refresh.CountedInputs))
	}
	if refresh.Lanes[0].FinalArtifact == nil || refresh.Lanes[1].FinalArtifact == nil {
		t.Fatalf("refresh lane final artifacts = %+v", refresh.Lanes)
	}
}

func TestBenchmarkMigrationPreservesStableSuiteSelection(t *testing.T) {
	t.Parallel()

	for _, suiteID := range []string{"go-benchmark-discovery-v1", "go-benchmark-refresh-v1"} {
		suite, err := NewBenchmarkRunner().LoadSuite(BenchmarkSuiteRequest{
			SuiteID:      suiteID,
			SuitesDir:    filepath.Join("..", "..", "testdata", "eval", "benchmarks"),
			FixturesRoot: filepath.Join("..", "..", "testdata", "eval", "fixtures"),
		})
		if err != nil {
			t.Fatalf("LoadSuite(%q) error = %v", suiteID, err)
		}
		if suite.SchemaVersion != repository.BenchmarkSuiteSchemaV2 {
			t.Fatalf("%s schema = %q, want %q", suiteID, suite.SchemaVersion, repository.BenchmarkSuiteSchemaV2)
		}
		if suite.Version != "v1" {
			t.Fatalf("%s version = %q, want v1", suiteID, suite.Version)
		}
		if err := suite.Boundary.Validate(); err != nil {
			t.Fatalf("%s boundary validate error = %v", suiteID, err)
		}
	}
}

func TestBenchmarkLaneDefinitions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	fixturesRoot := filepath.Join(root, "fixtures")
	suitesDir := filepath.Join(root, "benchmarks")
	writeEvalFixtureFile(t, filepath.Join(fixturesRoot, "go-benchmark", "v1", "repository", "go.mod"), "module fixture/benchmark\n\ngo 1.23.0\n")
	writeBenchmarkSuiteFile(t, filepath.Join(suitesDir, "suite.json"), validBenchmarkSuite())

	suite, err := NewBenchmarkRunner().LoadSuite(BenchmarkSuiteRequest{
		SuiteID:      "go-benchmark-discovery-v1",
		SuitesDir:    suitesDir,
		FixturesRoot: fixturesRoot,
	})
	if err != nil {
		t.Fatalf("LoadSuite() error = %v", err)
	}

	if suite.Lanes[0].StartMarker == "" || suite.Lanes[0].SuccessMarker == "" {
		t.Fatalf("lane markers missing: %+v", suite.Lanes[0])
	}
	if suite.Lanes[0].StopCondition.Marker != "target_identified" {
		t.Fatalf("discovery stop marker = %q", suite.Lanes[0].StopCondition.Marker)
	}
	if suite.Lanes[1].StopCondition.Marker != "context_ready" {
		t.Fatalf("context stop marker = %q", suite.Lanes[1].StopCondition.Marker)
	}
}

func TestBenchmarkDiscoveryTiming(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	runner := NewBenchmarkRunner()
	runner.Now = func() time.Time {
		current := now
		now = now.Add(250 * time.Millisecond)
		return current
	}
	runner.MkdirTemp = func(string, string) (string, error) {
		return t.TempDir(), nil
	}
	runner.CopyTree = func(src string, dst string) error {
		return copyEvalTree(src, dst)
	}
	runner.GitInit = func(context.Context, string) error { return nil }
	runner.RunCommand = func(context.Context, BenchmarkCommandInvocation) (BenchmarkCommandExecutionResult, error) {
		return BenchmarkCommandExecutionResult{ExitCode: 0}, nil
	}
	runner.RunTool = func(_ context.Context, invocation BenchmarkToolInvocation) (BenchmarkToolExecutionResult, error) {
		switch invocation.Name {
		case "optimusctx.repository_map":
			return BenchmarkToolExecutionResult{Payload: repository.RepositoryMap{
				RepositoryRoot: invocation.WorkingDir,
				Directories: []repository.RepositoryMapDirectory{{
					Path: "internal/http/handler",
					Files: []repository.RepositoryMapFile{{
						Path: "internal/http/handler/rollout.go",
					}},
				}},
			}}, nil
		case "optimusctx.symbol_lookup":
			return BenchmarkToolExecutionResult{Payload: repository.SymbolLookupResult{
				Matches: []repository.SymbolLookupMatch{{
					Path: "internal/http/handler/rollout.go",
					Name: "LoadRolloutConfig",
				}},
			}}, nil
		case "optimusctx.targeted_context":
			return BenchmarkToolExecutionResult{Payload: repository.TargetedContextResult{
				Path:   "internal/http/handler/rollout.go",
				Source: []string{"func LoadRolloutConfig() {}", "return cfg"},
			}}, nil
		default:
			t.Fatalf("unexpected tool %q", invocation.Name)
			return BenchmarkToolExecutionResult{}, nil
		}
	}

	result, err := runner.Run(context.Background(), BenchmarkRunRequest{
		SuiteID:      "go-benchmark-discovery-v1",
		SuitesDir:    filepath.Join("..", "..", "testdata", "eval", "benchmarks"),
		FixturesRoot: filepath.Join("..", "..", "testdata", "eval", "fixtures"),
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(result.Arms) != 2 {
		t.Fatalf("len(result.Arms) = %d, want 2", len(result.Arms))
	}

	discovery := result.Arms[0].LaneResults[0]
	if !discovery.Success {
		t.Fatalf("discovery lane = %+v, want success", discovery)
	}
	if discovery.Elapsed <= 0 {
		t.Fatalf("discovery elapsed = %s, want > 0", discovery.Elapsed)
	}
	if discovery.Effort.ActionCount == 0 {
		t.Fatalf("discovery effort = %+v, want actions", discovery.Effort)
	}

	contextLane := result.Arms[1].LaneResults[1]
	if !contextLane.Success {
		t.Fatalf("context lane = %+v, want success", contextLane)
	}
	if contextLane.Effort.BytesRead == 0 {
		t.Fatalf("context bytes = %d, want > 0", contextLane.Effort.BytesRead)
	}
	if !strings.Contains(strings.Join(contextLane.Effort.ConsultedArtifacts, ","), "internal/http/handler/rollout.go") {
		t.Fatalf("context consulted artifacts = %+v", contextLane.Effort.ConsultedArtifacts)
	}
}

func TestBenchmarkRefreshAfterChangeLane(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	runner := NewBenchmarkRunner()
	runner.Now = func() time.Time {
		current := now
		now = now.Add(250 * time.Millisecond)
		return current
	}
	runner.MkdirTemp = func(string, string) (string, error) {
		return t.TempDir(), nil
	}
	runner.CopyTree = func(src string, dst string) error {
		return copyEvalTree(src, dst)
	}
	runner.GitInit = func(context.Context, string) error { return nil }
	runner.RunCommand = func(_ context.Context, invocation BenchmarkCommandInvocation) (BenchmarkCommandExecutionResult, error) {
		if len(invocation.Args) > 0 && invocation.Args[0] == "refresh" {
			return BenchmarkCommandExecutionResult{ExitCode: 0}, nil
		}
		return BenchmarkCommandExecutionResult{ExitCode: 0}, nil
	}
	runner.RunTool = benchmarkMutationToolExecutor(t)

	result, err := runner.Run(context.Background(), BenchmarkRunRequest{
		SuitePath:     writeBenchmarkMutationSuite(t),
		FixturesRoot:  filepath.Join("..", "..", "testdata", "eval", "fixtures"),
		WorkspaceRoot: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	refreshLane := result.Arms[0].LaneResults[0]
	if !refreshLane.Success {
		t.Fatalf("refresh lane = %+v, want success", refreshLane)
	}
	if !refreshLane.SetupAppliedAt.Before(refreshLane.StartedAt) {
		t.Fatalf("setup applied at = %s, started at = %s, want setup before timing start", refreshLane.SetupAppliedAt, refreshLane.StartedAt)
	}
	if !strings.Contains(strings.Join(refreshLane.EvidencePaths, ","), "docs/notes.txt") {
		t.Fatalf("refresh evidence = %+v", refreshLane.EvidencePaths)
	}
}

func TestBenchmarkAgentInputProjection(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	runner := NewBenchmarkRunner()
	runner.Now = func() time.Time {
		current := now
		now = now.Add(250 * time.Millisecond)
		return current
	}
	runner.MkdirTemp = func(string, string) (string, error) { return t.TempDir(), nil }
	runner.CopyTree = func(src string, dst string) error { return copyEvalTree(src, dst) }
	runner.GitInit = func(context.Context, string) error { return nil }
	runner.RunCommand = func(context.Context, BenchmarkCommandInvocation) (BenchmarkCommandExecutionResult, error) {
		return BenchmarkCommandExecutionResult{ExitCode: 0}, nil
	}
	runner.RunTool = func(_ context.Context, invocation BenchmarkToolInvocation) (BenchmarkToolExecutionResult, error) {
		switch invocation.Name {
		case "optimusctx.repository_map":
			return BenchmarkToolExecutionResult{Payload: repository.RepositoryMap{
				Directories: []repository.RepositoryMapDirectory{{
					Path: "internal/http/handler",
					Files: []repository.RepositoryMapFile{{
						Path: "internal/http/handler/rollout.go",
					}},
				}},
			}}, nil
		case "optimusctx.symbol_lookup":
			return BenchmarkToolExecutionResult{Payload: repository.SymbolLookupResult{
				Matches: []repository.SymbolLookupMatch{{
					Path: "internal/http/handler/rollout.go",
					Name: "LoadRolloutConfig",
				}},
			}}, nil
		case "optimusctx.targeted_context":
			return BenchmarkToolExecutionResult{Payload: repository.TargetedContextResult{
				Path:   "internal/http/handler/rollout.go",
				Source: []string{"func LoadRolloutConfig() {}", "return cfg"},
			}}, nil
		default:
			t.Fatalf("unexpected tool %q", invocation.Name)
			return BenchmarkToolExecutionResult{}, nil
		}
	}
	suitePath := filepath.Join(t.TempDir(), "suite.json")
	writeBenchmarkSuiteFile(t, suitePath, validBenchmarkSuite())

	result, err := runner.Run(context.Background(), BenchmarkRunRequest{
		SuitePath:    suitePath,
		FixturesRoot: filepath.Join("..", "..", "testdata", "eval", "fixtures"),
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	baselineDiscovery := result.Arms[0].LaneResults[0]
	if len(baselineDiscovery.Attribution) != 2 {
		t.Fatalf("baseline discovery attribution = %+v, want counted input plus final-artifact verification", baselineDiscovery.Attribution)
	}
	if baselineDiscovery.Attribution[0].Boundary != repository.BenchmarkEvidenceBoundaryAgentInput {
		t.Fatalf("baseline discovery boundary = %q", baselineDiscovery.Attribution[0].Boundary)
	}
	if baselineDiscovery.Attribution[0].SourceKind != repository.BenchmarkTokenEstimateSourcePathEstimate {
		t.Fatalf("baseline discovery source kind = %q", baselineDiscovery.Attribution[0].SourceKind)
	}
	if baselineDiscovery.Attribution[1].Boundary != repository.BenchmarkEvidenceBoundaryFinalArtifactVerified {
		t.Fatalf("baseline discovery verification boundary = %q", baselineDiscovery.Attribution[1].Boundary)
	}

	baselineContext := result.Arms[0].LaneResults[1]
	if len(baselineContext.Attribution) != 2 {
		t.Fatalf("baseline context attribution = %+v, want counted input plus final-artifact verification", baselineContext.Attribution)
	}
	if baselineContext.Attribution[0].Boundary != repository.BenchmarkEvidenceBoundaryAgentInput {
		t.Fatalf("baseline context boundary = %q", baselineContext.Attribution[0].Boundary)
	}
	if baselineContext.Attribution[0].SourceKind != repository.BenchmarkTokenEstimateSourceBoundedFileContent {
		t.Fatalf("baseline source kind = %q", baselineContext.Attribution[0].SourceKind)
	}
	if baselineContext.Attribution[0].EstimatedBytes == 0 || baselineContext.Attribution[0].EstimatedTokens == 0 {
		t.Fatalf("baseline attribution = %+v, want estimated bytes/tokens", baselineContext.Attribution[0])
	}
	if baselineContext.Attribution[1].Boundary != repository.BenchmarkEvidenceBoundaryFinalArtifactVerified {
		t.Fatalf("baseline context verification boundary = %q", baselineContext.Attribution[1].Boundary)
	}

	treatmentDiscovery := result.Arms[1].LaneResults[0]
	if len(treatmentDiscovery.Attribution) != 4 {
		t.Fatalf("treatment discovery attribution = %+v, want raw repository-map + raw lookup + projected lookup + final-artifact verification", treatmentDiscovery.Attribution)
	}
	if treatmentDiscovery.Attribution[0].Boundary != repository.BenchmarkEvidenceBoundarySystemProvenance {
		t.Fatalf("repository map boundary = %q", treatmentDiscovery.Attribution[0].Boundary)
	}
	if treatmentDiscovery.Attribution[1].Boundary != repository.BenchmarkEvidenceBoundarySystemProvenance {
		t.Fatalf("symbol lookup provenance boundary = %q", treatmentDiscovery.Attribution[1].Boundary)
	}
	if treatmentDiscovery.Attribution[2].Boundary != repository.BenchmarkEvidenceBoundaryAgentInput {
		t.Fatalf("symbol lookup boundary = %q", treatmentDiscovery.Attribution[2].Boundary)
	}
	if treatmentDiscovery.Attribution[2].ArtifactType != repository.BenchmarkArtifactTypeExactLookup {
		t.Fatalf("lookup artifact type = %q", treatmentDiscovery.Attribution[2].ArtifactType)
	}
	if treatmentDiscovery.Attribution[3].Boundary != repository.BenchmarkEvidenceBoundaryFinalArtifactVerified {
		t.Fatalf("discovery verification boundary = %q", treatmentDiscovery.Attribution[3].Boundary)
	}

	treatmentContext := result.Arms[1].LaneResults[1]
	if len(treatmentContext.Attribution) != 3 {
		t.Fatalf("treatment context attribution = %+v, want provenance plus counted projection plus final-artifact verification", treatmentContext.Attribution)
	}
	if treatmentContext.Attribution[0].Boundary != repository.BenchmarkEvidenceBoundarySystemProvenance {
		t.Fatalf("targeted_context provenance boundary = %q", treatmentContext.Attribution[0].Boundary)
	}
	got := treatmentContext.Attribution[1]
	if got.Boundary != repository.BenchmarkEvidenceBoundaryAgentInput {
		t.Fatalf("counted context boundary = %q", got.Boundary)
	}
	if got.Tool != "optimusctx.targeted_context" {
		t.Fatalf("treatment tool = %q", got.Tool)
	}
	if got.ArtifactType != repository.BenchmarkArtifactTypeL2Context {
		t.Fatalf("artifact type = %q", got.ArtifactType)
	}
	if got.ReportLabel != repository.BenchmarkReportArtifactLabelL2Context {
		t.Fatalf("report label = %q", got.ReportLabel)
	}
	if got.SourceKind != repository.BenchmarkTokenEstimateSourceDirectPayload {
		t.Fatalf("source kind = %q", got.SourceKind)
	}
	if got.EstimatedBytes != int64(len("func LoadRolloutConfig() {}\nreturn cfg")) {
		t.Fatalf("estimated bytes = %d", got.EstimatedBytes)
	}
	if got.EstimatedTokens != repository.EstimateBenchmarkTokensFromBytes(got.EstimatedBytes) {
		t.Fatalf("estimated tokens = %d", got.EstimatedTokens)
	}
	if treatmentContext.Attribution[2].Boundary != repository.BenchmarkEvidenceBoundaryFinalArtifactVerified {
		t.Fatalf("context verification boundary = %q", treatmentContext.Attribution[2].Boundary)
	}
}

func TestBenchmarkAttributionBoundary(t *testing.T) {
	t.Parallel()

	clock := newDeterministicBenchmarkClock()
	runner := NewBenchmarkRunner()
	runner.Now = clock
	runner.MkdirTemp = func(string, string) (string, error) { return t.TempDir(), nil }
	runner.CopyTree = func(src string, dst string) error { return copyEvalTree(src, dst) }
	runner.GitInit = func(context.Context, string) error { return nil }
	runner.RunCommand = func(_ context.Context, invocation BenchmarkCommandInvocation) (BenchmarkCommandExecutionResult, error) {
		return BenchmarkCommandExecutionResult{ExitCode: 0, Stdout: "refresh completed\n"}, nil
	}
	runner.RunTool = benchmarkMutationToolExecutor(t)

	result, err := runner.Run(context.Background(), BenchmarkRunRequest{
		SuitePath:     writeBenchmarkMutationSuite(t),
		FixturesRoot:  filepath.Join("..", "..", "testdata", "eval", "fixtures"),
		WorkspaceRoot: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	refreshLane := result.Arms[1].LaneResults[0]
	if len(refreshLane.Attribution) != 5 {
		t.Fatalf("refresh lane attribution = %+v, want provenance plus final-artifact verification plus counted health projection", refreshLane.Attribution)
	}
	var foundRefreshOutput bool
	var foundRefreshMarker bool
	var foundReadinessVerification bool
	var foundHealthProvenance bool
	var foundHealthProjection bool
	for _, record := range refreshLane.Attribution {
		switch {
		case record.StepID == "refresh" && record.Command == repository.EvalCommandRefresh && record.Boundary == repository.BenchmarkEvidenceBoundarySystemProvenance && record.EstimatedTokens > 0:
			foundRefreshOutput = true
		case record.StepID == "refresh" && record.Command == repository.EvalCommandRefresh && record.Boundary == repository.BenchmarkEvidenceBoundarySystemProvenance && record.EstimatedTokens == 0:
			foundRefreshMarker = true
		case record.StepID == "refresh-readiness" && record.Boundary == repository.BenchmarkEvidenceBoundaryFinalArtifactVerified:
			foundReadinessVerification = true
		case record.StepID == "health" && record.Tool == "optimusctx.health" && record.Boundary == repository.BenchmarkEvidenceBoundarySystemProvenance:
			foundHealthProvenance = true
		case record.StepID == "health" && record.Tool == "optimusctx.health" && record.Boundary == repository.BenchmarkEvidenceBoundaryAgentInput:
			foundHealthProjection = true
		}
	}
	if !foundRefreshOutput || !foundRefreshMarker || !foundReadinessVerification || !foundHealthProvenance || !foundHealthProjection {
		t.Fatalf("refresh lane attribution = %+v, want refresh output + refresh marker + final-artifact verification + raw health + projected health summary", refreshLane.Attribution)
	}

	taskLane := result.Arms[1].LaneResults[1]
	if len(taskLane.Attribution) != 3 {
		t.Fatalf("task lane attribution = %+v, want provenance plus counted context projection plus final-artifact verification", taskLane.Attribution)
	}
	if taskLane.Attribution[0].Boundary != repository.BenchmarkEvidenceBoundarySystemProvenance {
		t.Fatalf("task provenance boundary = %q", taskLane.Attribution[0].Boundary)
	}
	if taskLane.Attribution[1].StepID != "context" {
		t.Fatalf("step id = %q, want context", taskLane.Attribution[1].StepID)
	}
	if taskLane.Attribution[1].Boundary != repository.BenchmarkEvidenceBoundaryAgentInput {
		t.Fatalf("task counted boundary = %q", taskLane.Attribution[1].Boundary)
	}
	if taskLane.Attribution[1].ArtifactType != repository.BenchmarkArtifactTypeL2Context {
		t.Fatalf("artifact type = %q, want l2_context", taskLane.Attribution[1].ArtifactType)
	}
	if taskLane.Attribution[1].SourceKind != repository.BenchmarkTokenEstimateSourceDirectPayload {
		t.Fatalf("source kind = %q", taskLane.Attribution[1].SourceKind)
	}
	if taskLane.Attribution[1].EstimatedTokens == 0 {
		t.Fatalf("context attribution = %+v, want non-zero estimated tokens", taskLane.Attribution[1])
	}
	if taskLane.Attribution[2].Boundary != repository.BenchmarkEvidenceBoundaryFinalArtifactVerified {
		t.Fatalf("task verification boundary = %q", taskLane.Attribution[2].Boundary)
	}
}

func TestBenchmarkLaneCompletionRequiresFinalArtifact(t *testing.T) {
	t.Parallel()

	runner := NewBenchmarkRunner()
	runner.MkdirTemp = func(string, string) (string, error) {
		return t.TempDir(), nil
	}
	runner.CopyTree = func(src string, dst string) error {
		return copyEvalTree(src, dst)
	}
	runner.GitInit = func(context.Context, string) error { return nil }
	runner.RunCommand = func(_ context.Context, invocation BenchmarkCommandInvocation) (BenchmarkCommandExecutionResult, error) {
		if reflect.DeepEqual(invocation.Args, []string{"refresh"}) {
			return BenchmarkCommandExecutionResult{ExitCode: 0}, nil
		}
		return BenchmarkCommandExecutionResult{}, nil
	}
	runner.RunTool = benchmarkMutationToolExecutor(t)

	result, err := runner.Run(context.Background(), BenchmarkRunRequest{
		SuitePath:     writeBenchmarkMutationSuite(t),
		FixturesRoot:  filepath.Join("..", "..", "testdata", "eval", "fixtures"),
		WorkspaceRoot: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	taskLane := result.Arms[1].LaneResults[1]
	if !taskLane.Success {
		t.Fatalf("task lane = %+v, want success", taskLane)
	}
	if taskLane.FinalArtifact == nil || !taskLane.FinalArtifact.Passed {
		t.Fatalf("task final artifact = %+v, want passing verification", taskLane.FinalArtifact)
	}
	if !strings.Contains(strings.Join(taskLane.EvidencePaths, ","), "docs/notes.txt") {
		t.Fatalf("task evidence = %+v", taskLane.EvidencePaths)
	}
	if !strings.Contains(strings.Join(taskLane.EvidencePaths, ","), "artifacts/updated-notes.txt") {
		t.Fatalf("task evidence = %+v, want normalized final artifact path", taskLane.EvidencePaths)
	}
}

func TestBenchmarkFinalArtifactValidation(t *testing.T) {
	t.Parallel()

	runner := NewBenchmarkRunner()
	runner.MkdirTemp = func(string, string) (string, error) { return t.TempDir(), nil }
	runner.CopyTree = func(src string, dst string) error { return copyEvalTree(src, dst) }
	runner.GitInit = func(context.Context, string) error { return nil }
	runner.RunCommand = func(_ context.Context, invocation BenchmarkCommandInvocation) (BenchmarkCommandExecutionResult, error) {
		return BenchmarkCommandExecutionResult{ExitCode: 0}, nil
	}
	runner.RunTool = func(_ context.Context, invocation BenchmarkToolInvocation) (BenchmarkToolExecutionResult, error) {
		switch invocation.Name {
		case "optimusctx.health":
			return BenchmarkToolExecutionResult{Payload: map[string]any{
				"freshness": "fresh",
			}}, nil
		case "optimusctx.targeted_context":
			return BenchmarkToolExecutionResult{Payload: repository.TargetedContextResult{
				Path:   "docs/notes.txt",
				Source: []string{"wrong content"},
			}}, nil
		default:
			t.Fatalf("unexpected tool %q", invocation.Name)
			return BenchmarkToolExecutionResult{}, nil
		}
	}

	result, err := runner.Run(context.Background(), BenchmarkRunRequest{
		SuitePath:     writeBenchmarkMutationSuite(t),
		FixturesRoot:  filepath.Join("..", "..", "testdata", "eval", "fixtures"),
		WorkspaceRoot: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	taskLane := result.Arms[1].LaneResults[1]
	if taskLane.Success {
		t.Fatalf("task lane = %+v, want failed final-artifact validation", taskLane)
	}
	if taskLane.FinalArtifact == nil || taskLane.FinalArtifact.Passed {
		t.Fatalf("task final artifact = %+v, want failure", taskLane.FinalArtifact)
	}
	if !strings.Contains(taskLane.FinalArtifact.FailureReason, "does not contain") {
		t.Fatalf("failure reason = %q", taskLane.FinalArtifact.FailureReason)
	}
}

func TestBenchmarkTextFromLaneStepsCombinesObservations(t *testing.T) {
	t.Parallel()

	text, ok := benchmarkTextFromLaneSteps(map[string]benchmarkStepObservation{
		"baseline-read-loader": {
			fileSliceContent: "func LoadConfig() string {\n\treturn value\n}",
		},
		"baseline-read-target": {
			fileSliceContent: "func LoadRolloutConfig() string {\n\treturn LoadConfig()\n}",
		},
	})
	if !ok {
		t.Fatal("benchmarkTextFromLaneSteps() should report combined text")
	}
	if !strings.Contains(text, "LoadConfig") || !strings.Contains(text, "LoadRolloutConfig") {
		t.Fatalf("combined text = %q, want both lane observations", text)
	}
}

func TestBenchmarkReadinessSummarySupportsCapitalizedHealthFields(t *testing.T) {
	t.Parallel()

	state := &benchmarkArmState{
		lanes: map[repository.BenchmarkLane]*benchmarkLaneState{
			repository.BenchmarkLaneRefreshReady: {
				definition: repository.BenchmarkLaneDefinition{
					FinalArtifact: &repository.BenchmarkFinalArtifactContract{
						ID:     "refresh-readiness",
						Name:   "Refresh readiness summary",
						Kind:   repository.BenchmarkFinalArtifactKindReadinessSummary,
						Path:   "artifacts/readiness.json",
						Format: repository.BenchmarkFinalArtifactFormatJSON,
					},
				},
				steps: map[string]benchmarkStepObservation{
					"opti-health": {
						payload: map[string]any{
							"Refresh": map[string]any{
								"Freshness":         "fresh",
								"CurrentGeneration": 7,
							},
						},
					},
				},
			},
		},
	}

	content, ok, err := state.renderLaneFinalArtifact(repository.BenchmarkLaneRefreshReady, *state.lanes[repository.BenchmarkLaneRefreshReady].definition.FinalArtifact)
	if err != nil {
		t.Fatalf("renderLaneFinalArtifact() error = %v", err)
	}
	if !ok {
		t.Fatal("renderLaneFinalArtifact() should return readiness content")
	}
	summary, ok := content.(map[string]any)
	if !ok {
		t.Fatalf("content type = %T, want map[string]any", content)
	}
	if summary["freshness"] != "fresh" {
		t.Fatalf("freshness = %#v, want fresh", summary["freshness"])
	}
	if got, ok := summary["generation"].(float64); !ok || got != 7 {
		t.Fatalf("generation = %#v, want 7", summary["generation"])
	}
	if summary["targetReady"] != true {
		t.Fatalf("targetReady = %#v, want true", summary["targetReady"])
	}
}

func TestBenchmarkRepeatedRuns(t *testing.T) {
	t.Parallel()

	repoRoot := initRepo(t)
	service := NewBenchmarkService()
	service.Runner = NewBenchmarkRunner()
	service.Runner.Now = newDeterministicBenchmarkClock()
	service.Runner.RunCommand = func(_ context.Context, invocation BenchmarkCommandInvocation) (BenchmarkCommandExecutionResult, error) {
		if reflect.DeepEqual(invocation.Args, []string{"refresh"}) {
			return BenchmarkCommandExecutionResult{ExitCode: 0}, nil
		}
		return BenchmarkCommandExecutionResult{ExitCode: 0}, nil
	}
	service.Runner.RunTool = benchmarkMutationToolExecutor(t)

	result, err := service.RunRepeated(context.Background(), BenchmarkRepeatedRunRequest{
		StartPath:     repoRoot,
		SuitePath:     writeBenchmarkMutationSuite(t),
		FixturesRoot:  filepath.Join("..", "..", "testdata", "eval", "fixtures"),
		WorkspaceRoot: t.TempDir(),
		Attempts:      2,
	})
	if err != nil {
		t.Fatalf("RunRepeated() error = %v", err)
	}
	if got, want := len(result.Attempts), 2; got != want {
		t.Fatalf("len(result.Attempts) = %d, want %d", got, want)
	}
	for index, attempt := range result.Attempts {
		if got, want := attempt.Attempt, index+1; got != want {
			t.Fatalf("attempt number = %d, want %d", got, want)
		}
		if len(attempt.Result.Arms) != 2 {
			t.Fatalf("attempt %d arms = %d, want 2", attempt.Attempt, len(attempt.Result.Arms))
		}
	}
	if !result.Summary.Verification.Passed {
		t.Fatalf("summary verification = %+v, want passed", result.Summary.Verification)
	}
	if result.Summary.AttemptCount != 2 {
		t.Fatalf("summary attempts = %d, want 2", result.Summary.AttemptCount)
	}
	if !strings.Contains(result.Summary.RerunCommand, "optimusctx eval benchmark export --suite-file") {
		t.Fatalf("rerun command = %q", result.Summary.RerunCommand)
	}
}

func validBenchmarkSuite() repository.BenchmarkSuiteDefinition {
	return repository.BenchmarkSuiteDefinition{
		SchemaVersion: repository.BenchmarkSuiteSchemaV2,
		ID:            "go-benchmark-discovery-v1",
		Version:       "v1",
		Name:          "Go benchmark discovery and context assembly",
		Boundary:      repository.DefaultBenchmarkBoundaryContract(),
		Fixture: repository.EvalFixtureRef{
			ID:          "go-benchmark",
			Version:     "v1",
			Path:        "go-benchmark/v1/repository",
			Materialize: repository.EvalFixtureModeCopyTree,
		},
		Task: repository.BenchmarkTaskDefinition{
			ID:           "handler-owner",
			Prompt:       "Find the rollout handler owner and assemble the exact surrounding context.",
			TargetPath:   "internal/http/handler/rollout.go",
			TargetSymbol: "LoadRolloutConfig",
			ContextPaths: []string{"internal/http/handler/rollout.go", "internal/config/loader.go"},
		},
		CountedInputs: []repository.BenchmarkCountedInputDefinition{
			{
				ID:         "baseline-discovery-paths",
				ArmKind:    repository.BenchmarkArmKindBaseline,
				Lane:       repository.BenchmarkLaneDiscovery,
				StepID:     "search",
				Name:       "Baseline discovery search results",
				Kind:       repository.BenchmarkCountedInputKindPathList,
				SourceKind: repository.BenchmarkTokenEstimateSourcePathEstimate,
				Path:       "internal",
			},
			{
				ID:         "baseline-context-slice",
				ArmKind:    repository.BenchmarkArmKindBaseline,
				Lane:       repository.BenchmarkLaneContextAssembly,
				StepID:     "read",
				Name:       "Baseline rollout handler slice",
				Kind:       repository.BenchmarkCountedInputKindFileSlice,
				SourceKind: repository.BenchmarkTokenEstimateSourceBoundedFileContent,
				Path:       "internal/http/handler/rollout.go",
				StartLine:  1,
				EndLine:    80,
			},
			{
				ID:           "treatment-symbol-path",
				ArmKind:      repository.BenchmarkArmKindOptimusCtx,
				Lane:         repository.BenchmarkLaneDiscovery,
				StepID:       "lookup",
				Name:         "Projected symbol lookup match",
				Kind:         repository.BenchmarkCountedInputKindJSONFieldProjection,
				SourceKind:   repository.BenchmarkTokenEstimateSourceDirectPayload,
				ArtifactType: repository.BenchmarkArtifactTypeExactLookup,
				ReportLabel:  repository.BenchmarkReportArtifactLabelExactLookup,
				JSONPath:     "Matches.0.Path",
			},
			{
				ID:           "treatment-context",
				ArmKind:      repository.BenchmarkArmKindOptimusCtx,
				Lane:         repository.BenchmarkLaneContextAssembly,
				StepID:       "context",
				Name:         "Treatment targeted context",
				Kind:         repository.BenchmarkCountedInputKindTextOutput,
				SourceKind:   repository.BenchmarkTokenEstimateSourceDirectPayload,
				ArtifactType: repository.BenchmarkArtifactTypeL2Context,
				ReportLabel:  repository.BenchmarkReportArtifactLabelL2Context,
				Path:         "artifacts/context.txt",
			},
		},
		Lanes: []repository.BenchmarkLaneDefinition{
			{
				Name:          repository.BenchmarkLaneDiscovery,
				StartMarker:   "discovery_started",
				SuccessMarker: "target_identified",
				StopCondition: repository.BenchmarkStopCondition{
					Kind:   repository.BenchmarkStopConditionKindMarker,
					Marker: "target_identified",
				},
				FinalArtifact: &repository.BenchmarkFinalArtifactContract{
					ID:     "target-locator",
					Name:   "Target locator",
					Kind:   repository.BenchmarkFinalArtifactKindTargetLocator,
					Path:   "artifacts/target.json",
					Format: repository.BenchmarkFinalArtifactFormatJSON,
					Normalization: repository.BenchmarkFinalArtifactNormalization{
						Mode:      repository.BenchmarkFinalArtifactNormalizationModeJSONFields,
						JSONPaths: []string{"path", "symbol"},
					},
					Assertions: []repository.BenchmarkFinalArtifactAssertion{
						{Kind: repository.EvalAssertionKindJSONFieldPresent, Path: "path"},
					},
				},
				Metrics: []repository.BenchmarkMetric{
					repository.BenchmarkMetricBroadSearchActions,
					repository.BenchmarkMetricConsultedArtifacts,
				},
			},
			{
				Name:          repository.BenchmarkLaneContextAssembly,
				StartMarker:   "context_started",
				SuccessMarker: "context_ready",
				StopCondition: repository.BenchmarkStopCondition{
					Kind:   repository.BenchmarkStopConditionKindMarker,
					Marker: "context_ready",
				},
				FinalArtifact: &repository.BenchmarkFinalArtifactContract{
					ID:     "context-bundle",
					Name:   "Context bundle",
					Kind:   repository.BenchmarkFinalArtifactKindContextBundle,
					Path:   "artifacts/context.txt",
					Format: repository.BenchmarkFinalArtifactFormatText,
					Normalization: repository.BenchmarkFinalArtifactNormalization{
						Mode:           repository.BenchmarkFinalArtifactNormalizationModeTextTrimmed,
						TrimWhitespace: true,
					},
					Assertions: []repository.BenchmarkFinalArtifactAssertion{
						{Kind: repository.EvalAssertionKindContains, Contains: "LoadRolloutConfig"},
					},
				},
				Metrics: []repository.BenchmarkMetric{
					repository.BenchmarkMetricFileReadActions,
					repository.BenchmarkMetricBytesRead,
				},
			},
		},
		Arms: []repository.BenchmarkArmDefinition{
			{
				Kind: repository.BenchmarkArmKindBaseline,
				Name: "Baseline",
				Steps: []repository.BenchmarkStep{
					{
						ID:   "search",
						Name: "Search text",
						Lane: repository.BenchmarkLaneDiscovery,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:  repository.BenchmarkBaselineActionSearchText,
							Query: "LoadRolloutConfig",
						},
					},
					{
						ID:   "discovery-done",
						Name: "Mark discovery complete",
						Lane: repository.BenchmarkLaneDiscovery,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:   repository.BenchmarkBaselineActionMarkLaneComplete,
							Marker: "target_identified",
						},
					},
					{
						ID:   "read",
						Name: "Read slice",
						Lane: repository.BenchmarkLaneContextAssembly,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:      repository.BenchmarkBaselineActionReadFileSlice,
							Path:      "internal/http/handler/rollout.go",
							StartLine: 1,
							EndLine:   80,
						},
					},
					{
						ID:   "context-done",
						Name: "Mark context complete",
						Lane: repository.BenchmarkLaneContextAssembly,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:   repository.BenchmarkBaselineActionMarkLaneComplete,
							Marker: "context_ready",
						},
					},
				},
			},
			{
				Kind: repository.BenchmarkArmKindOptimusCtx,
				Name: "OptimusCtx",
				Steps: []repository.BenchmarkStep{
					{
						ID:   "map",
						Name: "Repository map",
						Lane: repository.BenchmarkLaneDiscovery,
						Treatment: &repository.BenchmarkTreatmentAction{
							Surface: repository.BenchmarkTreatmentSurfaceMCP,
							Tool:    "optimusctx.repository_map",
						},
					},
					{
						ID:   "lookup",
						Name: "Symbol lookup",
						Lane: repository.BenchmarkLaneDiscovery,
						Treatment: &repository.BenchmarkTreatmentAction{
							Surface: repository.BenchmarkTreatmentSurfaceMCP,
							Tool:    "optimusctx.symbol_lookup",
						},
					},
					{
						ID:   "context",
						Name: "Targeted context",
						Lane: repository.BenchmarkLaneContextAssembly,
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

func writeBenchmarkSuiteFile(t *testing.T, path string, suite repository.BenchmarkSuiteDefinition) {
	t.Helper()

	data, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	writeEvalFixtureFile(t, path, string(data))
}

func writeBenchmarkMutationSuite(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "suite.json")
	writeBenchmarkSuiteFile(t, path, repository.BenchmarkSuiteDefinition{
		SchemaVersion: repository.BenchmarkSuiteSchemaV2,
		ID:            "go-benchmark-refresh-v1",
		Version:       "v1",
		Name:          "Go benchmark refresh and task completion",
		Boundary:      repository.DefaultBenchmarkBoundaryContract(),
		Fixture: repository.EvalFixtureRef{
			ID:          "go-worktree",
			Version:     "v1",
			Path:        "go-worktree/v1/repository",
			Materialize: repository.EvalFixtureModeCopyTree,
		},
		Task: repository.BenchmarkTaskDefinition{
			ID:         "docs-context",
			Prompt:     "Refresh after change and fetch the bounded updated notes context.",
			TargetPath: "docs/notes.txt",
		},
		CountedInputs: []repository.BenchmarkCountedInputDefinition{
			{
				ID:         "baseline-refresh-paths",
				ArmKind:    repository.BenchmarkArmKindBaseline,
				Lane:       repository.BenchmarkLaneRefreshReady,
				StepID:     "grep-note",
				Name:       "Baseline mutated note matches",
				Kind:       repository.BenchmarkCountedInputKindPathList,
				SourceKind: repository.BenchmarkTokenEstimateSourcePathEstimate,
				Path:       "docs/notes.txt",
			},
			{
				ID:           "treatment-health-summary",
				ArmKind:      repository.BenchmarkArmKindOptimusCtx,
				Lane:         repository.BenchmarkLaneRefreshReady,
				StepID:       "health",
				Name:         "Projected health freshness",
				Kind:         repository.BenchmarkCountedInputKindJSONFieldProjection,
				SourceKind:   repository.BenchmarkTokenEstimateSourceDirectPayload,
				ArtifactType: repository.BenchmarkArtifactTypeHealth,
				ReportLabel:  repository.BenchmarkReportArtifactLabelOperational,
				JSONPath:     "freshness",
			},
			{
				ID:         "baseline-task-slice",
				ArmKind:    repository.BenchmarkArmKindBaseline,
				Lane:       repository.BenchmarkLaneTaskCompletion,
				StepID:     "read-docs",
				Name:       "Baseline updated docs slice",
				Kind:       repository.BenchmarkCountedInputKindFileSlice,
				SourceKind: repository.BenchmarkTokenEstimateSourceBoundedFileContent,
				Path:       "docs/notes.txt",
				StartLine:  1,
				EndLine:    20,
			},
			{
				ID:           "treatment-context",
				ArmKind:      repository.BenchmarkArmKindOptimusCtx,
				Lane:         repository.BenchmarkLaneTaskCompletion,
				StepID:       "context",
				Name:         "Treatment bounded updated notes context",
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
				StopCondition: repository.BenchmarkStopCondition{
					Kind:   repository.BenchmarkStopConditionKindMarker,
					Marker: "refresh_ready",
				},
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
						{Kind: repository.EvalAssertionKindJSONFieldPresent, Path: "targetReady"},
					},
				},
				Metrics: []repository.BenchmarkMetric{
					repository.BenchmarkMetricTargetedLookupActions,
					repository.BenchmarkMetricConsultedArtifacts,
				},
			},
			{
				Name: repository.BenchmarkLaneTaskCompletion,
				Assertions: []repository.BenchmarkAssertion{{
					File:     "docs/notes.txt",
					Kind:     repository.EvalAssertionKindContains,
					Contains: "mutated benchmark note",
				}},
				StopCondition: repository.BenchmarkStopCondition{
					Kind:   repository.BenchmarkStopConditionKindMarker,
					Marker: "task_complete",
				},
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
				Metrics: []repository.BenchmarkMetric{
					repository.BenchmarkMetricFileReadActions,
					repository.BenchmarkMetricBytesRead,
				},
			},
		},
		Arms: []repository.BenchmarkArmDefinition{
			{
				Kind: repository.BenchmarkArmKindBaseline,
				Name: "Baseline",
				Steps: []repository.BenchmarkStep{
					{
						ID:   "git-files",
						Name: "Inspect tracked files",
						Lane: repository.BenchmarkLaneRefreshReady,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind: repository.BenchmarkBaselineActionGitListFiles,
						},
					},
					{
						ID:   "grep-note",
						Name: "Search mutated note",
						Lane: repository.BenchmarkLaneRefreshReady,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:  repository.BenchmarkBaselineActionGitGrep,
							Query: "mutated benchmark note",
						},
					},
					{
						ID:   "refresh-ready",
						Name: "Mark refresh ready",
						Lane: repository.BenchmarkLaneRefreshReady,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:   repository.BenchmarkBaselineActionMarkLaneComplete,
							Marker: "refresh_ready",
						},
					},
					{
						ID:   "read-docs",
						Name: "Read updated docs note",
						Lane: repository.BenchmarkLaneTaskCompletion,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:      repository.BenchmarkBaselineActionReadFileSlice,
							Path:      "docs/notes.txt",
							StartLine: 1,
							EndLine:   20,
						},
					},
					{
						ID:   "task-complete",
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
	})
	return path
}

func newDeterministicBenchmarkClock() func() time.Time {
	base := time.Date(2026, time.March, 16, 12, 0, 0, 0, time.UTC)
	var tick int
	return func() time.Time {
		current := base.Add(time.Duration(tick) * 250 * time.Millisecond)
		tick++
		return current
	}
}

func benchmarkMutationToolExecutor(t *testing.T) BenchmarkToolExecutor {
	t.Helper()

	return func(_ context.Context, invocation BenchmarkToolInvocation) (BenchmarkToolExecutionResult, error) {
		switch invocation.Name {
		case "optimusctx.health":
			return BenchmarkToolExecutionResult{Payload: map[string]any{
				"freshness": "fresh",
				"state":     "ready",
			}}, nil
		case "optimusctx.targeted_context":
			return BenchmarkToolExecutionResult{Payload: repository.TargetedContextResult{
				Path:   "docs/notes.txt",
				Source: []string{"mutated benchmark note"},
			}}, nil
		default:
			t.Fatalf("unexpected tool %q", invocation.Name)
			return BenchmarkToolExecutionResult{}, nil
		}
	}
}
