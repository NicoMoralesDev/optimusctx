package sqlite

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
)

func TestBudgetAnalysis(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	refreshedAt := time.Date(2026, 3, 14, 20, 0, 0, 0, time.UTC)
	if err := store.WriteRepositoryFreshness(ctx, repository.RepositoryFreshness{
		RepositoryID:           repoID,
		RootPath:               layout.RepoRoot,
		DetectionMode:          repository.DetectionModeGit,
		LastRefreshStartedAt:   refreshedAt,
		LastRefreshCompletedAt: refreshedAt.Add(time.Minute),
		LastRefreshReason:      repository.RefreshReasonManual,
		LastRefreshStatus:      repository.RefreshRunStatusSuccess,
		FreshnessStatus:        repository.FreshnessStatusFresh,
		CurrentGeneration:      8,
		LastRefreshGeneration:  8,
	}); err != nil {
		t.Fatalf("WriteRepositoryFreshness() error = %v", err)
	}

	run := createTestRefreshRun(t, ctx, store, repoID, 8)
	insertSizedTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{Path: "pkg/zeta.go", DirectoryPath: "pkg", Extension: ".go", Language: "go", ContentHash: "zeta", LastGeneration: 8}, 200)
	insertSizedTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{Path: "pkg/alpha.go", DirectoryPath: "pkg", Extension: ".go", Language: "go", ContentHash: "alpha", LastGeneration: 8}, 200)
	insertSizedTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{Path: "docs/guide.md", DirectoryPath: "docs", Extension: ".md", Language: "markdown", ContentHash: "guide", LastGeneration: 8}, 120)

	got, err := store.ReadBudgetAnalysis(ctx, repoID, repository.BudgetAnalysisRequest{
		GroupBy: repository.BudgetGroupByFile,
		Limit:   2,
	}, repository.BudgetEstimatePolicy{Name: "bytes_div_4_ceiling", BytesPerToken: 4})
	if err != nil {
		t.Fatalf("ReadBudgetAnalysis() error = %v", err)
	}

	if got.Summary.TotalCount != 3 || got.Summary.ReturnedCount != 2 || !got.Summary.Truncated {
		t.Fatalf("summary = %+v", got.Summary)
	}
	if got.Summary.TotalSizeBytes != 520 || got.Summary.TotalEstimatedTokens != 130 {
		t.Fatalf("summary totals = %+v", got.Summary)
	}
	if gotPaths := budgetHotspotPaths(got.Hotspots); !reflect.DeepEqual(gotPaths, []string{"pkg/alpha.go", "pkg/zeta.go"}) {
		t.Fatalf("hotspot paths = %v", gotPaths)
	}
	if got.Hotspots[0].EstimatedTokens != 50 || got.Hotspots[0].PercentOfTotalBytes <= 0 {
		t.Fatalf("first hotspot = %+v", got.Hotspots[0])
	}
}

func TestBudgetHotspots(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	refreshedAt := time.Date(2026, 3, 14, 20, 30, 0, 0, time.UTC)
	if err := store.WriteRepositoryFreshness(ctx, repository.RepositoryFreshness{
		RepositoryID:           repoID,
		RootPath:               layout.RepoRoot,
		DetectionMode:          repository.DetectionModeGit,
		LastRefreshStartedAt:   refreshedAt,
		LastRefreshCompletedAt: refreshedAt.Add(time.Minute),
		LastRefreshReason:      repository.RefreshReasonManual,
		LastRefreshStatus:      repository.RefreshRunStatusSuccess,
		FreshnessStatus:        repository.FreshnessStatusFresh,
		CurrentGeneration:      9,
		LastRefreshGeneration:  9,
	}); err != nil {
		t.Fatalf("WriteRepositoryFreshness() error = %v", err)
	}

	mustInsertDirectoryRecord(t, ctx, store, repoID, ".", nil, 4, 3, 520, 9)
	mustInsertDirectoryRecord(t, ctx, store, repoID, "pkg", ".", 2, 0, 400, 9)
	mustInsertDirectoryRecord(t, ctx, store, repoID, "pkg/internal", "pkg", 1, 0, 120, 9)
	mustInsertDirectoryRecord(t, ctx, store, repoID, "docs", ".", 1, 0, 400, 9)

	got, err := store.ReadBudgetAnalysis(ctx, repoID, repository.BudgetAnalysisRequest{
		GroupBy:    repository.BudgetGroupByDirectory,
		PathPrefix: "pkg",
		Limit:      5,
	}, repository.BudgetEstimatePolicy{Name: "bytes_div_4_ceiling", BytesPerToken: 4})
	if err != nil {
		t.Fatalf("ReadBudgetAnalysis(directory) error = %v", err)
	}

	if got.Summary.TotalCount != 2 || got.Summary.Truncated {
		t.Fatalf("summary = %+v", got.Summary)
	}
	if got.Summary.TotalEstimatedTokens != 130 {
		t.Fatalf("total estimated tokens = %d, want 130", got.Summary.TotalEstimatedTokens)
	}
	if gotPaths := budgetHotspotPaths(got.Hotspots); !reflect.DeepEqual(gotPaths, []string{"pkg", "pkg/internal"}) {
		t.Fatalf("hotspot paths = %v", gotPaths)
	}
	if got.Hotspots[0].IncludedFileCount != 2 || got.Hotspots[0].EstimatedTokens != 100 {
		t.Fatalf("first hotspot = %+v", got.Hotspots[0])
	}

	if _, err := store.ReadBudgetAnalysis(ctx, repoID, repository.BudgetAnalysisRequest{GroupBy: "symbol"}, repository.BudgetEstimatePolicy{Name: "bytes_div_4_ceiling", BytesPerToken: 4}); err == nil {
		t.Fatal("ReadBudgetAnalysis() expected error for unsupported group")
	}
}

func budgetHotspotPaths(hotspots []repository.BudgetHotspot) []string {
	paths := make([]string, 0, len(hotspots))
	for _, hotspot := range hotspots {
		paths = append(paths, hotspot.Path)
	}
	return paths
}
