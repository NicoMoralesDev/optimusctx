package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
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

type EvalRunner struct {
	LoadScenario              func(string) (repository.EvalScenarioDefinition, error)
	LoadScenarios             func(string) ([]repository.EvalScenarioDefinition, error)
	ValidateFixtureReferences func([]repository.EvalScenarioDefinition, string) error
	RunCommand                EvalCommandExecutor
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
		if err := prepareEvalStepArtifacts(workspaceRoot, step, artifactIndex); err != nil {
			runResult.FinishedAt = r.Now().UTC()
			runResult.Passed = false
			runResult.Artifacts = collectEvalArtifacts(scenario.Artifacts, seenArtifacts)
			return runResult, fmt.Errorf("scenario %q step %q: %w", scenario.ID, step.ID, err)
		}

		stepStartedAt := r.Now().UTC()
		execution, execErr := r.RunCommand(ctx, EvalCommandInvocation{
			Args:       buildEvalStepArgs(step, artifactIndex),
			WorkingDir: workspaceRoot,
		})
		stepFinishedAt := r.Now().UTC()

		stepResult := repository.EvalStepResult{
			Step:       step,
			StartedAt:  stepStartedAt,
			FinishedAt: stepFinishedAt,
			ExitCode:   execution.ExitCode,
			Passed:     execution.ExitCode == step.Expect.ExitCode,
			Stdout:     execution.Stdout,
			Stderr:     execution.Stderr,
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
		if execErr != nil {
			stepResult.Passed = false
			runResult.Steps = append(runResult.Steps, stepResult)
			runResult.FinishedAt = r.Now().UTC()
			runResult.Passed = false
			runResult.Artifacts = collectEvalArtifacts(scenario.Artifacts, seenArtifacts)
			return runResult, fmt.Errorf("scenario %q step %q: %w", scenario.ID, step.ID, execErr)
		}
		if execution.ExitCode != step.Expect.ExitCode {
			runResult.Steps = append(runResult.Steps, stepResult)
			runResult.FinishedAt = r.Now().UTC()
			runResult.Passed = false
			runResult.Artifacts = collectEvalArtifacts(scenario.Artifacts, seenArtifacts)
			return runResult, fmt.Errorf("scenario %q step %q failed: exit code %d, want %d", scenario.ID, step.ID, execution.ExitCode, step.Expect.ExitCode)
		}

		runResult.Steps = append(runResult.Steps, stepResult)
	}

	runResult.FinishedAt = r.Now().UTC()
	runResult.Passed = true
	runResult.Artifacts = collectEvalArtifacts(scenario.Artifacts, seenArtifacts)
	return runResult, nil
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
