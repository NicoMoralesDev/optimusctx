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
	runCommandInput  io.Reader = os.Stdin
	runCommandStderr io.Writer = os.Stderr
	runCommandServer           = func(ctx context.Context, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
		return mcp.ServeStdio(ctx, stdin, stdout, stderr)
	}
	runHealthService = func(ctx context.Context, workingDir string) (repository.HealthResult, error) {
		return app.NewHealthService().Health(ctx, workingDir, repository.HealthRequest{})
	}
	runInitService = func(ctx context.Context, workingDir string) (app.InitResult, error) {
		return app.NewInitService().Init(ctx, workingDir)
	}
	runRefreshService = func(ctx context.Context, workingDir string, reason repository.RefreshReason) (app.RefreshResult, error) {
		return app.NewRefreshService().Refresh(ctx, app.RefreshRequest{StartPath: workingDir, Reason: reason})
	}
	runWatchService = func(ctx context.Context, workingDir string, stdout io.Writer) (repository.WatchRunResult, error) {
		service := app.NewWatchService()
		service.ReportRefresh = func(report repository.WatchRefreshReport) {
			_, _ = io.WriteString(stdout, formatWatchRefreshReport(report))
		}
		return service.Run(ctx, repository.WatchRequest{StartPath: workingDir})
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

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			health, err := runHealthService(ctx, workingDir)
			if err != nil {
				if errors.Is(err, repository.ErrRepositoryNotFound) {
					return fmt.Errorf("no supported repository root found from %s; run `optimusctx init` inside a Git repository or an existing .optimusctx state directory", workingDir)
				}
				return err
			}

			if health.Summary.StateStatus == repository.HealthStateStatusMissing {
				if _, err := runInitService(ctx, workingDir); err != nil {
					return err
				}
			} else if health.Repository.Freshness == repository.FreshnessStatusStale || health.Repository.Freshness == repository.FreshnessStatusPartiallyDegraded {
				if _, err := runRefreshService(ctx, workingDir, repository.RefreshReasonWatch); err != nil {
					return err
				}
			}

			watchErr := make(chan error, 1)
			go func() {
				_, err := runWatchService(ctx, workingDir, stdout)
				watchErr <- err
			}()

			serverErr := runCommandServer(ctx, runCommandInput, stdout, runCommandStderr)
			cancel()
			watchRunErr := <-watchErr
			if serverErr != nil {
				return serverErr
			}
			if watchRunErr != nil && !errors.Is(watchRunErr, context.Canceled) {
				return watchRunErr
			}
			return nil
		},
	}
}

func healthErrNeedsInit(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == "read state metadata: open state metadata file: open" || false
}
