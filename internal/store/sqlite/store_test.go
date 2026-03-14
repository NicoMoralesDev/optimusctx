package sqlite

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/migrations"
)

func TestOpenOrCreateStoreInitializesEmptyDatabase(t *testing.T) {
	t.Parallel()

	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, err := OpenOrCreateStore(context.Background(), layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("OpenOrCreateStore() error = %v", err)
	}
	defer store.Close()

	if store.SchemaVersion() != migrations.CurrentVersion() {
		t.Fatalf("SchemaVersion() = %d, want %d", store.SchemaVersion(), migrations.CurrentVersion())
	}

	if _, err := os.Stat(layout.DatabasePath); err != nil {
		t.Fatalf("Stat(%q) error = %v", layout.DatabasePath, err)
	}

	metadata, err := layout.ReadMetadata()
	if err != nil {
		t.Fatalf("ReadMetadata() error = %v", err)
	}
	if metadata.SchemaVersion != migrations.CurrentVersion() {
		t.Fatalf("metadata schema version = %d", metadata.SchemaVersion)
	}

	var versionCount int
	if err := store.DB().QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&versionCount); err != nil {
		t.Fatalf("QueryRow(schema_migrations) error = %v", err)
	}
	if versionCount != migrations.CurrentVersion() {
		t.Fatalf("version count = %d, want %d", versionCount, migrations.CurrentVersion())
	}
}

func TestOpenOrCreateStoreIsIdempotent(t *testing.T) {
	t.Parallel()

	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	firstStore, err := OpenOrCreateStore(context.Background(), layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("first OpenOrCreateStore() error = %v", err)
	}

	root := repository.RepositoryRoot{
		RootPath:      layout.RepoRoot,
		DetectionMode: repository.DetectionModeGit,
		Fingerprint: repository.RepositoryFingerprint{
			RootPath:      layout.RepoRoot,
			GitCommonDir:  layout.RepoRoot + "/.git",
			GitHeadRef:    "main",
			GitHeadCommit: "0123456789abcdef0123456789abcdef01234567",
		},
	}

	record, err := firstStore.UpsertRepository(context.Background(), root, time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC))
	if err != nil {
		firstStore.Close()
		t.Fatalf("UpsertRepository() error = %v", err)
	}
	if record.ID == 0 {
		firstStore.Close()
		t.Fatal("repository record ID should be non-zero")
	}
	if err := firstStore.Close(); err != nil {
		t.Fatalf("firstStore.Close() error = %v", err)
	}

	secondStore, err := OpenOrCreateStore(context.Background(), layout, repository.DetectionModeOptimusCtxState)
	if err != nil {
		t.Fatalf("second OpenOrCreateStore() error = %v", err)
	}
	defer secondStore.Close()

	record, err = secondStore.UpsertRepository(context.Background(), root, time.Date(2026, 3, 14, 14, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("UpsertRepository() second error = %v", err)
	}
	if record.ID == 0 {
		t.Fatal("repository record ID should remain non-zero")
	}

	var repositoryCount int
	if err := secondStore.DB().QueryRow(`SELECT COUNT(*) FROM repositories`).Scan(&repositoryCount); err != nil {
		t.Fatalf("QueryRow(repositories) error = %v", err)
	}
	if repositoryCount != 1 {
		t.Fatalf("repository count = %d, want 1", repositoryCount)
	}
}

func TestSQLiteStoreReportsCorruptDatabase(t *testing.T) {
	t.Parallel()

	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}
	if _, err := layout.Ensure(repository.DetectionModeGit, 0, time.Now().UTC()); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	if err := os.WriteFile(layout.DatabasePath, []byte("not a sqlite database"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", layout.DatabasePath, err)
	}

	store, err := OpenOrCreateStore(context.Background(), layout, repository.DetectionModeGit)
	if err == nil {
		store.Close()
		t.Fatal("OpenOrCreateStore() expected error, got nil")
	}
}

func TestRefreshSchemaContracts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	freshness, err := store.ReadRepositoryFreshness(ctx, repoID)
	if err != nil {
		t.Fatalf("ReadRepositoryFreshness() error = %v", err)
	}

	if freshness.FreshnessStatus != repository.FreshnessStatusStale {
		t.Fatalf("freshness status = %q, want %q", freshness.FreshnessStatus, repository.FreshnessStatusStale)
	}
	if freshness.LastRefreshStatus != repository.RefreshRunStatusPending {
		t.Fatalf("last refresh status = %q, want %q", freshness.LastRefreshStatus, repository.RefreshRunStatusPending)
	}
	if freshness.CurrentGeneration != 0 || freshness.LastRefreshGeneration != 0 {
		t.Fatalf("unexpected generations = current %d last %d", freshness.CurrentGeneration, freshness.LastRefreshGeneration)
	}

	assertIndexColumns(t, store.DB(), "files", []string{"repository_id", "path"})
	assertIndexColumns(t, store.DB(), "directories", []string{"repository_id", "path"})
	assertIndexColumns(t, store.DB(), "refresh_runs", []string{"repository_id", "started_at"})
	assertIndexColumns(t, store.DB(), "refresh_file_events", []string{"repository_id", "path"})
}

func TestSnapshotReadModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	refreshedAt := time.Date(2026, 3, 14, 13, 30, 0, 0, time.UTC)
	if _, err := store.DB().ExecContext(ctx, `
		INSERT INTO directories (
			repository_id,
			path,
			parent_path,
			discovered_at,
			ignore_status,
			ignore_reason,
			subtree_fingerprint,
			included_file_count,
			included_directory_count,
			total_size_bytes,
			last_refreshed_at,
			last_refresh_generation
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, repoID, ".", nil, refreshedAt.Format(time.RFC3339), string(repository.IgnoreStatusIncluded), nil, "root-fingerprint", 2, 1, 64, refreshedAt.Format(time.RFC3339), 2); err != nil {
		t.Fatalf("insert root directory error = %v", err)
	}

	if _, err := store.DB().ExecContext(ctx, `
		INSERT INTO directories (
			repository_id,
			path,
			parent_path,
			discovered_at,
			ignore_status,
			ignore_reason,
			subtree_fingerprint,
			included_file_count,
			included_directory_count,
			total_size_bytes,
			last_refreshed_at,
			last_refresh_generation
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, repoID, "pkg", ".", refreshedAt.Format(time.RFC3339), string(repository.IgnoreStatusIncluded), nil, "pkg-fingerprint", 1, 0, 42, refreshedAt.Format(time.RFC3339), 2); err != nil {
		t.Fatalf("insert pkg directory error = %v", err)
	}

	run, err := store.CreateRefreshRun(ctx, repository.RefreshRunRecord{
		RepositoryID: repoID,
		Generation:   2,
		Reason:       repository.RefreshReasonManual,
		Status:       repository.RefreshRunStatusSuccess,
		StartedAt:    refreshedAt,
		CompletedAt:  refreshedAt.Add(2 * time.Minute),
		MetadataJSON: `{"paths":["pkg/file.go"]}`,
	})
	if err != nil {
		t.Fatalf("CreateRefreshRun() error = %v", err)
	}

	if _, err := store.DB().ExecContext(ctx, `
		INSERT INTO files (
			repository_id,
			path,
			directory_path,
			extension,
			language,
			size_bytes,
			content_hash,
			last_indexed_at,
			ignore_status,
			ignore_reason,
			fs_mod_time,
			discovered_at,
			updated_at,
			last_seen_generation,
			refresh_run_id,
			updated_reason
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, repoID, "pkg/file.go", "pkg", ".go", "go", 42, "abc123", refreshedAt.Format(time.RFC3339), string(repository.IgnoreStatusIncluded), nil, refreshedAt.Format(time.RFC3339), refreshedAt.Format(time.RFC3339), refreshedAt.Format(time.RFC3339), 2, run.ID, "content_changed"); err != nil {
		t.Fatalf("insert file error = %v", err)
	}

	if err := store.WriteRepositoryFreshness(ctx, repository.RepositoryFreshness{
		RepositoryID:           repoID,
		LastRefreshStartedAt:   refreshedAt,
		LastRefreshCompletedAt: run.CompletedAt,
		LastRefreshReason:      repository.RefreshReasonManual,
		LastRefreshStatus:      repository.RefreshRunStatusSuccess,
		FreshnessStatus:        repository.FreshnessStatusFresh,
		CurrentGeneration:      2,
		LastRefreshGeneration:  2,
	}); err != nil {
		t.Fatalf("WriteRepositoryFreshness() error = %v", err)
	}

	snapshot, err := store.LoadRepositorySnapshot(ctx, repoID)
	if err != nil {
		t.Fatalf("LoadRepositorySnapshot() error = %v", err)
	}

	if snapshot.Repository.CurrentGeneration != 2 || snapshot.Repository.LastRefreshGeneration != 2 {
		t.Fatalf("unexpected repository generations: %+v", snapshot.Repository)
	}
	if snapshot.Repository.FreshnessStatus != repository.FreshnessStatusFresh {
		t.Fatalf("freshness status = %q, want %q", snapshot.Repository.FreshnessStatus, repository.FreshnessStatusFresh)
	}

	if len(snapshot.Directories) != 2 {
		t.Fatalf("directory count = %d, want 2", len(snapshot.Directories))
	}
	if got := snapshot.Directories[0].Path; got != "." {
		t.Fatalf("first directory path = %q, want \".\"", got)
	}
	if snapshot.Directories[1].SubtreeFingerprint != "pkg-fingerprint" {
		t.Fatalf("pkg subtree fingerprint = %q", snapshot.Directories[1].SubtreeFingerprint)
	}

	if len(snapshot.Files) != 1 {
		t.Fatalf("file count = %d, want 1", len(snapshot.Files))
	}
	file := snapshot.Files[0]
	if file.Path != "pkg/file.go" {
		t.Fatalf("file path = %q", file.Path)
	}
	if file.LastSeenGeneration != 2 || file.RefreshRunID != run.ID {
		t.Fatalf("file generation/run = %d/%d", file.LastSeenGeneration, file.RefreshRunID)
	}
	if file.UpdatedReason != "content_changed" {
		t.Fatalf("updated reason = %q", file.UpdatedReason)
	}
}

func TestRepositoryFreshnessState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)

	initial, err := store.ReadRepositoryFreshness(ctx, repoID)
	if err != nil {
		store.Close()
		t.Fatalf("ReadRepositoryFreshness() initial error = %v", err)
	}
	if initial.FreshnessStatus != repository.FreshnessStatusStale {
		store.Close()
		t.Fatalf("initial freshness status = %q, want %q", initial.FreshnessStatus, repository.FreshnessStatusStale)
	}

	startedAt := time.Date(2026, 3, 14, 14, 0, 0, 0, time.UTC)
	completedAt := startedAt.Add(3 * time.Minute)
	if err := store.WriteRepositoryFreshness(ctx, repository.RepositoryFreshness{
		RepositoryID:           repoID,
		LastRefreshStartedAt:   startedAt,
		LastRefreshCompletedAt: completedAt,
		LastRefreshReason:      repository.RefreshReasonInit,
		LastRefreshStatus:      repository.RefreshRunStatusSuccess,
		FreshnessStatus:        repository.FreshnessStatusFresh,
		CurrentGeneration:      1,
		LastRefreshGeneration:  1,
	}); err != nil {
		store.Close()
		t.Fatalf("WriteRepositoryFreshness() fresh error = %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("store.Close() error = %v", err)
	}

	store, err = OpenOrCreateStore(ctx, layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("reopen OpenOrCreateStore() error = %v", err)
	}
	defer store.Close()

	persisted, err := store.ReadRepositoryFreshness(ctx, repoID)
	if err != nil {
		t.Fatalf("ReadRepositoryFreshness() persisted error = %v", err)
	}

	if persisted.FreshnessStatus != repository.FreshnessStatusFresh {
		t.Fatalf("persisted freshness status = %q, want %q", persisted.FreshnessStatus, repository.FreshnessStatusFresh)
	}
	if persisted.LastRefreshStatus != repository.RefreshRunStatusSuccess {
		t.Fatalf("persisted last refresh status = %q, want %q", persisted.LastRefreshStatus, repository.RefreshRunStatusSuccess)
	}
	if !persisted.LastRefreshCompletedAt.Equal(completedAt) {
		t.Fatalf("persisted completed_at = %v, want %v", persisted.LastRefreshCompletedAt, completedAt)
	}
	if persisted.CurrentGeneration != 1 || persisted.LastRefreshGeneration != 1 {
		t.Fatalf("persisted generations = current %d last %d", persisted.CurrentGeneration, persisted.LastRefreshGeneration)
	}
}

func TestRefreshRunPersistence(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	startedAt := time.Date(2026, 3, 14, 15, 0, 0, 0, time.UTC)
	run, err := store.CreateRefreshRun(ctx, repository.RefreshRunRecord{
		RepositoryID: repoID,
		Generation:   4,
		Reason:       repository.RefreshReasonManual,
		Status:       repository.RefreshRunStatusRunning,
		StartedAt:    startedAt,
		MetadataJSON: `{"mode":"incremental"}`,
	})
	if err != nil {
		store.Close()
		t.Fatalf("CreateRefreshRun() error = %v", err)
	}

	run.Status = repository.RefreshRunStatusSuccess
	run.CompletedAt = startedAt.Add(90 * time.Second)
	if err := store.UpdateRefreshRun(ctx, run); err != nil {
		store.Close()
		t.Fatalf("UpdateRefreshRun() error = %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("store.Close() error = %v", err)
	}

	store, err = OpenOrCreateStore(ctx, layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("reopen OpenOrCreateStore() error = %v", err)
	}
	defer store.Close()

	var generation int64
	var status string
	var completedAt string
	if err := store.DB().QueryRowContext(ctx, `
		SELECT generation, status, completed_at
		FROM refresh_runs
		WHERE repository_id = ? AND generation = ?
	`, repoID, 4).Scan(&generation, &status, &completedAt); err != nil {
		t.Fatalf("QueryRow(refresh_runs) error = %v", err)
	}

	if generation != 4 {
		t.Fatalf("generation = %d, want 4", generation)
	}
	if repository.RefreshRunStatus(status) != repository.RefreshRunStatusSuccess {
		t.Fatalf("status = %q, want %q", status, repository.RefreshRunStatusSuccess)
	}
	if completedAt != run.CompletedAt.Format(time.RFC3339) {
		t.Fatalf("completed_at = %q, want %q", completedAt, run.CompletedAt.Format(time.RFC3339))
	}
}

func TestDegradedRefreshMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)

	successAt := time.Date(2026, 3, 14, 16, 0, 0, 0, time.UTC)
	if err := store.WriteRepositoryFreshness(ctx, repository.RepositoryFreshness{
		RepositoryID:           repoID,
		LastRefreshStartedAt:   successAt,
		LastRefreshCompletedAt: successAt.Add(2 * time.Minute),
		LastRefreshReason:      repository.RefreshReasonInit,
		LastRefreshStatus:      repository.RefreshRunStatusSuccess,
		FreshnessStatus:        repository.FreshnessStatusFresh,
		CurrentGeneration:      3,
		LastRefreshGeneration:  3,
	}); err != nil {
		store.Close()
		t.Fatalf("WriteRepositoryFreshness() baseline error = %v", err)
	}

	run, err := store.CreateRefreshRun(ctx, repository.RefreshRunRecord{
		RepositoryID: repoID,
		Generation:   4,
		Reason:       repository.RefreshReasonManual,
		Status:       repository.RefreshRunStatusRunning,
		StartedAt:    successAt.Add(10 * time.Minute),
	})
	if err != nil {
		store.Close()
		t.Fatalf("CreateRefreshRun() error = %v", err)
	}

	run.Status = repository.RefreshRunStatusFailed
	run.FailureReason = "hashing interrupted"
	run.CompletedAt = run.StartedAt.Add(45 * time.Second)
	if err := store.UpdateRefreshRun(ctx, run); err != nil {
		store.Close()
		t.Fatalf("UpdateRefreshRun() error = %v", err)
	}

	if err := store.WriteRepositoryFreshness(ctx, repository.RepositoryFreshness{
		RepositoryID:           repoID,
		LastRefreshStartedAt:   run.StartedAt,
		LastRefreshCompletedAt: run.CompletedAt,
		LastRefreshReason:      repository.RefreshReasonManual,
		LastRefreshStatus:      repository.RefreshRunStatusFailed,
		FreshnessStatus:        repository.FreshnessStatusPartiallyDegraded,
		FreshnessReason:        run.FailureReason,
		CurrentGeneration:      4,
		LastRefreshGeneration:  3,
	}); err != nil {
		store.Close()
		t.Fatalf("WriteRepositoryFreshness() degraded error = %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("store.Close() error = %v", err)
	}

	store, err = OpenOrCreateStore(ctx, layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("reopen OpenOrCreateStore() error = %v", err)
	}
	defer store.Close()

	persisted, err := store.ReadRepositoryFreshness(ctx, repoID)
	if err != nil {
		t.Fatalf("ReadRepositoryFreshness() error = %v", err)
	}

	if persisted.FreshnessStatus != repository.FreshnessStatusPartiallyDegraded {
		t.Fatalf("freshness status = %q, want %q", persisted.FreshnessStatus, repository.FreshnessStatusPartiallyDegraded)
	}
	if persisted.FreshnessReason != "hashing interrupted" {
		t.Fatalf("freshness reason = %q", persisted.FreshnessReason)
	}
	if persisted.CurrentGeneration != 4 || persisted.LastRefreshGeneration != 3 {
		t.Fatalf("generations = current %d last %d", persisted.CurrentGeneration, persisted.LastRefreshGeneration)
	}
	if persisted.LastRefreshStatus != repository.RefreshRunStatusFailed {
		t.Fatalf("last refresh status = %q, want %q", persisted.LastRefreshStatus, repository.RefreshRunStatusFailed)
	}
}

func openStoreWithRepository(t *testing.T, ctx context.Context, layout state.Layout) (*Store, int64) {
	t.Helper()

	store, err := OpenOrCreateStore(ctx, layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("OpenOrCreateStore() error = %v", err)
	}

	record, err := store.UpsertRepository(ctx, testRepositoryRoot(layout), time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC))
	if err != nil {
		store.Close()
		t.Fatalf("UpsertRepository() error = %v", err)
	}

	return store, record.ID
}

func testRepositoryRoot(layout state.Layout) repository.RepositoryRoot {
	return repository.RepositoryRoot{
		RootPath:      layout.RepoRoot,
		DetectionMode: repository.DetectionModeGit,
		Fingerprint: repository.RepositoryFingerprint{
			RootPath:      layout.RepoRoot,
			GitCommonDir:  layout.RepoRoot + "/.git",
			GitHeadRef:    "refs/heads/main",
			GitHeadCommit: "0123456789abcdef0123456789abcdef01234567",
		},
	}
}

func assertIndexColumns(t *testing.T, db *sql.DB, tableName string, expected []string) {
	t.Helper()

	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = ?`, tableName)
	if err != nil {
		t.Fatalf("Query(index list) error = %v", err)
	}
	defer rows.Close()

	var indexNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("rows.Scan() error = %v", err)
		}
		indexNames = append(indexNames, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err() = %v", err)
	}

	slices.Sort(indexNames)
	for _, indexName := range indexNames {
		indexRows, err := db.Query(`SELECT name FROM pragma_index_info(?) ORDER BY seqno`, indexName)
		if err != nil {
			t.Fatalf("Query(pragma_index_info %q) error = %v", indexName, err)
		}

		var columns []string
		for indexRows.Next() {
			var column string
			if err := indexRows.Scan(&column); err != nil {
				indexRows.Close()
				t.Fatalf("indexRows.Scan() error = %v", err)
			}
			columns = append(columns, column)
		}
		if err := indexRows.Err(); err != nil {
			indexRows.Close()
			t.Fatalf("indexRows.Err() = %v", err)
		}
		indexRows.Close()

		if reflect.DeepEqual(columns, expected) {
			return
		}
	}

	t.Fatalf("index with columns %v not found on %s", expected, tableName)
}
