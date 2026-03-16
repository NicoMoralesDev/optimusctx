package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
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
