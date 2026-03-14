package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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
		if !strings.Contains(output, "schema version: 1") {
			t.Fatalf("output = %q, want schema version", output)
		}
		if !strings.Contains(output, "discovered files: 3") {
			t.Fatalf("output = %q, want discovered file count", output)
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
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	fn()
}
