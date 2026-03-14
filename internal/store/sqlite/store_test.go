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

func TestStructuralArtifactReadModels(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	run := createTestRefreshRun(t, ctx, store, repoID, 7)
	goFileID := insertTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{
		Path:           "internal/app/service.go",
		DirectoryPath:  "internal/app",
		Extension:      ".go",
		Language:       "go",
		ContentHash:    "hash-go-1",
		LastGeneration: 7,
	})
	unsupportedFileID := insertTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{
		Path:           "scripts/tool.py",
		DirectoryPath:  "scripts",
		Extension:      ".py",
		Language:       "python",
		ContentHash:    "hash-py-1",
		LastGeneration: 7,
	})

	candidates, err := store.ListExtractionCandidates(ctx, repoID)
	if err != nil {
		t.Fatalf("ListExtractionCandidates() error = %v", err)
	}
	if len(candidates) != 2 {
		t.Fatalf("candidate count = %d, want 2", len(candidates))
	}
	if candidates[0].Path != "internal/app/service.go" || candidates[0].Language != "go" {
		t.Fatalf("first candidate = %+v", candidates[0])
	}
	if candidates[1].Path != "scripts/tool.py" || candidates[1].Language != "python" {
		t.Fatalf("second candidate = %+v", candidates[1])
	}

	supportedExtraction := repository.FileStructuralArtifacts{
		Extraction: repository.FileExtractionRecord{
			RepositoryID:        repoID,
			FileID:              goFileID,
			Path:                "internal/app/service.go",
			Language:            "go",
			AdapterName:         "tree-sitter-go",
			GrammarVersion:      "v0.25.0",
			SourceContentHash:   "hash-go-1",
			SourceGeneration:    7,
			CoverageState:       repository.ExtractionCoverageStateSupported,
			SymbolCount:         2,
			TopLevelSymbolCount: 1,
			MaxSymbolDepth:      1,
			ExtractedAt:         time.Date(2026, 3, 14, 18, 0, 0, 0, time.UTC),
			RefreshRunID:        run.ID,
		},
		Symbols: []repository.SymbolRecord{
			{
				StableKey:          "function:service",
				Path:               "internal/app/service.go",
				Language:           "go",
				Kind:               "function",
				Name:               "Service",
				QualifiedName:      "Service",
				Ordinal:            0,
				Depth:              0,
				StartByte:          0,
				EndByte:            64,
				StartRow:           0,
				StartColumn:        0,
				EndRow:             4,
				EndColumn:          1,
				NameStartByte:      5,
				NameEndByte:        12,
				SignatureStartByte: 0,
				SignatureEndByte:   18,
				IsExported:         true,
			},
			{
				StableKey:       "method:service.run",
				ParentStableKey: "function:service",
				Path:            "internal/app/service.go",
				Language:        "go",
				Kind:            "method",
				Name:            "Run",
				QualifiedName:   "Service.Run",
				Ordinal:         1,
				Depth:           1,
				StartByte:       65,
				EndByte:         120,
				StartRow:        5,
				StartColumn:     0,
				EndRow:          8,
				EndColumn:       1,
				IsExported:      true,
			},
		},
	}
	if _, err := store.ReplaceFileArtifacts(ctx, supportedExtraction); err != nil {
		t.Fatalf("ReplaceFileArtifacts() supported error = %v", err)
	}

	if _, err := store.ReplaceFileArtifacts(ctx, repository.FileStructuralArtifacts{
		Extraction: repository.FileExtractionRecord{
			RepositoryID:      repoID,
			FileID:            unsupportedFileID,
			Path:              "scripts/tool.py",
			Language:          "python",
			AdapterName:       "none",
			GrammarVersion:    "none",
			SourceContentHash: "hash-py-1",
			SourceGeneration:  7,
			CoverageState:     repository.ExtractionCoverageStateUnsupported,
			CoverageReason:    repository.ExtractionCoverageReasonUnsupportedLanguage,
			ExtractedAt:       time.Date(2026, 3, 14, 18, 1, 0, 0, time.UTC),
			RefreshRunID:      run.ID,
		},
	}); err != nil {
		t.Fatalf("ReplaceFileArtifacts() unsupported error = %v", err)
	}

	extractions, err := store.ListFileExtractions(ctx, repoID)
	if err != nil {
		t.Fatalf("ListFileExtractions() error = %v", err)
	}
	if len(extractions) != 2 {
		t.Fatalf("extraction count = %d, want 2", len(extractions))
	}
	if extractions[0].Path != "internal/app/service.go" || extractions[0].CoverageState != repository.ExtractionCoverageStateSupported {
		t.Fatalf("supported extraction = %+v", extractions[0])
	}
	if extractions[1].CoverageReason != repository.ExtractionCoverageReasonUnsupportedLanguage {
		t.Fatalf("unsupported extraction reason = %q", extractions[1].CoverageReason)
	}

	symbols, err := store.ListSymbols(ctx, repoID)
	if err != nil {
		t.Fatalf("ListSymbols() error = %v", err)
	}
	if len(symbols) != 2 {
		t.Fatalf("symbol count = %d, want 2", len(symbols))
	}
	if symbols[1].ParentSymbolID == 0 {
		t.Fatalf("nested symbol parent should be persisted: %+v", symbols[1])
	}

	summary, err := store.ReadRepositoryStructuralCoverage(ctx, repoID)
	if err != nil {
		t.Fatalf("ReadRepositoryStructuralCoverage() error = %v", err)
	}
	if summary.IncludedFileCount != 2 || summary.ExtractionCount != 2 {
		t.Fatalf("coverage summary = %+v", summary)
	}
	if summary.SupportedCount != 1 || summary.UnsupportedCount != 1 || summary.FilesWithCoverageGap != 1 || summary.TotalSymbolCount != 2 {
		t.Fatalf("coverage summary = %+v", summary)
	}

	mapRecords, err := store.LoadRepositoryMapRecords(ctx, repoID)
	if err != nil {
		t.Fatalf("LoadRepositoryMapRecords() error = %v", err)
	}
	if len(mapRecords) != 2 {
		t.Fatalf("repository map file count = %d, want 2", len(mapRecords))
	}
	if len(mapRecords[0].Symbols) != 1 || mapRecords[0].Symbols[0].Name != "Service" {
		t.Fatalf("top-level repository map symbols = %+v", mapRecords[0].Symbols)
	}
	if mapRecords[1].CoverageState != repository.ExtractionCoverageStateUnsupported || len(mapRecords[1].Symbols) != 0 {
		t.Fatalf("unsupported repository map record = %+v", mapRecords[1])
	}
}

func TestReplaceFileArtifacts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	run := createTestRefreshRun(t, ctx, store, repoID, 8)
	replacedFileID := insertTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{
		Path:           "pkg/replaced.go",
		DirectoryPath:  "pkg",
		Extension:      ".go",
		Language:       "go",
		ContentHash:    "hash-v1",
		LastGeneration: 8,
	})
	stableFileID := insertTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{
		Path:           "pkg/stable.go",
		DirectoryPath:  "pkg",
		Extension:      ".go",
		Language:       "go",
		ContentHash:    "hash-stable",
		LastGeneration: 8,
	})

	first, err := store.ReplaceFileArtifacts(ctx, repository.FileStructuralArtifacts{
		Extraction: repository.FileExtractionRecord{
			RepositoryID:        repoID,
			FileID:              replacedFileID,
			Path:                "pkg/replaced.go",
			Language:            "go",
			AdapterName:         "tree-sitter-go",
			GrammarVersion:      "v0.25.0",
			SourceContentHash:   "hash-v1",
			SourceGeneration:    8,
			CoverageState:       repository.ExtractionCoverageStateSupported,
			SymbolCount:         2,
			TopLevelSymbolCount: 2,
			MaxSymbolDepth:      0,
			ExtractedAt:         time.Date(2026, 3, 14, 19, 0, 0, 0, time.UTC),
			RefreshRunID:        run.ID,
		},
		Symbols: []repository.SymbolRecord{
			testTopLevelSymbol("pkg/replaced.go", "go", "function", "Alpha", "alpha", 0),
			testTopLevelSymbol("pkg/replaced.go", "go", "function", "Beta", "beta", 1),
		},
	})
	if err != nil {
		t.Fatalf("ReplaceFileArtifacts() first error = %v", err)
	}
	if _, err := store.ReplaceFileArtifacts(ctx, repository.FileStructuralArtifacts{
		Extraction: repository.FileExtractionRecord{
			RepositoryID:        repoID,
			FileID:              stableFileID,
			Path:                "pkg/stable.go",
			Language:            "go",
			AdapterName:         "tree-sitter-go",
			GrammarVersion:      "v0.25.0",
			SourceContentHash:   "hash-stable",
			SourceGeneration:    8,
			CoverageState:       repository.ExtractionCoverageStateSupported,
			SymbolCount:         1,
			TopLevelSymbolCount: 1,
			MaxSymbolDepth:      0,
			ExtractedAt:         time.Date(2026, 3, 14, 19, 1, 0, 0, time.UTC),
			RefreshRunID:        run.ID,
		},
		Symbols: []repository.SymbolRecord{
			testTopLevelSymbol("pkg/stable.go", "go", "type", "Stable", "stable", 0),
		},
	}); err != nil {
		t.Fatalf("ReplaceFileArtifacts() stable error = %v", err)
	}

	second, err := store.ReplaceFileArtifacts(ctx, repository.FileStructuralArtifacts{
		Extraction: repository.FileExtractionRecord{
			RepositoryID:        repoID,
			FileID:              replacedFileID,
			Path:                "pkg/replaced.go",
			Language:            "go",
			AdapterName:         "tree-sitter-go",
			GrammarVersion:      "v0.25.1",
			SourceContentHash:   "hash-v2",
			SourceGeneration:    9,
			CoverageState:       repository.ExtractionCoverageStateSupported,
			SymbolCount:         1,
			TopLevelSymbolCount: 1,
			MaxSymbolDepth:      0,
			ExtractedAt:         time.Date(2026, 3, 14, 19, 2, 0, 0, time.UTC),
			RefreshRunID:        run.ID,
		},
		Symbols: []repository.SymbolRecord{
			testTopLevelSymbol("pkg/replaced.go", "go", "function", "Gamma", "gamma", 0),
		},
	})
	if err != nil {
		t.Fatalf("ReplaceFileArtifacts() second error = %v", err)
	}

	if second.ID != first.ID {
		t.Fatalf("file extraction row should be replaced in place: first=%d second=%d", first.ID, second.ID)
	}

	symbols, err := store.ListSymbols(ctx, repoID)
	if err != nil {
		t.Fatalf("ListSymbols() error = %v", err)
	}
	if len(symbols) != 2 {
		t.Fatalf("symbol count = %d, want 2", len(symbols))
	}
	if symbols[0].Path != "pkg/replaced.go" || symbols[0].Name != "Gamma" {
		t.Fatalf("replacement symbol = %+v", symbols[0])
	}
	if symbols[1].Path != "pkg/stable.go" || symbols[1].Name != "Stable" {
		t.Fatalf("stable symbol = %+v", symbols[1])
	}

	extractions, err := store.ListFileExtractions(ctx, repoID)
	if err != nil {
		t.Fatalf("ListFileExtractions() error = %v", err)
	}
	if len(extractions) != 2 {
		t.Fatalf("extraction count = %d, want 2", len(extractions))
	}
	if extractions[0].Path != "pkg/replaced.go" || extractions[0].SourceContentHash != "hash-v2" || extractions[0].SymbolCount != 1 {
		t.Fatalf("replaced extraction = %+v", extractions[0])
	}
	if extractions[1].Path != "pkg/stable.go" || extractions[1].SourceContentHash != "hash-stable" {
		t.Fatalf("stable extraction = %+v", extractions[1])
	}
}

func TestUnsupportedExtractionState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	run := createTestRefreshRun(t, ctx, store, repoID, 3)
	fileID := insertTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{
		Path:           "assets/template.mustache",
		DirectoryPath:  "assets",
		Extension:      ".mustache",
		Language:       "mustache",
		ContentHash:    "hash-template",
		LastGeneration: 3,
	})

	if _, err := store.ReplaceFileArtifacts(ctx, repository.FileStructuralArtifacts{
		Extraction: repository.FileExtractionRecord{
			RepositoryID:        repoID,
			FileID:              fileID,
			Path:                "assets/template.mustache",
			Language:            "mustache",
			AdapterName:         "legacy-adapter",
			GrammarVersion:      "v1",
			SourceContentHash:   "hash-template",
			SourceGeneration:    2,
			CoverageState:       repository.ExtractionCoverageStateSupported,
			SymbolCount:         1,
			TopLevelSymbolCount: 1,
			MaxSymbolDepth:      0,
			ExtractedAt:         time.Date(2026, 3, 14, 19, 55, 0, 0, time.UTC),
			RefreshRunID:        run.ID,
		},
		Symbols: []repository.SymbolRecord{
			testTopLevelSymbol("assets/template.mustache", "mustache", "template", "Legacy", "legacy", 0),
		},
	}); err != nil {
		t.Fatalf("ReplaceFileArtifacts() baseline error = %v", err)
	}

	if _, err := store.ReplaceFileArtifacts(ctx, repository.FileStructuralArtifacts{
		Extraction: repository.FileExtractionRecord{
			RepositoryID:      repoID,
			FileID:            fileID,
			Path:              "assets/template.mustache",
			Language:          "mustache",
			AdapterName:       "none",
			GrammarVersion:    "none",
			SourceContentHash: "hash-template",
			SourceGeneration:  3,
			CoverageState:     repository.ExtractionCoverageStateUnsupported,
			CoverageReason:    repository.ExtractionCoverageReasonUnsupportedLanguage,
			ExtractedAt:       time.Date(2026, 3, 14, 20, 0, 0, 0, time.UTC),
			RefreshRunID:      run.ID,
		},
	}); err != nil {
		t.Fatalf("ReplaceFileArtifacts() error = %v", err)
	}

	extractions, err := store.ListFileExtractions(ctx, repoID)
	if err != nil {
		t.Fatalf("ListFileExtractions() error = %v", err)
	}
	if len(extractions) != 1 || extractions[0].CoverageState != repository.ExtractionCoverageStateUnsupported || extractions[0].CoverageReason != repository.ExtractionCoverageReasonUnsupportedLanguage {
		t.Fatalf("unsupported extraction = %+v", extractions)
	}
	if extractions[0].SourceGeneration != 3 || extractions[0].SymbolCount != 0 {
		t.Fatalf("unsupported extraction should replace prior symbols: %+v", extractions[0])
	}

	symbols, err := store.ListSymbols(ctx, repoID)
	if err != nil {
		t.Fatalf("ListSymbols() error = %v", err)
	}
	if len(symbols) != 0 {
		t.Fatalf("unsupported file should not persist symbols, got %d", len(symbols))
	}
}

func TestPartialAndFailedExtractionState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	run := createTestRefreshRun(t, ctx, store, repoID, 11)
	partialFileID := insertTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{
		Path:           "pkg/partial.go",
		DirectoryPath:  "pkg",
		Extension:      ".go",
		Language:       "go",
		ContentHash:    "hash-partial",
		LastGeneration: 11,
	})
	failedFileID := insertTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{
		Path:           "pkg/failed.go",
		DirectoryPath:  "pkg",
		Extension:      ".go",
		Language:       "go",
		ContentHash:    "hash-failed",
		LastGeneration: 11,
	})

	if _, err := store.ReplaceFileArtifacts(ctx, repository.FileStructuralArtifacts{
		Extraction: repository.FileExtractionRecord{
			RepositoryID:        repoID,
			FileID:              partialFileID,
			Path:                "pkg/partial.go",
			Language:            "go",
			AdapterName:         "tree-sitter-go",
			GrammarVersion:      "v0.25.0",
			SourceContentHash:   "hash-partial",
			SourceGeneration:    11,
			CoverageState:       repository.ExtractionCoverageStatePartial,
			CoverageReason:      repository.ExtractionCoverageReasonParseError,
			ParserErrorCount:    2,
			HasErrorNodes:       true,
			SymbolCount:         1,
			TopLevelSymbolCount: 1,
			MaxSymbolDepth:      0,
			ExtractedAt:         time.Date(2026, 3, 14, 20, 30, 0, 0, time.UTC),
			RefreshRunID:        run.ID,
		},
		Symbols: []repository.SymbolRecord{
			testTopLevelSymbol("pkg/partial.go", "go", "function", "Recovered", "recovered", 0),
		},
	}); err != nil {
		t.Fatalf("ReplaceFileArtifacts() partial error = %v", err)
	}

	if _, err := store.ReplaceFileArtifacts(ctx, repository.FileStructuralArtifacts{
		Extraction: repository.FileExtractionRecord{
			RepositoryID:      repoID,
			FileID:            failedFileID,
			Path:              "pkg/failed.go",
			Language:          "go",
			AdapterName:       "tree-sitter-go",
			GrammarVersion:    "v0.25.0",
			SourceContentHash: "hash-failed",
			SourceGeneration:  11,
			CoverageState:     repository.ExtractionCoverageStateFailed,
			CoverageReason:    repository.ExtractionCoverageReasonAdapterError,
			ParserErrorCount:  4,
			HasErrorNodes:     true,
			ExtractedAt:       time.Date(2026, 3, 14, 20, 31, 0, 0, time.UTC),
			RefreshRunID:      run.ID,
		},
	}); err != nil {
		t.Fatalf("ReplaceFileArtifacts() failed error = %v", err)
	}

	if _, err := store.ReplaceFileArtifacts(ctx, repository.FileStructuralArtifacts{
		Extraction: repository.FileExtractionRecord{
			RepositoryID:      repoID,
			FileID:            partialFileID,
			Path:              "pkg/partial.go",
			Language:          "go",
			AdapterName:       "tree-sitter-go",
			GrammarVersion:    "v0.25.1",
			SourceContentHash: "hash-partial-v2",
			SourceGeneration:  12,
			CoverageState:     repository.ExtractionCoverageStateFailed,
			CoverageReason:    repository.ExtractionCoverageReasonParseError,
			ParserErrorCount:  6,
			HasErrorNodes:     true,
			ExtractedAt:       time.Date(2026, 3, 14, 20, 32, 0, 0, time.UTC),
			RefreshRunID:      run.ID,
		},
	}); err != nil {
		t.Fatalf("ReplaceFileArtifacts() partial->failed error = %v", err)
	}

	extractions, err := store.ListFileExtractions(ctx, repoID)
	if err != nil {
		t.Fatalf("ListFileExtractions() error = %v", err)
	}
	if len(extractions) != 2 {
		t.Fatalf("extraction count = %d, want 2", len(extractions))
	}
	if extractions[0].CoverageState != repository.ExtractionCoverageStateFailed || extractions[0].SymbolCount != 0 {
		t.Fatalf("failed extraction = %+v", extractions[0])
	}
	if extractions[1].CoverageState != repository.ExtractionCoverageStateFailed || !extractions[1].HasErrorNodes || extractions[1].SymbolCount != 0 || extractions[1].SourceGeneration != 12 {
		t.Fatalf("partial file should advance to failed generation without stale symbols: %+v", extractions[1])
	}

	symbols, err := store.ListSymbols(ctx, repoID)
	if err != nil {
		t.Fatalf("ListSymbols() error = %v", err)
	}
	if len(symbols) != 0 {
		t.Fatalf("symbols = %+v", symbols)
	}

	summary, err := store.ReadRepositoryStructuralCoverage(ctx, repoID)
	if err != nil {
		t.Fatalf("ReadRepositoryStructuralCoverage() error = %v", err)
	}
	if summary.PartialCount != 0 || summary.FailedCount != 2 || summary.FilesWithCoverageGap != 2 || summary.TotalSymbolCount != 0 {
		t.Fatalf("coverage summary = %+v", summary)
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

type testFileSeed struct {
	Path           string
	DirectoryPath  string
	Extension      string
	Language       string
	ContentHash    string
	LastGeneration int64
}

func createTestRefreshRun(t *testing.T, ctx context.Context, store *Store, repoID, generation int64) repository.RefreshRunRecord {
	t.Helper()

	run, err := store.CreateRefreshRun(ctx, repository.RefreshRunRecord{
		RepositoryID: repoID,
		Generation:   generation,
		Reason:       repository.RefreshReasonManual,
		Status:       repository.RefreshRunStatusSuccess,
		StartedAt:    time.Date(2026, 3, 14, 17, 0, 0, 0, time.UTC),
		CompletedAt:  time.Date(2026, 3, 14, 17, 1, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("CreateRefreshRun() error = %v", err)
	}
	return run
}

func insertTestFileRecord(t *testing.T, ctx context.Context, store *Store, repoID, refreshRunID int64, seed testFileSeed) int64 {
	t.Helper()

	discoveredAt := time.Date(2026, 3, 14, 16, 0, 0, 0, time.UTC).Format(time.RFC3339)
	updatedAt := time.Date(2026, 3, 14, 16, 5, 0, 0, time.UTC).Format(time.RFC3339)
	result, err := store.DB().ExecContext(ctx, `
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
	`,
		repoID,
		seed.Path,
		seed.DirectoryPath,
		seed.Extension,
		seed.Language,
		128,
		seed.ContentHash,
		updatedAt,
		string(repository.IgnoreStatusIncluded),
		nil,
		updatedAt,
		discoveredAt,
		updatedAt,
		seed.LastGeneration,
		refreshRunID,
		"content_changed",
	)
	if err != nil {
		t.Fatalf("insert file %q error = %v", seed.Path, err)
	}

	fileID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId() error = %v", err)
	}
	return fileID
}

func testTopLevelSymbol(path, language, kind, name, stableKey string, ordinal int64) repository.SymbolRecord {
	return repository.SymbolRecord{
		StableKey:          stableKey,
		Path:               path,
		Language:           language,
		Kind:               kind,
		Name:               name,
		QualifiedName:      name,
		Ordinal:            ordinal,
		Depth:              0,
		StartByte:          ordinal * 10,
		EndByte:            ordinal*10 + 8,
		StartRow:           ordinal,
		StartColumn:        0,
		EndRow:             ordinal,
		EndColumn:          8,
		NameStartByte:      ordinal*10 + 1,
		NameEndByte:        ordinal*10 + 1 + int64(len(name)),
		SignatureStartByte: ordinal * 10,
		SignatureEndByte:   ordinal*10 + 8,
		IsExported:         true,
	}
}
