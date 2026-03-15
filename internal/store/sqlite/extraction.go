package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
)

type RefreshStructuralStore interface {
	ListExtractionCandidatesByPath(ctx context.Context, paths []string) ([]repository.ExtractionCandidate, error)
	ReplaceFileArtifacts(ctx context.Context, artifacts repository.FileStructuralArtifacts) (repository.FileExtractionRecord, error)
	DeleteFileArtifactsByPath(ctx context.Context, paths []string) error
}

type refreshStructuralTx struct {
	tx           *sql.Tx
	repositoryID int64
}

func (s *Store) newRefreshStructuralTx(tx *sql.Tx, repositoryID int64) RefreshStructuralStore {
	return refreshStructuralTx{
		tx:           tx,
		repositoryID: repositoryID,
	}
}

func (s refreshStructuralTx) ListExtractionCandidatesByPath(ctx context.Context, paths []string) ([]repository.ExtractionCandidate, error) {
	return listExtractionCandidatesByPath(ctx, s.tx, s.repositoryID, paths)
}

func (s refreshStructuralTx) ReplaceFileArtifacts(ctx context.Context, artifacts repository.FileStructuralArtifacts) (repository.FileExtractionRecord, error) {
	return replaceFileArtifacts(ctx, s.tx, artifacts)
}

func (s refreshStructuralTx) DeleteFileArtifactsByPath(ctx context.Context, paths []string) error {
	return deleteFileArtifactsByPath(ctx, s.tx, s.repositoryID, paths)
}

func listExtractionCandidatesByPath(ctx context.Context, queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, repositoryID int64, paths []string) ([]repository.ExtractionCandidate, error) {
	if repositoryID == 0 {
		return nil, fmt.Errorf("list extraction candidates by path: repository ID is required")
	}
	if len(paths) == 0 {
		return nil, nil
	}

	placeholders := make([]string, 0, len(paths))
	args := make([]any, 0, len(paths)+2)
	args = append(args, repositoryID, string(repository.IgnoreStatusIncluded))
	for _, path := range paths {
		placeholders = append(placeholders, "?")
		args = append(args, path)
	}

	rows, err := queryer.QueryContext(ctx, `
		SELECT
			id,
			path,
			language,
			content_hash,
			last_seen_generation,
			refresh_run_id
		FROM files
		WHERE repository_id = ? AND ignore_status = ? AND path IN (`+strings.Join(placeholders, ",")+`)
		ORDER BY path
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list extraction candidates by path for repository %d: %w", repositoryID, err)
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
			return nil, fmt.Errorf("scan extraction candidate by path for repository %d: %w", repositoryID, err)
		}
		candidate.RepositoryID = repositoryID
		candidate.Language = language.String
		candidate.ContentHash = contentHash.String
		candidate.RefreshRunID = refreshRunID.Int64
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate extraction candidates by path for repository %d: %w", repositoryID, err)
	}

	return candidates, nil
}

func replaceFileArtifacts(ctx context.Context, tx *sql.Tx, artifacts repository.FileStructuralArtifacts) (repository.FileExtractionRecord, error) {
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

	if _, err := tx.ExecContext(ctx, `DELETE FROM symbols WHERE file_id = ?`, artifacts.Extraction.FileID); err != nil {
		return repository.FileExtractionRecord{}, fmt.Errorf("delete stale symbols for file %d: %w", artifacts.Extraction.FileID, err)
	}

	if _, err := tx.ExecContext(ctx, `
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
	); err != nil {
		return repository.FileExtractionRecord{}, fmt.Errorf("upsert file extraction for file %d: %w", artifacts.Extraction.FileID, err)
	}

	if err := tx.QueryRowContext(ctx, `SELECT id FROM file_extractions WHERE file_id = ?`, artifacts.Extraction.FileID).Scan(&artifacts.Extraction.ID); err != nil {
		return repository.FileExtractionRecord{}, fmt.Errorf("load file extraction for file %d: %w", artifacts.Extraction.FileID, err)
	}

	insertedSymbolIDs := make(map[string]int64, len(artifacts.Symbols))
	for i := range artifacts.Symbols {
		symbol := artifacts.Symbols[i]
		symbol.RepositoryID = artifacts.Extraction.RepositoryID
		symbol.FileID = artifacts.Extraction.FileID
		symbol.FileExtractionID = artifacts.Extraction.ID

		result, err := tx.ExecContext(ctx, `
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
		if err != nil {
			return repository.FileExtractionRecord{}, fmt.Errorf("insert symbol %q for file %d: %w", symbol.StableKey, artifacts.Extraction.FileID, err)
		}

		symbolID, err := result.LastInsertId()
		if err != nil {
			return repository.FileExtractionRecord{}, fmt.Errorf("load inserted symbol ID for file %d: %w", artifacts.Extraction.FileID, err)
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
		if _, err := tx.ExecContext(ctx, `UPDATE symbols SET parent_symbol_id = ? WHERE id = ?`, parentID, symbol.ID); err != nil {
			return repository.FileExtractionRecord{}, fmt.Errorf("update parent for symbol %q: %w", symbol.StableKey, err)
		}
	}

	return artifacts.Extraction, nil
}

func deleteFileArtifactsByPath(ctx context.Context, execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, repositoryID int64, paths []string) error {
	if repositoryID == 0 {
		return fmt.Errorf("delete file artifacts by path: repository ID is required")
	}
	if len(paths) == 0 {
		return nil
	}

	placeholders := make([]string, 0, len(paths))
	args := make([]any, 0, len(paths)+1)
	args = append(args, repositoryID)
	for _, path := range paths {
		placeholders = append(placeholders, "?")
		args = append(args, path)
	}

	if _, err := execer.ExecContext(ctx, `
		DELETE FROM file_extractions
		WHERE repository_id = ? AND path IN (`+strings.Join(placeholders, ",")+`)
	`, args...); err != nil {
		return fmt.Errorf("delete file artifacts by path for repository %d: %w", repositoryID, err)
	}

	return nil
}
