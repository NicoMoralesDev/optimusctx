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
		"# `optimusctx snippet` is deprecated; use `optimusctx status --client claude-desktop` for the current preview path.",
		"# OptimusCtx now serves MCP over `optimusctx run`.",
		"# You can paste this into a supported client config or preview the same contract with:",
		"#   optimusctx status --client claude-desktop --config /path/to/claude_desktop_config.json",
		"",
		strings.TrimSuffix(rendered, "\n"),
		"",
		"# Write the same registration explicitly with:",
		"#   optimusctx status --client claude-desktop --config /path/to/claude_desktop_config.json --write",
	}

	return strings.Join(lines, "\n") + "\n"
}
