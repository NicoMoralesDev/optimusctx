package sqlite

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
)

func TestTokenTree(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	refreshedAt := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	if err := store.WriteRepositoryFreshness(ctx, repository.RepositoryFreshness{
		RepositoryID:           repoID,
		RootPath:               layout.RepoRoot,
		DetectionMode:          repository.DetectionModeGit,
		LastRefreshStartedAt:   refreshedAt,
		LastRefreshCompletedAt: refreshedAt.Add(time.Minute),
		LastRefreshReason:      repository.RefreshReasonManual,
		LastRefreshStatus:      repository.RefreshRunStatusSuccess,
		FreshnessStatus:        repository.FreshnessStatusFresh,
		CurrentGeneration:      10,
		LastRefreshGeneration:  10,
	}); err != nil {
		t.Fatalf("WriteRepositoryFreshness() error = %v", err)
	}

	mustInsertDirectoryRecord(t, ctx, store, repoID, ".", nil, 5, 2, 720, 10)
	mustInsertDirectoryRecord(t, ctx, store, repoID, "pkg", ".", 3, 1, 500, 10)
	mustInsertDirectoryRecord(t, ctx, store, repoID, "pkg/internal", "pkg", 1, 0, 120, 10)
	mustInsertDirectoryRecord(t, ctx, store, repoID, "docs", ".", 1, 0, 180, 10)

	run := createTestRefreshRun(t, ctx, store, repoID, 10)
	insertSizedTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{Path: "pkg/alpha.go", DirectoryPath: "pkg", Extension: ".go", Language: "go", ContentHash: "alpha", LastGeneration: 10}, 200)
	insertSizedTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{Path: "pkg/zeta.go", DirectoryPath: "pkg", Extension: ".go", Language: "go", ContentHash: "zeta", LastGeneration: 10}, 180)
	insertSizedTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{Path: "pkg/internal/worker.go", DirectoryPath: "pkg/internal", Extension: ".go", Language: "go", ContentHash: "worker", LastGeneration: 10}, 120)
	insertSizedTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{Path: "docs/guide.md", DirectoryPath: "docs", Extension: ".md", Language: "markdown", ContentHash: "guide", LastGeneration: 10}, 180)
	insertSizedTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{Path: "README.md", DirectoryPath: ".", Extension: ".md", Language: "markdown", ContentHash: "readme", LastGeneration: 10}, 40)

	got, err := store.ReadTokenTree(ctx, repoID, repository.TokenTreeRequest{
		PathPrefix: "pkg",
		MaxDepth:   2,
		MaxNodes:   8,
	}, repository.BudgetEstimatePolicy{Name: "bytes_div_4_ceiling", BytesPerToken: 4})
	if err != nil {
		t.Fatalf("ReadTokenTree() error = %v", err)
	}

	if got.Summary.TotalNodeCount != 5 || got.Summary.DepthLimitedNodeCount != 5 || got.Summary.ReturnedNodeCount != 5 {
		t.Fatalf("summary counts = %+v", got.Summary)
	}
	if got.Summary.Truncated || got.Summary.TotalEstimatedTokens != 125 {
		t.Fatalf("summary bounds = %+v", got.Summary)
	}
	if got.Root.Path != "pkg" || got.Root.Kind != repository.TokenTreeNodeKindDirectory {
		t.Fatalf("root = %+v", got.Root)
	}
	if got.Root.IncludedFileCount != 3 || got.Root.IncludedDirectoryCount != 1 {
		t.Fatalf("root counts = %+v", got.Root)
	}
	if gotPaths := tokenTreeChildPaths(got.Root.Children); !reflect.DeepEqual(gotPaths, []string{"pkg/internal", "pkg/alpha.go", "pkg/zeta.go"}) {
		t.Fatalf("root child paths = %v", gotPaths)
	}
	if nested := tokenTreeChildPaths(got.Root.Children[0].Children); !reflect.DeepEqual(nested, []string{"pkg/internal/worker.go"}) {
		t.Fatalf("nested child paths = %v", nested)
	}
}

func TestTokenTreeBounds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	refreshedAt := time.Date(2026, 3, 15, 12, 30, 0, 0, time.UTC)
	if err := store.WriteRepositoryFreshness(ctx, repository.RepositoryFreshness{
		RepositoryID:           repoID,
		RootPath:               layout.RepoRoot,
		DetectionMode:          repository.DetectionModeGit,
		LastRefreshStartedAt:   refreshedAt,
		LastRefreshCompletedAt: refreshedAt.Add(time.Minute),
		LastRefreshReason:      repository.RefreshReasonManual,
		LastRefreshStatus:      repository.RefreshRunStatusSuccess,
		FreshnessStatus:        repository.FreshnessStatusFresh,
		CurrentGeneration:      11,
		LastRefreshGeneration:  11,
	}); err != nil {
		t.Fatalf("WriteRepositoryFreshness() error = %v", err)
	}

	mustInsertDirectoryRecord(t, ctx, store, repoID, ".", nil, 5, 2, 720, 11)
	mustInsertDirectoryRecord(t, ctx, store, repoID, "pkg", ".", 3, 1, 500, 11)
	mustInsertDirectoryRecord(t, ctx, store, repoID, "pkg/internal", "pkg", 1, 0, 120, 11)
	mustInsertDirectoryRecord(t, ctx, store, repoID, "docs", ".", 1, 0, 180, 11)

	run := createTestRefreshRun(t, ctx, store, repoID, 11)
	insertSizedTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{Path: "pkg/alpha.go", DirectoryPath: "pkg", Extension: ".go", Language: "go", ContentHash: "alpha", LastGeneration: 11}, 200)
	insertSizedTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{Path: "pkg/zeta.go", DirectoryPath: "pkg", Extension: ".go", Language: "go", ContentHash: "zeta", LastGeneration: 11}, 180)
	insertSizedTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{Path: "pkg/internal/worker.go", DirectoryPath: "pkg/internal", Extension: ".go", Language: "go", ContentHash: "worker", LastGeneration: 11}, 120)
	insertSizedTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{Path: "docs/guide.md", DirectoryPath: "docs", Extension: ".md", Language: "markdown", ContentHash: "guide", LastGeneration: 11}, 180)
	insertSizedTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{Path: "README.md", DirectoryPath: ".", Extension: ".md", Language: "markdown", ContentHash: "readme", LastGeneration: 11}, 40)

	got, err := store.ReadTokenTree(ctx, repoID, repository.TokenTreeRequest{
		MaxDepth: 1,
		MaxNodes: 3,
	}, repository.BudgetEstimatePolicy{Name: "bytes_div_4_ceiling", BytesPerToken: 4})
	if err != nil {
		t.Fatalf("ReadTokenTree() error = %v", err)
	}

	if !got.Summary.Truncated || !got.Summary.DepthTruncated || !got.Summary.NodeLimitTruncated {
		t.Fatalf("summary = %+v", got.Summary)
	}
	if got.Summary.TotalNodeCount != 9 || got.Summary.DepthLimitedNodeCount != 4 || got.Summary.ReturnedNodeCount != 3 {
		t.Fatalf("summary counts = %+v", got.Summary)
	}
	if got.Root.ReturnedChildCount != 2 || !got.Root.ChildrenTruncated {
		t.Fatalf("root truncation = %+v", got.Root)
	}
	if gotPaths := tokenTreeChildPaths(got.Root.Children); !reflect.DeepEqual(gotPaths, []string{"pkg", "docs"}) {
		t.Fatalf("root child paths = %v", gotPaths)
	}
	if !got.Root.Children[0].ChildrenTruncated || !got.Root.Children[1].ChildrenTruncated {
		t.Fatalf("depth truncation on children = %+v", got.Root.Children)
	}
}

func tokenTreeChildPaths(nodes []repository.TokenTreeNode) []string {
	paths := make([]string, 0, len(nodes))
	for _, node := range nodes {
		paths = append(paths, node.Path)
	}
	return paths
}
