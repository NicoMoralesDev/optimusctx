package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

func TestEvalWorkspaceMutations(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Initial() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "obsolete.txt"), "remove me\n")

	refresh := NewRefreshService()
	initial, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	})
	if err != nil {
		t.Fatalf("Refresh() baseline error = %v", err)
	}

	controls, err := applyEvalSetupActions(repoRoot, []repository.EvalSetupAction{
		{Kind: repository.EvalSetupActionWriteFile, Path: "README.md", Content: "# fixture\n"},
		{Kind: repository.EvalSetupActionOverwriteFile, Path: "main.go", Content: "package main\n\nfunc Mutated() {}\n"},
		{Kind: repository.EvalSetupActionDeleteFile, Path: "obsolete.txt"},
		{
			Kind: repository.EvalSetupActionSeedWatchStatus,
			WatchStatus: &repository.EvalWatchStatusSeed{
				PID:                    777,
				BinaryVersion:          "dev",
				StartedAt:              "2026-03-15T17:59:50Z",
				LastHeartbeatAt:        "2026-03-15T18:00:00Z",
				LastEventAt:            "2026-03-15T18:00:00Z",
				LastRefreshCompletedAt: "2026-03-15T18:00:00Z",
				LastRefreshGeneration:  initial.Generation,
				LastError:              "watch observer overflowed; falling back to full refresh",
			},
		},
		{
			Kind:           repository.EvalSetupActionInjectRefreshFailure,
			FailureStage:   "after_files",
			FailureMessage: "forced eval failure",
		},
	})
	if err != nil {
		t.Fatalf("applyEvalSetupActions() error = %v", err)
	}
	if controls.RefreshFailure == nil {
		t.Fatal("expected refresh failure control")
	}

	readme, err := os.ReadFile(filepath.Join(repoRoot, "README.md"))
	if err != nil {
		t.Fatalf("ReadFile(README.md) error = %v", err)
	}
	if string(readme) != "# fixture\n" {
		t.Fatalf("README.md = %q", string(readme))
	}
	mainContent, err := os.ReadFile(filepath.Join(repoRoot, "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(mainContent), "Mutated") {
		t.Fatalf("main.go = %q, want mutated content", string(mainContent))
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "obsolete.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("obsolete.txt should be deleted, err=%v", err)
	}

	watchStatusPath := filepath.Join(repoRoot, ".optimusctx", "tmp", repository.DefaultWatchStatusFilename)
	statusPayload, err := os.ReadFile(watchStatusPath)
	if err != nil {
		t.Fatalf("ReadFile(watch-status.json) error = %v", err)
	}
	var record repository.WatchStatusRecord
	if err := json.Unmarshal(statusPayload, &record); err != nil {
		t.Fatalf("Unmarshal(watch-status.json) error = %v", err)
	}
	if record.RepoRoot != repoRoot || record.LastRefreshGeneration != initial.Generation {
		t.Fatalf("watch status record = %+v", record)
	}
	if record.LastError != "watch observer overflowed; falling back to full refresh" {
		t.Fatalf("watch status error = %q", record.LastError)
	}

	_, refreshErr := withEvalStepControls(controls, func() (repository.EvalStepResult, error) {
		result, err := refresh.Refresh(context.Background(), RefreshRequest{
			StartPath: repoRoot,
			Reason:    repository.RefreshReasonManual,
		})
		if err == nil {
			t.Fatal("Refresh() error = nil, want injected failure")
		}
		if result.FreshnessStatus != repository.FreshnessStatusPartiallyDegraded {
			t.Fatalf("freshness = %q, want partially_degraded", result.FreshnessStatus)
		}
		if result.Generation != initial.Generation+1 {
			t.Fatalf("generation = %d, want %d", result.Generation, initial.Generation+1)
		}
		return repository.EvalStepResult{}, err
	})
	if refreshErr == nil || !strings.Contains(refreshErr.Error(), "forced eval failure") {
		t.Fatalf("Refresh() error = %v, want forced eval failure", refreshErr)
	}
}

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

func TestEvalMCPSessionExecution(t *testing.T) {
	sourceRoot := t.TempDir()
	fixturesRoot := filepath.Join(sourceRoot, "fixtures")
	scenarioPath := filepath.Join(sourceRoot, "scenarios", "mcp.json")

	writeEvalFixtureFile(t, filepath.Join(fixturesRoot, "go-basic", "v1", "repository", "main.go"), "package main\n")
	writeEvalScenarioFile(t, scenarioPath, repository.EvalScenarioDefinition{
		SchemaVersion: repository.EvalScenarioSchemaV1,
		ID:            "mcp",
		Version:       "v1",
		Name:          "MCP",
		Fixture: repository.EvalFixtureRef{
			ID:           "go-basic",
			Version:      "v1",
			Path:         "go-basic/v1/repository",
			Materialize:  repository.EvalFixtureModeCopyTree,
			WorkspaceDir: "workspace",
		},
		Steps: []repository.EvalScenarioStep{{
			ID:   "mcp-serve",
			Name: "Run MCP session",
			Kind: repository.EvalStepKindMCPSession,
			Session: &repository.EvalMCPSession{
				Requests: []repository.EvalMCPRequest{
					{ID: 1, Method: "initialize"},
					{Method: "notifications/initialized", Notification: true},
					{ID: 2, Method: "tools/list"},
					{ID: 3, Method: "tools/call", Params: map[string]any{
						"name": "optimusctx.repository_map",
					}},
				},
				TranscriptArtifact: "transcript",
				CaptureResponses: []repository.EvalMCPResponseCapture{
					{RequestID: 2, Artifact: "tools-list"},
					{RequestID: 3, Artifact: "repository-map"},
				},
			},
			Assert: []repository.EvalAssertion{
				{Kind: repository.EvalAssertionKindContains, Target: repository.EvalAssertionTargetStderr, Contains: "ready for stdio requests"},
				{Kind: repository.EvalAssertionKindContains, Target: repository.EvalAssertionTargetArtifact, Artifact: "tools-list", Contains: "optimusctx.repository_map"},
				{Kind: repository.EvalAssertionKindJSONFieldEquals, Target: repository.EvalAssertionTargetArtifact, Artifact: "repository-map", Path: "result.structuredContent.meta.cacheStatus", Equals: "persisted_only"},
			},
			CaptureArtifact: []string{"transcript", "tools-list", "repository-map", "session-stderr"},
		}},
		Artifacts: []repository.EvalArtifactRef{
			{ID: "transcript", Kind: repository.EvalArtifactKindFile, Path: "artifacts/transcript.json", Required: true},
			{ID: "tools-list", Kind: repository.EvalArtifactKindFile, Path: "artifacts/tools-list.json", Required: true},
			{ID: "repository-map", Kind: repository.EvalArtifactKindFile, Path: "artifacts/repository-map.json", Required: true},
			{ID: "session-stderr", Kind: repository.EvalArtifactKindStderr, Required: true},
		},
	})

	runner := NewEvalRunner()
	runner.Now = newDeterministicEvalClock()
	runner.RunCommand = func(_ context.Context, invocation EvalCommandInvocation) (EvalCommandExecutionResult, error) {
		t.Fatalf("RunCommand should not be called for MCP session steps: %+v", invocation)
		return EvalCommandExecutionResult{}, nil
	}
	runner.RunMCPSession = func(_ context.Context, invocation EvalMCPSessionInvocation) (EvalMCPSessionExecutionResult, error) {
		if invocation.WorkingDir == "" {
			t.Fatal("expected working directory")
		}
		if len(invocation.Session.Requests) != 4 {
			t.Fatalf("len(invocation.Session.Requests) = %d, want 4", len(invocation.Session.Requests))
		}
		return EvalMCPSessionExecutionResult{
			Stdout: "",
			Stderr: "optimusctx mcp: ready for stdio requests\n",
			Responses: []EvalMCPSessionResponse{
				{RequestID: 1, Response: map[string]any{"jsonrpc": "2.0", "id": float64(1), "result": map[string]any{"protocolVersion": "2026-03-15"}}},
				{RequestID: 2, Response: map[string]any{"jsonrpc": "2.0", "id": float64(2), "result": map[string]any{"tools": []any{map[string]any{"name": "optimusctx.repository_map"}}}}},
				{RequestID: 3, Response: map[string]any{"jsonrpc": "2.0", "id": float64(3), "result": map[string]any{"structuredContent": map[string]any{"meta": map[string]any{"cacheStatus": "persisted_only"}}}}},
			},
		}, nil
	}

	result, err := runner.Run(context.Background(), EvalRunRequest{
		ScenarioPath: scenarioPath,
		FixturesRoot: fixturesRoot,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !result.Passed {
		t.Fatalf("Run() passed = false, want true: %+v", result)
	}
	if len(result.Steps) != 1 {
		t.Fatalf("len(result.Steps) = %d, want 1", len(result.Steps))
	}
	step := result.Steps[0]
	if step.ExitCode != 0 || !step.Passed {
		t.Fatalf("step result = %+v", step)
	}
	if len(step.Artifacts) != 4 {
		t.Fatalf("len(step.Artifacts) = %d, want 4", len(step.Artifacts))
	}
	if _, err := os.Stat(filepath.Join(result.WorkspacePath, "artifacts", "transcript.json")); err != nil {
		t.Fatalf("expected transcript artifact: %v", err)
	}
}

func TestEvalRunnerPersistsMCPSessionEvidence(t *testing.T) {
	sourceRoot := t.TempDir()
	fixturesRoot := filepath.Join(sourceRoot, "fixtures")
	scenarioPath := filepath.Join(sourceRoot, "scenarios", "mcp-persist.json")

	writeEvalFixtureFile(t, filepath.Join(fixturesRoot, "go-basic", "v1", "repository", "main.go"), "package main\n")
	writeEvalScenarioFile(t, scenarioPath, repository.EvalScenarioDefinition{
		SchemaVersion: repository.EvalScenarioSchemaV1,
		ID:            "mcp-persist",
		Version:       "v1",
		Name:          "MCP Persist",
		Fixture: repository.EvalFixtureRef{
			ID:           "go-basic",
			Version:      "v1",
			Path:         "go-basic/v1/repository",
			Materialize:  repository.EvalFixtureModeCopyTree,
			WorkspaceDir: "workspace",
		},
		Steps: []repository.EvalScenarioStep{{
			ID:   "mcp-serve",
			Name: "Run MCP session",
			Kind: repository.EvalStepKindMCPSession,
			Session: &repository.EvalMCPSession{
				Requests: []repository.EvalMCPRequest{
					{ID: 1, Method: "initialize"},
					{Method: "notifications/initialized", Notification: true},
					{ID: 2, Method: "tools/list"},
				},
				TranscriptArtifact: "transcript",
				CaptureResponses: []repository.EvalMCPResponseCapture{
					{RequestID: 2, Artifact: "tools-list"},
				},
			},
			CaptureArtifact: []string{"session-stderr", "transcript", "tools-list"},
		}},
		Artifacts: []repository.EvalArtifactRef{
			{ID: "session-stderr", Kind: repository.EvalArtifactKindStderr, Required: true},
			{ID: "transcript", Kind: repository.EvalArtifactKindFile, Path: "artifacts/transcript.json", Required: true},
			{ID: "tools-list", Kind: repository.EvalArtifactKindFile, Path: "artifacts/tools-list.json", Required: true},
		},
	})

	runner := NewEvalRunner()
	runner.Now = newDeterministicEvalClock()
	runner.RunMCPSession = func(_ context.Context, invocation EvalMCPSessionInvocation) (EvalMCPSessionExecutionResult, error) {
		return EvalMCPSessionExecutionResult{
			Stderr: "optimusctx mcp: ready for stdio requests\n",
			Responses: []EvalMCPSessionResponse{
				{RequestID: 1, Response: map[string]any{"jsonrpc": "2.0", "id": float64(1), "result": map[string]any{"protocolVersion": "2026-03-15"}}},
				{RequestID: 2, Response: map[string]any{"jsonrpc": "2.0", "id": float64(2), "result": map[string]any{"tools": []any{map[string]any{"name": "optimusctx.repository_map"}}}}},
			},
		}, nil
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
	if len(stepRecords) != 1 {
		t.Fatalf("len(stepRecords) = %d, want 1", len(stepRecords))
	}
	if stepRecords[0].StderrPath == "" {
		t.Fatalf("expected stderr path for MCP step: %+v", stepRecords[0])
	}
	if stepRecords[0].Surface != "mcp" || stepRecords[0].Command != string(repository.EvalStepKindMCPSession) {
		t.Fatalf("step storage identity = %+v", stepRecords[0])
	}
	if _, err := os.Stat(stepRecords[0].StderrPath); err != nil {
		t.Fatalf("Stat(stderr path) error = %v", err)
	}
	transcript, ok := findEvalArtifactRecord(artifactRecords, "transcript")
	if !ok {
		t.Fatalf("missing transcript artifact: %+v", artifactRecords)
	}
	if transcript.StoredPath != filepath.Join(artifactRoot, "artifacts", "transcript.json") {
		t.Fatalf("transcript stored path = %q", transcript.StoredPath)
	}
	toolsList, ok := findEvalArtifactRecord(artifactRecords, "tools-list")
	if !ok {
		t.Fatalf("missing tools-list artifact: %+v", artifactRecords)
	}
	content, err := os.ReadFile(toolsList.StoredPath)
	if err != nil {
		t.Fatalf("ReadFile(tools-list) error = %v", err)
	}
	if !strings.Contains(string(content), "optimusctx.repository_map") {
		t.Fatalf("tools-list content = %q", string(content))
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

func TestEvalRequirementCoverageReport(t *testing.T) {
	repoRoot := initRepo(t)
	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, err := sqlite.OpenOrCreateStore(context.Background(), layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("OpenOrCreateStore() error = %v", err)
	}
	defer store.Close()

	repoRecord, err := store.UpsertRepository(context.Background(), testRepoRoot(repoRoot), time.Date(2026, 3, 16, 13, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("UpsertRepository() error = %v", err)
	}

	saveRun := func(startedAt time.Time, scenarioID string, scenarioName string, artifactName string) int64 {
		run, err := store.SaveEvalRun(context.Background(), sqlite.EvalRunRecord{
			RepositoryID:    repoRecord.ID,
			ScenarioID:      scenarioID,
			ScenarioVersion: "v1",
			FixtureID:       "go-basic",
			FixtureVersion:  "v1",
			Status:          sqlite.EvalRunStatusPassed,
			Passed:          true,
			WorkspacePath:   repoRoot,
			ArtifactRoot:    layout.EvalDir,
			StartedAt:       startedAt,
			CompletedAt:     startedAt.Add(time.Minute),
			MetadataJSON:    `{"scenarioName":"` + scenarioName + `"}`,
		}, nil, []sqlite.EvalArtifactRecord{{
			ArtifactID: artifactName,
			Kind:       "file",
			StoredPath: filepath.Join(layout.EvalDir, scenarioID, artifactName+".json"),
			Required:   true,
			Present:    true,
			SizeBytes:  128,
		}})
		if err != nil {
			t.Fatalf("SaveEvalRun(%s) error = %v", scenarioID, err)
		}
		return run.ID
	}

	recoveryRunID := saveRun(time.Date(2026, 3, 16, 13, 5, 0, 0, time.UTC), "mcp-go-recovery-v1", "MCP degraded-to-recovery flow on basic Go fixture", "refresh-recovered")
	saveRun(time.Date(2026, 3, 16, 13, 1, 0, 0, time.UTC), "mcp-go-basic-v1", "MCP initialize and tools/list on basic Go fixture", "mcp-transcript")
	saveRun(time.Date(2026, 3, 16, 13, 2, 0, 0, time.UTC), "mcp-go-worktree-v1", "MCP full tool flow on nested worktree fixture", "health-response")
	saveRun(time.Date(2026, 3, 16, 13, 3, 0, 0, time.UTC), "cli-go-stale-v1", "CLI stale and watch-diagnostic flow on basic Go fixture", "doctor-stdout")
	saveRun(time.Date(2026, 3, 16, 13, 4, 0, 0, time.UTC), "mcp-go-degraded-v1", "MCP degraded refresh flow on basic Go fixture", "refresh-error")

	service := NewEvalService()
	report, err := service.RequirementCoverageReport(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("RequirementCoverageReport() error = %v", err)
	}

	if report.RepositoryRoot != repoRoot {
		t.Fatalf("RepositoryRoot = %q, want %q", report.RepositoryRoot, repoRoot)
	}
	if report.EvalArtifactRoot != layout.EvalDir {
		t.Fatalf("EvalArtifactRoot = %q, want %q", report.EvalArtifactRoot, layout.EvalDir)
	}
	if got, want := len(report.ScenarioInventory), 5; got != want {
		t.Fatalf("len(ScenarioInventory) = %d, want %d", got, want)
	}
	if report.ScenarioInventory[0].ScenarioID != "cli-go-stale-v1" {
		t.Fatalf("first inventory scenario = %q, want cli-go-stale-v1", report.ScenarioInventory[0].ScenarioID)
	}
	if report.ScenarioInventory[4].ScenarioID != "mcp-go-worktree-v1" {
		t.Fatalf("last inventory scenario = %q, want mcp-go-worktree-v1", report.ScenarioInventory[4].ScenarioID)
	}

	if got, want := len(report.Requirements), 2; got != want {
		t.Fatalf("len(Requirements) = %d, want %d", got, want)
	}
	if !report.Requirements[0].Covered || report.Requirements[0].RequirementID != "EVAL-02" {
		t.Fatalf("EVAL-02 coverage = %+v", report.Requirements[0])
	}
	if !report.Requirements[1].Covered || report.Requirements[1].RequirementID != "EVAL-03" {
		t.Fatalf("EVAL-03 coverage = %+v", report.Requirements[1])
	}
	if got, want := report.Requirements[0].RerunCommands[0], "go run ./cmd/optimusctx eval --scenario mcp-go-basic-v1"; got != want {
		t.Fatalf("EVAL-02 rerun command = %q, want %q", got, want)
	}
	if got, want := report.Requirements[1].Evidence[2].RunID, recoveryRunID; got != want {
		t.Fatalf("recovery run ID = %d, want %d", got, want)
	}
	if got, want := report.Requirements[1].Evidence[2].ArtifactRoot, layout.EvalDir; got != want {
		t.Fatalf("recovery artifact root = %q, want %q", got, want)
	}
	if got, want := report.Requirements[1].Evidence[2].ArtifactPaths[0], filepath.Join(layout.EvalDir, "mcp-go-recovery-v1", "refresh-recovered.json"); got != want {
		t.Fatalf("recovery artifact path = %q, want %q", got, want)
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

func findEvalArtifactRecord(records []sqlite.EvalArtifactRecord, artifactID string) (sqlite.EvalArtifactRecord, bool) {
	for _, record := range records {
		if record.ArtifactID == artifactID {
			return record, true
		}
	}
	return sqlite.EvalArtifactRecord{}, false
}
