package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/migrations"

	_ "modernc.org/sqlite"
)

type Store struct {
	db            *sql.DB
	layout        state.Layout
	schemaVersion int
}

type RepositoryRecord struct {
	ID int64
}

func OpenOrCreateStore(ctx context.Context, layout state.Layout, repoDetectionMode string) (*Store, error) {
	if layout.DatabasePath == "" {
		return nil, fmt.Errorf("open sqlite store: database path is required")
	}
	for _, dir := range []string{layout.StateDir, layout.LogsDir, layout.TmpDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("prepare state directory %q: %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite", layout.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	store := &Store{
		db:            db,
		layout:        layout,
		schemaVersion: migrations.CurrentVersion(),
	}

	defer func() {
		if err != nil {
			_ = db.Close()
		}
	}()

	if _, err = db.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	if err = migrations.Apply(ctx, db); err != nil {
		return nil, fmt.Errorf("initialize sqlite schema: %w", err)
	}

	if _, err = layout.Ensure(repoDetectionMode, store.schemaVersion, time.Now().UTC()); err != nil {
		return nil, fmt.Errorf("sync state metadata: %w", err)
	}

	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) SchemaVersion() int {
	if s == nil {
		return 0
	}
	return s.schemaVersion
}

func (s *Store) DB() *sql.DB {
	if s == nil {
		return nil
	}
	return s.db
}

func (s *Store) UpsertRepository(ctx context.Context, root repository.RepositoryRoot, now time.Time) (RepositoryRecord, error) {
	if s == nil || s.db == nil {
		return RepositoryRecord{}, fmt.Errorf("upsert repository: store is not initialized")
	}
	if root.RootPath == "" {
		return RepositoryRecord{}, fmt.Errorf("upsert repository: root path is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	timestamp := now.UTC().Format(time.RFC3339)
	if _, err := s.db.ExecContext(
		ctx,
		`INSERT INTO repositories (
			root_path, root_real_path, detection_mode, git_common_dir, git_head_ref, git_head_commit, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(root_path) DO UPDATE SET
			root_real_path = excluded.root_real_path,
			detection_mode = excluded.detection_mode,
			git_common_dir = excluded.git_common_dir,
			git_head_ref = excluded.git_head_ref,
			git_head_commit = excluded.git_head_commit,
			updated_at = excluded.updated_at`,
		root.RootPath,
		root.Fingerprint.RootPath,
		root.DetectionMode,
		emptyToNil(root.Fingerprint.GitCommonDir),
		emptyToNil(root.Fingerprint.GitHeadRef),
		emptyToNil(root.Fingerprint.GitHeadCommit),
		timestamp,
		timestamp,
	); err != nil {
		return RepositoryRecord{}, fmt.Errorf("upsert repository %q: %w", root.RootPath, err)
	}

	var record RepositoryRecord
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM repositories WHERE root_path = ?`, root.RootPath).Scan(&record.ID); err != nil {
		return RepositoryRecord{}, fmt.Errorf("load repository %q: %w", root.RootPath, err)
	}

	return record, nil
}

func (s *Store) ReadRepositoryFreshness(ctx context.Context, repositoryID int64) (repository.RepositoryFreshness, error) {
	if s == nil || s.db == nil {
		return repository.RepositoryFreshness{}, fmt.Errorf("read repository freshness: store is not initialized")
	}

	var record repository.RepositoryFreshness
	var gitCommonDir, gitHeadRef, gitHeadCommit sql.NullString
	var lastRefreshStartedAt, lastRefreshCompletedAt sql.NullString
	var lastRefreshReason, freshnessReason sql.NullString
	var lastRefreshStatus, freshnessStatus string

	err := s.db.QueryRowContext(ctx, `
		SELECT
			id,
			root_path,
			detection_mode,
			git_common_dir,
			git_head_ref,
			git_head_commit,
			last_refresh_started_at,
			last_refresh_completed_at,
			last_refresh_reason,
			last_refresh_status,
			freshness_status,
			freshness_reason,
			current_refresh_generation,
			last_refresh_generation
		FROM repositories
		WHERE id = ?
	`, repositoryID).Scan(
		&record.RepositoryID,
		&record.RootPath,
		&record.DetectionMode,
		&gitCommonDir,
		&gitHeadRef,
		&gitHeadCommit,
		&lastRefreshStartedAt,
		&lastRefreshCompletedAt,
		&lastRefreshReason,
		&lastRefreshStatus,
		&freshnessStatus,
		&freshnessReason,
		&record.CurrentGeneration,
		&record.LastRefreshGeneration,
	)
	if err != nil {
		return repository.RepositoryFreshness{}, fmt.Errorf("read repository freshness %d: %w", repositoryID, err)
	}

	record.GitCommonDir = gitCommonDir.String
	record.GitHeadRef = gitHeadRef.String
	record.GitHeadCommit = gitHeadCommit.String
	record.LastRefreshStartedAt = parseOptionalRFC3339(lastRefreshStartedAt)
	record.LastRefreshCompletedAt = parseOptionalRFC3339(lastRefreshCompletedAt)
	record.LastRefreshReason = repository.RefreshReason(lastRefreshReason.String)
	record.LastRefreshStatus = repository.RefreshRunStatus(lastRefreshStatus)
	record.FreshnessStatus = repository.FreshnessStatus(freshnessStatus)
	record.FreshnessReason = freshnessReason.String

	return record, nil
}

func (s *Store) WriteRepositoryFreshness(ctx context.Context, record repository.RepositoryFreshness) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("write repository freshness: store is not initialized")
	}
	if record.RepositoryID == 0 {
		return fmt.Errorf("write repository freshness: repository ID is required")
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE repositories
		SET
			last_refresh_started_at = ?,
			last_refresh_completed_at = ?,
			last_refresh_reason = ?,
			last_refresh_status = ?,
			freshness_status = ?,
			freshness_reason = ?,
			current_refresh_generation = ?,
			last_refresh_generation = ?,
			updated_at = ?
		WHERE id = ?
	`,
		optionalTime(record.LastRefreshStartedAt),
		optionalTime(record.LastRefreshCompletedAt),
		emptyToNil(string(record.LastRefreshReason)),
		stringOrDefault(string(record.LastRefreshStatus), string(repository.RefreshRunStatusPending)),
		stringOrDefault(string(record.FreshnessStatus), string(repository.FreshnessStatusStale)),
		emptyToNil(record.FreshnessReason),
		record.CurrentGeneration,
		record.LastRefreshGeneration,
		time.Now().UTC().Format(time.RFC3339),
		record.RepositoryID,
	)
	if err != nil {
		return fmt.Errorf("write repository freshness %d: %w", record.RepositoryID, err)
	}

	return nil
}

func (s *Store) CreateRefreshRun(ctx context.Context, run repository.RefreshRunRecord) (repository.RefreshRunRecord, error) {
	if s == nil || s.db == nil {
		return repository.RefreshRunRecord{}, fmt.Errorf("create refresh run: store is not initialized")
	}
	if run.RepositoryID == 0 {
		return repository.RefreshRunRecord{}, fmt.Errorf("create refresh run: repository ID is required")
	}
	if run.Generation == 0 {
		return repository.RefreshRunRecord{}, fmt.Errorf("create refresh run: generation is required")
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	if run.Status == "" {
		run.Status = repository.RefreshRunStatusRunning
	}

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO refresh_runs (
			repository_id,
			generation,
			reason,
			status,
			failure_reason,
			started_at,
			completed_at,
			metadata_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		run.RepositoryID,
		run.Generation,
		string(run.Reason),
		string(run.Status),
		emptyToNil(run.FailureReason),
		run.StartedAt.UTC().Format(time.RFC3339),
		optionalTime(run.CompletedAt),
		emptyToNil(run.MetadataJSON),
	)
	if err != nil {
		return repository.RefreshRunRecord{}, fmt.Errorf("create refresh run for repository %d: %w", run.RepositoryID, err)
	}

	runID, err := result.LastInsertId()
	if err != nil {
		return repository.RefreshRunRecord{}, fmt.Errorf("load refresh run ID: %w", err)
	}
	run.ID = runID

	return run, nil
}

func (s *Store) UpdateRefreshRun(ctx context.Context, run repository.RefreshRunRecord) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("update refresh run: store is not initialized")
	}
	if run.ID == 0 {
		return fmt.Errorf("update refresh run: run ID is required")
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE refresh_runs
		SET
			status = ?,
			failure_reason = ?,
			completed_at = ?,
			metadata_json = ?
		WHERE id = ?
	`,
		string(run.Status),
		emptyToNil(run.FailureReason),
		optionalTime(run.CompletedAt),
		emptyToNil(run.MetadataJSON),
		run.ID,
	)
	if err != nil {
		return fmt.Errorf("update refresh run %d: %w", run.ID, err)
	}

	return nil
}

func (s *Store) LoadRepositorySnapshot(ctx context.Context, repositoryID int64) (repository.RepositorySnapshot, error) {
	if s == nil || s.db == nil {
		return repository.RepositorySnapshot{}, fmt.Errorf("load repository snapshot: store is not initialized")
	}

	freshness, err := s.ReadRepositoryFreshness(ctx, repositoryID)
	if err != nil {
		return repository.RepositorySnapshot{}, err
	}

	directories, err := s.loadDirectorySnapshot(ctx, repositoryID)
	if err != nil {
		return repository.RepositorySnapshot{}, err
	}

	files, err := s.loadFileSnapshot(ctx, repositoryID)
	if err != nil {
		return repository.RepositorySnapshot{}, err
	}

	return repository.RepositorySnapshot{
		Repository:  freshness,
		Directories: directories,
		Files:       files,
	}, nil
}

func (s *Store) loadDirectorySnapshot(ctx context.Context, repositoryID int64) ([]repository.DirectorySnapshotRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			path,
			parent_path,
			ignore_status,
			ignore_reason,
			discovered_at,
			subtree_fingerprint,
			included_file_count,
			included_directory_count,
			total_size_bytes,
			last_refreshed_at,
			last_refresh_generation
		FROM directories
		WHERE repository_id = ?
		ORDER BY path
	`, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("query directory snapshot for repository %d: %w", repositoryID, err)
	}
	defer rows.Close()

	var records []repository.DirectorySnapshotRecord
	for rows.Next() {
		var record repository.DirectorySnapshotRecord
		var parentPath, ignoreReason, subtreeFingerprint, lastRefreshedAt sql.NullString
		var ignoreStatus string
		var discoveredAt string

		if err := rows.Scan(
			&record.Path,
			&parentPath,
			&ignoreStatus,
			&ignoreReason,
			&discoveredAt,
			&subtreeFingerprint,
			&record.IncludedFileCount,
			&record.IncludedDirectoryCount,
			&record.TotalSizeBytes,
			&lastRefreshedAt,
			&record.LastRefreshGeneration,
		); err != nil {
			return nil, fmt.Errorf("scan directory snapshot for repository %d: %w", repositoryID, err)
		}

		record.ParentPath = normalizeNullablePath(parentPath)
		record.IgnoreStatus = repository.IgnoreStatus(ignoreStatus)
		record.IgnoreReason = repository.IgnoreReason(ignoreReason.String)
		record.DiscoveredAt = parseRequiredRFC3339(discoveredAt)
		record.SubtreeFingerprint = subtreeFingerprint.String
		record.LastRefreshedAt = parseOptionalRFC3339(lastRefreshedAt)
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate directory snapshot for repository %d: %w", repositoryID, err)
	}

	return records, nil
}

func (s *Store) loadFileSnapshot(ctx context.Context, repositoryID int64) ([]repository.PersistedFileSnapshotRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
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
		FROM files
		WHERE repository_id = ?
		ORDER BY path
	`, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("query file snapshot for repository %d: %w", repositoryID, err)
	}
	defer rows.Close()

	var records []repository.PersistedFileSnapshotRecord
	for rows.Next() {
		var record repository.PersistedFileSnapshotRecord
		var extension, language, contentHash sql.NullString
		var lastIndexedAt, ignoreReason, fsModTime sql.NullString
		var discoveredAt, updatedAt string
		var refreshRunID sql.NullInt64
		var updatedReason sql.NullString
		var ignoreStatus string

		if err := rows.Scan(
			&record.Path,
			&record.DirectoryPath,
			&extension,
			&language,
			&record.SizeBytes,
			&contentHash,
			&lastIndexedAt,
			&ignoreStatus,
			&ignoreReason,
			&fsModTime,
			&discoveredAt,
			&updatedAt,
			&record.LastSeenGeneration,
			&refreshRunID,
			&updatedReason,
		); err != nil {
			return nil, fmt.Errorf("scan file snapshot for repository %d: %w", repositoryID, err)
		}

		record.DirectoryPath = normalizeStoredDirectoryPath(record.DirectoryPath)
		record.Extension = extension.String
		record.LanguageHint = language.String
		record.ContentHash = contentHash.String
		record.LastIndexedAt = parseOptionalRFC3339(lastIndexedAt)
		record.IgnoreStatus = repository.IgnoreStatus(ignoreStatus)
		record.IgnoreReason = repository.IgnoreReason(ignoreReason.String)
		record.FilesystemModTime = parseOptionalRFC3339(fsModTime)
		record.DiscoveredAt = parseRequiredRFC3339(discoveredAt)
		record.UpdatedAt = parseRequiredRFC3339(updatedAt)
		record.RefreshRunID = refreshRunID.Int64
		record.UpdatedReason = updatedReason.String
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate file snapshot for repository %d: %w", repositoryID, err)
	}

	return records, nil
}

func emptyToNil(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func stringOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func optionalTime(timestamp time.Time) any {
	if timestamp.IsZero() {
		return nil
	}
	return timestamp.UTC().Format(time.RFC3339)
}

func normalizeNullablePath(value sql.NullString) string {
	if !value.Valid || value.String == "" {
		return ""
	}
	return filepath.ToSlash(value.String)
}

func normalizeStoredDirectoryPath(path string) string {
	if path == "" {
		return "."
	}
	return filepath.ToSlash(path)
}

func parseOptionalRFC3339(value sql.NullString) time.Time {
	if !value.Valid || value.String == "" {
		return time.Time{}
	}
	return parseRequiredRFC3339(value.String)
}

func parseRequiredRFC3339(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
