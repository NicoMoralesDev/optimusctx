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
	StopCondition BenchmarkStopCondition `json:"stopCondition"`
	Metrics       []BenchmarkMetric      `json:"metrics"`
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
	StartedAt   time.Time                `json:"startedAt"`
	FinishedAt  time.Time                `json:"finishedAt"`
	LaneResults []BenchmarkLaneRunResult `json:"laneResults"`
}

type BenchmarkLaneRunResult struct {
	Lane          BenchmarkLane       `json:"lane"`
	StartMarker   string              `json:"startMarker"`
	SuccessMarker string              `json:"successMarker"`
	StopMarker    string              `json:"stopMarker"`
	StartedAt     time.Time           `json:"startedAt"`
	FinishedAt    time.Time           `json:"finishedAt"`
	Elapsed       time.Duration       `json:"elapsed"`
	Success       bool                `json:"success"`
	Effort        BenchmarkLaneEffort `json:"effort"`
}

type BenchmarkLaneEffort struct {
	ActionCount           int64    `json:"actionCount"`
	BroadSearchActions    int64    `json:"broadSearchActions"`
	TargetedLookupActions int64    `json:"targetedLookupActions"`
	FileReadActions       int64    `json:"fileReadActions"`
	BytesRead             int64    `json:"bytesRead"`
	ConsultedArtifacts    []string `json:"consultedArtifacts,omitempty"`
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
		seen[lane.Name] = lane
	}
	return seen, nil
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
