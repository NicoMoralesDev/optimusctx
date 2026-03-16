package app

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

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
				Name: repository.BenchmarkLaneDiscovery,
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
				Name: repository.BenchmarkLaneContextAssembly,
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
