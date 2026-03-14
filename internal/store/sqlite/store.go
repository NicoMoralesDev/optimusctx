package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
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

func emptyToNil(value string) any {
	if value == "" {
		return nil
	}
	return value
}
