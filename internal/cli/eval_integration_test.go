package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/mcp"
	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

func TestEvalMCPInitializeAndToolsList(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "--scenario", "mcp-go-basic-v1"}, &stdout); err != nil {
			t.Fatalf("Execute(eval mcp-go-basic-v1) error = %v", err)
		}
		assertContains(t, stdout.String(), "scenario id: mcp-go-basic-v1")
		assertContains(t, stdout.String(), "status: passed")
		assertContains(t, stdout.String(), "step init: passed exit=0")
		assertContains(t, stdout.String(), "step refresh: passed exit=0")
		assertContains(t, stdout.String(), "step mcp-serve: passed exit=0")
	})
}

func TestEvalMCPToolFlows(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "--scenario", "mcp-go-worktree-v1"}, &stdout); err != nil {
			t.Fatalf("Execute(eval mcp-go-worktree-v1) error = %v", err)
		}
		assertContains(t, stdout.String(), "scenario id: mcp-go-worktree-v1")
		assertContains(t, stdout.String(), "status: passed")
		assertContains(t, stdout.String(), "step init: passed exit=0")
		assertContains(t, stdout.String(), "step refresh: passed exit=0")
		assertContains(t, stdout.String(), "step mcp-serve: passed exit=0")
	})
}

func TestBenchmarkDiscoveryLane(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		runner := app.NewBenchmarkRunner()
		runner.RunCommand = func(ctx context.Context, invocation app.BenchmarkCommandInvocation) (app.BenchmarkCommandExecutionResult, error) {
			execution, err := executeEvalCLICommand(ctx, app.EvalCommandInvocation{
				Args:       invocation.Args,
				WorkingDir: invocation.WorkingDir,
			})
			return app.BenchmarkCommandExecutionResult{
				Stdout:   execution.Stdout,
				Stderr:   execution.Stderr,
				ExitCode: execution.ExitCode,
			}, err
		}
		runner.RunTool = func(ctx context.Context, invocation app.BenchmarkToolInvocation) (app.BenchmarkToolExecutionResult, error) {
			session := repository.EvalMCPSession{
				Requests: []repository.EvalMCPRequest{
					{ID: 1, Method: "initialize", Params: mcp.InitializeParams{
						ClientInfo:      mcp.ClientInfo{Name: "benchmark-test", Version: "1.0.0"},
						ProtocolVersion: "2024-11-05",
					}},
					{Method: "notifications/initialized", Notification: true},
					{ID: 2, Method: "tools/call", Params: mcp.CallToolParams{
						Name:      invocation.Name,
						Arguments: invocation.Arguments,
					}},
				},
			}
			execution, err := executeEvalCLIMCPSession(ctx, app.EvalMCPSessionInvocation{
				WorkingDir: invocation.WorkingDir,
				Session:    session,
			})
			if err != nil {
				return app.BenchmarkToolExecutionResult{}, err
			}
			if len(execution.Responses) == 0 {
				t.Fatal("expected MCP responses")
			}
			payload, err := decodeBenchmarkToolPayload(execution.Responses[len(execution.Responses)-1].Response)
			if err != nil {
				return app.BenchmarkToolExecutionResult{}, err
			}
			return app.BenchmarkToolExecutionResult{Payload: payload}, nil
		}

		result, err := runner.Run(context.Background(), app.BenchmarkRunRequest{
			SuiteID:      "go-benchmark-discovery-v1",
			SuitesDir:    filepath.Join(repoRoot, "testdata", "eval", "benchmarks"),
			FixturesRoot: filepath.Join(repoRoot, "testdata", "eval", "fixtures"),
		})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if len(result.Arms) != 2 {
			t.Fatalf("len(result.Arms) = %d, want 2", len(result.Arms))
		}
		discovery := result.Arms[1].LaneResults[0]
		if discovery.StopMarker != "target_identified" || !discovery.Success {
			t.Fatalf("discovery lane = %+v", discovery)
		}
		if discovery.Effort.TargetedLookupActions == 0 && discovery.Effort.BroadSearchActions == 0 {
			t.Fatalf("discovery effort = %+v", discovery.Effort)
		}
	})
}

func TestBenchmarkRefreshAfterChangeComparison(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		runner := app.NewBenchmarkRunner()
		runner.RunCommand = func(ctx context.Context, invocation app.BenchmarkCommandInvocation) (app.BenchmarkCommandExecutionResult, error) {
			execution, err := executeEvalCLICommand(ctx, app.EvalCommandInvocation{
				Args:       invocation.Args,
				WorkingDir: invocation.WorkingDir,
			})
			return app.BenchmarkCommandExecutionResult{
				Stdout:   execution.Stdout,
				Stderr:   execution.Stderr,
				ExitCode: execution.ExitCode,
			}, err
		}
		runner.RunTool = func(ctx context.Context, invocation app.BenchmarkToolInvocation) (app.BenchmarkToolExecutionResult, error) {
			session := repository.EvalMCPSession{
				Requests: []repository.EvalMCPRequest{
					{ID: 1, Method: "initialize", Params: mcp.InitializeParams{
						ClientInfo:      mcp.ClientInfo{Name: "benchmark-test", Version: "1.0.0"},
						ProtocolVersion: "2024-11-05",
					}},
					{Method: "notifications/initialized", Notification: true},
					{ID: 2, Method: "tools/call", Params: mcp.CallToolParams{
						Name:      invocation.Name,
						Arguments: invocation.Arguments,
					}},
				},
			}
			execution, err := executeEvalCLIMCPSession(ctx, app.EvalMCPSessionInvocation{
				WorkingDir: invocation.WorkingDir,
				Session:    session,
			})
			if err != nil {
				return app.BenchmarkToolExecutionResult{}, err
			}
			payload, err := decodeBenchmarkToolPayload(execution.Responses[len(execution.Responses)-1].Response)
			if err != nil {
				return app.BenchmarkToolExecutionResult{}, err
			}
			return app.BenchmarkToolExecutionResult{Payload: payload}, nil
		}

		result, err := runner.Run(context.Background(), app.BenchmarkRunRequest{
			SuiteID:      "go-benchmark-refresh-v1",
			SuitesDir:    filepath.Join(repoRoot, "testdata", "eval", "benchmarks"),
			FixturesRoot: filepath.Join(repoRoot, "testdata", "eval", "fixtures"),
		})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if len(result.Arms) != 2 {
			t.Fatalf("len(result.Arms) = %d, want 2", len(result.Arms))
		}

		baselineRefresh := result.Arms[0].LaneResults[0]
		if !baselineRefresh.Success || baselineRefresh.StopMarker != "refresh_ready" {
			t.Fatalf("baseline refresh lane = %+v", baselineRefresh)
		}
		if !strings.Contains(strings.Join(baselineRefresh.EvidencePaths, ","), "docs/notes.txt") {
			t.Fatalf("baseline refresh evidence = %+v", baselineRefresh.EvidencePaths)
		}

		treatmentRefresh := result.Arms[1].LaneResults[0]
		if !treatmentRefresh.Success || treatmentRefresh.StopMarker != "refresh_ready" {
			t.Fatalf("treatment refresh lane = %+v", treatmentRefresh)
		}
		if treatmentRefresh.Effort.ActionCount < 2 {
			t.Fatalf("treatment refresh effort = %+v, want cli + mcp actions", treatmentRefresh.Effort)
		}
		if !treatmentRefresh.SetupAppliedAt.Before(treatmentRefresh.StartedAt) {
			t.Fatalf("refresh timing should start after mutation: setup=%s start=%s", treatmentRefresh.SetupAppliedAt, treatmentRefresh.StartedAt)
		}
	})
}

func TestBenchmarkVerificationWorkflow(t *testing.T) {
	t.Run("stable methodology passes", func(t *testing.T) {
		repoRoot := initCLIRepo(t)
		seedCommittedEvalFixtures(t, repoRoot)
		withWorkingDirectory(t, repoRoot, func() {
			service := newCLIBenchmarkService(t)
			result, err := service.RunRepeated(context.Background(), app.BenchmarkRepeatedRunRequest{
				StartPath:    repoRoot,
				SuiteID:      "go-benchmark-refresh-v1",
				SuitesDir:    filepath.Join(repoRoot, "testdata", "eval", "benchmarks"),
				FixturesRoot: filepath.Join(repoRoot, "testdata", "eval", "fixtures"),
				Attempts:     2,
			})
			if err != nil {
				t.Fatalf("RunRepeated() error = %v", err)
			}
			if !result.Summary.Verification.Passed {
				t.Fatalf("verification = %+v", result.Summary.Verification)
			}
			if got, want := len(result.Summary.Arms), 2; got != want {
				t.Fatalf("len(summary.Arms) = %d, want %d", got, want)
			}
			if !strings.Contains(result.Summary.RerunCommand, "optimusctx eval benchmark export --suite go-benchmark-refresh-v1 --attempts 2") {
				t.Fatalf("rerun command = %q", result.Summary.RerunCommand)
			}
		})
	})

	t.Run("drift fails verification", func(t *testing.T) {
		repoRoot := initCLIRepo(t)
		seedCommittedEvalFixtures(t, repoRoot)
		withWorkingDirectory(t, repoRoot, func() {
			service := newCLIBenchmarkService(t)
			suitePath := filepath.Join(repoRoot, "testdata", "eval", "benchmarks", "go-benchmark-refresh-v1.json")
			baseLoadSuiteFile := service.Runner.LoadSuiteFile
			var loadCount int
			service.Runner.LoadSuiteFile = func(path string) (repository.BenchmarkSuiteDefinition, error) {
				suite, err := baseLoadSuiteFile(path)
				if err != nil {
					return repository.BenchmarkSuiteDefinition{}, err
				}
				if path == suitePath {
					loadCount++
					if loadCount > 2 {
						suite.Lanes[0].StopCondition.Marker = "refresh_ready_drifted"
					}
				}
				return suite, nil
			}

			verification, err := service.VerifyMethodology(context.Background(), app.BenchmarkRepeatedRunRequest{
				StartPath:    repoRoot,
				SuitePath:    suitePath,
				FixturesRoot: filepath.Join(repoRoot, "testdata", "eval", "fixtures"),
				Attempts:     2,
			})
			if err != nil {
				t.Fatalf("VerifyMethodology() error = %v", err)
			}
			if verification.Passed {
				t.Fatalf("verification = %+v, want drift failure", verification)
			}
			if !strings.Contains(verification.FailureReason, "drifted from frozen methodology") &&
				!strings.Contains(verification.FailureReason, "did not satisfy stop condition") {
				t.Fatalf("failure reason = %q", verification.FailureReason)
			}
		})
	})
}

func TestBenchmarkHumanReadableReport(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "benchmark", "report", "--suite", "go-benchmark-refresh-v1", "--attempts", "2"}, &stdout); err != nil {
			t.Fatalf("Execute(eval benchmark report) error = %v", err)
		}
		output := stdout.String()
		assertContains(t, output, "benchmark report")
		assertContains(t, output, "suite: go-benchmark-refresh-v1@v1")
		assertContains(t, output, "lane comparison")
		assertContains(t, output, "rerun")
	})
}

func TestBenchmarkAttributionTables(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		render := func(suiteID string) string {
			var stdout bytes.Buffer
			if err := NewRootCommand().Execute([]string{"eval", "benchmark", "report", "--suite", suiteID, "--attempts", "2"}, &stdout); err != nil {
				t.Fatalf("Execute(eval benchmark report %s) error = %v", suiteID, err)
			}
			return stdout.String()
		}
		refreshOutput := render("go-benchmark-refresh-v1")
		assertContains(t, refreshOutput, "treatment artifact attribution")
		assertContains(t, refreshOutput, "L2 Context")
		assertContains(t, refreshOutput, "Pack Export")

		discoveryOutput := render("go-benchmark-discovery-v1")
		assertContains(t, discoveryOutput, "Repository Map")
		assertContains(t, discoveryOutput, "Exact Lookup")
	})
}

func TestBenchmarkReportWordingGuards(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "benchmark", "report", "--suite", "go-benchmark-refresh-v1", "--attempts", "2"}, &stdout); err != nil {
			t.Fatalf("Execute(eval benchmark report) error = %v", err)
		}
		output := stdout.String()
		assertContains(t, output, "estimated tokens use bytes_div_4_ceiling")
		assertContains(t, output, "not provider-billed token invoices")
		for _, banned := range []string{"provider billing", "statistically significant", "universal savings"} {
			if strings.Contains(output, banned) {
				t.Fatalf("report should not contain %q:\n%s", banned, output)
			}
		}
	})
}

func TestBenchmarkMilestoneVerification(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "benchmark", "export", "--suite", "go-benchmark-refresh-v1", "--attempts", "2"}, &stdout); err != nil {
			t.Fatalf("Execute(eval benchmark export) error = %v", err)
		}

		stdout.Reset()
		if err := NewRootCommand().Execute([]string{"eval", "benchmark", "verify", "--suite", "go-benchmark-refresh-v1", "--attempts", "2"}, &stdout); err != nil {
			t.Fatalf("Execute(eval benchmark verify) error = %v", err)
		}
		output := stdout.String()
		assertContains(t, output, "benchmark milestone verification")
		assertContains(t, output, "status: passed")
		assertContains(t, output, "reproducibility verification: passed")
		assertContains(t, output, "report wording verification: passed")
		assertContains(t, output, "methodology fingerprint:")
	})
}

func TestBenchmarkVerificationWordingGuards(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "benchmark", "export", "--suite", "go-benchmark-refresh-v1", "--attempts", "2"}, &stdout); err != nil {
			t.Fatalf("Execute(eval benchmark export) error = %v", err)
		}

		stdout.Reset()
		if err := NewRootCommand().Execute([]string{"eval", "benchmark", "verify", "--suite", "go-benchmark-refresh-v1", "--attempts", "2"}, &stdout); err != nil {
			t.Fatalf("Execute(eval benchmark verify) error = %v", err)
		}
		output := stdout.String()
		assertContains(t, output, "report wording verification: passed")
		for _, banned := range []string{"provider billing", "statistically significant", "universal savings"} {
			if strings.Contains(output, banned) {
				t.Fatalf("verification output should not contain %q:\n%s", banned, output)
			}
		}
	})
}

func TestEvalMCPArtifactsPersist(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "--scenario", "mcp-go-worktree-v1"}, &stdout); err != nil {
			t.Fatalf("Execute(eval mcp-go-worktree-v1) error = %v", err)
		}
		assertContains(t, stdout.String(), "status: passed")
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

	run, steps, artifacts, err := store.LoadEvalRun(context.Background(), 1)
	if err != nil {
		t.Fatalf("LoadEvalRun(1) error = %v", err)
	}
	if run.ScenarioID != "mcp-go-worktree-v1" || !run.Passed {
		t.Fatalf("run = %+v", run)
	}
	if run.ArtifactRoot != layout.EvalRunDir(1) {
		t.Fatalf("artifact root = %q, want %q", run.ArtifactRoot, layout.EvalRunDir(1))
	}
	if len(steps) != 3 {
		t.Fatalf("len(steps) = %d, want 3", len(steps))
	}
	if steps[2].StderrPath == "" {
		t.Fatalf("missing MCP stderr path: %+v", steps[2])
	}
	if _, err := os.Stat(steps[2].StderrPath); err != nil {
		t.Fatalf("Stat(stderr path) error = %v", err)
	}

	transcript, ok := findEvalArtifactRecord(artifacts, "mcp-transcript")
	if !ok {
		t.Fatalf("missing transcript artifact: %+v", artifacts)
	}
	if transcript.StoredPath != filepath.Join(layout.EvalRunDir(1), "artifacts", "mcp-worktree-transcript.json") {
		t.Fatalf("transcript stored path = %q", transcript.StoredPath)
	}
	if _, err := os.Stat(transcript.StoredPath); err != nil {
		t.Fatalf("Stat(transcript path) error = %v", err)
	}

	healthArtifact, ok := findEvalArtifactRecord(artifacts, "health-response")
	if !ok {
		t.Fatalf("missing health artifact: %+v", artifacts)
	}
	content, err := os.ReadFile(healthArtifact.StoredPath)
	if err != nil {
		t.Fatalf("ReadFile(health artifact) error = %v", err)
	}
	assertContains(t, string(content), "\"Initialized\": true")
}

func TestEvalStaleAndDegradedScenarios(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		for _, scenarioID := range []string{"cli-go-stale-v1", "mcp-go-degraded-v1"} {
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

	staleRun, _, staleArtifacts, err := store.LoadEvalRun(context.Background(), 1)
	if err != nil {
		t.Fatalf("LoadEvalRun(1) error = %v", err)
	}
	if staleRun.ScenarioID != "cli-go-stale-v1" || !staleRun.Passed {
		t.Fatalf("stale run = %+v", staleRun)
	}
	doctorArtifact, ok := findEvalArtifactRecord(staleArtifacts, "doctor-stdout")
	if !ok {
		t.Fatalf("missing doctor artifact: %+v", staleArtifacts)
	}
	doctorContent, err := os.ReadFile(doctorArtifact.StoredPath)
	if err != nil {
		t.Fatalf("ReadFile(doctor artifact) error = %v", err)
	}
	assertContains(t, string(doctorContent), "freshness: stale")
	assertContains(t, string(doctorContent), "summary: watch heartbeat is stale")

	degradedRun, _, degradedArtifacts, err := store.LoadEvalRun(context.Background(), 2)
	if err != nil {
		t.Fatalf("LoadEvalRun(2) error = %v", err)
	}
	if degradedRun.ScenarioID != "mcp-go-degraded-v1" || !degradedRun.Passed {
		t.Fatalf("degraded run = %+v", degradedRun)
	}
	refreshError, ok := findEvalArtifactRecord(degradedArtifacts, "refresh-error")
	if !ok {
		t.Fatalf("missing refresh-error artifact: %+v", degradedArtifacts)
	}
	refreshErrorContent, err := os.ReadFile(refreshError.StoredPath)
	if err != nil {
		t.Fatalf("ReadFile(refresh-error) error = %v", err)
	}
	assertContains(t, string(refreshErrorContent), "forced eval failure")

	healthArtifact, ok := findEvalArtifactRecord(degradedArtifacts, "health-response")
	if !ok {
		t.Fatalf("missing health artifact: %+v", degradedArtifacts)
	}
	healthContent, err := os.ReadFile(healthArtifact.StoredPath)
	if err != nil {
		t.Fatalf("ReadFile(health artifact) error = %v", err)
	}
	assertContains(t, string(healthContent), "\"freshness\": \"partially_degraded\"")

	repositoryMapArtifact, ok := findEvalArtifactRecord(degradedArtifacts, "repository-map")
	if !ok {
		t.Fatalf("missing repository-map artifact: %+v", degradedArtifacts)
	}
	repositoryMapContent, err := os.ReadFile(repositoryMapArtifact.StoredPath)
	if err != nil {
		t.Fatalf("ReadFile(repository-map artifact) error = %v", err)
	}
	assertContains(t, string(repositoryMapContent), "\"Name\"")
}

func TestEvalRecoveryScenarios(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "--scenario", "mcp-go-recovery-v1"}, &stdout); err != nil {
			t.Fatalf("Execute(eval mcp-go-recovery-v1) error = %v", err)
		}
		assertContains(t, stdout.String(), "scenario id: mcp-go-recovery-v1")
		assertContains(t, stdout.String(), "status: passed")
		assertContains(t, stdout.String(), "step mcp-recovery: passed exit=0")
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

	run, _, artifacts, err := store.LoadEvalRun(context.Background(), 1)
	if err != nil {
		t.Fatalf("LoadEvalRun(1) error = %v", err)
	}
	if run.ScenarioID != "mcp-go-recovery-v1" || !run.Passed {
		t.Fatalf("run = %+v", run)
	}
	refreshArtifact, ok := findEvalArtifactRecord(artifacts, "refresh-recovered")
	if !ok {
		t.Fatalf("missing refresh-recovered artifact: %+v", artifacts)
	}
	refreshContent, err := os.ReadFile(refreshArtifact.StoredPath)
	if err != nil {
		t.Fatalf("ReadFile(refresh-recovered) error = %v", err)
	}
	assertContains(t, string(refreshContent), "\"generation\": 4")
	assertContains(t, string(refreshContent), "\"freshness\": \"fresh\"")

	repositoryMapArtifact, ok := findEvalArtifactRecord(artifacts, "repository-map-recovered")
	if !ok {
		t.Fatalf("missing repository-map-recovered artifact: %+v", artifacts)
	}
	repositoryMapContent, err := os.ReadFile(repositoryMapArtifact.StoredPath)
	if err != nil {
		t.Fatalf("ReadFile(repository-map-recovered) error = %v", err)
	}
	assertContains(t, string(repositoryMapContent), "\"RecoveredName\"")

	healthArtifact, ok := findEvalArtifactRecord(artifacts, "health-recovered")
	if !ok {
		t.Fatalf("missing health-recovered artifact: %+v", artifacts)
	}
	healthContent, err := os.ReadFile(healthArtifact.StoredPath)
	if err != nil {
		t.Fatalf("ReadFile(health-recovered) error = %v", err)
	}
	assertContains(t, string(healthContent), "\"generation\": 4")
	assertContains(t, string(healthContent), "\"freshness\": \"fresh\"")
}

func TestEvalRequirementCoverageReport(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		for _, scenarioID := range []string{
			"mcp-go-basic-v1",
			"mcp-go-worktree-v1",
			"cli-go-stale-v1",
			"mcp-go-degraded-v1",
			"mcp-go-recovery-v1",
		} {
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

	report, err := app.NewEvalService().RequirementCoverageReport(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("RequirementCoverageReport() error = %v", err)
	}
	if report.EvalArtifactRoot != layout.EvalDir {
		t.Fatalf("EvalArtifactRoot = %q, want %q", report.EvalArtifactRoot, layout.EvalDir)
	}
	if got, want := len(report.Requirements), 2; got != want {
		t.Fatalf("len(Requirements) = %d, want %d", got, want)
	}
	for _, requirement := range report.Requirements {
		if !requirement.Covered {
			t.Fatalf("requirement should be covered: %+v", requirement)
		}
		for _, evidence := range requirement.Evidence {
			if evidence.ArtifactRoot == "" {
				t.Fatalf("evidence missing artifact root: %+v", evidence)
			}
			if len(evidence.ArtifactPaths) == 0 {
				t.Fatalf("evidence missing artifact paths: %+v", evidence)
			}
			assertContains(t, evidence.RerunCommand, "go run ./cmd/optimusctx eval --scenario ")
		}
	}
	if report.Requirements[0].RequirementID != "EVAL-02" || report.Requirements[1].RequirementID != "EVAL-03" {
		t.Fatalf("requirements = %+v", report.Requirements)
	}
	assertContains(t, report.Requirements[0].Evidence[0].ScenarioID, "mcp-go-")
	assertContains(t, report.Requirements[1].Evidence[0].ArtifactRoot, filepath.Join(repoRoot, ".optimusctx", "eval", "run-"))
}

func TestEvalStateTransitionsPersistEvidence(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		for _, scenarioID := range []string{"cli-go-stale-v1", "mcp-go-degraded-v1", "mcp-go-recovery-v1"} {
			var stdout bytes.Buffer
			if err := NewRootCommand().Execute([]string{"eval", "--scenario", scenarioID}, &stdout); err != nil {
				t.Fatalf("Execute(eval %s) error = %v", scenarioID, err)
			}
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

	checkArtifact := func(runID int64, scenarioID string, artifactID string, fragments ...string) {
		run, _, artifacts, err := store.LoadEvalRun(context.Background(), runID)
		if err != nil {
			t.Fatalf("LoadEvalRun(%d) error = %v", runID, err)
		}
		if run.ScenarioID != scenarioID || run.ArtifactRoot != layout.EvalRunDir(runID) {
			t.Fatalf("run %d = %+v", runID, run)
		}
		record, ok := findEvalArtifactRecord(artifacts, artifactID)
		if !ok {
			t.Fatalf("run %d missing artifact %q: %+v", runID, artifactID, artifacts)
		}
		if _, err := os.Stat(record.StoredPath); err != nil {
			t.Fatalf("Stat(%s) error = %v", record.StoredPath, err)
		}
		content, err := os.ReadFile(record.StoredPath)
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", record.StoredPath, err)
		}
		for _, fragment := range fragments {
			assertContains(t, string(content), fragment)
		}
	}

	checkArtifact(1, "cli-go-stale-v1", "doctor-stdout", "freshness: stale", "summary: watch heartbeat is stale")
	checkArtifact(2, "mcp-go-degraded-v1", "refresh-error", "forced eval failure")
	checkArtifact(2, "mcp-go-degraded-v1", "health-response", "\"freshness\": \"partially_degraded\"")
	checkArtifact(3, "mcp-go-recovery-v1", "refresh-recovered", "\"generation\": 4", "\"freshness\": \"fresh\"")
	checkArtifact(3, "mcp-go-recovery-v1", "repository-map-recovered", "\"RecoveredName\"")
}

func TestEvalRecoveryAdvancesGeneration(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "--scenario", "mcp-go-recovery-v1"}, &stdout); err != nil {
			t.Fatalf("Execute(eval mcp-go-recovery-v1) error = %v", err)
		}
		assertContains(t, stdout.String(), "step mcp-recovery: passed exit=0")
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

	run, _, artifacts, err := store.LoadEvalRun(context.Background(), 1)
	if err != nil {
		t.Fatalf("LoadEvalRun(1) error = %v", err)
	}
	if run.ScenarioID != "mcp-go-recovery-v1" || !run.Passed {
		t.Fatalf("run = %+v", run)
	}

	refreshArtifact, ok := findEvalArtifactRecord(artifacts, "refresh-recovered")
	if !ok {
		t.Fatalf("missing refresh-recovered artifact: %+v", artifacts)
	}
	refreshContent, err := os.ReadFile(refreshArtifact.StoredPath)
	if err != nil {
		t.Fatalf("ReadFile(refresh-recovered) error = %v", err)
	}
	assertContains(t, string(refreshContent), "\"generation\": 4")
	assertContains(t, string(refreshContent), "\"freshness\": \"fresh\"")

	healthArtifact, ok := findEvalArtifactRecord(artifacts, "health-recovered")
	if !ok {
		t.Fatalf("missing health-recovered artifact: %+v", artifacts)
	}
	healthContent, err := os.ReadFile(healthArtifact.StoredPath)
	if err != nil {
		t.Fatalf("ReadFile(health-recovered) error = %v", err)
	}
	assertContains(t, string(healthContent), "\"generation\": 4")
	assertContains(t, string(healthContent), "\"freshness\": \"fresh\"")
}

func TestEvalMCPScenariosRerun(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	scenarios := []string{"mcp-go-basic-v1", "mcp-go-worktree-v1"}
	withWorkingDirectory(t, repoRoot, func() {
		for _, scenarioID := range scenarios {
			for run := 0; run < 2; run++ {
				var stdout bytes.Buffer
				if err := NewRootCommand().Execute([]string{"eval", "--scenario", scenarioID}, &stdout); err != nil {
					t.Fatalf("Execute(eval %s) #%d error = %v", scenarioID, run+1, err)
				}
				assertContains(t, stdout.String(), "scenario id: "+scenarioID)
				assertContains(t, stdout.String(), "status: passed")
				assertContains(t, stdout.String(), "step mcp-serve: passed exit=0")
			}
		}
	})

	for _, fixtureID := range []string{"go-basic", "go-worktree"} {
		fixtureRoot := filepath.Join(repoRoot, "testdata", "eval", "fixtures", fixtureID, "v1", "repository")
		if _, err := os.Stat(filepath.Join(fixtureRoot, ".optimusctx")); !os.IsNotExist(err) {
			t.Fatalf("fixture source %q should not be mutated by reruns, err=%v", fixtureID, err)
		}
	}

	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}
	store, err := sqlite.OpenOrCreateStore(context.Background(), layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("OpenOrCreateStore() error = %v", err)
	}
	defer store.Close()

	for runID := int64(1); runID <= 4; runID++ {
		run, _, artifacts, err := store.LoadEvalRun(context.Background(), runID)
		if err != nil {
			t.Fatalf("LoadEvalRun(%d) error = %v", runID, err)
		}
		if run.ArtifactRoot != layout.EvalRunDir(runID) {
			t.Fatalf("run %d artifact root = %q, want %q", runID, run.ArtifactRoot, layout.EvalRunDir(runID))
		}
		if run.ScenarioID != "mcp-go-worktree-v1" {
			continue
		}
		transcript, ok := findEvalArtifactRecord(artifacts, "mcp-transcript")
		if !ok {
			t.Fatalf("run %d missing transcript artifact: %+v", runID, artifacts)
		}
		wantPath := filepath.Join(layout.EvalRunDir(runID), "artifacts", "mcp-worktree-transcript.json")
		if transcript.StoredPath != wantPath {
			t.Fatalf("run %d transcript path = %q, want %q", runID, transcript.StoredPath, wantPath)
		}
	}
}

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
	packOne, ok := findEvalArtifactRecord(artifactsOne, "pack-output")
	if !ok {
		t.Fatalf("pack artifact missing from run one: %+v", artifactsOne)
	}
	packTwo, ok := findEvalArtifactRecord(artifactsTwo, "pack-output")
	if !ok {
		t.Fatalf("pack artifact missing from run two: %+v", artifactsTwo)
	}
	if packOne.StoredPath != packOnePath {
		t.Fatalf("artifact one stored path = %q", packOne.StoredPath)
	}
	if packTwo.StoredPath != packTwoPath {
		t.Fatalf("artifact two stored path = %q", packTwo.StoredPath)
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
	if packOne.SizeBytes == 0 || packTwo.SizeBytes == 0 {
		t.Fatalf("artifact metadata sizes = %d and %d, want non-zero", packOne.SizeBytes, packTwo.SizeBytes)
	}
}

func seedCommittedEvalFixtures(t *testing.T, repoRoot string) {
	t.Helper()

	sourceRoot := filepath.Join("..", "..", "testdata", "eval")
	copyCLITree(t, sourceRoot, filepath.Join(repoRoot, "testdata", "eval"))
}

func TestBenchmarkArtifactAttribution(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		service := newCLIBenchmarkService(t)
		if _, err := service.RunRepeated(context.Background(), app.BenchmarkRepeatedRunRequest{
			StartPath:    repoRoot,
			SuiteID:      "go-benchmark-discovery-v1",
			SuitesDir:    filepath.Join(repoRoot, "testdata", "eval", "benchmarks"),
			FixturesRoot: filepath.Join(repoRoot, "testdata", "eval", "fixtures"),
			Attempts:     1,
		}); err != nil {
			t.Fatalf("RunRepeated() error = %v", err)
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

	repositoryID, err := store.LookupRepositoryID(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("LookupRepositoryID() error = %v", err)
	}
	runs, err := store.ListBenchmarkRuns(context.Background(), repositoryID, "go-benchmark-discovery-v1", "v1")
	if err != nil {
		t.Fatalf("ListBenchmarkRuns() error = %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("len(runs) = %d, want 2", len(runs))
	}

	var found bool
	for _, run := range runs {
		if run.Run.ArmKind != repository.BenchmarkArmKindOptimusCtx {
			continue
		}
		for _, sample := range run.Samples {
			if sample.Sample.Lane != repository.BenchmarkLaneContextAssembly {
				continue
			}
			var metadata struct {
				Attribution []repository.BenchmarkArtifactConsumption `json:"attribution"`
			}
			if err := json.Unmarshal([]byte(sample.Sample.MetadataJSON), &metadata); err != nil {
				t.Fatalf("sample metadata json: %v", err)
			}
			if len(metadata.Attribution) == 0 {
				t.Fatalf("metadata attribution missing: %s", sample.Sample.MetadataJSON)
			}
			if metadata.Attribution[0].StepID != "opti-targeted-context" {
				t.Fatalf("step id = %q", metadata.Attribution[0].StepID)
			}
			if metadata.Attribution[0].ReportLabel != repository.BenchmarkReportArtifactLabelL2Context {
				t.Fatalf("report label = %q", metadata.Attribution[0].ReportLabel)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("did not find persisted optimusctx context-assembly attribution")
	}
}

func TestBenchmarkEvidenceBundleGeneration(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "benchmark", "export", "--suite", "go-benchmark-refresh-v1", "--attempts", "2"}, &stdout); err != nil {
			t.Fatalf("Execute(eval benchmark export) error = %v", err)
		}

		var bundle repository.BenchmarkEvidenceBundle
		if err := json.Unmarshal(stdout.Bytes(), &bundle); err != nil {
			t.Fatalf("Unmarshal(bundle) error = %v", err)
		}
		if bundle.SchemaVersion != repository.BenchmarkEvidenceBundleSchemaV1 {
			t.Fatalf("SchemaVersion = %q", bundle.SchemaVersion)
		}
		if len(bundle.Attempts) != 2 {
			t.Fatalf("len(bundle.Attempts) = %d, want 2", len(bundle.Attempts))
		}
		if len(bundle.Comparison) != 2 {
			t.Fatalf("len(bundle.Comparison) = %d, want 2", len(bundle.Comparison))
		}
	})
}

func TestBenchmarkExportContainsMethodologyIdentity(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "benchmark", "export", "--suite", "go-benchmark-refresh-v1", "--attempts", "2"}, &stdout); err != nil {
			t.Fatalf("Execute(eval benchmark export) error = %v", err)
		}

		var bundle repository.BenchmarkEvidenceBundle
		if err := json.Unmarshal(stdout.Bytes(), &bundle); err != nil {
			t.Fatalf("Unmarshal(bundle) error = %v", err)
		}
		if bundle.SuiteID != "go-benchmark-refresh-v1" {
			t.Fatalf("SuiteID = %q", bundle.SuiteID)
		}
		if bundle.FixtureID != "go-worktree" {
			t.Fatalf("FixtureID = %q", bundle.FixtureID)
		}
		if bundle.MethodologyFingerprint == "" {
			t.Fatal("MethodologyFingerprint should not be empty")
		}
		if bundle.TokenEstimateContract.Policy.Name != repository.BenchmarkTokenEstimatorPolicyName {
			t.Fatalf("Policy.Name = %q", bundle.TokenEstimateContract.Policy.Name)
		}
		if bundle.TokenEstimateContract.BillingDisambiguator != repository.BenchmarkTokenEstimateBillingDisambiguator {
			t.Fatalf("BillingDisambiguator = %q", bundle.TokenEstimateContract.BillingDisambiguator)
		}
		foundPackLabel := false
		for _, attempt := range bundle.Attempts {
			for _, arm := range attempt.Arms {
				for _, lane := range arm.Lanes {
					for _, attribution := range lane.Attribution {
						if attribution.ReportLabel == repository.BenchmarkReportArtifactLabelPackExport {
							foundPackLabel = true
						}
					}
				}
			}
		}
		if !foundPackLabel {
			t.Fatalf("bundle missing BNCH-02 report labels: %+v", bundle.Attempts)
		}
	})
}

func TestBenchmarkExportCLIPath(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		outputPath := filepath.Join("artifacts", "benchmark-evidence.json")
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "benchmark", "export", "--suite", "go-benchmark-refresh-v1", "--attempts", "2", "--output", outputPath}, &stdout); err != nil {
			t.Fatalf("Execute(eval benchmark export --output) error = %v", err)
		}
		assertContains(t, stdout.String(), "benchmark evidence written:")

		data, err := os.ReadFile(filepath.Join(repoRoot, outputPath))
		if err != nil {
			t.Fatalf("ReadFile(outputPath) error = %v", err)
		}
		var bundle repository.BenchmarkEvidenceBundle
		if err := json.Unmarshal(data, &bundle); err != nil {
			t.Fatalf("Unmarshal(output bundle) error = %v", err)
		}
		if !strings.Contains(bundle.RerunCommand, "optimusctx eval benchmark export --suite go-benchmark-refresh-v1 --attempts 2") {
			t.Fatalf("RerunCommand = %q", bundle.RerunCommand)
		}
	})
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

func findEvalArtifactRecord(records []sqlite.EvalArtifactRecord, artifactID string) (sqlite.EvalArtifactRecord, bool) {
	for _, record := range records {
		if record.ArtifactID == artifactID {
			return record, true
		}
	}
	return sqlite.EvalArtifactRecord{}, false
}

func newCLIBenchmarkService(t *testing.T) app.BenchmarkService {
	t.Helper()

	service := app.NewBenchmarkService()
	service.Runner = app.NewBenchmarkRunner()
	service.Runner.RunCommand = func(ctx context.Context, invocation app.BenchmarkCommandInvocation) (app.BenchmarkCommandExecutionResult, error) {
		execution, err := executeEvalCLICommand(ctx, app.EvalCommandInvocation{
			Args:       invocation.Args,
			WorkingDir: invocation.WorkingDir,
		})
		return app.BenchmarkCommandExecutionResult{
			Stdout:   execution.Stdout,
			Stderr:   execution.Stderr,
			ExitCode: execution.ExitCode,
		}, err
	}
	service.Runner.RunTool = func(ctx context.Context, invocation app.BenchmarkToolInvocation) (app.BenchmarkToolExecutionResult, error) {
		session := repository.EvalMCPSession{
			Requests: []repository.EvalMCPRequest{
				{ID: 1, Method: "initialize", Params: mcp.InitializeParams{
					ClientInfo:      mcp.ClientInfo{Name: "benchmark-test", Version: "1.0.0"},
					ProtocolVersion: "2024-11-05",
				}},
				{Method: "notifications/initialized", Notification: true},
				{ID: 2, Method: "tools/call", Params: mcp.CallToolParams{
					Name:      invocation.Name,
					Arguments: invocation.Arguments,
				}},
			},
		}
		execution, err := executeEvalCLIMCPSession(ctx, app.EvalMCPSessionInvocation{
			WorkingDir: invocation.WorkingDir,
			Session:    session,
		})
		if err != nil {
			return app.BenchmarkToolExecutionResult{}, err
		}
		payload, err := decodeBenchmarkToolPayload(execution.Responses[len(execution.Responses)-1].Response)
		if err != nil {
			return app.BenchmarkToolExecutionResult{}, err
		}
		return app.BenchmarkToolExecutionResult{Payload: payload}, nil
	}
	return service
}
