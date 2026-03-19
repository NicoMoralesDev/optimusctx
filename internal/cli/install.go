package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

var installExecutablePath = os.Executable

func newInstallCommand() *Command {
	return &Command{
		Name:    "install",
		Summary: "Deprecated: preview or write supported MCP client registration",
		Run: func(stdout io.Writer, args []string) error {
			_, _ = io.WriteString(stdout, "warning: `optimusctx install` is deprecated; use `optimusctx status --client <client> [--write]` instead\n")
			return runInstallCommand(stdout, args)
		},
	}
}

func runInstallCommand(stdout io.Writer, args []string) error {
	request := app.InstallRequest{}
	binaryFlagProvided := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help":
			writeInstallHelp(stdout)
			return nil
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
			binaryFlagProvided = true
			i = next
		case "--write":
			request.Write = true
		default:
			if len(arg) > 0 && arg[0] == '-' {
				return fmt.Errorf("install does not accept flag %q", arg)
			}
			return fmt.Errorf("install does not accept argument %q", arg)
		}
	}

	if request.ClientID == "" {
		return errors.New("install requires --client")
	}

	if !binaryFlagProvided {
		binaryPath, err := installExecutablePath()
		if err != nil {
			return fmt.Errorf("resolve executable path: %w", err)
		}
		request.BinaryPath = normalizeInstallBinaryPath(binaryPath)
	}

	result, err := app.NewInstallService().Register(context.Background(), request)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "client: %s\nconfig path: %s\nmode: %s\n\n%s", result.Rendered.Client.DisplayName, result.Rendered.ConfigPath, result.Rendered.Mode, result.Rendered.Content); err != nil {
		return err
	}
	if result.Wrote {
		_, err = io.WriteString(stdout, "status: wrote config\n")
		return err
	}
	_, err = io.WriteString(stdout, "status: preview only\n")
	return err
}

func requireInstallValue(args []string, index int, flag string) (string, int, error) {
	next := index + 1
	if next >= len(args) {
		return "", index, fmt.Errorf("install flag %s requires a value", flag)
	}
	value := args[next]
	if len(value) > 0 && value[0] == '-' {
		return "", index, fmt.Errorf("install flag %s requires a value", flag)
	}
	return value, next, nil
}

func writeInstallHelp(stdout io.Writer) {
	_, _ = io.WriteString(stdout, "Usage:\n  optimusctx install --client <client> [--config <path>] [--binary <path>] [--write]\n\nDeprecated command. Prefer `optimusctx status --client <client> [--write]`.\n")
}

func normalizeInstallBinaryPath(runtimeBinaryPath string) string {
	if looksEphemeralRuntimeBinary(runtimeBinaryPath) {
		return repository.DefaultServeCommandName
	}
	return repository.CanonicalServeCommandPath("")
}

func looksEphemeralRuntimeBinary(path string) bool {
	path = strings.ReplaceAll(strings.TrimSpace(path), "\\", "/")
	return strings.Contains(path, "/.cache/go-build/") || strings.Contains(path, "/go-build/")
}
