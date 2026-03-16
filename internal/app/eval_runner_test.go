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

func TestEvalRunnerExecutesCLIWorkflow(t *testing.T) {
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
				Setup: []repository.EvalSetupAction{
					{Kind: repository.EvalSetupActionWriteFile, Path: "README.md", Content: "# fixture\n"},
				},
				Assert: []repository.EvalAssertion{
					{Kind: repository.EvalAssertionKindContains, Target: repository.EvalAssertionTargetStdout, Contains: "initialized"},
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
		if reflect.DeepEqual(invocation.Args, []string{"init"}) {
			content, err := os.ReadFile(filepath.Join(invocation.WorkingDir, "README.md"))
			if err != nil {
				t.Fatalf("setup file missing: %v", err)
			}
			if string(content) != "# fixture\n" {
				t.Fatalf("setup file content = %q", string(content))
			}
			return EvalCommandExecutionResult{Stdout: "initialized\n", ExitCode: 0}, nil
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

func TestEvalCLIWorkflowAssertions(t *testing.T) {
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
				Assert: []repository.EvalAssertion{
					{Kind: repository.EvalAssertionKindContains, Target: repository.EvalAssertionTargetStdout, Contains: "initialized"},
					{Kind: repository.EvalAssertionKindContains, Target: repository.EvalAssertionTargetStderr, Contains: "warning"},
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
				Setup: []repository.EvalSetupAction{
					{Kind: repository.EvalSetupActionOverwriteFile, Path: "state.json", Content: "{\"prepared\":true}\n"},
				},
				Assert: []repository.EvalAssertion{
					{Kind: repository.EvalAssertionKindJSONFieldPresent, Target: repository.EvalAssertionTargetArtifact, Artifact: "refresh-json", Path: "summary.freshness"},
					{Kind: repository.EvalAssertionKindJSONFieldEquals, Target: repository.EvalAssertionTargetArtifact, Artifact: "refresh-json", Path: "summary.status", Equals: "ok"},
				},
				CaptureArtifact: []string{"refresh-json"},
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
				Assert: []repository.EvalAssertion{
					{Kind: repository.EvalAssertionKindJSONFieldEquals, Target: repository.EvalAssertionTargetArtifact, Artifact: "pack", Path: "status", Equals: "ok"},
				},
				CaptureArtifact: []string{"pack"},
			},
		},
		Artifacts: []repository.EvalArtifactRef{
			{ID: "stdout", Kind: repository.EvalArtifactKindStdout, Required: true},
			{ID: "stderr", Kind: repository.EvalArtifactKindStderr, Required: true},
			{ID: "refresh-json", Kind: repository.EvalArtifactKindFile, Path: "artifacts/refresh.json", Required: true},
			{ID: "pack", Kind: repository.EvalArtifactKindFile, Path: "artifacts/pack.json", Required: true},
		},
	})

	runner := NewEvalRunner()
	runner.Now = newDeterministicEvalClock()
	runner.RunCommand = func(_ context.Context, invocation EvalCommandInvocation) (EvalCommandExecutionResult, error) {
		if reflect.DeepEqual(invocation.Args, []string{"refresh"}) {
			content, err := os.ReadFile(filepath.Join(invocation.WorkingDir, "state.json"))
			if err != nil {
				t.Fatalf("refresh setup file missing: %v", err)
			}
			if string(content) != "{\"prepared\":true}\n" {
				t.Fatalf("refresh setup file content = %q", string(content))
			}
			writeEvalFixtureFile(t, filepath.Join(invocation.WorkingDir, "artifacts", "refresh.json"), "{\"summary\":{\"freshness\":\"fresh\",\"status\":\"ok\"}}\n")
			return EvalCommandExecutionResult{Stdout: "refreshed\n", ExitCode: 0}, nil
		}
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
	if len(result.Artifacts) != 4 {
		t.Fatalf("len(result.Artifacts) = %d, want 4", len(result.Artifacts))
	}
}

func TestEvalRunnerPersistsCLIArtifacts(t *testing.T) {
	sourceRoot := t.TempDir()
	fixturesRoot := filepath.Join(sourceRoot, "fixtures")
	scenarioPath := filepath.Join(sourceRoot, "scenarios", "persist.json")

	writeEvalFixtureFile(t, filepath.Join(fixturesRoot, "go-basic", "v1", "repository", "main.go"), "package main\n")
	writeEvalScenarioFile(t, scenarioPath, repository.EvalScenarioDefinition{
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
				CaptureArtifact: []string{"init-stdout"},
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
			{ID: "init-stdout", Kind: repository.EvalArtifactKindStdout, Required: true},
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
		return EvalCommandExecutionResult{Stdout: "initialized\n", ExitCode: 0}, nil
	}

	result, err := runner.Run(context.Background(), EvalRunRequest{
		ScenarioPath: scenarioPath,
		FixturesRoot: fixturesRoot,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	artifactRoot := filepath.Join(sourceRoot, "persisted")
	stepRecords, artifactRecords, err := persistEvalEvidence(artifactRoot, result)
	if err != nil {
		t.Fatalf("persistEvalEvidence() error = %v", err)
	}
	if len(stepRecords) != 3 || len(artifactRecords) != 2 {
		t.Fatalf("persisted record counts = %d steps, %d artifacts", len(stepRecords), len(artifactRecords))
	}
	if stepRecords[0].StdoutPath == "" {
		t.Fatalf("expected init stdout path, got %+v", stepRecords[0])
	}
	if _, err := os.Stat(stepRecords[0].StdoutPath); err != nil {
		t.Fatalf("Stat(init stdout) error = %v", err)
	}
	if artifactRecords[1].StoredPath != filepath.Join(artifactRoot, "artifacts", "pack.json") {
		t.Fatalf("pack stored path = %q", artifactRecords[1].StoredPath)
	}
	content, err := os.ReadFile(artifactRecords[1].StoredPath)
	if err != nil {
		t.Fatalf("ReadFile(pack stored path) error = %v", err)
	}
	if string(content) != "{\"status\":\"ok\"}\n" {
		t.Fatalf("pack stored content = %q", string(content))
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

func TestEvalRunnerReportsExecutionFailuresWithStepContext(t *testing.T) {
	sourceRoot := t.TempDir()
	fixturesRoot := filepath.Join(sourceRoot, "fixtures")
	scenarioPath := filepath.Join(sourceRoot, "scenarios", "failure.json")

	writeEvalFixtureFile(t, filepath.Join(fixturesRoot, "go-basic", "v1", "repository", "main.go"), "package main\n")
	writeEvalScenarioFile(t, scenarioPath, repository.EvalScenarioDefinition{
		SchemaVersion: repository.EvalScenarioSchemaV1,
		ID:            "failure",
		Version:       "v1",
		Name:          "Failure",
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

	runner := NewEvalRunner()
	runner.Now = newDeterministicEvalClock()
	runner.RunCommand = func(_ context.Context, invocation EvalCommandInvocation) (EvalCommandExecutionResult, error) {
		if !reflect.DeepEqual(invocation.Args, []string{"init"}) {
			t.Fatalf("invocation.Args = %v, want init", invocation.Args)
		}
		return EvalCommandExecutionResult{ExitCode: 1, Stderr: "boom\n"}, nil
	}

	result, err := runner.Run(context.Background(), EvalRunRequest{
		ScenarioPath: scenarioPath,
		FixturesRoot: fixturesRoot,
	})
	if err == nil || !strings.Contains(err.Error(), `scenario "failure" step "init" failed: exit code 1, want 0`) {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Passed {
		t.Fatalf("Run() passed = true, want false: %+v", result)
	}
	if len(result.Steps) != 1 || result.Steps[0].ExitCode != 1 {
		t.Fatalf("Run() steps = %+v", result.Steps)
	}
}

func TestEvalScenarioMaterializesFixture(t *testing.T) {
	sourceRoot := t.TempDir()
	fixturesRoot := filepath.Join(sourceRoot, "fixtures")
	scenarioPath := filepath.Join(sourceRoot, "scenarios", "materialize.json")

	writeEvalFixtureFile(t, filepath.Join(fixturesRoot, "go-basic", "v1", "repository", "main.go"), "package main\n")
	writeEvalScenarioFile(t, scenarioPath, repository.EvalScenarioDefinition{
		SchemaVersion: repository.EvalScenarioSchemaV1,
		ID:            "materialize",
		Version:       "v1",
		Name:          "Materialize",
		Fixture: repository.EvalFixtureRef{
			ID:           "go-basic",
			Version:      "v1",
			Path:         "go-basic/v1/repository",
			Materialize:  repository.EvalFixtureModeCopyTree,
			WorkspaceDir: "workspace",
		},
		Steps: []repository.EvalScenarioStep{{
			ID:   "init",
			Name: "Initialize",
			Kind: repository.EvalStepKindCommand,
			Expect: repository.EvalExpectedCommand{
				Surface:  repository.EvalCommandSurfaceCLI,
				Command:  repository.EvalCommandInit,
				ExitCode: 0,
			},
		}},
	})

	workspaces := []string{
		filepath.Join(sourceRoot, "runs", "first"),
		filepath.Join(sourceRoot, "runs", "second"),
	}
	var materialized []string
	runner := NewEvalRunner()
	runner.Now = newDeterministicEvalClock()
	runner.MkdirTemp = func(string, string) (string, error) {
		path := workspaces[len(materialized)]
		materialized = append(materialized, path)
		return path, nil
	}
	runner.RunCommand = func(_ context.Context, invocation EvalCommandInvocation) (EvalCommandExecutionResult, error) {
		if _, err := os.Stat(filepath.Join(invocation.WorkingDir, "main.go")); err != nil {
			t.Fatalf("materialized fixture missing main.go: %v", err)
		}
		if _, err := os.Stat(filepath.Join(invocation.WorkingDir, ".git")); err != nil {
			t.Fatalf("materialized fixture missing .git: %v", err)
		}
		return EvalCommandExecutionResult{Stdout: "ok\n", ExitCode: 0}, nil
	}

	for run := 0; run < len(workspaces); run++ {
		result, err := runner.Run(context.Background(), EvalRunRequest{
			ScenarioPath: scenarioPath,
			FixturesRoot: fixturesRoot,
		})
		if err != nil {
			t.Fatalf("Run() #%d error = %v", run+1, err)
		}
		if !result.Passed {
			t.Fatalf("Run() #%d passed = false, want true", run+1)
		}
	}

	for _, root := range workspaces {
		if _, err := os.Stat(filepath.Join(root, "workspace", "main.go")); err != nil {
			t.Fatalf("workspace %q missing main.go: %v", root, err)
		}
	}
	if _, err := os.Stat(filepath.Join(fixturesRoot, "go-basic", "v1", "repository", ".git")); !os.IsNotExist(err) {
		t.Fatalf("fixture source should not gain .git metadata, err=%v", err)
	}
}

func TestEvalScenarioRerun(t *testing.T) {
	sourceRoot := t.TempDir()
	fixturesRoot := filepath.Join(sourceRoot, "fixtures")
	scenariosDir := filepath.Join(sourceRoot, "scenarios")

	writeEvalFixtureFile(t, filepath.Join(fixturesRoot, "go-basic", "v1", "repository", "main.go"), "package main\n")
	writeEvalScenarioFile(t, filepath.Join(scenariosDir, "rerun.json"), repository.EvalScenarioDefinition{
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

	workspaces := []string{
		filepath.Join(sourceRoot, "runs", "first"),
		filepath.Join(sourceRoot, "runs", "second"),
	}
	runIndex := 0
	runner := NewEvalRunner()
	runner.Now = newDeterministicEvalClock()
	runner.MkdirTemp = func(string, string) (string, error) {
		path := workspaces[runIndex]
		runIndex++
		return path, nil
	}
	runner.RunCommand = func(_ context.Context, invocation EvalCommandInvocation) (EvalCommandExecutionResult, error) {
		sentinel := filepath.Join(invocation.WorkingDir, "manual-only.txt")
		if len(invocation.Args) > 0 && invocation.Args[0] == "init" {
			if _, err := os.Stat(sentinel); err == nil {
				t.Fatalf("rerun reused prior workspace state at %q", sentinel)
			}
			writeEvalFixtureFile(t, sentinel, "transient\n")
		}
		return EvalCommandExecutionResult{Stdout: "ok\n", ExitCode: 0}, nil
	}

	for run := 0; run < len(workspaces); run++ {
		result, err := runner.Run(context.Background(), EvalRunRequest{
			ScenarioID:   "rerun",
			ScenariosDir: scenariosDir,
			FixturesRoot: fixturesRoot,
		})
		if err != nil {
			t.Fatalf("Run() #%d error = %v", run+1, err)
		}
		if !result.Passed {
			t.Fatalf("Run() #%d passed = false, want true", run+1)
		}
	}

	if _, err := os.Stat(filepath.Join(fixturesRoot, "go-basic", "v1", "repository", "manual-only.txt")); !os.IsNotExist(err) {
		t.Fatalf("fixture source should remain immutable across reruns, err=%v", err)
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
