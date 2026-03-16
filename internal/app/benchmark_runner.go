package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
)

type BenchmarkSuiteRequest struct {
	SuiteID      string
	SuitePath    string
	SuitesDir    string
	FixturesRoot string
}

type BenchmarkRunRequest struct {
	SuiteID       string
	SuitePath     string
	SuitesDir     string
	FixturesRoot  string
	WorkspaceRoot string
}

type BenchmarkCommandInvocation struct {
	Args       []string
	WorkingDir string
}

type BenchmarkCommandExecutionResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type BenchmarkCommandExecutor func(context.Context, BenchmarkCommandInvocation) (BenchmarkCommandExecutionResult, error)

type BenchmarkToolInvocation struct {
	Name       string
	Arguments  map[string]any
	WorkingDir string
}

type BenchmarkToolExecutionResult struct {
	Payload any
}

type BenchmarkToolExecutor func(context.Context, BenchmarkToolInvocation) (BenchmarkToolExecutionResult, error)

type BenchmarkRunner struct {
	LoadSuiteFile             func(string) (repository.BenchmarkSuiteDefinition, error)
	LoadSuiteFiles            func(string) ([]repository.BenchmarkSuiteDefinition, error)
	ValidateFixtureReferences func([]repository.BenchmarkSuiteDefinition, string) error
	RunCommand                BenchmarkCommandExecutor
	RunTool                   BenchmarkToolExecutor
	MkdirTemp                 func(string, string) (string, error)
	CopyTree                  func(string, string) error
	GitInit                   func(context.Context, string) error
	Now                       func() time.Time
}

func NewBenchmarkRunner() BenchmarkRunner {
	return BenchmarkRunner{
		LoadSuiteFile:             repository.LoadBenchmarkSuite,
		LoadSuiteFiles:            repository.LoadBenchmarkSuites,
		ValidateFixtureReferences: repository.ValidateBenchmarkFixtureReferences,
		MkdirTemp:                 os.MkdirTemp,
		CopyTree:                  copyEvalTree,
		GitInit: func(ctx context.Context, dir string) error {
			cmd := exec.CommandContext(ctx, "git", "init")
			cmd.Dir = dir
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("git init %s: %w (%s)", dir, err, string(output))
			}
			return nil
		},
		Now: time.Now,
	}
}

func (r BenchmarkRunner) LoadSuite(request BenchmarkSuiteRequest) (repository.BenchmarkSuiteDefinition, error) {
	r = r.withDefaults()

	if request.SuitePath != "" {
		if request.SuiteID != "" || request.SuitesDir != "" {
			return repository.BenchmarkSuiteDefinition{}, errors.New("suitePath cannot be combined with suiteID or suitesDir")
		}
		suite, err := r.LoadSuiteFile(request.SuitePath)
		if err != nil {
			return repository.BenchmarkSuiteDefinition{}, err
		}
		if err := r.ValidateFixtureReferences([]repository.BenchmarkSuiteDefinition{suite}, request.FixturesRoot); err != nil {
			return repository.BenchmarkSuiteDefinition{}, err
		}
		return suite, nil
	}

	if request.SuiteID == "" {
		return repository.BenchmarkSuiteDefinition{}, errors.New("suiteID or suitePath is required")
	}
	if request.SuitesDir == "" {
		return repository.BenchmarkSuiteDefinition{}, errors.New("suitesDir is required when selecting by suiteID")
	}

	suites, err := r.LoadSuiteFiles(request.SuitesDir)
	if err != nil {
		return repository.BenchmarkSuiteDefinition{}, err
	}
	if err := r.ValidateFixtureReferences(suites, request.FixturesRoot); err != nil {
		return repository.BenchmarkSuiteDefinition{}, err
	}
	for _, suite := range suites {
		if suite.ID == request.SuiteID {
			return suite, nil
		}
	}
	return repository.BenchmarkSuiteDefinition{}, fmt.Errorf("benchmark suite %q not found", request.SuiteID)
}

func (r BenchmarkRunner) LoadSuites(request BenchmarkSuiteRequest) ([]repository.BenchmarkSuiteDefinition, error) {
	r = r.withDefaults()
	if request.SuitesDir == "" {
		return nil, errors.New("suitesDir is required")
	}
	suites, err := r.LoadSuiteFiles(request.SuitesDir)
	if err != nil {
		return nil, err
	}
	if err := r.ValidateFixtureReferences(suites, request.FixturesRoot); err != nil {
		return nil, err
	}
	return suites, nil
}

func (r BenchmarkRunner) Run(ctx context.Context, request BenchmarkRunRequest) (repository.BenchmarkRunResult, error) {
	r = r.withDefaults()
	if err := validateBenchmarkRunRequest(request); err != nil {
		return repository.BenchmarkRunResult{}, err
	}

	suite, err := r.LoadSuite(BenchmarkSuiteRequest{
		SuiteID:      request.SuiteID,
		SuitePath:    request.SuitePath,
		SuitesDir:    request.SuitesDir,
		FixturesRoot: request.FixturesRoot,
	})
	if err != nil {
		return repository.BenchmarkRunResult{}, err
	}

	workspaceRoot, err := r.prepareWorkspace(ctx, request, suite)
	if err != nil {
		return repository.BenchmarkRunResult{}, err
	}
	if err := r.bootstrapTreatmentWorkspace(ctx, workspaceRoot); err != nil {
		return repository.BenchmarkRunResult{}, err
	}

	result := repository.BenchmarkRunResult{
		SchemaVersion: repository.BenchmarkSuiteSchemaV1,
		SuiteID:       suite.ID,
		SuiteVersion:  suite.Version,
		FixtureID:     suite.Fixture.ID,
		FixturePath:   suite.Fixture.Path,
		WorkspacePath: workspaceRoot,
		StartedAt:     r.Now().UTC(),
	}

	for _, arm := range suite.Arms {
		armResult, err := r.runArm(ctx, workspaceRoot, suite, arm)
		if err != nil {
			result.FinishedAt = r.Now().UTC()
			return result, err
		}
		result.Arms = append(result.Arms, armResult)
	}
	result.FinishedAt = r.Now().UTC()
	return result, nil
}

func validateBenchmarkRunRequest(request BenchmarkRunRequest) error {
	hasID := strings.TrimSpace(request.SuiteID) != ""
	hasPath := strings.TrimSpace(request.SuitePath) != ""
	if hasID == hasPath {
		return errors.New("benchmark runner requires exactly one of suite ID or suite path")
	}
	if strings.TrimSpace(request.FixturesRoot) == "" {
		return errors.New("fixtures root is required")
	}
	if hasID && strings.TrimSpace(request.SuitesDir) == "" {
		return errors.New("suites directory is required when selecting by suite ID")
	}
	return nil
}

func (r BenchmarkRunner) prepareWorkspace(ctx context.Context, request BenchmarkRunRequest, suite repository.BenchmarkSuiteDefinition) (string, error) {
	tempRootParent := request.WorkspaceRoot
	tempRoot, err := r.MkdirTemp(tempRootParent, "optimusctx-benchmark-")
	if err != nil {
		return "", fmt.Errorf("create benchmark workspace: %w", err)
	}

	workspaceDirName := suite.Fixture.WorkspaceDir
	if workspaceDirName == "" {
		workspaceDirName = "workspace"
	}
	workspaceRoot := filepath.Join(tempRoot, filepath.FromSlash(workspaceDirName))
	fixtureRoot := filepath.Join(request.FixturesRoot, filepath.FromSlash(suite.Fixture.Path))
	if err := r.CopyTree(fixtureRoot, workspaceRoot); err != nil {
		return "", fmt.Errorf("materialize benchmark fixture %q: %w", suite.Fixture.ID, err)
	}
	if err := r.GitInit(ctx, workspaceRoot); err != nil {
		return "", err
	}
	return workspaceRoot, nil
}

func (r BenchmarkRunner) bootstrapTreatmentWorkspace(ctx context.Context, workspaceRoot string) error {
	if r.RunCommand == nil {
		return nil
	}
	for _, args := range [][]string{{"init"}, {"refresh"}} {
		execution, err := r.RunCommand(ctx, BenchmarkCommandInvocation{
			Args:       args,
			WorkingDir: workspaceRoot,
		})
		if err != nil {
			return fmt.Errorf("bootstrap benchmark workspace with %q: %w", strings.Join(args, " "), err)
		}
		if execution.ExitCode != 0 {
			return fmt.Errorf("bootstrap benchmark workspace with %q: exit code %d", strings.Join(args, " "), execution.ExitCode)
		}
	}
	return nil
}

func (r BenchmarkRunner) runArm(ctx context.Context, workspaceRoot string, suite repository.BenchmarkSuiteDefinition, arm repository.BenchmarkArmDefinition) (repository.BenchmarkArmRunResult, error) {
	state := newBenchmarkArmState(suite, r.Now)
	result := repository.BenchmarkArmRunResult{
		Kind:      arm.Kind,
		Name:      arm.Name,
		StartedAt: r.Now().UTC(),
	}

	for _, step := range arm.Steps {
		var err error
		switch arm.Kind {
		case repository.BenchmarkArmKindBaseline:
			err = r.executeBaselineStep(workspaceRoot, step, state)
		case repository.BenchmarkArmKindOptimusCtx:
			err = r.executeTreatmentStep(ctx, workspaceRoot, suite, step, state)
		default:
			err = fmt.Errorf("unsupported benchmark arm %q", arm.Kind)
		}
		if err != nil {
			return repository.BenchmarkArmRunResult{}, fmt.Errorf("benchmark arm %q step %q: %w", arm.Kind, step.ID, err)
		}
	}

	result.FinishedAt = r.Now().UTC()
	result.LaneResults = state.results()
	return result, nil
}

func (r BenchmarkRunner) executeBaselineStep(workspaceRoot string, step repository.BenchmarkStep, state *benchmarkArmState) error {
	startedAt := state.now()
	state.startLane(step.Lane, startedAt)

	action := step.Baseline
	switch action.Kind {
	case repository.BenchmarkBaselineActionListTree, repository.BenchmarkBaselineActionGitListFiles:
		if _, err := walkBenchmarkFiles(workspaceRoot, action.Path); err != nil {
			return err
		}
		state.recordBroadSearch(step.Lane)
	case repository.BenchmarkBaselineActionSearchText, repository.BenchmarkBaselineActionGitGrep:
		matches, err := searchBenchmarkQuery(workspaceRoot, action.Path, action.Query)
		if err != nil {
			return err
		}
		state.recordBroadSearch(step.Lane)
		for _, match := range matches {
			state.addArtifact(step.Lane, match)
		}
	case repository.BenchmarkBaselineActionReadFileSlice:
		content, err := readBenchmarkFileSlice(filepath.Join(workspaceRoot, filepath.FromSlash(action.Path)), action.StartLine, action.EndLine)
		if err != nil {
			return err
		}
		state.recordFileRead(step.Lane, action.Path, int64(len(content)))
	case repository.BenchmarkBaselineActionMarkLaneComplete:
		state.markLaneComplete(step.Lane, action.Marker, startedAt)
	default:
		return fmt.Errorf("unsupported baseline action %q", action.Kind)
	}
	return nil
}

func (r BenchmarkRunner) executeTreatmentStep(ctx context.Context, workspaceRoot string, suite repository.BenchmarkSuiteDefinition, step repository.BenchmarkStep, state *benchmarkArmState) error {
	if step.Treatment == nil {
		return errors.New("treatment step is missing action")
	}
	startedAt := state.now()
	state.startLane(step.Lane, startedAt)

	switch step.Treatment.Surface {
	case repository.BenchmarkTreatmentSurfaceCLI:
		if r.RunCommand == nil {
			return errors.New("benchmark command executor is not configured")
		}
		invocation := BenchmarkCommandInvocation{
			Args:       buildBenchmarkCLIArgs(step.Treatment),
			WorkingDir: workspaceRoot,
		}
		execution, err := r.RunCommand(ctx, invocation)
		if err != nil {
			return err
		}
		if execution.ExitCode != 0 {
			return fmt.Errorf("cli command %q exited with %d", strings.Join(invocation.Args, " "), execution.ExitCode)
		}
		state.recordTargetedLookup(step.Lane)
		if strings.TrimSpace(execution.Stdout) != "" {
			state.addArtifact(step.Lane, "stdout")
		}
		if strings.TrimSpace(execution.Stderr) != "" {
			state.addArtifact(step.Lane, "stderr")
		}
	case repository.BenchmarkTreatmentSurfaceMCP:
		if r.RunTool == nil {
			return errors.New("benchmark MCP tool executor is not configured")
		}
		invocation := BenchmarkToolInvocation{
			Name:       step.Treatment.Tool,
			Arguments:  buildBenchmarkToolArguments(suite, workspaceRoot, step.Treatment.Tool, state),
			WorkingDir: workspaceRoot,
		}
		execution, err := r.RunTool(ctx, invocation)
		if err != nil {
			return err
		}
		if err := state.applyToolResult(step.Lane, step.Treatment.Tool, suite.Task, execution.Payload, startedAt); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported treatment surface %q", step.Treatment.Surface)
	}
	return nil
}

func buildBenchmarkCLIArgs(action *repository.BenchmarkTreatmentAction) []string {
	var args []string
	switch action.Command {
	case repository.EvalCommandInit:
		args = append(args, "init")
	case repository.EvalCommandRefresh:
		args = append(args, "refresh")
	case repository.EvalCommandDoctor:
		args = append(args, "doctor")
	case repository.EvalCommandPackExport:
		args = append(args, "pack", "export")
	}
	args = append(args, action.Args...)
	return args
}

func buildBenchmarkToolArguments(suite repository.BenchmarkSuiteDefinition, workspaceRoot string, tool string, state *benchmarkArmState) map[string]any {
	args := map[string]any{"startPath": workspaceRoot}
	switch tool {
	case "optimusctx.symbol_lookup":
		args["name"] = suite.Task.TargetSymbol
	case "optimusctx.targeted_context":
		if state.targetStableKey != "" {
			args["stableKey"] = state.targetStableKey
		} else {
			args["path"] = suite.Task.TargetPath
			args["startLine"] = 1
			args["endLine"] = 1
		}
		args["beforeLines"] = 0
		args["afterLines"] = 40
	}
	return args
}

func walkBenchmarkFiles(workspaceRoot string, relativeRoot string) ([]string, error) {
	root := workspaceRoot
	if strings.TrimSpace(relativeRoot) != "" {
		root = filepath.Join(workspaceRoot, filepath.FromSlash(relativeRoot))
	}
	var matches []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".optimusctx" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(workspaceRoot, path)
		if err != nil {
			return err
		}
		matches = append(matches, filepath.ToSlash(rel))
		return nil
	})
	return matches, err
}

func searchBenchmarkQuery(workspaceRoot string, relativeRoot string, query string) ([]string, error) {
	files, err := walkBenchmarkFiles(workspaceRoot, relativeRoot)
	if err != nil {
		return nil, err
	}
	var matches []string
	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(file)))
		if err != nil {
			return nil, err
		}
		if strings.Contains(string(content), query) {
			matches = append(matches, file)
		}
	}
	return matches, nil
}

func readBenchmarkFileSlice(path string, startLine int, endLine int) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	if startLine > len(lines) {
		return "", nil
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	return strings.Join(lines[startLine-1:endLine], "\n"), nil
}

func (r BenchmarkRunner) withDefaults() BenchmarkRunner {
	defaults := NewBenchmarkRunner()
	if r.LoadSuiteFile == nil {
		r.LoadSuiteFile = defaults.LoadSuiteFile
	}
	if r.LoadSuiteFiles == nil {
		r.LoadSuiteFiles = defaults.LoadSuiteFiles
	}
	if r.ValidateFixtureReferences == nil {
		r.ValidateFixtureReferences = defaults.ValidateFixtureReferences
	}
	if r.MkdirTemp == nil {
		r.MkdirTemp = defaults.MkdirTemp
	}
	if r.CopyTree == nil {
		r.CopyTree = defaults.CopyTree
	}
	if r.GitInit == nil {
		r.GitInit = defaults.GitInit
	}
	if r.Now == nil {
		r.Now = defaults.Now
	}
	return r
}

type benchmarkLaneState struct {
	definition repository.BenchmarkLaneDefinition
	startedAt  time.Time
	finishedAt time.Time
	success    bool
	effort     repository.BenchmarkLaneEffort
}

type benchmarkArmState struct {
	lanes           map[repository.BenchmarkLane]*benchmarkLaneState
	order           []repository.BenchmarkLane
	nowFn           func() time.Time
	targetStableKey string
}

func newBenchmarkArmState(suite repository.BenchmarkSuiteDefinition, nowFn func() time.Time) *benchmarkArmState {
	state := &benchmarkArmState{
		lanes: make(map[repository.BenchmarkLane]*benchmarkLaneState, len(suite.Lanes)),
		nowFn: nowFn,
	}
	for _, lane := range suite.Lanes {
		lane.StartMarker = lane.StartMarkerName()
		lane.SuccessMarker = lane.SuccessMarkerName()
		state.lanes[lane.Name] = &benchmarkLaneState{definition: lane}
		state.order = append(state.order, lane.Name)
	}
	return state
}

func (s *benchmarkArmState) now() time.Time {
	return s.nowFn().UTC()
}

func (s *benchmarkArmState) startLane(lane repository.BenchmarkLane, startedAt time.Time) {
	current := s.lanes[lane]
	if current.startedAt.IsZero() {
		current.startedAt = startedAt.UTC()
	}
}

func (s *benchmarkArmState) markLaneComplete(lane repository.BenchmarkLane, marker string, finishedAt time.Time) {
	current := s.lanes[lane]
	if marker != current.definition.StopCondition.Marker {
		return
	}
	current.finishedAt = finishedAt.UTC()
	current.success = true
}

func (s *benchmarkArmState) recordBroadSearch(lane repository.BenchmarkLane) {
	current := s.lanes[lane]
	current.effort.ActionCount++
	current.effort.BroadSearchActions++
}

func (s *benchmarkArmState) recordTargetedLookup(lane repository.BenchmarkLane) {
	current := s.lanes[lane]
	current.effort.ActionCount++
	current.effort.TargetedLookupActions++
}

func (s *benchmarkArmState) recordFileRead(lane repository.BenchmarkLane, artifact string, bytesRead int64) {
	current := s.lanes[lane]
	current.effort.ActionCount++
	current.effort.FileReadActions++
	current.effort.BytesRead += bytesRead
	s.addArtifact(lane, artifact)
}

func (s *benchmarkArmState) addArtifact(lane repository.BenchmarkLane, artifact string) {
	current := s.lanes[lane]
	if artifact == "" {
		return
	}
	if slices.Contains(current.effort.ConsultedArtifacts, artifact) {
		return
	}
	current.effort.ConsultedArtifacts = append(current.effort.ConsultedArtifacts, artifact)
}

func (s *benchmarkArmState) applyToolResult(lane repository.BenchmarkLane, tool string, task repository.BenchmarkTaskDefinition, payload any, finishedAt time.Time) error {
	current := s.lanes[lane]
	switch tool {
	case "optimusctx.repository_map":
		var result repository.RepositoryMap
		if err := decodeBenchmarkPayload(payload, &result); err != nil {
			return err
		}
		current.effort.ActionCount++
		current.effort.BroadSearchActions++
		for _, directory := range result.Directories {
			for _, file := range directory.Files {
				s.addArtifact(lane, file.Path)
			}
		}
	case "optimusctx.symbol_lookup":
		var result repository.SymbolLookupResult
		if err := decodeBenchmarkPayload(payload, &result); err != nil {
			return err
		}
		current.effort.ActionCount++
		current.effort.TargetedLookupActions++
		for _, match := range result.Matches {
			s.addArtifact(lane, match.Path)
			if match.Path == task.TargetPath && match.Name == task.TargetSymbol {
				s.targetStableKey = match.StableKey
				current.finishedAt = finishedAt.UTC()
				current.success = true
			}
		}
	case "optimusctx.targeted_context":
		var result repository.TargetedContextResult
		if err := decodeBenchmarkPayload(payload, &result); err != nil {
			return err
		}
		current.effort.ActionCount++
		current.effort.FileReadActions++
		current.effort.BytesRead += int64(len(strings.Join(result.Source, "\n")))
		s.addArtifact(lane, result.Path)
		if result.Path == task.TargetPath {
			current.finishedAt = finishedAt.UTC()
			current.success = true
		}
	default:
		current.effort.ActionCount++
		current.effort.TargetedLookupActions++
	}
	return nil
}

func (s *benchmarkArmState) results() []repository.BenchmarkLaneRunResult {
	results := make([]repository.BenchmarkLaneRunResult, 0, len(s.order))
	for _, name := range s.order {
		current := s.lanes[name]
		if current.finishedAt.IsZero() {
			current.finishedAt = current.startedAt
		}
		results = append(results, repository.BenchmarkLaneRunResult{
			Lane:          name,
			StartMarker:   current.definition.StartMarkerName(),
			SuccessMarker: current.definition.SuccessMarkerName(),
			StopMarker:    current.definition.StopCondition.Marker,
			StartedAt:     current.startedAt,
			FinishedAt:    current.finishedAt,
			Elapsed:       current.finishedAt.Sub(current.startedAt),
			Success:       current.success,
			Effort:        current.effort,
		})
	}
	return results
}

func decodeBenchmarkPayload(payload any, target any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
