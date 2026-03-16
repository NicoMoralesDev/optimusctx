package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
)

const EvalScenarioSchemaV1 = "optimusctx/eval-scenario@v1"

type EvalStepKind string

const (
	EvalStepKindCommand EvalStepKind = "command"
)

type EvalCommandSurface string

const (
	EvalCommandSurfaceCLI EvalCommandSurface = "cli"
)

type EvalCommandName string

const (
	EvalCommandInit       EvalCommandName = "init"
	EvalCommandRefresh    EvalCommandName = "refresh"
	EvalCommandDoctor     EvalCommandName = "doctor"
	EvalCommandPackExport EvalCommandName = "pack_export"
)

type EvalFixtureMode string

const (
	EvalFixtureModeCopyTree EvalFixtureMode = "copy_tree"
)

type EvalArtifactKind string

const (
	EvalArtifactKindStdout EvalArtifactKind = "stdout"
	EvalArtifactKindStderr EvalArtifactKind = "stderr"
	EvalArtifactKindFile   EvalArtifactKind = "file"
)

type EvalFixtureRef struct {
	ID           string          `json:"id"`
	Version      string          `json:"version"`
	Path         string          `json:"path"`
	Materialize  EvalFixtureMode `json:"materialize"`
	WorkspaceDir string          `json:"workspaceDir,omitempty"`
}

type EvalExpectedCommand struct {
	Surface  EvalCommandSurface `json:"surface"`
	Command  EvalCommandName    `json:"command"`
	Args     []string           `json:"args,omitempty"`
	ExitCode int                `json:"exitCode"`
}

type EvalArtifactRef struct {
	ID       string           `json:"id"`
	Kind     EvalArtifactKind `json:"kind"`
	Path     string           `json:"path,omitempty"`
	Required bool             `json:"required"`
}

type EvalScenarioStep struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Kind            EvalStepKind        `json:"kind"`
	Expect          EvalExpectedCommand `json:"expect"`
	CaptureArtifact []string            `json:"captureArtifact,omitempty"`
}

type EvalScenarioDefinition struct {
	SchemaVersion string             `json:"schemaVersion"`
	ID            string             `json:"id"`
	Version       string             `json:"version"`
	Name          string             `json:"name"`
	Description   string             `json:"description,omitempty"`
	Fixture       EvalFixtureRef     `json:"fixture"`
	Steps         []EvalScenarioStep `json:"steps"`
	Artifacts     []EvalArtifactRef  `json:"artifacts,omitempty"`
}

type EvalArtifactResult struct {
	Ref      EvalArtifactRef `json:"ref"`
	Location string          `json:"location,omitempty"`
	Present  bool            `json:"present"`
	Bytes    int64           `json:"bytes,omitempty"`
}

type EvalStepResult struct {
	Step       EvalScenarioStep     `json:"step"`
	StartedAt  time.Time            `json:"startedAt"`
	FinishedAt time.Time            `json:"finishedAt"`
	ExitCode   int                  `json:"exitCode"`
	Passed     bool                 `json:"passed"`
	Stdout     string               `json:"stdout,omitempty"`
	Stderr     string               `json:"stderr,omitempty"`
	Artifacts  []EvalArtifactResult `json:"artifacts,omitempty"`
}

type EvalRunResult struct {
	SchemaVersion         string                 `json:"schemaVersion"`
	ScenarioID            string                 `json:"scenarioId"`
	Scenario              EvalScenarioDefinition `json:"scenario"`
	WorkspacePath         string                 `json:"workspacePath,omitempty"`
	StartedAt             time.Time              `json:"startedAt"`
	FinishedAt            time.Time              `json:"finishedAt"`
	Passed                bool                   `json:"passed"`
	PersistedRunID        int64                  `json:"persistedRunId,omitempty"`
	PersistedArtifactRoot string                 `json:"persistedArtifactRoot,omitempty"`
	Steps                 []EvalStepResult       `json:"steps"`
	Artifacts             []EvalArtifactResult   `json:"artifacts,omitempty"`
}

func (s EvalScenarioDefinition) Validate() error {
	if s.SchemaVersion != EvalScenarioSchemaV1 {
		return fmt.Errorf("schemaVersion must be %q", EvalScenarioSchemaV1)
	}
	if strings.TrimSpace(s.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(s.Version) == "" {
		return errors.New("version is required")
	}
	if strings.TrimSpace(s.Name) == "" {
		return errors.New("name is required")
	}
	if err := s.Fixture.validate(); err != nil {
		return fmt.Errorf("fixture: %w", err)
	}
	if len(s.Steps) == 0 {
		return errors.New("at least one step is required")
	}
	if err := validateEvalArtifacts(s.Artifacts); err != nil {
		return err
	}
	if err := validateEvalSteps(s.Steps, s.Artifacts); err != nil {
		return err
	}
	return nil
}

func (f EvalFixtureRef) validate() error {
	if strings.TrimSpace(f.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(f.Version) == "" {
		return errors.New("version is required")
	}
	if strings.TrimSpace(f.Path) == "" {
		return errors.New("path is required")
	}
	if f.Materialize != EvalFixtureModeCopyTree {
		return fmt.Errorf("materialize must be %q", EvalFixtureModeCopyTree)
	}
	return nil
}

func validateEvalArtifacts(artifacts []EvalArtifactRef) error {
	seen := make(map[string]struct{}, len(artifacts))
	for _, artifact := range artifacts {
		if strings.TrimSpace(artifact.ID) == "" {
			return errors.New("artifacts[].id is required")
		}
		if _, ok := seen[artifact.ID]; ok {
			return fmt.Errorf("artifacts[%q]: duplicate artifact id", artifact.ID)
		}
		seen[artifact.ID] = struct{}{}
		switch artifact.Kind {
		case EvalArtifactKindStdout, EvalArtifactKindStderr:
			if artifact.Path != "" {
				return fmt.Errorf("artifacts[%q]: path must be empty for %s", artifact.ID, artifact.Kind)
			}
		case EvalArtifactKindFile:
			if strings.TrimSpace(artifact.Path) == "" {
				return fmt.Errorf("artifacts[%q]: path is required for file artifacts", artifact.ID)
			}
		default:
			return fmt.Errorf("artifacts[%q]: unsupported kind %q", artifact.ID, artifact.Kind)
		}
	}
	return nil
}

func validateEvalSteps(steps []EvalScenarioStep, artifacts []EvalArtifactRef) error {
	artifactIDs := make(map[string]struct{}, len(artifacts))
	for _, artifact := range artifacts {
		artifactIDs[artifact.ID] = struct{}{}
	}

	seen := make(map[string]struct{}, len(steps))
	seenInit := false
	seenRefresh := false

	for idx, step := range steps {
		if strings.TrimSpace(step.ID) == "" {
			return fmt.Errorf("steps[%d]: id is required", idx)
		}
		if _, ok := seen[step.ID]; ok {
			return fmt.Errorf("steps[%d]: duplicate step id %q", idx, step.ID)
		}
		seen[step.ID] = struct{}{}
		if strings.TrimSpace(step.Name) == "" {
			return fmt.Errorf("steps[%d]: name is required", idx)
		}
		if step.Kind != EvalStepKindCommand {
			return fmt.Errorf("steps[%d]: kind must be %q", idx, EvalStepKindCommand)
		}
		if step.Expect.Surface != EvalCommandSurfaceCLI {
			return fmt.Errorf("steps[%d]: surface must be %q", idx, EvalCommandSurfaceCLI)
		}
		if step.Expect.ExitCode < 0 {
			return fmt.Errorf("steps[%d]: exitCode must be >= 0", idx)
		}
		switch step.Expect.Command {
		case EvalCommandInit:
			seenInit = true
		case EvalCommandRefresh:
			if !seenInit {
				return fmt.Errorf("steps[%d]: refresh requires a prior init step", idx)
			}
			seenRefresh = true
		case EvalCommandDoctor:
			if !seenInit {
				return fmt.Errorf("steps[%d]: doctor requires a prior init step", idx)
			}
		case EvalCommandPackExport:
			if !seenRefresh {
				return fmt.Errorf("steps[%d]: pack_export requires a prior refresh step", idx)
			}
		default:
			return fmt.Errorf("steps[%d]: unsupported command %q", idx, step.Expect.Command)
		}
		for _, artifactID := range step.CaptureArtifact {
			if _, ok := artifactIDs[artifactID]; !ok {
				return fmt.Errorf("steps[%d]: captureArtifact references unknown artifact %q", idx, artifactID)
			}
		}
	}

	return nil
}

func LoadEvalScenario(path string) (EvalScenarioDefinition, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return EvalScenarioDefinition{}, err
	}

	var scenario EvalScenarioDefinition
	if err := json.Unmarshal(content, &scenario); err != nil {
		return EvalScenarioDefinition{}, fmt.Errorf("decode scenario %s: %w", path, err)
	}
	if err := scenario.Validate(); err != nil {
		return EvalScenarioDefinition{}, fmt.Errorf("validate scenario %s: %w", path, err)
	}

	return scenario, nil
}

func LoadEvalScenarios(dir string) ([]EvalScenarioDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var scenarios []EvalScenarioDefinition
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		scenario, err := LoadEvalScenario(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, scenario)
	}

	sort.Slice(scenarios, func(i, j int) bool {
		return scenarios[i].ID < scenarios[j].ID
	})
	if err := validateUniqueScenarioIDs(scenarios); err != nil {
		return nil, err
	}
	return scenarios, nil
}

func validateUniqueScenarioIDs(scenarios []EvalScenarioDefinition) error {
	seen := make(map[string]struct{}, len(scenarios))
	for _, scenario := range scenarios {
		if _, ok := seen[scenario.ID]; ok {
			return fmt.Errorf("duplicate scenario id %q", scenario.ID)
		}
		seen[scenario.ID] = struct{}{}
	}
	return nil
}

func ValidateEvalFixtureReferences(scenarios []EvalScenarioDefinition, fixturesRoot string) error {
	for _, scenario := range scenarios {
		fixturePath := filepath.Join(fixturesRoot, filepath.FromSlash(scenario.Fixture.Path))
		info, err := os.Stat(fixturePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("scenario %q fixture path %q does not exist", scenario.ID, scenario.Fixture.Path)
			}
			return fmt.Errorf("scenario %q fixture path %q: %w", scenario.ID, scenario.Fixture.Path, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("scenario %q fixture path %q is not a directory", scenario.ID, scenario.Fixture.Path)
		}
		if !strings.Contains(scenario.Fixture.Path, scenario.Fixture.ID) {
			return fmt.Errorf("scenario %q fixture path %q must contain fixture id %q", scenario.ID, scenario.Fixture.Path, scenario.Fixture.ID)
		}
		if !strings.Contains(scenario.Fixture.Path, scenario.Fixture.Version) {
			return fmt.Errorf("scenario %q fixture path %q must contain fixture version %q", scenario.ID, scenario.Fixture.Path, scenario.Fixture.Version)
		}
	}
	return nil
}

func EvalSupportsCommandSequence(commands ...EvalCommandName) bool {
	return slices.Equal(commands, []EvalCommandName{
		EvalCommandInit,
		EvalCommandRefresh,
		EvalCommandDoctor,
		EvalCommandPackExport,
	})
}
