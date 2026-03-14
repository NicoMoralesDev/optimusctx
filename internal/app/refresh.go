package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	refreshcore "github.com/niccrow/optimusctx/internal/refresh"
	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

type RefreshRequest struct {
	StartPath     string
	Reason        repository.RefreshReason
	ForceFull     bool
	ChangedHint   []string
	InjectFailure func(string) error
}

type RefreshResult struct {
	RepositoryRoot      string
	StatePath           string
	SchemaVersion       int
	Reason              repository.RefreshReason
	AddedFiles          int
	ChangedContentFiles int
	DeletedFiles        int
	MovedFiles          int
	NewlyIgnoredFiles   int
	ReincludedFiles     int
	ChangedFiles        int
	UnchangedFiles      int
	AffectedDirectories int
	Generation          int64
	FreshnessStatus     repository.FreshnessStatus
}

type RefreshService struct {
	Locator       repository.Locator
	OpenStore     func(context.Context, state.Layout, string) (*sqlite.Store, error)
	Discover      func(string, repository.RepositorySnapshot, bool) (repository.DiscoveryResult, error)
	ResolveLayout func(string) (state.Layout, error)
	Now           func() time.Time
}

func NewRefreshService() RefreshService {
	return RefreshService{
		Locator: repository.NewLocator(),
		OpenStore: func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		},
		Discover: func(root string, snapshot repository.RepositorySnapshot, forceFull bool) (repository.DiscoveryResult, error) {
			discovery := repository.NewDiscovery(root)
			if forceFull {
				return discovery.Walk()
			}
			return discovery.WalkWithPersistedSnapshot(snapshot)
		},
		ResolveLayout: state.ResolveLayout,
		Now:           time.Now,
	}
}

func (s RefreshService) Refresh(ctx context.Context, request RefreshRequest) (RefreshResult, error) {
	root, err := s.Locator.Resolve(request.StartPath)
	if err != nil {
		if errors.Is(err, repository.ErrRepositoryNotFound) {
			return RefreshResult{}, fmt.Errorf("resolve repository root: %w", err)
		}
		return RefreshResult{}, fmt.Errorf("resolve repository root: %w", err)
	}
	if request.Reason == "" {
		request.Reason = repository.RefreshReasonManual
	}

	layoutResolver := s.ResolveLayout
	if layoutResolver == nil {
		layoutResolver = state.ResolveLayout
	}
	layout, err := layoutResolver(root.RootPath)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("resolve state layout: %w", err)
	}

	openStore := s.OpenStore
	if openStore == nil {
		openStore = func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		}
	}
	store, err := openStore(ctx, layout, root.DetectionMode)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("open state store: %w", err)
	}
	defer store.Close()

	nowFn := s.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	startedAt := nowFn().UTC()

	repoRecord, err := store.UpsertRepository(ctx, root, startedAt)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("ensure repository metadata: %w", err)
	}

	persistedSnapshot, err := store.LoadRepositorySnapshot(ctx, repoRecord.ID)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("load persisted snapshot: %w", err)
	}

	discover := s.Discover
	if discover == nil {
		discover = func(root string, snapshot repository.RepositorySnapshot, forceFull bool) (repository.DiscoveryResult, error) {
			discovery := repository.NewDiscovery(root)
			if forceFull {
				return discovery.Walk()
			}
			return discovery.WalkWithPersistedSnapshot(snapshot)
		}
	}
	currentResult, err := discover(root.RootPath, persistedSnapshot, request.ForceFull)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("discover repository snapshot: %w", err)
	}

	currentSnapshot := refreshcore.CurrentSnapshot(currentResult)
	persistedRefreshSnapshot := refreshcore.PersistedSnapshot(persistedSnapshot)
	diff := refreshcore.DiffSnapshots(currentSnapshot, persistedRefreshSnapshot)
	affected := refreshcore.AffectedDirectories(currentSnapshot, persistedRefreshSnapshot, diff)
	fingerprints := refreshcore.ComputeSubtreeFingerprints(currentSnapshot, persistedRefreshSnapshot, affected)

	applyResult, err := store.ApplyRefreshPlan(ctx, sqlite.ApplyRefreshPlanRequest{
		RepositoryID:      repoRecord.ID,
		RepositoryRoot:    root,
		Reason:            request.Reason,
		StartedAt:         startedAt,
		CompletedAt:       nowFn().UTC(),
		CurrentResult:     currentResult,
		PersistedSnapshot: persistedSnapshot,
		Diff:              diff,
		AffectedPaths:     affected,
		Fingerprints:      fingerprints,
		InjectFailure:     request.InjectFailure,
	})
	if err != nil {
		failureResult, failureLoadErr := buildRefreshFailureResult(ctx, store, repoRecord.ID, layout.StateDir, root.RootPath, diff, affected, request)
		if failureLoadErr != nil {
			return RefreshResult{}, err
		}
		return failureResult, err
	}

	return RefreshResult{
		RepositoryRoot:      root.RootPath,
		StatePath:           layout.StateDir,
		SchemaVersion:       store.SchemaVersion(),
		Reason:              request.Reason,
		AddedFiles:          len(diff.Added),
		ChangedContentFiles: len(diff.Changed),
		DeletedFiles:        len(diff.Deleted),
		MovedFiles:          len(diff.Moved),
		NewlyIgnoredFiles:   len(diff.NewlyIgnored),
		ReincludedFiles:     len(diff.Reincluded),
		ChangedFiles:        changedFileCount(diff),
		UnchangedFiles:      len(diff.Unchanged),
		AffectedDirectories: len(affected),
		Generation:          applyResult.Generation,
		FreshnessStatus:     applyResult.FreshnessStatus,
	}, nil
}

func buildRefreshFailureResult(
	ctx context.Context,
	store *sqlite.Store,
	repositoryID int64,
	statePath string,
	rootPath string,
	diff refreshcore.Diff,
	affected []string,
	request RefreshRequest,
) (RefreshResult, error) {
	freshness, err := store.ReadRepositoryFreshness(ctx, repositoryID)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("read refresh failure state: %w", err)
	}

	return RefreshResult{
		RepositoryRoot:      rootPath,
		StatePath:           statePath,
		SchemaVersion:       store.SchemaVersion(),
		Reason:              request.Reason,
		AddedFiles:          len(diff.Added),
		ChangedContentFiles: len(diff.Changed),
		DeletedFiles:        len(diff.Deleted),
		MovedFiles:          len(diff.Moved),
		NewlyIgnoredFiles:   len(diff.NewlyIgnored),
		ReincludedFiles:     len(diff.Reincluded),
		ChangedFiles:        changedFileCount(diff),
		UnchangedFiles:      len(diff.Unchanged),
		AffectedDirectories: len(affected),
		Generation:          freshness.CurrentGeneration,
		FreshnessStatus:     freshness.FreshnessStatus,
	}, nil
}

func changedFileCount(diff refreshcore.Diff) int {
	return len(diff.Added) +
		len(diff.Changed) +
		len(diff.Deleted) +
		len(diff.Moved) +
		len(diff.NewlyIgnored) +
		len(diff.Reincluded)
}
