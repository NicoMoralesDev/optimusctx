package app

import (
	"strings"
	"testing"
)

func TestSnippetGeneratorRender(t *testing.T) {
	output := NewSnippetGenerator().Render()

	required := []string{
		"# OptimusCtx manual integration snippet",
		"OptimusCtx now serves MCP over `optimusctx mcp serve`.",
		"optimusctx install --client claude-desktop",
		"\"mcpServers\"",
		"\"command\": \"optimusctx\"",
		"\"args\": [",
		"\"mcp\"",
		"\"serve\"",
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
