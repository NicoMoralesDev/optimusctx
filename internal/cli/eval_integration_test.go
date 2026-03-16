package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/niccrow/optimusctx/internal/app"
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
