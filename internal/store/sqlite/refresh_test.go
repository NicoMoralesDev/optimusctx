package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	refreshcore "github.com/niccrow/optimusctx/internal/refresh"
	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
)

func TestApplyRefreshPlan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	root := testRepositoryRoot(layout)
	base := buildDiscoveryResult(root, time.Date(2026, 3, 14, 18, 0, 0, 0, time.UTC),
		[]repository.DirectoryRecord{
			directory(".", "", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 18, 0, 0, 0, time.UTC)),
			directory("pkg", ".", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 18, 0, 0, 0, time.UTC)),
		},
		[]repository.FileRecord{
			file("README.md", ".", "markdown", 32, "hash-readme-1", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 18, 0, 0, 0, time.UTC)),
			file("pkg/main.go", "pkg", "go", 64, "hash-main-1", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 18, 0, 0, 0, time.UTC)),
		},
	)

	baseResult, err := store.ApplyRefreshPlan(ctx, buildRefreshRequest(repoID, root, repository.RepositorySnapshot{}, base, repository.RefreshReasonInit, nil))
	if err != nil {
		t.Fatalf("ApplyRefreshPlan() baseline error = %v", err)
	}

	baseSnapshot, err := store.LoadRepositorySnapshot(ctx, repoID)
	if err != nil {
		t.Fatalf("LoadRepositorySnapshot() baseline error = %v", err)
	}

	readmeBefore := persistedFileByPath(baseSnapshot.Files, "README.md")
	mainBefore := persistedFileByPath(baseSnapshot.Files, "pkg/main.go")

	next := buildDiscoveryResult(root, time.Date(2026, 3, 14, 18, 5, 0, 0, time.UTC),
		[]repository.DirectoryRecord{
			directory(".", "", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 18, 5, 0, 0, time.UTC)),
			directory("pkg", ".", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 18, 5, 0, 0, time.UTC)),
			directory("pkg/internal", "pkg", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 18, 5, 0, 0, time.UTC)),
		},
		[]repository.FileRecord{
			file("README.md", ".", "markdown", 32, "hash-readme-1", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 18, 5, 0, 0, time.UTC)),
			file("pkg/main.go", "pkg", "go", 96, "hash-main-2", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 18, 5, 0, 0, time.UTC)),
			file("pkg/internal/helper.go", "pkg/internal", "go", 48, "hash-helper-1", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 18, 5, 0, 0, time.UTC)),
		},
	)

	incrementalResult, err := store.ApplyRefreshPlan(ctx, buildRefreshRequest(repoID, root, baseSnapshot, next, repository.RefreshReasonManual, nil))
	if err != nil {
		t.Fatalf("ApplyRefreshPlan() incremental error = %v", err)
	}

	if incrementalResult.Generation != baseResult.Generation+1 {
		t.Fatalf("generation = %d, want %d", incrementalResult.Generation, baseResult.Generation+1)
	}

	snapshot, err := store.LoadRepositorySnapshot(ctx, repoID)
	if err != nil {
		t.Fatalf("LoadRepositorySnapshot() error = %v", err)
	}

	readmeAfter := persistedFileByPath(snapshot.Files, "README.md")
	if !readmeAfter.UpdatedAt.Equal(readmeBefore.UpdatedAt) {
		t.Fatalf("README updated_at changed from %v to %v", readmeBefore.UpdatedAt, readmeAfter.UpdatedAt)
	}
	if readmeAfter.LastSeenGeneration != readmeBefore.LastSeenGeneration {
		t.Fatalf("README last_seen_generation changed from %d to %d", readmeBefore.LastSeenGeneration, readmeAfter.LastSeenGeneration)
	}

	mainAfter := persistedFileByPath(snapshot.Files, "pkg/main.go")
	if mainAfter.ContentHash != "hash-main-2" {
		t.Fatalf("pkg/main.go content hash = %q", mainAfter.ContentHash)
	}
	if mainAfter.LastSeenGeneration != incrementalResult.Generation {
		t.Fatalf("pkg/main.go generation = %d, want %d", mainAfter.LastSeenGeneration, incrementalResult.Generation)
	}
	if mainAfter.UpdatedReason != "content_changed" {
		t.Fatalf("pkg/main.go updated reason = %q", mainAfter.UpdatedReason)
	}
	if mainAfter.RefreshRunID != incrementalResult.RefreshRunID {
		t.Fatalf("pkg/main.go refresh run ID = %d, want %d", mainAfter.RefreshRunID, incrementalResult.RefreshRunID)
	}
	if reflect.DeepEqual(mainBefore, mainAfter) {
		t.Fatal("pkg/main.go should have changed after incremental refresh")
	}

	if snapshot.Repository.FreshnessStatus != repository.FreshnessStatusFresh {
		t.Fatalf("freshness status = %q, want %q", snapshot.Repository.FreshnessStatus, repository.FreshnessStatusFresh)
	}
	if snapshot.Repository.LastRefreshGeneration != incrementalResult.Generation {
		t.Fatalf("last refresh generation = %d, want %d", snapshot.Repository.LastRefreshGeneration, incrementalResult.Generation)
	}

	assertDirectoryAggregate(t, snapshot.Directories, ".", 3, 2)
	assertDirectoryAggregate(t, snapshot.Directories, "pkg", 2, 1)
	assertDirectoryAggregate(t, snapshot.Directories, "pkg/internal", 1, 0)

	assertRefreshEventTypes(t, store.DB(), repoID, incrementalResult.RefreshRunID, []string{"added", "changed"})
}

func TestIncrementalRefreshTransaction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	root := testRepositoryRoot(layout)
	base := buildDiscoveryResult(root, time.Date(2026, 3, 14, 19, 0, 0, 0, time.UTC),
		[]repository.DirectoryRecord{
			directory(".", "", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 19, 0, 0, 0, time.UTC)),
		},
		[]repository.FileRecord{
			file("main.go", ".", "go", 32, "hash-main-a", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 19, 0, 0, 0, time.UTC)),
		},
	)

	if _, err := store.ApplyRefreshPlan(ctx, buildRefreshRequest(repoID, root, repository.RepositorySnapshot{}, base, repository.RefreshReasonInit, nil)); err != nil {
		t.Fatalf("ApplyRefreshPlan() baseline error = %v", err)
	}

	before, err := store.LoadRepositorySnapshot(ctx, repoID)
	if err != nil {
		t.Fatalf("LoadRepositorySnapshot() before error = %v", err)
	}

	next := buildDiscoveryResult(root, time.Date(2026, 3, 14, 19, 5, 0, 0, time.UTC),
		[]repository.DirectoryRecord{
			directory(".", "", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 19, 5, 0, 0, time.UTC)),
			directory("pkg", ".", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 19, 5, 0, 0, time.UTC)),
		},
		[]repository.FileRecord{
			file("main.go", ".", "go", 64, "hash-main-b", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 19, 5, 0, 0, time.UTC)),
			file("pkg/new.go", "pkg", "go", 24, "hash-new", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 19, 5, 0, 0, time.UTC)),
		},
	)

	_, err = store.ApplyRefreshPlan(ctx, buildRefreshRequest(repoID, root, before, next, repository.RefreshReasonManual, func(stage string) error {
		if stage == "after_files" {
			return errors.New("forced after file updates")
		}
		return nil
	}))
	if err == nil || !strings.Contains(err.Error(), "forced after file updates") {
		t.Fatalf("ApplyRefreshPlan() error = %v, want injected failure", err)
	}

	after, err := store.LoadRepositorySnapshot(ctx, repoID)
	if err != nil {
		t.Fatalf("LoadRepositorySnapshot() after error = %v", err)
	}
	if !reflect.DeepEqual(before.Files, after.Files) {
		t.Fatalf("files changed after rolled back refresh:\nbefore=%#v\nafter=%#v", before.Files, after.Files)
	}
	if !reflect.DeepEqual(before.Directories, after.Directories) {
		t.Fatalf("directories changed after rolled back refresh:\nbefore=%#v\nafter=%#v", before.Directories, after.Directories)
	}
}

func TestDeletedFilesAreRemoved(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	root := testRepositoryRoot(layout)
	base := buildDiscoveryResult(root, time.Date(2026, 3, 14, 20, 0, 0, 0, time.UTC),
		[]repository.DirectoryRecord{
			directory(".", "", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 20, 0, 0, 0, time.UTC)),
			directory("docs", ".", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 20, 0, 0, 0, time.UTC)),
		},
		[]repository.FileRecord{
			file("docs/readme.md", "docs", "markdown", 80, "hash-docs", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 20, 0, 0, 0, time.UTC)),
			file("keep.go", ".", "go", 24, "hash-keep", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 20, 0, 0, 0, time.UTC)),
		},
	)

	if _, err := store.ApplyRefreshPlan(ctx, buildRefreshRequest(repoID, root, repository.RepositorySnapshot{}, base, repository.RefreshReasonInit, nil)); err != nil {
		t.Fatalf("ApplyRefreshPlan() baseline error = %v", err)
	}

	persisted, err := store.LoadRepositorySnapshot(ctx, repoID)
	if err != nil {
		t.Fatalf("LoadRepositorySnapshot() error = %v", err)
	}

	next := buildDiscoveryResult(root, time.Date(2026, 3, 14, 20, 10, 0, 0, time.UTC),
		[]repository.DirectoryRecord{
			directory(".", "", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 20, 10, 0, 0, time.UTC)),
		},
		[]repository.FileRecord{
			file("keep.go", ".", "go", 24, "hash-keep", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 20, 10, 0, 0, time.UTC)),
		},
	)

	result, err := store.ApplyRefreshPlan(ctx, buildRefreshRequest(repoID, root, persisted, next, repository.RefreshReasonManual, nil))
	if err != nil {
		t.Fatalf("ApplyRefreshPlan() delete error = %v", err)
	}

	assertFileMissing(t, store.DB(), repoID, "docs/readme.md")
	assertDirectoryMissing(t, store.DB(), repoID, "docs")
	assertRefreshEventTypes(t, store.DB(), repoID, result.RefreshRunID, []string{"deleted"})
}

func TestDegradedRefreshState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	root := testRepositoryRoot(layout)
	base := buildDiscoveryResult(root, time.Date(2026, 3, 14, 21, 0, 0, 0, time.UTC),
		[]repository.DirectoryRecord{
			directory(".", "", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 21, 0, 0, 0, time.UTC)),
		},
		[]repository.FileRecord{
			file("main.go", ".", "go", 48, "hash-main-initial", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 21, 0, 0, 0, time.UTC)),
		},
	)

	initialResult, err := store.ApplyRefreshPlan(ctx, buildRefreshRequest(repoID, root, repository.RepositorySnapshot{}, base, repository.RefreshReasonInit, nil))
	if err != nil {
		t.Fatalf("ApplyRefreshPlan() baseline error = %v", err)
	}

	persisted, err := store.LoadRepositorySnapshot(ctx, repoID)
	if err != nil {
		t.Fatalf("LoadRepositorySnapshot() persisted error = %v", err)
	}

	next := buildDiscoveryResult(root, time.Date(2026, 3, 14, 21, 10, 0, 0, time.UTC),
		[]repository.DirectoryRecord{
			directory(".", "", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 21, 10, 0, 0, time.UTC)),
		},
		[]repository.FileRecord{
			file("main.go", ".", "go", 96, "hash-main-updated", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 21, 10, 0, 0, time.UTC)),
		},
	)

	_, err = store.ApplyRefreshPlan(ctx, buildRefreshRequest(repoID, root, persisted, next, repository.RefreshReasonManual, func(stage string) error {
		if stage == "after_directories" {
			return errors.New("reconcile interrupted")
		}
		return nil
	}))
	if err == nil || !strings.Contains(err.Error(), "reconcile interrupted") {
		t.Fatalf("ApplyRefreshPlan() error = %v, want reconcile interrupted", err)
	}

	freshness, err := store.ReadRepositoryFreshness(ctx, repoID)
	if err != nil {
		t.Fatalf("ReadRepositoryFreshness() error = %v", err)
	}
	if freshness.FreshnessStatus != repository.FreshnessStatusPartiallyDegraded {
		t.Fatalf("freshness status = %q, want %q", freshness.FreshnessStatus, repository.FreshnessStatusPartiallyDegraded)
	}
	if freshness.LastRefreshStatus != repository.RefreshRunStatusFailed {
		t.Fatalf("last refresh status = %q, want %q", freshness.LastRefreshStatus, repository.RefreshRunStatusFailed)
	}
	if freshness.LastRefreshGeneration != initialResult.Generation {
		t.Fatalf("last refresh generation = %d, want %d", freshness.LastRefreshGeneration, initialResult.Generation)
	}
	if freshness.CurrentGeneration != initialResult.Generation+1 {
		t.Fatalf("current generation = %d, want %d", freshness.CurrentGeneration, initialResult.Generation+1)
	}

	var status string
	var failureReason sql.NullString
	if err := store.DB().QueryRowContext(ctx, `
		SELECT status, failure_reason
		FROM refresh_runs
		WHERE repository_id = ? AND generation = ?
	`, repoID, initialResult.Generation+1).Scan(&status, &failureReason); err != nil {
		t.Fatalf("QueryRow(refresh_runs) error = %v", err)
	}
	if repository.RefreshRunStatus(status) != repository.RefreshRunStatusFailed {
		t.Fatalf("refresh run status = %q, want %q", status, repository.RefreshRunStatusFailed)
	}
	if !failureReason.Valid || failureReason.String != "reconcile interrupted" {
		t.Fatalf("failure_reason = %#v, want reconcile interrupted", failureReason)
	}
}

func TestDegradedRefreshRecovery(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	root := testRepositoryRoot(layout)
	base := buildDiscoveryResult(root, time.Date(2026, 3, 14, 21, 30, 0, 0, time.UTC),
		[]repository.DirectoryRecord{
			directory(".", "", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 21, 30, 0, 0, time.UTC)),
		},
		[]repository.FileRecord{
			file("main.go", ".", "go", 48, "hash-main-initial", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 21, 30, 0, 0, time.UTC)),
		},
	)

	initialResult, err := store.ApplyRefreshPlan(ctx, buildRefreshRequest(repoID, root, repository.RepositorySnapshot{}, base, repository.RefreshReasonInit, nil))
	if err != nil {
		t.Fatalf("ApplyRefreshPlan() baseline error = %v", err)
	}

	beforeFailure, err := store.LoadRepositorySnapshot(ctx, repoID)
	if err != nil {
		t.Fatalf("LoadRepositorySnapshot() before failure error = %v", err)
	}

	failedResult := buildDiscoveryResult(root, time.Date(2026, 3, 14, 21, 40, 0, 0, time.UTC),
		[]repository.DirectoryRecord{
			directory(".", "", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 21, 40, 0, 0, time.UTC)),
		},
		[]repository.FileRecord{
			file("main.go", ".", "go", 96, "hash-main-updated", repository.IgnoreStatusIncluded, "", time.Date(2026, 3, 14, 21, 40, 0, 0, time.UTC)),
		},
	)

	_, err = store.ApplyRefreshPlan(ctx, buildRefreshRequest(repoID, root, beforeFailure, failedResult, repository.RefreshReasonManual, func(stage string) error {
		if stage == "after_files" {
			return errors.New("forced after file updates")
		}
		return nil
	}))
	if err == nil || !strings.Contains(err.Error(), "forced after file updates") {
		t.Fatalf("ApplyRefreshPlan() error = %v, want injected failure", err)
	}

	afterFailure, err := store.LoadRepositorySnapshot(ctx, repoID)
	if err != nil {
		t.Fatalf("LoadRepositorySnapshot() after failure error = %v", err)
	}
	if !reflect.DeepEqual(beforeFailure.Files, afterFailure.Files) {
		t.Fatalf("files changed after failed refresh:\nbefore=%#v\nafter=%#v", beforeFailure.Files, afterFailure.Files)
	}
	if afterFailure.Repository.LastRefreshGeneration != initialResult.Generation {
		t.Fatalf("last refresh generation = %d, want %d", afterFailure.Repository.LastRefreshGeneration, initialResult.Generation)
	}
	if afterFailure.Repository.CurrentGeneration != initialResult.Generation+1 {
		t.Fatalf("current generation = %d, want %d", afterFailure.Repository.CurrentGeneration, initialResult.Generation+1)
	}
	if afterFailure.Repository.FreshnessStatus != repository.FreshnessStatusPartiallyDegraded {
		t.Fatalf("freshness status = %q, want %q", afterFailure.Repository.FreshnessStatus, repository.FreshnessStatusPartiallyDegraded)
	}

	recovered, err := store.ApplyRefreshPlan(ctx, buildRefreshRequest(repoID, root, afterFailure, failedResult, repository.RefreshReasonManual, nil))
	if err != nil {
		t.Fatalf("ApplyRefreshPlan() recovery error = %v", err)
	}
	if recovered.Generation != initialResult.Generation+2 {
		t.Fatalf("generation = %d, want %d", recovered.Generation, initialResult.Generation+2)
	}
	if recovered.FreshnessStatus != repository.FreshnessStatusFresh {
		t.Fatalf("freshness status = %q, want %q", recovered.FreshnessStatus, repository.FreshnessStatusFresh)
	}

	afterRecovery, err := store.LoadRepositorySnapshot(ctx, repoID)
	if err != nil {
		t.Fatalf("LoadRepositorySnapshot() after recovery error = %v", err)
	}
	mainFile := persistedFileByPath(afterRecovery.Files, "main.go")
	if mainFile.ContentHash != "hash-main-updated" {
		t.Fatalf("main.go content hash = %q, want hash-main-updated", mainFile.ContentHash)
	}
	if afterRecovery.Repository.LastRefreshGeneration != recovered.Generation {
		t.Fatalf("last refresh generation = %d, want %d", afterRecovery.Repository.LastRefreshGeneration, recovered.Generation)
	}
	if afterRecovery.Repository.FreshnessStatus != repository.FreshnessStatusFresh {
		t.Fatalf("freshness status = %q, want %q", afterRecovery.Repository.FreshnessStatus, repository.FreshnessStatusFresh)
	}
}

func buildRefreshRequest(repoID int64, root repository.RepositoryRoot, persisted repository.RepositorySnapshot, current repository.DiscoveryResult, reason repository.RefreshReason, inject func(string) error) ApplyRefreshPlanRequest {
	currentSnapshot := refreshcore.CurrentSnapshot(current)
	persistedRefreshSnapshot := refreshcore.PersistedSnapshot(persisted)
	diff := refreshcore.DiffSnapshots(currentSnapshot, persistedRefreshSnapshot)
	affected := refreshcore.AffectedDirectories(currentSnapshot, persistedRefreshSnapshot, diff)
	fingerprints := refreshcore.ComputeSubtreeFingerprints(currentSnapshot, persistedRefreshSnapshot, affected)

	return ApplyRefreshPlanRequest{
		RepositoryID:      repoID,
		RepositoryRoot:    root,
		Reason:            reason,
		StartedAt:         current.Files[0].DiscoveredAt,
		CompletedAt:       current.Files[0].DiscoveredAt.Add(30 * time.Second),
		CurrentResult:     current,
		PersistedSnapshot: persisted,
		Diff:              diff,
		AffectedPaths:     affected,
		Fingerprints:      fingerprints,
		InjectFailure:     inject,
	}
}

func buildDiscoveryResult(root repository.RepositoryRoot, discoveredAt time.Time, directories []repository.DirectoryRecord, files []repository.FileRecord) repository.DiscoveryResult {
	return repository.DiscoveryResult{
		Repository: repository.RepositoryRecord{
			RootPath:      root.RootPath,
			DetectionMode: root.DetectionMode,
			GitCommonDir:  root.Fingerprint.GitCommonDir,
			GitHeadRef:    root.Fingerprint.GitHeadRef,
			GitHeadCommit: root.Fingerprint.GitHeadCommit,
		},
		Directories: directories,
		Files:       files,
	}
}

func directory(path, parent string, status repository.IgnoreStatus, reason repository.IgnoreReason, discoveredAt time.Time) repository.DirectoryRecord {
	return repository.DirectoryRecord{
		Path:         path,
		ParentPath:   parent,
		IgnoreStatus: status,
		IgnoreReason: reason,
		DiscoveredAt: discoveredAt,
	}
}

func file(path, dir, language string, size int64, hash string, status repository.IgnoreStatus, reason repository.IgnoreReason, discoveredAt time.Time) repository.FileRecord {
	return repository.FileRecord{
		Path:              path,
		DirectoryPath:     dir,
		Extension:         extensionFromLanguage(language),
		LanguageHint:      language,
		SizeBytes:         size,
		ContentHash:       hash,
		LastIndexedAt:     discoveredAt,
		FilesystemModTime: discoveredAt,
		IgnoreStatus:      status,
		IgnoreReason:      reason,
		DiscoveredAt:      discoveredAt,
	}
}

func extensionFromLanguage(language string) string {
	switch language {
	case "go":
		return ".go"
	case "markdown":
		return ".md"
	default:
		return ""
	}
}

func persistedFileByPath(files []repository.PersistedFileSnapshotRecord, path string) repository.PersistedFileSnapshotRecord {
	for _, file := range files {
		if file.Path == path {
			return file
		}
	}
	return repository.PersistedFileSnapshotRecord{}
}

func assertDirectoryAggregate(t *testing.T, directories []repository.DirectorySnapshotRecord, path string, fileCount, directoryCount int64) {
	t.Helper()

	for _, directory := range directories {
		if directory.Path != path {
			continue
		}
		if directory.IncludedFileCount != fileCount || directory.IncludedDirectoryCount != directoryCount {
			t.Fatalf("%s aggregates = files %d dirs %d, want files %d dirs %d", path, directory.IncludedFileCount, directory.IncludedDirectoryCount, fileCount, directoryCount)
		}
		return
	}

	t.Fatalf("directory %q not found", path)
}

func assertRefreshEventTypes(t *testing.T, db *sql.DB, repoID, refreshRunID int64, want []string) {
	t.Helper()

	rows, err := db.Query(`SELECT event_type FROM refresh_file_events WHERE repository_id = ? AND refresh_run_id = ? ORDER BY event_type`, repoID, refreshRunID)
	if err != nil {
		t.Fatalf("Query(refresh_file_events) error = %v", err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var eventType string
		if err := rows.Scan(&eventType); err != nil {
			t.Fatalf("rows.Scan() error = %v", err)
		}
		got = append(got, eventType)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err() = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("event types = %v, want %v", got, want)
	}
}

func assertFileMissing(t *testing.T, db *sql.DB, repoID int64, path string) {
	t.Helper()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM files WHERE repository_id = ? AND path = ?`, repoID, path).Scan(&count); err != nil {
		t.Fatalf("QueryRow(files) error = %v", err)
	}
	if count != 0 {
		t.Fatalf("expected file %q to be removed, count=%d", path, count)
	}
}

func assertDirectoryMissing(t *testing.T, db *sql.DB, repoID int64, path string) {
	t.Helper()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM directories WHERE repository_id = ? AND path = ?`, repoID, path).Scan(&count); err != nil {
		t.Fatalf("QueryRow(directories) error = %v", err)
	}
	if count != 0 {
		t.Fatalf("expected directory %q to be removed, count=%d", path, count)
	}
}
