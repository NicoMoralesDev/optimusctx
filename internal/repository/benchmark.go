package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const BenchmarkSuiteSchemaV1 = "optimusctx/benchmark-suite@v1"
const BenchmarkSuiteSchemaV2 = "optimusctx/benchmark-suite@v2"
const BenchmarkEvidenceBundleSchemaV1 = "optimusctx/benchmark-evidence@v1"
const BenchmarkEvidenceBundleSchemaV2 = "optimusctx/benchmark-evidence@v2"

const (
	BenchmarkTokenEstimatorPolicyName          = "bytes_div_4_ceiling"
	BenchmarkTokenEstimateUsageClaim           = "estimated workflow-consumed tokens"
	BenchmarkTokenEstimateBillingDisambiguator = "not provider-billed token invoices"
)

type BenchmarkArmKind string

const (
	BenchmarkArmKindBaseline   BenchmarkArmKind = "baseline"
	BenchmarkArmKindOptimusCtx BenchmarkArmKind = "optimusctx"
)

type BenchmarkLane string

const (
	BenchmarkLaneDiscovery       BenchmarkLane = "discovery"
	BenchmarkLaneContextAssembly BenchmarkLane = "context_assembly"
	BenchmarkLaneRefreshReady    BenchmarkLane = "refresh_after_change"
	BenchmarkLaneTaskCompletion  BenchmarkLane = "task_completion"
)

type BenchmarkMetric string

const (
	BenchmarkMetricBroadSearchActions    BenchmarkMetric = "broad_search_actions"
	BenchmarkMetricTargetedLookupActions BenchmarkMetric = "targeted_lookup_actions"
	BenchmarkMetricFileReadActions       BenchmarkMetric = "file_read_actions"
	BenchmarkMetricBytesRead             BenchmarkMetric = "bytes_read"
	BenchmarkMetricConsultedArtifacts    BenchmarkMetric = "consulted_artifacts"
)

type BenchmarkStopConditionKind string

const (
	BenchmarkStopConditionKindMarker BenchmarkStopConditionKind = "marker"
)

type BenchmarkBaselineActionKind string

const (
	BenchmarkBaselineActionListTree         BenchmarkBaselineActionKind = "list_tree"
	BenchmarkBaselineActionSearchText       BenchmarkBaselineActionKind = "search_text"
	BenchmarkBaselineActionReadFileSlice    BenchmarkBaselineActionKind = "read_file_slice"
	BenchmarkBaselineActionGitListFiles     BenchmarkBaselineActionKind = "git_list_files"
	BenchmarkBaselineActionGitGrep          BenchmarkBaselineActionKind = "git_grep"
	BenchmarkBaselineActionMarkLaneComplete BenchmarkBaselineActionKind = "mark_lane_complete"
)

type BenchmarkTreatmentSurface string

const (
	BenchmarkTreatmentSurfaceCLI BenchmarkTreatmentSurface = "cli"
	BenchmarkTreatmentSurfaceMCP BenchmarkTreatmentSurface = "mcp"
)

type BenchmarkArtifactType string

const (
	BenchmarkArtifactTypeRepositoryMap BenchmarkArtifactType = "repository_map"
	BenchmarkArtifactTypeExactLookup   BenchmarkArtifactType = "exact_lookup"
	BenchmarkArtifactTypeL2Context     BenchmarkArtifactType = "l2_context"
	BenchmarkArtifactTypePackExport    BenchmarkArtifactType = "pack_export"
	BenchmarkArtifactTypeHealth        BenchmarkArtifactType = "health"
	BenchmarkArtifactTypeRefresh       BenchmarkArtifactType = "refresh"
)

type BenchmarkReportArtifactLabel string

const (
	BenchmarkReportArtifactLabelRepositoryMap BenchmarkReportArtifactLabel = "repository_map"
	BenchmarkReportArtifactLabelExactLookup   BenchmarkReportArtifactLabel = "exact_lookup"
	BenchmarkReportArtifactLabelL2Context     BenchmarkReportArtifactLabel = "l2_context"
	BenchmarkReportArtifactLabelPackExport    BenchmarkReportArtifactLabel = "pack_export"
	BenchmarkReportArtifactLabelOperational   BenchmarkReportArtifactLabel = "operational"
)

type BenchmarkTokenEstimateSourceKind string

const (
	BenchmarkTokenEstimateSourceBoundedFileContent BenchmarkTokenEstimateSourceKind = "bounded_file_content"
	BenchmarkTokenEstimateSourceDirectPayload      BenchmarkTokenEstimateSourceKind = "direct_payload"
	BenchmarkTokenEstimateSourcePathEstimate       BenchmarkTokenEstimateSourceKind = "path_estimate"
	BenchmarkTokenEstimateSourcePackExportSection  BenchmarkTokenEstimateSourceKind = "pack_export_section_estimate"
)

type BenchmarkBoundaryCountPolicy string

const (
	BenchmarkBoundaryCountPolicyDeclaredAgentInputsOnly BenchmarkBoundaryCountPolicy = "declared_agent_inputs_only"
)

type BenchmarkBoundaryProvenancePolicy string

const (
	BenchmarkBoundaryProvenancePolicyPersistSystemOutputs BenchmarkBoundaryProvenancePolicy = "persist_system_outputs"
)

type BenchmarkBoundaryFinalArtifactPolicy string

const (
	BenchmarkBoundaryFinalArtifactPolicyRequiredPerLaneOrTask BenchmarkBoundaryFinalArtifactPolicy = "required_per_lane_or_task"
)

type BenchmarkBoundaryContract struct {
	CountedInputs  BenchmarkBoundaryCountPolicy         `json:"countedInputs"`
	SystemOutputs  BenchmarkBoundaryProvenancePolicy    `json:"systemOutputs"`
	FinalArtifacts BenchmarkBoundaryFinalArtifactPolicy `json:"finalArtifacts"`
}

type BenchmarkCountedInputKind string

const (
	BenchmarkCountedInputKindPathList            BenchmarkCountedInputKind = "path_list"
	BenchmarkCountedInputKindFileSlice           BenchmarkCountedInputKind = "file_slice"
	BenchmarkCountedInputKindJSONFieldProjection BenchmarkCountedInputKind = "json_field_projection"
	BenchmarkCountedInputKindTextOutput          BenchmarkCountedInputKind = "text_output"
	BenchmarkCountedInputKindPackSection         BenchmarkCountedInputKind = "pack_section"
)

type BenchmarkFinalArtifactKind string

const (
	BenchmarkFinalArtifactKindTargetLocator    BenchmarkFinalArtifactKind = "target_locator"
	BenchmarkFinalArtifactKindContextBundle    BenchmarkFinalArtifactKind = "context_bundle"
	BenchmarkFinalArtifactKindReadinessSummary BenchmarkFinalArtifactKind = "readiness_summary"
	BenchmarkFinalArtifactKindTaskOutput       BenchmarkFinalArtifactKind = "task_output"
)

type BenchmarkFinalArtifactFormat string

const (
	BenchmarkFinalArtifactFormatText BenchmarkFinalArtifactFormat = "text"
	BenchmarkFinalArtifactFormatJSON BenchmarkFinalArtifactFormat = "json"
)

type BenchmarkFinalArtifactNormalizationMode string

const (
	BenchmarkFinalArtifactNormalizationModeTextTrimmed BenchmarkFinalArtifactNormalizationMode = "text_trimmed"
	BenchmarkFinalArtifactNormalizationModeTextLines   BenchmarkFinalArtifactNormalizationMode = "text_lines"
	BenchmarkFinalArtifactNormalizationModeJSONExact   BenchmarkFinalArtifactNormalizationMode = "json_exact"
	BenchmarkFinalArtifactNormalizationModeJSONFields  BenchmarkFinalArtifactNormalizationMode = "json_fields"
)

type BenchmarkCountedInputDefinition struct {
	ID           string                           `json:"id"`
	ArmKind      BenchmarkArmKind                 `json:"armKind"`
	Lane         BenchmarkLane                    `json:"lane"`
	StepID       string                           `json:"stepId"`
	Name         string                           `json:"name"`
	Kind         BenchmarkCountedInputKind        `json:"kind"`
	SourceKind   BenchmarkTokenEstimateSourceKind `json:"sourceKind"`
	ArtifactType BenchmarkArtifactType            `json:"artifactType,omitempty"`
	ReportLabel  BenchmarkReportArtifactLabel     `json:"reportLabel,omitempty"`
	Path         string                           `json:"path,omitempty"`
	JSONPath     string                           `json:"jsonPath,omitempty"`
	StartLine    int                              `json:"startLine,omitempty"`
	EndLine      int                              `json:"endLine,omitempty"`
}

type BenchmarkFinalArtifactNormalization struct {
	Mode           BenchmarkFinalArtifactNormalizationMode `json:"mode"`
	JSONPaths      []string                                `json:"jsonPaths,omitempty"`
	TrimWhitespace bool                                    `json:"trimWhitespace,omitempty"`
	SortLines      bool                                    `json:"sortLines,omitempty"`
}

type BenchmarkFinalArtifactAssertion struct {
	Kind     EvalAssertionKind `json:"kind"`
	Path     string            `json:"path,omitempty"`
	Contains string            `json:"contains,omitempty"`
	Equals   any               `json:"equals,omitempty"`
}

type BenchmarkFinalArtifactContract struct {
	ID            string                              `json:"id"`
	Name          string                              `json:"name"`
	Kind          BenchmarkFinalArtifactKind          `json:"kind"`
	Path          string                              `json:"path"`
	Format        BenchmarkFinalArtifactFormat        `json:"format"`
	Normalization BenchmarkFinalArtifactNormalization `json:"normalization"`
	Assertions    []BenchmarkFinalArtifactAssertion   `json:"assert"`
}

type BenchmarkTokenEstimateContract struct {
	Policy               BudgetEstimatePolicy `json:"policy"`
	UsageClaim           string               `json:"usageClaim"`
	BillingDisambiguator string               `json:"billingDisambiguator"`
}

type BenchmarkEvidenceVerification struct {
	Passed            bool     `json:"passed"`
	FailureReason     string   `json:"failureReason,omitempty"`
	DriftReasons      []string `json:"driftReasons,omitempty"`
	InvalidRunReasons []string `json:"invalidRunReasons,omitempty"`
}

type BenchmarkEvidenceBoundary string

const (
	BenchmarkEvidenceBoundaryAgentInput            BenchmarkEvidenceBoundary = "agent_input"
	BenchmarkEvidenceBoundarySystemProvenance      BenchmarkEvidenceBoundary = "system_provenance"
	BenchmarkEvidenceBoundaryFinalArtifactVerified BenchmarkEvidenceBoundary = "final_artifact_verification"
)

type BenchmarkLaneFinalArtifactVerification struct {
	ContractID    string `json:"contractId"`
	Path          string `json:"path"`
	Passed        bool   `json:"passed"`
	FailureReason string `json:"failureReason,omitempty"`
}

type BenchmarkLaneFinalArtifactSnapshot struct {
	Lane     BenchmarkLane                  `json:"lane"`
	Contract BenchmarkFinalArtifactContract `json:"contract"`
}

type BenchmarkMethodologySnapshot struct {
	SuiteSchemaVersion string                               `json:"suiteSchemaVersion"`
	Boundary           BenchmarkBoundaryContract            `json:"boundary"`
	CountedInputs      []BenchmarkCountedInputDefinition    `json:"countedInputs"`
	TaskFinalArtifact  *BenchmarkFinalArtifactContract      `json:"taskFinalArtifact,omitempty"`
	LaneFinalArtifacts []BenchmarkLaneFinalArtifactSnapshot `json:"laneFinalArtifacts,omitempty"`
}

type BenchmarkEvidenceInt64Stats struct {
	Min    int64 `json:"min"`
	Max    int64 `json:"max"`
	Median int64 `json:"median"`
	Mean   int64 `json:"mean"`
}

type BenchmarkEvidenceLaneSummary struct {
	Lane                   BenchmarkLane               `json:"lane"`
	AttemptCount           int                         `json:"attemptCount"`
	SuccessCount           int                         `json:"successCount"`
	InvalidAttemptCount    int                         `json:"invalidAttemptCount"`
	ElapsedMS              BenchmarkEvidenceInt64Stats `json:"elapsedMs"`
	ActionCount            BenchmarkEvidenceInt64Stats `json:"actionCount"`
	BroadSearchActions     BenchmarkEvidenceInt64Stats `json:"broadSearchActions"`
	TargetedLookupActions  BenchmarkEvidenceInt64Stats `json:"targetedLookupActions"`
	FileReadActions        BenchmarkEvidenceInt64Stats `json:"fileReadActions"`
	BytesRead              BenchmarkEvidenceInt64Stats `json:"bytesRead"`
	ConsultedArtifacts     []string                    `json:"consultedArtifacts,omitempty"`
	RejectedAttemptReasons []string                    `json:"rejectedAttemptReasons,omitempty"`
}

type BenchmarkEvidenceArmSummary struct {
	ArmKind BenchmarkArmKind               `json:"armKind"`
	ArmName string                         `json:"armName"`
	Lanes   []BenchmarkEvidenceLaneSummary `json:"lanes"`
}

type BenchmarkEvidenceLane struct {
	Lane           BenchmarkLane                           `json:"lane"`
	StartMarker    string                                  `json:"startMarker"`
	SuccessMarker  string                                  `json:"successMarker"`
	StopMarker     string                                  `json:"stopMarker"`
	SetupAppliedAt time.Time                               `json:"setupAppliedAt,omitempty"`
	StartedAt      time.Time                               `json:"startedAt"`
	FinishedAt     time.Time                               `json:"finishedAt"`
	ElapsedMS      int64                                   `json:"elapsedMs"`
	Success        bool                                    `json:"success"`
	EvidencePaths  []string                                `json:"evidencePaths,omitempty"`
	Effort         BenchmarkLaneEffort                     `json:"effort"`
	FinalArtifact  *BenchmarkLaneFinalArtifactVerification `json:"finalArtifact,omitempty"`
	Attribution    []BenchmarkArtifactConsumption          `json:"attribution,omitempty"`
}

type BenchmarkEvidenceArmAttempt struct {
	Kind          BenchmarkArmKind        `json:"kind"`
	Name          string                  `json:"name"`
	WorkspacePath string                  `json:"workspacePath,omitempty"`
	StartedAt     time.Time               `json:"startedAt"`
	FinishedAt    time.Time               `json:"finishedAt"`
	Lanes         []BenchmarkEvidenceLane `json:"lanes"`
}

type BenchmarkEvidenceAttempt struct {
	Attempt int                           `json:"attempt"`
	Arms    []BenchmarkEvidenceArmAttempt `json:"arms"`
}

type BenchmarkEvidenceBundle struct {
	SchemaVersion          string                         `json:"schemaVersion"`
	GeneratedAt            time.Time                      `json:"generatedAt"`
	RepositoryRoot         string                         `json:"repositoryRoot"`
	SuiteID                string                         `json:"suiteId"`
	SuiteVersion           string                         `json:"suiteVersion"`
	FixtureID              string                         `json:"fixtureId"`
	FixturePath            string                         `json:"fixturePath"`
	Methodology            BenchmarkMethodologySnapshot   `json:"methodology"`
	TokenEstimateContract  BenchmarkTokenEstimateContract `json:"tokenEstimateContract"`
	MethodologyFingerprint string                         `json:"methodologyFingerprint"`
	RerunCommand           string                         `json:"rerunCommand"`
	Verification           BenchmarkEvidenceVerification  `json:"verification"`
	Comparison             []BenchmarkEvidenceArmSummary  `json:"comparison"`
	Attempts               []BenchmarkEvidenceAttempt     `json:"attempts"`
}

type BenchmarkArtifactConsumption struct {
	StepID          string                           `json:"stepId"`
	StepName        string                           `json:"stepName,omitempty"`
	Lane            BenchmarkLane                    `json:"lane"`
	Boundary        BenchmarkEvidenceBoundary        `json:"boundary"`
	Surface         BenchmarkTreatmentSurface        `json:"surface,omitempty"`
	Command         EvalCommandName                  `json:"command,omitempty"`
	Tool            string                           `json:"tool,omitempty"`
	ArtifactType    BenchmarkArtifactType            `json:"artifactType,omitempty"`
	ReportLabel     BenchmarkReportArtifactLabel     `json:"reportLabel,omitempty"`
	SourceKind      BenchmarkTokenEstimateSourceKind `json:"sourceKind"`
	ArtifactPath    string                           `json:"artifactPath,omitempty"`
	EstimatedBytes  int64                            `json:"estimatedBytes,omitempty"`
	EstimatedTokens int64                            `json:"estimatedTokens,omitempty"`
}

type BenchmarkSuiteDefinition struct {
	SchemaVersion string                            `json:"schemaVersion"`
	ID            string                            `json:"id"`
	Version       string                            `json:"version"`
	Name          string                            `json:"name"`
	Description   string                            `json:"description,omitempty"`
	Boundary      BenchmarkBoundaryContract         `json:"boundary"`
	Fixture       EvalFixtureRef                    `json:"fixture"`
	Task          BenchmarkTaskDefinition           `json:"task"`
	CountedInputs []BenchmarkCountedInputDefinition `json:"countedInputs"`
	Lanes         []BenchmarkLaneDefinition         `json:"lanes"`
	Arms          []BenchmarkArmDefinition          `json:"arms"`
}

type BenchmarkTaskDefinition struct {
	ID                 string                          `json:"id"`
	Prompt             string                          `json:"prompt"`
	TargetPath         string                          `json:"targetPath,omitempty"`
	TargetSymbol       string                          `json:"targetSymbol,omitempty"`
	ContextPaths       []string                        `json:"contextPaths,omitempty"`
	CompletionArtifact string                          `json:"completionArtifact,omitempty"`
	FinalArtifact      *BenchmarkFinalArtifactContract `json:"finalArtifact,omitempty"`
}

type BenchmarkLaneDefinition struct {
	Name          BenchmarkLane                   `json:"name"`
	Description   string                          `json:"description,omitempty"`
	StartMarker   string                          `json:"startMarker,omitempty"`
	SuccessMarker string                          `json:"successMarker,omitempty"`
	Setup         []EvalSetupAction               `json:"setup,omitempty"`
	Assertions    []BenchmarkAssertion            `json:"assert,omitempty"`
	FinalArtifact *BenchmarkFinalArtifactContract `json:"finalArtifact,omitempty"`
	StopCondition BenchmarkStopCondition          `json:"stopCondition"`
	Metrics       []BenchmarkMetric               `json:"metrics"`
}

type BenchmarkAssertion struct {
	File     string            `json:"file"`
	Kind     EvalAssertionKind `json:"kind"`
	Path     string            `json:"path,omitempty"`
	Contains string            `json:"contains,omitempty"`
	Equals   any               `json:"equals,omitempty"`
}

type BenchmarkStopCondition struct {
	Kind        BenchmarkStopConditionKind `json:"kind"`
	Marker      string                     `json:"marker,omitempty"`
	Description string                     `json:"description,omitempty"`
}

type BenchmarkArmDefinition struct {
	Kind        BenchmarkArmKind `json:"kind"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Steps       []BenchmarkStep  `json:"steps"`
}

type BenchmarkStep struct {
	ID        string                    `json:"id"`
	Name      string                    `json:"name"`
	Lane      BenchmarkLane             `json:"lane"`
	Baseline  *BenchmarkBaselineAction  `json:"baseline,omitempty"`
	Treatment *BenchmarkTreatmentAction `json:"treatment,omitempty"`
}

type BenchmarkBaselineAction struct {
	Kind      BenchmarkBaselineActionKind `json:"kind"`
	Path      string                      `json:"path,omitempty"`
	Query     string                      `json:"query,omitempty"`
	StartLine int                         `json:"startLine,omitempty"`
	EndLine   int                         `json:"endLine,omitempty"`
	Marker    string                      `json:"marker,omitempty"`
}

type BenchmarkTreatmentAction struct {
	Surface BenchmarkTreatmentSurface `json:"surface"`
	Command EvalCommandName           `json:"command,omitempty"`
	Args    []string                  `json:"args,omitempty"`
	Tool    string                    `json:"tool,omitempty"`
}

type BenchmarkRunResult struct {
	SchemaVersion string                  `json:"schemaVersion"`
	SuiteID       string                  `json:"suiteId"`
	SuiteVersion  string                  `json:"suiteVersion"`
	FixtureID     string                  `json:"fixtureId"`
	FixturePath   string                  `json:"fixturePath"`
	WorkspacePath string                  `json:"workspacePath"`
	StartedAt     time.Time               `json:"startedAt"`
	FinishedAt    time.Time               `json:"finishedAt"`
	Arms          []BenchmarkArmRunResult `json:"arms"`
}

type BenchmarkArmRunResult struct {
	Kind        BenchmarkArmKind         `json:"kind"`
	Name        string                   `json:"name"`
	Workspace   string                   `json:"workspacePath,omitempty"`
	StartedAt   time.Time                `json:"startedAt"`
	FinishedAt  time.Time                `json:"finishedAt"`
	LaneResults []BenchmarkLaneRunResult `json:"laneResults"`
}

type BenchmarkLaneRunResult struct {
	Lane           BenchmarkLane                           `json:"lane"`
	StartMarker    string                                  `json:"startMarker"`
	SuccessMarker  string                                  `json:"successMarker"`
	StopMarker     string                                  `json:"stopMarker"`
	SetupAppliedAt time.Time                               `json:"setupAppliedAt,omitempty"`
	StartedAt      time.Time                               `json:"startedAt"`
	FinishedAt     time.Time                               `json:"finishedAt"`
	Elapsed        time.Duration                           `json:"elapsed"`
	Success        bool                                    `json:"success"`
	Setup          []EvalSetupAction                       `json:"setup,omitempty"`
	Assertions     []BenchmarkAssertion                    `json:"assert,omitempty"`
	EvidencePaths  []string                                `json:"evidencePaths,omitempty"`
	Effort         BenchmarkLaneEffort                     `json:"effort"`
	FinalArtifact  *BenchmarkLaneFinalArtifactVerification `json:"finalArtifact,omitempty"`
	Attribution    []BenchmarkArtifactConsumption          `json:"attribution,omitempty"`
}

type BenchmarkLaneEffort struct {
	ActionCount           int64    `json:"actionCount"`
	BroadSearchActions    int64    `json:"broadSearchActions"`
	TargetedLookupActions int64    `json:"targetedLookupActions"`
	FileReadActions       int64    `json:"fileReadActions"`
	BytesRead             int64    `json:"bytesRead"`
	ConsultedArtifacts    []string `json:"consultedArtifacts,omitempty"`
}

func DefaultBenchmarkTokenEstimateContract() BenchmarkTokenEstimateContract {
	return BenchmarkTokenEstimateContract{
		Policy: BudgetEstimatePolicy{
			Name:          BenchmarkTokenEstimatorPolicyName,
			BytesPerToken: 4,
		},
		UsageClaim:           BenchmarkTokenEstimateUsageClaim,
		BillingDisambiguator: BenchmarkTokenEstimateBillingDisambiguator,
	}
}

func EstimateBenchmarkTokensFromBytes(byteCount int64) int64 {
	if byteCount <= 0 {
		return 0
	}
	policy := DefaultBenchmarkTokenEstimateContract().Policy
	return (byteCount + policy.BytesPerToken - 1) / policy.BytesPerToken
}

func (c BenchmarkArtifactConsumption) CountsTowardEstimatedTokens() bool {
	return c.Boundary == BenchmarkEvidenceBoundaryAgentInput
}

func NormalizeBenchmarkEvidenceBundle(bundle BenchmarkEvidenceBundle) BenchmarkEvidenceBundle {
	bundle.Methodology = NormalizeBenchmarkMethodologySnapshot(bundle.Methodology)
	bundle.Attempts = append([]BenchmarkEvidenceAttempt(nil), bundle.Attempts...)
	sort.SliceStable(bundle.Attempts, func(i, j int) bool {
		return bundle.Attempts[i].Attempt < bundle.Attempts[j].Attempt
	})
	for attemptIndex := range bundle.Attempts {
		arms := append([]BenchmarkEvidenceArmAttempt(nil), bundle.Attempts[attemptIndex].Arms...)
		sort.SliceStable(arms, func(i, j int) bool {
			if arms[i].Kind == arms[j].Kind {
				return arms[i].Name < arms[j].Name
			}
			return benchmarkArmSortKey(arms[i].Kind) < benchmarkArmSortKey(arms[j].Kind)
		})
		for armIndex := range arms {
			lanes := append([]BenchmarkEvidenceLane(nil), arms[armIndex].Lanes...)
			sort.SliceStable(lanes, func(i, j int) bool {
				return benchmarkLaneSortKey(lanes[i].Lane) < benchmarkLaneSortKey(lanes[j].Lane)
			})
			for laneIndex := range lanes {
				lanes[laneIndex].EvidencePaths = benchmarkUniqueSorted(lanes[laneIndex].EvidencePaths)
				lanes[laneIndex].Effort.ConsultedArtifacts = benchmarkUniqueSorted(lanes[laneIndex].Effort.ConsultedArtifacts)
				lanes[laneIndex].Attribution = append([]BenchmarkArtifactConsumption(nil), lanes[laneIndex].Attribution...)
				sort.SliceStable(lanes[laneIndex].Attribution, func(i, j int) bool {
					return compareBenchmarkAttribution(lanes[laneIndex].Attribution[i], lanes[laneIndex].Attribution[j]) < 0
				})
				if lanes[laneIndex].FinalArtifact != nil {
					current := *lanes[laneIndex].FinalArtifact
					lanes[laneIndex].FinalArtifact = &current
				}
			}
			arms[armIndex].Lanes = lanes
		}
		bundle.Attempts[attemptIndex].Arms = arms
	}

	bundle.Comparison = append([]BenchmarkEvidenceArmSummary(nil), bundle.Comparison...)
	sort.SliceStable(bundle.Comparison, func(i, j int) bool {
		if bundle.Comparison[i].ArmKind == bundle.Comparison[j].ArmKind {
			return bundle.Comparison[i].ArmName < bundle.Comparison[j].ArmName
		}
		return benchmarkArmSortKey(bundle.Comparison[i].ArmKind) < benchmarkArmSortKey(bundle.Comparison[j].ArmKind)
	})
	for armIndex := range bundle.Comparison {
		lanes := append([]BenchmarkEvidenceLaneSummary(nil), bundle.Comparison[armIndex].Lanes...)
		sort.SliceStable(lanes, func(i, j int) bool {
			return benchmarkLaneSortKey(lanes[i].Lane) < benchmarkLaneSortKey(lanes[j].Lane)
		})
		for laneIndex := range lanes {
			lanes[laneIndex].ConsultedArtifacts = benchmarkUniqueSorted(lanes[laneIndex].ConsultedArtifacts)
			lanes[laneIndex].RejectedAttemptReasons = benchmarkUniqueSorted(lanes[laneIndex].RejectedAttemptReasons)
		}
		bundle.Comparison[armIndex].Lanes = lanes
	}
	bundle.Verification.DriftReasons = benchmarkUniqueSorted(bundle.Verification.DriftReasons)
	bundle.Verification.InvalidRunReasons = benchmarkUniqueSorted(bundle.Verification.InvalidRunReasons)
	return bundle
}

func MarshalBenchmarkEvidenceBundle(bundle BenchmarkEvidenceBundle) ([]byte, error) {
	return json.MarshalIndent(NormalizeBenchmarkEvidenceBundle(bundle), "", "  ")
}

func benchmarkArmSortKey(kind BenchmarkArmKind) int {
	switch kind {
	case BenchmarkArmKindBaseline:
		return 0
	case BenchmarkArmKindOptimusCtx:
		return 1
	default:
		return 2
	}
}

func benchmarkLaneSortKey(lane BenchmarkLane) int {
	switch lane {
	case BenchmarkLaneDiscovery:
		return 0
	case BenchmarkLaneContextAssembly:
		return 1
	case BenchmarkLaneRefreshReady:
		return 2
	case BenchmarkLaneTaskCompletion:
		return 3
	default:
		return 4
	}
}

func compareBenchmarkAttribution(left BenchmarkArtifactConsumption, right BenchmarkArtifactConsumption) int {
	leftKey := strings.Join([]string{
		string(left.Lane),
		left.StepID,
		left.StepName,
		string(left.Boundary),
		string(left.Surface),
		string(left.Command),
		left.Tool,
		string(left.ArtifactType),
		string(left.ReportLabel),
		string(left.SourceKind),
		left.ArtifactPath,
		fmt.Sprintf("%020d", left.EstimatedBytes),
		fmt.Sprintf("%020d", left.EstimatedTokens),
	}, "|")
	rightKey := strings.Join([]string{
		string(right.Lane),
		right.StepID,
		right.StepName,
		string(right.Boundary),
		string(right.Surface),
		string(right.Command),
		right.Tool,
		string(right.ArtifactType),
		string(right.ReportLabel),
		string(right.SourceKind),
		right.ArtifactPath,
		fmt.Sprintf("%020d", right.EstimatedBytes),
		fmt.Sprintf("%020d", right.EstimatedTokens),
	}, "|")
	switch {
	case leftKey < rightKey:
		return -1
	case leftKey > rightKey:
		return 1
	default:
		return 0
	}
}

func NormalizeBenchmarkMethodologySnapshot(snapshot BenchmarkMethodologySnapshot) BenchmarkMethodologySnapshot {
	snapshot.CountedInputs = append([]BenchmarkCountedInputDefinition(nil), snapshot.CountedInputs...)
	sort.SliceStable(snapshot.CountedInputs, func(i, j int) bool {
		left := snapshot.CountedInputs[i]
		right := snapshot.CountedInputs[j]
		leftKey := strings.Join([]string{string(left.ArmKind), string(left.Lane), left.StepID, left.ID}, "|")
		rightKey := strings.Join([]string{string(right.ArmKind), string(right.Lane), right.StepID, right.ID}, "|")
		return leftKey < rightKey
	})
	snapshot.LaneFinalArtifacts = append([]BenchmarkLaneFinalArtifactSnapshot(nil), snapshot.LaneFinalArtifacts...)
	sort.SliceStable(snapshot.LaneFinalArtifacts, func(i, j int) bool {
		left := snapshot.LaneFinalArtifacts[i]
		right := snapshot.LaneFinalArtifacts[j]
		leftKey := strings.Join([]string{string(left.Lane), left.Contract.ID, left.Contract.Path}, "|")
		rightKey := strings.Join([]string{string(right.Lane), right.Contract.ID, right.Contract.Path}, "|")
		return leftKey < rightKey
	})
	if snapshot.TaskFinalArtifact != nil {
		current := *snapshot.TaskFinalArtifact
		snapshot.TaskFinalArtifact = &current
	}
	return snapshot
}

func benchmarkUniqueSorted(items []string) []string {
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		set[item] = struct{}{}
	}
	ordered := make([]string, 0, len(set))
	for item := range set {
		ordered = append(ordered, item)
	}
	sort.Strings(ordered)
	return ordered
}

func BenchmarkArtifactTypeForTool(tool string) (BenchmarkArtifactType, bool) {
	switch tool {
	case "optimusctx.repository_map":
		return BenchmarkArtifactTypeRepositoryMap, true
	case "optimusctx.symbol_lookup", "optimusctx.structure_lookup":
		return BenchmarkArtifactTypeExactLookup, true
	case "optimusctx.layered_context_l1", "optimusctx.targeted_context", "optimusctx.pack":
		return BenchmarkArtifactTypeL2Context, true
	case "optimusctx.health":
		return BenchmarkArtifactTypeHealth, true
	case "optimusctx.refresh":
		return BenchmarkArtifactTypeRefresh, true
	default:
		return "", false
	}
}

func BenchmarkArtifactTypeForCommand(command EvalCommandName) (BenchmarkArtifactType, bool) {
	switch command {
	case EvalCommandPackExport:
		return BenchmarkArtifactTypePackExport, true
	case EvalCommandRefresh:
		return BenchmarkArtifactTypeRefresh, true
	case EvalCommandDoctor:
		return BenchmarkArtifactTypeHealth, true
	default:
		return "", false
	}
}

func BenchmarkReportLabelForArtifactType(artifactType BenchmarkArtifactType) BenchmarkReportArtifactLabel {
	switch artifactType {
	case BenchmarkArtifactTypeRepositoryMap:
		return BenchmarkReportArtifactLabelRepositoryMap
	case BenchmarkArtifactTypeExactLookup:
		return BenchmarkReportArtifactLabelExactLookup
	case BenchmarkArtifactTypeL2Context:
		return BenchmarkReportArtifactLabelL2Context
	case BenchmarkArtifactTypePackExport:
		return BenchmarkReportArtifactLabelPackExport
	case BenchmarkArtifactTypeHealth, BenchmarkArtifactTypeRefresh:
		return BenchmarkReportArtifactLabelOperational
	default:
		return ""
	}
}

func (s BenchmarkSuiteDefinition) Validate() error {
	if s.SchemaVersion != BenchmarkSuiteSchemaV2 {
		return fmt.Errorf("schemaVersion must be %q", BenchmarkSuiteSchemaV2)
	}
	if strings.TrimSpace(s.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(s.Version) == "" {
		return errors.New("version is required")
	}
	if strings.TrimSpace(s.Name) == "" {
		return errors.New("name is required")
	}
	if err := s.Boundary.validate(); err != nil {
		return fmt.Errorf("boundary: %w", err)
	}
	if err := s.Fixture.validate(); err != nil {
		return fmt.Errorf("fixture: %w", err)
	}
	if err := s.Task.validate(); err != nil {
		return fmt.Errorf("task: %w", err)
	}
	if len(s.Lanes) == 0 {
		return errors.New("at least one lane is required")
	}
	laneDefs, err := validateBenchmarkLanes(s.Lanes)
	if err != nil {
		return err
	}
	if len(s.Arms) != 2 {
		return errors.New("exactly two arms are required")
	}
	stepRefs, err := validateBenchmarkArms(s.Arms, laneDefs)
	if err != nil {
		return err
	}
	if len(s.CountedInputs) == 0 {
		return errors.New("countedInputs: at least one declared agent input is required")
	}
	if err := validateBenchmarkCountedInputs(s.CountedInputs, stepRefs); err != nil {
		return fmt.Errorf("countedInputs: %w", err)
	}
	for idx, lane := range s.Lanes {
		if lane.FinalArtifact == nil && s.Task.FinalArtifact == nil {
			return fmt.Errorf("lanes[%d]: lane %q requires finalArtifact or task.finalArtifact", idx, lane.Name)
		}
	}
	return nil
}

func (t BenchmarkTaskDefinition) validate() error {
	if strings.TrimSpace(t.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(t.Prompt) == "" {
		return errors.New("prompt is required")
	}
	if strings.TrimSpace(t.TargetPath) == "" && strings.TrimSpace(t.TargetSymbol) == "" {
		return errors.New("targetPath or targetSymbol is required")
	}
	for idx, path := range t.ContextPaths {
		if err := validateEvalRelativePath(path); err != nil {
			return fmt.Errorf("contextPaths[%d]: %w", idx, err)
		}
	}
	if t.CompletionArtifact != "" {
		return errors.New("completionArtifact is not supported in v2; use finalArtifact")
	}
	if t.FinalArtifact != nil {
		if err := t.FinalArtifact.validate(); err != nil {
			return fmt.Errorf("finalArtifact: %w", err)
		}
	}
	return nil
}

func validateBenchmarkLanes(lanes []BenchmarkLaneDefinition) (map[BenchmarkLane]BenchmarkLaneDefinition, error) {
	seen := make(map[BenchmarkLane]BenchmarkLaneDefinition, len(lanes))
	for idx, lane := range lanes {
		switch lane.Name {
		case BenchmarkLaneDiscovery, BenchmarkLaneContextAssembly, BenchmarkLaneRefreshReady, BenchmarkLaneTaskCompletion:
		default:
			return nil, fmt.Errorf("lanes[%d]: unsupported lane %q", idx, lane.Name)
		}
		if _, exists := seen[lane.Name]; exists {
			return nil, fmt.Errorf("duplicate lane %q", lane.Name)
		}
		if strings.TrimSpace(lane.StartMarker) == "" {
			lane.StartMarker = lane.StartMarkerName()
		}
		if strings.TrimSpace(lane.SuccessMarker) == "" {
			lane.SuccessMarker = lane.SuccessMarkerName()
		}
		if err := lane.StopCondition.validate(); err != nil {
			return nil, fmt.Errorf("lanes[%d].stopCondition: %w", idx, err)
		}
		if err := validateEvalSetupActions(lane.Setup); err != nil {
			return nil, fmt.Errorf("lanes[%d].setup: %w", idx, err)
		}
		if err := validateBenchmarkAssertions(lane.Assertions); err != nil {
			return nil, fmt.Errorf("lanes[%d].assert: %w", idx, err)
		}
		if lane.FinalArtifact != nil {
			if err := lane.FinalArtifact.validate(); err != nil {
				return nil, fmt.Errorf("lanes[%d].finalArtifact: %w", idx, err)
			}
		}
		if len(lane.Metrics) == 0 {
			return nil, fmt.Errorf("lanes[%d]: at least one metric is required", idx)
		}
		metrics := make(map[BenchmarkMetric]struct{}, len(lane.Metrics))
		for metricIdx, metric := range lane.Metrics {
			switch metric {
			case BenchmarkMetricBroadSearchActions, BenchmarkMetricTargetedLookupActions, BenchmarkMetricFileReadActions, BenchmarkMetricBytesRead, BenchmarkMetricConsultedArtifacts:
			default:
				return nil, fmt.Errorf("lanes[%d].metrics[%d]: unsupported metric %q", idx, metricIdx, metric)
			}
			if _, exists := metrics[metric]; exists {
				return nil, fmt.Errorf("lanes[%d]: duplicate metric %q", idx, metric)
			}
			metrics[metric] = struct{}{}
		}
		if (lane.Name == BenchmarkLaneRefreshReady || lane.Name == BenchmarkLaneTaskCompletion) && len(lane.Assertions) == 0 {
			return nil, fmt.Errorf("lanes[%d]: lane %q requires at least one assertion", idx, lane.Name)
		}
		seen[lane.Name] = lane
	}
	return seen, nil
}

func validateBenchmarkAssertions(assertions []BenchmarkAssertion) error {
	for idx, assertion := range assertions {
		if err := validateEvalRelativePath(assertion.File); err != nil {
			return fmt.Errorf("assert[%d].file: %w", idx, err)
		}
		switch assertion.Kind {
		case EvalAssertionKindContains:
			if assertion.Contains == "" {
				return fmt.Errorf("assert[%d]: contains is required for %q", idx, assertion.Kind)
			}
			if assertion.Path != "" || assertion.Equals != nil {
				return fmt.Errorf("assert[%d]: path/equals must be empty for %q", idx, assertion.Kind)
			}
		case EvalAssertionKindJSONFieldPresent:
			if strings.TrimSpace(assertion.Path) == "" {
				return fmt.Errorf("assert[%d]: path is required for %q", idx, assertion.Kind)
			}
			if assertion.Contains != "" || assertion.Equals != nil {
				return fmt.Errorf("assert[%d]: contains/equals must be empty for %q", idx, assertion.Kind)
			}
		case EvalAssertionKindJSONFieldEquals:
			if strings.TrimSpace(assertion.Path) == "" {
				return fmt.Errorf("assert[%d]: path is required for %q", idx, assertion.Kind)
			}
			if assertion.Contains != "" {
				return fmt.Errorf("assert[%d]: contains must be empty for %q", idx, assertion.Kind)
			}
		default:
			return fmt.Errorf("assert[%d]: unsupported kind %q", idx, assertion.Kind)
		}
	}
	return nil
}

func (s BenchmarkStopCondition) validate() error {
	switch s.Kind {
	case BenchmarkStopConditionKindMarker:
		if strings.TrimSpace(s.Marker) == "" {
			return errors.New("marker is required")
		}
	default:
		return fmt.Errorf("unsupported stop condition kind %q", s.Kind)
	}
	return nil
}

func (l BenchmarkLaneDefinition) StartMarkerName() string {
	if strings.TrimSpace(l.StartMarker) != "" {
		return l.StartMarker
	}
	return string(l.Name) + "_started"
}

func (l BenchmarkLaneDefinition) SuccessMarkerName() string {
	if strings.TrimSpace(l.SuccessMarker) != "" {
		return l.SuccessMarker
	}
	return l.StopCondition.Marker
}

type benchmarkStepRef struct {
	ArmKind BenchmarkArmKind
	Lane    BenchmarkLane
	Step    BenchmarkStep
}

func validateBenchmarkArms(arms []BenchmarkArmDefinition, lanes map[BenchmarkLane]BenchmarkLaneDefinition) (map[string]benchmarkStepRef, error) {
	seenKinds := map[BenchmarkArmKind]struct{}{}
	stepRefs := make(map[string]benchmarkStepRef)
	for idx, arm := range arms {
		switch arm.Kind {
		case BenchmarkArmKindBaseline, BenchmarkArmKindOptimusCtx:
		default:
			return nil, fmt.Errorf("arms[%d]: unsupported kind %q", idx, arm.Kind)
		}
		if _, exists := seenKinds[arm.Kind]; exists {
			return nil, fmt.Errorf("duplicate arm kind %q", arm.Kind)
		}
		seenKinds[arm.Kind] = struct{}{}
		if strings.TrimSpace(arm.Name) == "" {
			return nil, fmt.Errorf("arms[%d]: name is required", idx)
		}
		if len(arm.Steps) == 0 {
			return nil, fmt.Errorf("arms[%d]: at least one step is required", idx)
		}
		armStepRefs, err := validateBenchmarkArmSteps(idx, arm, lanes)
		if err != nil {
			return nil, err
		}
		for stepID, ref := range armStepRefs {
			stepRefs[stepID] = ref
		}
	}
	if _, ok := seenKinds[BenchmarkArmKindBaseline]; !ok {
		return nil, errors.New("baseline arm is required")
	}
	if _, ok := seenKinds[BenchmarkArmKindOptimusCtx]; !ok {
		return nil, errors.New("optimusctx arm is required")
	}
	return stepRefs, nil
}

func validateBenchmarkArmSteps(armIdx int, arm BenchmarkArmDefinition, lanes map[BenchmarkLane]BenchmarkLaneDefinition) (map[string]benchmarkStepRef, error) {
	seenSteps := make(map[string]struct{}, len(arm.Steps))
	markers := make(map[BenchmarkLane]map[string]struct{})
	stepRefs := make(map[string]benchmarkStepRef, len(arm.Steps))
	for stepIdx, step := range arm.Steps {
		if strings.TrimSpace(step.ID) == "" {
			return nil, fmt.Errorf("arms[%d].steps[%d]: id is required", armIdx, stepIdx)
		}
		if _, exists := seenSteps[step.ID]; exists {
			return nil, fmt.Errorf("arms[%d]: duplicate step id %q", armIdx, step.ID)
		}
		seenSteps[step.ID] = struct{}{}
		if strings.TrimSpace(step.Name) == "" {
			return nil, fmt.Errorf("arms[%d].steps[%d]: name is required", armIdx, stepIdx)
		}
		if _, ok := lanes[step.Lane]; !ok {
			return nil, fmt.Errorf("arms[%d].steps[%d]: unknown lane %q", armIdx, stepIdx, step.Lane)
		}

		switch arm.Kind {
		case BenchmarkArmKindBaseline:
			if step.Baseline == nil || step.Treatment != nil {
				return nil, fmt.Errorf("arms[%d].steps[%d]: baseline arm must use only baseline actions", armIdx, stepIdx)
			}
			if err := step.Baseline.validate(); err != nil {
				return nil, fmt.Errorf("arms[%d].steps[%d].baseline: %w", armIdx, stepIdx, err)
			}
			if step.Baseline.Marker != "" {
				if markers[step.Lane] == nil {
					markers[step.Lane] = make(map[string]struct{})
				}
				markers[step.Lane][step.Baseline.Marker] = struct{}{}
			}
		case BenchmarkArmKindOptimusCtx:
			if step.Treatment == nil || step.Baseline != nil {
				return nil, fmt.Errorf("arms[%d].steps[%d]: optimusctx arm must use only treatment actions", armIdx, stepIdx)
			}
			if err := step.Treatment.validate(); err != nil {
				return nil, fmt.Errorf("arms[%d].steps[%d].treatment: %w", armIdx, stepIdx, err)
			}
		}
		stepRefs[step.ID] = benchmarkStepRef{ArmKind: arm.Kind, Lane: step.Lane, Step: step}
	}

	for laneName, laneDef := range lanes {
		if !armHasLane(arm, laneName) {
			return nil, fmt.Errorf("arms[%d]: lane %q is missing", armIdx, laneName)
		}
		if arm.Kind == BenchmarkArmKindBaseline && laneDef.StopCondition.Kind == BenchmarkStopConditionKindMarker {
			if _, ok := markers[laneName][laneDef.StopCondition.Marker]; !ok {
				return nil, fmt.Errorf("arms[%d]: lane %q does not emit stop marker %q", armIdx, laneName, laneDef.StopCondition.Marker)
			}
		}
	}
	return stepRefs, nil
}

func armHasLane(arm BenchmarkArmDefinition, lane BenchmarkLane) bool {
	for _, step := range arm.Steps {
		if step.Lane == lane {
			return true
		}
	}
	return false
}

func (a BenchmarkBaselineAction) validate() error {
	switch a.Kind {
	case BenchmarkBaselineActionListTree, BenchmarkBaselineActionGitListFiles:
		if a.Query != "" || a.StartLine != 0 || a.EndLine != 0 || a.Marker != "" {
			return fmt.Errorf("%q does not allow query, line bounds, or marker", a.Kind)
		}
		if a.Path != "" {
			if err := validateEvalRelativePath(a.Path); err != nil {
				return err
			}
		}
	case BenchmarkBaselineActionSearchText, BenchmarkBaselineActionGitGrep:
		if strings.TrimSpace(a.Query) == "" {
			return fmt.Errorf("%q requires query", a.Kind)
		}
		if a.StartLine != 0 || a.EndLine != 0 || a.Marker != "" {
			return fmt.Errorf("%q does not allow line bounds or marker", a.Kind)
		}
		if a.Path != "" {
			if err := validateEvalRelativePath(a.Path); err != nil {
				return err
			}
		}
	case BenchmarkBaselineActionReadFileSlice:
		if err := validateEvalRelativePath(a.Path); err != nil {
			return err
		}
		if a.StartLine <= 0 || a.EndLine <= 0 || a.EndLine < a.StartLine {
			return errors.New("read_file_slice requires positive startLine/endLine with endLine >= startLine")
		}
		if a.Query != "" || a.Marker != "" {
			return errors.New("read_file_slice does not allow query or marker")
		}
	case BenchmarkBaselineActionMarkLaneComplete:
		if strings.TrimSpace(a.Marker) == "" {
			return errors.New("mark_lane_complete requires marker")
		}
		if a.Path != "" || a.Query != "" || a.StartLine != 0 || a.EndLine != 0 {
			return errors.New("mark_lane_complete does not allow path, query, or line bounds")
		}
	default:
		return fmt.Errorf("unsupported baseline action %q", a.Kind)
	}
	return nil
}

func (a BenchmarkTreatmentAction) validate() error {
	switch a.Surface {
	case BenchmarkTreatmentSurfaceCLI:
		if a.Tool != "" {
			return errors.New("cli treatment does not allow tool")
		}
		if !EvalSupportsCommandSequence(a.Command) && !benchmarkSupportsCLICommand(a.Command) {
			return fmt.Errorf("unsupported cli command %q", a.Command)
		}
	case BenchmarkTreatmentSurfaceMCP:
		if a.Command != "" {
			return errors.New("mcp treatment does not allow command")
		}
		if !benchmarkSupportsMCPTool(a.Tool) {
			return fmt.Errorf("unsupported mcp tool %q", a.Tool)
		}
	default:
		return fmt.Errorf("unsupported treatment surface %q", a.Surface)
	}
	return nil
}

func benchmarkSupportsCLICommand(command EvalCommandName) bool {
	switch command {
	case EvalCommandInit, EvalCommandRefresh, EvalCommandDoctor, EvalCommandPackExport:
		return true
	default:
		return false
	}
}

func benchmarkSupportsMCPTool(tool string) bool {
	switch tool {
	case "optimusctx.repository_map", "optimusctx.layered_context_l0", "optimusctx.layered_context_l1",
		"optimusctx.symbol_lookup", "optimusctx.structure_lookup", "optimusctx.targeted_context",
		"optimusctx.refresh", "optimusctx.token_tree", "optimusctx.pack", "optimusctx.health":
		return true
	default:
		return false
	}
}

func DefaultBenchmarkBoundaryContract() BenchmarkBoundaryContract {
	return BenchmarkBoundaryContract{
		CountedInputs:  BenchmarkBoundaryCountPolicyDeclaredAgentInputsOnly,
		SystemOutputs:  BenchmarkBoundaryProvenancePolicyPersistSystemOutputs,
		FinalArtifacts: BenchmarkBoundaryFinalArtifactPolicyRequiredPerLaneOrTask,
	}
}

func (c BenchmarkBoundaryContract) validate() error {
	switch c.CountedInputs {
	case BenchmarkBoundaryCountPolicyDeclaredAgentInputsOnly:
	default:
		return fmt.Errorf("countedInputs must be %q", BenchmarkBoundaryCountPolicyDeclaredAgentInputsOnly)
	}
	switch c.SystemOutputs {
	case BenchmarkBoundaryProvenancePolicyPersistSystemOutputs:
	default:
		return fmt.Errorf("systemOutputs must be %q", BenchmarkBoundaryProvenancePolicyPersistSystemOutputs)
	}
	switch c.FinalArtifacts {
	case BenchmarkBoundaryFinalArtifactPolicyRequiredPerLaneOrTask:
	default:
		return fmt.Errorf("finalArtifacts must be %q", BenchmarkBoundaryFinalArtifactPolicyRequiredPerLaneOrTask)
	}
	return nil
}

func (c BenchmarkBoundaryContract) Validate() error {
	return c.validate()
}

func validateBenchmarkCountedInputs(inputs []BenchmarkCountedInputDefinition, stepRefs map[string]benchmarkStepRef) error {
	seen := make(map[string]struct{}, len(inputs))
	for idx, input := range inputs {
		if strings.TrimSpace(input.ID) == "" {
			return fmt.Errorf("[%d].id: is required", idx)
		}
		if _, ok := seen[input.ID]; ok {
			return fmt.Errorf("duplicate id %q", input.ID)
		}
		seen[input.ID] = struct{}{}
		if strings.TrimSpace(input.Name) == "" {
			return fmt.Errorf("[%d].name: is required", idx)
		}
		ref, ok := stepRefs[input.StepID]
		if !ok {
			return fmt.Errorf("[%d].stepId: unknown step %q", idx, input.StepID)
		}
		if input.ArmKind != ref.ArmKind {
			return fmt.Errorf("[%d].armKind: step %q belongs to %q", idx, input.StepID, ref.ArmKind)
		}
		if input.Lane != ref.Lane {
			return fmt.Errorf("[%d].lane: step %q belongs to %q", idx, input.StepID, ref.Lane)
		}
		if err := input.validate(); err != nil {
			return fmt.Errorf("[%d]: %w", idx, err)
		}
	}
	return nil
}

func (i BenchmarkCountedInputDefinition) validate() error {
	switch i.ArmKind {
	case BenchmarkArmKindBaseline, BenchmarkArmKindOptimusCtx:
	default:
		return fmt.Errorf("armKind %q is unsupported", i.ArmKind)
	}
	switch i.Lane {
	case BenchmarkLaneDiscovery, BenchmarkLaneContextAssembly, BenchmarkLaneRefreshReady, BenchmarkLaneTaskCompletion:
	default:
		return fmt.Errorf("lane %q is unsupported", i.Lane)
	}
	switch i.Kind {
	case BenchmarkCountedInputKindPathList, BenchmarkCountedInputKindFileSlice, BenchmarkCountedInputKindJSONFieldProjection, BenchmarkCountedInputKindTextOutput, BenchmarkCountedInputKindPackSection:
	default:
		return fmt.Errorf("kind %q is unsupported", i.Kind)
	}
	switch i.SourceKind {
	case BenchmarkTokenEstimateSourcePathEstimate, BenchmarkTokenEstimateSourceBoundedFileContent, BenchmarkTokenEstimateSourceDirectPayload, BenchmarkTokenEstimateSourcePackExportSection:
	default:
		return fmt.Errorf("sourceKind %q is unsupported", i.SourceKind)
	}
	if i.Path != "" {
		if err := validateEvalRelativePath(i.Path); err != nil {
			return fmt.Errorf("path: %w", err)
		}
	}
	if i.JSONPath != "" && strings.TrimSpace(i.JSONPath) == "" {
		return errors.New("jsonPath cannot be blank")
	}
	if i.ArtifactType != "" {
		expectedLabel := BenchmarkReportLabelForArtifactType(i.ArtifactType)
		if expectedLabel == "" {
			return fmt.Errorf("artifactType %q is unsupported", i.ArtifactType)
		}
		if i.ReportLabel != "" && i.ReportLabel != expectedLabel {
			return fmt.Errorf("reportLabel must be %q for artifactType %q", expectedLabel, i.ArtifactType)
		}
	}
	switch i.Kind {
	case BenchmarkCountedInputKindPathList:
		if i.SourceKind != BenchmarkTokenEstimateSourcePathEstimate {
			return fmt.Errorf("path_list requires sourceKind %q", BenchmarkTokenEstimateSourcePathEstimate)
		}
		if i.Path == "" {
			return errors.New("path_list requires path")
		}
		if i.JSONPath != "" || i.StartLine != 0 || i.EndLine != 0 {
			return errors.New("path_list does not allow jsonPath or line bounds")
		}
	case BenchmarkCountedInputKindFileSlice:
		if i.SourceKind != BenchmarkTokenEstimateSourceBoundedFileContent {
			return fmt.Errorf("file_slice requires sourceKind %q", BenchmarkTokenEstimateSourceBoundedFileContent)
		}
		if i.Path == "" {
			return errors.New("file_slice requires path")
		}
		if i.StartLine <= 0 || i.EndLine <= 0 || i.EndLine < i.StartLine {
			return errors.New("file_slice requires positive startLine/endLine with endLine >= startLine")
		}
		if i.JSONPath != "" {
			return errors.New("file_slice does not allow jsonPath")
		}
	case BenchmarkCountedInputKindJSONFieldProjection:
		if i.SourceKind != BenchmarkTokenEstimateSourceDirectPayload {
			return fmt.Errorf("json_field_projection requires sourceKind %q", BenchmarkTokenEstimateSourceDirectPayload)
		}
		if strings.TrimSpace(i.JSONPath) == "" {
			return errors.New("json_field_projection requires jsonPath")
		}
		if i.StartLine != 0 || i.EndLine != 0 {
			return errors.New("json_field_projection does not allow line bounds")
		}
	case BenchmarkCountedInputKindTextOutput:
		if i.Path == "" {
			return errors.New("text_output requires path")
		}
		if i.JSONPath != "" || i.StartLine != 0 || i.EndLine != 0 {
			return errors.New("text_output does not allow jsonPath or line bounds")
		}
	case BenchmarkCountedInputKindPackSection:
		if i.SourceKind != BenchmarkTokenEstimateSourcePackExportSection {
			return fmt.Errorf("pack_section requires sourceKind %q", BenchmarkTokenEstimateSourcePackExportSection)
		}
		if i.Path == "" {
			return errors.New("pack_section requires path")
		}
		if i.JSONPath != "" || i.StartLine != 0 || i.EndLine != 0 {
			return errors.New("pack_section does not allow jsonPath or line bounds")
		}
	}
	if benchmarkRequiresProjection(i.ArtifactType) && i.SourceKind == BenchmarkTokenEstimateSourceDirectPayload && i.Kind != BenchmarkCountedInputKindJSONFieldProjection {
		return fmt.Errorf("%q direct payloads must use json_field_projection", i.ArtifactType)
	}
	if i.ArmKind == BenchmarkArmKindBaseline && (i.ArtifactType != "" || i.ReportLabel != "") {
		return errors.New("baseline counted inputs must not declare artifactType or reportLabel")
	}
	return nil
}

func (i BenchmarkCountedInputDefinition) Validate() error {
	return i.validate()
}

func benchmarkRequiresProjection(artifactType BenchmarkArtifactType) bool {
	switch artifactType {
	case BenchmarkArtifactTypeRepositoryMap, BenchmarkArtifactTypeHealth, BenchmarkArtifactTypeRefresh:
		return true
	default:
		return false
	}
}

func (c BenchmarkFinalArtifactContract) validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(c.Name) == "" {
		return errors.New("name is required")
	}
	switch c.Kind {
	case BenchmarkFinalArtifactKindTargetLocator, BenchmarkFinalArtifactKindContextBundle, BenchmarkFinalArtifactKindReadinessSummary, BenchmarkFinalArtifactKindTaskOutput:
	default:
		return fmt.Errorf("kind %q is unsupported", c.Kind)
	}
	if err := validateEvalRelativePath(c.Path); err != nil {
		return fmt.Errorf("path: %w", err)
	}
	switch c.Format {
	case BenchmarkFinalArtifactFormatText, BenchmarkFinalArtifactFormatJSON:
	default:
		return fmt.Errorf("format %q is unsupported", c.Format)
	}
	if err := c.Normalization.validate(c.Format); err != nil {
		return fmt.Errorf("normalization: %w", err)
	}
	if len(c.Assertions) == 0 {
		return errors.New("at least one assertion is required")
	}
	for idx, assertion := range c.Assertions {
		if err := assertion.validate(); err != nil {
			return fmt.Errorf("assert[%d]: %w", idx, err)
		}
	}
	return nil
}

func (c BenchmarkFinalArtifactContract) Validate() error {
	return c.validate()
}

func (n BenchmarkFinalArtifactNormalization) validate(format BenchmarkFinalArtifactFormat) error {
	switch n.Mode {
	case BenchmarkFinalArtifactNormalizationModeTextTrimmed, BenchmarkFinalArtifactNormalizationModeTextLines:
		if format != BenchmarkFinalArtifactFormatText {
			return fmt.Errorf("%q requires format %q", n.Mode, BenchmarkFinalArtifactFormatText)
		}
		if len(n.JSONPaths) > 0 {
			return errors.New("text normalization does not allow jsonPaths")
		}
	case BenchmarkFinalArtifactNormalizationModeJSONExact:
		if format != BenchmarkFinalArtifactFormatJSON {
			return fmt.Errorf("%q requires format %q", n.Mode, BenchmarkFinalArtifactFormatJSON)
		}
		if len(n.JSONPaths) > 0 {
			return errors.New("json_exact does not allow jsonPaths")
		}
	case BenchmarkFinalArtifactNormalizationModeJSONFields:
		if format != BenchmarkFinalArtifactFormatJSON {
			return fmt.Errorf("%q requires format %q", n.Mode, BenchmarkFinalArtifactFormatJSON)
		}
		if len(n.JSONPaths) == 0 {
			return errors.New("json_fields requires at least one jsonPath")
		}
		for idx, item := range n.JSONPaths {
			if strings.TrimSpace(item) == "" {
				return fmt.Errorf("jsonPaths[%d]: is required", idx)
			}
		}
	default:
		return fmt.Errorf("mode %q is unsupported", n.Mode)
	}
	return nil
}

func (a BenchmarkFinalArtifactAssertion) validate() error {
	switch a.Kind {
	case EvalAssertionKindContains:
		if a.Contains == "" {
			return fmt.Errorf("contains is required for %q", a.Kind)
		}
		if a.Path != "" || a.Equals != nil {
			return fmt.Errorf("path/equals must be empty for %q", a.Kind)
		}
	case EvalAssertionKindJSONFieldPresent:
		if strings.TrimSpace(a.Path) == "" {
			return fmt.Errorf("path is required for %q", a.Kind)
		}
		if a.Contains != "" || a.Equals != nil {
			return fmt.Errorf("contains/equals must be empty for %q", a.Kind)
		}
	case EvalAssertionKindJSONFieldEquals:
		if strings.TrimSpace(a.Path) == "" {
			return fmt.Errorf("path is required for %q", a.Kind)
		}
		if a.Contains != "" {
			return fmt.Errorf("contains must be empty for %q", a.Kind)
		}
	default:
		return fmt.Errorf("unsupported kind %q", a.Kind)
	}
	return nil
}

func BenchmarkMethodologyFromSuite(suite BenchmarkSuiteDefinition) BenchmarkMethodologySnapshot {
	snapshot := BenchmarkMethodologySnapshot{
		SuiteSchemaVersion: suite.SchemaVersion,
		Boundary:           suite.Boundary,
		CountedInputs:      append([]BenchmarkCountedInputDefinition(nil), suite.CountedInputs...),
	}
	if suite.Task.FinalArtifact != nil {
		current := *suite.Task.FinalArtifact
		snapshot.TaskFinalArtifact = &current
	}
	snapshot.LaneFinalArtifacts = make([]BenchmarkLaneFinalArtifactSnapshot, 0, len(suite.Lanes))
	for _, lane := range suite.Lanes {
		if lane.FinalArtifact == nil {
			continue
		}
		snapshot.LaneFinalArtifacts = append(snapshot.LaneFinalArtifacts, BenchmarkLaneFinalArtifactSnapshot{
			Lane:     lane.Name,
			Contract: *lane.FinalArtifact,
		})
	}
	return NormalizeBenchmarkMethodologySnapshot(snapshot)
}

func LoadBenchmarkSuite(path string) (BenchmarkSuiteDefinition, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return BenchmarkSuiteDefinition{}, err
	}

	var suite BenchmarkSuiteDefinition
	if err := json.Unmarshal(content, &suite); err != nil {
		return BenchmarkSuiteDefinition{}, fmt.Errorf("decode benchmark suite %s: %w", path, err)
	}
	if err := suite.Validate(); err != nil {
		return BenchmarkSuiteDefinition{}, fmt.Errorf("validate benchmark suite %s: %w", path, err)
	}
	return suite, nil
}

func LoadBenchmarkSuites(dir string) ([]BenchmarkSuiteDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var suites []BenchmarkSuiteDefinition
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		suite, err := LoadBenchmarkSuite(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		suites = append(suites, suite)
	}

	sort.Slice(suites, func(i, j int) bool {
		return suites[i].ID < suites[j].ID
	})
	seen := make(map[string]struct{}, len(suites))
	for _, suite := range suites {
		if _, exists := seen[suite.ID]; exists {
			return nil, fmt.Errorf("duplicate benchmark suite id %q", suite.ID)
		}
		seen[suite.ID] = struct{}{}
	}
	return suites, nil
}

func ValidateBenchmarkFixtureReferences(suites []BenchmarkSuiteDefinition, fixturesRoot string) error {
	for _, suite := range suites {
		fixturePath := filepath.Join(fixturesRoot, filepath.FromSlash(suite.Fixture.Path))
		info, err := os.Stat(fixturePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("suite %q fixture path %q does not exist", suite.ID, suite.Fixture.Path)
			}
			return fmt.Errorf("suite %q fixture path %q: %w", suite.ID, suite.Fixture.Path, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("suite %q fixture path %q is not a directory", suite.ID, suite.Fixture.Path)
		}
		if !strings.Contains(suite.Fixture.Path, suite.Fixture.ID) {
			return fmt.Errorf("suite %q fixture path %q must contain fixture id %q", suite.ID, suite.Fixture.Path, suite.Fixture.ID)
		}
		if !strings.Contains(suite.Fixture.Path, suite.Fixture.Version) {
			return fmt.Errorf("suite %q fixture path %q must contain fixture version %q", suite.ID, suite.Fixture.Path, suite.Fixture.Version)
		}
	}
	return nil
}
