package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestEvalRunnerExecutesScenario(t *testing.T) {
	sourceRoot := t.TempDir()
	fixturesRoot := filepath.Join(sourceRoot, "fixtures")
	scenariosDir := filepath.Join(sourceRoot, "scenarios")

	writeEvalFixtureFile(t, filepath.Join(fixturesRoot, "go-basic", "v1", "repository", "main.go"), "package main\n")
	writeEvalScenarioFile(t, filepath.Join(scenariosDir, "basic.json"), repository.EvalScenarioDefinition{
		SchemaVersion: repository.EvalScenarioSchemaV1,
		ID:            "basic",
		Version:       "v1",
		Name:          "Basic",
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
				ID:   "doctor",
				Name: "Doctor",
				Kind: repository.EvalStepKindCommand,
				Expect: repository.EvalExpectedCommand{
					Surface:  repository.EvalCommandSurfaceCLI,
					Command:  repository.EvalCommandDoctor,
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

	var invocations []EvalCommandInvocation
	runner := NewEvalRunner()
	runner.Now = newDeterministicEvalClock()
	runner.RunCommand = func(_ context.Context, invocation EvalCommandInvocation) (EvalCommandExecutionResult, error) {
		invocations = append(invocations, invocation)
		if _, err := os.Stat(filepath.Join(invocation.WorkingDir, "main.go")); err != nil {
			t.Fatalf("materialized fixture missing main.go: %v", err)
		}
		if _, err := os.Stat(filepath.Join(invocation.WorkingDir, ".git")); err != nil {
			t.Fatalf("materialized fixture missing .git directory: %v", err)
		}
		if reflect.DeepEqual(invocation.Args, []string{"pack", "export", "--format", "json", "--output", "artifacts/pack.json"}) {
			writeEvalFixtureFile(t, filepath.Join(invocation.WorkingDir, "artifacts", "pack.json"), "{\"ok\":true}\n")
		}
		return EvalCommandExecutionResult{Stdout: "ok\n", ExitCode: 0}, nil
	}

	result, err := runner.Run(context.Background(), EvalRunRequest{
		ScenarioID:   "basic",
		ScenariosDir: scenariosDir,
		FixturesRoot: fixturesRoot,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !result.Passed {
		t.Fatalf("Run() passed = false, want true: %+v", result)
	}

	gotArgs := make([][]string, 0, len(invocations))
	for _, invocation := range invocations {
		gotArgs = append(gotArgs, invocation.Args)
	}
	wantArgs := [][]string{
		{"init"},
		{"refresh"},
		{"doctor"},
		{"pack", "export", "--format", "json", "--output", "artifacts/pack.json"},
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("invocation args = %v, want %v", gotArgs, wantArgs)
	}
}

func TestEvalRunnerCapturesStepResults(t *testing.T) {
	sourceRoot := t.TempDir()
	fixturesRoot := filepath.Join(sourceRoot, "fixtures")
	scenarioPath := filepath.Join(sourceRoot, "scenarios", "capture.json")

	writeEvalFixtureFile(t, filepath.Join(fixturesRoot, "go-basic", "v1", "repository", "main.go"), "package main\n")
	writeEvalScenarioFile(t, scenarioPath, repository.EvalScenarioDefinition{
		SchemaVersion: repository.EvalScenarioSchemaV1,
		ID:            "capture",
		Version:       "v1",
		Name:          "Capture",
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
				CaptureArtifact: []string{"stdout", "stderr"},
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
					ExitCode: 0,
				},
				CaptureArtifact: []string{"pack"},
			},
		},
		Artifacts: []repository.EvalArtifactRef{
			{ID: "stdout", Kind: repository.EvalArtifactKindStdout, Required: true},
			{ID: "stderr", Kind: repository.EvalArtifactKindStderr, Required: true},
			{ID: "pack", Kind: repository.EvalArtifactKindFile, Path: "artifacts/pack.json", Required: true},
		},
	})

	runner := NewEvalRunner()
	runner.Now = newDeterministicEvalClock()
	runner.RunCommand = func(_ context.Context, invocation EvalCommandInvocation) (EvalCommandExecutionResult, error) {
		if strings.HasPrefix(strings.Join(invocation.Args, " "), "pack export") {
			writeEvalFixtureFile(t, filepath.Join(invocation.WorkingDir, "artifacts", "pack.json"), "{\"status\":\"ok\"}\n")
			return EvalCommandExecutionResult{Stdout: "packed\n", ExitCode: 0}, nil
		}
		return EvalCommandExecutionResult{Stdout: "initialized\n", Stderr: "warning\n", ExitCode: 0}, nil
	}

	result, err := runner.Run(context.Background(), EvalRunRequest{
		ScenarioPath: scenarioPath,
		FixturesRoot: fixturesRoot,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(result.Steps) != 3 {
		t.Fatalf("len(result.Steps) = %d, want 3", len(result.Steps))
	}
	if !result.Passed {
		t.Fatalf("Run() passed = false, want true: %+v", result)
	}

	first := result.Steps[0]
	if first.Stdout != "initialized\n" || first.Stderr != "warning\n" {
		t.Fatalf("first step output = stdout %q stderr %q", first.Stdout, first.Stderr)
	}
	if !first.Passed || first.ExitCode != 0 {
		t.Fatalf("first step result = %+v", first)
	}
	if len(first.Artifacts) != 2 {
		t.Fatalf("len(first.Artifacts) = %d, want 2", len(first.Artifacts))
	}
	if first.Artifacts[0].Bytes == 0 || first.Artifacts[1].Bytes == 0 {
		t.Fatalf("first step artifact bytes = %+v", first.Artifacts)
	}
	if !first.StartedAt.Before(first.FinishedAt) {
		t.Fatalf("first step timing invalid: start=%s finish=%s", first.StartedAt, first.FinishedAt)
	}

	second := result.Steps[2]
	if len(second.Artifacts) != 1 {
		t.Fatalf("len(second.Artifacts) = %d, want 1", len(second.Artifacts))
	}
	if second.Artifacts[0].Location == "" || second.Artifacts[0].Bytes == 0 {
		t.Fatalf("second step artifact = %+v", second.Artifacts[0])
	}
	if len(result.Artifacts) != 3 {
		t.Fatalf("len(result.Artifacts) = %d, want 3", len(result.Artifacts))
	}
}

func TestEvalRunnerAppliesDefaultsToPartialConfiguration(t *testing.T) {
	sourceRoot := t.TempDir()
	fixturesRoot := filepath.Join(sourceRoot, "fixtures")
	scenarioPath := filepath.Join(sourceRoot, "scenarios", "defaults.json")

	writeEvalFixtureFile(t, filepath.Join(fixturesRoot, "go-basic", "v1", "repository", "main.go"), "package main\n")
	writeEvalScenarioFile(t, scenarioPath, repository.EvalScenarioDefinition{
		SchemaVersion: repository.EvalScenarioSchemaV1,
		ID:            "defaults",
		Version:       "v1",
		Name:          "Defaults",
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
		},
	})

	runner := EvalRunner{
		RunCommand: func(_ context.Context, invocation EvalCommandInvocation) (EvalCommandExecutionResult, error) {
			if _, err := os.Stat(filepath.Join(invocation.WorkingDir, ".git")); err != nil {
				t.Fatalf("expected default git initialization: %v", err)
			}
			return EvalCommandExecutionResult{Stdout: "ok\n", ExitCode: 0}, nil
		},
		Now: newDeterministicEvalClock(),
	}

	result, err := runner.Run(context.Background(), EvalRunRequest{
		ScenarioPath: scenarioPath,
		FixturesRoot: fixturesRoot,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.ScenarioID != "defaults" || !result.Passed {
		t.Fatalf("Run() result = %+v", result)
	}
}

func writeEvalScenarioFile(t *testing.T, path string, scenario repository.EvalScenarioDefinition) {
	t.Helper()

	data, err := json.MarshalIndent(scenario, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	writeEvalFixtureFile(t, path, string(data))
}

func writeEvalFixtureFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func newDeterministicEvalClock() func() time.Time {
	base := time.Date(2026, time.March, 15, 12, 0, 0, 0, time.UTC)
	var tick int
	return func() time.Time {
		current := base.Add(time.Duration(tick) * time.Second)
		tick++
		return current
	}
}
