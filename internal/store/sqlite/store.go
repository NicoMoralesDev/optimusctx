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

const (
	defaultLayeredContextLanguageLimit  = 5
	defaultLayeredContextMajorAreaLimit = 5
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

func (s *Store) LookupRepositoryID(ctx context.Context, rootPath string) (int64, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("lookup repository ID: store is not initialized")
	}
	if rootPath == "" {
		return 0, fmt.Errorf("lookup repository ID: root path is required")
	}

	var repositoryID int64
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM repositories WHERE root_path = ?`, rootPath).Scan(&repositoryID); err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("lookup repository ID for %q: %w", rootPath, err)
		}
		return 0, fmt.Errorf("lookup repository ID for %q: %w", rootPath, err)
	}

	return repositoryID, nil
}

func (s *Store) UpsertRepository(ctx context.Context, root repository.RepositoryRoot, now time.Time) (RepositoryRecord, error) {
	if s == nil || s.db == nil {
		return RepositoryRecord{}, fmt.Errorf("upsert repository: store is not initialized")
	}
	return s.upsertRepository(ctx, s.db, root, now)
}

type repositoryExecer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (s *Store) upsertRepository(ctx context.Context, execer repositoryExecer, root repository.RepositoryRoot, now time.Time) (RepositoryRecord, error) {
	if root.RootPath == "" {
		return RepositoryRecord{}, fmt.Errorf("upsert repository: root path is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	timestamp := now.UTC().Format(time.RFC3339)
	if _, err := execer.ExecContext(
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
	if err := execer.QueryRowContext(ctx, `SELECT id FROM repositories WHERE root_path = ?`, root.RootPath).Scan(&record.ID); err != nil {
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

func (s *Store) ListExtractionCandidates(ctx context.Context, repositoryID int64) ([]repository.ExtractionCandidate, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("list extraction candidates: store is not initialized")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id,
			path,
			language,
			content_hash,
			last_seen_generation,
			refresh_run_id
		FROM files
		WHERE repository_id = ? AND ignore_status = ?
		ORDER BY path
	`, repositoryID, string(repository.IgnoreStatusIncluded))
	if err != nil {
		return nil, fmt.Errorf("list extraction candidates for repository %d: %w", repositoryID, err)
	}
	defer rows.Close()

	var candidates []repository.ExtractionCandidate
	for rows.Next() {
		var candidate repository.ExtractionCandidate
		var language, contentHash sql.NullString
		var refreshRunID sql.NullInt64
		if err := rows.Scan(
			&candidate.FileID,
			&candidate.Path,
			&language,
			&contentHash,
			&candidate.SourceGeneration,
			&refreshRunID,
		); err != nil {
			return nil, fmt.Errorf("scan extraction candidate for repository %d: %w", repositoryID, err)
		}

		candidate.RepositoryID = repositoryID
		candidate.Language = language.String
		candidate.ContentHash = contentHash.String
		candidate.RefreshRunID = refreshRunID.Int64
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate extraction candidates for repository %d: %w", repositoryID, err)
	}

	return candidates, nil
}

func (s *Store) ReplaceFileArtifacts(ctx context.Context, artifacts repository.FileStructuralArtifacts) (repository.FileExtractionRecord, error) {
	if s == nil || s.db == nil {
		return repository.FileExtractionRecord{}, fmt.Errorf("replace file artifacts: store is not initialized")
	}
	if artifacts.Extraction.RepositoryID == 0 {
		return repository.FileExtractionRecord{}, fmt.Errorf("replace file artifacts: repository ID is required")
	}
	if artifacts.Extraction.FileID == 0 {
		return repository.FileExtractionRecord{}, fmt.Errorf("replace file artifacts: file ID is required")
	}
	if artifacts.Extraction.Path == "" {
		return repository.FileExtractionRecord{}, fmt.Errorf("replace file artifacts: path is required")
	}
	if artifacts.Extraction.Language == "" {
		return repository.FileExtractionRecord{}, fmt.Errorf("replace file artifacts: language is required")
	}
	if artifacts.Extraction.AdapterName == "" {
		return repository.FileExtractionRecord{}, fmt.Errorf("replace file artifacts: adapter name is required")
	}
	if artifacts.Extraction.GrammarVersion == "" {
		return repository.FileExtractionRecord{}, fmt.Errorf("replace file artifacts: grammar version is required")
	}
	if artifacts.Extraction.SourceContentHash == "" {
		return repository.FileExtractionRecord{}, fmt.Errorf("replace file artifacts: source content hash is required")
	}
	if artifacts.Extraction.SourceGeneration == 0 {
		return repository.FileExtractionRecord{}, fmt.Errorf("replace file artifacts: source generation is required")
	}
	if artifacts.Extraction.CoverageState == "" {
		return repository.FileExtractionRecord{}, fmt.Errorf("replace file artifacts: coverage state is required")
	}
	if artifacts.Extraction.ExtractedAt.IsZero() {
		artifacts.Extraction.ExtractedAt = time.Now().UTC()
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return repository.FileExtractionRecord{}, fmt.Errorf("begin replace file artifacts transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM symbols WHERE file_id = ?`, artifacts.Extraction.FileID); err != nil {
		return repository.FileExtractionRecord{}, fmt.Errorf("delete stale symbols for file %d: %w", artifacts.Extraction.FileID, err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO file_extractions (
			repository_id,
			file_id,
			path,
			language,
			adapter_name,
			grammar_version,
			source_content_hash,
			source_generation,
			coverage_state,
			coverage_reason,
			parser_error_count,
			has_error_nodes,
			symbol_count,
			top_level_symbol_count,
			max_symbol_depth,
			extracted_at,
			refresh_run_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(file_id) DO UPDATE SET
			repository_id = excluded.repository_id,
			path = excluded.path,
			language = excluded.language,
			adapter_name = excluded.adapter_name,
			grammar_version = excluded.grammar_version,
			source_content_hash = excluded.source_content_hash,
			source_generation = excluded.source_generation,
			coverage_state = excluded.coverage_state,
			coverage_reason = excluded.coverage_reason,
			parser_error_count = excluded.parser_error_count,
			has_error_nodes = excluded.has_error_nodes,
			symbol_count = excluded.symbol_count,
			top_level_symbol_count = excluded.top_level_symbol_count,
			max_symbol_depth = excluded.max_symbol_depth,
			extracted_at = excluded.extracted_at,
			refresh_run_id = excluded.refresh_run_id
	`,
		artifacts.Extraction.RepositoryID,
		artifacts.Extraction.FileID,
		artifacts.Extraction.Path,
		artifacts.Extraction.Language,
		artifacts.Extraction.AdapterName,
		artifacts.Extraction.GrammarVersion,
		artifacts.Extraction.SourceContentHash,
		artifacts.Extraction.SourceGeneration,
		string(artifacts.Extraction.CoverageState),
		emptyToNil(string(artifacts.Extraction.CoverageReason)),
		artifacts.Extraction.ParserErrorCount,
		boolToInt(artifacts.Extraction.HasErrorNodes),
		artifacts.Extraction.SymbolCount,
		artifacts.Extraction.TopLevelSymbolCount,
		artifacts.Extraction.MaxSymbolDepth,
		artifacts.Extraction.ExtractedAt.UTC().Format(time.RFC3339),
		nullableInt64(artifacts.Extraction.RefreshRunID),
	)
	if err != nil {
		return repository.FileExtractionRecord{}, fmt.Errorf("upsert file extraction for file %d: %w", artifacts.Extraction.FileID, err)
	}

	if err = tx.QueryRowContext(ctx, `SELECT id FROM file_extractions WHERE file_id = ?`, artifacts.Extraction.FileID).Scan(&artifacts.Extraction.ID); err != nil {
		return repository.FileExtractionRecord{}, fmt.Errorf("load file extraction for file %d: %w", artifacts.Extraction.FileID, err)
	}

	insertedSymbolIDs := make(map[string]int64, len(artifacts.Symbols))
	for i := range artifacts.Symbols {
		symbol := artifacts.Symbols[i]
		symbol.RepositoryID = artifacts.Extraction.RepositoryID
		symbol.FileID = artifacts.Extraction.FileID
		symbol.FileExtractionID = artifacts.Extraction.ID

		result, execErr := tx.ExecContext(ctx, `
			INSERT INTO symbols (
				repository_id,
				file_id,
				file_extraction_id,
				stable_key,
				parent_symbol_id,
				path,
				language,
				kind,
				name,
				qualified_name,
				ordinal,
				depth,
				start_byte,
				end_byte,
				start_row,
				start_column,
				end_row,
				end_column,
				name_start_byte,
				name_end_byte,
				signature_start_byte,
				signature_end_byte,
				is_exported
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			symbol.RepositoryID,
			symbol.FileID,
			symbol.FileExtractionID,
			symbol.StableKey,
			nil,
			symbol.Path,
			symbol.Language,
			symbol.Kind,
			symbol.Name,
			emptyToNil(symbol.QualifiedName),
			symbol.Ordinal,
			symbol.Depth,
			symbol.StartByte,
			symbol.EndByte,
			symbol.StartRow,
			symbol.StartColumn,
			symbol.EndRow,
			symbol.EndColumn,
			nullableInt64(symbol.NameStartByte),
			nullableInt64(symbol.NameEndByte),
			nullableInt64(symbol.SignatureStartByte),
			nullableInt64(symbol.SignatureEndByte),
			boolToInt(symbol.IsExported),
		)
		if execErr != nil {
			err = fmt.Errorf("insert symbol %q for file %d: %w", symbol.StableKey, artifacts.Extraction.FileID, execErr)
			return repository.FileExtractionRecord{}, err
		}

		symbolID, idErr := result.LastInsertId()
		if idErr != nil {
			err = fmt.Errorf("load inserted symbol ID for file %d: %w", artifacts.Extraction.FileID, idErr)
			return repository.FileExtractionRecord{}, err
		}

		symbol.ID = symbolID
		insertedSymbolIDs[symbol.StableKey] = symbolID
		artifacts.Symbols[i] = symbol
	}

	for _, symbol := range artifacts.Symbols {
		if symbol.ParentStableKey == "" {
			continue
		}
		parentID, ok := insertedSymbolIDs[symbol.ParentStableKey]
		if !ok {
			return repository.FileExtractionRecord{}, fmt.Errorf("insert symbol %q for file %d: parent %q not found", symbol.StableKey, artifacts.Extraction.FileID, symbol.ParentStableKey)
		}
		if _, err = tx.ExecContext(ctx, `UPDATE symbols SET parent_symbol_id = ? WHERE id = ?`, parentID, symbol.ID); err != nil {
			return repository.FileExtractionRecord{}, fmt.Errorf("update parent for symbol %q: %w", symbol.StableKey, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return repository.FileExtractionRecord{}, fmt.Errorf("commit replace file artifacts transaction: %w", err)
	}

	return artifacts.Extraction, nil
}

func (s *Store) DeleteFileArtifacts(ctx context.Context, fileID int64) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("delete file artifacts: store is not initialized")
	}
	if fileID == 0 {
		return fmt.Errorf("delete file artifacts: file ID is required")
	}

	if _, err := s.db.ExecContext(ctx, `DELETE FROM file_extractions WHERE file_id = ?`, fileID); err != nil {
		return fmt.Errorf("delete file artifacts for file %d: %w", fileID, err)
	}
	return nil
}

func (s *Store) ListFileExtractions(ctx context.Context, repositoryID int64) ([]repository.FileExtractionRecord, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("list file extractions: store is not initialized")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id,
			repository_id,
			file_id,
			path,
			language,
			adapter_name,
			grammar_version,
			source_content_hash,
			source_generation,
			coverage_state,
			coverage_reason,
			parser_error_count,
			has_error_nodes,
			symbol_count,
			top_level_symbol_count,
			max_symbol_depth,
			extracted_at,
			refresh_run_id
		FROM file_extractions
		WHERE repository_id = ?
		ORDER BY path
	`, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("list file extractions for repository %d: %w", repositoryID, err)
	}
	defer rows.Close()

	var records []repository.FileExtractionRecord
	for rows.Next() {
		record, scanErr := scanFileExtraction(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan file extraction for repository %d: %w", repositoryID, scanErr)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate file extractions for repository %d: %w", repositoryID, err)
	}

	return records, nil
}

func (s *Store) ListSymbols(ctx context.Context, repositoryID int64) ([]repository.SymbolRecord, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("list symbols: store is not initialized")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id,
			repository_id,
			file_id,
			file_extraction_id,
			stable_key,
			parent_symbol_id,
			path,
			language,
			kind,
			name,
			qualified_name,
			ordinal,
			depth,
			start_byte,
			end_byte,
			start_row,
			start_column,
			end_row,
			end_column,
			name_start_byte,
			name_end_byte,
			signature_start_byte,
			signature_end_byte,
			is_exported
		FROM symbols
		WHERE repository_id = ?
		ORDER BY path, ordinal, id
	`, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("list symbols for repository %d: %w", repositoryID, err)
	}
	defer rows.Close()

	var records []repository.SymbolRecord
	for rows.Next() {
		record, scanErr := scanSymbol(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan symbol for repository %d: %w", repositoryID, scanErr)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate symbols for repository %d: %w", repositoryID, err)
	}

	return records, nil
}

func (s *Store) ReadRepositoryStructuralCoverage(ctx context.Context, repositoryID int64) (repository.RepositoryStructuralCoverageSummary, error) {
	if s == nil || s.db == nil {
		return repository.RepositoryStructuralCoverageSummary{}, fmt.Errorf("read repository structural coverage: store is not initialized")
	}

	var summary repository.RepositoryStructuralCoverageSummary
	summary.RepositoryID = repositoryID

	err := s.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN ignore_status = ? THEN 1 ELSE 0 END), 0),
			COUNT(fe.id),
			COALESCE(SUM(CASE WHEN fe.coverage_state = 'supported' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN fe.coverage_state = 'partial' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN fe.coverage_state = 'unsupported' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN fe.coverage_state = 'failed' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN fe.coverage_state = 'skipped' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN fe.coverage_state IN ('partial', 'unsupported', 'failed', 'skipped') THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(fe.symbol_count), 0)
		FROM files f
		LEFT JOIN file_extractions fe ON fe.file_id = f.id
		WHERE f.repository_id = ?
	`, string(repository.IgnoreStatusIncluded), repositoryID).Scan(
		&summary.IncludedFileCount,
		&summary.ExtractionCount,
		&summary.SupportedCount,
		&summary.PartialCount,
		&summary.UnsupportedCount,
		&summary.FailedCount,
		&summary.SkippedCount,
		&summary.FilesWithCoverageGap,
		&summary.TotalSymbolCount,
	)
	if err != nil {
		return repository.RepositoryStructuralCoverageSummary{}, fmt.Errorf("read repository structural coverage %d: %w", repositoryID, err)
	}

	return summary, nil
}

func (s *Store) LoadRepositoryMapRecords(ctx context.Context, repositoryID int64) ([]repository.RepositoryMapFileRecord, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("load repository map records: store is not initialized")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			f.id,
			f.path,
			f.directory_path,
			f.language,
			f.ignore_status,
			fe.coverage_state,
			fe.coverage_reason,
			fe.symbol_count,
			fe.top_level_symbol_count,
			fe.max_symbol_depth,
			fe.source_generation
		FROM files f
		LEFT JOIN file_extractions fe ON fe.file_id = f.id
		WHERE f.repository_id = ? AND f.ignore_status = ?
		ORDER BY f.path
	`, repositoryID, string(repository.IgnoreStatusIncluded))
	if err != nil {
		return nil, fmt.Errorf("load repository map records for repository %d: %w", repositoryID, err)
	}
	defer rows.Close()

	symbols, err := s.ListSymbols(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	symbolsByFileID := make(map[int64][]repository.SymbolRecord)
	for _, symbol := range symbols {
		if symbol.Depth != 0 {
			continue
		}
		symbolsByFileID[symbol.FileID] = append(symbolsByFileID[symbol.FileID], symbol)
	}

	var records []repository.RepositoryMapFileRecord
	for rows.Next() {
		var record repository.RepositoryMapFileRecord
		var language, coverageState, coverageReason sql.NullString
		var symbolCount, topLevelSymbolCount, maxSymbolDepth, sourceGeneration sql.NullInt64
		var ignoreStatus string
		if err := rows.Scan(
			&record.FileID,
			&record.Path,
			&record.DirectoryPath,
			&language,
			&ignoreStatus,
			&coverageState,
			&coverageReason,
			&symbolCount,
			&topLevelSymbolCount,
			&maxSymbolDepth,
			&sourceGeneration,
		); err != nil {
			return nil, fmt.Errorf("scan repository map record for repository %d: %w", repositoryID, err)
		}

		record.DirectoryPath = normalizeStoredDirectoryPath(record.DirectoryPath)
		record.Language = language.String
		record.IgnoreStatus = repository.IgnoreStatus(ignoreStatus)
		record.CoverageState = repository.ExtractionCoverageState(coverageState.String)
		record.CoverageReason = repository.ExtractionCoverageReason(coverageReason.String)
		record.SymbolCount = symbolCount.Int64
		record.TopLevelSymbolCount = topLevelSymbolCount.Int64
		record.MaxSymbolDepth = maxSymbolDepth.Int64
		record.SourceGeneration = sourceGeneration.Int64
		if record.CoverageState == "" {
			record.CoverageState = repository.ExtractionCoverageStateSkipped
		}
		if record.CoverageState == repository.ExtractionCoverageStateSupported || record.CoverageState == repository.ExtractionCoverageStatePartial {
			record.Symbols = symbolsByFileID[record.FileID]
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate repository map records for repository %d: %w", repositoryID, err)
	}

	return records, nil
}

func (s *Store) LoadRepositoryMapDirectories(ctx context.Context, repositoryID int64) ([]repository.RepositoryMapDirectoryRecord, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("load repository map directories: store is not initialized")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			path,
			parent_path,
			included_file_count,
			included_directory_count,
			total_size_bytes,
			last_refresh_generation
		FROM directories
		WHERE repository_id = ? AND ignore_status = ?
		ORDER BY path
	`, repositoryID, string(repository.IgnoreStatusIncluded))
	if err != nil {
		return nil, fmt.Errorf("load repository map directories for repository %d: %w", repositoryID, err)
	}
	defer rows.Close()

	var records []repository.RepositoryMapDirectoryRecord
	for rows.Next() {
		var record repository.RepositoryMapDirectoryRecord
		var parentPath sql.NullString
		if err := rows.Scan(
			&record.Path,
			&parentPath,
			&record.IncludedFileCount,
			&record.IncludedDirectoryCount,
			&record.TotalSizeBytes,
			&record.LastRefreshGeneration,
		); err != nil {
			return nil, fmt.Errorf("scan repository map directory for repository %d: %w", repositoryID, err)
		}

		record.Path = normalizeStoredDirectoryPath(record.Path)
		record.ParentPath = normalizeNullablePath(parentPath)
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate repository map directories for repository %d: %w", repositoryID, err)
	}

	return records, nil
}

func (s *Store) ReadLayeredContextL0(ctx context.Context, repositoryID int64) (repository.LayeredContextL0, error) {
	if s == nil || s.db == nil {
		return repository.LayeredContextL0{}, fmt.Errorf("read layered context l0: store is not initialized")
	}

	freshness, err := s.ReadRepositoryFreshness(ctx, repositoryID)
	if err != nil {
		return repository.LayeredContextL0{}, fmt.Errorf("read layered context l0 freshness: %w", err)
	}

	result := repository.LayeredContextL0{
		Repository: repository.LayeredContextEnvelope{
			RepositoryRoot: freshness.RootPath,
			Generation:     freshness.LastRefreshGeneration,
			Freshness:      freshness.FreshnessStatus,
		},
		Identity: repository.LayeredContextRepositoryIdentity{
			RootPath:      freshness.RootPath,
			DetectionMode: freshness.DetectionMode,
			GitHeadRef:    freshness.GitHeadRef,
			GitHeadCommit: freshness.GitHeadCommit,
		},
	}

	languages, err := s.loadLayeredContextLanguages(ctx, repositoryID, defaultLayeredContextLanguageLimit)
	if err != nil {
		return repository.LayeredContextL0{}, err
	}
	result.Languages = languages

	majorAreas, err := s.loadLayeredContextMajorAreas(ctx, repositoryID, defaultLayeredContextMajorAreaLimit)
	if err != nil {
		return repository.LayeredContextL0{}, err
	}
	result.MajorAreas = majorAreas

	return result, nil
}

func (s *Store) loadLayeredContextLanguages(ctx context.Context, repositoryID int64, limit int) ([]repository.LayeredContextLanguageSummary, error) {
	if limit <= 0 {
		limit = defaultLayeredContextLanguageLimit
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			COALESCE(NULLIF(language, ''), 'unknown') AS language,
			COUNT(*) AS file_count,
			COALESCE(SUM(size_bytes), 0) AS total_size_bytes
		FROM files
		WHERE repository_id = ? AND ignore_status = ?
		GROUP BY COALESCE(NULLIF(language, ''), 'unknown')
		ORDER BY file_count DESC, total_size_bytes DESC, language ASC
		LIMIT ?
	`, repositoryID, string(repository.IgnoreStatusIncluded), limit)
	if err != nil {
		return nil, fmt.Errorf("load layered context l0 languages for repository %d: %w", repositoryID, err)
	}
	defer rows.Close()

	var summaries []repository.LayeredContextLanguageSummary
	for rows.Next() {
		var summary repository.LayeredContextLanguageSummary
		if err := rows.Scan(&summary.Language, &summary.FileCount, &summary.TotalSizeBytes); err != nil {
			return nil, fmt.Errorf("scan layered context l0 language for repository %d: %w", repositoryID, err)
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate layered context l0 languages for repository %d: %w", repositoryID, err)
	}

	return summaries, nil
}

func (s *Store) loadLayeredContextMajorAreas(ctx context.Context, repositoryID int64, limit int) ([]repository.LayeredContextMajorAreaSummary, error) {
	if limit <= 0 {
		limit = defaultLayeredContextMajorAreaLimit
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT path, kind, included_file_count, total_size_bytes
		FROM (
			SELECT
				d.path AS path,
				'directory' AS kind,
				d.included_file_count,
				d.total_size_bytes
			FROM directories d
			WHERE d.repository_id = ? AND d.ignore_status = ? AND d.parent_path = '.'
			UNION ALL
			SELECT
				'.' AS path,
				'root_files' AS kind,
				COUNT(*) AS included_file_count,
				COALESCE(SUM(size_bytes), 0) AS total_size_bytes
			FROM files
			WHERE repository_id = ? AND ignore_status = ? AND directory_path = '.'
		)
		WHERE included_file_count > 0
		ORDER BY total_size_bytes DESC, included_file_count DESC, path ASC
		LIMIT ?
	`, repositoryID, string(repository.IgnoreStatusIncluded), repositoryID, string(repository.IgnoreStatusIncluded), limit)
	if err != nil {
		return nil, fmt.Errorf("load layered context l0 major areas for repository %d: %w", repositoryID, err)
	}
	defer rows.Close()

	var summaries []repository.LayeredContextMajorAreaSummary
	for rows.Next() {
		var summary repository.LayeredContextMajorAreaSummary
		var kind string
		if err := rows.Scan(&summary.Path, &kind, &summary.IncludedFileCount, &summary.TotalSizeBytes); err != nil {
			return nil, fmt.Errorf("scan layered context l0 major area for repository %d: %w", repositoryID, err)
		}
		summary.Path = normalizeStoredDirectoryPath(summary.Path)
		summary.Kind = repository.MajorAreaKind(kind)
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate layered context l0 major areas for repository %d: %w", repositoryID, err)
	}

	return summaries, nil
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

func scanFileExtraction(scanner interface {
	Scan(dest ...any) error
}) (repository.FileExtractionRecord, error) {
	var record repository.FileExtractionRecord
	var coverageReason, extractedAt sql.NullString
	var refreshRunID sql.NullInt64
	var coverageState string
	var hasErrorNodes int64

	err := scanner.Scan(
		&record.ID,
		&record.RepositoryID,
		&record.FileID,
		&record.Path,
		&record.Language,
		&record.AdapterName,
		&record.GrammarVersion,
		&record.SourceContentHash,
		&record.SourceGeneration,
		&coverageState,
		&coverageReason,
		&record.ParserErrorCount,
		&hasErrorNodes,
		&record.SymbolCount,
		&record.TopLevelSymbolCount,
		&record.MaxSymbolDepth,
		&extractedAt,
		&refreshRunID,
	)
	if err != nil {
		return repository.FileExtractionRecord{}, err
	}

	record.CoverageState = repository.ExtractionCoverageState(coverageState)
	record.CoverageReason = repository.ExtractionCoverageReason(coverageReason.String)
	record.HasErrorNodes = hasErrorNodes != 0
	record.ExtractedAt = parseOptionalRFC3339(extractedAt)
	record.RefreshRunID = refreshRunID.Int64

	return record, nil
}

func scanSymbol(scanner interface {
	Scan(dest ...any) error
}) (repository.SymbolRecord, error) {
	var record repository.SymbolRecord
	var parentSymbolID sql.NullInt64
	var qualifiedName sql.NullString
	var nameStartByte, nameEndByte, signatureStartByte, signatureEndByte sql.NullInt64
	var isExported int64

	err := scanner.Scan(
		&record.ID,
		&record.RepositoryID,
		&record.FileID,
		&record.FileExtractionID,
		&record.StableKey,
		&parentSymbolID,
		&record.Path,
		&record.Language,
		&record.Kind,
		&record.Name,
		&qualifiedName,
		&record.Ordinal,
		&record.Depth,
		&record.StartByte,
		&record.EndByte,
		&record.StartRow,
		&record.StartColumn,
		&record.EndRow,
		&record.EndColumn,
		&nameStartByte,
		&nameEndByte,
		&signatureStartByte,
		&signatureEndByte,
		&isExported,
	)
	if err != nil {
		return repository.SymbolRecord{}, err
	}

	record.ParentSymbolID = parentSymbolID.Int64
	record.QualifiedName = qualifiedName.String
	record.NameStartByte = nameStartByte.Int64
	record.NameEndByte = nameEndByte.Int64
	record.SignatureStartByte = signatureStartByte.Int64
	record.SignatureEndByte = signatureEndByte.Int64
	record.IsExported = isExported != 0

	return record, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func nullableInt64(value int64) any {
	if value == 0 {
		return nil
	}
	return value
}
