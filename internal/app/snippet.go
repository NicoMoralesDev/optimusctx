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
	document, err := repository.MergeClientConfig(nil, repository.DefaultMCPServerName, repository.NewServeCommand("/absolute/path/to/optimusctx"))
	if err != nil {
		panic(err)
	}
	rendered, err := repository.RenderClientConfig(document)
	if err != nil {
		panic(err)
	}

	lines := []string{
		"# OptimusCtx manual integration snippet",
		"# OptimusCtx now serves MCP over `optimusctx mcp serve`.",
		"# You can paste this into a supported client config or preview the same contract with:",
		"#   optimusctx install --client claude-desktop --config /path/to/claude_desktop_config.json",
		"",
		strings.TrimSuffix(rendered, "\n"),
		"",
		"# Write the same registration explicitly with:",
		"#   optimusctx install --client claude-desktop --config /path/to/claude_desktop_config.json --write",
	}

	return strings.Join(lines, "\n") + "\n"
}
