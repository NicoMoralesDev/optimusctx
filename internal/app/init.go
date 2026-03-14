package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

type InitService struct {
	Locator       repository.Locator
	Refresh       func(context.Context, RefreshRequest) (RefreshResult, error)
	OpenStore     func(context.Context, state.Layout, string) (*sqlite.Store, error)
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
		Refresh: func(ctx context.Context, request RefreshRequest) (RefreshResult, error) {
			return NewRefreshService().Refresh(ctx, request)
		},
		OpenStore: func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
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

	refreshFn := s.Refresh
	if refreshFn == nil {
		refreshFn = func(ctx context.Context, request RefreshRequest) (RefreshResult, error) {
			return NewRefreshService().Refresh(ctx, request)
		}
	}
	refreshResult, err := refreshFn(ctx, RefreshRequest{
		StartPath: root.RootPath,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	})
	if err != nil {
		return InitResult{}, fmt.Errorf("refresh repository inventory: %w", err)
	}

	repoRecord, err := store.UpsertRepository(ctx, root, time.Now().UTC())
	if err != nil {
		return InitResult{}, fmt.Errorf("load repository metadata: %w", err)
	}
	snapshot, err := store.LoadRepositorySnapshot(ctx, repoRecord.ID)
	if err != nil {
		return InitResult{}, fmt.Errorf("load persisted snapshot: %w", err)
	}

	result := InitResult{
		RepositoryRoot: refreshResult.RepositoryRoot,
		StatePath:      refreshResult.StatePath,
		SchemaVersion:  refreshResult.SchemaVersion,
		FileCount:      len(snapshot.Files),
		DirectoryCount: len(snapshot.Directories),
	}
	for _, file := range snapshot.Files {
		if file.IgnoreStatus == repository.IgnoreStatusIncluded {
			result.IncludedFiles++
			continue
		}
		result.IgnoredFiles++
	}

	return result, nil
}
