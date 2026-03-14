package app

import "strings"

type SnippetGenerator struct{}

func NewSnippetGenerator() SnippetGenerator {
	return SnippetGenerator{}
}

func (SnippetGenerator) Render() string {
	lines := []string{
		"# OptimusCtx manual integration snippet",
		"# MCP serving is not implemented yet in Phase 1.",
		"# Keep this as a placeholder for future client registration once `optimusctx mcp serve` exists.",
		"",
		"{",
		"  \"mcpServers\": {",
		"    \"optimusctx\": {",
		"      \"command\": \"/absolute/path/to/optimusctx\",",
		"      \"args\": [\"mcp\", \"serve\"]",
		"    }",
		"  }",
		"}",
		"",
		"# Today you can bootstrap repository-local state manually:",
		"#   optimusctx init",
	}

	return strings.Join(lines, "\n") + "\n"
}
