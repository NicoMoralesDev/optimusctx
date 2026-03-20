package cli

import (
	"bytes"
	"context"
	"io"
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
			Client:     repository.SupportedClient{ID: repository.ClientClaudeCLI, DisplayName: "Claude CLI"},
			ConfigPath: "command",
			Mode:       repository.RenderModePreview,
			Content:    "claude mcp add --transport stdio --scope local optimusctx -- optimusctx run\n",
		}}, nil
	}

	withWorkingDirectory(t, repoRoot, func() {
		writeCLIFile(t, repoRoot+"/main.go", "package main\n")
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init", "--client", "claude-cli"}, &stdout); err != nil {
			t.Fatalf("Execute(init --client) error = %v", err)
		}
		output := stdout.String()
		for _, want := range []string{"client: Claude CLI", "config path: command", "mode: preview", "claude mcp add --transport stdio --scope local optimusctx -- optimusctx run", "status: preview only", "next step: rerun `optimusctx init --client claude-cli --write` to apply this setup, then use `optimusctx run`"} {
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
		for _, want := range []string{"client: Claude CLI", "config path: command", "mode: preview", "claude mcp add --transport stdio --scope project optimusctx -- optimusctx run", "status: preview only", "next step: rerun `optimusctx init --client claude-cli --scope project --write` to apply this setup, then use `optimusctx run`"} {
			if !strings.Contains(output, want) {
				t.Fatalf("missing %q in:\n%s", want, output)
			}
		}
	})
}

func TestInitCommandInteractiveSkipOnboarding(t *testing.T) {
	repoRoot := initCLIRepo(t)
	previousInstall := initInstallService
	previousPrompt := initShouldPrompt
	previousInput := initPromptInput
	t.Cleanup(func() {
		initInstallService = previousInstall
		initShouldPrompt = previousPrompt
		initPromptInput = previousInput
	})
	initShouldPrompt = func(io.Writer) bool { return true }
	initPromptInput = strings.NewReader("\n")
	initInstallService = func(context.Context, app.InstallRequest) (app.InstallResult, error) {
		t.Fatal("interactive skip should not call install service")
		return app.InstallResult{}, nil
	}

	withWorkingDirectory(t, repoRoot, func() {
		writeCLIFile(t, repoRoot+"/main.go", "package main\n")
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init"}, &stdout); err != nil {
			t.Fatalf("Execute(init) error = %v", err)
		}
		output := stdout.String()
		for _, want := range []string{
			"set up a supported MCP client now?",
			"1. Claude Desktop",
			"4. Codex CLI",
			"next step: use `optimusctx init --client <client> [--write]` for claude-desktop, claude-cli, codex-app, or codex-cli when you're ready, then use `optimusctx run`",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("missing %q in:\n%s", want, output)
			}
		}
	})
}

func TestInitCommandInteractiveChoosesClientPreview(t *testing.T) {
	repoRoot := initCLIRepo(t)
	previousInstall := initInstallService
	previousPrompt := initShouldPrompt
	previousInput := initPromptInput
	t.Cleanup(func() {
		initInstallService = previousInstall
		initShouldPrompt = previousPrompt
		initPromptInput = previousInput
	})
	initShouldPrompt = func(io.Writer) bool { return true }
	initPromptInput = strings.NewReader("3\n1\n")
	initInstallService = func(ctx context.Context, request app.InstallRequest) (app.InstallResult, error) {
		if request.ClientID != "codex-app" {
			t.Fatalf("client = %q", request.ClientID)
		}
		if request.Write {
			t.Fatal("preview path should not set write")
		}
		return app.InstallResult{Rendered: repository.RenderedClientConfig{
			Client:     repository.SupportedClient{ID: repository.ClientCodexApp, DisplayName: "Codex App"},
			ConfigPath: "/home/test/.codex/config.toml",
			Mode:       repository.RenderModePreview,
			Content:    "[mcp_servers.optimusctx]\ncommand = \"optimusctx\"\nargs = [\"run\"]\n",
		}}, nil
	}

	withWorkingDirectory(t, repoRoot, func() {
		writeCLIFile(t, repoRoot+"/main.go", "package main\n")
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init"}, &stdout); err != nil {
			t.Fatalf("Execute(init) error = %v", err)
		}
		output := stdout.String()
		for _, want := range []string{
			"Choose [1-4, Enter to skip]:",
			"How should OptimusCtx continue?",
			"client: Codex App",
			"mode: preview",
			"status: preview only",
			"next step: rerun `optimusctx init --client codex-app --write` to apply this setup, then use `optimusctx run`",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("missing %q in:\n%s", want, output)
			}
		}
	})
}

func TestInitCommandInteractiveClaudeCLIWriteUsesScope(t *testing.T) {
	repoRoot := initCLIRepo(t)
	previousInstall := initInstallService
	previousPrompt := initShouldPrompt
	previousInput := initPromptInput
	t.Cleanup(func() {
		initInstallService = previousInstall
		initShouldPrompt = previousPrompt
		initPromptInput = previousInput
	})
	initShouldPrompt = func(io.Writer) bool { return true }
	initPromptInput = strings.NewReader("2\n2\n2\n")
	initInstallService = func(ctx context.Context, request app.InstallRequest) (app.InstallResult, error) {
		if request.ClientID != "claude-cli" {
			t.Fatalf("client = %q", request.ClientID)
		}
		if request.Scope != "project" {
			t.Fatalf("scope = %q", request.Scope)
		}
		if !request.Write {
			t.Fatal("interactive write path should set write")
		}
		return app.InstallResult{Rendered: repository.RenderedClientConfig{
			Client:     repository.SupportedClient{ID: repository.ClientClaudeCLI, DisplayName: "Claude CLI"},
			ConfigPath: "command",
			Mode:       repository.RenderModeWrite,
			Content:    "claude mcp add --transport stdio --scope project optimusctx -- optimusctx run\n",
		}, Wrote: true}, nil
	}

	withWorkingDirectory(t, repoRoot, func() {
		writeCLIFile(t, repoRoot+"/main.go", "package main\n")
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init"}, &stdout); err != nil {
			t.Fatalf("Execute(init) error = %v", err)
		}
		output := stdout.String()
		for _, want := range []string{
			"Claude CLI scope:",
			"client: Claude CLI",
			"mode: write",
			"status: wrote config",
			"next step: use the registered Claude CLI MCP setup with `optimusctx run`",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("missing %q in:\n%s", want, output)
			}
		}
	})
}
