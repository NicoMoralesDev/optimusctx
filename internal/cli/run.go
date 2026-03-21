package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/mcp"
	"github.com/niccrow/optimusctx/internal/repository"
)

var (
	runCommandInput          io.Reader = os.Stdin
	runCommandStderr         io.Writer = os.Stderr
	runResolveRepositoryRoot           = func(workingDir string) (repository.RepositoryRoot, error) {
		return repository.NewLocator().ResolveWithoutFingerprint(workingDir)
	}
	runCommandServer = func(ctx context.Context, repoRoot string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
		return mcp.ServeStdioWithObserver(ctx, stdin, stdout, stderr, app.MCPServerObserver{
			RepoRoot: repoRoot,
			Store:    app.NewMCPActivityStore(),
		})
	}
	runHealthService = func(ctx context.Context, rootPath string) (repository.HealthResult, error) {
		return app.NewHealthService().HealthForRootPath(ctx, rootPath, repository.HealthRequest{})
	}
	runInitService = func(ctx context.Context, workingDir string) (app.InitResult, error) {
		return app.NewInitService().Init(ctx, workingDir)
	}
	runRefreshService = func(ctx context.Context, workingDir string, reason repository.RefreshReason) (app.RefreshResult, error) {
		return app.NewRefreshService().Refresh(ctx, app.RefreshRequest{StartPath: workingDir, Reason: reason})
	}
	runWatchService = func(ctx context.Context, workingDir string, errout io.Writer) (repository.WatchRunResult, error) {
		service := app.NewWatchService()
		service.ReportRefresh = func(report repository.WatchRefreshReport) {
			_, _ = io.WriteString(errout, formatWatchRefreshReport(report))
		}
		return service.RunForRootPath(ctx, workingDir, repository.WatchRequest{StartPath: workingDir})
	}
)

func newRunCommand() *Command {
	return &Command{
		Name:    "run",
		Summary: "Run the agent-facing OptimusCtx runtime over STDIO",
		Run: func(stdout io.Writer, args []string) error {
			for _, arg := range args {
				switch arg {
				case "-h", "--help":
					_, err := io.WriteString(stdout, "Usage:\n  optimusctx run\n\nRun the agent-facing OptimusCtx runtime over STDIO. This is the canonical entrypoint for MCP clients.\n")
					return err
				default:
					if len(arg) > 0 && arg[0] == '-' {
						return fmt.Errorf("run does not accept flags; got %q", arg)
					}
					return fmt.Errorf("run does not accept arguments; got %q", arg)
				}
			}

			workingDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve working directory: %w", err)
			}

			repoRoot, err := runResolveRepositoryRoot(workingDir)
			if err != nil {
				if errors.Is(err, repository.ErrRepositoryNotFound) {
					return fmt.Errorf("no supported repository root found from %s; run `optimusctx init` inside a Git repository or an existing .optimusctx state directory", workingDir)
				}
				return err
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			backgroundErr := make(chan error, 1)
			go func() {
				backgroundErr <- runRuntimeBootstrap(ctx, repoRoot.RootPath, runCommandStderr)
			}()

			serverErr := runCommandServer(ctx, repoRoot.RootPath, runCommandInput, stdout, runCommandStderr)
			cancel()
			backgroundRunErr := <-backgroundErr
			if serverErr != nil {
				return serverErr
			}
			if backgroundRunErr != nil && !errors.Is(backgroundRunErr, context.Canceled) {
				return backgroundRunErr
			}
			return nil
		},
	}
}

func runRuntimeBootstrap(ctx context.Context, rootPath string, errout io.Writer) error {
	health, err := runHealthService(ctx, rootPath)
	if err != nil {
		reportRuntimeBootstrapError(errout, err)
		return err
	}

	switch {
	case health.Summary.StateStatus == repository.HealthStateStatusMissing:
		if _, err := runInitService(ctx, rootPath); err != nil {
			reportRuntimeBootstrapError(errout, err)
			return err
		}
	case health.Repository.Freshness == repository.FreshnessStatusStale || health.Repository.Freshness == repository.FreshnessStatusPartiallyDegraded:
		if _, err := runRefreshService(ctx, rootPath, repository.RefreshReasonWatch); err != nil {
			reportRuntimeBootstrapError(errout, err)
			return err
		}
	}

	_, err = runWatchService(ctx, rootPath, errout)
	return err
}

func reportRuntimeBootstrapError(errout io.Writer, err error) {
	if errout == nil || err == nil || errors.Is(err, context.Canceled) {
		return
	}
	_, _ = fmt.Fprintf(errout, "optimusctx runtime bootstrap warning: %v\n", err)
}

func healthErrNeedsInit(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == "read state metadata: open state metadata file: open" || false
}
