package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

var (
	evalCommandService func(context.Context, app.EvalRunRequest) (repository.EvalRunResult, error)
	evalCommandMu      sync.Mutex
)

func newEvalCommand() *Command {
	return &Command{
		Name:    "eval",
		Summary: "Run versioned evaluation scenarios through the shipped CLI boundary",
		Run: func(stdout io.Writer, args []string) error {
			for _, arg := range args {
				if arg == "-h" || arg == "--help" {
					return writeEvalHelp(stdout)
				}
			}

			request, err := parseEvalArgs(args)
			if err != nil {
				return err
			}

			workingDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve working directory: %w", err)
			}
			root, err := repository.ResolveRepositoryRoot(workingDir)
			if err != nil {
				return fmt.Errorf("resolve source repository root: %w", err)
			}

			request.ScenariosDir = filepath.Join(root.RootPath, "testdata", "eval", "scenarios")
			request.FixturesRoot = filepath.Join(root.RootPath, "testdata", "eval", "fixtures")
			if request.ScenarioPath != "" && !filepath.IsAbs(request.ScenarioPath) {
				request.ScenarioPath = filepath.Join(workingDir, request.ScenarioPath)
			}

			result, err := runEvalCommandService(context.Background(), request)
			if result.ScenarioID != "" {
				if _, writeErr := io.WriteString(stdout, formatEvalRunSummary(result)); writeErr != nil {
					return writeErr
				}
			}
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func runEvalCommandService(ctx context.Context, request app.EvalRunRequest) (repository.EvalRunResult, error) {
	if evalCommandService != nil {
		return evalCommandService(ctx, request)
	}

	runner := app.NewEvalRunner()
	runner.RunCommand = executeEvalCLICommand
	return runner.Run(ctx, request)
}

func parseEvalArgs(args []string) (app.EvalRunRequest, error) {
	var request app.EvalRunRequest

	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--scenario":
			index++
			if index >= len(args) {
				return app.EvalRunRequest{}, fmt.Errorf("%s requires a value", arg)
			}
			request.ScenarioID = strings.TrimSpace(args[index])
			if request.ScenarioID == "" {
				return app.EvalRunRequest{}, fmt.Errorf("%s requires a non-empty value", arg)
			}
		case "--scenario-file":
			index++
			if index >= len(args) {
				return app.EvalRunRequest{}, fmt.Errorf("%s requires a value", arg)
			}
			request.ScenarioPath = strings.TrimSpace(args[index])
			if request.ScenarioPath == "" {
				return app.EvalRunRequest{}, fmt.Errorf("%s requires a non-empty value", arg)
			}
		default:
			if len(arg) > 0 && arg[0] == '-' {
				return app.EvalRunRequest{}, fmt.Errorf("unknown eval flag %q", arg)
			}
			return app.EvalRunRequest{}, fmt.Errorf("eval does not accept arguments; got %q", arg)
		}
	}

	if (request.ScenarioID == "") == (request.ScenarioPath == "") {
		return app.EvalRunRequest{}, errors.New("eval requires exactly one of --scenario or --scenario-file")
	}

	return request, nil
}

func executeEvalCLICommand(_ context.Context, invocation app.EvalCommandInvocation) (app.EvalCommandExecutionResult, error) {
	evalCommandMu.Lock()
	defer evalCommandMu.Unlock()

	originalDir, err := os.Getwd()
	if err != nil {
		return app.EvalCommandExecutionResult{}, fmt.Errorf("capture working directory: %w", err)
	}
	if err := os.Chdir(invocation.WorkingDir); err != nil {
		return app.EvalCommandExecutionResult{}, fmt.Errorf("switch working directory: %w", err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	var stdout bytes.Buffer
	result := app.EvalCommandExecutionResult{}
	err = NewRootCommand().Execute(invocation.Args, &stdout)
	result.Stdout = stdout.String()
	if err != nil {
		result.Stderr = err.Error() + "\n"
		result.ExitCode = 1
		return result, nil
	}
	return result, nil
}

func writeEvalHelp(stdout io.Writer) error {
	_, err := io.WriteString(stdout, "Usage:\n  optimusctx eval [--scenario ID | --scenario-file PATH]\n\nRun one versioned evaluation scenario through the shipped CLI boundary.\n")
	return err
}

func formatEvalRunSummary(result repository.EvalRunResult) string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "scenario id: %s\nstatus: %s\nstarted at: %s\nfinished at: %s\nsteps: %d\n",
		result.ScenarioID,
		renderEvalStatus(result.Passed),
		result.StartedAt.UTC().Format(time.RFC3339),
		result.FinishedAt.UTC().Format(time.RFC3339),
		len(result.Steps),
	)
	for _, step := range result.Steps {
		duration := step.FinishedAt.Sub(step.StartedAt).Round(time.Millisecond)
		_, _ = fmt.Fprintf(&b, "step %s: %s exit=%d duration=%s\n", step.Step.ID, renderEvalStatus(step.Passed), step.ExitCode, duration)
	}
	return b.String()
}

func renderEvalStatus(passed bool) string {
	if passed {
		return "passed"
	}
	return "failed"
}
