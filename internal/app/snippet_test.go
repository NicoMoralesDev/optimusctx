package app

import (
	"strings"
	"testing"
)

func TestSnippetRender(t *testing.T) {
	output := NewSnippetGenerator().Render()

	required := []string{
		"# OptimusCtx manual integration snippet",
		"# Supported native clients: claude-desktop, claude-cli, codex-app, codex-cli, gemini-cli",
		"`optimusctx snippet` is deprecated; use `optimusctx init --client <client> [--write]`",
		"OptimusCtx now serves MCP over `optimusctx run`.",
		"optimusctx init --client claude-cli --scope local",
		"optimusctx init --client codex-app --config /path/to/.codex/config.toml",
		"optimusctx init --client codex-cli --config /path/to/.codex/config.toml",
		"optimusctx init --client gemini-cli --config /path/to/.gemini/settings.json",
		"optimusctx init --client <client> [--write]",
		"\"mcpServers\"",
		"\"command\": \"optimusctx\"",
		"\"args\": [",
		"\"run\"",
		"\"optimusctx\": {",
	}
	for _, fragment := range required {
		if !strings.Contains(output, fragment) {
			t.Fatalf("Render() output missing %q: %q", fragment, output)
		}
	}

	if !strings.HasSuffix(output, "\n") {
		t.Fatalf("Render() should end with newline: %q", output)
	}
	if strings.Contains(output, "use `optimusctx status --client claude-desktop` for the current preview path") {
		t.Fatalf("Render() output should not keep the Claude Desktop-only preview path: %q", output)
	}
	if strings.Contains(output, "optimusctx status --client <client> [--write]") {
		t.Fatalf("Render() output should not keep the stale status-led onboarding path: %q", output)
	}
}

func TestSnippetInstallCommandAlignment(t *testing.T) {
	output := NewSnippetGenerator().Render()

	if strings.Contains(output, "/absolute/path/to/optimusctx") {
		t.Fatalf("Render() output should not contain placeholder path: %q", output)
	}
	if !strings.Contains(output, "\"command\": \"optimusctx\"") {
		t.Fatalf("Render() output missing canonical command: %q", output)
	}
}
