package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

type InitService struct {
	Locator       repository.Locator
	OpenStore     func(context.Context, state.Layout, string) (*sqlite.Store, error)
	Discover      func(string) (repository.DiscoveryResult, error)
	ResolveLayout func(string) (state.Layout, error)
}

type InitResult struct {
	RepositoryRoot string
	StatePath      string
	SchemaVersion  int
	FileCount      int
	IncludedFiles  int
	IgnoredFiles   int
	DirectoryCount int
}

func NewInitService() InitService {
	return InitService{
		Locator: repository.NewLocator(),
		OpenStore: func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		},
		Discover: func(root string) (repository.DiscoveryResult, error) {
			return repository.NewDiscovery(root).Walk()
		},
		ResolveLayout: state.ResolveLayout,
	}
}

func (s InitService) Init(ctx context.Context, startPath string) (InitResult, error) {
	root, err := s.Locator.Resolve(startPath)
	if err != nil {
		if errors.Is(err, repository.ErrRepositoryNotFound) {
			return InitResult{}, fmt.Errorf("resolve repository root: %w", err)
		}
		return InitResult{}, fmt.Errorf("resolve repository root: %w", err)
	}

	layoutResolver := s.ResolveLayout
	if layoutResolver == nil {
		layoutResolver = state.ResolveLayout
	}

	layout, err := layoutResolver(root.RootPath)
	if err != nil {
		return InitResult{}, fmt.Errorf("resolve state layout: %w", err)
	}

	openStore := s.OpenStore
	if openStore == nil {
		openStore = func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		}
	}

	store, err := openStore(ctx, layout, root.DetectionMode)
	if err != nil {
		return InitResult{}, fmt.Errorf("open state store: %w", err)
	}
	defer store.Close()

	discover := s.Discover
	if discover == nil {
		discover = func(repoRoot string) (repository.DiscoveryResult, error) {
			return repository.NewDiscovery(repoRoot).Walk()
		}
	}

	discoveryResult, err := discover(root.RootPath)
	if err != nil {
		return InitResult{}, fmt.Errorf("discover repository inventory: %w", err)
	}

	if err := persistInventory(ctx, store, root, discoveryResult); err != nil {
		return InitResult{}, err
	}

	result := InitResult{
		RepositoryRoot: root.RootPath,
		StatePath:      layout.StateDir,
		SchemaVersion:  store.SchemaVersion(),
		FileCount:      len(discoveryResult.Files),
		DirectoryCount: len(discoveryResult.Directories),
	}

	for _, file := range discoveryResult.Files {
		if file.IgnoreStatus == repository.IgnoreStatusIncluded {
			result.IncludedFiles++
			continue
		}
		result.IgnoredFiles++
	}

	return result, nil
}

func persistInventory(ctx context.Context, store *sqlite.Store, root repository.RepositoryRoot, result repository.DiscoveryResult) error {
	tx, err := store.DB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin inventory transaction: %w", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	repoRecord, err := store.UpsertRepository(ctx, root, inventoryTimestamp(result))
	if err != nil {
		return fmt.Errorf("persist repository metadata: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM directories WHERE repository_id = ?`, repoRecord.ID); err != nil {
		return fmt.Errorf("clear persisted directories: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM files WHERE repository_id = ?`, repoRecord.ID); err != nil {
		return fmt.Errorf("clear persisted files: %w", err)
	}

	if err := insertDirectories(ctx, tx, repoRecord.ID, result.Directories); err != nil {
		return err
	}
	if err := insertFiles(ctx, tx, repoRecord.ID, result.Files); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit inventory transaction: %w", err)
	}
	tx = nil
	return nil
}

func inventoryTimestamp(result repository.DiscoveryResult) time.Time {
	for _, directory := range result.Directories {
		if directory.Path == "." && !directory.DiscoveredAt.IsZero() {
			return directory.DiscoveredAt
		}
	}
	for _, file := range result.Files {
		if !file.DiscoveredAt.IsZero() {
			return file.DiscoveredAt
		}
	}
	return time.Now().UTC()
}

func insertDirectories(ctx context.Context, tx *sql.Tx, repositoryID int64, directories []repository.DirectoryRecord) error {
	statement, err := tx.PrepareContext(ctx, `INSERT INTO directories (
		repository_id, path, parent_path, discovered_at, ignore_status, ignore_reason
	) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare directory insert: %w", err)
	}
	defer statement.Close()

	for _, directory := range directories {
		if _, err := statement.ExecContext(
			ctx,
			repositoryID,
			directory.Path,
			emptyDirectoryParent(directory.ParentPath),
			directory.DiscoveredAt.UTC().Format(time.RFC3339),
			string(directory.IgnoreStatus),
			optionalIgnoreReason(directory.IgnoreStatus, directory.IgnoreReason),
		); err != nil {
			return fmt.Errorf("insert directory %q: %w", directory.Path, err)
		}
	}

	return nil
}

func insertFiles(ctx context.Context, tx *sql.Tx, repositoryID int64, files []repository.FileRecord) error {
	statement, err := tx.PrepareContext(ctx, `INSERT INTO files (
		repository_id, path, directory_path, extension, language, size_bytes, content_hash, last_indexed_at,
		ignore_status, ignore_reason, fs_mod_time, discovered_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare file insert: %w", err)
	}
	defer statement.Close()

	for _, file := range files {
		discoveredAt := file.DiscoveredAt.UTC().Format(time.RFC3339)
		if _, err := statement.ExecContext(
			ctx,
			repositoryID,
			file.Path,
			normalizeDirectoryPath(file.DirectoryPath),
			emptyStringToNil(file.Extension),
			file.LanguageHint,
			file.SizeBytes,
			optionalContentHash(file.IgnoreStatus, file.ContentHash),
			optionalTime(file.LastIndexedAt),
			string(file.IgnoreStatus),
			optionalIgnoreReason(file.IgnoreStatus, file.IgnoreReason),
			optionalTime(file.FilesystemModTime),
			discoveredAt,
			discoveredAt,
		); err != nil {
			return fmt.Errorf("insert file %q: %w", file.Path, err)
		}
	}

	return nil
}

func emptyDirectoryParent(parent string) any {
	if parent == "" {
		return nil
	}
	return normalizeDirectoryPath(parent)
}

func normalizeDirectoryPath(path string) string {
	if path == "" {
		return "."
	}
	return filepath.ToSlash(path)
}

func emptyStringToNil(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func optionalContentHash(status repository.IgnoreStatus, hash string) any {
	if status != repository.IgnoreStatusIncluded || hash == "" {
		return nil
	}
	return hash
}

func optionalTime(timestamp time.Time) any {
	if timestamp.IsZero() {
		return nil
	}
	return timestamp.UTC().Format(time.RFC3339)
}

func optionalIgnoreReason(status repository.IgnoreStatus, reason repository.IgnoreReason) any {
	if status == repository.IgnoreStatusIncluded || reason == repository.IgnoreReasonNone {
		return nil
	}
	return string(reason)
}
