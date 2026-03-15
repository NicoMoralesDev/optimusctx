package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
