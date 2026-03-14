package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	refreshcore "github.com/niccrow/optimusctx/internal/refresh"
	"github.com/niccrow/optimusctx/internal/repository"
)

type ApplyRefreshPlanRequest struct {
	RepositoryRoot    repository.RepositoryRoot
	RepositoryID      int64
	Reason            repository.RefreshReason
	StartedAt         time.Time
	CompletedAt       time.Time
	CurrentResult     repository.DiscoveryResult
	PersistedSnapshot repository.RepositorySnapshot
	Diff              refreshcore.Diff
	AffectedPaths     []string
	Fingerprints      map[string]string
	MetadataJSON      string
	InjectFailure     func(stage string) error
}

type ApplyRefreshPlanResult struct {
	RepositoryID          int64
	RefreshRunID          int64
	Generation            int64
	LastRefreshGeneration int64
	FreshnessStatus       repository.FreshnessStatus
}

func (s *Store) ApplyRefreshPlan(ctx context.Context, req ApplyRefreshPlanRequest) (ApplyRefreshPlanResult, error) {
	if s == nil || s.db == nil {
		return ApplyRefreshPlanResult{}, fmt.Errorf("apply refresh plan: store is not initialized")
	}
	if req.RepositoryRoot.RootPath == "" {
		return ApplyRefreshPlanResult{}, fmt.Errorf("apply refresh plan: repository root is required")
	}
	if req.StartedAt.IsZero() {
		req.StartedAt = time.Now().UTC()
	}
	if req.CompletedAt.IsZero() {
		req.CompletedAt = req.StartedAt
	}
	if req.Reason == "" {
		req.Reason = repository.RefreshReasonManual
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ApplyRefreshPlanResult{}, fmt.Errorf("begin refresh transaction: %w", err)
	}

	rolledBack := false
	defer func() {
		if !rolledBack && tx != nil {
			_ = tx.Rollback()
		}
	}()

	repoRecord, err := s.upsertRepository(ctx, tx, req.RepositoryRoot, req.StartedAt)
	if err != nil {
		rolledBack = true
		_ = tx.Rollback()
		return ApplyRefreshPlanResult{}, fmt.Errorf("ensure repository for refresh: %w", err)
	}
	req.RepositoryID = repoRecord.ID

	priorFreshness, err := readRepositoryFreshnessTx(ctx, tx, req.RepositoryID)
	if err != nil {
		rolledBack = true
		_ = tx.Rollback()
		return ApplyRefreshPlanResult{}, err
	}

	generation := priorFreshness.CurrentGeneration + 1
	run, err := createRefreshRunTx(ctx, tx, repository.RefreshRunRecord{
		RepositoryID: req.RepositoryID,
		Generation:   generation,
		Reason:       req.Reason,
		Status:       repository.RefreshRunStatusRunning,
		StartedAt:    req.StartedAt,
		MetadataJSON: req.MetadataJSON,
	})
	if err != nil {
		rolledBack = true
		_ = tx.Rollback()
		return ApplyRefreshPlanResult{}, err
	}

	if err := applyFileChanges(ctx, tx, req, run.ID, generation); err != nil {
		rolledBack = true
		_ = tx.Rollback()
		recordErr := s.recordRefreshFailure(ctx, req, priorFreshness, generation, err)
		if recordErr != nil {
			return ApplyRefreshPlanResult{}, fmt.Errorf("apply refresh plan: %v (plus failure recording error: %w)", err, recordErr)
		}
		return ApplyRefreshPlanResult{}, fmt.Errorf("apply refresh plan: %w", err)
	}

	if err := maybeInjectRefreshFailure(req, "after_files"); err != nil {
		rolledBack = true
		_ = tx.Rollback()
		recordErr := s.recordRefreshFailure(ctx, req, priorFreshness, generation, err)
		if recordErr != nil {
			return ApplyRefreshPlanResult{}, fmt.Errorf("apply refresh plan: %v (plus failure recording error: %w)", err, recordErr)
		}
		return ApplyRefreshPlanResult{}, fmt.Errorf("apply refresh plan: %w", err)
	}

	if err := applyDirectoryChanges(ctx, tx, req, generation, req.CompletedAt); err != nil {
		rolledBack = true
		_ = tx.Rollback()
		recordErr := s.recordRefreshFailure(ctx, req, priorFreshness, generation, err)
		if recordErr != nil {
			return ApplyRefreshPlanResult{}, fmt.Errorf("apply refresh plan: %v (plus failure recording error: %w)", err, recordErr)
		}
		return ApplyRefreshPlanResult{}, fmt.Errorf("apply refresh plan: %w", err)
	}

	if err := maybeInjectRefreshFailure(req, "after_directories"); err != nil {
		rolledBack = true
		_ = tx.Rollback()
		recordErr := s.recordRefreshFailure(ctx, req, priorFreshness, generation, err)
		if recordErr != nil {
			return ApplyRefreshPlanResult{}, fmt.Errorf("apply refresh plan: %v (plus failure recording error: %w)", err, recordErr)
		}
		return ApplyRefreshPlanResult{}, fmt.Errorf("apply refresh plan: %w", err)
	}

	if err := insertRefreshEvents(ctx, tx, req, run.ID, req.CompletedAt); err != nil {
		rolledBack = true
		_ = tx.Rollback()
		recordErr := s.recordRefreshFailure(ctx, req, priorFreshness, generation, err)
		if recordErr != nil {
			return ApplyRefreshPlanResult{}, fmt.Errorf("apply refresh plan: %v (plus failure recording error: %w)", err, recordErr)
		}
		return ApplyRefreshPlanResult{}, fmt.Errorf("apply refresh plan: %w", err)
	}

	run.Status = repository.RefreshRunStatusSuccess
	run.CompletedAt = req.CompletedAt
	if run.MetadataJSON == "" {
		run.MetadataJSON = buildRefreshMetadataJSON(req)
	}
	if err := updateRefreshRunTx(ctx, tx, run); err != nil {
		rolledBack = true
		_ = tx.Rollback()
		return ApplyRefreshPlanResult{}, err
	}

	if err := writeRepositoryFreshnessTx(ctx, tx, repository.RepositoryFreshness{
		RepositoryID:           req.RepositoryID,
		LastRefreshStartedAt:   req.StartedAt,
		LastRefreshCompletedAt: req.CompletedAt,
		LastRefreshReason:      req.Reason,
		LastRefreshStatus:      repository.RefreshRunStatusSuccess,
		FreshnessStatus:        repository.FreshnessStatusFresh,
		CurrentGeneration:      generation,
		LastRefreshGeneration:  generation,
	}); err != nil {
		rolledBack = true
		_ = tx.Rollback()
		return ApplyRefreshPlanResult{}, err
	}

	if err := tx.Commit(); err != nil {
		rolledBack = true
		_ = tx.Rollback()
		return ApplyRefreshPlanResult{}, fmt.Errorf("commit refresh transaction: %w", err)
	}
	rolledBack = true
	tx = nil

	return ApplyRefreshPlanResult{
		RepositoryID:          req.RepositoryID,
		RefreshRunID:          run.ID,
		Generation:            generation,
		LastRefreshGeneration: generation,
		FreshnessStatus:       repository.FreshnessStatusFresh,
	}, nil
}

func (s *Store) recordRefreshFailure(ctx context.Context, req ApplyRefreshPlanRequest, prior repository.RepositoryFreshness, generation int64, applyErr error) error {
	run, err := s.CreateRefreshRun(ctx, repository.RefreshRunRecord{
		RepositoryID: req.RepositoryID,
		Generation:   generation,
		Reason:       req.Reason,
		Status:       repository.RefreshRunStatusRunning,
		StartedAt:    req.StartedAt,
		MetadataJSON: req.MetadataJSON,
	})
	if err != nil {
		return fmt.Errorf("create failure refresh run: %w", err)
	}

	run.Status = repository.RefreshRunStatusFailed
	run.FailureReason = applyErr.Error()
	run.CompletedAt = req.CompletedAt
	if run.MetadataJSON == "" {
		run.MetadataJSON = buildRefreshMetadataJSON(req)
	}
	if err := s.UpdateRefreshRun(ctx, run); err != nil {
		return fmt.Errorf("update failure refresh run: %w", err)
	}

	return s.WriteRepositoryFreshness(ctx, repository.RepositoryFreshness{
		RepositoryID:           req.RepositoryID,
		LastRefreshStartedAt:   req.StartedAt,
		LastRefreshCompletedAt: req.CompletedAt,
		LastRefreshReason:      req.Reason,
		LastRefreshStatus:      repository.RefreshRunStatusFailed,
		FreshnessStatus:        repository.FreshnessStatusPartiallyDegraded,
		FreshnessReason:        applyErr.Error(),
		CurrentGeneration:      generation,
		LastRefreshGeneration:  prior.LastRefreshGeneration,
	})
}

func applyFileChanges(ctx context.Context, tx *sql.Tx, req ApplyRefreshPlanRequest, refreshRunID, generation int64) error {
	currentFiles := currentFilesByPath(req.CurrentResult.Files)

	if err := upsertFiles(ctx, tx, req.RepositoryID, refreshRunID, generation, req.CompletedAt, currentFiles, req.Diff.Added, "added"); err != nil {
		return err
	}
	if err := upsertFiles(ctx, tx, req.RepositoryID, refreshRunID, generation, req.CompletedAt, currentFiles, req.Diff.Changed, "content_changed"); err != nil {
		return err
	}
	if err := upsertFiles(ctx, tx, req.RepositoryID, refreshRunID, generation, req.CompletedAt, currentFiles, req.Diff.NewlyIgnored, "newly_ignored"); err != nil {
		return err
	}
	if err := upsertFiles(ctx, tx, req.RepositoryID, refreshRunID, generation, req.CompletedAt, currentFiles, req.Diff.Reincluded, "reincluded"); err != nil {
		return err
	}

	for _, change := range req.Diff.Moved {
		if _, err := tx.ExecContext(ctx, `DELETE FROM files WHERE repository_id = ? AND path = ?`, req.RepositoryID, change.PreviousPath); err != nil {
			return fmt.Errorf("delete moved file %q: %w", change.PreviousPath, err)
		}
		if err := upsertSingleFile(ctx, tx, req.RepositoryID, refreshRunID, generation, req.CompletedAt, currentFiles[change.Path], "moved"); err != nil {
			return err
		}
	}

	for _, change := range req.Diff.Deleted {
		if _, err := tx.ExecContext(ctx, `DELETE FROM files WHERE repository_id = ? AND path = ?`, req.RepositoryID, change.Path); err != nil {
			return fmt.Errorf("delete file %q: %w", change.Path, err)
		}
	}

	return nil
}

func upsertFiles(ctx context.Context, tx *sql.Tx, repositoryID, refreshRunID, generation int64, refreshedAt time.Time, currentFiles map[string]repository.FileRecord, changes []refreshcore.FileChange, reason string) error {
	for _, change := range changes {
		file, ok := currentFiles[change.Path]
		if !ok {
			return fmt.Errorf("load current file %q for upsert", change.Path)
		}
		if err := upsertSingleFile(ctx, tx, repositoryID, refreshRunID, generation, refreshedAt, file, reason); err != nil {
			return err
		}
	}
	return nil
}

func upsertSingleFile(ctx context.Context, tx *sql.Tx, repositoryID, refreshRunID, generation int64, refreshedAt time.Time, file repository.FileRecord, reason string) error {
	discoveredAt := file.DiscoveredAt
	if discoveredAt.IsZero() {
		discoveredAt = refreshedAt
	}

	_, err := tx.ExecContext(ctx, `
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
		ON CONFLICT(repository_id, path) DO UPDATE SET
			directory_path = excluded.directory_path,
			extension = excluded.extension,
			language = excluded.language,
			size_bytes = excluded.size_bytes,
			content_hash = excluded.content_hash,
			last_indexed_at = excluded.last_indexed_at,
			ignore_status = excluded.ignore_status,
			ignore_reason = excluded.ignore_reason,
			fs_mod_time = excluded.fs_mod_time,
			updated_at = excluded.updated_at,
			last_seen_generation = excluded.last_seen_generation,
			refresh_run_id = excluded.refresh_run_id,
			updated_reason = excluded.updated_reason
	`,
		repositoryID,
		file.Path,
		normalizeStoredDirectory(file.DirectoryPath),
		emptyToNil(file.Extension),
		emptyToNil(file.LanguageHint),
		file.SizeBytes,
		optionalContentHashForFile(file),
		optionalTime(file.LastIndexedAt),
		string(file.IgnoreStatus),
		optionalIgnoreReasonForFile(file),
		optionalTime(file.FilesystemModTime),
		discoveredAt.UTC().Format(time.RFC3339),
		refreshedAt.UTC().Format(time.RFC3339),
		generation,
		refreshRunID,
		reason,
	)
	if err != nil {
		return fmt.Errorf("upsert file %q: %w", file.Path, err)
	}
	return nil
}

func applyDirectoryChanges(ctx context.Context, tx *sql.Tx, req ApplyRefreshPlanRequest, generation int64, refreshedAt time.Time) error {
	currentDirs := currentDirectoriesByPath(req.CurrentResult.Directories)
	persistedDirs := persistedDirectoriesByPath(req.PersistedSnapshot.Directories)
	metrics := computeDirectoryMetrics(req.CurrentResult, req.Fingerprints)

	for path := range persistedDirs {
		if _, ok := currentDirs[path]; ok {
			continue
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM directories WHERE repository_id = ? AND path = ?`, req.RepositoryID, path); err != nil {
			return fmt.Errorf("delete directory %q: %w", path, err)
		}
	}

	affected := req.AffectedPaths
	if len(affected) == 0 {
		for path := range currentDirs {
			affected = append(affected, path)
		}
		sort.Strings(affected)
	}

	for _, path := range affected {
		directory, ok := currentDirs[path]
		if !ok {
			continue
		}
		metric := metrics[path]
		if _, err := tx.ExecContext(ctx, `
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
			ON CONFLICT(repository_id, path) DO UPDATE SET
				parent_path = excluded.parent_path,
				discovered_at = excluded.discovered_at,
				ignore_status = excluded.ignore_status,
				ignore_reason = excluded.ignore_reason,
				subtree_fingerprint = excluded.subtree_fingerprint,
				included_file_count = excluded.included_file_count,
				included_directory_count = excluded.included_directory_count,
				total_size_bytes = excluded.total_size_bytes,
				last_refreshed_at = excluded.last_refreshed_at,
				last_refresh_generation = excluded.last_refresh_generation
		`,
			req.RepositoryID,
			directory.Path,
			emptyDirectoryParentPath(directory.ParentPath),
			directory.DiscoveredAt.UTC().Format(time.RFC3339),
			string(directory.IgnoreStatus),
			optionalDirectoryIgnoreReason(directory),
			emptyToNil(metric.Fingerprint),
			metric.IncludedFileCount,
			metric.IncludedDirectoryCount,
			metric.TotalSizeBytes,
			refreshedAt.UTC().Format(time.RFC3339),
			generation,
		); err != nil {
			return fmt.Errorf("upsert directory %q: %w", path, err)
		}
	}

	return nil
}

func insertRefreshEvents(ctx context.Context, tx *sql.Tx, req ApplyRefreshPlanRequest, refreshRunID int64, occurredAt time.Time) error {
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO refresh_file_events (
			repository_id,
			refresh_run_id,
			path,
			previous_path,
			event_type,
			content_hash,
			occurred_at,
			metadata_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare refresh file event insert: %w", err)
	}
	defer stmt.Close()

	writeEvent := func(change refreshcore.FileChange, eventType string) error {
		metadata := eventMetadata(change)
		_, err := stmt.ExecContext(
			ctx,
			req.RepositoryID,
			refreshRunID,
			change.Path,
			emptyToNil(change.PreviousPath),
			eventType,
			emptyToNil(change.ContentHash),
			occurredAt.UTC().Format(time.RFC3339),
			emptyToNil(metadata),
		)
		if err != nil {
			return fmt.Errorf("insert refresh event for %q: %w", change.Path, err)
		}
		return nil
	}

	for _, item := range []struct {
		changes []refreshcore.FileChange
		event   string
	}{
		{req.Diff.Added, "added"},
		{req.Diff.Changed, "changed"},
		{req.Diff.Deleted, "deleted"},
		{req.Diff.Moved, "moved"},
		{req.Diff.NewlyIgnored, "newly_ignored"},
		{req.Diff.Reincluded, "reincluded"},
	} {
		for _, change := range item.changes {
			if err := writeEvent(change, item.event); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildRefreshMetadataJSON(req ApplyRefreshPlanRequest) string {
	payload := map[string]any{
		"reason":               req.Reason,
		"added":                len(req.Diff.Added),
		"changed":              len(req.Diff.Changed),
		"deleted":              len(req.Diff.Deleted),
		"moved":                len(req.Diff.Moved),
		"newly_ignored":        len(req.Diff.NewlyIgnored),
		"reincluded":           len(req.Diff.Reincluded),
		"affected_directories": req.AffectedPaths,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func eventMetadata(change refreshcore.FileChange) string {
	if change.PreviousPath == "" && change.PreviousDirectory == "" {
		return ""
	}
	payload := map[string]string{}
	if change.PreviousDirectory != "" {
		payload["previous_directory"] = change.PreviousDirectory
	}
	if change.DirectoryPath != "" {
		payload["directory"] = change.DirectoryPath
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func maybeInjectRefreshFailure(req ApplyRefreshPlanRequest, stage string) error {
	if req.InjectFailure == nil {
		return nil
	}
	return req.InjectFailure(stage)
}

type directoryMetrics struct {
	Fingerprint            string
	IncludedFileCount      int64
	IncludedDirectoryCount int64
	TotalSizeBytes         int64
}

func computeDirectoryMetrics(result repository.DiscoveryResult, fingerprints map[string]string) map[string]directoryMetrics {
	metrics := make(map[string]directoryMetrics, len(result.Directories))
	children := make(map[string][]string, len(result.Directories))
	filesByDirectory := make(map[string][]repository.FileRecord, len(result.Directories))
	order := make([]string, 0, len(result.Directories))

	for _, directory := range result.Directories {
		path := normalizeStoredDirectory(directory.Path)
		order = append(order, path)
		parent := normalizeParentPath(directory.ParentPath)
		children[parent] = append(children[parent], path)
	}
	for _, file := range result.Files {
		filesByDirectory[normalizeStoredDirectory(file.DirectoryPath)] = append(filesByDirectory[normalizeStoredDirectory(file.DirectoryPath)], file)
	}

	sort.Slice(order, func(i, j int) bool {
		return directoryDepth(order[i]) > directoryDepth(order[j])
	})

	for _, path := range order {
		metric := directoryMetrics{Fingerprint: fingerprints[path]}
		for _, file := range filesByDirectory[path] {
			if file.IgnoreStatus != repository.IgnoreStatusIncluded {
				continue
			}
			metric.IncludedFileCount++
			metric.TotalSizeBytes += file.SizeBytes
		}
		for _, child := range children[path] {
			if child == path {
				continue
			}
			childMetric := metrics[child]
			metric.IncludedFileCount += childMetric.IncludedFileCount
			metric.IncludedDirectoryCount += childMetric.IncludedDirectoryCount
			metric.TotalSizeBytes += childMetric.TotalSizeBytes
			if child != "." {
				childRecord := findDirectory(result.Directories, child)
				if childRecord.IgnoreStatus == repository.IgnoreStatusIncluded {
					metric.IncludedDirectoryCount++
				}
			}
		}
		metrics[path] = metric
	}

	return metrics
}

func findDirectory(directories []repository.DirectoryRecord, path string) repository.DirectoryRecord {
	for _, directory := range directories {
		if normalizeStoredDirectory(directory.Path) == path {
			return directory
		}
	}
	return repository.DirectoryRecord{}
}

func currentFilesByPath(files []repository.FileRecord) map[string]repository.FileRecord {
	index := make(map[string]repository.FileRecord, len(files))
	for _, file := range files {
		index[file.Path] = file
	}
	return index
}

func currentDirectoriesByPath(directories []repository.DirectoryRecord) map[string]repository.DirectoryRecord {
	index := make(map[string]repository.DirectoryRecord, len(directories))
	for _, directory := range directories {
		index[normalizeStoredDirectory(directory.Path)] = directory
	}
	return index
}

func persistedDirectoriesByPath(directories []repository.DirectorySnapshotRecord) map[string]repository.DirectorySnapshotRecord {
	index := make(map[string]repository.DirectorySnapshotRecord, len(directories))
	for _, directory := range directories {
		index[normalizeStoredDirectory(directory.Path)] = directory
	}
	return index
}

func normalizeStoredDirectory(path string) string {
	if path == "" || path == "." {
		return "."
	}
	return filepath.ToSlash(path)
}

func normalizeParentPath(path string) string {
	if path == "" {
		return ""
	}
	return normalizeStoredDirectory(path)
}

func emptyDirectoryParentPath(path string) any {
	if path == "" {
		return nil
	}
	return normalizeStoredDirectory(path)
}

func optionalContentHashForFile(file repository.FileRecord) any {
	if file.IgnoreStatus != repository.IgnoreStatusIncluded || file.ContentHash == "" {
		return nil
	}
	return file.ContentHash
}

func optionalIgnoreReasonForFile(file repository.FileRecord) any {
	if file.IgnoreStatus == repository.IgnoreStatusIncluded || file.IgnoreReason == "" {
		return nil
	}
	return string(file.IgnoreReason)
}

func optionalDirectoryIgnoreReason(directory repository.DirectoryRecord) any {
	if directory.IgnoreStatus == repository.IgnoreStatusIncluded || directory.IgnoreReason == "" {
		return nil
	}
	return string(directory.IgnoreReason)
}

func directoryDepth(path string) int {
	if path == "." {
		return 0
	}
	depth := 1
	for _, char := range path {
		if char == '/' {
			depth++
		}
	}
	return depth
}

func readRepositoryFreshnessTx(ctx context.Context, tx *sql.Tx, repositoryID int64) (repository.RepositoryFreshness, error) {
	var record repository.RepositoryFreshness
	var gitCommonDir, gitHeadRef, gitHeadCommit sql.NullString
	var lastRefreshStartedAt, lastRefreshCompletedAt sql.NullString
	var lastRefreshReason, freshnessReason sql.NullString
	var lastRefreshStatus, freshnessStatus string

	err := tx.QueryRowContext(ctx, `
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

func writeRepositoryFreshnessTx(ctx context.Context, tx *sql.Tx, record repository.RepositoryFreshness) error {
	_, err := tx.ExecContext(ctx, `
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

func createRefreshRunTx(ctx context.Context, tx *sql.Tx, run repository.RefreshRunRecord) (repository.RefreshRunRecord, error) {
	result, err := tx.ExecContext(ctx, `
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
	run.ID, err = result.LastInsertId()
	if err != nil {
		return repository.RefreshRunRecord{}, fmt.Errorf("load refresh run ID: %w", err)
	}
	return run, nil
}

func updateRefreshRunTx(ctx context.Context, tx *sql.Tx, run repository.RefreshRunRecord) error {
	_, err := tx.ExecContext(ctx, `
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
