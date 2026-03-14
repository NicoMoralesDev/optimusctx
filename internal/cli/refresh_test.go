package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

func TestRefreshCommand(t *testing.T) {
	repoRoot := initCLIRepo(t)

	previous := refreshCommandService
	t.Cleanup(func() {
		refreshCommandService = previous
	})

	refreshCommandService = func(ctx context.Context, workingDir string) (app.RefreshResult, error) {
		if workingDir != repoRoot {
			t.Fatalf("workingDir = %q, want %q", workingDir, repoRoot)
		}
		return app.RefreshResult{
			RepositoryRoot:      repoRoot,
			Generation:          7,
			FreshnessStatus:     repository.FreshnessStatusPartiallyDegraded,
			AddedFiles:          2,
			ChangedContentFiles: 3,
			DeletedFiles:        1,
			MovedFiles:          4,
			NewlyIgnoredFiles:   5,
			ReincludedFiles:     1,
			UnchangedFiles:      6,
		}, nil
	}

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"refresh"}, &stdout); err != nil {
			t.Fatalf("Execute(refresh) error = %v", err)
		}

		output := stdout.String()
		assertContains(t, output, "repository root: "+repoRoot)
		assertContains(t, output, "refresh generation: 7")
		assertContains(t, output, "freshness: partially degraded")
		assertContains(t, output, "added files: 3")
		assertContains(t, output, "changed files: 3")
		assertContains(t, output, "deleted files: 1")
		assertContains(t, output, "moved files: 4")
		assertContains(t, output, "newly ignored files: 5")
		assertContains(t, output, "unchanged files: 6")
	})
}

func TestRefreshCommandErrors(t *testing.T) {
	repoRoot := initCLIRepo(t)

	t.Run("rejects arguments", func(t *testing.T) {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"refresh", "unexpected"}, &stdout)
		if err == nil || err.Error() != "refresh does not accept arguments" {
			t.Fatalf("Execute(refresh unexpected) error = %v", err)
		}
		if stdout.Len() != 0 {
			t.Fatalf("stdout = %q, want empty", stdout.String())
		}
	})

	t.Run("reports unsupported working directory", func(t *testing.T) {
		workdir := t.TempDir()
		previous := refreshCommandService
		t.Cleanup(func() {
			refreshCommandService = previous
		})
		refreshCommandService = func(ctx context.Context, workingDir string) (app.RefreshResult, error) {
			return app.RefreshResult{}, errors.New("resolve repository root: " + repository.ErrRepositoryNotFound.Error())
		}

		withWorkingDirectory(t, workdir, func() {
			var stdout bytes.Buffer
			err := NewRootCommand().Execute([]string{"refresh"}, &stdout)
			if err == nil {
				t.Fatal("Execute(refresh) expected error, got nil")
			}
			if !strings.Contains(err.Error(), "no supported repository root found from "+workdir) {
				t.Fatalf("error = %v, want unsupported repository root message", err)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
		})
	})

	t.Run("returns service errors", func(t *testing.T) {
		previous := refreshCommandService
		t.Cleanup(func() {
			refreshCommandService = previous
		})
		refreshCommandService = func(ctx context.Context, workingDir string) (app.RefreshResult, error) {
			return app.RefreshResult{}, errors.New("boom")
		}

		withWorkingDirectory(t, repoRoot, func() {
			var stdout bytes.Buffer
			err := NewRootCommand().Execute([]string{"refresh"}, &stdout)
			if err == nil || err.Error() != "boom" {
				t.Fatalf("Execute(refresh) error = %v, want boom", err)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
		})
	})
}

func TestRenderFreshnessStatus(t *testing.T) {
	if got := renderFreshnessStatus(repository.FreshnessStatusPartiallyDegraded); got != "partially degraded" {
		t.Fatalf("renderFreshnessStatus(partially_degraded) = %q", got)
	}
	if got := renderFreshnessStatus(repository.FreshnessStatusFresh); got != "fresh" {
		t.Fatalf("renderFreshnessStatus(fresh) = %q", got)
	}
}

func assertContains(t *testing.T, output string, fragment string) {
	t.Helper()

	if !strings.Contains(output, fragment) {
		t.Fatalf("output = %q, want fragment %q", output, fragment)
	}
}
