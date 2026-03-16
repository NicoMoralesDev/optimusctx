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
		SchemaVersion: repository.BenchmarkSuiteSchemaV1,
		ID:            "invalid-baseline-v1",
		Version:       "v1",
		Name:          "Invalid baseline",
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
		},
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

func TestBenchmarkTaskCompletionLane(t *testing.T) {
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
		if reflect.DeepEqual(invocation.Args, []string{"pack", "export", "--format", "json", "--output", "artifacts/pack.json"}) {
			writeEvalFixtureFile(t, filepath.Join(invocation.WorkingDir, "artifacts", "pack.json"), "{\"documents\":[\"docs/notes.txt\"]}\n")
			return BenchmarkCommandExecutionResult{ExitCode: 0}, nil
		}
		return BenchmarkCommandExecutionResult{}, nil
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
	if !taskLane.Success {
		t.Fatalf("task lane = %+v, want success", taskLane)
	}
	if !strings.Contains(strings.Join(taskLane.EvidencePaths, ","), "artifacts/pack.json") {
		t.Fatalf("task evidence = %+v", taskLane.EvidencePaths)
	}
}

func validBenchmarkSuite() repository.BenchmarkSuiteDefinition {
	return repository.BenchmarkSuiteDefinition{
		SchemaVersion: repository.BenchmarkSuiteSchemaV1,
		ID:            "go-benchmark-discovery-v1",
		Version:       "v1",
		Name:          "Go benchmark discovery and context assembly",
		Fixture: repository.EvalFixtureRef{
			ID:          "go-benchmark",
			Version:     "v1",
			Path:        "go-benchmark/v1/repository",
			Materialize: repository.EvalFixtureModeCopyTree,
		},
		Task: repository.BenchmarkTaskDefinition{
			ID:                 "handler-owner",
			Prompt:             "Find the rollout handler owner and assemble the exact surrounding context.",
			TargetPath:         "internal/http/handler/rollout.go",
			TargetSymbol:       "LoadRolloutConfig",
			ContextPaths:       []string{"internal/http/handler/rollout.go", "internal/config/loader.go"},
			CompletionArtifact: "artifacts/context-pack.txt",
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
		SchemaVersion: repository.BenchmarkSuiteSchemaV1,
		ID:            "go-benchmark-refresh-v1",
		Version:       "v1",
		Name:          "Go benchmark refresh and task completion",
		Fixture: repository.EvalFixtureRef{
			ID:          "go-worktree",
			Version:     "v1",
			Path:        "go-worktree/v1/repository",
			Materialize: repository.EvalFixtureModeCopyTree,
		},
		Task: repository.BenchmarkTaskDefinition{
			ID:                 "docs-pack",
			Prompt:             "Refresh after change and export a pack artifact.",
			TargetPath:         "docs/notes.txt",
			CompletionArtifact: "artifacts/pack.json",
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
						ID:   "pack",
						Name: "Export pack",
						Lane: repository.BenchmarkLaneTaskCompletion,
						Treatment: &repository.BenchmarkTreatmentAction{
							Surface: repository.BenchmarkTreatmentSurfaceCLI,
							Command: repository.EvalCommandPackExport,
							Args:    []string{"--format", "json", "--output", "artifacts/pack.json"},
						},
					},
				},
			},
		},
	})
	return path
}
