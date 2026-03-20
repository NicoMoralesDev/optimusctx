package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

var (
	statusHealthService = func(ctx context.Context, workingDir string) (repository.HealthResult, error) {
		return app.NewHealthService().Health(ctx, workingDir, repository.HealthRequest{})
	}
	statusWatchService = func(ctx context.Context, workingDir string) (repository.WatchStatusResult, error) {
		return app.NewWatchService().Status(ctx, workingDir, 0)
	}
)

const supportedClientList = "claude-desktop, claude-cli, codex-app, codex-cli"

func newStatusCommand() *Command {
	return &Command{
		Name:    "status",
		Summary: "Show short runtime status without mutating MCP client configuration",
		Run: func(stdout io.Writer, args []string) error {
			return runStatusCommand(stdout, args)
		},
	}
}

func runStatusCommand(stdout io.Writer, args []string) error {
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			_, err := io.WriteString(stdout, "Usage:\n  optimusctx status\n\nShow the current repository/runtime status without mutating MCP client configuration.\n")
			return err
		default:
			if len(arg) > 0 && arg[0] == '-' {
				return fmt.Errorf("status does not accept flag %q", arg)
			}
			return fmt.Errorf("status does not accept argument %q", arg)
		}
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	ctx := context.Background()
	health, err := statusHealthService(ctx, workingDir)
	if err != nil {
		return err
	}
	watch, err := statusWatchService(ctx, workingDir)
	if err != nil {
		return err
	}

	var b strings.Builder
	status := deriveShortRuntimeStatus(health, watch)
	nextAction := deriveShortNextAction(health, watch)
	_, _ = fmt.Fprintf(&b, "repository root: %s\nruntime status: %s\nstate status: %s\nfreshness: %s\ngeneration: %d\nlast refresh completed: %s\nwatch status: %s\nwatch reason: %s\nmcp command: %s %s\nsupported clients: %s\nnext action: %s\n",
		health.Repository.RepositoryRoot,
		status,
		health.Summary.StateStatus,
		renderFreshnessStatus(health.Repository.Freshness),
		health.Repository.Generation,
		formatDoctorTime(health.Refresh.LastRefreshCompletedAt),
		watch.Status,
		watch.Reason,
		repository.NewServeCommand("").Command,
		strings.Join(repository.NewServeCommand("").Args, " "),
		supportedClientList,
		nextAction,
	)

	_, err = io.WriteString(stdout, b.String())
	return err
}

func deriveShortRuntimeStatus(health repository.HealthResult, watch repository.WatchStatusResult) string {
	switch {
	case health.Summary.StateStatus == repository.HealthStateStatusMissing:
		return "not_initialized"
	case health.Repository.Freshness == repository.FreshnessStatusPartiallyDegraded || watch.Status == repository.WatchStatusKindStale:
		return "degraded"
	case watch.Status == repository.WatchStatusKindRunning:
		return "running"
	default:
		return "idle"
	}
}

func deriveShortNextAction(health repository.HealthResult, watch repository.WatchStatusResult) string {
	switch {
	case health.Summary.StateStatus == repository.HealthStateStatusMissing:
		return "run `optimusctx init` in the repository root"
	case health.Repository.Freshness == repository.FreshnessStatusStale:
		return "run `optimusctx run` to refresh and serve the runtime"
	case health.Repository.Freshness == repository.FreshnessStatusPartiallyDegraded:
		return "run `optimusctx doctor` and then `optimusctx run` to recover the runtime"
	case watch.Status == repository.WatchStatusKindStale:
		return "restart with `optimusctx run` or inspect deeper with `optimusctx doctor`"
	default:
		return "runtime is ready; rerun `optimusctx init --client <client> [--write]` for claude-desktop, claude-cli, codex-app, or codex-cli to preview or register MCP. Registered MCP clients should launch `optimusctx run` automatically, and manual `optimusctx run` remains the direct/debug path."
	}
}
