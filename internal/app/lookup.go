package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

type LookupService struct {
	Locator       repository.Locator
	OpenStore     func(context.Context, state.Layout, string) (*sqlite.Store, error)
	ResolveLayout func(string) (state.Layout, error)
}

func NewLookupService() LookupService {
	return LookupService{
		Locator: repository.NewLocator(),
		OpenStore: func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		},
		ResolveLayout: state.ResolveLayout,
	}
}

func (s LookupService) SymbolLookup(ctx context.Context, startPath string, request repository.SymbolLookupRequest) (repository.SymbolLookupResult, error) {
	store, repositoryID, err := s.openLookupStore(ctx, startPath)
	if err != nil {
		return repository.SymbolLookupResult{}, err
	}
	defer store.Close()

	result, err := store.ReadSymbolLookup(ctx, repositoryID, request)
	if err != nil {
		return repository.SymbolLookupResult{}, fmt.Errorf("load symbol lookup: %w", err)
	}

	return result, nil
}

func (s LookupService) ResolveSymbolByStableKey(ctx context.Context, startPath string, stableKey string) (repository.SymbolLookupMatch, error) {
	store, repositoryID, err := s.openLookupStore(ctx, startPath)
	if err != nil {
		return repository.SymbolLookupMatch{}, err
	}
	defer store.Close()

	return s.resolveSymbolByStableKey(ctx, store, repositoryID, stableKey)
}

func (s LookupService) resolveSymbolByStableKey(ctx context.Context, store *sqlite.Store, repositoryID int64, stableKey string) (repository.SymbolLookupMatch, error) {
	if store == nil || store.DB() == nil {
		return repository.SymbolLookupMatch{}, fmt.Errorf("load symbol anchor: store is not initialized")
	}
	stableKey = strings.TrimSpace(stableKey)
	if stableKey == "" {
		return repository.SymbolLookupMatch{}, fmt.Errorf("load symbol anchor: stable key is required")
	}

	var match repository.SymbolLookupMatch
	err := store.DB().QueryRowContext(ctx, `
		SELECT
			stable_key,
			path,
			COALESCE(NULLIF(language, ''), 'unknown') AS language,
			kind,
			name,
			COALESCE(qualified_name, '') AS qualified_name,
			ordinal,
			start_row,
			start_column,
			end_row,
			end_column
		FROM symbols
		WHERE repository_id = ? AND stable_key = ?
		LIMIT 1
	`, repositoryID, stableKey).Scan(
		&match.StableKey,
		&match.Path,
		&match.Language,
		&match.Kind,
		&match.Name,
		&match.QualifiedName,
		&match.Ordinal,
		&match.StartRow,
		&match.StartColumn,
		&match.EndRow,
		&match.EndColumn,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.SymbolLookupMatch{}, fmt.Errorf("load symbol anchor: stable key %q not found", stableKey)
		}
		return repository.SymbolLookupMatch{}, fmt.Errorf("load symbol anchor: %w", err)
	}

	return match, nil
}

func (s LookupService) StructureLookup(ctx context.Context, startPath string, request repository.StructureLookupRequest) (repository.StructureLookupResult, error) {
	store, repositoryID, err := s.openLookupStore(ctx, startPath)
	if err != nil {
		return repository.StructureLookupResult{}, err
	}
	defer store.Close()

	result, err := store.ReadStructureLookup(ctx, repositoryID, request)
	if err != nil {
		return repository.StructureLookupResult{}, fmt.Errorf("load structure lookup: %w", err)
	}

	return result, nil
}

type symbolAnchor struct {
	Path     string
	StartRow int64
	EndRow   int64
}

func (s LookupService) loadSymbolAnchor(ctx context.Context, store *sqlite.Store, repositoryID int64, stableKey string) (symbolAnchor, error) {
	if stableKey == "" {
		return symbolAnchor{}, fmt.Errorf("load symbol anchor: stable key is required")
	}

	var anchor symbolAnchor
	if err := store.DB().QueryRowContext(ctx, `
		SELECT path, start_row, end_row
		FROM symbols
		WHERE repository_id = ? AND stable_key = ?
	`, repositoryID, stableKey).Scan(&anchor.Path, &anchor.StartRow, &anchor.EndRow); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return symbolAnchor{}, fmt.Errorf("load symbol anchor: stable key %q not found", stableKey)
		}
		return symbolAnchor{}, fmt.Errorf("load symbol anchor: %w", err)
	}

	return anchor, nil
}

func (s LookupService) openLookupStore(ctx context.Context, startPath string) (*sqlite.Store, int64, error) {
	root, err := s.Locator.Resolve(startPath)
	if err != nil {
		if errors.Is(err, repository.ErrRepositoryNotFound) {
			return nil, 0, fmt.Errorf("resolve repository root: %w", err)
		}
		return nil, 0, fmt.Errorf("resolve repository root: %w", err)
	}

	layoutResolver := s.ResolveLayout
	if layoutResolver == nil {
		layoutResolver = state.ResolveLayout
	}
	layout, err := layoutResolver(root.RootPath)
	if err != nil {
		return nil, 0, fmt.Errorf("resolve state layout: %w", err)
	}

	openStore := s.OpenStore
	if openStore == nil {
		openStore = func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		}
	}
	store, err := openStore(ctx, layout, root.DetectionMode)
	if err != nil {
		return nil, 0, fmt.Errorf("open state store: %w", err)
	}

	repositoryID, err := store.LookupRepositoryID(ctx, root.RootPath)
	if err != nil {
		_ = store.Close()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, fmt.Errorf("load repository metadata: %w", err)
		}
		return nil, 0, fmt.Errorf("load repository metadata: %w", err)
	}

	return store, repositoryID, nil
}
