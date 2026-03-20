package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

func TestRunCommand(t *testing.T) {
	previousServe := runCommandServer
	previousInput := runCommandInput
	previousStderr := runCommandStderr
	previousHealth := runHealthService
	previousInit := runInitService
	previousRefresh := runRefreshService
	t.Cleanup(func() {
		runCommandServer = previousServe
		runCommandInput = previousInput
		runCommandStderr = previousStderr
		runHealthService = previousHealth
		runInitService = previousInit
		runRefreshService = previousRefresh
	})

	var called bool
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runCommandInput = bytes.NewBufferString("")
	runCommandStderr = &stderr
	runHealthService = func(ctx context.Context, workingDir string) (repository.HealthResult, error) {
		return repository.HealthResult{
			Repository: repository.LayeredContextEnvelope{RepositoryRoot: "/repo", Generation: 7, Freshness: repository.FreshnessStatusFresh},
			Summary:    repository.HealthSummary{StateStatus: repository.HealthStateStatusReady},
		}, nil
	}
	runInitService = func(ctx context.Context, workingDir string) (app.InitResult, error) {
		t.Fatal("init should not be called for healthy state")
		return app.InitResult{}, nil
	}
	runRefreshService = func(ctx context.Context, workingDir string, reason repository.RefreshReason) (app.RefreshResult, error) {
		t.Fatal("refresh should not be called for fresh state")
		return app.RefreshResult{}, nil
	}
	runCommandServer = func(ctx context.Context, repoRoot string, stdin io.Reader, serveStdout io.Writer, serveStderr io.Writer) error {
		called = true
		if repoRoot != "/repo" {
			t.Fatalf("repo root = %q, want /repo", repoRoot)
		}
		if stdin != runCommandInput {
			t.Fatal("run stdin did not use configured input")
		}
		if serveStdout != &stdout {
			t.Fatal("run stdout did not use command stdout")
		}
		if serveStderr != &stderr {
			t.Fatal("run stderr did not use configured stderr")
		}
		_, _ = io.WriteString(serveStderr, "optimusctx mcp: ready for stdio requests\n")
		return nil
	}

	if err := NewRootCommand().Execute([]string{"run"}, &stdout); err != nil {
		t.Fatalf("Execute(run) error = %v", err)
	}
	if !called {
		t.Fatal("run server was not called")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() != "optimusctx mcp: ready for stdio requests\n" {
		t.Fatalf("stderr = %q, want readiness signal", stderr.String())
	}
}

func TestRunCommandBootstrapsMissingState(t *testing.T) {
	previousServe := runCommandServer
	previousInput := runCommandInput
	previousStderr := runCommandStderr
	previousHealth := runHealthService
	previousInit := runInitService
	previousRefresh := runRefreshService
	t.Cleanup(func() {
		runCommandServer = previousServe
		runCommandInput = previousInput
		runCommandStderr = previousStderr
		runHealthService = previousHealth
		runInitService = previousInit
		runRefreshService = previousRefresh
	})

	var calledInit bool
	var calledServer bool
	runCommandInput = bytes.NewBufferString("")
	runHealthService = func(ctx context.Context, workingDir string) (repository.HealthResult, error) {
		return repository.HealthResult{
			Summary: repository.HealthSummary{StateStatus: repository.HealthStateStatusMissing},
		}, nil
	}
	runInitService = func(ctx context.Context, workingDir string) (app.InitResult, error) {
		calledInit = true
		return app.InitResult{RepositoryRoot: "/repo"}, nil
	}
	runRefreshService = func(ctx context.Context, workingDir string, reason repository.RefreshReason) (app.RefreshResult, error) {
		t.Fatal("refresh should not be called immediately after init bootstrap")
		return app.RefreshResult{}, nil
	}
	runCommandServer = func(ctx context.Context, repoRoot string, stdin io.Reader, serveStdout io.Writer, serveStderr io.Writer) error {
		calledServer = true
		return nil
	}

	var stdout bytes.Buffer
	if err := NewRootCommand().Execute([]string{"run"}, &stdout); err != nil {
		t.Fatalf("Execute(run) error = %v", err)
	}
	if !calledInit {
		t.Fatal("run should call init when state is missing")
	}
	if !calledServer {
		t.Fatal("run should start server after init bootstrap")
	}
}

func TestRunCommandRefreshesStaleState(t *testing.T) {
	previousServe := runCommandServer
	previousInput := runCommandInput
	previousStderr := runCommandStderr
	previousHealth := runHealthService
	previousInit := runInitService
	previousRefresh := runRefreshService
	t.Cleanup(func() {
		runCommandServer = previousServe
		runCommandInput = previousInput
		runCommandStderr = previousStderr
		runHealthService = previousHealth
		runInitService = previousInit
		runRefreshService = previousRefresh
	})

	var calledRefresh bool
	var calledServer bool
	runCommandInput = bytes.NewBufferString("")
	runHealthService = func(ctx context.Context, workingDir string) (repository.HealthResult, error) {
		return repository.HealthResult{
			Repository: repository.LayeredContextEnvelope{RepositoryRoot: "/repo", Freshness: repository.FreshnessStatusStale},
			Summary:    repository.HealthSummary{StateStatus: repository.HealthStateStatusReady},
		}, nil
	}
	runInitService = func(ctx context.Context, workingDir string) (app.InitResult, error) {
		t.Fatal("init should not be called for existing state")
		return app.InitResult{}, nil
	}
	runRefreshService = func(ctx context.Context, workingDir string, reason repository.RefreshReason) (app.RefreshResult, error) {
		calledRefresh = true
		if reason != repository.RefreshReasonWatch {
			t.Fatalf("refresh reason = %q, want watch", reason)
		}
		return app.RefreshResult{RepositoryRoot: "/repo", FreshnessStatus: repository.FreshnessStatusFresh}, nil
	}
	runCommandServer = func(ctx context.Context, repoRoot string, stdin io.Reader, serveStdout io.Writer, serveStderr io.Writer) error {
		calledServer = true
		if repoRoot != "/repo" {
			t.Fatalf("repo root = %q, want /repo", repoRoot)
		}
		return nil
	}

	var stdout bytes.Buffer
	if err := NewRootCommand().Execute([]string{"run"}, &stdout); err != nil {
		t.Fatalf("Execute(run) error = %v", err)
	}
	if !calledRefresh {
		t.Fatal("run should refresh stale state before serving")
	}
	if !calledServer {
		t.Fatal("run should start server after refresh")
	}
}

func TestRunCommandRejectsUnsupportedArguments(t *testing.T) {
	var stdout bytes.Buffer

	err := NewRootCommand().Execute([]string{"run", "--verbose"}, &stdout)
	if err == nil || err.Error() != "run does not accept flags; got \"--verbose\"" {
		t.Fatalf("Execute(run --verbose) error = %v", err)
	}
}

func TestRunCommandRejectsUnsupportedWorkingDirectory(t *testing.T) {
	previousHealth := runHealthService
	t.Cleanup(func() { runHealthService = previousHealth })
	runHealthService = func(ctx context.Context, workingDir string) (repository.HealthResult, error) {
		return repository.HealthResult{}, repository.ErrRepositoryNotFound
	}

	var stdout bytes.Buffer
	err := NewRootCommand().Execute([]string{"run"}, &stdout)
	if err == nil || !strings.Contains(err.Error(), "no supported repository root found") {
		t.Fatalf("Execute(run) error = %v", err)
	}
}

func TestRootHelpListsRun(t *testing.T) {
	var stdout bytes.Buffer
	if err := NewRootCommand().Execute([]string{"--help"}, &stdout); err != nil {
		t.Fatalf("Execute(--help) error = %v", err)
	}
	if !strings.Contains(stdout.String(), "run       Run the agent-facing OptimusCtx runtime over STDIO") {
		t.Fatalf("help missing run command:\n%s", stdout.String())
	}
}
