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
	statusInstallPreviewService = func(ctx context.Context, request app.InstallRequest) (app.InstallResult, error) {
		return app.NewInstallService().Register(ctx, request)
	}
)

func newStatusCommand() *Command {
	return &Command{
		Name:    "status",
		Summary: "Show short runtime status and optional MCP client config preview",
		Run: func(stdout io.Writer, args []string) error {
			return runStatusCommand(stdout, args)
		},
	}
}

func runStatusCommand(stdout io.Writer, args []string) error {
	request := app.InstallRequest{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help":
			_, err := io.WriteString(stdout, "Usage:\n  optimusctx status [--client <client>] [--config <path>] [--binary <path>] [--write]\n\nShow the current repository/runtime status. When --client is provided, also preview or write MCP client registration.\n")
			return err
		case "--client":
			value, next, err := requireInstallValue(args, i, arg)
			if err != nil {
				return err
			}
			request.ClientID = value
			i = next
		case "--config":
			value, next, err := requireInstallValue(args, i, arg)
			if err != nil {
				return err
			}
			request.ConfigPath = value
			i = next
		case "--binary":
			value, next, err := requireInstallValue(args, i, arg)
			if err != nil {
				return err
			}
			request.BinaryPath = value
			i = next
		case "--write":
			request.Write = true
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
	_, _ = fmt.Fprintf(&b, "repository root: %s\nruntime status: %s\nstate status: %s\nfreshness: %s\ngeneration: %d\nlast refresh completed: %s\nwatch status: %s\nwatch reason: %s\nmcp command: %s %s\nnext action: %s\n",
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
		nextAction,
	)

	if request.ClientID != "" {
		if request.BinaryPath == "" {
			request.BinaryPath = repository.DefaultServeCommandName
		}
		preview, err := statusInstallPreviewService(ctx, request)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(&b, "\nclient: %s\nconfig path: %s\nmode: %s\n\n%s", preview.Rendered.Client.DisplayName, preview.Rendered.ConfigPath, preview.Rendered.Mode, preview.Rendered.Content)
		for _, note := range preview.Rendered.Notes {
			_, _ = fmt.Fprintf(&b, "note: %s\n", note)
		}
		if preview.Wrote {
			_, _ = io.WriteString(&b, "status: wrote config\n")
		} else {
			_, _ = io.WriteString(&b, "status: preview only\n")
		}
	}

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
		return "runtime is ready; point your MCP client at `optimusctx run`"
	}
}
