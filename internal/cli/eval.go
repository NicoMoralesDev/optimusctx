package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/mcp"
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
			request.StartPath = workingDir
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
	runner.RunMCPSession = executeEvalCLIMCPSession
	service := app.NewEvalService()
	service.Runner = runner
	return service.Run(ctx, request)
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

func executeEvalCLIMCPSession(_ context.Context, invocation app.EvalMCPSessionInvocation) (app.EvalMCPSessionExecutionResult, error) {
	evalCommandMu.Lock()
	defer evalCommandMu.Unlock()

	originalDir, err := os.Getwd()
	if err != nil {
		return app.EvalMCPSessionExecutionResult{}, fmt.Errorf("capture working directory: %w", err)
	}
	if err := os.Chdir(invocation.WorkingDir); err != nil {
		return app.EvalMCPSessionExecutionResult{}, fmt.Errorf("switch working directory: %w", err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	var input bytes.Buffer
	for _, request := range invocation.Session.Requests {
		frame := mcp.Request{
			JSONRPC: "2.0",
			Method:  request.Method,
			Params:  request.Params,
		}
		if request.ID > 0 {
			frame.ID = request.ID
		}
		if err := writeEvalMCPFrame(&input, frame); err != nil {
			return app.EvalMCPSessionExecutionResult{}, fmt.Errorf("encode MCP request %q: %w", request.Method, err)
		}
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	prevInput := mcpServeInput
	prevStderr := mcpServeStderr
	mcpServeInput = &input
	mcpServeStderr = &stderr
	defer func() {
		mcpServeInput = prevInput
		mcpServeStderr = prevStderr
	}()

	if err := NewRootCommand().Execute([]string{"mcp", "serve"}, &stdout); err != nil {
		return app.EvalMCPSessionExecutionResult{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
		}, err
	}

	responses, err := readEvalMCPResponses(&stdout)
	if err != nil {
		return app.EvalMCPSessionExecutionResult{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
		}, err
	}

	return app.EvalMCPSessionExecutionResult{
		Stdout:    stdout.String(),
		Stderr:    stderr.String(),
		Responses: responses,
	}, nil
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

func writeEvalMCPFrame(writer io.Writer, request mcp.Request) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "Content-Length: %d\r\n\r\n", len(payload)); err != nil {
		return err
	}
	_, err = writer.Write(payload)
	return err
}

func readEvalMCPResponses(reader io.Reader) ([]app.EvalMCPSessionResponse, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var responses []app.EvalMCPSessionResponse
	for len(data) > 0 {
		headerEnd := bytes.Index(data, []byte("\r\n\r\n"))
		if headerEnd < 0 {
			return nil, errors.New("invalid MCP response frame header")
		}
		header := string(data[:headerEnd])
		contentLength, err := parseEvalMCPContentLength(header)
		if err != nil {
			return nil, err
		}

		frameStart := headerEnd + len("\r\n\r\n")
		if len(data) < frameStart+contentLength {
			return nil, errors.New("truncated MCP response frame")
		}

		frame := data[frameStart : frameStart+contentLength]
		var response map[string]any
		if err := json.Unmarshal(frame, &response); err != nil {
			return nil, fmt.Errorf("decode MCP response: %w", err)
		}

		requestID, ok := normalizeEvalResponseID(response["id"])
		if ok {
			responses = append(responses, app.EvalMCPSessionResponse{
				RequestID: requestID,
				Response:  response,
			})
		}

		data = data[frameStart+contentLength:]
	}

	return responses, nil
}

func parseEvalMCPContentLength(header string) (int, error) {
	for _, line := range strings.Split(header, "\r\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(parts[0]), "Content-Length") {
			value := strings.TrimSpace(parts[1])
			length, err := strconv.Atoi(value)
			if err != nil || length < 0 {
				return 0, fmt.Errorf("invalid Content-Length %q", value)
			}
			return length, nil
		}
	}
	return 0, errors.New("missing Content-Length header")
}

func normalizeEvalResponseID(value any) (int64, bool) {
	switch id := value.(type) {
	case float64:
		return int64(id), true
	case int64:
		return id, true
	case int:
		return int64(id), true
	default:
		return 0, false
	}
}
