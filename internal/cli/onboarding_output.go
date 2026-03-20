package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

func writeOnboardingResult(stdout io.Writer, request app.InstallRequest, result app.InstallResult) error {
	target := describeOnboardingTarget(request, result.Rendered)
	if _, err := fmt.Fprintf(stdout, "\nclient: %s\ndestination: %s\n%s: %s\n", result.Rendered.Client.DisplayName, target.destination, target.locationLabel, target.location); err != nil {
		return err
	}
	if result.Wrote {
		if _, err := io.WriteString(stdout, "status: configured\n"); err != nil {
			return err
		}
		if _, err := io.WriteString(stdout, "runtime: your registered MCP client should launch `optimusctx run` automatically when it connects\n"); err != nil {
			return err
		}
		if _, err := io.WriteString(stdout, "verify: confirm your host exposes `optimusctx.*` tools such as `optimusctx.repository_map`, `optimusctx.symbol_lookup`, and `optimusctx.health`\n"); err != nil {
			return err
		}
		_, err := io.WriteString(stdout, "manual fallback: run `optimusctx run` yourself only for direct STDIO use or debugging\n")
		return err
	}
	if _, err := fmt.Fprintf(stdout, "\nreview this change first:\n\n%s", ensureTrailingNewline(result.Rendered.Content)); err != nil {
		return err
	}
	if _, err := io.WriteString(stdout, "status: ready to configure\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stdout, "next step: rerun `%s` to apply this setup\n", renderInitWriteCommand(request)); err != nil {
		return err
	}
	if _, err := io.WriteString(stdout, "runtime after apply: your MCP client should launch `optimusctx run` automatically when it connects\n"); err != nil {
		return err
	}
	_, err := io.WriteString(stdout, "manual fallback: run `optimusctx run` yourself only for direct STDIO use or debugging\n")
	return err
}

func renderDefaultInitNextStep() string {
	return "next step: use `optimusctx init --client <client>` to review the change for claude-desktop, claude-cli, codex-app, or codex-cli, or add `--write` to configure one right away\nruntime after registration: your MCP client should launch `optimusctx run` automatically when it connects\nmanual fallback: run `optimusctx run` yourself only for direct STDIO use or debugging\n"
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

type onboardingTarget struct {
	destination   string
	locationLabel string
	location      string
}

func describeOnboardingTarget(request app.InstallRequest, rendered repository.RenderedClientConfig) onboardingTarget {
	switch rendered.Client.ID {
	case repository.ClientClaudeDesktop:
		return onboardingTarget{
			destination:   "Claude Desktop app config",
			locationLabel: "config path",
			location:      rendered.ConfigPath,
		}
	case repository.ClientClaudeCLI:
		scope := strings.TrimSpace(request.Scope)
		if scope == "" {
			scope = repository.ClaudeCLIScopeLocal
		}
		label := map[string]string{
			repository.ClaudeCLIScopeLocal:   "Your current Claude setup",
			repository.ClaudeCLIScopeProject: "This project",
			repository.ClaudeCLIScopeUser:    "Your Claude user profile",
		}[scope]
		if label == "" {
			label = "Claude CLI"
		}
		return onboardingTarget{
			destination:   label,
			locationLabel: "native target",
			location:      fmt.Sprintf("claude mcp add --scope %s", scope),
		}
	case repository.ClientCodexApp, repository.ClientCodexCLI:
		repoLocalPath := ""
		if strings.TrimSpace(request.RepoRoot) != "" {
			repoLocalPath = filepath.Join(request.RepoRoot, ".codex", "config.toml")
		}
		destination := "Your shared Codex config"
		if repoLocalPath != "" && rendered.ConfigPath == repoLocalPath {
			destination = "This repo only"
		}
		return onboardingTarget{
			destination:   destination,
			locationLabel: "config path",
			location:      rendered.ConfigPath,
		}
	default:
		return onboardingTarget{
			destination:   "Manual MCP host config",
			locationLabel: "config path",
			location:      rendered.ConfigPath,
		}
	}
}
