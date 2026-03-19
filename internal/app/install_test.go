package app

import (
	"context"
	"strings"
	"testing"
)

func TestInstallServiceSupportsGenericPreview(t *testing.T) {
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

func TestInstallServiceRejectsGenericWrite(t *testing.T) {
	service := NewInstallService()
	_, err := service.Register(context.Background(), InstallRequest{ClientID: "generic", Write: true})
	if err == nil || !strings.Contains(err.Error(), "does not support --write") {
		t.Fatalf("Register(generic write) error = %v", err)
	}
}
