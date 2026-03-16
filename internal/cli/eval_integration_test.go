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

func TestEvalScenarioRerun(t *testing.T) {
	repoRoot := initCLIRepo(t)
	fixtureRoot := filepath.Join(repoRoot, "testdata", "eval", "fixtures", "go-basic", "v1", "repository")
	writeCLIFile(t, filepath.Join(fixtureRoot, "main.go"), "package main\n")
	writeCLIFile(t, filepath.Join(fixtureRoot, "go.mod"), "module fixture/basic\n\ngo 1.23.0\n")
	writeEvalScenarioJSON(t, filepath.Join(repoRoot, "testdata", "eval", "scenarios", "rerun.json"), repository.EvalScenarioDefinition{
		SchemaVersion: repository.EvalScenarioSchemaV1,
		ID:            "rerun",
		Version:       "v1",
		Name:          "Rerun",
		Fixture: repository.EvalFixtureRef{
			ID:           "go-basic",
			Version:      "v1",
			Path:         "go-basic/v1/repository",
			Materialize:  repository.EvalFixtureModeCopyTree,
			WorkspaceDir: "workspace",
		},
		Steps: []repository.EvalScenarioStep{
			{
				ID:   "init",
				Name: "Initialize",
				Kind: repository.EvalStepKindCommand,
				Expect: repository.EvalExpectedCommand{
					Surface:  repository.EvalCommandSurfaceCLI,
					Command:  repository.EvalCommandInit,
					ExitCode: 0,
				},
			},
			{
				ID:   "refresh",
				Name: "Refresh",
				Kind: repository.EvalStepKindCommand,
				Expect: repository.EvalExpectedCommand{
					Surface:  repository.EvalCommandSurfaceCLI,
					Command:  repository.EvalCommandRefresh,
					ExitCode: 0,
				},
			},
		},
	})

	withWorkingDirectory(t, repoRoot, func() {
		for run := 0; run < 2; run++ {
			var stdout bytes.Buffer
			if err := NewRootCommand().Execute([]string{"eval", "--scenario", "rerun"}, &stdout); err != nil {
				t.Fatalf("Execute(eval rerun) #%d error = %v", run+1, err)
			}
			assertContains(t, stdout.String(), "scenario id: rerun")
			assertContains(t, stdout.String(), "status: passed")
			assertContains(t, stdout.String(), "step init: passed exit=0")
			assertContains(t, stdout.String(), "step refresh: passed exit=0")
		}
	})

	if _, err := os.Stat(filepath.Join(fixtureRoot, ".optimusctx")); !os.IsNotExist(err) {
		t.Fatalf("fixture source should not be mutated by reruns, err=%v", err)
	}
}

func TestEvalArtifactsPersistAcrossRun(t *testing.T) {
	repoRoot := initCLIRepo(t)
	fixtureRoot := filepath.Join(repoRoot, "testdata", "eval", "fixtures", "go-basic", "v1", "repository")
	writeCLIFile(t, filepath.Join(fixtureRoot, "main.go"), "package main\n")
	writeCLIFile(t, filepath.Join(fixtureRoot, "go.mod"), "module fixture/basic\n\ngo 1.23.0\n")
	writeEvalScenarioJSON(t, filepath.Join(repoRoot, "testdata", "eval", "scenarios", "persist.json"), repository.EvalScenarioDefinition{
		SchemaVersion: repository.EvalScenarioSchemaV1,
		ID:            "persist",
		Version:       "v1",
		Name:          "Persist",
		Fixture: repository.EvalFixtureRef{
			ID:           "go-basic",
			Version:      "v1",
			Path:         "go-basic/v1/repository",
			Materialize:  repository.EvalFixtureModeCopyTree,
			WorkspaceDir: "workspace",
		},
		Steps: []repository.EvalScenarioStep{
			{
				ID:   "init",
				Name: "Initialize",
				Kind: repository.EvalStepKindCommand,
				Expect: repository.EvalExpectedCommand{
					Surface:  repository.EvalCommandSurfaceCLI,
					Command:  repository.EvalCommandInit,
					ExitCode: 0,
				},
			},
			{
				ID:   "refresh",
				Name: "Refresh",
				Kind: repository.EvalStepKindCommand,
				Expect: repository.EvalExpectedCommand{
					Surface:  repository.EvalCommandSurfaceCLI,
					Command:  repository.EvalCommandRefresh,
					ExitCode: 0,
				},
			},
			{
				ID:   "pack",
				Name: "Pack",
				Kind: repository.EvalStepKindCommand,
				Expect: repository.EvalExpectedCommand{
					Surface:  repository.EvalCommandSurfaceCLI,
					Command:  repository.EvalCommandPackExport,
					Args:     []string{"--format", "json"},
					ExitCode: 0,
				},
				CaptureArtifact: []string{"pack"},
			},
		},
		Artifacts: []repository.EvalArtifactRef{
			{ID: "pack", Kind: repository.EvalArtifactKindFile, Path: "artifacts/pack.json", Required: true},
		},
	})

	withWorkingDirectory(t, repoRoot, func() {
		for run := 0; run < 2; run++ {
			var stdout bytes.Buffer
			if err := NewRootCommand().Execute([]string{"eval", "--scenario", "persist"}, &stdout); err != nil {
				t.Fatalf("Execute(eval persist) #%d error = %v", run+1, err)
			}
			assertContains(t, stdout.String(), "scenario id: persist")
			assertContains(t, stdout.String(), "status: passed")
			assertContains(t, stdout.String(), "step pack: passed exit=0")
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
	if len(stepsOne) != 3 || len(stepsTwo) != 3 {
		t.Fatalf("step counts = %d and %d, want 3", len(stepsOne), len(stepsTwo))
	}
	if stepsOne[0].StdoutPath == "" || stepsTwo[0].StdoutPath == "" {
		t.Fatalf("persisted stdout paths missing: %+v %+v", stepsOne[0], stepsTwo[0])
	}
	if len(artifactsOne) != 1 || len(artifactsTwo) != 1 {
		t.Fatalf("artifact counts = %d and %d, want 1", len(artifactsOne), len(artifactsTwo))
	}
	if artifactsOne[0].StoredPath != filepath.Join(layout.EvalRunDir(1), "artifacts", "pack.json") {
		t.Fatalf("artifact one stored path = %q", artifactsOne[0].StoredPath)
	}
	if artifactsTwo[0].StoredPath != filepath.Join(layout.EvalRunDir(2), "artifacts", "pack.json") {
		t.Fatalf("artifact two stored path = %q", artifactsTwo[0].StoredPath)
	}

	infoOne, err := os.Stat(artifactsOne[0].StoredPath)
	if err != nil {
		t.Fatalf("Stat(artifact one) error = %v", err)
	}
	infoTwo, err := os.Stat(artifactsTwo[0].StoredPath)
	if err != nil {
		t.Fatalf("Stat(artifact two) error = %v", err)
	}
	if infoOne.Size() == 0 || infoTwo.Size() == 0 {
		t.Fatalf("artifact sizes = %d and %d, want non-zero", infoOne.Size(), infoTwo.Size())
	}
	if artifactsOne[0].SizeBytes != artifactsTwo[0].SizeBytes {
		t.Fatalf("artifact metadata sizes = %d and %d", artifactsOne[0].SizeBytes, artifactsTwo[0].SizeBytes)
	}
}
