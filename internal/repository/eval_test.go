package repository

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestEvalScenarioContracts(t *testing.T) {
	t.Parallel()

	scenario := EvalScenarioDefinition{
		SchemaVersion: EvalScenarioSchemaV1,
		ID:            "cli-go-basic-v1",
		Version:       "v1",
		Name:          "CLI flow on basic Go fixture",
		Description:   "Defines the canonical schema for Phase 9 scenario loading.",
		Fixture: EvalFixtureRef{
			ID:           "go-basic",
			Version:      "v1",
			Path:         "go-basic/v1/repository",
			Materialize:  EvalFixtureModeCopyTree,
			WorkspaceDir: "workspace",
		},
		Steps: []EvalScenarioStep{
			{
				ID:   "init",
				Name: "Initialize repository state",
				Kind: EvalStepKindCommand,
				Expect: EvalExpectedCommand{
					Surface:  EvalCommandSurfaceCLI,
					Command:  EvalCommandInit,
					ExitCode: 0,
				},
				Setup: []EvalSetupAction{
					{Kind: EvalSetupActionWriteFile, Path: "README.md", Content: "# fixture\n"},
				},
				Assert: []EvalAssertion{
					{Kind: EvalAssertionKindContains, Target: EvalAssertionTargetStdout, Contains: "initialized"},
				},
				CaptureArtifact: []string{"init-stdout"},
			},
			{
				ID:   "refresh",
				Name: "Refresh repository inventory",
				Kind: EvalStepKindCommand,
				Expect: EvalExpectedCommand{
					Surface:  EvalCommandSurfaceCLI,
					Command:  EvalCommandRefresh,
					ExitCode: 0,
				},
			},
			{
				ID:   "doctor",
				Name: "Check repository health",
				Kind: EvalStepKindCommand,
				Expect: EvalExpectedCommand{
					Surface:  EvalCommandSurfaceCLI,
					Command:  EvalCommandDoctor,
					ExitCode: 0,
				},
				Assert: []EvalAssertion{
					{Kind: EvalAssertionKindJSONFieldPresent, Target: EvalAssertionTargetArtifact, Artifact: "doctor-json", Path: "repository.root_path"},
				},
				CaptureArtifact: []string{"doctor-stdout", "doctor-json"},
			},
			{
				ID:   "pack-export",
				Name: "Export packed context",
				Kind: EvalStepKindCommand,
				Expect: EvalExpectedCommand{
					Surface:  EvalCommandSurfaceCLI,
					Command:  EvalCommandPackExport,
					Args:     []string{"--format", "json"},
					ExitCode: 0,
				},
				CaptureArtifact: []string{"pack-file"},
			},
		},
		Artifacts: []EvalArtifactRef{
			{ID: "init-stdout", Kind: EvalArtifactKindStdout, Required: true},
			{ID: "doctor-stdout", Kind: EvalArtifactKindStdout, Required: true},
			{ID: "doctor-json", Kind: EvalArtifactKindFile, Path: "artifacts/doctor.json", Required: false},
			{ID: "pack-file", Kind: EvalArtifactKindFile, Path: "artifacts/pack.json", Required: true},
		},
	}

	if err := scenario.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	encoded, err := json.MarshalIndent(scenario, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}

	var decoded EvalScenarioDefinition
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !reflect.DeepEqual(decoded, scenario) {
		t.Fatalf("decoded scenario mismatch\n got: %#v\nwant: %#v", decoded, scenario)
	}
}

func TestEvalAssertions(t *testing.T) {
	t.Parallel()

	scenario := EvalScenarioDefinition{
		SchemaVersion: EvalScenarioSchemaV1,
		ID:            "assertions-v1",
		Version:       "v1",
		Name:          "Assertions",
		Fixture: EvalFixtureRef{
			ID:          "go-basic",
			Version:     "v1",
			Path:        "go-basic/v1/repository",
			Materialize: EvalFixtureModeCopyTree,
		},
		Steps: []EvalScenarioStep{
			{
				ID:   "init",
				Name: "Initialize",
				Kind: EvalStepKindCommand,
				Expect: EvalExpectedCommand{
					Surface:  EvalCommandSurfaceCLI,
					Command:  EvalCommandInit,
					ExitCode: 0,
				},
				Assert: []EvalAssertion{
					{Kind: EvalAssertionKindContains, Target: EvalAssertionTargetStdout, Contains: "initialized"},
				},
			},
			{
				ID:   "refresh",
				Name: "Refresh",
				Kind: EvalStepKindCommand,
				Expect: EvalExpectedCommand{
					Surface:  EvalCommandSurfaceCLI,
					Command:  EvalCommandRefresh,
					ExitCode: 0,
				},
				Assert: []EvalAssertion{
					{Kind: EvalAssertionKindJSONFieldPresent, Target: EvalAssertionTargetArtifact, Artifact: "refresh-json", Path: "summary.freshness"},
					{Kind: EvalAssertionKindJSONFieldEquals, Target: EvalAssertionTargetArtifact, Artifact: "refresh-json", Path: "summary.status", Equals: "ok"},
				},
				CaptureArtifact: []string{"refresh-json"},
			},
		},
		Artifacts: []EvalArtifactRef{
			{ID: "refresh-json", Kind: EvalArtifactKindFile, Path: "artifacts/refresh.json", Required: true},
		},
	}

	if err := scenario.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	encoded, err := json.Marshal(scenario)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded EvalScenarioDefinition
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !reflect.DeepEqual(decoded, scenario) {
		t.Fatalf("decoded scenario mismatch\n got: %#v\nwant: %#v", decoded, scenario)
	}
}

func TestEvalScenarioValidation(t *testing.T) {
	t.Parallel()

	t.Run("rejects refresh before init", func(t *testing.T) {
		scenario := EvalScenarioDefinition{
			SchemaVersion: EvalScenarioSchemaV1,
			ID:            "bad-order-v1",
			Version:       "v1",
			Name:          "Bad order",
			Fixture: EvalFixtureRef{
				ID:          "go-basic",
				Version:     "v1",
				Path:        "go-basic/v1/repository",
				Materialize: EvalFixtureModeCopyTree,
			},
			Steps: []EvalScenarioStep{
				{
					ID:   "refresh",
					Name: "Refresh first",
					Kind: EvalStepKindCommand,
					Expect: EvalExpectedCommand{
						Surface:  EvalCommandSurfaceCLI,
						Command:  EvalCommandRefresh,
						ExitCode: 0,
					},
				},
			},
		}

		err := scenario.Validate()
		if err == nil || !strings.Contains(err.Error(), "refresh requires a prior init step") {
			t.Fatalf("Validate() error = %v, want refresh ordering error", err)
		}
	})

	t.Run("rejects unknown artifact reference", func(t *testing.T) {
		scenario := EvalScenarioDefinition{
			SchemaVersion: EvalScenarioSchemaV1,
			ID:            "bad-artifact-v1",
			Version:       "v1",
			Name:          "Bad artifact",
			Fixture: EvalFixtureRef{
				ID:          "go-basic",
				Version:     "v1",
				Path:        filepath.ToSlash(filepath.Join("go-basic", "v1", "repository")),
				Materialize: EvalFixtureModeCopyTree,
			},
			Steps: []EvalScenarioStep{
				{
					ID:   "init",
					Name: "Initialize",
					Kind: EvalStepKindCommand,
					Expect: EvalExpectedCommand{
						Surface:  EvalCommandSurfaceCLI,
						Command:  EvalCommandInit,
						ExitCode: 0,
					},
					CaptureArtifact: []string{"missing"},
				},
			},
		}

		err := scenario.Validate()
		if err == nil || !strings.Contains(err.Error(), `unknown artifact "missing"`) {
			t.Fatalf("Validate() error = %v, want missing artifact error", err)
		}
	})

	t.Run("rejects pack export before refresh", func(t *testing.T) {
		scenario := EvalScenarioDefinition{
			SchemaVersion: EvalScenarioSchemaV1,
			ID:            "bad-pack-order-v1",
			Version:       "v1",
			Name:          "Bad pack order",
			Fixture: EvalFixtureRef{
				ID:          "go-basic",
				Version:     "v1",
				Path:        "go-basic/v1/repository",
				Materialize: EvalFixtureModeCopyTree,
			},
			Steps: []EvalScenarioStep{
				{
					ID:   "init",
					Name: "Initialize",
					Kind: EvalStepKindCommand,
					Expect: EvalExpectedCommand{
						Surface:  EvalCommandSurfaceCLI,
						Command:  EvalCommandInit,
						ExitCode: 0,
					},
				},
				{
					ID:   "pack-export",
					Name: "Export too early",
					Kind: EvalStepKindCommand,
					Expect: EvalExpectedCommand{
						Surface:  EvalCommandSurfaceCLI,
						Command:  EvalCommandPackExport,
						ExitCode: 0,
					},
				},
			},
		}

		err := scenario.Validate()
		if err == nil || !strings.Contains(err.Error(), "pack_export requires a prior refresh step") {
			t.Fatalf("Validate() error = %v, want pack_export ordering error", err)
		}
	})

	t.Run("rejects setup paths outside workspace", func(t *testing.T) {
		scenario := EvalScenarioDefinition{
			SchemaVersion: EvalScenarioSchemaV1,
			ID:            "bad-setup-v1",
			Version:       "v1",
			Name:          "Bad setup",
			Fixture: EvalFixtureRef{
				ID:          "go-basic",
				Version:     "v1",
				Path:        "go-basic/v1/repository",
				Materialize: EvalFixtureModeCopyTree,
			},
			Steps: []EvalScenarioStep{
				{
					ID:   "init",
					Name: "Initialize",
					Kind: EvalStepKindCommand,
					Expect: EvalExpectedCommand{
						Surface:  EvalCommandSurfaceCLI,
						Command:  EvalCommandInit,
						ExitCode: 0,
					},
					Setup: []EvalSetupAction{
						{Kind: EvalSetupActionDeleteFile, Path: "../escape.txt"},
					},
				},
			},
		}

		err := scenario.Validate()
		if err == nil || !strings.Contains(err.Error(), "must stay within the workspace") {
			t.Fatalf("Validate() error = %v, want workspace boundary error", err)
		}
	})

	t.Run("rejects artifact assertions without known artifact refs", func(t *testing.T) {
		scenario := EvalScenarioDefinition{
			SchemaVersion: EvalScenarioSchemaV1,
			ID:            "bad-assert-v1",
			Version:       "v1",
			Name:          "Bad assertion",
			Fixture: EvalFixtureRef{
				ID:          "go-basic",
				Version:     "v1",
				Path:        "go-basic/v1/repository",
				Materialize: EvalFixtureModeCopyTree,
			},
			Steps: []EvalScenarioStep{
				{
					ID:   "init",
					Name: "Initialize",
					Kind: EvalStepKindCommand,
					Expect: EvalExpectedCommand{
						Surface:  EvalCommandSurfaceCLI,
						Command:  EvalCommandInit,
						ExitCode: 0,
					},
					Assert: []EvalAssertion{
						{Kind: EvalAssertionKindJSONFieldPresent, Target: EvalAssertionTargetArtifact, Artifact: "missing", Path: "summary.status"},
					},
				},
			},
		}

		err := scenario.Validate()
		if err == nil || !strings.Contains(err.Error(), `unknown artifact "missing"`) {
			t.Fatalf("Validate() error = %v, want missing artifact assertion error", err)
		}
	})
}

func TestEvalFixtureReferences(t *testing.T) {
	t.Parallel()

	scenariosDir := filepath.Join("..", "..", "testdata", "eval", "scenarios")
	fixturesRoot := filepath.Join("..", "..", "testdata", "eval", "fixtures")

	scenarios, err := LoadEvalScenarios(scenariosDir)
	if err != nil {
		t.Fatalf("LoadEvalScenarios() error = %v", err)
	}
	if len(scenarios) == 0 {
		t.Fatal("LoadEvalScenarios() returned no scenarios")
	}

	gotIDs := []string{scenarios[0].ID, scenarios[1].ID}
	wantIDs := []string{"cli-go-basic-v1", "cli-go-worktree-v1"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("loaded scenario IDs = %v, want %v", gotIDs, wantIDs)
	}

	if err := ValidateEvalFixtureReferences(scenarios, fixturesRoot); err != nil {
		t.Fatalf("ValidateEvalFixtureReferences() error = %v", err)
	}

	expectedFixtures := map[string]EvalFixtureRef{
		"cli-go-basic-v1": {
			ID:           "go-basic",
			Version:      "v1",
			Path:         "go-basic/v1/repository",
			Materialize:  EvalFixtureModeCopyTree,
			WorkspaceDir: "workspace",
		},
		"cli-go-worktree-v1": {
			ID:           "go-worktree",
			Version:      "v1",
			Path:         "go-worktree/v1/repository",
			Materialize:  EvalFixtureModeCopyTree,
			WorkspaceDir: "workspace",
		},
	}
	expectedCommands := map[string][]EvalCommandName{
		"cli-go-basic-v1":    {EvalCommandInit, EvalCommandRefresh, EvalCommandDoctor, EvalCommandPackExport},
		"cli-go-worktree-v1": {EvalCommandInit, EvalCommandRefresh, EvalCommandDoctor, EvalCommandPackExport},
	}

	for _, scenario := range scenarios {
		if got := scenario.Fixture; !reflect.DeepEqual(got, expectedFixtures[scenario.ID]) {
			t.Fatalf("scenario %q fixture = %#v, want %#v", scenario.ID, got, expectedFixtures[scenario.ID])
		}

		var commands []EvalCommandName
		for _, step := range scenario.Steps {
			commands = append(commands, step.Expect.Command)
		}
		if !reflect.DeepEqual(commands, expectedCommands[scenario.ID]) {
			t.Fatalf("scenario %q commands = %v, want %v", scenario.ID, commands, expectedCommands[scenario.ID])
		}
		for _, step := range scenario.Steps {
			for _, action := range step.Setup {
				if err := validateEvalRelativePath(action.Path); err != nil {
					t.Fatalf("scenario %q setup path %q invalid: %v", scenario.ID, action.Path, err)
				}
			}
		}
	}
}
