package repository

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestBenchmarkTokenAttributionContract(t *testing.T) {
	t.Parallel()

	contract := DefaultBenchmarkTokenEstimateContract()
	if contract.Policy.Name != "bytes_div_4_ceiling" {
		t.Fatalf("Policy.Name = %q, want bytes_div_4_ceiling", contract.Policy.Name)
	}
	if contract.Policy.BytesPerToken != 4 {
		t.Fatalf("Policy.BytesPerToken = %d, want 4", contract.Policy.BytesPerToken)
	}
	if contract.UsageClaim != "estimated workflow-consumed tokens" {
		t.Fatalf("UsageClaim = %q", contract.UsageClaim)
	}
	if contract.BillingDisambiguator != "not provider-billed token invoices" {
		t.Fatalf("BillingDisambiguator = %q", contract.BillingDisambiguator)
	}
	if got := EstimateBenchmarkTokensFromBytes(17); got != 5 {
		t.Fatalf("EstimateBenchmarkTokensFromBytes(17) = %d, want 5", got)
	}
	if got := EstimateBenchmarkTokensFromBytes(0); got != 0 {
		t.Fatalf("EstimateBenchmarkTokensFromBytes(0) = %d, want 0", got)
	}
}

func TestBenchmarkArtifactTypeAttribution(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		tool      string
		command   EvalCommandName
		wantType  BenchmarkArtifactType
		wantLabel BenchmarkReportArtifactLabel
	}{
		{name: "repository map tool", tool: "optimusctx.repository_map", wantType: BenchmarkArtifactTypeRepositoryMap, wantLabel: BenchmarkReportArtifactLabelRepositoryMap},
		{name: "symbol lookup tool", tool: "optimusctx.symbol_lookup", wantType: BenchmarkArtifactTypeExactLookup, wantLabel: BenchmarkReportArtifactLabelExactLookup},
		{name: "structure lookup tool", tool: "optimusctx.structure_lookup", wantType: BenchmarkArtifactTypeExactLookup, wantLabel: BenchmarkReportArtifactLabelExactLookup},
		{name: "targeted context tool", tool: "optimusctx.targeted_context", wantType: BenchmarkArtifactTypeL2Context, wantLabel: BenchmarkReportArtifactLabelL2Context},
		{name: "layered context tool", tool: "optimusctx.layered_context_l1", wantType: BenchmarkArtifactTypeL2Context, wantLabel: BenchmarkReportArtifactLabelL2Context},
		{name: "pack export command", command: EvalCommandPackExport, wantType: BenchmarkArtifactTypePackExport, wantLabel: BenchmarkReportArtifactLabelPackExport},
		{name: "refresh command", command: EvalCommandRefresh, wantType: BenchmarkArtifactTypeRefresh, wantLabel: BenchmarkReportArtifactLabelOperational},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				gotType BenchmarkArtifactType
				ok      bool
			)
			if tc.tool != "" {
				gotType, ok = BenchmarkArtifactTypeForTool(tc.tool)
			} else {
				gotType, ok = BenchmarkArtifactTypeForCommand(tc.command)
			}
			if !ok {
				t.Fatalf("expected attribution mapping for %+v", tc)
			}
			if gotType != tc.wantType {
				t.Fatalf("artifact type = %q, want %q", gotType, tc.wantType)
			}
			if got := BenchmarkReportLabelForArtifactType(gotType); got != tc.wantLabel {
				t.Fatalf("report label = %q, want %q", got, tc.wantLabel)
			}
		})
	}

	if _, ok := BenchmarkArtifactTypeForTool("optimusctx.unknown"); ok {
		t.Fatal("unknown tool unexpectedly mapped")
	}
	if got := BenchmarkReportLabelForArtifactType(""); got != "" {
		t.Fatalf("empty artifact label = %q, want empty", got)
	}
}

func TestBenchmarkSuiteContracts(t *testing.T) {
	t.Parallel()

	suite := BenchmarkSuiteDefinition{
		SchemaVersion: BenchmarkSuiteSchemaV1,
		ID:            "go-benchmark-discovery-v1",
		Version:       "v1",
		Name:          "Go benchmark discovery and context assembly",
		Description:   "Canonical paired benchmark contract for Phase 11.",
		Fixture: EvalFixtureRef{
			ID:           "go-benchmark",
			Version:      "v1",
			Path:         "go-benchmark/v1/repository",
			Materialize:  EvalFixtureModeCopyTree,
			WorkspaceDir: "workspace",
		},
		Task: BenchmarkTaskDefinition{
			ID:                 "handler-owner",
			Prompt:             "Find the HTTP handler that owns the rollout configuration path and assemble the exact supporting context.",
			TargetPath:         "internal/http/handler/rollout.go",
			TargetSymbol:       "LoadRolloutConfig",
			ContextPaths:       []string{"internal/http/handler/rollout.go", "internal/config/loader.go"},
			CompletionArtifact: "artifacts/context-pack.txt",
		},
		Lanes: []BenchmarkLaneDefinition{
			{
				Name: BenchmarkLaneDiscovery,
				StopCondition: BenchmarkStopCondition{
					Kind:   BenchmarkStopConditionKindMarker,
					Marker: "target_identified",
				},
				Metrics: []BenchmarkMetric{BenchmarkMetricBroadSearchActions, BenchmarkMetricConsultedArtifacts},
			},
			{
				Name: BenchmarkLaneContextAssembly,
				StopCondition: BenchmarkStopCondition{
					Kind:   BenchmarkStopConditionKindMarker,
					Marker: "context_ready",
				},
				Metrics: []BenchmarkMetric{BenchmarkMetricFileReadActions, BenchmarkMetricBytesRead},
			},
		},
		Arms: []BenchmarkArmDefinition{
			{
				Kind: BenchmarkArmKindBaseline,
				Name: "Baseline exact-search workflow",
				Steps: []BenchmarkStep{
					{
						ID:   "baseline-search",
						Name: "Search for rollout symbol",
						Lane: BenchmarkLaneDiscovery,
						Baseline: &BenchmarkBaselineAction{
							Kind:  BenchmarkBaselineActionSearchText,
							Query: "LoadRolloutConfig",
						},
					},
					{
						ID:   "baseline-discovery-done",
						Name: "Mark target identified",
						Lane: BenchmarkLaneDiscovery,
						Baseline: &BenchmarkBaselineAction{
							Kind:   BenchmarkBaselineActionMarkLaneComplete,
							Marker: "target_identified",
						},
					},
					{
						ID:   "baseline-read",
						Name: "Read bounded file slice",
						Lane: BenchmarkLaneContextAssembly,
						Baseline: &BenchmarkBaselineAction{
							Kind:      BenchmarkBaselineActionReadFileSlice,
							Path:      "internal/http/handler/rollout.go",
							StartLine: 1,
							EndLine:   80,
						},
					},
					{
						ID:   "baseline-context-done",
						Name: "Mark context ready",
						Lane: BenchmarkLaneContextAssembly,
						Baseline: &BenchmarkBaselineAction{
							Kind:   BenchmarkBaselineActionMarkLaneComplete,
							Marker: "context_ready",
						},
					},
				},
			},
			{
				Kind: BenchmarkArmKindOptimusCtx,
				Name: "OptimusCtx lookup workflow",
				Steps: []BenchmarkStep{
					{
						ID:   "opti-repository-map",
						Name: "Inspect repository map",
						Lane: BenchmarkLaneDiscovery,
						Treatment: &BenchmarkTreatmentAction{
							Surface: BenchmarkTreatmentSurfaceMCP,
							Tool:    "optimusctx.repository_map",
						},
					},
					{
						ID:   "opti-symbol-lookup",
						Name: "Lookup rollout symbol",
						Lane: BenchmarkLaneDiscovery,
						Treatment: &BenchmarkTreatmentAction{
							Surface: BenchmarkTreatmentSurfaceMCP,
							Tool:    "optimusctx.symbol_lookup",
						},
					},
					{
						ID:   "opti-targeted-context",
						Name: "Fetch exact context",
						Lane: BenchmarkLaneContextAssembly,
						Treatment: &BenchmarkTreatmentAction{
							Surface: BenchmarkTreatmentSurfaceMCP,
							Tool:    "optimusctx.targeted_context",
						},
					},
				},
			},
		},
	}

	if err := suite.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	encoded, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}

	var decoded BenchmarkSuiteDefinition
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !reflect.DeepEqual(decoded, suite) {
		t.Fatalf("decoded suite mismatch\n got: %#v\nwant: %#v", decoded, suite)
	}
}

func TestBenchmarkMutationLaneContracts(t *testing.T) {
	t.Parallel()

	suite := BenchmarkSuiteDefinition{
		SchemaVersion: BenchmarkSuiteSchemaV1,
		ID:            "go-benchmark-refresh-v1",
		Version:       "v1",
		Name:          "Go benchmark refresh and task completion",
		Fixture: EvalFixtureRef{
			ID:          "go-worktree",
			Version:     "v1",
			Path:        "go-worktree/v1/repository",
			Materialize: EvalFixtureModeCopyTree,
		},
		Task: BenchmarkTaskDefinition{
			ID:                 "docs-pack",
			Prompt:             "Refresh after mutation and export a bounded pack.",
			TargetPath:         "docs/notes.txt",
			CompletionArtifact: "artifacts/pack.json",
		},
		Lanes: []BenchmarkLaneDefinition{
			{
				Name: BenchmarkLaneRefreshReady,
				Setup: []EvalSetupAction{{
					Kind:    EvalSetupActionOverwriteFile,
					Path:    "docs/notes.txt",
					Content: "mutated benchmark note\n",
				}},
				Assertions: []BenchmarkAssertion{{
					File:     "docs/notes.txt",
					Kind:     EvalAssertionKindContains,
					Contains: "mutated benchmark note",
				}},
				StopCondition: BenchmarkStopCondition{
					Kind:   BenchmarkStopConditionKindMarker,
					Marker: "refresh_ready",
				},
				Metrics: []BenchmarkMetric{BenchmarkMetricConsultedArtifacts},
			},
			{
				Name: BenchmarkLaneTaskCompletion,
				Assertions: []BenchmarkAssertion{{
					File:     "docs/notes.txt",
					Kind:     EvalAssertionKindContains,
					Contains: "mutated benchmark note",
				}},
				StopCondition: BenchmarkStopCondition{
					Kind:   BenchmarkStopConditionKindMarker,
					Marker: "task_complete",
				},
				Metrics: []BenchmarkMetric{BenchmarkMetricFileReadActions},
			},
		},
		Arms: []BenchmarkArmDefinition{
			{
				Kind: BenchmarkArmKindBaseline,
				Name: "Baseline",
				Steps: []BenchmarkStep{
					{
						ID:   "git-grep",
						Name: "Search mutated note",
						Lane: BenchmarkLaneRefreshReady,
						Baseline: &BenchmarkBaselineAction{
							Kind:  BenchmarkBaselineActionGitGrep,
							Query: "mutated benchmark note",
						},
					},
					{
						ID:   "refresh-done",
						Name: "Mark refresh complete",
						Lane: BenchmarkLaneRefreshReady,
						Baseline: &BenchmarkBaselineAction{
							Kind:   BenchmarkBaselineActionMarkLaneComplete,
							Marker: "refresh_ready",
						},
					},
					{
						ID:   "task-read",
						Name: "Read docs note",
						Lane: BenchmarkLaneTaskCompletion,
						Baseline: &BenchmarkBaselineAction{
							Kind:      BenchmarkBaselineActionReadFileSlice,
							Path:      "docs/notes.txt",
							StartLine: 1,
							EndLine:   20,
						},
					},
					{
						ID:   "task-done",
						Name: "Mark task complete",
						Lane: BenchmarkLaneTaskCompletion,
						Baseline: &BenchmarkBaselineAction{
							Kind:   BenchmarkBaselineActionMarkLaneComplete,
							Marker: "task_complete",
						},
					},
				},
			},
			{
				Kind: BenchmarkArmKindOptimusCtx,
				Name: "Treatment",
				Steps: []BenchmarkStep{
					{
						ID:   "refresh",
						Name: "Refresh repository",
						Lane: BenchmarkLaneRefreshReady,
						Treatment: &BenchmarkTreatmentAction{
							Surface: BenchmarkTreatmentSurfaceCLI,
							Command: EvalCommandRefresh,
						},
					},
					{
						ID:   "pack",
						Name: "Export pack",
						Lane: BenchmarkLaneTaskCompletion,
						Treatment: &BenchmarkTreatmentAction{
							Surface: BenchmarkTreatmentSurfaceCLI,
							Command: EvalCommandPackExport,
							Args:    []string{"--format", "json", "--output", "artifacts/pack.json"},
						},
					},
				},
			},
		},
	}

	if err := suite.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestBaselineActionValidation(t *testing.T) {
	t.Parallel()

	t.Run("rejects file slice without bounds", func(t *testing.T) {
		action := BenchmarkBaselineAction{
			Kind: BenchmarkBaselineActionReadFileSlice,
			Path: "internal/http/handler/rollout.go",
		}
		if err := action.validate(); err == nil || !strings.Contains(err.Error(), "requires positive startLine/endLine") {
			t.Fatalf("validate() error = %v, want missing line bounds", err)
		}
	})

	t.Run("rejects baseline arm using treatment action", func(t *testing.T) {
		suite := BenchmarkSuiteDefinition{
			SchemaVersion: BenchmarkSuiteSchemaV1,
			ID:            "invalid-v1",
			Version:       "v1",
			Name:          "Invalid",
			Fixture: EvalFixtureRef{
				ID:          "go-benchmark",
				Version:     "v1",
				Path:        "go-benchmark/v1/repository",
				Materialize: EvalFixtureModeCopyTree,
			},
			Task: BenchmarkTaskDefinition{
				ID:         "target",
				Prompt:     "Find target",
				TargetPath: "internal/http/handler/rollout.go",
			},
			Lanes: []BenchmarkLaneDefinition{{
				Name: BenchmarkLaneDiscovery,
				StopCondition: BenchmarkStopCondition{
					Kind:   BenchmarkStopConditionKindMarker,
					Marker: "target_identified",
				},
				Metrics: []BenchmarkMetric{BenchmarkMetricBroadSearchActions},
			}},
			Arms: []BenchmarkArmDefinition{
				{
					Kind: BenchmarkArmKindBaseline,
					Name: "Baseline",
					Steps: []BenchmarkStep{{
						ID:   "bad",
						Name: "Bad",
						Lane: BenchmarkLaneDiscovery,
						Treatment: &BenchmarkTreatmentAction{
							Surface: BenchmarkTreatmentSurfaceMCP,
							Tool:    "optimusctx.repository_map",
						},
					}},
				},
				{
					Kind: BenchmarkArmKindOptimusCtx,
					Name: "Treatment",
					Steps: []BenchmarkStep{{
						ID:   "ok",
						Name: "OK",
						Lane: BenchmarkLaneDiscovery,
						Treatment: &BenchmarkTreatmentAction{
							Surface: BenchmarkTreatmentSurfaceMCP,
							Tool:    "optimusctx.repository_map",
						},
					}},
				},
			},
		}

		err := suite.Validate()
		if err == nil || !strings.Contains(err.Error(), "baseline arm must use only baseline actions") {
			t.Fatalf("Validate() error = %v, want baseline restriction failure", err)
		}
	})
}
