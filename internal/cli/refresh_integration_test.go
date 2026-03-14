package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

func TestRefreshIntegration(t *testing.T) {
	repoRoot := initCLIRepo(t)
	writeCLIFile(t, filepath.Join(repoRoot, "main.go"), "package main\n")
	writeCLIFile(t, filepath.Join(repoRoot, "README.md"), "# Repo\n")

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init"}, &stdout); err != nil {
			t.Fatalf("Execute(init) error = %v", err)
		}
	})

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"refresh"}, &stdout); err != nil {
			t.Fatalf("Execute(refresh) error = %v", err)
		}

		output := stdout.String()
		assertContains(t, output, "repository root: "+repoRoot)
		assertContains(t, output, "refresh generation: 2")
		assertContains(t, output, "freshness: fresh")
		assertContains(t, output, "added files: 0")
		assertContains(t, output, "changed files: 0")
		assertContains(t, output, "deleted files: 0")
		assertContains(t, output, "moved files: 0")
		assertContains(t, output, "newly ignored files: 0")
		assertContains(t, output, "unchanged files: 2")
	})
}

func TestFreshnessStateCLI(t *testing.T) {
	repoRoot := initCLIRepo(t)
	writeCLIFile(t, filepath.Join(repoRoot, ".gitignore"), "")
	writeCLIFile(t, filepath.Join(repoRoot, "README.md"), "# Repo\n")
	writeCLIFile(t, filepath.Join(repoRoot, "main.go"), "package main\n")
	writeCLIFile(t, filepath.Join(repoRoot, "move-me.txt"), "move me\n")
	writeCLIFile(t, filepath.Join(repoRoot, "delete-me.txt"), "delete me\n")
	writeCLIFile(t, filepath.Join(repoRoot, "ignored-later.log"), "baseline log\n")

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init"}, &stdout); err != nil {
			t.Fatalf("Execute(init) error = %v", err)
		}
	})

	writeCLIFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc refreshed() {}\n")
	writeCLIFile(t, filepath.Join(repoRoot, "added.go"), "package main\n")
	writeCLIFile(t, filepath.Join(repoRoot, ".gitignore"), "*.log\n")
	if err := os.Remove(filepath.Join(repoRoot, "delete-me.txt")); err != nil {
		t.Fatalf("Remove(delete-me.txt) error = %v", err)
	}
	if err := os.Rename(filepath.Join(repoRoot, "move-me.txt"), filepath.Join(repoRoot, "moved.txt")); err != nil {
		t.Fatalf("Rename(move-me.txt) error = %v", err)
	}

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"refresh"}, &stdout); err != nil {
			t.Fatalf("Execute(refresh) error = %v", err)
		}

		output := stdout.String()
		assertContains(t, output, "refresh generation: 2")
		assertContains(t, output, "freshness: fresh")
		assertContains(t, output, "added files: 1")
		assertContains(t, output, "changed files: 2")
		assertContains(t, output, "deleted files: 1")
		assertContains(t, output, "moved files: 1")
		assertContains(t, output, "newly ignored files: 1")
		assertContains(t, output, "unchanged files: 1")
	})
}

func TestDegradedRefreshRecovery(t *testing.T) {
	repoRoot := initCLIRepo(t)
	writeCLIFile(t, filepath.Join(repoRoot, "main.go"), "package main\n")

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init"}, &stdout); err != nil {
			t.Fatalf("Execute(init) error = %v", err)
		}
	})

	writeCLIFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc degraded() {}\n")

	previous := refreshCommandService
	t.Cleanup(func() {
		refreshCommandService = previous
	})
	refreshCommandService = func(ctx context.Context, workingDir string) (app.RefreshResult, error) {
		return app.NewRefreshService().Refresh(ctx, app.RefreshRequest{
			StartPath: workingDir,
			Reason:    repository.RefreshReasonManual,
			InjectFailure: func(stage string) error {
				if stage == "after_files" {
					return errors.New("forced after file updates")
				}
				return nil
			},
		})
	}

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"refresh"}, &stdout)
		if err == nil || !strings.Contains(err.Error(), "forced after file updates") {
			t.Fatalf("Execute(refresh) error = %v, want injected failure", err)
		}

		output := stdout.String()
		assertContains(t, output, "refresh generation: 2")
		assertContains(t, output, "freshness: partially degraded")
		assertContains(t, output, "changed files: 1")
	})

	refreshCommandService = previous
	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"refresh"}, &stdout); err != nil {
			t.Fatalf("Execute(refresh recovery) error = %v", err)
		}

		output := stdout.String()
		assertContains(t, output, "refresh generation: 3")
		assertContains(t, output, "freshness: fresh")
		assertContains(t, output, "changed files: 1")
	})
}
