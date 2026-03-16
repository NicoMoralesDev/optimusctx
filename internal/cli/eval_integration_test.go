package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/buildinfo"
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

func TestDistributionSmokeFlow(t *testing.T) {
	previousVersion := buildinfo.Version
	previousCommit := buildinfo.Commit
	previousBuildDate := buildinfo.BuildDate
	t.Cleanup(func() {
		buildinfo.Version = previousVersion
		buildinfo.Commit = previousCommit
		buildinfo.BuildDate = previousBuildDate
	})

	buildinfo.Version = "v1.1.0"
	buildinfo.Commit = "abc1234"
	buildinfo.BuildDate = "2026-03-16T18:00:00Z"

	repoRoot := initCLIRepo(t)
	writeCLIFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc main() {}\n")

	versionOutput := runDistributionCLICommand(t, repoRoot, "version")
	if got, want := versionOutput, "optimusctx version=v1.1.0 commit=abc1234 build_date=2026-03-16T18:00:00Z\n"; got != want {
		t.Fatalf("version output = %q, want %q", got, want)
	}

	initOutput := runDistributionCLICommand(t, repoRoot, "init")
	assertContains(t, initOutput, "repository root: "+repoRoot)
	assertContains(t, initOutput, "freshness: fresh")

	doctorOutput := runDistributionCLICommand(t, repoRoot, "doctor")
	assertContains(t, doctorOutput, "overall status: healthy")
	assertContains(t, doctorOutput, "runtime version: v1.1.0")
	assertContains(t, doctorOutput, "freshness: fresh")
	assertContains(t, doctorOutput, "snippet available: yes")

	snippetOutput := runDistributionCLICommand(t, repoRoot, "snippet")
	assertContains(t, snippetOutput, "# OptimusCtx manual integration snippet")
	assertContains(t, snippetOutput, "optimusctx install --client claude-desktop")

	configPath := filepath.Join(repoRoot, "claude_desktop_config.json")
	installOutput := runDistributionCLICommand(t, repoRoot, "install", "--client", "claude-desktop", "--config", configPath)
	assertContains(t, installOutput, "mode: preview")
	assertContains(t, installOutput, "status: preview only")
	assertContains(t, installOutput, "config path: "+configPath)
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("install preview should not write config file: %v", err)
	}

	snippetDocument := extractConfigDocument(t, snippetOutput)
	installDocument := extractConfigDocument(t, installOutput)
	if got, want := installDocument.MCPServers[repository.DefaultMCPServerName], snippetDocument.MCPServers[repository.DefaultMCPServerName]; got.Command != want.Command || strings.Join(got.Args, " ") != strings.Join(want.Args, " ") {
		t.Fatalf("install preview = %+v, snippet = %+v", got, want)
	}
}

func TestReleaseVerificationCommands(t *testing.T) {
	readme := readCLIRepoFile(t, "README.md")
	guide := readCLIRepoFile(t, "docs/install-and-verify.md")

	assertContains(t, readme, "docs/install-and-verify.md")
	assertContains(t, readme, "MCP registration is explicit and opt-in")

	requiredGuideFragments := []string{
		"optimusctx version",
		"optimusctx doctor",
		"optimusctx snippet",
		"optimusctx install --client claude-desktop",
		"status: preview only",
		"brew install niccrow/tap/optimusctx",
		"scoop install niccrow/optimusctx",
	}
	for _, fragment := range requiredGuideFragments {
		assertContains(t, guide, fragment)
	}

	for _, banned := range []string{
		"go run ./cmd/optimusctx",
		"go install ./cmd/optimusctx",
	} {
		if strings.Contains(guide, banned) {
			t.Fatalf("docs/install-and-verify.md should not contain %q", banned)
		}
	}

	versionIndex := strings.Index(guide, "optimusctx version")
	doctorIndex := strings.Index(guide, "optimusctx doctor")
	snippetIndex := strings.Index(guide, "optimusctx snippet")
	installIndex := strings.Index(guide, "optimusctx install --client claude-desktop")
	if !(versionIndex >= 0 && doctorIndex > versionIndex && snippetIndex > doctorIndex && installIndex > snippetIndex) {
		t.Fatalf("guide command order is wrong: version=%d doctor=%d snippet=%d install=%d", versionIndex, doctorIndex, snippetIndex, installIndex)
	}
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

func runDistributionCLICommand(t *testing.T, workingDir string, args ...string) string {
	t.Helper()

	execution, err := executeEvalCLICommand(context.Background(), app.EvalCommandInvocation{
		Args:       args,
		WorkingDir: workingDir,
	})
	if err != nil {
		t.Fatalf("executeEvalCLICommand(%v) error = %v", args, err)
	}
	if execution.ExitCode != 0 {
		t.Fatalf("executeEvalCLICommand(%v) exit=%d stderr=%q stdout=%q", args, execution.ExitCode, execution.Stderr, execution.Stdout)
	}
	return execution.Stdout
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
		assertContains(t, refreshOutput, "counted agent-input attribution")
		assertContains(t, refreshOutput, "L2 Context")
		assertContains(t, refreshOutput, "Operational")

		discoveryOutput := render("go-benchmark-discovery-v1")
		assertContains(t, discoveryOutput, "Repository Map")
		assertContains(t, discoveryOutput, "Exact Lookup")
	})
}

func TestBenchmarkReportWordingGuards(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)
	suitePath := seedV2BenchmarkEvidence(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "benchmark", "report", "--suite-file", suitePath}, &stdout); err != nil {
			t.Fatalf("Execute(eval benchmark report) error = %v", err)
		}
		output := stdout.String()
		assertContains(t, output, "estimated tokens use bytes_div_4_ceiling")
		assertContains(t, output, "not provider-billed token invoices")
		assertContains(t, output, "fixes benchmark truthfulness")
		for _, banned := range []string{"provider-billed token truth", "statistically significant", "universal savings"} {
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
	suitePath := seedV2BenchmarkEvidence(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "benchmark", "export", "--suite-file", suitePath}, &stdout); err != nil {
			t.Fatalf("Execute(eval benchmark export) error = %v", err)
		}

		var bundle repository.BenchmarkEvidenceBundle
		if err := json.Unmarshal(stdout.Bytes(), &bundle); err != nil {
			t.Fatalf("Unmarshal(bundle) error = %v", err)
		}
		if bundle.SuiteID != "go-benchmark-refresh-v2" {
			t.Fatalf("SuiteID = %q", bundle.SuiteID)
		}
		if bundle.FixtureID != "go-worktree" {
			t.Fatalf("FixtureID = %q", bundle.FixtureID)
		}
		if bundle.SchemaVersion != repository.BenchmarkEvidenceBundleSchemaV2 {
			t.Fatalf("SchemaVersion = %q", bundle.SchemaVersion)
		}
		if bundle.Methodology.SuiteSchemaVersion != repository.BenchmarkSuiteSchemaV2 {
			t.Fatalf("Methodology.SuiteSchemaVersion = %q", bundle.Methodology.SuiteSchemaVersion)
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
		foundOperationalLabel := false
		foundL2ContextLabel := false
		for _, attempt := range bundle.Attempts {
			for _, arm := range attempt.Arms {
				for _, lane := range arm.Lanes {
					for _, attribution := range lane.Attribution {
						switch attribution.ReportLabel {
						case repository.BenchmarkReportArtifactLabelOperational:
							foundOperationalLabel = true
						case repository.BenchmarkReportArtifactLabelL2Context:
							foundL2ContextLabel = true
						case repository.BenchmarkReportArtifactLabelPackExport:
							t.Fatalf("bundle should not contain pack export labels after fairness fix: %+v", bundle.Attempts)
						}
					}
				}
			}
		}
		if !foundOperationalLabel || !foundL2ContextLabel {
			t.Fatalf("bundle missing expected report labels: %+v", bundle.Attempts)
		}
	})
}

func TestBenchmarkExportCLIPath(t *testing.T) {
	repoRoot := initCLIRepo(t)
	seedCommittedEvalFixtures(t, repoRoot)
	suitePath := seedV2BenchmarkEvidence(t, repoRoot)

	withWorkingDirectory(t, repoRoot, func() {
		outputPath := filepath.Join("artifacts", "benchmark-evidence.json")
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"eval", "benchmark", "export", "--suite-file", suitePath, "--output", outputPath}, &stdout); err != nil {
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
		if !strings.Contains(bundle.RerunCommand, "optimusctx eval benchmark export --suite-file") {
			t.Fatalf("RerunCommand = %q", bundle.RerunCommand)
		}
	})
}

func seedV2BenchmarkEvidence(t *testing.T, repoRoot string) string {
	t.Helper()

	fixturePath := filepath.Join(repoRoot, "testdata", "eval", "fixtures", "go-worktree", "v2", "repository")
	if err := os.MkdirAll(fixturePath, 0o755); err != nil {
		t.Fatalf("MkdirAll(fixturePath) error = %v", err)
	}
	suitePath := filepath.Join(repoRoot, "testdata", "eval", "benchmarks", "go-benchmark-refresh-v2.json")
	data, err := json.MarshalIndent(cliBenchmarkSuiteV2(), "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent(cliBenchmarkSuiteV2) error = %v", err)
	}
	if err := os.WriteFile(suitePath, data, 0o644); err != nil {
		t.Fatalf("WriteFile(suitePath) error = %v", err)
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

	record, err := store.UpsertRepository(context.Background(), repository.RepositoryRoot{
		RootPath:      repoRoot,
		DetectionMode: repository.DetectionModeGit,
		Fingerprint: repository.RepositoryFingerprint{
			RootPath:     repoRoot,
			GitCommonDir: filepath.Join(repoRoot, ".git"),
		},
	}, time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("UpsertRepository() error = %v", err)
	}
	for attempt := 1; attempt <= 2; attempt++ {
		for _, persisted := range sqlite.BenchmarkPersistedArmsFromResult(record.ID, attempt, cliBenchmarkRunResult(repoRoot)) {
			if _, _, err := store.SaveBenchmarkRun(context.Background(), persisted.Run, persisted.Samples); err != nil {
				t.Fatalf("SaveBenchmarkRun(attempt=%d) error = %v", attempt, err)
			}
		}
	}
	return suitePath
}

func cliBenchmarkSuiteV2() repository.BenchmarkSuiteDefinition {
	return repository.BenchmarkSuiteDefinition{
		SchemaVersion: repository.BenchmarkSuiteSchemaV2,
		ID:            "go-benchmark-refresh-v2",
		Version:       "v2",
		Name:          "Go benchmark refresh and task completion",
		Boundary:      repository.DefaultBenchmarkBoundaryContract(),
		Fixture: repository.EvalFixtureRef{
			ID:          "go-worktree",
			Version:     "v2",
			Path:        "go-worktree/v2/repository",
			Materialize: repository.EvalFixtureModeCopyTree,
		},
		Task: repository.BenchmarkTaskDefinition{
			ID:         "docs-pack",
			Prompt:     "Refresh after mutation and export bounded context.",
			TargetPath: "docs/notes.txt",
		},
		CountedInputs: []repository.BenchmarkCountedInputDefinition{
			{
				ID:         "baseline-readiness-paths",
				ArmKind:    repository.BenchmarkArmKindBaseline,
				Lane:       repository.BenchmarkLaneRefreshReady,
				StepID:     "baseline-readiness",
				Name:       "Baseline readiness path hints",
				Kind:       repository.BenchmarkCountedInputKindPathList,
				SourceKind: repository.BenchmarkTokenEstimateSourcePathEstimate,
				Path:       "docs/notes.txt",
			},
			{
				ID:           "treatment-health-summary",
				ArmKind:      repository.BenchmarkArmKindOptimusCtx,
				Lane:         repository.BenchmarkLaneRefreshReady,
				StepID:       "health",
				Name:         "Projected health summary",
				Kind:         repository.BenchmarkCountedInputKindJSONFieldProjection,
				SourceKind:   repository.BenchmarkTokenEstimateSourceDirectPayload,
				ArtifactType: repository.BenchmarkArtifactTypeHealth,
				ReportLabel:  repository.BenchmarkReportArtifactLabelOperational,
				JSONPath:     "refresh.freshness",
			},
			{
				ID:         "baseline-task-slice",
				ArmKind:    repository.BenchmarkArmKindBaseline,
				Lane:       repository.BenchmarkLaneTaskCompletion,
				StepID:     "baseline-context",
				Name:       "Baseline updated notes slice",
				Kind:       repository.BenchmarkCountedInputKindFileSlice,
				SourceKind: repository.BenchmarkTokenEstimateSourceBoundedFileContent,
				Path:       "docs/notes.txt",
				StartLine:  1,
				EndLine:    20,
			},
			{
				ID:           "treatment-updated-context",
				ArmKind:      repository.BenchmarkArmKindOptimusCtx,
				Lane:         repository.BenchmarkLaneTaskCompletion,
				StepID:       "context",
				Name:         "Treatment updated notes context",
				Kind:         repository.BenchmarkCountedInputKindTextOutput,
				SourceKind:   repository.BenchmarkTokenEstimateSourceDirectPayload,
				ArtifactType: repository.BenchmarkArtifactTypeL2Context,
				ReportLabel:  repository.BenchmarkReportArtifactLabelL2Context,
				Path:         "artifacts/updated-notes.txt",
			},
		},
		Lanes: []repository.BenchmarkLaneDefinition{
			{
				Name: repository.BenchmarkLaneRefreshReady,
				Assertions: []repository.BenchmarkAssertion{{
					File:     "docs/notes.txt",
					Kind:     repository.EvalAssertionKindContains,
					Contains: "mutated benchmark note",
				}},
				FinalArtifact: &repository.BenchmarkFinalArtifactContract{
					ID:     "refresh-readiness",
					Name:   "Refresh readiness summary",
					Kind:   repository.BenchmarkFinalArtifactKindReadinessSummary,
					Path:   "artifacts/readiness.json",
					Format: repository.BenchmarkFinalArtifactFormatJSON,
					Normalization: repository.BenchmarkFinalArtifactNormalization{
						Mode:      repository.BenchmarkFinalArtifactNormalizationModeJSONFields,
						JSONPaths: []string{"freshness", "targetReady"},
					},
					Assertions: []repository.BenchmarkFinalArtifactAssertion{
						{Kind: repository.EvalAssertionKindJSONFieldPresent, Path: "freshness"},
					},
				},
				StopCondition: repository.BenchmarkStopCondition{
					Kind:   repository.BenchmarkStopConditionKindMarker,
					Marker: "refresh_ready",
				},
				Metrics: []repository.BenchmarkMetric{repository.BenchmarkMetricConsultedArtifacts},
			},
			{
				Name: repository.BenchmarkLaneTaskCompletion,
				Assertions: []repository.BenchmarkAssertion{{
					File:     "docs/notes.txt",
					Kind:     repository.EvalAssertionKindContains,
					Contains: "mutated benchmark note",
				}},
				FinalArtifact: &repository.BenchmarkFinalArtifactContract{
					ID:     "updated-notes-context",
					Name:   "Updated notes context",
					Kind:   repository.BenchmarkFinalArtifactKindTaskOutput,
					Path:   "artifacts/updated-notes.txt",
					Format: repository.BenchmarkFinalArtifactFormatText,
					Normalization: repository.BenchmarkFinalArtifactNormalization{
						Mode:           repository.BenchmarkFinalArtifactNormalizationModeTextTrimmed,
						TrimWhitespace: true,
					},
					Assertions: []repository.BenchmarkFinalArtifactAssertion{
						{Kind: repository.EvalAssertionKindContains, Contains: "mutated benchmark note"},
					},
				},
				StopCondition: repository.BenchmarkStopCondition{
					Kind:   repository.BenchmarkStopConditionKindMarker,
					Marker: "task_complete",
				},
				Metrics: []repository.BenchmarkMetric{repository.BenchmarkMetricFileReadActions},
			},
		},
		Arms: []repository.BenchmarkArmDefinition{
			{
				Kind: repository.BenchmarkArmKindBaseline,
				Name: "Baseline",
				Steps: []repository.BenchmarkStep{
					{
						ID:   "baseline-readiness",
						Name: "Search mutated note",
						Lane: repository.BenchmarkLaneRefreshReady,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:  repository.BenchmarkBaselineActionGitGrep,
							Query: "mutated benchmark note",
						},
					},
					{
						ID:   "baseline-ready",
						Name: "Mark refresh ready",
						Lane: repository.BenchmarkLaneRefreshReady,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:   repository.BenchmarkBaselineActionMarkLaneComplete,
							Marker: "refresh_ready",
						},
					},
					{
						ID:   "baseline-context",
						Name: "Read docs slice",
						Lane: repository.BenchmarkLaneTaskCompletion,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:      repository.BenchmarkBaselineActionReadFileSlice,
							Path:      "docs/notes.txt",
							StartLine: 1,
							EndLine:   20,
						},
					},
					{
						ID:   "baseline-done",
						Name: "Mark task complete",
						Lane: repository.BenchmarkLaneTaskCompletion,
						Baseline: &repository.BenchmarkBaselineAction{
							Kind:   repository.BenchmarkBaselineActionMarkLaneComplete,
							Marker: "task_complete",
						},
					},
				},
			},
			{
				Kind: repository.BenchmarkArmKindOptimusCtx,
				Name: "OptimusCtx",
				Steps: []repository.BenchmarkStep{
					{
						ID:   "refresh",
						Name: "Refresh repository",
						Lane: repository.BenchmarkLaneRefreshReady,
						Treatment: &repository.BenchmarkTreatmentAction{
							Surface: repository.BenchmarkTreatmentSurfaceCLI,
							Command: repository.EvalCommandRefresh,
						},
					},
					{
						ID:   "health",
						Name: "Check repository health",
						Lane: repository.BenchmarkLaneRefreshReady,
						Treatment: &repository.BenchmarkTreatmentAction{
							Surface: repository.BenchmarkTreatmentSurfaceMCP,
							Tool:    "optimusctx.health",
						},
					},
					{
						ID:   "context",
						Name: "Fetch bounded updated notes context",
						Lane: repository.BenchmarkLaneTaskCompletion,
						Treatment: &repository.BenchmarkTreatmentAction{
							Surface: repository.BenchmarkTreatmentSurfaceMCP,
							Tool:    "optimusctx.targeted_context",
						},
					},
				},
			},
		},
	}
}

func cliBenchmarkRunResult(repositoryRoot string) repository.BenchmarkRunResult {
	return repository.BenchmarkRunResult{
		SchemaVersion: repository.BenchmarkSuiteSchemaV2,
		SuiteID:       "go-benchmark-refresh-v2",
		SuiteVersion:  "v2",
		FixtureID:     "go-worktree",
		FixturePath:   "go-worktree/v2/repository",
		WorkspacePath: repositoryRoot,
		Arms: []repository.BenchmarkArmRunResult{
			{
				Kind:       repository.BenchmarkArmKindBaseline,
				Name:       "Baseline",
				Workspace:  repositoryRoot,
				StartedAt:  time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, 3, 16, 20, 0, 2, 0, time.UTC),
				LaneResults: []repository.BenchmarkLaneRunResult{
					{
						Lane:          repository.BenchmarkLaneRefreshReady,
						StartMarker:   "refresh_after_change_started",
						SuccessMarker: "refresh_ready",
						StopMarker:    "refresh_ready",
						StartedAt:     time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC),
						FinishedAt:    time.Date(2026, 3, 16, 20, 0, 1, 0, time.UTC),
						Elapsed:       time.Second,
						Success:       true,
						EvidencePaths: []string{"docs/notes.txt"},
						FinalArtifact: &repository.BenchmarkLaneFinalArtifactVerification{
							ContractID: "refresh-readiness",
							Path:       "artifacts/readiness.json",
							Passed:     true,
						},
						Effort: repository.BenchmarkLaneEffort{
							ActionCount:        2,
							BytesRead:          128,
							ConsultedArtifacts: []string{"docs/notes.txt"},
						},
						Attribution: []repository.BenchmarkArtifactConsumption{{
							StepID:          "baseline-readiness",
							StepName:        "Search docs note",
							Lane:            repository.BenchmarkLaneRefreshReady,
							Boundary:        repository.BenchmarkEvidenceBoundaryAgentInput,
							SourceKind:      repository.BenchmarkTokenEstimateSourcePathEstimate,
							ArtifactPath:    "docs/notes.txt",
							EstimatedBytes:  48,
							EstimatedTokens: 12,
						}},
					},
					{
						Lane:          repository.BenchmarkLaneTaskCompletion,
						StartMarker:   "task_completion_started",
						SuccessMarker: "task_complete",
						StopMarker:    "task_complete",
						StartedAt:     time.Date(2026, 3, 16, 20, 0, 1, 0, time.UTC),
						FinishedAt:    time.Date(2026, 3, 16, 20, 0, 2, 0, time.UTC),
						Elapsed:       time.Second,
						Success:       true,
						EvidencePaths: []string{"docs/notes.txt"},
						FinalArtifact: &repository.BenchmarkLaneFinalArtifactVerification{
							ContractID: "updated-notes-context",
							Path:       "artifacts/updated-notes.txt",
							Passed:     true,
						},
						Effort: repository.BenchmarkLaneEffort{
							ActionCount:        1,
							FileReadActions:    1,
							BytesRead:          512,
							ConsultedArtifacts: []string{"docs/notes.txt"},
						},
						Attribution: []repository.BenchmarkArtifactConsumption{{
							StepID:          "baseline-context",
							StepName:        "Read docs slice",
							Lane:            repository.BenchmarkLaneTaskCompletion,
							Boundary:        repository.BenchmarkEvidenceBoundaryAgentInput,
							SourceKind:      repository.BenchmarkTokenEstimateSourceBoundedFileContent,
							ArtifactPath:    "docs/notes.txt",
							EstimatedBytes:  88,
							EstimatedTokens: 22,
						}},
					},
				},
			},
			{
				Kind:       repository.BenchmarkArmKindOptimusCtx,
				Name:       "OptimusCtx",
				Workspace:  repositoryRoot,
				StartedAt:  time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, 3, 16, 20, 0, 2, 0, time.UTC),
				LaneResults: []repository.BenchmarkLaneRunResult{
					{
						Lane:          repository.BenchmarkLaneRefreshReady,
						StartMarker:   "refresh_after_change_started",
						SuccessMarker: "refresh_ready",
						StopMarker:    "refresh_ready",
						StartedAt:     time.Date(2026, 3, 16, 20, 0, 0, 0, time.UTC),
						FinishedAt:    time.Date(2026, 3, 16, 20, 0, 1, 0, time.UTC),
						Elapsed:       time.Second,
						Success:       true,
						EvidencePaths: []string{"docs/notes.txt", "stdout"},
						FinalArtifact: &repository.BenchmarkLaneFinalArtifactVerification{
							ContractID: "refresh-readiness",
							Path:       "artifacts/readiness.json",
							Passed:     true,
						},
						Effort: repository.BenchmarkLaneEffort{
							ActionCount:           2,
							TargetedLookupActions: 2,
							ConsultedArtifacts:    []string{"docs/notes.txt", "stdout"},
						},
						Attribution: []repository.BenchmarkArtifactConsumption{
							{
								StepID:          "refresh",
								StepName:        "Refresh repository",
								Lane:            repository.BenchmarkLaneRefreshReady,
								Boundary:        repository.BenchmarkEvidenceBoundarySystemProvenance,
								Surface:         repository.BenchmarkTreatmentSurfaceCLI,
								Command:         repository.EvalCommandRefresh,
								ArtifactType:    repository.BenchmarkArtifactTypeRefresh,
								ReportLabel:     repository.BenchmarkReportArtifactLabelOperational,
								SourceKind:      repository.BenchmarkTokenEstimateSourceDirectPayload,
								ArtifactPath:    "stdout",
								EstimatedBytes:  32,
								EstimatedTokens: 8,
							},
							{
								StepID:          "health",
								StepName:        "Check repository health",
								Lane:            repository.BenchmarkLaneRefreshReady,
								Boundary:        repository.BenchmarkEvidenceBoundaryAgentInput,
								Surface:         repository.BenchmarkTreatmentSurfaceMCP,
								Tool:            "optimusctx.health",
								ArtifactType:    repository.BenchmarkArtifactTypeHealth,
								ReportLabel:     repository.BenchmarkReportArtifactLabelOperational,
								SourceKind:      repository.BenchmarkTokenEstimateSourceDirectPayload,
								ArtifactPath:    "artifacts/readiness.json",
								EstimatedBytes:  48,
								EstimatedTokens: 12,
							},
						},
					},
					{
						Lane:          repository.BenchmarkLaneTaskCompletion,
						StartMarker:   "task_completion_started",
						SuccessMarker: "task_complete",
						StopMarker:    "task_complete",
						StartedAt:     time.Date(2026, 3, 16, 20, 0, 1, 0, time.UTC),
						FinishedAt:    time.Date(2026, 3, 16, 20, 0, 2, 0, time.UTC),
						Elapsed:       time.Second,
						Success:       true,
						EvidencePaths: []string{"docs/notes.txt"},
						FinalArtifact: &repository.BenchmarkLaneFinalArtifactVerification{
							ContractID: "updated-notes-context",
							Path:       "artifacts/updated-notes.txt",
							Passed:     true,
						},
						Effort: repository.BenchmarkLaneEffort{
							ActionCount:           1,
							TargetedLookupActions: 1,
							BytesRead:             72,
							ConsultedArtifacts:    []string{"docs/notes.txt"},
						},
						Attribution: []repository.BenchmarkArtifactConsumption{{
							StepID:          "context",
							StepName:        "Fetch bounded updated notes context",
							Lane:            repository.BenchmarkLaneTaskCompletion,
							Boundary:        repository.BenchmarkEvidenceBoundaryAgentInput,
							Surface:         repository.BenchmarkTreatmentSurfaceMCP,
							Tool:            "optimusctx.targeted_context",
							ArtifactType:    repository.BenchmarkArtifactTypeL2Context,
							ReportLabel:     repository.BenchmarkReportArtifactLabelL2Context,
							SourceKind:      repository.BenchmarkTokenEstimateSourceDirectPayload,
							ArtifactPath:    "artifacts/updated-notes.txt",
							EstimatedBytes:  72,
							EstimatedTokens: 18,
						}},
					},
				},
			},
		},
	}
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
