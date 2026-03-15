package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/niccrow/optimusctx/internal/extract"
	"github.com/niccrow/optimusctx/internal/extract/adapter/goextract"
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
	ExtractionQueued    int
	ArtifactsReplaced   int
	ArtifactsDeleted    int
	UnsupportedFiles    int
	FailedExtractions   int
	Generation          int64
	FreshnessStatus     repository.FreshnessStatus
}

type RefreshService struct {
	Locator            repository.Locator
	OpenStore          func(context.Context, state.Layout, string) (*sqlite.Store, error)
	Discover           func(string, repository.RepositorySnapshot, bool) (repository.DiscoveryResult, error)
	ExtractionRegistry extract.Registry
	ExtractionEngine   extract.Engine
	ReadFile           func(string) ([]byte, error)
	ResolveLayout      func(string) (state.Layout, error)
	Now                func() time.Time
}

func NewRefreshService() RefreshService {
	registry, err := extract.NewRegistry(goextract.New())
	if err != nil {
		panic(fmt.Sprintf("create extraction registry: %v", err))
	}
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
		ExtractionRegistry: registry,
		ExtractionEngine:   extract.NewEngine(registry),
		ReadFile:           os.ReadFile,
		ResolveLayout:      state.ResolveLayout,
		Now:                time.Now,
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
	extractionPlan := buildExtractionPlan(diff)
	changedHint := normalizeChangedHints(root.RootPath, request.ChangedHint)
	var extractionStats extractionStats

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
		MetadataJSON:      buildRefreshMetadataJSON(request, diff, affected, changedHint),
		InjectFailure:     request.InjectFailure,
		ApplyStructuralArtifacts: func(ctx context.Context, structuralStore sqlite.RefreshStructuralStore) error {
			stats, err := s.applyStructuralArtifacts(ctx, root.RootPath, extractionPlan, structuralStore)
			if err != nil {
				return err
			}
			extractionStats = stats
			return nil
		},
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
		ExtractionQueued:    extractionStats.Queued,
		ArtifactsReplaced:   extractionStats.Replaced,
		ArtifactsDeleted:    extractionStats.Deleted,
		UnsupportedFiles:    extractionStats.Unsupported,
		FailedExtractions:   extractionStats.Failed,
		Generation:          applyResult.Generation,
		FreshnessStatus:     applyResult.FreshnessStatus,
	}, nil
}

func normalizeChangedHints(root string, hints []string) []string {
	if len(hints) == 0 {
		return nil
	}
	unique := make(map[string]struct{}, len(hints))
	for _, hint := range hints {
		hint = filepath.ToSlash(strings.TrimSpace(hint))
		if hint == "" || hint == "." {
			continue
		}
		if filepath.IsAbs(hint) {
			rel, err := filepath.Rel(root, hint)
			if err != nil {
				continue
			}
			hint = filepath.ToSlash(rel)
		}
		hint = filepath.ToSlash(filepath.Clean(hint))
		if hint == "." || hint == ".." || strings.HasPrefix(hint, "../") {
			continue
		}
		unique[hint] = struct{}{}
	}
	if len(unique) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(unique))
	for hint := range unique {
		normalized = append(normalized, hint)
	}
	sort.Strings(normalized)
	return normalized
}

func buildRefreshMetadataJSON(request RefreshRequest, diff refreshcore.Diff, affected []string, changedHint []string) string {
	payload := map[string]any{
		"reason":               request.Reason,
		"force_full":           request.ForceFull,
		"changed_hint":         changedHint,
		"added":                len(diff.Added),
		"changed":              len(diff.Changed),
		"deleted":              len(diff.Deleted),
		"moved":                len(diff.Moved),
		"newly_ignored":        len(diff.NewlyIgnored),
		"reincluded":           len(diff.Reincluded),
		"affected_directories": affected,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(encoded)
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

type extractionPlan struct {
	candidatePaths []string
	deletePaths    []string
}

type extractionStats struct {
	Queued      int
	Replaced    int
	Deleted     int
	Unsupported int
	Failed      int
}

func buildExtractionPlan(diff refreshcore.Diff) extractionPlan {
	candidatePaths := make(map[string]struct{})
	deletePaths := make(map[string]struct{})

	for _, change := range diff.Added {
		candidatePaths[change.Path] = struct{}{}
	}
	for _, change := range diff.Changed {
		candidatePaths[change.Path] = struct{}{}
	}
	for _, change := range diff.Reincluded {
		candidatePaths[change.Path] = struct{}{}
	}
	for _, change := range diff.Moved {
		candidatePaths[change.Path] = struct{}{}
		deletePaths[change.PreviousPath] = struct{}{}
	}
	for _, change := range diff.Deleted {
		deletePaths[change.Path] = struct{}{}
	}
	for _, change := range diff.NewlyIgnored {
		deletePaths[change.Path] = struct{}{}
	}

	return extractionPlan{
		candidatePaths: sortedPaths(candidatePaths),
		deletePaths:    sortedPaths(deletePaths),
	}
}

func sortedPaths(values map[string]struct{}) []string {
	paths := make([]string, 0, len(values))
	for path := range values {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func (s RefreshService) applyStructuralArtifacts(
	ctx context.Context,
	repositoryRoot string,
	plan extractionPlan,
	store sqlite.RefreshStructuralStore,
) (extractionStats, error) {
	var stats extractionStats
	if len(plan.deletePaths) > 0 {
		if err := store.DeleteFileArtifactsByPath(ctx, plan.deletePaths); err != nil {
			return extractionStats{}, err
		}
		stats.Deleted = len(plan.deletePaths)
	}
	if len(plan.candidatePaths) == 0 {
		return stats, nil
	}

	candidates, err := store.ListExtractionCandidatesByPath(ctx, plan.candidatePaths)
	if err != nil {
		return extractionStats{}, err
	}

	stats.Queued = len(candidates)
	nowFn := s.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	readFile := s.ReadFile
	if readFile == nil {
		readFile = os.ReadFile
	}

	for _, candidate := range candidates {
		extractedAt := nowFn().UTC()
		artifacts, kind := s.extractCandidate(ctx, repositoryRoot, candidate, extractedAt, readFile)
		if _, err := store.ReplaceFileArtifacts(ctx, artifacts); err != nil {
			return extractionStats{}, err
		}
		stats.Replaced++
		switch kind {
		case repository.ExtractionCoverageStateUnsupported:
			stats.Unsupported++
		case repository.ExtractionCoverageStateFailed:
			stats.Failed++
		}
	}

	return stats, nil
}

func (s RefreshService) extractCandidate(
	ctx context.Context,
	repositoryRoot string,
	candidate repository.ExtractionCandidate,
	extractedAt time.Time,
	readFile func(string) ([]byte, error),
) (repository.FileStructuralArtifacts, repository.ExtractionCoverageState) {
	adapter, ok := s.ExtractionRegistry.Resolve(candidate)
	if !ok {
		return extract.UnsupportedArtifacts(candidate, extractedAt), repository.ExtractionCoverageStateUnsupported
	}

	sourcePath := filepath.Join(repositoryRoot, filepath.FromSlash(candidate.Path))
	content, err := readFile(sourcePath)
	if err != nil {
		return failedArtifacts(candidate, adapter, extractedAt), repository.ExtractionCoverageStateFailed
	}

	artifacts, err := s.ExtractionEngine.Extract(ctx, extract.Request{
		Candidate:   candidate,
		Content:     content,
		ExtractedAt: extractedAt,
	})
	if err != nil {
		return failedArtifacts(candidate, adapter, extractedAt), repository.ExtractionCoverageStateFailed
	}

	return artifacts, artifacts.Extraction.CoverageState
}

func failedArtifacts(candidate repository.ExtractionCandidate, adapter extract.Adapter, extractedAt time.Time) repository.FileStructuralArtifacts {
	return extract.BuildArtifacts(candidate, adapter.Name(), adapter.GrammarVersion(), extract.Result{
		CoverageState:  repository.ExtractionCoverageStateFailed,
		CoverageReason: repository.ExtractionCoverageReasonAdapterError,
	}, extractedAt)
}
