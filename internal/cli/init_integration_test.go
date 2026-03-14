package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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

func TestSnippetCommandPrintsStdoutOnlyWithoutCreatingState(t *testing.T) {
	repoRoot := initCLIRepo(t)
	writeCLIFile(t, filepath.Join(repoRoot, "README.md"), "# Repo\n")

	before := snapshotRelativeFiles(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"snippet"}, &stdout); err != nil {
			t.Fatalf("Execute(snippet) error = %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "# OptimusCtx manual integration snippet") {
			t.Fatalf("output = %q, want snippet header", output)
		}
		if !strings.Contains(output, "MCP serving is not implemented yet in Phase 1.") {
			t.Fatalf("output = %q, want truthful MCP placeholder", output)
		}
	})

	after := snapshotRelativeFiles(t, repoRoot)
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("repository files changed after snippet:\nbefore=%v\nafter=%v", before, after)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, ".optimusctx")); !os.IsNotExist(err) {
		t.Fatalf(".optimusctx state directory should not exist after snippet: %v", err)
	}
}

func TestSnippetCommandLeavesExistingStateUntouched(t *testing.T) {
	repoRoot := initCLIRepo(t)
	writeCLIFile(t, filepath.Join(repoRoot, "main.go"), "package main\n")

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init"}, &stdout); err != nil {
			t.Fatalf("Execute(init) error = %v", err)
		}
	})

	stateDir := filepath.Join(repoRoot, ".optimusctx")
	stateBefore, err := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if err != nil {
		t.Fatalf("ReadFile(state.json) error = %v", err)
	}
	dbInfoBefore, err := os.Stat(filepath.Join(stateDir, "db.sqlite"))
	if err != nil {
		t.Fatalf("Stat(db.sqlite) error = %v", err)
	}

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"snippet"}, &stdout); err != nil {
			t.Fatalf("Execute(snippet) error = %v", err)
		}
		if strings.Count(stdout.String(), "\n") == 0 {
			t.Fatalf("snippet output = %q, want multi-line content", stdout.String())
		}
	})

	stateAfter, err := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if err != nil {
		t.Fatalf("ReadFile(state.json) after snippet error = %v", err)
	}
	dbInfoAfter, err := os.Stat(filepath.Join(stateDir, "db.sqlite"))
	if err != nil {
		t.Fatalf("Stat(db.sqlite) after snippet error = %v", err)
	}

	if !bytes.Equal(stateAfter, stateBefore) {
		t.Fatalf("state.json changed after snippet:\nbefore=%s\nafter=%s", stateBefore, stateAfter)
	}
	if dbInfoAfter.ModTime() != dbInfoBefore.ModTime() || dbInfoAfter.Size() != dbInfoBefore.Size() {
		t.Fatalf("db.sqlite changed after snippet: before=%v/%d after=%v/%d", dbInfoBefore.ModTime(), dbInfoBefore.Size(), dbInfoAfter.ModTime(), dbInfoAfter.Size())
	}
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

func snapshotRelativeFiles(t *testing.T, root string) []string {
	t.Helper()

	var paths []string
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	}); err != nil {
		t.Fatalf("WalkDir(%q) error = %v", root, err)
	}
	return paths
}
