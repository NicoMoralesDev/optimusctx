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

type RepositoryContextService struct {
	Locator       repository.Locator
	OpenStore     func(context.Context, state.Layout, string) (*sqlite.Store, error)
	ResolveLayout func(string) (state.Layout, error)
}

func NewRepositoryContextService() RepositoryContextService {
	return RepositoryContextService{
		Locator: repository.NewLocator(),
		OpenStore: func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		},
		ResolveLayout: state.ResolveLayout,
	}
}

func (s RepositoryContextService) LayeredContextL0(ctx context.Context, startPath string) (repository.LayeredContextL0, error) {
	store, repositoryID, err := s.openRepositoryContextStore(ctx, startPath)
	if err != nil {
		return repository.LayeredContextL0{}, err
	}
	defer store.Close()

	result, err := store.ReadLayeredContextL0(ctx, repositoryID)
	if err != nil {
		return repository.LayeredContextL0{}, fmt.Errorf("load layered context l0: %w", err)
	}

	return result, nil
}

func (s RepositoryContextService) LayeredContextL1(ctx context.Context, startPath string) (repository.LayeredContextL1, error) {
	store, repositoryID, err := s.openRepositoryContextStore(ctx, startPath)
	if err != nil {
		return repository.LayeredContextL1{}, err
	}
	defer store.Close()

	result, err := store.ReadLayeredContextL1(ctx, repositoryID)
	if err != nil {
		return repository.LayeredContextL1{}, fmt.Errorf("load layered context l1: %w", err)
	}

	return result, nil
}

func (s RepositoryContextService) openRepositoryContextStore(ctx context.Context, startPath string) (*sqlite.Store, int64, error) {
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
