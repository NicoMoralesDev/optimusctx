package repository

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryLocatorPrefersGitRootFromNestedDirectory(t *testing.T) {
	repoRoot := t.TempDir()
	runGitCommand(t, repoRoot, "init")
	runGitCommand(t, repoRoot, "checkout", "-b", "main")
	writeFile(t, filepath.Join(repoRoot, "README.md"), "optimusctx\n")
	runGitCommand(t, repoRoot, "add", "README.md")
	runGitCommand(t, repoRoot, "-c", "user.name=Test User", "-c", "user.email=test@example.com", "commit", "-m", "initial commit")

	nested := filepath.Join(repoRoot, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	root, err := NewLocator().Resolve(nested)
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}

	if root.RootPath != repoRoot {
		t.Fatalf("root path = %q, want %q", root.RootPath, repoRoot)
	}
	if root.DetectionMode != DetectionModeGit {
		t.Fatalf("detection mode = %q, want %q", root.DetectionMode, DetectionModeGit)
	}
	if root.Fingerprint.GitCommonDir == "" {
		t.Fatal("git common dir should not be empty")
	}
	if root.Fingerprint.GitHeadRef != "main" {
		t.Fatalf("head ref = %q, want main", root.Fingerprint.GitHeadRef)
	}
	if len(root.Fingerprint.GitHeadCommit) != 40 {
		t.Fatalf("head commit length = %d, want 40", len(root.Fingerprint.GitHeadCommit))
	}
}

func TestRepositoryLocatorFallsBackToOptimusCtxState(t *testing.T) {
	rootDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(rootDir, ".optimusctx"), 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}

	nested := filepath.Join(rootDir, "nested", "dir")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	root, err := ResolveRepositoryRoot(nested)
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}

	if root.RootPath != rootDir {
		t.Fatalf("root path = %q, want %q", root.RootPath, rootDir)
	}
	if root.DetectionMode != DetectionModeOptimusCtxState {
		t.Fatalf("detection mode = %q, want %q", root.DetectionMode, DetectionModeOptimusCtxState)
	}
	if root.Fingerprint.GitCommonDir != "" || root.Fingerprint.GitHeadRef != "" || root.Fingerprint.GitHeadCommit != "" {
		t.Fatal("fallback fingerprint should not include git metadata")
	}
}

func TestRepositoryLocatorNormalizesSymlinkedStartPath(t *testing.T) {
	repoRoot := t.TempDir()
	runGitCommand(t, repoRoot, "init")

	realNested := filepath.Join(repoRoot, "real", "nested")
	if err := os.MkdirAll(realNested, 0o755); err != nil {
		t.Fatalf("mkdir real nested: %v", err)
	}

	symlinkPath := filepath.Join(t.TempDir(), "repo-link")
	if err := os.Symlink(realNested, symlinkPath); err != nil {
		t.Fatalf("symlink nested path: %v", err)
	}

	root, err := ResolveRepositoryRoot(symlinkPath)
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}

	if root.RootPath != repoRoot {
		t.Fatalf("root path = %q, want %q", root.RootPath, repoRoot)
	}
}

func TestResolveRepositoryRootReturnsNotFoundWhenNoRepositoryExists(t *testing.T) {
	root, err := ResolveRepositoryRoot(t.TempDir())
	if !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrRepositoryNotFound)
	}
	if root != (RepositoryRoot{}) {
		t.Fatalf("unexpected root result: %+v", root)
	}
}

func runGitCommand(t *testing.T, dir string, args ...string) string {
	t.Helper()

	output, err := runGit(dir, args...)
	if err != nil {
		t.Fatalf("git %s failed: %v", strings.Join(args, " "), err)
	}
	return output
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
