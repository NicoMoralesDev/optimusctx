package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

type EvalRunRequest struct {
	ScenarioID    string
	ScenarioPath  string
	ScenariosDir  string
	FixturesRoot  string
	WorkspaceRoot string
	StartPath     string
}

type EvalCommandInvocation struct {
	Args       []string
	WorkingDir string
}

type EvalCommandExecutionResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type EvalCommandExecutor func(context.Context, EvalCommandInvocation) (EvalCommandExecutionResult, error)

type EvalMCPSessionInvocation struct {
	WorkingDir string
	Session    repository.EvalMCPSession
}

type EvalMCPSessionResponse struct {
	RequestID int64
	Response  any
}

type EvalMCPSessionExecutionResult struct {
	Stdout    string
	Stderr    string
	Responses []EvalMCPSessionResponse
}

type EvalMCPSessionExecutor func(context.Context, EvalMCPSessionInvocation) (EvalMCPSessionExecutionResult, error)

type evalRefreshFailureControl struct {
	Stage   string
	Message string
}

type evalStepControls struct {
	RefreshFailure *evalRefreshFailureControl
}

var (
	evalRefreshFailureMu     sync.Mutex
	evalRefreshFailureActive *evalRefreshFailureControl
)

type EvalRunner struct {
	LoadScenario              func(string) (repository.EvalScenarioDefinition, error)
	LoadScenarios             func(string) ([]repository.EvalScenarioDefinition, error)
	ValidateFixtureReferences func([]repository.EvalScenarioDefinition, string) error
	RunCommand                EvalCommandExecutor
	RunMCPSession             EvalMCPSessionExecutor
	MkdirTemp                 func(string, string) (string, error)
	CopyTree                  func(string, string) error
	GitInit                   func(context.Context, string) error
	Now                       func() time.Time
}

func NewEvalRunner() EvalRunner {
	return EvalRunner{
		LoadScenario:              repository.LoadEvalScenario,
		LoadScenarios:             repository.LoadEvalScenarios,
		ValidateFixtureReferences: repository.ValidateEvalFixtureReferences,
		RunCommand: func(context.Context, EvalCommandInvocation) (EvalCommandExecutionResult, error) {
			return EvalCommandExecutionResult{}, errors.New("eval runner command executor is not configured")
		},
		RunMCPSession: func(context.Context, EvalMCPSessionInvocation) (EvalMCPSessionExecutionResult, error) {
			return EvalMCPSessionExecutionResult{}, errors.New("eval runner MCP session executor is not configured")
		},
		MkdirTemp: os.MkdirTemp,
		CopyTree:  copyEvalTree,
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

func (r EvalRunner) Run(ctx context.Context, request EvalRunRequest) (repository.EvalRunResult, error) {
	r = r.withDefaults()

	if err := validateEvalRunRequest(request); err != nil {
		return repository.EvalRunResult{}, err
	}

	scenario, err := r.loadScenario(request)
	if err != nil {
		return repository.EvalRunResult{}, err
	}
	if err := r.ValidateFixtureReferences([]repository.EvalScenarioDefinition{scenario}, request.FixturesRoot); err != nil {
		return repository.EvalRunResult{}, err
	}

	workspaceRoot, err := r.prepareWorkspace(ctx, request, scenario)
	if err != nil {
		return repository.EvalRunResult{}, err
	}

	runStartedAt := r.Now().UTC()
	runResult := repository.EvalRunResult{
		SchemaVersion: repository.EvalScenarioSchemaV1,
		ScenarioID:    scenario.ID,
		Scenario:      scenario,
		WorkspacePath: workspaceRoot,
		StartedAt:     runStartedAt,
	}

	artifactIndex := indexEvalArtifacts(scenario.Artifacts)
	seenArtifacts := make(map[string]repository.EvalArtifactResult, len(artifactIndex))

	for _, step := range scenario.Steps {
		controls, err := applyEvalSetupActions(workspaceRoot, step.Setup)
		if err != nil {
			runResult.FinishedAt = r.Now().UTC()
			runResult.Passed = false
			runResult.Artifacts = collectEvalArtifacts(scenario.Artifacts, seenArtifacts)
			return runResult, fmt.Errorf("scenario %q step %q: %w", scenario.ID, step.ID, err)
		}
		if err := prepareEvalStepArtifacts(workspaceRoot, step, artifactIndex); err != nil {
			runResult.FinishedAt = r.Now().UTC()
			runResult.Passed = false
			runResult.Artifacts = collectEvalArtifacts(scenario.Artifacts, seenArtifacts)
			return runResult, fmt.Errorf("scenario %q step %q: %w", scenario.ID, step.ID, err)
		}

		stepStartedAt := r.Now().UTC()
		stepResult, execErr := r.executeEvalStep(ctx, workspaceRoot, step, artifactIndex, controls)
		stepFinishedAt := r.Now().UTC()
		stepResult.Step = step
		stepResult.StartedAt = stepStartedAt
		stepResult.FinishedAt = stepFinishedAt
		execution := EvalCommandExecutionResult{
			Stdout:   stepResult.Stdout,
			Stderr:   stepResult.Stderr,
			ExitCode: stepResult.ExitCode,
		}

		stepArtifacts, artifactErr := captureEvalArtifacts(workspaceRoot, step, artifactIndex, execution)
		stepResult.Artifacts = stepArtifacts
		for _, artifact := range stepArtifacts {
			seenArtifacts[artifact.Ref.ID] = artifact
		}
		if artifactErr != nil {
			stepResult.Passed = false
			runResult.Steps = append(runResult.Steps, stepResult)
			runResult.FinishedAt = r.Now().UTC()
			runResult.Passed = false
			runResult.Artifacts = collectEvalArtifacts(scenario.Artifacts, seenArtifacts)
			return runResult, fmt.Errorf("scenario %q step %q: %w", scenario.ID, step.ID, artifactErr)
		}
		if assertionErr := evaluateEvalAssertions(step, stepArtifacts, artifactIndex, execution); assertionErr != nil {
			stepResult.Passed = false
			runResult.Steps = append(runResult.Steps, stepResult)
			runResult.FinishedAt = r.Now().UTC()
			runResult.Passed = false
			runResult.Artifacts = collectEvalArtifacts(scenario.Artifacts, seenArtifacts)
			return runResult, fmt.Errorf("scenario %q step %q: %w", scenario.ID, step.ID, assertionErr)
		}
		if execErr != nil {
			stepResult.Passed = false
			runResult.Steps = append(runResult.Steps, stepResult)
			runResult.FinishedAt = r.Now().UTC()
			runResult.Passed = false
			runResult.Artifacts = collectEvalArtifacts(scenario.Artifacts, seenArtifacts)
			return runResult, fmt.Errorf("scenario %q step %q: %w", scenario.ID, step.ID, execErr)
		}
		if step.Kind == repository.EvalStepKindCommand && stepResult.ExitCode != step.Expect.ExitCode {
			runResult.Steps = append(runResult.Steps, stepResult)
			runResult.FinishedAt = r.Now().UTC()
			runResult.Passed = false
			runResult.Artifacts = collectEvalArtifacts(scenario.Artifacts, seenArtifacts)
			return runResult, fmt.Errorf("scenario %q step %q failed: exit code %d, want %d", scenario.ID, step.ID, stepResult.ExitCode, step.Expect.ExitCode)
		}

		runResult.Steps = append(runResult.Steps, stepResult)
	}

	runResult.FinishedAt = r.Now().UTC()
	runResult.Passed = true
	runResult.Artifacts = collectEvalArtifacts(scenario.Artifacts, seenArtifacts)
	return runResult, nil
}

func applyEvalSetupActions(workspaceRoot string, actions []repository.EvalSetupAction) (evalStepControls, error) {
	var controls evalStepControls
	for _, action := range actions {
		switch action.Kind {
		case repository.EvalSetupActionWriteFile:
			path := filepath.Join(workspaceRoot, filepath.FromSlash(action.Path))
			if _, err := os.Stat(path); err == nil {
				return evalStepControls{}, fmt.Errorf("setup action %q requires missing path %q", action.Kind, action.Path)
			} else if !errors.Is(err, os.ErrNotExist) {
				return evalStepControls{}, fmt.Errorf("stat setup path %q: %w", action.Path, err)
			}
			if err := writeEvalSetupFile(path, action.Content); err != nil {
				return evalStepControls{}, fmt.Errorf("write setup file %q: %w", action.Path, err)
			}
		case repository.EvalSetupActionOverwriteFile:
			path := filepath.Join(workspaceRoot, filepath.FromSlash(action.Path))
			if err := writeEvalSetupFile(path, action.Content); err != nil {
				return evalStepControls{}, fmt.Errorf("overwrite setup file %q: %w", action.Path, err)
			}
		case repository.EvalSetupActionDeleteFile:
			path := filepath.Join(workspaceRoot, filepath.FromSlash(action.Path))
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return evalStepControls{}, fmt.Errorf("delete setup file %q: %w", action.Path, err)
			}
		case repository.EvalSetupActionSeedWatchStatus:
			if action.WatchStatus == nil {
				return evalStepControls{}, errors.New("seed_watch_status requires watchStatus")
			}
			if err := seedEvalWatchStatus(workspaceRoot, *action.WatchStatus); err != nil {
				return evalStepControls{}, err
			}
		case repository.EvalSetupActionSetRepositoryState:
			if action.RepositoryState == nil {
				return evalStepControls{}, errors.New("set_repository_state requires repositoryState")
			}
			if err := seedEvalRepositoryState(workspaceRoot, *action.RepositoryState); err != nil {
				return evalStepControls{}, err
			}
		case repository.EvalSetupActionInjectRefreshFailure:
			controls.RefreshFailure = &evalRefreshFailureControl{
				Stage:   action.FailureStage,
				Message: action.FailureMessage,
			}
		default:
			return evalStepControls{}, fmt.Errorf("unsupported setup action %q", action.Kind)
		}
	}
	return controls, nil
}

func writeEvalSetupFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func prepareEvalStepArtifacts(workspaceRoot string, step repository.EvalScenarioStep, artifactIndex map[string]repository.EvalArtifactRef) error {
	for _, artifactID := range step.CaptureArtifact {
		artifact, ok := artifactIndex[artifactID]
		if !ok || artifact.Kind != repository.EvalArtifactKindFile || artifact.Path == "" {
			continue
		}
		if err := os.MkdirAll(filepath.Join(workspaceRoot, filepath.Dir(filepath.FromSlash(artifact.Path))), 0o755); err != nil {
			return fmt.Errorf("prepare artifact path %q: %w", artifact.ID, err)
		}
	}
	return nil
}

func validateEvalRunRequest(request EvalRunRequest) error {
	hasID := request.ScenarioID != ""
	hasPath := request.ScenarioPath != ""
	if hasID == hasPath {
		return errors.New("eval runner requires exactly one of scenario ID or scenario path")
	}
	if request.FixturesRoot == "" {
		return errors.New("fixtures root is required")
	}
	if hasID && request.ScenariosDir == "" {
		return errors.New("scenarios directory is required when selecting by scenario ID")
	}
	return nil
}

func (r EvalRunner) withDefaults() EvalRunner {
	defaults := NewEvalRunner()
	if r.LoadScenario == nil {
		r.LoadScenario = defaults.LoadScenario
	}
	if r.LoadScenarios == nil {
		r.LoadScenarios = defaults.LoadScenarios
	}
	if r.ValidateFixtureReferences == nil {
		r.ValidateFixtureReferences = defaults.ValidateFixtureReferences
	}
	if r.RunCommand == nil {
		r.RunCommand = defaults.RunCommand
	}
	if r.RunMCPSession == nil {
		r.RunMCPSession = defaults.RunMCPSession
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

func (r EvalRunner) loadScenario(request EvalRunRequest) (repository.EvalScenarioDefinition, error) {
	if request.ScenarioPath != "" {
		return r.LoadScenario(request.ScenarioPath)
	}

	scenarios, err := r.LoadScenarios(request.ScenariosDir)
	if err != nil {
		return repository.EvalScenarioDefinition{}, err
	}
	for _, scenario := range scenarios {
		if scenario.ID == request.ScenarioID {
			return scenario, nil
		}
	}
	return repository.EvalScenarioDefinition{}, fmt.Errorf("unknown scenario %q", request.ScenarioID)
}

func (r EvalRunner) prepareWorkspace(ctx context.Context, request EvalRunRequest, scenario repository.EvalScenarioDefinition) (string, error) {
	tempRootParent := request.WorkspaceRoot
	tempRoot, err := r.MkdirTemp(tempRootParent, "optimusctx-eval-")
	if err != nil {
		return "", fmt.Errorf("create eval workspace: %w", err)
	}

	workspaceDirName := scenario.Fixture.WorkspaceDir
	if workspaceDirName == "" {
		workspaceDirName = "workspace"
	}
	workspaceRoot := filepath.Join(tempRoot, filepath.FromSlash(workspaceDirName))
	fixtureRoot := filepath.Join(request.FixturesRoot, filepath.FromSlash(scenario.Fixture.Path))

	if err := r.CopyTree(fixtureRoot, workspaceRoot); err != nil {
		return "", fmt.Errorf("materialize fixture %q: %w", scenario.Fixture.ID, err)
	}
	if err := r.GitInit(ctx, workspaceRoot); err != nil {
		return "", err
	}

	return workspaceRoot, nil
}

func seedEvalWatchStatus(workspaceRoot string, seed repository.EvalWatchStatusSeed) error {
	record := repository.WatchStatusRecord{
		PID:                   int(seed.PID),
		RepoRoot:              workspaceRoot,
		BinaryVersion:         seed.BinaryVersion,
		StartedAt:             seed.StartedAt,
		LastHeartbeatAt:       seed.LastHeartbeatAt,
		LastEventAt:           seed.LastEventAt,
		LastRefreshStartedAt:  seed.LastRefreshStartedAt,
		LastRefreshDoneAt:     seed.LastRefreshCompletedAt,
		LastRefreshGeneration: seed.LastRefreshGeneration,
		LastError:             seed.LastError,
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("encode watch status: %w", err)
	}
	data = append(data, '\n')
	statusPath := filepath.Join(workspaceRoot, ".optimusctx", "tmp", repository.DefaultWatchStatusFilename)
	if err := writeEvalSetupFile(statusPath, string(data)); err != nil {
		return fmt.Errorf("seed watch status: %w", err)
	}
	return nil
}

func seedEvalRepositoryState(workspaceRoot string, seed repository.EvalRepositoryStateSeed) error {
	ctx := context.Background()
	root, err := repository.ResolveRepositoryRoot(workspaceRoot)
	if err != nil {
		return fmt.Errorf("resolve repository root for state seed: %w", err)
	}
	layout, err := state.ResolveLayout(workspaceRoot)
	if err != nil {
		return fmt.Errorf("resolve state layout for state seed: %w", err)
	}
	store, err := sqlite.OpenOrCreateStore(ctx, layout, root.DetectionMode)
	if err != nil {
		return fmt.Errorf("open state store for state seed: %w", err)
	}
	defer store.Close()

	record, err := store.UpsertRepository(ctx, root, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("upsert repository for state seed: %w", err)
	}
	freshness, err := store.ReadRepositoryFreshness(ctx, record.ID)
	if err != nil {
		return fmt.Errorf("read repository freshness for state seed: %w", err)
	}
	freshness.FreshnessStatus = seed.FreshnessStatus
	freshness.FreshnessReason = seed.FreshnessReason
	if seed.LastRefreshStatus != "" {
		freshness.LastRefreshStatus = seed.LastRefreshStatus
	}
	if err := store.WriteRepositoryFreshness(ctx, freshness); err != nil {
		return fmt.Errorf("write repository freshness seed: %w", err)
	}
	return nil
}

func (r EvalRunner) executeEvalStep(ctx context.Context, workspaceRoot string, step repository.EvalScenarioStep, artifactIndex map[string]repository.EvalArtifactRef, controls evalStepControls) (repository.EvalStepResult, error) {
	return withEvalStepControls(controls, func() (repository.EvalStepResult, error) {
		switch step.Kind {
		case repository.EvalStepKindCommand:
			execution, err := r.RunCommand(ctx, EvalCommandInvocation{
				Args:       buildEvalStepArgs(step, artifactIndex),
				WorkingDir: workspaceRoot,
			})
			return repository.EvalStepResult{
				ExitCode: execution.ExitCode,
				Passed:   execution.ExitCode == step.Expect.ExitCode,
				Stdout:   execution.Stdout,
				Stderr:   execution.Stderr,
			}, err
		case repository.EvalStepKindMCPSession:
			if step.Session == nil {
				return repository.EvalStepResult{}, errors.New("mcp_session step is missing session configuration")
			}
			execution, err := r.RunMCPSession(ctx, EvalMCPSessionInvocation{
				WorkingDir: workspaceRoot,
				Session:    *step.Session,
			})
			if writeErr := persistEvalMCPSessionArtifacts(workspaceRoot, *step.Session, artifactIndex, execution); writeErr != nil {
				return repository.EvalStepResult{
					ExitCode: 1,
					Passed:   false,
					Stdout:   execution.Stdout,
					Stderr:   execution.Stderr,
				}, writeErr
			}
			result := repository.EvalStepResult{
				ExitCode: 0,
				Passed:   err == nil,
				Stdout:   execution.Stdout,
				Stderr:   execution.Stderr,
			}
			if err != nil {
				result.ExitCode = 1
			}
			return result, err
		default:
			return repository.EvalStepResult{}, fmt.Errorf("unsupported step kind %q", step.Kind)
		}
	})
}

func withEvalStepControls(controls evalStepControls, fn func() (repository.EvalStepResult, error)) (repository.EvalStepResult, error) {
	evalRefreshFailureMu.Lock()
	previous := evalRefreshFailureActive
	evalRefreshFailureActive = controls.RefreshFailure
	evalRefreshFailureMu.Unlock()
	defer func() {
		evalRefreshFailureMu.Lock()
		evalRefreshFailureActive = previous
		evalRefreshFailureMu.Unlock()
	}()
	return fn()
}

func currentEvalRefreshFailure() *evalRefreshFailureControl {
	evalRefreshFailureMu.Lock()
	defer evalRefreshFailureMu.Unlock()
	if evalRefreshFailureActive == nil {
		return nil
	}
	copy := *evalRefreshFailureActive
	return &copy
}

func buildEvalStepArgs(step repository.EvalScenarioStep, artifactIndex map[string]repository.EvalArtifactRef) []string {
	var args []string
	switch step.Expect.Command {
	case repository.EvalCommandInit:
		args = append(args, "init")
	case repository.EvalCommandRefresh:
		args = append(args, "refresh")
	case repository.EvalCommandDoctor:
		args = append(args, "doctor")
	case repository.EvalCommandPackExport:
		args = append(args, "pack", "export")
	default:
		return nil
	}
	args = append(args, step.Expect.Args...)

	if step.Expect.Command == repository.EvalCommandPackExport && !slices.Contains(args, "--output") {
		for _, artifactID := range step.CaptureArtifact {
			artifact, ok := artifactIndex[artifactID]
			if ok && artifact.Kind == repository.EvalArtifactKindFile {
				args = append(args, "--output", filepath.ToSlash(artifact.Path))
				break
			}
		}
	}

	return args
}

func persistEvalMCPSessionArtifacts(workspaceRoot string, session repository.EvalMCPSession, artifactIndex map[string]repository.EvalArtifactRef, execution EvalMCPSessionExecutionResult) error {
	if session.TranscriptArtifact != "" {
		artifact := artifactIndex[session.TranscriptArtifact]
		payload := map[string]any{
			"requests":  session.Requests,
			"responses": execution.Responses,
		}
		if err := writeEvalJSONArtifact(workspaceRoot, artifact, payload); err != nil {
			return fmt.Errorf("persist transcript artifact %q: %w", artifact.ID, err)
		}
	}

	responsesByID := make(map[int64]any, len(execution.Responses))
	for _, response := range execution.Responses {
		responsesByID[response.RequestID] = response.Response
	}
	for _, capture := range session.CaptureResponses {
		artifact := artifactIndex[capture.Artifact]
		response, ok := responsesByID[capture.RequestID]
		if !ok {
			return fmt.Errorf("missing MCP response for request id %d", capture.RequestID)
		}
		if err := writeEvalJSONArtifact(workspaceRoot, artifact, response); err != nil {
			return fmt.Errorf("persist response artifact %q: %w", artifact.ID, err)
		}
	}

	return nil
}

func writeEvalJSONArtifact(workspaceRoot string, artifact repository.EvalArtifactRef, payload any) error {
	if artifact.Kind != repository.EvalArtifactKindFile {
		return fmt.Errorf("artifact %q must be a file artifact", artifact.ID)
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("encode artifact payload: %w", err)
	}
	data = append(data, '\n')
	path := filepath.Join(workspaceRoot, filepath.FromSlash(artifact.Path))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func captureEvalArtifacts(workspaceRoot string, step repository.EvalScenarioStep, artifactIndex map[string]repository.EvalArtifactRef, execution EvalCommandExecutionResult) ([]repository.EvalArtifactResult, error) {
	results := make([]repository.EvalArtifactResult, 0, len(step.CaptureArtifact))
	for _, artifactID := range step.CaptureArtifact {
		artifact := artifactIndex[artifactID]
		result := repository.EvalArtifactResult{Ref: artifact}
		switch artifact.Kind {
		case repository.EvalArtifactKindStdout:
			result.Present = execution.Stdout != ""
			result.Bytes = int64(len(execution.Stdout))
			if result.Present {
				result.Location = "stdout"
			}
		case repository.EvalArtifactKindStderr:
			result.Present = execution.Stderr != ""
			result.Bytes = int64(len(execution.Stderr))
			if result.Present {
				result.Location = "stderr"
			}
		case repository.EvalArtifactKindFile:
			path := filepath.Join(workspaceRoot, filepath.FromSlash(artifact.Path))
			info, err := os.Stat(path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					if artifact.Required {
						return nil, fmt.Errorf("missing required artifact %q at %s", artifact.ID, artifact.Path)
					}
					results = append(results, result)
					continue
				}
				return nil, fmt.Errorf("stat artifact %q: %w", artifact.ID, err)
			}
			result.Location = path
			result.Present = true
			result.Bytes = info.Size()
		}

		if artifact.Required && !result.Present {
			return nil, fmt.Errorf("missing required artifact %q", artifact.ID)
		}
		results = append(results, result)
	}
	return results, nil
}

func evaluateEvalAssertions(step repository.EvalScenarioStep, artifacts []repository.EvalArtifactResult, artifactIndex map[string]repository.EvalArtifactRef, execution EvalCommandExecutionResult) error {
	if len(step.Assert) == 0 {
		return nil
	}

	resultsByID := make(map[string]repository.EvalArtifactResult, len(artifacts))
	for _, artifact := range artifacts {
		resultsByID[artifact.Ref.ID] = artifact
	}

	for idx, assertion := range step.Assert {
		content, err := resolveEvalAssertionContent(assertion, resultsByID, artifactIndex, execution)
		if err != nil {
			return fmt.Errorf("assert[%d]: %w", idx, err)
		}
		switch assertion.Kind {
		case repository.EvalAssertionKindContains:
			if !strings.Contains(content, assertion.Contains) {
				return fmt.Errorf("assert[%d]: content missing substring %q", idx, assertion.Contains)
			}
		case repository.EvalAssertionKindJSONFieldPresent:
			if _, ok, err := evalJSONField(content, assertion.Path); err != nil {
				return fmt.Errorf("assert[%d]: %w", idx, err)
			} else if !ok {
				return fmt.Errorf("assert[%d]: json field %q not present", idx, assertion.Path)
			}
		case repository.EvalAssertionKindJSONFieldEquals:
			value, ok, err := evalJSONField(content, assertion.Path)
			if err != nil {
				return fmt.Errorf("assert[%d]: %w", idx, err)
			}
			if !ok {
				return fmt.Errorf("assert[%d]: json field %q not present", idx, assertion.Path)
			}
			if !evalJSONValuesEqual(value, assertion.Equals) {
				return fmt.Errorf("assert[%d]: json field %q = %#v, want %#v", idx, assertion.Path, value, assertion.Equals)
			}
		default:
			return fmt.Errorf("assert[%d]: unsupported kind %q", idx, assertion.Kind)
		}
	}

	return nil
}

func resolveEvalAssertionContent(assertion repository.EvalAssertion, artifacts map[string]repository.EvalArtifactResult, artifactIndex map[string]repository.EvalArtifactRef, execution EvalCommandExecutionResult) (string, error) {
	switch assertion.Target {
	case repository.EvalAssertionTargetStdout:
		return execution.Stdout, nil
	case repository.EvalAssertionTargetStderr:
		return execution.Stderr, nil
	case repository.EvalAssertionTargetArtifact:
		result, ok := artifacts[assertion.Artifact]
		if !ok {
			ref, refOK := artifactIndex[assertion.Artifact]
			if !refOK {
				return "", fmt.Errorf("unknown artifact %q", assertion.Artifact)
			}
			return "", fmt.Errorf("artifact %q not captured for assertion target %q", ref.ID, assertion.Target)
		}
		if !result.Present {
			return "", fmt.Errorf("artifact %q is not present", assertion.Artifact)
		}
		switch result.Ref.Kind {
		case repository.EvalArtifactKindStdout:
			return execution.Stdout, nil
		case repository.EvalArtifactKindStderr:
			return execution.Stderr, nil
		case repository.EvalArtifactKindFile:
			data, err := os.ReadFile(result.Location)
			if err != nil {
				return "", fmt.Errorf("read artifact %q: %w", assertion.Artifact, err)
			}
			return string(data), nil
		default:
			return "", fmt.Errorf("unsupported artifact kind %q", result.Ref.Kind)
		}
	default:
		return "", fmt.Errorf("unsupported target %q", assertion.Target)
	}
}

func evalJSONField(content string, path string) (any, bool, error) {
	var decoded any
	if err := json.Unmarshal([]byte(content), &decoded); err != nil {
		return nil, false, fmt.Errorf("decode json assertion payload: %w", err)
	}

	current := decoded
	for _, segment := range strings.Split(path, ".") {
		switch node := current.(type) {
		case map[string]any:
			next, ok := node[segment]
			if !ok {
				return nil, false, nil
			}
			current = next
		case []any:
			index, err := strconv.Atoi(segment)
			if err != nil {
				return nil, false, fmt.Errorf("path segment %q must be an array index", segment)
			}
			if index < 0 || index >= len(node) {
				return nil, false, nil
			}
			current = node[index]
		default:
			return nil, false, nil
		}
	}
	return current, true, nil
}

func evalJSONValuesEqual(left any, right any) bool {
	return bytes.Equal(mustMarshalEvalValue(left), mustMarshalEvalValue(right))
}

func mustMarshalEvalValue(value any) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		return []byte(fmt.Sprintf("%#v", value))
	}
	return data
}

func indexEvalArtifacts(artifacts []repository.EvalArtifactRef) map[string]repository.EvalArtifactRef {
	index := make(map[string]repository.EvalArtifactRef, len(artifacts))
	for _, artifact := range artifacts {
		index[artifact.ID] = artifact
	}
	return index
}

func collectEvalArtifacts(artifacts []repository.EvalArtifactRef, seen map[string]repository.EvalArtifactResult) []repository.EvalArtifactResult {
	results := make([]repository.EvalArtifactResult, 0, len(seen))
	for _, artifact := range artifacts {
		result, ok := seen[artifact.ID]
		if ok {
			results = append(results, result)
		}
	}
	return results
}

func copyEvalTree(src string, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", src)
	}
	if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyEvalTree(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		if err := copyEvalFile(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func copyEvalFile(src string, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	info, err := input.Stat()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	output, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	return err
}
