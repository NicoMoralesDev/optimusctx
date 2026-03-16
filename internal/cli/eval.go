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
	evalCommandService              func(context.Context, app.EvalRunRequest) (repository.EvalRunResult, error)
	benchmarkEvidenceCommandService func(context.Context, app.BenchmarkEvidenceBundleRequest) (repository.BenchmarkEvidenceBundle, error)
	benchmarkReportCommandService   func(context.Context, app.BenchmarkHumanReportRequest) (string, error)
	benchmarkVerifyCommandService   func(context.Context, app.BenchmarkMilestoneVerificationRequest) (app.BenchmarkMilestoneVerificationResult, error)
	evalCommandMu                   sync.Mutex
)

func newEvalCommand() *Command {
	return &Command{
		Name:    "eval",
		Summary: "Run versioned evaluation scenarios through the shipped CLI boundary",
		Run: func(stdout io.Writer, args []string) error {
			if len(args) > 0 && args[0] == "benchmark" {
				return runEvalBenchmarkCommand(stdout, args[1:])
			}
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

func runBenchmarkEvidenceCommandService(ctx context.Context, request app.BenchmarkEvidenceBundleRequest) (repository.BenchmarkEvidenceBundle, error) {
	if benchmarkEvidenceCommandService != nil {
		return benchmarkEvidenceCommandService(ctx, request)
	}
	service := app.NewBenchmarkService()
	service.Runner = app.NewBenchmarkRunner()
	service.Runner.RunCommand = func(ctx context.Context, invocation app.BenchmarkCommandInvocation) (app.BenchmarkCommandExecutionResult, error) {
		execution, err := executeEvalCLICommand(ctx, app.EvalCommandInvocation{
			Args:       invocation.Args,
			WorkingDir: invocation.WorkingDir,
		})
		return app.BenchmarkCommandExecutionResult{
			Stdout:   execution.Stdout,
			Stderr:   execution.Stderr,
			ExitCode: execution.ExitCode,
		}, err
	}
	service.Runner.RunTool = func(ctx context.Context, invocation app.BenchmarkToolInvocation) (app.BenchmarkToolExecutionResult, error) {
		session := repository.EvalMCPSession{
			Requests: []repository.EvalMCPRequest{
				{ID: 1, Method: "initialize", Params: mcp.InitializeParams{
					ClientInfo:      mcp.ClientInfo{Name: "benchmark-export", Version: "1.0.0"},
					ProtocolVersion: "2024-11-05",
				}},
				{Method: "notifications/initialized", Notification: true},
				{ID: 2, Method: "tools/call", Params: mcp.CallToolParams{
					Name:      invocation.Name,
					Arguments: invocation.Arguments,
				}},
			},
		}
		execution, err := executeEvalCLIMCPSession(ctx, app.EvalMCPSessionInvocation{
			WorkingDir: invocation.WorkingDir,
			Session:    session,
		})
		if err != nil {
			return app.BenchmarkToolExecutionResult{}, err
		}
		payload, err := decodeBenchmarkToolPayload(execution.Responses[len(execution.Responses)-1].Response)
		if err != nil {
			return app.BenchmarkToolExecutionResult{}, err
		}
		return app.BenchmarkToolExecutionResult{Payload: payload}, nil
	}
	return service.ExportEvidenceBundle(ctx, request)
}

func runBenchmarkReportCommandService(ctx context.Context, request app.BenchmarkHumanReportRequest) (string, error) {
	if benchmarkReportCommandService != nil {
		return benchmarkReportCommandService(ctx, request)
	}
	service := app.NewBenchmarkService()
	service.Runner = app.NewBenchmarkRunner()
	service.Runner.RunCommand = func(ctx context.Context, invocation app.BenchmarkCommandInvocation) (app.BenchmarkCommandExecutionResult, error) {
		execution, err := executeEvalCLICommand(ctx, app.EvalCommandInvocation{
			Args:       invocation.Args,
			WorkingDir: invocation.WorkingDir,
		})
		return app.BenchmarkCommandExecutionResult{
			Stdout:   execution.Stdout,
			Stderr:   execution.Stderr,
			ExitCode: execution.ExitCode,
		}, err
	}
	service.Runner.RunTool = func(ctx context.Context, invocation app.BenchmarkToolInvocation) (app.BenchmarkToolExecutionResult, error) {
		session := repository.EvalMCPSession{
			Requests: []repository.EvalMCPRequest{
				{ID: 1, Method: "initialize", Params: mcp.InitializeParams{
					ClientInfo:      mcp.ClientInfo{Name: "benchmark-report", Version: "1.0.0"},
					ProtocolVersion: "2024-11-05",
				}},
				{Method: "notifications/initialized", Notification: true},
				{ID: 2, Method: "tools/call", Params: mcp.CallToolParams{
					Name:      invocation.Name,
					Arguments: invocation.Arguments,
				}},
			},
		}
		execution, err := executeEvalCLIMCPSession(ctx, app.EvalMCPSessionInvocation{
			WorkingDir: invocation.WorkingDir,
			Session:    session,
		})
		if err != nil {
			return app.BenchmarkToolExecutionResult{}, err
		}
		payload, err := decodeBenchmarkToolPayload(execution.Responses[len(execution.Responses)-1].Response)
		if err != nil {
			return app.BenchmarkToolExecutionResult{}, err
		}
		return app.BenchmarkToolExecutionResult{Payload: payload}, nil
	}
	return service.RenderHumanReport(ctx, request)
}

func runBenchmarkVerifyCommandService(ctx context.Context, request app.BenchmarkMilestoneVerificationRequest) (app.BenchmarkMilestoneVerificationResult, error) {
	if benchmarkVerifyCommandService != nil {
		return benchmarkVerifyCommandService(ctx, request)
	}
	service := app.NewBenchmarkService()
	service.Runner = app.NewBenchmarkRunner()
	service.Runner.RunCommand = func(ctx context.Context, invocation app.BenchmarkCommandInvocation) (app.BenchmarkCommandExecutionResult, error) {
		execution, err := executeEvalCLICommand(ctx, app.EvalCommandInvocation{
			Args:       invocation.Args,
			WorkingDir: invocation.WorkingDir,
		})
		return app.BenchmarkCommandExecutionResult{
			Stdout:   execution.Stdout,
			Stderr:   execution.Stderr,
			ExitCode: execution.ExitCode,
		}, err
	}
	service.Runner.RunTool = func(ctx context.Context, invocation app.BenchmarkToolInvocation) (app.BenchmarkToolExecutionResult, error) {
		session := repository.EvalMCPSession{
			Requests: []repository.EvalMCPRequest{
				{ID: 1, Method: "initialize", Params: mcp.InitializeParams{
					ClientInfo:      mcp.ClientInfo{Name: "benchmark-verify", Version: "1.0.0"},
					ProtocolVersion: "2024-11-05",
				}},
				{Method: "notifications/initialized", Notification: true},
				{ID: 2, Method: "tools/call", Params: mcp.CallToolParams{
					Name:      invocation.Name,
					Arguments: invocation.Arguments,
				}},
			},
		}
		execution, err := executeEvalCLIMCPSession(ctx, app.EvalMCPSessionInvocation{
			WorkingDir: invocation.WorkingDir,
			Session:    session,
		})
		if err != nil {
			return app.BenchmarkToolExecutionResult{}, err
		}
		payload, err := decodeBenchmarkToolPayload(execution.Responses[len(execution.Responses)-1].Response)
		if err != nil {
			return app.BenchmarkToolExecutionResult{}, err
		}
		return app.BenchmarkToolExecutionResult{Payload: payload}, nil
	}
	return service.VerifyMilestoneEvidence(ctx, request)
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
	_, err := io.WriteString(stdout, "Usage:\n  optimusctx eval [--scenario ID | --scenario-file PATH]\n  optimusctx eval benchmark export [--suite ID | --suite-file PATH] [--attempts N] [--output PATH]\n  optimusctx eval benchmark report [--suite ID | --suite-file PATH] [--attempts N] [--output PATH]\n  optimusctx eval benchmark verify [--suite ID | --suite-file PATH] [--attempts N]\n\nRun versioned evaluation scenarios or render deterministic counted-input benchmark evidence through the shipped CLI boundary.\n")
	return err
}

func runEvalBenchmarkCommand(stdout io.Writer, args []string) error {
	if len(args) == 0 {
		return errors.New("eval benchmark requires a subcommand")
	}
	switch args[0] {
	case "export":
		request, outputPath, err := parseBenchmarkExportArgs(args[1:])
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
		request.StartPath = workingDir
		if request.SuitePath == "" {
			request.SuitesDir = filepath.Join(root.RootPath, "testdata", "eval", "benchmarks")
		}
		request.FixturesRoot = filepath.Join(root.RootPath, "testdata", "eval", "fixtures")
		if request.SuitePath != "" && !filepath.IsAbs(request.SuitePath) {
			request.SuitePath = filepath.Join(workingDir, request.SuitePath)
		}

		bundle, err := runBenchmarkEvidenceCommandService(context.Background(), request)
		if err != nil {
			return err
		}
		payload, err := repository.MarshalBenchmarkEvidenceBundle(bundle)
		if err != nil {
			return fmt.Errorf("encode benchmark evidence bundle: %w", err)
		}
		if outputPath != "" {
			if !filepath.IsAbs(outputPath) {
				outputPath = filepath.Join(workingDir, outputPath)
			}
			if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
				return fmt.Errorf("create benchmark export directory: %w", err)
			}
			if err := os.WriteFile(outputPath, payload, 0o644); err != nil {
				return fmt.Errorf("write benchmark evidence bundle: %w", err)
			}
			_, err = fmt.Fprintf(stdout, "benchmark evidence written: %s\n", outputPath)
			return err
		}
		_, err = stdout.Write(payload)
		return err
	case "report":
		request, outputPath, err := parseBenchmarkReportArgs(args[1:])
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
		request.StartPath = workingDir
		if request.SuitePath == "" {
			request.SuitesDir = filepath.Join(root.RootPath, "testdata", "eval", "benchmarks")
		}
		request.FixturesRoot = filepath.Join(root.RootPath, "testdata", "eval", "fixtures")
		if request.SuitePath != "" && !filepath.IsAbs(request.SuitePath) {
			request.SuitePath = filepath.Join(workingDir, request.SuitePath)
		}

		report, err := runBenchmarkReportCommandService(context.Background(), request)
		if err != nil {
			return err
		}
		if outputPath != "" {
			if !filepath.IsAbs(outputPath) {
				outputPath = filepath.Join(workingDir, outputPath)
			}
			if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
				return fmt.Errorf("create benchmark report directory: %w", err)
			}
			if err := os.WriteFile(outputPath, []byte(report), 0o644); err != nil {
				return fmt.Errorf("write benchmark report: %w", err)
			}
			_, err = fmt.Fprintf(stdout, "benchmark report written: %s\n", outputPath)
			return err
		}
		_, err = io.WriteString(stdout, report)
		return err
	case "verify":
		request, err := parseBenchmarkVerifyArgs(args[1:])
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
		request.StartPath = workingDir
		if request.SuitePath == "" {
			request.SuitesDir = filepath.Join(root.RootPath, "testdata", "eval", "benchmarks")
		}
		request.FixturesRoot = filepath.Join(root.RootPath, "testdata", "eval", "fixtures")
		if request.SuitePath != "" && !filepath.IsAbs(request.SuitePath) {
			request.SuitePath = filepath.Join(workingDir, request.SuitePath)
		}

		result, err := runBenchmarkVerifyCommandService(context.Background(), request)
		if err != nil {
			return err
		}
		_, err = io.WriteString(stdout, formatBenchmarkVerificationSummary(result))
		return err
	default:
		return fmt.Errorf("unknown eval benchmark subcommand %q", args[0])
	}
}

func parseBenchmarkExportArgs(args []string) (app.BenchmarkEvidenceBundleRequest, string, error) {
	request, outputPath, err := parseBenchmarkArgsCommon(args, "export")
	if err != nil {
		return app.BenchmarkEvidenceBundleRequest{}, "", err
	}
	return request, outputPath, nil
}

func parseBenchmarkReportArgs(args []string) (app.BenchmarkHumanReportRequest, string, error) {
	request, outputPath, err := parseBenchmarkArgsCommon(args, "report")
	if err != nil {
		return app.BenchmarkHumanReportRequest{}, "", err
	}
	return app.BenchmarkHumanReportRequest{
		SuiteID:   request.SuiteID,
		SuitePath: request.SuitePath,
		Attempts:  request.Attempts,
	}, outputPath, nil
}

func parseBenchmarkVerifyArgs(args []string) (app.BenchmarkMilestoneVerificationRequest, error) {
	request, _, err := parseBenchmarkArgsCommon(args, "verify")
	if err != nil {
		return app.BenchmarkMilestoneVerificationRequest{}, err
	}
	return app.BenchmarkMilestoneVerificationRequest{
		SuiteID:   request.SuiteID,
		SuitePath: request.SuitePath,
		Attempts:  request.Attempts,
	}, nil
}

func parseBenchmarkArgsCommon(args []string, subcommand string) (app.BenchmarkEvidenceBundleRequest, string, error) {
	var request app.BenchmarkEvidenceBundleRequest
	var outputPath string
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--suite":
			index++
			if index >= len(args) {
				return app.BenchmarkEvidenceBundleRequest{}, "", fmt.Errorf("%s requires a value", arg)
			}
			request.SuiteID = strings.TrimSpace(args[index])
			if request.SuiteID == "" {
				return app.BenchmarkEvidenceBundleRequest{}, "", fmt.Errorf("%s requires a non-empty value", arg)
			}
		case "--suite-file":
			index++
			if index >= len(args) {
				return app.BenchmarkEvidenceBundleRequest{}, "", fmt.Errorf("%s requires a value", arg)
			}
			request.SuitePath = strings.TrimSpace(args[index])
			if request.SuitePath == "" {
				return app.BenchmarkEvidenceBundleRequest{}, "", fmt.Errorf("%s requires a non-empty value", arg)
			}
		case "--attempts":
			index++
			if index >= len(args) {
				return app.BenchmarkEvidenceBundleRequest{}, "", fmt.Errorf("%s requires a value", arg)
			}
			value, err := strconv.Atoi(strings.TrimSpace(args[index]))
			if err != nil || value <= 0 {
				return app.BenchmarkEvidenceBundleRequest{}, "", fmt.Errorf("%s requires a positive integer", arg)
			}
			request.Attempts = value
		case "--output":
			if subcommand == "verify" {
				return app.BenchmarkEvidenceBundleRequest{}, "", fmt.Errorf("%s does not accept %s", subcommand, arg)
			}
			index++
			if index >= len(args) {
				return app.BenchmarkEvidenceBundleRequest{}, "", fmt.Errorf("%s requires a value", arg)
			}
			outputPath = strings.TrimSpace(args[index])
			if outputPath == "" {
				return app.BenchmarkEvidenceBundleRequest{}, "", fmt.Errorf("%s requires a non-empty value", arg)
			}
		default:
			return app.BenchmarkEvidenceBundleRequest{}, "", fmt.Errorf("unknown eval benchmark %s flag %q", subcommand, arg)
		}
	}
	if (request.SuiteID == "") == (request.SuitePath == "") {
		return app.BenchmarkEvidenceBundleRequest{}, "", fmt.Errorf("eval benchmark %s requires exactly one of --suite or --suite-file", subcommand)
	}
	return request, outputPath, nil
}

func formatBenchmarkVerificationSummary(result app.BenchmarkMilestoneVerificationResult) string {
	var b strings.Builder
	status := "passed"
	if !result.Passed {
		status = "failed"
	}
	_, _ = fmt.Fprintf(&b, "benchmark milestone verification\nsuite: %s@%s\nfixture: %s (%s)\nattempts: %d\nstatus: %s\nmethodology fingerprint: %s\npersisted methodology fingerprint: %s\nrerun command: %s\n",
		result.SuiteID,
		result.SuiteVersion,
		result.FixtureID,
		result.FixturePath,
		result.AttemptCount,
		status,
		result.MethodologyFingerprint,
		result.PersistedFingerprint,
		result.RerunCommand,
	)
	if result.Verification.Passed {
		_, _ = fmt.Fprintf(&b, "repeated-run verification: passed\n")
	} else {
		_, _ = fmt.Fprintf(&b, "repeated-run verification: failed\n")
	}
	if result.ReproducibilityPassed {
		_, _ = fmt.Fprintf(&b, "reproducibility verification: passed\n")
	} else {
		_, _ = fmt.Fprintf(&b, "reproducibility verification: failed\n")
	}
	if result.ReportVerificationPassed {
		_, _ = fmt.Fprintf(&b, "report wording verification: passed\n")
	} else {
		_, _ = fmt.Fprintf(&b, "report wording verification: failed\n")
	}
	if len(result.ReproducibilityDriftReasons) > 0 || len(result.ReportVerificationReasons) > 0 || len(result.Verification.DriftReasons) > 0 {
		_, _ = fmt.Fprintf(&b, "drift reasons\n")
		for _, reason := range append(append([]string(nil), result.ReproducibilityDriftReasons...), result.ReportVerificationReasons...) {
			_, _ = fmt.Fprintf(&b, "- %s\n", reason)
		}
		for _, reason := range result.Verification.DriftReasons {
			_, _ = fmt.Fprintf(&b, "- %s\n", reason)
		}
	}
	if result.FailureReason != "" {
		_, _ = fmt.Fprintf(&b, "failure reason: %s\n", result.FailureReason)
	}
	return b.String()
}

func decodeBenchmarkToolPayload(response any) (any, error) {
	data, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	var frame struct {
		Result struct {
			StructuredContent struct {
				Data any `json:"data"`
			} `json:"structuredContent"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &frame); err != nil {
		return nil, err
	}
	return frame.Result.StructuredContent.Data, nil
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
