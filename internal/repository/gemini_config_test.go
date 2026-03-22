package repository

import (
	"strings"
	"testing"
)

func TestRenderGeminiConfig(t *testing.T) {
	got, err := RenderGeminiConfig("", NewServeCommand(""))
	if err != nil {
		t.Fatalf("RenderGeminiConfig() error = %v", err)
	}

	const want = "{\n  \"mcpServers\": {\n    \"optimusctx\": {\n      \"args\": [\n        \"run\"\n      ],\n      \"command\": \"optimusctx\"\n    }\n  }\n}\n"
	if got != want {
		t.Fatalf("RenderGeminiConfig() = %q, want %q", got, want)
	}
}

func TestMergeGeminiConfigPreservesUnrelatedSettings(t *testing.T) {
	const existing = "{\n  \"theme\": \"dark\",\n  \"mcpServers\": {\n    \"other\": {\n      \"command\": \"other\",\n      \"args\": [\"serve\"]\n    }\n  }\n}\n"

	got, err := MergeGeminiConfig([]byte(existing), "", NewServeCommand(""))
	if err != nil {
		t.Fatalf("MergeGeminiConfig() error = %v", err)
	}
	for _, want := range []string{
		"\"theme\": \"dark\"",
		"\"other\": {",
		"\"optimusctx\": {",
		"\"command\": \"optimusctx\"",
		"\"run\"",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("merged config missing %q:\n%s", want, got)
		}
	}
}

func TestGeminiConfigServerNames(t *testing.T) {
	const existing = "{\n  \"mcpServers\": {\n    \"optimusctx\": {\n      \"command\": \"optimusctx\",\n      \"args\": [\"run\"]\n    },\n    \"other\": {\n      \"command\": \"other\",\n      \"args\": [\"serve\"]\n    }\n  }\n}\n"

	got, err := GeminiConfigServerNames([]byte(existing))
	if err != nil {
		t.Fatalf("GeminiConfigServerNames() error = %v", err)
	}
	if !got["optimusctx"] || !got["other"] {
		t.Fatalf("GeminiConfigServerNames() = %+v, want optimusctx and other", got)
	}
}
