package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"

	_ "modernc.org/sqlite"
)

type HealthService struct {
	Locator       repository.Locator
	ResolveLayout func(string) (state.Layout, error)
	ReadMetadata  func(state.Layout) (state.Metadata, error)
	Stat          func(string) (os.FileInfo, error)
	OpenDB        func(string) (*sql.DB, error)
}

func NewHealthService() HealthService {
	return HealthService{
		Locator:       repository.NewLocator(),
		ResolveLayout: state.ResolveLayout,
		ReadMetadata: func(layout state.Layout) (state.Metadata, error) {
			return layout.ReadMetadata()
		},
		Stat: os.Stat,
		OpenDB: func(path string) (*sql.DB, error) {
			return sql.Open("sqlite", "file:"+path+"?mode=ro")
		},
	}
}

func (s HealthService) Health(ctx context.Context, startPath string, request repository.HealthRequest) (repository.HealthResult, error) {
	root, err := s.Locator.Resolve(startPath)
	if err != nil {
		return repository.HealthResult{}, fmt.Errorf("resolve repository root: %w", err)
	}

	layoutResolver := s.ResolveLayout
	if layoutResolver == nil {
		layoutResolver = state.ResolveLayout
	}
	layout, err := layoutResolver(root.RootPath)
	if err != nil {
		return repository.HealthResult{}, fmt.Errorf("resolve state layout: %w", err)
	}

	statFn := s.Stat
	if statFn == nil {
		statFn = os.Stat
	}
	result := repository.HealthResult{
		Repository: repository.LayeredContextEnvelope{
			RepositoryRoot: root.RootPath,
		},
		Identity: repository.LayeredContextRepositoryIdentity{
			RootPath:      root.RootPath,
			DetectionMode: root.DetectionMode,
			GitHeadRef:    root.Fingerprint.GitHeadRef,
			GitHeadCommit: root.Fingerprint.GitHeadCommit,
		},
		Request: request,
		State: repository.HealthStateLayout{
			StateDir:     healthPathStatus(statFn, layout.StateDir),
			MetadataFile: healthPathStatus(statFn, layout.MetadataPath),
			DatabaseFile: healthPathStatus(statFn, layout.DatabasePath),
			LogsDir:      healthPathStatus(statFn, layout.LogsDir),
			TmpDir:       healthPathStatus(statFn, layout.TmpDir),
		},
	}
	result.Summary.StateStatus = healthStateStatus(result.State)

	readMetadata := s.ReadMetadata
	if readMetadata == nil {
		readMetadata = func(layout state.Layout) (state.Metadata, error) {
			return layout.ReadMetadata()
		}
	}
	if result.State.MetadataFile.Exists {
		metadata, err := readMetadata(layout)
		if err != nil {
			return repository.HealthResult{}, fmt.Errorf("read state metadata: %w", err)
		}
		result.Metadata = repository.HealthStateMetadata{
			Present:           true,
			FormatVersion:     metadata.FormatVersion,
			RepoRoot:          metadata.RepoRoot,
			RepoDetectionMode: metadata.RepoDetectionMode,
			CreatedAt:         parseRFC3339OrZero(metadata.CreatedAt),
			UpdatedAt:         parseRFC3339OrZero(metadata.UpdatedAt),
			RuntimeVersion:    metadata.RuntimeVersion,
			SchemaVersion:     metadata.SchemaVersion,
		}
	}

	if !result.State.DatabaseFile.Exists {
		result.Summary.Initialized = result.Metadata.Present
		return result, nil
	}

	openDB := s.OpenDB
	if openDB == nil {
		openDB = func(path string) (*sql.DB, error) {
			return sql.Open("sqlite", "file:"+path+"?mode=ro")
		}
	}
	db, err := openDB(layout.DatabasePath)
	if err != nil {
		return repository.HealthResult{}, fmt.Errorf("open state database: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return repository.HealthResult{}, fmt.Errorf("ping state database: %w", err)
	}

	refresh, err := readHealthRefresh(ctx, db, root.RootPath)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return repository.HealthResult{}, fmt.Errorf("read repository freshness: %w", err)
		}
		result.Summary.Initialized = result.Metadata.Present
		return result, nil
	}

	result.Refresh = refresh
	result.Repository.Generation = refresh.LastRefreshGeneration
	result.Repository.Freshness = refresh.Freshness
	result.Identity.DetectionMode = nonEmptyOrFallback(refresh.DetectionMode, result.Identity.DetectionMode)
	result.Identity.GitHeadRef = nonEmptyOrFallback(refresh.GitHeadRef, result.Identity.GitHeadRef)
	result.Identity.GitHeadCommit = nonEmptyOrFallback(refresh.GitHeadCommit, result.Identity.GitHeadCommit)
	result.Summary.RepositoryRegistered = true
	result.Summary.Initialized = result.Metadata.Present && refresh.Present
	return result, nil
}

func readHealthRefresh(ctx context.Context, db *sql.DB, rootPath string) (repository.HealthRefreshDiagnostics, error) {
	var refresh repository.HealthRefreshDiagnostics
	var detectionMode string
	var gitHeadRef sql.NullString
	var gitHeadCommit sql.NullString
	var lastRefreshStartedAt sql.NullString
	var lastRefreshCompletedAt sql.NullString
	var lastRefreshReason sql.NullString
	var lastRefreshStatus sql.NullString
	var freshnessStatus sql.NullString
	var freshnessReason sql.NullString

	err := db.QueryRowContext(ctx, `
		SELECT
			id,
			detection_mode,
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
		WHERE root_path = ?
	`, rootPath).Scan(
		&refresh.RepositoryID,
		&detectionMode,
		&gitHeadRef,
		&gitHeadCommit,
		&lastRefreshStartedAt,
		&lastRefreshCompletedAt,
		&lastRefreshReason,
		&lastRefreshStatus,
		&freshnessStatus,
		&freshnessReason,
		&refresh.CurrentGeneration,
		&refresh.LastRefreshGeneration,
	)
	if err != nil {
		return repository.HealthRefreshDiagnostics{}, err
	}

	refresh.Present = true
	refresh.DetectionMode = detectionMode
	refresh.GitHeadRef = gitHeadRef.String
	refresh.GitHeadCommit = gitHeadCommit.String
	refresh.LastRefreshStartedAt = parseNullableRFC3339(lastRefreshStartedAt)
	refresh.LastRefreshCompletedAt = parseNullableRFC3339(lastRefreshCompletedAt)
	refresh.LastRefreshReason = repository.RefreshReason(lastRefreshReason.String)
	refresh.LastRefreshStatus = repository.RefreshRunStatus(lastRefreshStatus.String)
	refresh.Freshness = repository.FreshnessStatus(freshnessStatus.String)
	refresh.FreshnessReason = freshnessReason.String
	return refresh, nil
}

func healthPathStatus(statFn func(string) (os.FileInfo, error), path string) repository.HealthPathStatus {
	_, err := statFn(path)
	return repository.HealthPathStatus{
		Path:   path,
		Exists: err == nil,
	}
}

func healthStateStatus(layout repository.HealthStateLayout) repository.HealthStateStatus {
	statuses := []repository.HealthPathStatus{
		layout.StateDir,
		layout.MetadataFile,
		layout.DatabaseFile,
		layout.LogsDir,
		layout.TmpDir,
	}
	existing := 0
	for _, status := range statuses {
		if status.Exists {
			existing++
		}
	}
	switch {
	case existing == 0:
		return repository.HealthStateStatusMissing
	case existing == len(statuses):
		return repository.HealthStateStatusReady
	default:
		return repository.HealthStateStatusPartial
	}
}

func parseRFC3339OrZero(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func parseNullableRFC3339(value sql.NullString) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return parseRFC3339OrZero(value.String)
}

func nonEmptyOrFallback(value string, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
