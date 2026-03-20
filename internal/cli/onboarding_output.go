package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

func writeOnboardingResult(stdout io.Writer, request app.InstallRequest, result app.InstallResult) error {
	if _, err := fmt.Fprintf(stdout, "\nclient: %s\nconfig path: %s\nmode: %s\n\n%s", result.Rendered.Client.DisplayName, result.Rendered.ConfigPath, result.Rendered.Mode, ensureTrailingNewline(result.Rendered.Content)); err != nil {
		return err
	}
	for _, note := range result.Rendered.Notes {
		if _, err := fmt.Fprintf(stdout, "note: %s\n", note); err != nil {
			return err
		}
	}
	if result.Wrote {
		if _, err := io.WriteString(stdout, "status: wrote config\n"); err != nil {
			return err
		}
		_, err := fmt.Fprintf(stdout, "next step: use the registered %s MCP setup with `optimusctx run`\n", result.Rendered.Client.DisplayName)
		return err
	}
	if _, err := io.WriteString(stdout, "status: preview only\n"); err != nil {
		return err
	}
	_, err := fmt.Fprintf(stdout, "next step: rerun `%s` to apply this setup, then use `optimusctx run`\n", renderInitWriteCommand(request))
	return err
}

func renderDefaultInitNextStep() string {
	return "next step: use `optimusctx init --client <client> [--write]` for claude-desktop, claude-cli, codex-app, or codex-cli when you're ready, then use `optimusctx run`\n"
}

func renderInitWriteCommand(request app.InstallRequest) string {
	args := []string{"optimusctx", "init"}
	if strings.TrimSpace(request.ClientID) != "" {
		args = append(args, "--client", request.ClientID)
	}
	if strings.TrimSpace(request.Scope) != "" {
		args = append(args, "--scope", request.Scope)
	}
	if strings.TrimSpace(request.ConfigPath) != "" {
		args = append(args, "--config", request.ConfigPath)
	}
	if binaryPath := strings.TrimSpace(request.BinaryPath); binaryPath != "" && binaryPath != repository.DefaultServeCommandName {
		args = append(args, "--binary", binaryPath)
	}
	args = append(args, "--write")

	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		if needsShellQuote(arg) {
			quoted = append(quoted, strconv.Quote(arg))
			continue
		}
		quoted = append(quoted, arg)
	}
	return strings.Join(quoted, " ")
}

func needsShellQuote(value string) bool {
	if value == "" {
		return true
	}
	return strings.ContainsAny(value, " \t\n\"'")
}
