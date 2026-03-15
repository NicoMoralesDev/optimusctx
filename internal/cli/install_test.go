package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestSnippetCommandPrintsUpdatedMCPContract(t *testing.T) {
	repoRoot := initCLIRepo(t)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"snippet"}, &stdout); err != nil {
			t.Fatalf("Execute(snippet) error = %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "optimusctx mcp serve") {
			t.Fatalf("output = %q, want serve contract", output)
		}
		if !strings.Contains(output, "optimusctx install --client claude-desktop") {
			t.Fatalf("output = %q, want install guidance", output)
		}

		document := extractConfigDocument(t, output)
		command, ok := document.MCPServers[repository.DefaultMCPServerName]
		if !ok {
			t.Fatalf("snippet config missing %q entry: %+v", repository.DefaultMCPServerName, document.MCPServers)
		}
		if command.Command != "/absolute/path/to/optimusctx" {
			t.Fatalf("command = %q, want placeholder binary path", command.Command)
		}
		if strings.Join(command.Args, " ") != "mcp serve" {
			t.Fatalf("args = %v, want [mcp serve]", command.Args)
		}
	})
}

func TestInstallRegistrationDryRun(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "Claude", "claude_desktop_config.json")
	var stdout bytes.Buffer

	if err := NewRootCommand().Execute([]string{
		"install",
		"--client", "claude-desktop",
		"--config", configPath,
		"--binary", "/tmp/optimusctx",
	}, &stdout); err != nil {
		t.Fatalf("Execute(install dry-run) error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "mode: preview") {
		t.Fatalf("output = %q, want preview mode", output)
	}
	if !strings.Contains(output, "status: preview only") {
		t.Fatalf("output = %q, want preview status", output)
	}
	if !strings.Contains(output, "config path: "+configPath) {
		t.Fatalf("output = %q, want config path", output)
	}
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("config file should not exist after dry-run: %v", err)
	}

	document := extractConfigDocument(t, output)
	command := document.MCPServers[repository.DefaultMCPServerName]
	if command.Command != "/tmp/optimusctx" {
		t.Fatalf("command = %q, want /tmp/optimusctx", command.Command)
	}
	if strings.Join(command.Args, " ") != "mcp serve" {
		t.Fatalf("args = %v, want [mcp serve]", command.Args)
	}

	snippetDocument := extractConfigDocument(t, appSnippetOutput(t))
	if got, want := document.MCPServers[repository.DefaultMCPServerName].Args, snippetDocument.MCPServers[repository.DefaultMCPServerName].Args; strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("install args = %v, snippet args = %v", got, want)
	}
}

func TestInstallRegistrationConsent(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "Claude", "claude_desktop_config.json")
	existing := []byte("{\n  \"mcpServers\": {\n    \"existing\": {\n      \"command\": \"/bin/existing\",\n      \"args\": [\"serve\"]\n    }\n  }\n}\n")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(configPath, existing, 0o644); err != nil {
		t.Fatalf("WriteFile(existing config) error = %v", err)
	}

	var stdout bytes.Buffer
	if err := NewRootCommand().Execute([]string{
		"install",
		"--client", "claude-desktop",
		"--config", configPath,
		"--binary", "/tmp/optimusctx",
		"--write",
	}, &stdout); err != nil {
		t.Fatalf("Execute(install --write) error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "mode: write") {
		t.Fatalf("output = %q, want write mode", output)
	}
	if !strings.Contains(output, "status: wrote config") {
		t.Fatalf("output = %q, want wrote status", output)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(config) error = %v", err)
	}
	document := extractConfigDocument(t, string(content))
	if _, ok := document.MCPServers["existing"]; !ok {
		t.Fatalf("written config missing existing server: %+v", document.MCPServers)
	}
	command, ok := document.MCPServers[repository.DefaultMCPServerName]
	if !ok {
		t.Fatalf("written config missing optimusctx server: %+v", document.MCPServers)
	}
	if command.Command != "/tmp/optimusctx" {
		t.Fatalf("command = %q, want /tmp/optimusctx", command.Command)
	}
}

func TestInstallCommandRejectsUnsupportedClient(t *testing.T) {
	var stdout bytes.Buffer

	err := NewRootCommand().Execute([]string{"install", "--client", "unknown", "--config", filepath.Join(t.TempDir(), "client.json"), "--binary", "/tmp/optimusctx"}, &stdout)
	if err == nil || err.Error() != "unsupported client \"unknown\"" {
		t.Fatalf("Execute(install unknown client) error = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func appSnippetOutput(t *testing.T) string {
	t.Helper()

	var stdout bytes.Buffer
	if err := NewRootCommand().Execute([]string{"snippet"}, &stdout); err != nil {
		t.Fatalf("Execute(snippet) error = %v", err)
	}
	return stdout.String()
}

func extractConfigDocument(t *testing.T, output string) repository.ClientConfigDocument {
	t.Helper()

	start := strings.Index(output, "{")
	end := strings.LastIndex(output, "}")
	if start == -1 || end == -1 || end < start {
		t.Fatalf("output missing JSON object: %q", output)
	}

	var document repository.ClientConfigDocument
	if err := json.Unmarshal([]byte(output[start:end+1]), &document); err != nil {
		t.Fatalf("Unmarshal(config JSON) error = %v\noutput=%q", err, output[start:end+1])
	}
	return document
}
