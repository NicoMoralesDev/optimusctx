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
		if request.ClientID != "claude-cli" {
			t.Fatalf("client = %q", request.ClientID)
		}
		return app.InstallResult{Rendered: repository.RenderedClientConfig{
			Client: repository.SupportedClient{ID: repository.ClientClaudeCLI, DisplayName: "Claude CLI"},
			ConfigPath: "command",
			Mode: repository.RenderModePreview,
			Content: "claude mcp add --transport stdio --scope local optimusctx -- optimusctx run\n",
		}}, nil
	}

	withWorkingDirectory(t, repoRoot, func() {
		writeCLIFile(t, repoRoot+"/main.go", "package main\n")
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init", "--client", "claude-cli"}, &stdout); err != nil {
			t.Fatalf("Execute(init --client) error = %v", err)
		}
		output := stdout.String()
		for _, want := range []string{"client: Claude CLI", "config path: command", "mode: preview", "claude mcp add --transport stdio --scope local optimusctx -- optimusctx run", "status: preview only"} {
			if !strings.Contains(output, want) {
				t.Fatalf("missing %q in:\n%s", want, output)
			}
		}
		if strings.Contains(output, "runnote:") {
			t.Fatalf("preview output should keep note lines separated from command content:\n%s", output)
		}
	})
}

func TestInitCommandClaudeCLIPreviewUsesScope(t *testing.T) {
	repoRoot := initCLIRepo(t)
	previous := initInstallService
	t.Cleanup(func() { initInstallService = previous })
	initInstallService = func(ctx context.Context, request app.InstallRequest) (app.InstallResult, error) {
		if request.ClientID != "claude-cli" {
			t.Fatalf("client = %q", request.ClientID)
		}
		if request.Scope != "project" {
			t.Fatalf("scope = %q", request.Scope)
		}
		if request.Write {
			t.Fatal("preview request should not set write")
		}
		return app.InstallResult{Rendered: repository.RenderedClientConfig{
			Client:     repository.SupportedClient{ID: repository.ClientClaudeCLI, DisplayName: "Claude CLI"},
			ConfigPath: "command",
			Mode:       repository.RenderModePreview,
			Content:    "claude mcp add --transport stdio --scope project optimusctx -- optimusctx run\n",
		}}, nil
	}

	withWorkingDirectory(t, repoRoot, func() {
		writeCLIFile(t, repoRoot+"/main.go", "package main\n")
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init", "--client", "claude-cli", "--scope", "project"}, &stdout); err != nil {
			t.Fatalf("Execute(init --client claude-cli --scope project) error = %v", err)
		}
		output := stdout.String()
		for _, want := range []string{"client: Claude CLI", "config path: command", "mode: preview", "claude mcp add --transport stdio --scope project optimusctx -- optimusctx run", "status: preview only"} {
			if !strings.Contains(output, want) {
				t.Fatalf("missing %q in:\n%s", want, output)
			}
		}
	})
}
