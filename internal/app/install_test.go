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
}

func TestInstallServiceRejectsGenericWrite(t *testing.T) {
	service := NewInstallService()
	_, err := service.Register(context.Background(), InstallRequest{ClientID: "generic", Write: true})
	if err == nil || !strings.Contains(err.Error(), "does not support --write") {
		t.Fatalf("Register(generic write) error = %v", err)
	}
}
