package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

func TestEvalCommand(t *testing.T) {
	t.Run("routes scenario selection through the service", func(t *testing.T) {
		repoRoot := initCLIRepo(t)
		previous := evalCommandService
		t.Cleanup(func() {
			evalCommandService = previous
		})

		evalCommandService = func(ctx context.Context, request app.EvalRunRequest) (repository.EvalRunResult, error) {
			if ctx == nil {
				t.Fatal("context should not be nil")
			}
			if request.ScenarioID != "cli-go-basic-v1" {
				t.Fatalf("ScenarioID = %q, want cli-go-basic-v1", request.ScenarioID)
			}
			if request.ScenariosDir != filepath.Join(repoRoot, "testdata", "eval", "scenarios") {
				t.Fatalf("ScenariosDir = %q", request.ScenariosDir)
			}
			if request.FixturesRoot != filepath.Join(repoRoot, "testdata", "eval", "fixtures") {
				t.Fatalf("FixturesRoot = %q", request.FixturesRoot)
			}
			return repository.EvalRunResult{
				ScenarioID: "cli-go-basic-v1",
				StartedAt:  time.Date(2026, time.March, 15, 12, 0, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, time.March, 15, 12, 0, 2, 0, time.UTC),
				Passed:     true,
				Steps: []repository.EvalStepResult{
					{
						Step:       repository.EvalScenarioStep{ID: "init"},
						StartedAt:  time.Date(2026, time.March, 15, 12, 0, 0, 0, time.UTC),
						FinishedAt: time.Date(2026, time.March, 15, 12, 0, 1, 0, time.UTC),
						ExitCode:   0,
						Passed:     true,
					},
				},
			}, nil
		}

		withWorkingDirectory(t, repoRoot, func() {
			var stdout bytes.Buffer
			if err := NewRootCommand().Execute([]string{"eval", "--scenario", "cli-go-basic-v1"}, &stdout); err != nil {
				t.Fatalf("Execute(eval --scenario) error = %v", err)
			}
			output := stdout.String()
			assertContains(t, output, "scenario id: cli-go-basic-v1")
			assertContains(t, output, "status: passed")
			assertContains(t, output, "step init: passed exit=0")
		})
	})

	t.Run("runs a scenario file through the real command boundary", func(t *testing.T) {
		repoRoot := initCLIRepo(t)
		scenarioPath := filepath.Join(repoRoot, "testdata", "eval", "scenarios", "basic.json")
		writeCLIFile(t, filepath.Join(repoRoot, "testdata", "eval", "fixtures", "go-basic", "v1", "repository", "main.go"), "package main\n")
		writeCLIFile(t, filepath.Join(repoRoot, "testdata", "eval", "fixtures", "go-basic", "v1", "repository", "go.mod"), "module fixture/basic\n\ngo 1.23.0\n")
		writeEvalScenarioJSON(t, scenarioPath, repository.EvalScenarioDefinition{
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
			},
		})

		withWorkingDirectory(t, repoRoot, func() {
			var stdout bytes.Buffer
			if err := NewRootCommand().Execute([]string{"eval", "--scenario-file", scenarioPath}, &stdout); err != nil {
				t.Fatalf("Execute(eval --scenario-file) error = %v", err)
			}
			output := stdout.String()
			assertContains(t, output, "scenario id: basic")
			assertContains(t, output, "status: passed")
			assertContains(t, output, "step init: passed exit=0")
			assertContains(t, output, "step refresh: passed exit=0")
		})
	})
}

func TestEvalCommandRejectsInvalidArgs(t *testing.T) {
	repoRoot := initCLIRepo(t)

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing selector",
			args: []string{"eval"},
			want: "eval requires exactly one of --scenario or --scenario-file",
		},
		{
			name: "both selectors",
			args: []string{"eval", "--scenario", "one", "--scenario-file", "two.json"},
			want: "eval requires exactly one of --scenario or --scenario-file",
		},
		{
			name: "unknown flag",
			args: []string{"eval", "--verbose"},
			want: `unknown eval flag "--verbose"`,
		},
		{
			name: "unexpected argument",
			args: []string{"eval", "basic"},
			want: `eval does not accept arguments; got "basic"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withWorkingDirectory(t, repoRoot, func() {
				var stdout bytes.Buffer
				err := NewRootCommand().Execute(tt.args, &stdout)
				if err == nil || err.Error() != tt.want {
					t.Fatalf("Execute(%v) error = %v, want %q", tt.args, err, tt.want)
				}
				if stdout.Len() != 0 {
					t.Fatalf("stdout = %q, want empty", stdout.String())
				}
			})
		})
	}

	t.Run("surfaces runtime execution failures separately from argument errors", func(t *testing.T) {
		previous := evalCommandService
		t.Cleanup(func() {
			evalCommandService = previous
		})
		evalCommandService = func(context.Context, app.EvalRunRequest) (repository.EvalRunResult, error) {
			return repository.EvalRunResult{
				ScenarioID: "broken",
				StartedAt:  time.Date(2026, time.March, 15, 12, 0, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, time.March, 15, 12, 0, 1, 0, time.UTC),
				Passed:     false,
				Steps: []repository.EvalStepResult{
					{
						Step:       repository.EvalScenarioStep{ID: "init"},
						StartedAt:  time.Date(2026, time.March, 15, 12, 0, 0, 0, time.UTC),
						FinishedAt: time.Date(2026, time.March, 15, 12, 0, 1, 0, time.UTC),
						ExitCode:   1,
					},
				},
			}, errors.New("scenario \"broken\" step \"init\" failed")
		}

		withWorkingDirectory(t, repoRoot, func() {
			var stdout bytes.Buffer
			err := NewRootCommand().Execute([]string{"eval", "--scenario", "broken"}, &stdout)
			if err == nil || err.Error() != `scenario "broken" step "init" failed` {
				t.Fatalf("Execute(eval broken) error = %v", err)
			}
			assertContains(t, stdout.String(), "scenario id: broken")
			assertContains(t, stdout.String(), "status: failed")
		})
	})

	t.Run("returns a specific error for unknown scenario IDs", func(t *testing.T) {
		repoRoot := initCLIRepo(t)
		writeCLIFile(t, filepath.Join(repoRoot, "testdata", "eval", "fixtures", "go-basic", "v1", "repository", "main.go"), "package main\n")
		writeEvalScenarioJSON(t, filepath.Join(repoRoot, "testdata", "eval", "scenarios", "known.json"), repository.EvalScenarioDefinition{
			SchemaVersion: repository.EvalScenarioSchemaV1,
			ID:            "known",
			Version:       "v1",
			Name:          "Known",
			Fixture: repository.EvalFixtureRef{
				ID:          "go-basic",
				Version:     "v1",
				Path:        "go-basic/v1/repository",
				Materialize: repository.EvalFixtureModeCopyTree,
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

		withWorkingDirectory(t, repoRoot, func() {
			var stdout bytes.Buffer
			err := NewRootCommand().Execute([]string{"eval", "--scenario", "missing"}, &stdout)
			if err == nil || err.Error() != `unknown scenario "missing"` {
				t.Fatalf("Execute(eval --scenario missing) error = %v", err)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
		})
	})
}

func writeEvalScenarioJSON(t *testing.T, path string, scenario repository.EvalScenarioDefinition) {
	t.Helper()

	data, err := json.MarshalIndent(scenario, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	writeCLIFile(t, path, string(data))
}
