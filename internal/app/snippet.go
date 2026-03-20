package app

import (
	"strings"

	"github.com/niccrow/optimusctx/internal/repository"
)

type SnippetGenerator struct{}

func NewSnippetGenerator() SnippetGenerator {
	return SnippetGenerator{}
}

func (SnippetGenerator) Render() string {
	document, err := repository.MergeClientConfig(nil, repository.DefaultMCPServerName, repository.NewServeCommand(""))
	if err != nil {
		panic(err)
	}
	rendered, err := repository.RenderClientConfig(document)
	if err != nil {
		panic(err)
	}

	lines := []string{
		"# OptimusCtx manual integration snippet",
		"# Supported native clients: claude-desktop, claude-cli, codex-app, codex-cli",
		"# `optimusctx snippet` is deprecated; use `optimusctx status --client <client> [--write]` for the current supported-client surface.",
		"# OptimusCtx now serves MCP over `optimusctx run`.",
		"# You can paste this into a supported client config or preview the same contract with:",
		"#   optimusctx status --client claude-cli --scope local",
		"#   optimusctx status --client codex-app --config /path/to/.codex/config.toml",
		"#   optimusctx status --client codex-cli --config /path/to/.codex/config.toml",
		"",
		strings.TrimSuffix(rendered, "\n"),
		"",
		"# Write the same registration explicitly with:",
		"#   optimusctx status --client <client> [--write]",
		"# Supported native clients: claude-desktop, claude-cli, codex-app, codex-cli",
	}

	return strings.Join(lines, "\n") + "\n"
}
