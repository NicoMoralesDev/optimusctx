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

	suite := benchmarkSuiteV2ForTest()
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
		SchemaVersion: BenchmarkSuiteSchemaV2,
		ID:            "go-benchmark-refresh-v2",
		Version:       "v2",
		Name:          "Go benchmark refresh and task completion",
		Boundary:      DefaultBenchmarkBoundaryContract(),
		Fixture: EvalFixtureRef{
			ID:          "go-worktree",
			Version:     "v2",
			Path:        "go-worktree/v2/repository",
			Materialize: EvalFixtureModeCopyTree,
		},
		Task: BenchmarkTaskDefinition{
			ID:         "docs-pack",
			Prompt:     "Refresh after mutation and export bounded context.",
			TargetPath: "docs/notes.txt",
		},
		CountedInputs: []BenchmarkCountedInputDefinition{
			{
				ID:         "baseline-readiness-paths",
				ArmKind:    BenchmarkArmKindBaseline,
				Lane:       BenchmarkLaneRefreshReady,
				StepID:     "git-grep",
				Name:       "Baseline readiness path hints",
				Kind:       BenchmarkCountedInputKindPathList,
				SourceKind: BenchmarkTokenEstimateSourcePathEstimate,
				Path:       "docs/notes.txt",
			},
			{
				ID:           "treatment-health-summary",
				ArmKind:      BenchmarkArmKindOptimusCtx,
				Lane:         BenchmarkLaneRefreshReady,
				StepID:       "health",
				Name:         "Projected readiness summary",
				Kind:         BenchmarkCountedInputKindJSONFieldProjection,
				SourceKind:   BenchmarkTokenEstimateSourceDirectPayload,
				ArtifactType: BenchmarkArtifactTypeHealth,
				ReportLabel:  BenchmarkReportArtifactLabelOperational,
				JSONPath:     "refresh.freshness",
			},
			{
				ID:         "baseline-task-slice",
				ArmKind:    BenchmarkArmKindBaseline,
				Lane:       BenchmarkLaneTaskCompletion,
				StepID:     "task-read",
				Name:       "Baseline updated notes slice",
				Kind:       BenchmarkCountedInputKindFileSlice,
				SourceKind: BenchmarkTokenEstimateSourceBoundedFileContent,
				Path:       "docs/notes.txt",
				StartLine:  1,
				EndLine:    20,
			},
			{
				ID:           "treatment-context-output",
				ArmKind:      BenchmarkArmKindOptimusCtx,
				Lane:         BenchmarkLaneTaskCompletion,
				StepID:       "context",
				Name:         "Projected updated notes context",
				Kind:         BenchmarkCountedInputKindTextOutput,
				SourceKind:   BenchmarkTokenEstimateSourceDirectPayload,
				ArtifactType: BenchmarkArtifactTypeL2Context,
				ReportLabel:  BenchmarkReportArtifactLabelL2Context,
				Path:         "artifacts/updated-notes.txt",
			},
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
				FinalArtifact: &BenchmarkFinalArtifactContract{
					ID:     "refresh-readiness",
					Name:   "Refresh readiness summary",
					Kind:   BenchmarkFinalArtifactKindReadinessSummary,
					Path:   "artifacts/readiness.json",
					Format: BenchmarkFinalArtifactFormatJSON,
					Normalization: BenchmarkFinalArtifactNormalization{
						Mode:      BenchmarkFinalArtifactNormalizationModeJSONFields,
						JSONPaths: []string{"freshness", "targetReady"},
					},
					Assertions: []BenchmarkFinalArtifactAssertion{
						{Kind: EvalAssertionKindJSONFieldPresent, Path: "freshness"},
					},
				},
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
				FinalArtifact: &BenchmarkFinalArtifactContract{
					ID:     "updated-notes-context",
					Name:   "Updated notes context",
					Kind:   BenchmarkFinalArtifactKindTaskOutput,
					Path:   "artifacts/updated-notes.txt",
					Format: BenchmarkFinalArtifactFormatText,
					Normalization: BenchmarkFinalArtifactNormalization{
						Mode:           BenchmarkFinalArtifactNormalizationModeTextTrimmed,
						TrimWhitespace: true,
					},
					Assertions: []BenchmarkFinalArtifactAssertion{
						{Kind: EvalAssertionKindContains, Contains: "mutated benchmark note"},
					},
				},
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
						ID:   "health",
						Name: "Inspect health",
						Lane: BenchmarkLaneRefreshReady,
						Treatment: &BenchmarkTreatmentAction{
							Surface: BenchmarkTreatmentSurfaceMCP,
							Tool:    "optimusctx.health",
						},
					},
					{
						ID:   "context",
						Name: "Fetch exact updated notes context",
						Lane: BenchmarkLaneTaskCompletion,
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
}

func TestBenchmarkSchemaValidationRejectsAmbiguousCounting(t *testing.T) {
	t.Parallel()

	t.Run("rejects raw repository map counting without projection", func(t *testing.T) {
		suite := benchmarkSuiteV2ForTest()
		suite.CountedInputs[2] = BenchmarkCountedInputDefinition{
			ID:           "treatment-raw-repository-map",
			ArmKind:      BenchmarkArmKindOptimusCtx,
			Lane:         BenchmarkLaneDiscovery,
			StepID:       "opti-repository-map",
			Name:         "Raw repository map payload",
			Kind:         BenchmarkCountedInputKindTextOutput,
			SourceKind:   BenchmarkTokenEstimateSourceDirectPayload,
			ArtifactType: BenchmarkArtifactTypeRepositoryMap,
			ReportLabel:  BenchmarkReportArtifactLabelRepositoryMap,
			Path:         "artifacts/repository-map.json",
		}

		err := suite.Validate()
		if err == nil || !strings.Contains(err.Error(), "repository_map") {
			t.Fatalf("Validate() error = %v, want repository_map projection failure", err)
		}
	})

	t.Run("rejects legacy completionArtifact", func(t *testing.T) {
		suite := benchmarkSuiteV2ForTest()
		suite.Task.FinalArtifact = nil
		suite.Task.CompletionArtifact = "artifacts/context-pack.txt"
		suite.Lanes[0].FinalArtifact = nil
		suite.Lanes[1].FinalArtifact = nil

		err := suite.Validate()
		if err == nil || !strings.Contains(err.Error(), "completionArtifact") {
			t.Fatalf("Validate() error = %v, want completionArtifact rejection", err)
		}
	})

	t.Run("rejects missing final artifact coverage", func(t *testing.T) {
		suite := benchmarkSuiteV2ForTest()
		suite.Task.FinalArtifact = nil
		suite.Lanes[0].FinalArtifact = nil

		err := suite.Validate()
		if err == nil || !strings.Contains(err.Error(), "requires finalArtifact or task.finalArtifact") {
			t.Fatalf("Validate() error = %v, want missing finalArtifact rejection", err)
		}
	})

	t.Run("rejects baseline artifact typing", func(t *testing.T) {
		suite := benchmarkSuiteV2ForTest()
		suite.CountedInputs[0].ArtifactType = BenchmarkArtifactTypeExactLookup
		suite.CountedInputs[0].ReportLabel = BenchmarkReportArtifactLabelExactLookup

		err := suite.Validate()
		if err == nil || !strings.Contains(err.Error(), "baseline counted inputs") {
			t.Fatalf("Validate() error = %v, want baseline counted-input rejection", err)
		}
	})
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
		suite := benchmarkSuiteV2ForTest()
		suite.Arms[0].Steps[0].Baseline = nil
		suite.Arms[0].Steps[0].Treatment = &BenchmarkTreatmentAction{
			Surface: BenchmarkTreatmentSurfaceMCP,
			Tool:    "optimusctx.repository_map",
		}

		err := suite.Validate()
		if err == nil || !strings.Contains(err.Error(), "baseline arm must use only baseline actions") {
			t.Fatalf("Validate() error = %v, want baseline restriction failure", err)
		}
	})
}

func benchmarkSuiteV2ForTest() BenchmarkSuiteDefinition {
	return BenchmarkSuiteDefinition{
		SchemaVersion: BenchmarkSuiteSchemaV2,
		ID:            "go-benchmark-discovery-v2",
		Version:       "v2",
		Name:          "Go benchmark discovery and context assembly",
		Description:   "Canonical paired benchmark contract for Phase 14.",
		Boundary:      DefaultBenchmarkBoundaryContract(),
		Fixture: EvalFixtureRef{
			ID:           "go-benchmark",
			Version:      "v2",
			Path:         "go-benchmark/v2/repository",
			Materialize:  EvalFixtureModeCopyTree,
			WorkspaceDir: "workspace",
		},
		Task: BenchmarkTaskDefinition{
			ID:           "handler-owner",
			Prompt:       "Find the HTTP handler that owns the rollout configuration path and assemble the exact supporting context.",
			TargetPath:   "internal/http/handler/rollout.go",
			TargetSymbol: "LoadRolloutConfig",
			ContextPaths: []string{
				"internal/http/handler/rollout.go",
				"internal/config/loader.go",
			},
			FinalArtifact: &BenchmarkFinalArtifactContract{
				ID:     "task-context-pack",
				Name:   "Canonical context pack",
				Kind:   BenchmarkFinalArtifactKindContextBundle,
				Path:   "artifacts/context-pack.txt",
				Format: BenchmarkFinalArtifactFormatText,
				Normalization: BenchmarkFinalArtifactNormalization{
					Mode:           BenchmarkFinalArtifactNormalizationModeTextLines,
					TrimWhitespace: true,
					SortLines:      true,
				},
				Assertions: []BenchmarkFinalArtifactAssertion{
					{Kind: EvalAssertionKindContains, Contains: "LoadRolloutConfig"},
				},
			},
		},
		CountedInputs: []BenchmarkCountedInputDefinition{
			{
				ID:         "baseline-search-results",
				ArmKind:    BenchmarkArmKindBaseline,
				Lane:       BenchmarkLaneDiscovery,
				StepID:     "baseline-search",
				Name:       "Baseline symbol matches",
				Kind:       BenchmarkCountedInputKindPathList,
				SourceKind: BenchmarkTokenEstimateSourcePathEstimate,
				Path:       "internal/http/handler/rollout.go",
			},
			{
				ID:         "baseline-context-slice",
				ArmKind:    BenchmarkArmKindBaseline,
				Lane:       BenchmarkLaneContextAssembly,
				StepID:     "baseline-read",
				Name:       "Baseline bounded file slice",
				Kind:       BenchmarkCountedInputKindFileSlice,
				SourceKind: BenchmarkTokenEstimateSourceBoundedFileContent,
				Path:       "internal/http/handler/rollout.go",
				StartLine:  1,
				EndLine:    80,
			},
			{
				ID:           "treatment-symbol-match",
				ArmKind:      BenchmarkArmKindOptimusCtx,
				Lane:         BenchmarkLaneDiscovery,
				StepID:       "opti-symbol-lookup",
				Name:         "Projected symbol lookup match",
				Kind:         BenchmarkCountedInputKindJSONFieldProjection,
				SourceKind:   BenchmarkTokenEstimateSourceDirectPayload,
				ArtifactType: BenchmarkArtifactTypeExactLookup,
				ReportLabel:  BenchmarkReportArtifactLabelExactLookup,
				JSONPath:     "matches[0].path",
			},
			{
				ID:           "treatment-context-bundle",
				ArmKind:      BenchmarkArmKindOptimusCtx,
				Lane:         BenchmarkLaneContextAssembly,
				StepID:       "opti-targeted-context",
				Name:         "Bounded context payload",
				Kind:         BenchmarkCountedInputKindTextOutput,
				SourceKind:   BenchmarkTokenEstimateSourceDirectPayload,
				ArtifactType: BenchmarkArtifactTypeL2Context,
				ReportLabel:  BenchmarkReportArtifactLabelL2Context,
				Path:         "artifacts/context-pack.txt",
			},
		},
		Lanes: []BenchmarkLaneDefinition{
			{
				Name: BenchmarkLaneDiscovery,
				FinalArtifact: &BenchmarkFinalArtifactContract{
					ID:     "discovery-target-locator",
					Name:   "Target locator",
					Kind:   BenchmarkFinalArtifactKindTargetLocator,
					Path:   "artifacts/target.json",
					Format: BenchmarkFinalArtifactFormatJSON,
					Normalization: BenchmarkFinalArtifactNormalization{
						Mode:      BenchmarkFinalArtifactNormalizationModeJSONFields,
						JSONPaths: []string{"path", "symbol"},
					},
					Assertions: []BenchmarkFinalArtifactAssertion{
						{Kind: EvalAssertionKindJSONFieldEquals, Path: "path", Equals: "internal/http/handler/rollout.go"},
					},
				},
				StopCondition: BenchmarkStopCondition{
					Kind:   BenchmarkStopConditionKindMarker,
					Marker: "target_identified",
				},
				Metrics: []BenchmarkMetric{BenchmarkMetricBroadSearchActions, BenchmarkMetricConsultedArtifacts},
			},
			{
				Name: BenchmarkLaneContextAssembly,
				FinalArtifact: &BenchmarkFinalArtifactContract{
					ID:     "context-ready-pack",
					Name:   "Context-ready pack",
					Kind:   BenchmarkFinalArtifactKindContextBundle,
					Path:   "artifacts/context-pack.txt",
					Format: BenchmarkFinalArtifactFormatText,
					Normalization: BenchmarkFinalArtifactNormalization{
						Mode:           BenchmarkFinalArtifactNormalizationModeTextTrimmed,
						TrimWhitespace: true,
					},
					Assertions: []BenchmarkFinalArtifactAssertion{
						{Kind: EvalAssertionKindContains, Contains: "internal/http/handler/rollout.go"},
					},
				},
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
}
