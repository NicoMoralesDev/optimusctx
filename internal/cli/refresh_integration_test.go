package cli

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

func TestRefreshIntegration(t *testing.T) {
	fixture := cliRefreshFixture{repoRoot: initCLIRepo(t)}
	writeCLIFile(t, filepath.Join(fixture.repoRoot, "main.go"), "package main\n")
	writeCLIFile(t, filepath.Join(fixture.repoRoot, "README.md"), "# Repo\n")

	fixture.runInit(t)
	output := fixture.runRefresh(t)

	assertContains(t, output, "repository root: "+fixture.repoRoot)
	assertContains(t, output, "refresh generation: 2")
	assertContains(t, output, "freshness: fresh")
	assertContains(t, output, "added files: 0")
	assertContains(t, output, "changed files: 0")
	assertContains(t, output, "deleted files: 0")
	assertContains(t, output, "moved files: 0")
	assertContains(t, output, "newly ignored files: 0")
	assertContains(t, output, "unchanged files: 2")
}

func TestTrackedMutationRefreshCounts(t *testing.T) {
	fixture := newCLIRefreshFixture(t)
	fixture.runInit(t)
	fixture.applyTrackedMutations(t)

	output := fixture.runRefresh(t)
	assertContains(t, output, "refresh generation: 2")
	assertContains(t, output, "freshness: fresh")
	assertContains(t, output, "added files: 1")
	assertContains(t, output, "changed files: 2")
	assertContains(t, output, "deleted files: 1")
	assertContains(t, output, "moved files: 1")
	assertContains(t, output, "newly ignored files: 1")
	assertContains(t, output, "unchanged files: 1")

	noOpOutput := fixture.runRefresh(t)
	assertContains(t, noOpOutput, "refresh generation: 3")
	assertContains(t, noOpOutput, "freshness: fresh")
	assertContains(t, noOpOutput, "added files: 0")
	assertContains(t, noOpOutput, "changed files: 0")
	assertContains(t, noOpOutput, "deleted files: 0")
	assertContains(t, noOpOutput, "moved files: 0")
	assertContains(t, noOpOutput, "newly ignored files: 0")
	assertContains(t, noOpOutput, "unchanged files: 5")
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
