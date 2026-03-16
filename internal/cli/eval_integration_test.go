package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

func TestEvalCLIScenariosRerun(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	scenarios := []string{"cli-go-basic-v1", "cli-go-worktree-v1"}
	withWorkingDirectory(t, repoRoot, func() {
		for _, scenarioID := range scenarios {
			for run := 0; run < 2; run++ {
				var stdout bytes.Buffer
				if err := NewRootCommand().Execute([]string{"eval", "--scenario", scenarioID}, &stdout); err != nil {
					t.Fatalf("Execute(eval %s) #%d error = %v", scenarioID, run+1, err)
				}
				assertContains(t, stdout.String(), "scenario id: "+scenarioID)
				assertContains(t, stdout.String(), "status: passed")
				assertContains(t, stdout.String(), "step init: passed exit=0")
				assertContains(t, stdout.String(), "step refresh: passed exit=0")
				assertContains(t, stdout.String(), "step doctor: passed exit=0")
				assertContains(t, stdout.String(), "step pack-export: passed exit=0")
			}
		}
	})

	for _, fixtureID := range []string{"go-basic", "go-worktree"} {
		fixtureRoot := filepath.Join(repoRoot, "testdata", "eval", "fixtures", fixtureID, "v1", "repository")
		if _, err := os.Stat(filepath.Join(fixtureRoot, ".optimusctx")); !os.IsNotExist(err) {
			t.Fatalf("fixture source %q should not be mutated by reruns, err=%v", fixtureID, err)
		}
	}
}

func TestEvalArtifactsPersistAcrossRun(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		for _, scenarioID := range []string{"cli-go-basic-v1", "cli-go-worktree-v1"} {
			var stdout bytes.Buffer
			if err := NewRootCommand().Execute([]string{"eval", "--scenario", scenarioID}, &stdout); err != nil {
				t.Fatalf("Execute(eval %s) error = %v", scenarioID, err)
			}
			assertContains(t, stdout.String(), "scenario id: "+scenarioID)
			assertContains(t, stdout.String(), "status: passed")
		}
	})

	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}
	store, err := sqlite.OpenOrCreateStore(context.Background(), layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("OpenOrCreateStore() error = %v", err)
	}
	defer store.Close()

	runOne, stepsOne, artifactsOne, err := store.LoadEvalRun(context.Background(), 1)
	if err != nil {
		t.Fatalf("LoadEvalRun(1) error = %v", err)
	}
	runTwo, stepsTwo, artifactsTwo, err := store.LoadEvalRun(context.Background(), 2)
	if err != nil {
		t.Fatalf("LoadEvalRun(2) error = %v", err)
	}

	if runOne.ArtifactRoot != layout.EvalRunDir(1) || runTwo.ArtifactRoot != layout.EvalRunDir(2) {
		t.Fatalf("artifact roots = %q and %q", runOne.ArtifactRoot, runTwo.ArtifactRoot)
	}
	if !runOne.Passed || !runTwo.Passed {
		t.Fatalf("runs should pass: %+v %+v", runOne, runTwo)
	}
	if len(stepsOne) != 4 || len(stepsTwo) != 4 {
		t.Fatalf("step counts = %d and %d, want 4", len(stepsOne), len(stepsTwo))
	}
	if stepsOne[0].StdoutPath == "" || stepsOne[2].StdoutPath == "" || stepsTwo[0].StdoutPath == "" || stepsTwo[2].StdoutPath == "" {
		t.Fatalf("persisted stdout paths missing: %+v %+v %+v %+v", stepsOne[0], stepsOne[2], stepsTwo[0], stepsTwo[2])
	}
	if len(artifactsOne) != 4 || len(artifactsTwo) != 3 {
		t.Fatalf("artifact counts = %d and %d, want 4 and 3", len(artifactsOne), len(artifactsTwo))
	}

	packOnePath := filepath.Join(layout.EvalRunDir(1), "artifacts", "pack-basic.json")
	packTwoPath := filepath.Join(layout.EvalRunDir(2), "artifacts", "pack-worktree.json")
	if artifactsOne[len(artifactsOne)-1].StoredPath != packOnePath {
		t.Fatalf("artifact one stored path = %q", artifactsOne[len(artifactsOne)-1].StoredPath)
	}
	if artifactsTwo[len(artifactsTwo)-1].StoredPath != packTwoPath {
		t.Fatalf("artifact two stored path = %q", artifactsTwo[len(artifactsTwo)-1].StoredPath)
	}

	infoOne, err := os.Stat(packOnePath)
	if err != nil {
		t.Fatalf("Stat(pack one) error = %v", err)
	}
	infoTwo, err := os.Stat(packTwoPath)
	if err != nil {
		t.Fatalf("Stat(pack two) error = %v", err)
	}
	if infoOne.Size() == 0 || infoTwo.Size() == 0 {
		t.Fatalf("artifact sizes = %d and %d, want non-zero", infoOne.Size(), infoTwo.Size())
	}
	if artifactsOne[len(artifactsOne)-1].SizeBytes == 0 || artifactsTwo[len(artifactsTwo)-1].SizeBytes == 0 {
		t.Fatalf("artifact metadata sizes = %d and %d, want non-zero", artifactsOne[len(artifactsOne)-1].SizeBytes, artifactsTwo[len(artifactsTwo)-1].SizeBytes)
	}
}

func seedCommittedEvalFixtures(t *testing.T, repoRoot string) {
	t.Helper()

	sourceRoot := filepath.Join("..", "..", "testdata", "eval")
	copyCLITree(t, sourceRoot, filepath.Join(repoRoot, "testdata", "eval"))
}

func copyCLITree(t *testing.T, src string, dst string) {
	t.Helper()

	info, err := os.Stat(src)
	if err != nil {
		t.Fatalf("Stat(%s) error = %v", src, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", src)
	}
	if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("ReadDir(%s) error = %v", src, err)
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			copyCLITree(t, srcPath, dstPath)
			continue
		}
		content, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", srcPath, err)
		}
		writeCLIFile(t, dstPath, string(content))
	}
}
