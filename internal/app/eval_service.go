package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

type EvalService struct {
	Locator       repository.Locator
	Runner        EvalRunner
	OpenStore     func(context.Context, state.Layout, string) (*sqlite.Store, error)
	ResolveLayout func(string) (state.Layout, error)
	Now           func() time.Time
}

type EvalRequirementCoverageReport struct {
	RepositoryRoot    string
	GeneratedAt       time.Time
	EvalArtifactRoot  string
	ScenarioInventory []EvalScenarioCoverageSummary
	Requirements      []EvalRequirementCoverage
}

type EvalScenarioCoverageSummary struct {
	ScenarioID     string
	ScenarioName   string
	Status         string
	Passed         bool
	RunID          int64
	FixtureID      string
	FixtureVersion string
	ArtifactRoot   string
	RerunCommand   string
	ArtifactPaths  []string
}

type EvalRequirementCoverage struct {
	RequirementID string
	Summary       string
	Covered       bool
	ScenarioIDs   []string
	RerunCommands []string
	Evidence      []EvalScenarioCoverageSummary
}

func NewEvalService() EvalService {
	return EvalService{
		Locator: repository.NewLocator(),
		Runner:  NewEvalRunner(),
		OpenStore: func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		},
		ResolveLayout: state.ResolveLayout,
		Now:           time.Now,
	}
}

func (s EvalService) Run(ctx context.Context, request EvalRunRequest) (repository.EvalRunResult, error) {
	root, err := s.Locator.Resolve(request.StartPath)
	if err != nil {
		return repository.EvalRunResult{}, fmt.Errorf("resolve repository root: %w", err)
	}

	layoutResolver := s.ResolveLayout
	if layoutResolver == nil {
		layoutResolver = state.ResolveLayout
	}
	layout, err := layoutResolver(root.RootPath)
	if err != nil {
		return repository.EvalRunResult{}, fmt.Errorf("resolve state layout: %w", err)
	}

	openStore := s.OpenStore
	if openStore == nil {
		openStore = func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		}
	}

	store, err := openStore(ctx, layout, root.DetectionMode)
	if err != nil {
		return repository.EvalRunResult{}, fmt.Errorf("open eval store: %w", err)
	}
	defer store.Close()

	repoRecord, err := store.UpsertRepository(ctx, root, s.nowUTC())
	if err != nil {
		return repository.EvalRunResult{}, fmt.Errorf("persist repository metadata: %w", err)
	}

	runner := s.Runner.withDefaults()
	result, runErr := runner.Run(ctx, request)
	if result.ScenarioID == "" {
		return result, runErr
	}

	record := sqlite.EvalRunRecord{
		RepositoryID:    repoRecord.ID,
		ScenarioID:      result.Scenario.ID,
		ScenarioVersion: result.Scenario.Version,
		FixtureID:       result.Scenario.Fixture.ID,
		FixtureVersion:  result.Scenario.Fixture.Version,
		Status:          sqlite.EvalRunStatusRunning,
		Passed:          false,
		WorkspacePath:   result.WorkspacePath,
		ArtifactRoot:    layout.EvalDir,
		StartedAt:       result.StartedAt.UTC(),
	}
	record, err = store.SaveEvalRun(ctx, record, nil, nil)
	if err != nil {
		return result, combineEvalErrors(runErr, fmt.Errorf("persist eval run header: %w", err))
	}

	artifactRoot := layout.EvalRunDir(record.ID)
	if err := os.MkdirAll(artifactRoot, 0o755); err != nil {
		return result, combineEvalErrors(runErr, fmt.Errorf("create eval artifact root: %w", err))
	}

	stepRecords, artifactRecords, err := persistEvalEvidence(artifactRoot, result)
	if err != nil {
		return result, combineEvalErrors(runErr, err)
	}

	record.Status = evalRunStatus(result, runErr)
	record.Passed = result.Passed
	record.ArtifactRoot = artifactRoot
	record.CompletedAt = result.FinishedAt.UTC()
	record.MetadataJSON = mustMarshalEvalMetadata(map[string]any{
		"scenarioName": result.Scenario.Name,
		"workspace":    result.WorkspacePath,
	})
	record, err = store.SaveEvalRun(ctx, record, stepRecords, artifactRecords)
	if err != nil {
		return result, combineEvalErrors(runErr, fmt.Errorf("persist eval run details: %w", err))
	}

	result.PersistedRunID = record.ID
	result.PersistedArtifactRoot = artifactRoot
	return result, runErr
}

func (s EvalService) RequirementCoverageReport(ctx context.Context, startPath string) (EvalRequirementCoverageReport, error) {
	root, err := s.Locator.Resolve(startPath)
	if err != nil {
		return EvalRequirementCoverageReport{}, fmt.Errorf("resolve repository root: %w", err)
	}

	layoutResolver := s.ResolveLayout
	if layoutResolver == nil {
		layoutResolver = state.ResolveLayout
	}
	layout, err := layoutResolver(root.RootPath)
	if err != nil {
		return EvalRequirementCoverageReport{}, fmt.Errorf("resolve state layout: %w", err)
	}

	openStore := s.OpenStore
	if openStore == nil {
		openStore = func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		}
	}

	store, err := openStore(ctx, layout, root.DetectionMode)
	if err != nil {
		return EvalRequirementCoverageReport{}, fmt.Errorf("open eval store: %w", err)
	}
	defer store.Close()

	repositoryID, err := store.LookupRepositoryID(ctx, root.RootPath)
	if err != nil {
		return EvalRequirementCoverageReport{}, fmt.Errorf("lookup repository: %w", err)
	}
	runs, err := store.ListEvalRuns(ctx, repositoryID)
	if err != nil {
		return EvalRequirementCoverageReport{}, fmt.Errorf("list eval runs: %w", err)
	}

	latestByScenario := make(map[string]EvalScenarioCoverageSummary, len(runs))
	for _, run := range runs {
		if _, exists := latestByScenario[run.ScenarioID]; exists {
			continue
		}
		_, _, artifacts, err := store.LoadEvalRun(ctx, run.ID)
		if err != nil {
			return EvalRequirementCoverageReport{}, fmt.Errorf("load eval run %d: %w", run.ID, err)
		}
		latestByScenario[run.ScenarioID] = evalScenarioCoverageSummary(run, artifacts)
	}

	inventoryIDs := make([]string, 0, len(latestByScenario))
	for scenarioID := range latestByScenario {
		inventoryIDs = append(inventoryIDs, scenarioID)
	}
	sort.Strings(inventoryIDs)

	inventory := make([]EvalScenarioCoverageSummary, 0, len(inventoryIDs))
	for _, scenarioID := range inventoryIDs {
		inventory = append(inventory, latestByScenario[scenarioID])
	}

	requirements := make([]EvalRequirementCoverage, 0, len(evalRequirementCoverageDefinitions))
	for _, definition := range evalRequirementCoverageDefinitions {
		requirement := EvalRequirementCoverage{
			RequirementID: definition.RequirementID,
			Summary:       definition.Summary,
			ScenarioIDs:   append([]string(nil), definition.ScenarioIDs...),
		}
		requirement.RerunCommands = make([]string, 0, len(definition.ScenarioIDs))
		requirement.Evidence = make([]EvalScenarioCoverageSummary, 0, len(definition.ScenarioIDs))
		requirement.Covered = true
		for _, scenarioID := range definition.ScenarioIDs {
			requirement.RerunCommands = append(requirement.RerunCommands, evalScenarioRerunCommand(scenarioID))
			summary, ok := latestByScenario[scenarioID]
			if !ok {
				requirement.Covered = false
				continue
			}
			requirement.Evidence = append(requirement.Evidence, summary)
			if !summary.Passed {
				requirement.Covered = false
			}
		}
		requirements = append(requirements, requirement)
	}

	return EvalRequirementCoverageReport{
		RepositoryRoot:    root.RootPath,
		GeneratedAt:       s.nowUTC(),
		EvalArtifactRoot:  layout.EvalDir,
		ScenarioInventory: inventory,
		Requirements:      requirements,
	}, nil
}

func (s EvalService) nowUTC() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

type evalRequirementCoverageDefinition struct {
	RequirementID string
	Summary       string
	ScenarioIDs   []string
}

var evalRequirementCoverageDefinitions = []evalRequirementCoverageDefinition{
	{
		RequirementID: "EVAL-02",
		Summary:       "Repeatable MCP evidence must prove the shipped stdio readiness, initialize, tools/list, and full query or ops tool surface from persisted transcript and tool-response artifacts.",
		ScenarioIDs:   []string{"mcp-go-basic-v1", "mcp-go-worktree-v1"},
	},
	{
		RequirementID: "EVAL-03",
		Summary:       "Functional evidence must prove stale, partially degraded, and recovered runtime behavior through the shipped CLI or MCP surfaces with persisted repo-local artifacts.",
		ScenarioIDs:   []string{"cli-go-stale-v1", "mcp-go-degraded-v1", "mcp-go-recovery-v1"},
	},
}

func evalScenarioCoverageSummary(run sqlite.EvalRunRecord, artifacts []sqlite.EvalArtifactRecord) EvalScenarioCoverageSummary {
	paths := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		if artifact.StoredPath == "" {
			continue
		}
		paths = append(paths, artifact.StoredPath)
	}
	sort.Strings(paths)

	return EvalScenarioCoverageSummary{
		ScenarioID:     run.ScenarioID,
		ScenarioName:   evalScenarioDisplayName(run),
		Status:         string(run.Status),
		Passed:         run.Passed,
		RunID:          run.ID,
		FixtureID:      run.FixtureID,
		FixtureVersion: run.FixtureVersion,
		ArtifactRoot:   run.ArtifactRoot,
		RerunCommand:   evalScenarioRerunCommand(run.ScenarioID),
		ArtifactPaths:  paths,
	}
}

func evalScenarioDisplayName(run sqlite.EvalRunRecord) string {
	if run.MetadataJSON == "" {
		return run.ScenarioID
	}
	var metadata map[string]any
	if err := json.Unmarshal([]byte(run.MetadataJSON), &metadata); err != nil {
		return run.ScenarioID
	}
	name, _ := metadata["scenarioName"].(string)
	if name == "" {
		return run.ScenarioID
	}
	return name
}

func evalScenarioRerunCommand(scenarioID string) string {
	return fmt.Sprintf("go run ./cmd/optimusctx eval --scenario %s", scenarioID)
}

func persistEvalEvidence(artifactRoot string, result repository.EvalRunResult) ([]sqlite.EvalStepRecord, []sqlite.EvalArtifactRecord, error) {
	stepRecords := make([]sqlite.EvalStepRecord, 0, len(result.Steps))
	artifactRecords := make([]sqlite.EvalArtifactRecord, 0, len(result.Artifacts))
	seenArtifacts := make(map[string]struct{}, len(result.Artifacts))

	for index, step := range result.Steps {
		stepDir := filepath.Join(artifactRoot, step.Step.ID)
		surface, command := evalStepStorageIdentity(step.Step)
		record := sqlite.EvalStepRecord{
			StepID:       step.Step.ID,
			Ordinal:      index,
			Name:         step.Step.Name,
			Surface:      surface,
			Command:      command,
			ArgsJSON:     mustMarshalEvalMetadata(buildEvalStepArgs(step.Step, indexEvalArtifacts(result.Scenario.Artifacts))),
			ExitCode:     step.ExitCode,
			Passed:       step.Passed,
			StartedAt:    step.StartedAt.UTC(),
			FinishedAt:   step.FinishedAt.UTC(),
			MetadataJSON: mustMarshalEvalMetadata(map[string]any{"artifactCount": len(step.Artifacts)}),
		}
		if step.Stdout != "" {
			record.StdoutPath = filepath.Join(stepDir, "stdout.txt")
			if err := writeEvalArtifactFile(record.StdoutPath, []byte(step.Stdout)); err != nil {
				return nil, nil, fmt.Errorf("persist stdout for step %q: %w", step.Step.ID, err)
			}
		}
		if step.Stderr != "" {
			record.StderrPath = filepath.Join(stepDir, "stderr.txt")
			if err := writeEvalArtifactFile(record.StderrPath, []byte(step.Stderr)); err != nil {
				return nil, nil, fmt.Errorf("persist stderr for step %q: %w", step.Step.ID, err)
			}
		}
		stepRecords = append(stepRecords, record)

		for _, artifact := range step.Artifacts {
			storedPath, err := persistStepArtifact(stepDir, artifact, record)
			if err != nil {
				return nil, nil, fmt.Errorf("persist artifact %q: %w", artifact.Ref.ID, err)
			}
			artifactRecords = append(artifactRecords, sqlite.EvalArtifactRecord{
				StepID:       step.Step.ID,
				ArtifactID:   artifact.Ref.ID,
				Kind:         string(artifact.Ref.Kind),
				LogicalPath:  artifact.Ref.Path,
				StoredPath:   storedPath,
				Required:     artifact.Ref.Required,
				Present:      artifact.Present,
				SizeBytes:    artifact.Bytes,
				MetadataJSON: mustMarshalEvalMetadata(map[string]any{"location": artifact.Location}),
			})
			seenArtifacts[artifact.Ref.ID] = struct{}{}
		}
	}

	for _, artifact := range result.Artifacts {
		if _, ok := seenArtifacts[artifact.Ref.ID]; ok {
			continue
		}
		storedPath, err := persistArtifactResult(artifactRoot, artifact)
		if err != nil {
			return nil, nil, fmt.Errorf("persist artifact %q: %w", artifact.Ref.ID, err)
		}
		artifactRecords = append(artifactRecords, sqlite.EvalArtifactRecord{
			ArtifactID:   artifact.Ref.ID,
			Kind:         string(artifact.Ref.Kind),
			LogicalPath:  artifact.Ref.Path,
			StoredPath:   storedPath,
			Required:     artifact.Ref.Required,
			Present:      artifact.Present,
			SizeBytes:    artifact.Bytes,
			MetadataJSON: mustMarshalEvalMetadata(map[string]any{"location": artifact.Location}),
		})
	}

	return stepRecords, artifactRecords, nil
}

func evalStepStorageIdentity(step repository.EvalScenarioStep) (string, string) {
	switch step.Kind {
	case repository.EvalStepKindCommand:
		return string(step.Expect.Surface), string(step.Expect.Command)
	case repository.EvalStepKindMCPSession:
		return "mcp", string(step.Kind)
	default:
		return string(step.Kind), string(step.Kind)
	}
}

func persistStepArtifact(stepDir string, artifact repository.EvalArtifactResult, record sqlite.EvalStepRecord) (string, error) {
	switch artifact.Ref.Kind {
	case repository.EvalArtifactKindStdout:
		return record.StdoutPath, nil
	case repository.EvalArtifactKindStderr:
		return record.StderrPath, nil
	default:
		return persistArtifactResult(filepath.Dir(stepDir), artifact)
	}
}

func persistArtifactResult(artifactRoot string, artifact repository.EvalArtifactResult) (string, error) {
	switch artifact.Ref.Kind {
	case repository.EvalArtifactKindStdout, repository.EvalArtifactKindStderr:
		return artifact.Location, nil
	case repository.EvalArtifactKindFile:
		if artifact.Ref.Path == "" {
			return "", nil
		}
		storedPath := filepath.Join(artifactRoot, filepath.FromSlash(artifact.Ref.Path))
		if !artifact.Present || artifact.Location == "" {
			return storedPath, nil
		}
		if err := copyEvalFile(artifact.Location, storedPath); err != nil {
			return "", err
		}
		return storedPath, nil
	default:
		return "", nil
	}
}

func writeEvalArtifactFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func evalRunStatus(result repository.EvalRunResult, runErr error) sqlite.EvalRunStatus {
	if result.Passed {
		return sqlite.EvalRunStatusPassed
	}
	if runErr != nil && len(result.Steps) == 0 {
		return sqlite.EvalRunStatusError
	}
	return sqlite.EvalRunStatusFailed
}

func combineEvalErrors(runErr error, persistErr error) error {
	if persistErr == nil {
		return runErr
	}
	if runErr == nil {
		return persistErr
	}
	return fmt.Errorf("%w; %v", runErr, persistErr)
}

func mustMarshalEvalMetadata(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(data)
}
