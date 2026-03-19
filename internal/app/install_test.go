package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallServiceSupportsGenericPreview(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{ClientID: "generic"})
	if err != nil {
		t.Fatalf("Register(generic) error = %v", err)
	}
	if result.Wrote {
		t.Fatal("generic preview should not write")
	}
	if result.Rendered.Client.ID != "generic" {
		t.Fatalf("client id = %q", result.Rendered.Client.ID)
	}
	if result.Rendered.ConfigPath != "manual" {
		t.Fatalf("config path = %q", result.Rendered.ConfigPath)
	}
	if !strings.Contains(result.Rendered.Content, "\"command\": \"optimusctx\"") || !strings.Contains(result.Rendered.Content, "\"run\"") {
		t.Fatalf("content missing run command: %s", result.Rendered.Content)
	}
	if len(result.Rendered.Notes) == 0 {
		t.Fatal("generic preview should include notes")
	}

	for _, clientID := range []string{"claude-cli", "codex-app", "codex-cli"} {
		namedResult, err := service.Register(context.Background(), InstallRequest{ClientID: clientID})
		if err != nil {
			t.Fatalf("Register(%s) error = %v", clientID, err)
		}
		if namedResult.Rendered.ConfigPath == "manual" {
			t.Fatalf("%s should not reuse the manual generic fallback", clientID)
		}
	}
}

func TestInstallServiceSupportsClaudeCLIPreview(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{ClientID: "claude-cli"})
	if err != nil {
		t.Fatalf("Register(claude-cli) error = %v", err)
	}
	if result.Wrote {
		t.Fatal("claude-cli preview should not write")
	}
	if result.Rendered.Client.ID != "claude-cli" {
		t.Fatalf("client id = %q", result.Rendered.Client.ID)
	}
	if result.Rendered.ConfigPath != "command" {
		t.Fatalf("config path = %q", result.Rendered.ConfigPath)
	}
	if result.Rendered.Mode != "preview" {
		t.Fatalf("mode = %q", result.Rendered.Mode)
	}
	if result.Rendered.Content != "claude mcp add --transport stdio optimusctx -- optimusctx run" {
		t.Fatalf("content = %q", result.Rendered.Content)
	}
	if len(result.Rendered.Notes) == 0 {
		t.Fatal("claude-cli preview should include notes")
	}
}

func TestInstallServiceSupportsCodexAppPreview(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{ClientID: "codex-app"})
	if err != nil {
		t.Fatalf("Register(codex-app) error = %v", err)
	}
	if result.Wrote {
		t.Fatal("codex-app preview should not write")
	}
	if result.Rendered.Client.ID != "codex-app" {
		t.Fatalf("client id = %q", result.Rendered.Client.ID)
	}
	if result.Rendered.Client.DisplayName != "Codex App" {
		t.Fatalf("display name = %q", result.Rendered.Client.DisplayName)
	}
	if got, want := result.Rendered.ConfigPath, filepath.Join(homeDir, ".codex", "config.toml"); got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}
	if result.Rendered.Mode != "preview" {
		t.Fatalf("mode = %q", result.Rendered.Mode)
	}
	if !strings.Contains(result.Rendered.Content, "[mcp_servers.optimusctx]") {
		t.Fatalf("content missing optimusctx table: %s", result.Rendered.Content)
	}
	if !strings.Contains(result.Rendered.Content, `command = "optimusctx"`) {
		t.Fatalf("content missing command line: %s", result.Rendered.Content)
	}
	if len(result.Rendered.Notes) == 0 {
		t.Fatal("codex-app preview should include notes")
	}
}

func TestInstallServiceSupportsCodexCLIPreview(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{ClientID: "codex-cli"})
	if err != nil {
		t.Fatalf("Register(codex-cli) error = %v", err)
	}
	if result.Wrote {
		t.Fatal("codex-cli preview should not write")
	}
	if result.Rendered.Client.ID != "codex-cli" {
		t.Fatalf("client id = %q", result.Rendered.Client.ID)
	}
	if result.Rendered.Client.DisplayName != "Codex CLI" {
		t.Fatalf("display name = %q", result.Rendered.Client.DisplayName)
	}
	if got, want := result.Rendered.ConfigPath, filepath.Join(homeDir, ".codex", "config.toml"); got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}
	if result.Rendered.Mode != "preview" {
		t.Fatalf("mode = %q", result.Rendered.Mode)
	}
	if !strings.Contains(result.Rendered.Content, "[mcp_servers.optimusctx]") {
		t.Fatalf("content missing optimusctx table: %s", result.Rendered.Content)
	}
	if !strings.Contains(result.Rendered.Content, `command = "optimusctx"`) {
		t.Fatalf("content missing command line: %s", result.Rendered.Content)
	}
	if len(result.Rendered.Notes) == 0 {
		t.Fatal("codex-cli preview should include notes")
	}
}

func TestInstallServicePreservesExistingCodexConfigPreview(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	configPath := filepath.Join(homeDir, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	const existing = `model = "gpt-5"

[mcp_servers.other]
command = "other"
args = ["serve"]

[profiles.default]
approval_policy = "on-request"
`
	if err := os.WriteFile(configPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{ClientID: "codex-app"})
	if err != nil {
		t.Fatalf("Register(codex-app) error = %v", err)
	}
	if !strings.Contains(result.Rendered.Content, `model = "gpt-5"`) {
		t.Fatalf("content missing existing model: %s", result.Rendered.Content)
	}
	if !strings.Contains(result.Rendered.Content, "[mcp_servers.other]") {
		t.Fatalf("content missing existing server: %s", result.Rendered.Content)
	}
	if !strings.Contains(result.Rendered.Content, "[profiles.default]") {
		t.Fatalf("content missing profile table: %s", result.Rendered.Content)
	}
	if !strings.Contains(result.Rendered.Content, "[mcp_servers.optimusctx]") {
		t.Fatalf("content missing optimusctx table: %s", result.Rendered.Content)
	}
	if strings.Count(result.Rendered.Content, "[mcp_servers.optimusctx]") != 1 {
		t.Fatalf("optimusctx table duplicated: %s", result.Rendered.Content)
	}
}

func TestInstallServiceRejectsGenericWrite(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	service := NewInstallService()
	_, err := service.Register(context.Background(), InstallRequest{ClientID: "generic", Write: true})
	if err == nil || !strings.Contains(err.Error(), "does not support --write") {
		t.Fatalf("Register(generic write) error = %v", err)
	}
}
