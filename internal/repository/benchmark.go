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
const BenchmarkEvidenceBundleSchemaV1 = "optimusctx/benchmark-evidence@v1"

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
	Lane           BenchmarkLane                  `json:"lane"`
	StartMarker    string                         `json:"startMarker"`
	SuccessMarker  string                         `json:"successMarker"`
	StopMarker     string                         `json:"stopMarker"`
	SetupAppliedAt time.Time                      `json:"setupAppliedAt,omitempty"`
	StartedAt      time.Time                      `json:"startedAt"`
	FinishedAt     time.Time                      `json:"finishedAt"`
	ElapsedMS      int64                          `json:"elapsedMs"`
	Success        bool                           `json:"success"`
	EvidencePaths  []string                       `json:"evidencePaths,omitempty"`
	Effort         BenchmarkLaneEffort            `json:"effort"`
	Attribution    []BenchmarkArtifactConsumption `json:"attribution,omitempty"`
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
	SchemaVersion string                    `json:"schemaVersion"`
	ID            string                    `json:"id"`
	Version       string                    `json:"version"`
	Name          string                    `json:"name"`
	Description   string                    `json:"description,omitempty"`
	Fixture       EvalFixtureRef            `json:"fixture"`
	Task          BenchmarkTaskDefinition   `json:"task"`
	Lanes         []BenchmarkLaneDefinition `json:"lanes"`
	Arms          []BenchmarkArmDefinition  `json:"arms"`
}

type BenchmarkTaskDefinition struct {
	ID                 string   `json:"id"`
	Prompt             string   `json:"prompt"`
	TargetPath         string   `json:"targetPath,omitempty"`
	TargetSymbol       string   `json:"targetSymbol,omitempty"`
	ContextPaths       []string `json:"contextPaths,omitempty"`
	CompletionArtifact string   `json:"completionArtifact,omitempty"`
}

type BenchmarkLaneDefinition struct {
	Name          BenchmarkLane          `json:"name"`
	Description   string                 `json:"description,omitempty"`
	StartMarker   string                 `json:"startMarker,omitempty"`
	SuccessMarker string                 `json:"successMarker,omitempty"`
	Setup         []EvalSetupAction      `json:"setup,omitempty"`
	Assertions    []BenchmarkAssertion   `json:"assert,omitempty"`
	StopCondition BenchmarkStopCondition `json:"stopCondition"`
	Metrics       []BenchmarkMetric      `json:"metrics"`
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
	Lane           BenchmarkLane                  `json:"lane"`
	StartMarker    string                         `json:"startMarker"`
	SuccessMarker  string                         `json:"successMarker"`
	StopMarker     string                         `json:"stopMarker"`
	SetupAppliedAt time.Time                      `json:"setupAppliedAt,omitempty"`
	StartedAt      time.Time                      `json:"startedAt"`
	FinishedAt     time.Time                      `json:"finishedAt"`
	Elapsed        time.Duration                  `json:"elapsed"`
	Success        bool                           `json:"success"`
	Setup          []EvalSetupAction              `json:"setup,omitempty"`
	Assertions     []BenchmarkAssertion           `json:"assert,omitempty"`
	EvidencePaths  []string                       `json:"evidencePaths,omitempty"`
	Effort         BenchmarkLaneEffort            `json:"effort"`
	Attribution    []BenchmarkArtifactConsumption `json:"attribution,omitempty"`
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

func NormalizeBenchmarkEvidenceBundle(bundle BenchmarkEvidenceBundle) BenchmarkEvidenceBundle {
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
	if s.SchemaVersion != BenchmarkSuiteSchemaV1 {
		return fmt.Errorf("schemaVersion must be %q", BenchmarkSuiteSchemaV1)
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
	return validateBenchmarkArms(s.Arms, laneDefs)
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
		if err := validateEvalRelativePath(t.CompletionArtifact); err != nil {
			return fmt.Errorf("completionArtifact: %w", err)
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

func validateBenchmarkArms(arms []BenchmarkArmDefinition, lanes map[BenchmarkLane]BenchmarkLaneDefinition) error {
	seenKinds := map[BenchmarkArmKind]struct{}{}
	for idx, arm := range arms {
		switch arm.Kind {
		case BenchmarkArmKindBaseline, BenchmarkArmKindOptimusCtx:
		default:
			return fmt.Errorf("arms[%d]: unsupported kind %q", idx, arm.Kind)
		}
		if _, exists := seenKinds[arm.Kind]; exists {
			return fmt.Errorf("duplicate arm kind %q", arm.Kind)
		}
		seenKinds[arm.Kind] = struct{}{}
		if strings.TrimSpace(arm.Name) == "" {
			return fmt.Errorf("arms[%d]: name is required", idx)
		}
		if len(arm.Steps) == 0 {
			return fmt.Errorf("arms[%d]: at least one step is required", idx)
		}
		if err := validateBenchmarkArmSteps(idx, arm, lanes); err != nil {
			return err
		}
	}
	if _, ok := seenKinds[BenchmarkArmKindBaseline]; !ok {
		return errors.New("baseline arm is required")
	}
	if _, ok := seenKinds[BenchmarkArmKindOptimusCtx]; !ok {
		return errors.New("optimusctx arm is required")
	}
	return nil
}

func validateBenchmarkArmSteps(armIdx int, arm BenchmarkArmDefinition, lanes map[BenchmarkLane]BenchmarkLaneDefinition) error {
	seenSteps := make(map[string]struct{}, len(arm.Steps))
	markers := make(map[BenchmarkLane]map[string]struct{})
	for stepIdx, step := range arm.Steps {
		if strings.TrimSpace(step.ID) == "" {
			return fmt.Errorf("arms[%d].steps[%d]: id is required", armIdx, stepIdx)
		}
		if _, exists := seenSteps[step.ID]; exists {
			return fmt.Errorf("arms[%d]: duplicate step id %q", armIdx, step.ID)
		}
		seenSteps[step.ID] = struct{}{}
		if strings.TrimSpace(step.Name) == "" {
			return fmt.Errorf("arms[%d].steps[%d]: name is required", armIdx, stepIdx)
		}
		if _, ok := lanes[step.Lane]; !ok {
			return fmt.Errorf("arms[%d].steps[%d]: unknown lane %q", armIdx, stepIdx, step.Lane)
		}

		switch arm.Kind {
		case BenchmarkArmKindBaseline:
			if step.Baseline == nil || step.Treatment != nil {
				return fmt.Errorf("arms[%d].steps[%d]: baseline arm must use only baseline actions", armIdx, stepIdx)
			}
			if err := step.Baseline.validate(); err != nil {
				return fmt.Errorf("arms[%d].steps[%d].baseline: %w", armIdx, stepIdx, err)
			}
			if step.Baseline.Marker != "" {
				if markers[step.Lane] == nil {
					markers[step.Lane] = make(map[string]struct{})
				}
				markers[step.Lane][step.Baseline.Marker] = struct{}{}
			}
		case BenchmarkArmKindOptimusCtx:
			if step.Treatment == nil || step.Baseline != nil {
				return fmt.Errorf("arms[%d].steps[%d]: optimusctx arm must use only treatment actions", armIdx, stepIdx)
			}
			if err := step.Treatment.validate(); err != nil {
				return fmt.Errorf("arms[%d].steps[%d].treatment: %w", armIdx, stepIdx, err)
			}
		}
	}

	for laneName, laneDef := range lanes {
		if !armHasLane(arm, laneName) {
			return fmt.Errorf("arms[%d]: lane %q is missing", armIdx, laneName)
		}
		if arm.Kind == BenchmarkArmKindBaseline && laneDef.StopCondition.Kind == BenchmarkStopConditionKindMarker {
			if _, ok := markers[laneName][laneDef.StopCondition.Marker]; !ok {
				return fmt.Errorf("arms[%d]: lane %q does not emit stop marker %q", armIdx, laneName, laneDef.StopCondition.Marker)
			}
		}
	}
	return nil
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
