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
	previousResolve := runResolveRepositoryRoot
	previousHealth := runHealthService
	previousInit := runInitService
	previousRefresh := runRefreshService
	previousWatch := runWatchService
	t.Cleanup(func() {
		runCommandServer = previousServe
		runCommandInput = previousInput
		runCommandStderr = previousStderr
		runResolveRepositoryRoot = previousResolve
		runHealthService = previousHealth
		runInitService = previousInit
		runRefreshService = previousRefresh
		runWatchService = previousWatch
	})

	var called bool
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runCommandInput = bytes.NewBufferString("")
	runCommandStderr = &stderr
	runResolveRepositoryRoot = func(workingDir string) (repository.RepositoryRoot, error) {
		return repository.RepositoryRoot{RootPath: "/repo"}, nil
	}
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
	previousResolve := runResolveRepositoryRoot
	previousHealth := runHealthService
	previousInit := runInitService
	previousRefresh := runRefreshService
	previousWatch := runWatchService
	t.Cleanup(func() {
		runCommandServer = previousServe
		runCommandInput = previousInput
		runCommandStderr = previousStderr
		runResolveRepositoryRoot = previousResolve
		runHealthService = previousHealth
		runInitService = previousInit
		runRefreshService = previousRefresh
		runWatchService = previousWatch
	})

	var calledInit bool
	var calledServer bool
	runCommandInput = bytes.NewBufferString("")
	runResolveRepositoryRoot = func(workingDir string) (repository.RepositoryRoot, error) {
		return repository.RepositoryRoot{RootPath: "/repo"}, nil
	}
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
	runWatchService = func(ctx context.Context, workingDir string, errout io.Writer) (repository.WatchRunResult, error) {
		return repository.WatchRunResult{}, nil
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
	previousResolve := runResolveRepositoryRoot
	previousHealth := runHealthService
	previousInit := runInitService
	previousRefresh := runRefreshService
	previousWatch := runWatchService
	t.Cleanup(func() {
		runCommandServer = previousServe
		runCommandInput = previousInput
		runCommandStderr = previousStderr
		runResolveRepositoryRoot = previousResolve
		runHealthService = previousHealth
		runInitService = previousInit
		runRefreshService = previousRefresh
		runWatchService = previousWatch
	})

	var calledRefresh bool
	var calledServer bool
	runCommandInput = bytes.NewBufferString("")
	runResolveRepositoryRoot = func(workingDir string) (repository.RepositoryRoot, error) {
		return repository.RepositoryRoot{RootPath: "/repo"}, nil
	}
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
	runWatchService = func(ctx context.Context, workingDir string, errout io.Writer) (repository.WatchRunResult, error) {
		return repository.WatchRunResult{}, nil
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
	previousResolve := runResolveRepositoryRoot
	t.Cleanup(func() { runResolveRepositoryRoot = previousResolve })
	runResolveRepositoryRoot = func(workingDir string) (repository.RepositoryRoot, error) {
		return repository.RepositoryRoot{}, repository.ErrRepositoryNotFound
	}

	var stdout bytes.Buffer
	err := NewRootCommand().Execute([]string{"run"}, &stdout)
	if err == nil || !strings.Contains(err.Error(), "no supported repository root found") {
		t.Fatalf("Execute(run) error = %v", err)
	}
}

func TestRunCommandRoutesWatchReportsToStderr(t *testing.T) {
	previousServe := runCommandServer
	previousInput := runCommandInput
	previousStderr := runCommandStderr
	previousResolve := runResolveRepositoryRoot
	previousHealth := runHealthService
	previousInit := runInitService
	previousRefresh := runRefreshService
	previousWatch := runWatchService
	t.Cleanup(func() {
		runCommandServer = previousServe
		runCommandInput = previousInput
		runCommandStderr = previousStderr
		runResolveRepositoryRoot = previousResolve
		runHealthService = previousHealth
		runInitService = previousInit
		runRefreshService = previousRefresh
		runWatchService = previousWatch
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runCommandInput = bytes.NewBufferString("")
	runCommandStderr = &stderr
	runResolveRepositoryRoot = func(workingDir string) (repository.RepositoryRoot, error) {
		return repository.RepositoryRoot{RootPath: "/repo"}, nil
	}
	runHealthService = func(ctx context.Context, workingDir string) (repository.HealthResult, error) {
		return repository.HealthResult{
			Repository: repository.LayeredContextEnvelope{RepositoryRoot: "/repo", Freshness: repository.FreshnessStatusFresh},
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
	runWatchService = func(ctx context.Context, workingDir string, errout io.Writer) (repository.WatchRunResult, error) {
		if errout != &stderr {
			t.Fatal("watch output should use stderr during MCP run")
		}
		_, _ = io.WriteString(errout, "watch refresh: reason=watch generation=1 freshness=fresh changed=0 unchanged=1 affected_directories=1 force_full=false error=n/a\n")
		return repository.WatchRunResult{}, nil
	}
	runCommandServer = func(ctx context.Context, repoRoot string, stdin io.Reader, serveStdout io.Writer, serveStderr io.Writer) error {
		_, _ = io.WriteString(serveStderr, "optimusctx mcp: ready for stdio requests\n")
		return nil
	}

	if err := NewRootCommand().Execute([]string{"run"}, &stdout); err != nil {
		t.Fatalf("Execute(run) error = %v", err)
	}
	if strings.Contains(stdout.String(), "watch refresh:") {
		t.Fatalf("stdout should not contain watch reports: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "watch refresh:") {
		t.Fatalf("stderr should contain watch report: %q", stderr.String())
	}
}

func TestRunCommandStartsServerBeforeInitCompletes(t *testing.T) {
	previousServe := runCommandServer
	previousInput := runCommandInput
	previousStderr := runCommandStderr
	previousResolve := runResolveRepositoryRoot
	previousHealth := runHealthService
	previousInit := runInitService
	previousRefresh := runRefreshService
	previousWatch := runWatchService
	t.Cleanup(func() {
		runCommandServer = previousServe
		runCommandInput = previousInput
		runCommandStderr = previousStderr
		runResolveRepositoryRoot = previousResolve
		runHealthService = previousHealth
		runInitService = previousInit
		runRefreshService = previousRefresh
		runWatchService = previousWatch
	})

	var order []string
	initRelease := make(chan struct{})
	runCommandInput = bytes.NewBufferString("")
	runResolveRepositoryRoot = func(workingDir string) (repository.RepositoryRoot, error) {
		return repository.RepositoryRoot{RootPath: "/repo"}, nil
	}
	runHealthService = func(ctx context.Context, workingDir string) (repository.HealthResult, error) {
		return repository.HealthResult{
			Repository: repository.LayeredContextEnvelope{RepositoryRoot: "/repo"},
			Identity:   repository.LayeredContextRepositoryIdentity{RootPath: "/repo"},
			Summary:    repository.HealthSummary{StateStatus: repository.HealthStateStatusMissing},
		}, nil
	}
	runInitService = func(ctx context.Context, workingDir string) (app.InitResult, error) {
		order = append(order, "init-start")
		<-initRelease
		order = append(order, "init-done")
		return app.InitResult{RepositoryRoot: "/repo"}, nil
	}
	runRefreshService = func(ctx context.Context, workingDir string, reason repository.RefreshReason) (app.RefreshResult, error) {
		t.Fatal("refresh should not be called when state is missing")
		return app.RefreshResult{}, nil
	}
	runWatchService = func(ctx context.Context, workingDir string, errout io.Writer) (repository.WatchRunResult, error) {
		order = append(order, "watch-start")
		return repository.WatchRunResult{}, nil
	}
	runCommandServer = func(ctx context.Context, repoRoot string, stdin io.Reader, serveStdout io.Writer, serveStderr io.Writer) error {
		order = append(order, "server-start")
		close(initRelease)
		return nil
	}

	var stdout bytes.Buffer
	if err := NewRootCommand().Execute([]string{"run"}, &stdout); err != nil {
		t.Fatalf("Execute(run) error = %v", err)
	}

	if len(order) < 3 {
		t.Fatalf("unexpected order = %v", order)
	}
	serverIndex := indexOfRunOrder(order, "server-start")
	initDoneIndex := indexOfRunOrder(order, "init-done")
	if serverIndex == -1 || initDoneIndex == -1 {
		t.Fatalf("unexpected order = %v", order)
	}
	if serverIndex > initDoneIndex {
		t.Fatalf("server should start before init completes: %v", order)
	}
}

func TestRunCommandStartsServerBeforeRefreshCompletes(t *testing.T) {
	previousServe := runCommandServer
	previousInput := runCommandInput
	previousStderr := runCommandStderr
	previousResolve := runResolveRepositoryRoot
	previousHealth := runHealthService
	previousInit := runInitService
	previousRefresh := runRefreshService
	previousWatch := runWatchService
	t.Cleanup(func() {
		runCommandServer = previousServe
		runCommandInput = previousInput
		runCommandStderr = previousStderr
		runResolveRepositoryRoot = previousResolve
		runHealthService = previousHealth
		runInitService = previousInit
		runRefreshService = previousRefresh
		runWatchService = previousWatch
	})

	var order []string
	refreshRelease := make(chan struct{})
	runCommandInput = bytes.NewBufferString("")
	runResolveRepositoryRoot = func(workingDir string) (repository.RepositoryRoot, error) {
		return repository.RepositoryRoot{RootPath: "/repo"}, nil
	}
	runHealthService = func(ctx context.Context, workingDir string) (repository.HealthResult, error) {
		return repository.HealthResult{
			Repository: repository.LayeredContextEnvelope{RepositoryRoot: "/repo", Freshness: repository.FreshnessStatusStale},
			Identity:   repository.LayeredContextRepositoryIdentity{RootPath: "/repo"},
			Summary:    repository.HealthSummary{StateStatus: repository.HealthStateStatusReady},
		}, nil
	}
	runInitService = func(ctx context.Context, workingDir string) (app.InitResult, error) {
		t.Fatal("init should not be called for existing state")
		return app.InitResult{}, nil
	}
	runRefreshService = func(ctx context.Context, workingDir string, reason repository.RefreshReason) (app.RefreshResult, error) {
		order = append(order, "refresh-start")
		<-refreshRelease
		order = append(order, "refresh-done")
		return app.RefreshResult{RepositoryRoot: "/repo", FreshnessStatus: repository.FreshnessStatusFresh}, nil
	}
	runWatchService = func(ctx context.Context, workingDir string, errout io.Writer) (repository.WatchRunResult, error) {
		order = append(order, "watch-start")
		return repository.WatchRunResult{}, nil
	}
	runCommandServer = func(ctx context.Context, repoRoot string, stdin io.Reader, serveStdout io.Writer, serveStderr io.Writer) error {
		order = append(order, "server-start")
		close(refreshRelease)
		return nil
	}

	var stdout bytes.Buffer
	if err := NewRootCommand().Execute([]string{"run"}, &stdout); err != nil {
		t.Fatalf("Execute(run) error = %v", err)
	}

	if len(order) < 3 {
		t.Fatalf("unexpected order = %v", order)
	}
	serverIndex := indexOfRunOrder(order, "server-start")
	refreshDoneIndex := indexOfRunOrder(order, "refresh-done")
	if serverIndex == -1 || refreshDoneIndex == -1 {
		t.Fatalf("unexpected order = %v", order)
	}
	if serverIndex > refreshDoneIndex {
		t.Fatalf("server should start before refresh completes: %v", order)
	}
}

func TestRunCommandStartsServerBeforeHealthCompletes(t *testing.T) {
	previousServe := runCommandServer
	previousInput := runCommandInput
	previousStderr := runCommandStderr
	previousResolve := runResolveRepositoryRoot
	previousHealth := runHealthService
	previousInit := runInitService
	previousRefresh := runRefreshService
	previousWatch := runWatchService
	t.Cleanup(func() {
		runCommandServer = previousServe
		runCommandInput = previousInput
		runCommandStderr = previousStderr
		runResolveRepositoryRoot = previousResolve
		runHealthService = previousHealth
		runInitService = previousInit
		runRefreshService = previousRefresh
		runWatchService = previousWatch
	})

	var order []string
	healthRelease := make(chan struct{})
	runCommandInput = bytes.NewBufferString("")
	runResolveRepositoryRoot = func(workingDir string) (repository.RepositoryRoot, error) {
		return repository.RepositoryRoot{RootPath: "/repo"}, nil
	}
	runHealthService = func(ctx context.Context, workingDir string) (repository.HealthResult, error) {
		order = append(order, "health-start")
		<-healthRelease
		order = append(order, "health-done")
		return repository.HealthResult{
			Repository: repository.LayeredContextEnvelope{RepositoryRoot: "/repo", Freshness: repository.FreshnessStatusFresh},
			Summary:    repository.HealthSummary{StateStatus: repository.HealthStateStatusReady},
		}, nil
	}
	runInitService = func(ctx context.Context, workingDir string) (app.InitResult, error) {
		t.Fatal("init should not be called for healthy state")
		return app.InitResult{}, nil
	}
	runRefreshService = func(ctx context.Context, workingDir string, reason repository.RefreshReason) (app.RefreshResult, error) {
		t.Fatal("refresh should not be called for healthy state")
		return app.RefreshResult{}, nil
	}
	runWatchService = func(ctx context.Context, workingDir string, errout io.Writer) (repository.WatchRunResult, error) {
		order = append(order, "watch-start")
		return repository.WatchRunResult{}, nil
	}
	runCommandServer = func(ctx context.Context, repoRoot string, stdin io.Reader, serveStdout io.Writer, serveStderr io.Writer) error {
		order = append(order, "server-start")
		close(healthRelease)
		return nil
	}

	var stdout bytes.Buffer
	if err := NewRootCommand().Execute([]string{"run"}, &stdout); err != nil {
		t.Fatalf("Execute(run) error = %v", err)
	}

	serverIndex := indexOfRunOrder(order, "server-start")
	healthDoneIndex := indexOfRunOrder(order, "health-done")
	if serverIndex == -1 || healthDoneIndex == -1 {
		t.Fatalf("unexpected order = %v", order)
	}
	if serverIndex > healthDoneIndex {
		t.Fatalf("server should start before health completes: %v", order)
	}
}

func indexOfRunOrder(order []string, want string) int {
	for i, got := range order {
		if got == want {
			return i
		}
	}
	return -1
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
