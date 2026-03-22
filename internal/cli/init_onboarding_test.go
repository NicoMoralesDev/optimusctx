package cli

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
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
		for _, want := range []string{
			"client: Claude CLI",
			"destination: Your current Claude setup",
			"native target: claude mcp add --scope local",
			"review this change first:",
			"claude mcp add --transport stdio --scope local optimusctx -- optimusctx run",
			"status: ready to configure",
			"next step: rerun `optimusctx init --client claude-cli --write` to apply this setup",
			"runtime after apply: your MCP client should launch `optimusctx run` automatically when it connects",
			"manual fallback: run `optimusctx run` yourself only for direct STDIO use or debugging",
		} {
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
		for _, want := range []string{
			"client: Claude CLI",
			"destination: This project",
			"native target: claude mcp add --scope project",
			"review this change first:",
			"claude mcp add --transport stdio --scope project optimusctx -- optimusctx run",
			"status: ready to configure",
			"next step: rerun `optimusctx init --client claude-cli --scope project --write` to apply this setup",
			"runtime after apply: your MCP client should launch `optimusctx run` automatically when it connects",
		} {
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
			"next step: use `optimusctx init --client <client>` to review the change for claude-desktop, claude-cli, codex-app, or codex-cli, or add `--write` to configure one right away",
			"runtime after registration: your MCP client should launch `optimusctx run` automatically when it connects",
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
	initPromptInput = strings.NewReader("3\n\n2\n")
	initInstallService = func(ctx context.Context, request app.InstallRequest) (app.InstallResult, error) {
		if request.ClientID != "codex-app" {
			t.Fatalf("client = %q", request.ClientID)
		}
		if request.ConfigPath != filepath.Join(repoRoot, ".codex", "config.toml") {
			t.Fatalf("config path = %q", request.ConfigPath)
		}
		if request.Write {
			t.Fatal("review-first path should not set write")
		}
		return app.InstallResult{Rendered: repository.RenderedClientConfig{
			Client:     repository.SupportedClient{ID: repository.ClientCodexApp, DisplayName: "Codex App"},
			ConfigPath: filepath.Join(repoRoot, ".codex", "config.toml"),
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
			"Where should Codex App use OptimusCtx?",
			filepath.Join(repoRoot, ".codex", "config.toml"),
			"How should OptimusCtx continue?",
			"client: Codex App",
			"destination: This repo only",
			"review this change first:",
			"status: ready to configure",
			"next step: rerun `optimusctx init --client codex-app --config " + filepath.Join(repoRoot, ".codex", "config.toml") + " --write` to apply this setup",
			"runtime after apply: your MCP client should launch `optimusctx run` automatically when it connects",
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
	initPromptInput = strings.NewReader("2\n\n1\n")
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
			"Where should Claude CLI register OptimusCtx?",
			"client: Claude CLI",
			"destination: This project",
			"native target: claude mcp add --scope project",
			"status: wrote config",
			"runtime after host pickup: your MCP client should launch `optimusctx run` automatically when it connects",
			"verify host pickup with `optimusctx status`: it should eventually show registration evidence, last MCP initialize/tools discovery, and recent `optimusctx.*` tool calls",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("missing %q in:\n%s", want, output)
			}
		}
	})
}

func TestInitCommandInteractiveCodexCLISharedConfigChoice(t *testing.T) {
	repoRoot := initCLIRepo(t)
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	sharedPath := filepath.Join(homeDir, ".codex", "config.toml")

	previousInstall := initInstallService
	previousPrompt := initShouldPrompt
	previousInput := initPromptInput
	t.Cleanup(func() {
		initInstallService = previousInstall
		initShouldPrompt = previousPrompt
		initPromptInput = previousInput
	})
	initShouldPrompt = func(io.Writer) bool { return true }
	initPromptInput = strings.NewReader("4\n2\n1\n")
	initInstallService = func(ctx context.Context, request app.InstallRequest) (app.InstallResult, error) {
		if request.ClientID != "codex-cli" {
			t.Fatalf("client = %q", request.ClientID)
		}
		if request.ConfigPath != "" {
			t.Fatalf("shared config path should stay implicit, got %q", request.ConfigPath)
		}
		if !request.Write {
			t.Fatal("configure-now path should set write")
		}
		return app.InstallResult{Rendered: repository.RenderedClientConfig{
			Client:     repository.SupportedClient{ID: repository.ClientCodexCLI, DisplayName: "Codex CLI"},
			ConfigPath: sharedPath,
			Mode:       repository.RenderModeWrite,
			Content:    "[mcp_servers.optimusctx]\ncommand = \"optimusctx\"\nargs = [\"run\"]\n",
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
			"Where should Codex CLI use OptimusCtx?",
			sharedPath,
			"destination: Your shared Codex config",
			"config path: " + sharedPath,
			"status: wrote config",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("missing %q in:\n%s", want, output)
			}
		}
	})
}

func TestInitCommandInteractiveCodexAppSharedConfigChoice(t *testing.T) {
	repoRoot := initCLIRepo(t)
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	sharedPath := filepath.Join(homeDir, ".codex", "config.toml")
	t.Setenv("OPTIMUSCTX_CODEX_APP_CONFIG", sharedPath)

	previousInstall := initInstallService
	previousPrompt := initShouldPrompt
	previousInput := initPromptInput
	t.Cleanup(func() {
		initInstallService = previousInstall
		initShouldPrompt = previousPrompt
		initPromptInput = previousInput
	})
	initShouldPrompt = func(io.Writer) bool { return true }
	initPromptInput = strings.NewReader("3\n2\n1\n")
	initInstallService = func(ctx context.Context, request app.InstallRequest) (app.InstallResult, error) {
		if request.ClientID != "codex-app" {
			t.Fatalf("client = %q", request.ClientID)
		}
		if request.ConfigPath != "" {
			t.Fatalf("shared config path should stay implicit, got %q", request.ConfigPath)
		}
		if !request.Write {
			t.Fatal("configure-now path should set write")
		}
		return app.InstallResult{
			Rendered: repository.RenderedClientConfig{
				Client:     repository.SupportedClient{ID: repository.ClientCodexApp, DisplayName: "Codex App"},
				ConfigPath: sharedPath,
				Mode:       repository.RenderModeWrite,
				Content:    "[mcp_servers.optimusctx]\ncommand = \"optimusctx\"\nargs = [\"run\"]\n",
			},
			Guidance: &repository.RenderedGuidance{
				Label: "Codex agent guidance",
				Path:  filepath.Join(homeDir, ".codex", "AGENTS.md"),
			},
			Wrote: true,
		}, nil
	}

	withWorkingDirectory(t, repoRoot, func() {
		writeCLIFile(t, repoRoot+"/main.go", "package main\n")
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init"}, &stdout); err != nil {
			t.Fatalf("Execute(init) error = %v", err)
		}
		output := stdout.String()
		for _, want := range []string{
			"Where should Codex App use OptimusCtx?",
			sharedPath,
			"destination: Your shared Codex config",
			"config path: " + sharedPath,
			"agent guidance: Codex agent guidance",
			"agent guidance path: " + filepath.Join(homeDir, ".codex", "AGENTS.md"),
			"status: wrote config",
			"agent guidance status: wrote guidance",
			"runtime after host pickup: your MCP client should launch `optimusctx run` automatically when it connects",
			"verify host pickup with `optimusctx status`: it should eventually show registration evidence, last MCP initialize/tools discovery, and recent `optimusctx.*` tool calls",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("missing %q in:\n%s", want, output)
			}
		}
	})
}
