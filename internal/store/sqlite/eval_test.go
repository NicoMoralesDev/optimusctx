package sqlite

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/state"
)

func TestApplyMigrationsCreatesEvalTables(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	assertTablesExist(t, store.DB(), "eval_runs", "eval_steps", "eval_artifacts")
	assertIndexColumns(t, store.DB(), "eval_runs", []string{"repository_id", "started_at"})
	assertIndexColumns(t, store.DB(), "eval_runs", []string{"repository_id", "scenario_id", "started_at"})
	assertIndexColumns(t, store.DB(), "eval_steps", []string{"eval_run_id", "ordinal"})
	assertIndexColumns(t, store.DB(), "eval_artifacts", []string{"eval_run_id", "step_id"})

	run, err := store.SaveEvalRun(ctx, EvalRunRecord{
		RepositoryID:    repoID,
		ScenarioID:      "init-refresh-v1",
		ScenarioVersion: "v1",
		FixtureID:       "sample-go",
		FixtureVersion:  "2026-03-15",
		Status:          EvalRunStatusRunning,
		WorkspacePath:   layout.RepoRoot,
		ArtifactRoot:    layout.EvalRunDir(1),
		StartedAt:       time.Date(2026, 3, 15, 16, 0, 0, 0, time.UTC),
	}, nil, nil)
	if err != nil {
		t.Fatalf("SaveEvalRun() error = %v", err)
	}
	if run.ID == 0 {
		t.Fatal("eval run ID should be assigned")
	}
}

func TestEvalRunPersistence(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	startedAt := time.Date(2026, 3, 15, 17, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(3 * time.Minute)
	run, err := store.SaveEvalRun(ctx, EvalRunRecord{
		RepositoryID:    repoID,
		ScenarioID:      "doctor-pack-v1",
		ScenarioVersion: "v1",
		FixtureID:       "sample-go",
		FixtureVersion:  "2026-03-15",
		Status:          EvalRunStatusPassed,
		Passed:          true,
		WorkspacePath:   layout.RepoRoot,
		ArtifactRoot:    layout.EvalRunDir(42),
		StartedAt:       startedAt,
		CompletedAt:     finishedAt,
		MetadataJSON:    `{"runner":"cli"}`,
	}, []EvalStepRecord{
		{
			StepID:       "init",
			Ordinal:      0,
			Name:         "Initialize repository state",
			Surface:      "cli",
			Command:      "init",
			ArgsJSON:     `["init"]`,
			ExitCode:     0,
			Passed:       true,
			StartedAt:    startedAt,
			FinishedAt:   startedAt.Add(30 * time.Second),
			StdoutPath:   layout.EvalRunDir(42) + "/init/stdout.txt",
			MetadataJSON: `{"exitCode":0}`,
		},
		{
			StepID:       "refresh",
			Ordinal:      1,
			Name:         "Refresh repository state",
			Surface:      "cli",
			Command:      "refresh",
			ArgsJSON:     `["refresh"]`,
			ExitCode:     0,
			Passed:       true,
			StartedAt:    startedAt.Add(45 * time.Second),
			FinishedAt:   finishedAt,
			StdoutPath:   layout.EvalRunDir(42) + "/refresh/stdout.txt",
			StderrPath:   layout.EvalRunDir(42) + "/refresh/stderr.txt",
			MetadataJSON: `{"generation":2}`,
		},
	}, nil)
	if err != nil {
		t.Fatalf("SaveEvalRun() error = %v", err)
	}

	gotRun, gotSteps, gotArtifacts, err := store.LoadEvalRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("LoadEvalRun() error = %v", err)
	}

	if !reflect.DeepEqual(gotArtifacts, []EvalArtifactRecord(nil)) {
		t.Fatalf("artifacts = %+v, want nil", gotArtifacts)
	}

	run.CreatedAt = gotRun.CreatedAt
	run.UpdatedAt = gotRun.UpdatedAt
	if !reflect.DeepEqual(gotRun, run) {
		t.Fatalf("run mismatch\n got=%+v\nwant=%+v", gotRun, run)
	}

	wantSteps := []EvalStepRecord{
		{
			ID:           gotSteps[0].ID,
			EvalRunID:    run.ID,
			StepID:       "init",
			Ordinal:      0,
			Name:         "Initialize repository state",
			Surface:      "cli",
			Command:      "init",
			ArgsJSON:     `["init"]`,
			ExitCode:     0,
			Passed:       true,
			StartedAt:    startedAt,
			FinishedAt:   startedAt.Add(30 * time.Second),
			StdoutPath:   layout.EvalRunDir(42) + "/init/stdout.txt",
			StderrPath:   "",
			MetadataJSON: `{"exitCode":0}`,
		},
		{
			ID:           gotSteps[1].ID,
			EvalRunID:    run.ID,
			StepID:       "refresh",
			Ordinal:      1,
			Name:         "Refresh repository state",
			Surface:      "cli",
			Command:      "refresh",
			ArgsJSON:     `["refresh"]`,
			ExitCode:     0,
			Passed:       true,
			StartedAt:    startedAt.Add(45 * time.Second),
			FinishedAt:   finishedAt,
			StdoutPath:   layout.EvalRunDir(42) + "/refresh/stdout.txt",
			StderrPath:   layout.EvalRunDir(42) + "/refresh/stderr.txt",
			MetadataJSON: `{"generation":2}`,
		},
	}
	if !reflect.DeepEqual(gotSteps, wantSteps) {
		t.Fatalf("steps mismatch\n got=%+v\nwant=%+v", gotSteps, wantSteps)
	}
}

func TestEvalArtifactPersistence(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	run, err := store.SaveEvalRun(ctx, EvalRunRecord{
		RepositoryID:    repoID,
		ScenarioID:      "pack-export-v1",
		ScenarioVersion: "v1",
		FixtureID:       "sample-go",
		FixtureVersion:  "2026-03-15",
		Status:          EvalRunStatusFailed,
		WorkspacePath:   layout.RepoRoot,
		ArtifactRoot:    layout.EvalRunDir(7),
		StartedAt:       time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC),
		CompletedAt:     time.Date(2026, 3, 15, 18, 1, 0, 0, time.UTC),
	}, []EvalStepRecord{{
		StepID:     "pack-export",
		Ordinal:    0,
		Name:       "Export packed context",
		Surface:    "cli",
		Command:    "pack_export",
		ExitCode:   1,
		Passed:     false,
		StartedAt:  time.Date(2026, 3, 15, 18, 0, 10, 0, time.UTC),
		FinishedAt: time.Date(2026, 3, 15, 18, 1, 0, 0, time.UTC),
	}}, []EvalArtifactRecord{
		{
			StepID:       "pack-export",
			ArtifactID:   "stderr",
			Kind:         "stderr",
			StoredPath:   layout.EvalRunDir(7) + "/pack-export/stderr.txt",
			Required:     true,
			Present:      true,
			SizeBytes:    128,
			MetadataJSON: `{"lineCount":4}`,
		},
		{
			ArtifactID:   "pack-json",
			Kind:         "file",
			LogicalPath:  "artifacts/pack.json",
			StoredPath:   layout.EvalRunDir(7) + "/artifacts/pack.json",
			Required:     true,
			Present:      false,
			SizeBytes:    0,
			MetadataJSON: `{"missing":true}`,
		},
	})
	if err != nil {
		t.Fatalf("SaveEvalRun() error = %v", err)
	}

	_, _, artifacts, err := store.LoadEvalRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("LoadEvalRun() error = %v", err)
	}

	want := []EvalArtifactRecord{
		{
			ID:           artifacts[0].ID,
			EvalRunID:    run.ID,
			ArtifactID:   "pack-json",
			Kind:         "file",
			LogicalPath:  "artifacts/pack.json",
			StoredPath:   layout.EvalRunDir(7) + "/artifacts/pack.json",
			Required:     true,
			Present:      false,
			SizeBytes:    0,
			MetadataJSON: `{"missing":true}`,
		},
		{
			ID:           artifacts[1].ID,
			EvalRunID:    run.ID,
			StepID:       "pack-export",
			ArtifactID:   "stderr",
			Kind:         "stderr",
			StoredPath:   layout.EvalRunDir(7) + "/pack-export/stderr.txt",
			Required:     true,
			Present:      true,
			SizeBytes:    128,
			MetadataJSON: `{"lineCount":4}`,
		},
	}
	if !reflect.DeepEqual(artifacts, want) {
		t.Fatalf("artifacts mismatch\n got=%+v\nwant=%+v", artifacts, want)
	}
}

func assertTablesExist(t *testing.T, db *sql.DB, tableNames ...string) {
	t.Helper()

	for _, tableName := range tableNames {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, tableName).Scan(&count); err != nil {
			t.Fatalf("QueryRow(%q) error = %v", tableName, err)
		}
		if count != 1 {
			t.Fatalf("table %q missing", tableName)
		}
	}
}
