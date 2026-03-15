package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

var (
	watchRunCommandService = func(ctx context.Context, workingDir string) (repository.WatchRunResult, error) {
		return app.NewWatchService().Run(ctx, repository.WatchRequest{StartPath: workingDir})
	}
	watchStatusCommandService = func(ctx context.Context, workingDir string) (repository.WatchStatusResult, error) {
		return app.NewWatchService().Status(ctx, workingDir, 0)
	}
)

func newWatchCommand() *Command {
	return &Command{
		Name:    "watch",
		Summary: "Run or inspect the optional repository watch process",
		Run: func(stdout io.Writer, args []string) error {
			if len(args) == 0 {
				writeWatchHelp(stdout)
				return errors.New("watch requires a subcommand")
			}

			switch args[0] {
			case "-h", "--help", "help":
				writeWatchHelp(stdout)
				return nil
			case "run":
				return runWatchRunCommand(stdout, args[1:])
			case "status":
				return runWatchStatusCommand(stdout, args[1:])
			default:
				writeWatchHelp(stdout)
				return fmt.Errorf("unknown watch subcommand %q", args[0])
			}
		},
	}
}

func runWatchRunCommand(stdout io.Writer, args []string) error {
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			_, err := io.WriteString(stdout, "Usage:\n  optimusctx watch run\n\nRun the optional repository watch process in the foreground.\n")
			return err
		default:
			if len(arg) > 0 && arg[0] == '-' {
				return fmt.Errorf("watch run does not accept flags; got %q", arg)
			}
			return fmt.Errorf("watch run does not accept arguments; got %q", arg)
		}
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	_, err = watchRunCommandService(ctx, workingDir)
	return err
}

func runWatchStatusCommand(stdout io.Writer, args []string) error {
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			_, err := io.WriteString(stdout, "Usage:\n  optimusctx watch status\n\nShow repo-scoped watch liveness from ephemeral state.\n")
			return err
		default:
			if len(arg) > 0 && arg[0] == '-' {
				return fmt.Errorf("watch status does not accept flags; got %q", arg)
			}
			return fmt.Errorf("watch status does not accept arguments; got %q", arg)
		}
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	result, err := watchStatusCommandService(context.Background(), workingDir)
	if err != nil {
		return err
	}

	_, err = io.WriteString(stdout, formatWatchStatus(result))
	return err
}

func writeWatchHelp(stdout io.Writer) {
	_, _ = io.WriteString(stdout, "Usage:\n  optimusctx watch <command>\n\nAvailable Commands:\n  run      Run the optional repository watch process in the foreground\n  status   Show repo-scoped watch liveness from ephemeral state\n")
}

func formatWatchStatus(result repository.WatchStatusResult) string {
	record := result.Record
	return fmt.Sprintf(
		"repository root: %s\nstatus: %s\nreason: %s\nstatus path: %s\npid: %d\nstarted at: %s\nlast heartbeat: %s\nlast event: %s\nlast refresh completed: %s\nlast refresh generation: %d\nlast error: %s\n",
		result.RepositoryRoot,
		result.Status,
		result.Reason,
		result.StatusPath,
		record.PID,
		renderWatchValue(record.StartedAt),
		renderWatchValue(record.LastHeartbeatAt),
		renderWatchValue(record.LastEventAt),
		renderWatchValue(record.LastRefreshDoneAt),
		record.LastRefreshGeneration,
		renderWatchValue(record.LastError),
	)
}

func renderWatchValue(value string) string {
	if value == "" {
		return "n/a"
	}
	return value
}
