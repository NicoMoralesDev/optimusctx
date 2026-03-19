package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

func TestInitCommandClientPreview(t *testing.T) {
	repoRoot := initCLIRepo(t)
	previous := initInstallService
	t.Cleanup(func() { initInstallService = previous })
	initInstallService = func(ctx context.Context, request app.InstallRequest) (app.InstallResult, error) {
		if request.ClientID != "claude-desktop" {
			t.Fatalf("client = %q", request.ClientID)
		}
		return app.InstallResult{Rendered: repository.RenderedClientConfig{
			Client: repository.SupportedClients()[0],
			ConfigPath: "/tmp/claude.json",
			Mode: repository.RenderModePreview,
			Content: "{\n  \"mcpServers\": {\n    \"optimusctx\": {\n      \"command\": \"optimusctx\",\n      \"args\": [\"run\"]\n    }\n  }\n}\n",
		}}, nil
	}

	withWorkingDirectory(t, repoRoot, func() {
		writeCLIFile(t, repoRoot+"/main.go", "package main\n")
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init", "--client", "claude-desktop"}, &stdout); err != nil {
			t.Fatalf("Execute(init --client) error = %v", err)
		}
		output := stdout.String()
		for _, want := range []string{"client: Claude Desktop", "config path: /tmp/claude.json", "mode: preview", "\"args\": [\"run\"]", "status: preview only"} {
			if !strings.Contains(output, want) {
				t.Fatalf("missing %q in:\n%s", want, output)
			}
		}
	})
}
