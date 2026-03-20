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

var initInstallService = func(ctx context.Context, request app.InstallRequest) (app.InstallResult, error) {
	return app.NewInstallService().Register(ctx, request)
}

func newInitCommand() *Command {
	return &Command{
		Name:    "init",
		Summary: "Initialize repository-local OptimusCtx state",
		Run: func(stdout io.Writer, args []string) error {
			request := app.InstallRequest{}
			for i := 0; i < len(args); i++ {
				arg := args[i]
				switch arg {
				case "-h", "--help":
					_, err := io.WriteString(stdout, "Usage:\n  optimusctx init [--client <client>] [--config <path>] [--binary <path>] [--scope <local|project|user>] [--write]\n\nInitialize repository-local OptimusCtx state. When --client is provided, also preview or write MCP client registration as part of onboarding.\n")
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
				case "--scope":
					value, next, err := requireInstallValue(args, i, arg)
					if err != nil {
						return err
					}
					request.Scope = value
					i = next
				case "--write":
					request.Write = true
				default:
					if len(arg) > 0 && arg[0] == '-' {
						return fmt.Errorf("init does not accept flag %q", arg)
					}
					return fmt.Errorf("init does not accept argument %q", arg)
				}
			}

			workingDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve working directory: %w", err)
			}

			result, err := app.NewInitService().Init(context.Background(), workingDir)
			if err != nil {
				if err == repository.ErrRepositoryNotFound || (err != nil && strings.Contains(err.Error(), repository.ErrRepositoryNotFound.Error())) {
					return fmt.Errorf("no supported repository root found from %s; run `optimusctx init` inside a Git repository or an existing .optimusctx state directory", workingDir)
				}
				return err
			}

			if _, err := fmt.Fprintf(
				stdout,
				"repository root: %s\nstate directory: %s\nschema version: %d\nrefresh generation: %d\nfreshness: %s\ndiscovered files: %d\n",
				result.RepositoryRoot,
				result.StatePath,
				result.SchemaVersion,
				result.Generation,
				renderFreshnessStatus(result.FreshnessStatus),
				result.FileCount,
			); err != nil {
				return err
			}

			if request.ClientID == "" {
				_, err = io.WriteString(stdout, "\nnext step: rerun `optimusctx init --client <client> [--write]` for claude-desktop, claude-cli, codex-app, or codex-cli to preview or register MCP during onboarding, then use `optimusctx run`\n")
				return err
			}

			if request.BinaryPath == "" {
				request.BinaryPath = repository.DefaultServeCommandName
			}
			installResult, err := initInstallService(context.Background(), request)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(stdout, "\nclient: %s\nconfig path: %s\nmode: %s\n\n%s", installResult.Rendered.Client.DisplayName, installResult.Rendered.ConfigPath, installResult.Rendered.Mode, ensureTrailingNewline(installResult.Rendered.Content)); err != nil {
				return err
			}
			for _, note := range installResult.Rendered.Notes {
				if _, err := fmt.Fprintf(stdout, "note: %s\n", note); err != nil {
					return err
				}
			}
			if installResult.Wrote {
				_, err = io.WriteString(stdout, "status: wrote config\n")
				return err
			}
			_, err = io.WriteString(stdout, "status: preview only\n")
			return err
		},
	}
}
