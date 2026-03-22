package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/store/migrations"
)

func TestInitCommandInitializesFromNestedRepositoryPath(t *testing.T) {
	repoRoot := initCLIRepo(t)
	writeCLIFile(t, filepath.Join(repoRoot, ".gitignore"), "*.tmp\nignored-dir/\n")
	writeCLIFile(t, filepath.Join(repoRoot, "src", "main.go"), "package main\n")
	writeCLIFile(t, filepath.Join(repoRoot, "ignored.tmp"), "ignored\n")
	writeCLIFile(t, filepath.Join(repoRoot, "ignored-dir", "nested.txt"), "ignored dir\n")

	nestedDir := filepath.Join(repoRoot, "src")
	withWorkingDirectory(t, nestedDir, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init"}, &stdout); err != nil {
			t.Fatalf("Execute(init) error = %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "repository root: "+repoRoot) {
			t.Fatalf("output = %q, want repository root", output)
		}
		if !strings.Contains(output, "state directory: "+filepath.Join(repoRoot, ".optimusctx")) {
			t.Fatalf("output = %q, want state directory", output)
		}
		if !strings.Contains(output, "schema version: "+strconv.Itoa(migrations.CurrentVersion())) {
			t.Fatalf("output = %q, want schema version", output)
		}
		if !strings.Contains(output, "refresh generation: 1") {
			t.Fatalf("output = %q, want refresh generation", output)
		}
		if !strings.Contains(output, "freshness: fresh") {
			t.Fatalf("output = %q, want freshness", output)
		}
		if !strings.Contains(output, "discovered files: 3") {
			t.Fatalf("output = %q, want discovered file count", output)
		}
		if !strings.Contains(output, "next step: use `optimusctx init --client <client>` to review the change for claude-desktop, claude-cli, codex-app, codex-cli, gemini-cli, or cursor-cli, or add `--write` to configure one right away") {
			t.Fatalf("output = %q, want onboarding next step", output)
		}
		if !strings.Contains(output, "runtime after registration: your MCP client should launch `optimusctx run` automatically when it connects") {
			t.Fatalf("output = %q, want automatic runtime handoff", output)
		}
	})

	if _, err := os.Stat(filepath.Join(repoRoot, ".optimusctx", "db.sqlite")); err != nil {
		t.Fatalf("expected sqlite state database: %v", err)
	}
}

func TestInitIntegrationReportsUnsupportedWorkingDirectory(t *testing.T) {
	workdir := t.TempDir()

	withWorkingDirectory(t, workdir, func() {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"init"}, &stdout)
		if err == nil {
			t.Fatal("Execute(init) expected error, got nil")
		}
		if !strings.Contains(err.Error(), "no supported repository root found") {
			t.Fatalf("error = %v, want unsupported repository root message", err)
		}
		if stdout.Len() != 0 {
			t.Fatalf("stdout = %q, want empty", stdout.String())
		}
	})
}

func initCLIRepo(t *testing.T) string {
	t.Helper()

	repoRoot := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, output)
	}
	return repoRoot
}

func writeCLIFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func withWorkingDirectory(t *testing.T, dir string, fn func()) {
	t.Helper()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%q) error = %v", dir, err)
	}
	defer func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	}()

	fn()
}

func TestInitCommandHelpIncludesScope(t *testing.T) {
	var stdout bytes.Buffer
	if err := NewRootCommand().Execute([]string{"init", "--help"}, &stdout); err != nil {
		t.Fatalf("Execute(init --help) error = %v", err)
	}

	output := stdout.String()
	want := "optimusctx init [--client <client>] [--config <path>] [--binary <path>] [--scope <local|project|user>] [--write]"
	if !strings.Contains(output, want) {
		t.Fatalf("help output missing %q:\n%s", want, output)
	}
}

func TestInitCommandCodexPreviewDoesNotDumpWholeConfig(t *testing.T) {
	repoRoot := initCLIRepo(t)
	writeCLIFile(t, filepath.Join(repoRoot, "main.go"), "package main\n")

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	configPath := filepath.Join(homeDir, ".codex", "config.toml")
	writeCLIFile(t, configPath, `model = "gpt-5.4"
model_reasoning_effort = "high"

[agents.gsd-debugger]
config_file = "agents/gsd-debugger.toml"

[mcp_servers.linear]
url = "https://mcp.linear.app/mcp"

[projects."/home/nico/projects/genera-platita"]
trust_level = "trusted"
`)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init", "--client", "codex-cli"}, &stdout); err != nil {
			t.Fatalf("Execute(init --client codex-cli) error = %v", err)
		}

		output := stdout.String()
		for _, want := range []string{
			"client: Codex CLI",
			"destination: Your shared Codex config",
			"config path: " + configPath,
			"review this change first:",
			"[mcp_servers.optimusctx]",
			`command = "optimusctx"`,
			`args = ["run"]`,
			"status: ready to configure",
			"next step: rerun `optimusctx init --client codex-cli --write` to apply this setup",
			"runtime after apply: your MCP client should launch `optimusctx run` automatically when it connects",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("missing %q in:\n%s", want, output)
			}
		}
		for _, unwanted := range []string{
			`model = "gpt-5.4"`,
			`[agents.gsd-debugger]`,
			`[mcp_servers.linear]`,
			`[projects."/home/nico/projects/genera-platita"]`,
		} {
			if strings.Contains(output, unwanted) {
				t.Fatalf("output should not dump unrelated existing Codex config %q:\n%s", unwanted, output)
			}
		}
	})
}
