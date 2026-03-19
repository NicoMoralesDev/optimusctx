package repository

import (
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
)

func TestRenderCodexConfig(t *testing.T) {
	rendered, err := RenderCodexConfig("", NewServeCommand(""))
	if err != nil {
		t.Fatalf("RenderCodexConfig() error = %v", err)
	}

	if !strings.Contains(rendered, "[mcp_servers.optimusctx]") {
		t.Fatalf("RenderCodexConfig() missing server table: %s", rendered)
	}

	const wantCommand = `command = "optimusctx"`
	if !strings.Contains(rendered, wantCommand) {
		t.Fatalf("RenderCodexConfig() missing command line %q in %s", wantCommand, rendered)
	}

	document := decodeCodexConfigForTest(t, rendered)
	server := decodeCodexServerForTest(t, document, "optimusctx")
	if got := stringValueForTest(t, server, "command"); got != "optimusctx" {
		t.Fatalf("command = %q", got)
	}
	if got := stringSliceValueForTest(t, server, "args"); len(got) != 1 || got[0] != "run" {
		t.Fatalf("args = %v", got)
	}
}

func TestMergeCodexConfigPreservesExistingContent(t *testing.T) {
	const existing = `
model = "gpt-5"

[mcp_servers.other]
command = "other"
args = ["serve"]

[profiles.default]
approval_policy = "on-request"
`

	rendered, err := MergeCodexConfig([]byte(existing), "", NewServeCommand(""))
	if err != nil {
		t.Fatalf("MergeCodexConfig() error = %v", err)
	}

	document := decodeCodexConfigForTest(t, rendered)
	if got := stringValueForTest(t, document, "model"); got != "gpt-5" {
		t.Fatalf("model = %q", got)
	}

	other := decodeCodexServerForTest(t, document, "other")
	if got := stringValueForTest(t, other, "command"); got != "other" {
		t.Fatalf("other command = %q", got)
	}

	profiles := mapValueForTest(t, document, "profiles")
	defaultProfile := mapValueForTest(t, profiles, "default")
	if got := stringValueForTest(t, defaultProfile, "approval_policy"); got != "on-request" {
		t.Fatalf("approval_policy = %q", got)
	}

	optimus := decodeCodexServerForTest(t, document, "optimusctx")
	if got := stringValueForTest(t, optimus, "command"); got != "optimusctx" {
		t.Fatalf("optimusctx command = %q", got)
	}
}

func TestMergeCodexConfigIsIdempotent(t *testing.T) {
	first, err := MergeCodexConfig(nil, "", NewServeCommand(""))
	if err != nil {
		t.Fatalf("first MergeCodexConfig() error = %v", err)
	}

	second, err := MergeCodexConfig([]byte(first), "", NewServeCommand(""))
	if err != nil {
		t.Fatalf("second MergeCodexConfig() error = %v", err)
	}

	if strings.Count(second, "[mcp_servers.optimusctx]") != 1 {
		t.Fatalf("optimusctx server table duplicated in %s", second)
	}

	firstDocument := decodeCodexConfigForTest(t, first)
	secondDocument := decodeCodexConfigForTest(t, second)
	if got, want := firstDocument, secondDocument; !mapsEqualForTest(got, want) {
		t.Fatalf("second merge changed config:\nfirst: %#v\nsecond: %#v", got, want)
	}
}

func decodeCodexConfigForTest(t *testing.T, content string) map[string]any {
	t.Helper()

	var document map[string]any
	if err := toml.Unmarshal([]byte(content), &document); err != nil {
		t.Fatalf("toml.Unmarshal() error = %v", err)
	}
	return document
}

func decodeCodexServerForTest(t *testing.T, document map[string]any, serverName string) map[string]any {
	t.Helper()

	servers := mapValueForTest(t, document, codexMCPServersKey)
	return mapValueForTest(t, servers, serverName)
}

func mapValueForTest(t *testing.T, source map[string]any, key string) map[string]any {
	t.Helper()

	raw, ok := source[key]
	if !ok {
		t.Fatalf("missing key %q in %#v", key, source)
	}

	value, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("key %q has type %T", key, raw)
	}
	return value
}

func stringValueForTest(t *testing.T, source map[string]any, key string) string {
	t.Helper()

	raw, ok := source[key]
	if !ok {
		t.Fatalf("missing key %q in %#v", key, source)
	}

	value, ok := raw.(string)
	if !ok {
		t.Fatalf("key %q has type %T", key, raw)
	}
	return value
}

func stringSliceValueForTest(t *testing.T, source map[string]any, key string) []string {
	t.Helper()

	raw, ok := source[key]
	if !ok {
		t.Fatalf("missing key %q in %#v", key, source)
	}

	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("key %q has type %T", key, raw)
	}

	values := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			t.Fatalf("array item has type %T", item)
		}
		values = append(values, value)
	}
	return values
}

func mapsEqualForTest(left, right map[string]any) bool {
	leftEncoded, leftErr := toml.Marshal(left)
	rightEncoded, rightErr := toml.Marshal(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return string(leftEncoded) == string(rightEncoded)
}
