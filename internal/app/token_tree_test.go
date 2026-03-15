package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestTokenTree(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\n"+strings.Repeat("a", 200))
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "zeta.go"), "package pkg\n\n"+strings.Repeat("z", 180))
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "internal", "worker.go"), "package internal\n\n"+strings.Repeat("w", 120))
	writeRepoFile(t, filepath.Join(repoRoot, "docs", "guide.md"), strings.Repeat("g", 180))

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	got, err := NewTokenTreeService().Analyze(context.Background(), filepath.Join(repoRoot, "pkg"), repository.TokenTreeRequest{
		PathPrefix: "pkg",
		MaxDepth:   2,
		MaxNodes:   8,
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if got.Policy.BytesPerToken != 4 || got.Summary.TotalEstimatedTokens == 0 {
		t.Fatalf("policy/summary = %+v %+v", got.Policy, got.Summary)
	}
	if got.Root.Path != "pkg" || got.Root.IncludedFileCount != 3 || got.Root.IncludedDirectoryCount != 1 {
		t.Fatalf("root = %+v", got.Root)
	}
	if gotPaths := tokenTreePaths(got.Root.Children); !reflect.DeepEqual(gotPaths, []string{"pkg/internal", "pkg/alpha.go", "pkg/zeta.go"}) {
		t.Fatalf("root child paths = %v", gotPaths)
	}

	if err := os.Remove(filepath.Join(repoRoot, "pkg", "alpha.go")); err != nil {
		t.Fatalf("Remove(alpha.go) error = %v", err)
	}
	if err := os.Remove(filepath.Join(repoRoot, "pkg", "zeta.go")); err != nil {
		t.Fatalf("Remove(zeta.go) error = %v", err)
	}
	if err := os.Remove(filepath.Join(repoRoot, "pkg", "internal", "worker.go")); err != nil {
		t.Fatalf("Remove(worker.go) error = %v", err)
	}
	if err := os.Remove(filepath.Join(repoRoot, "docs", "guide.md")); err != nil {
		t.Fatalf("Remove(guide.md) error = %v", err)
	}

	persisted, err := NewTokenTreeService().Analyze(context.Background(), repoRoot, repository.TokenTreeRequest{
		PathPrefix: "pkg",
		MaxDepth:   2,
		MaxNodes:   8,
	})
	if err != nil {
		t.Fatalf("Analyze() after delete error = %v", err)
	}
	if gotPaths := tokenTreePaths(persisted.Root.Children); !reflect.DeepEqual(gotPaths, []string{"pkg/internal", "pkg/alpha.go", "pkg/zeta.go"}) {
		t.Fatalf("persisted root child paths = %v", gotPaths)
	}
}

func TestTokenTreeBounds(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\n"+strings.Repeat("a", 200))
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "zeta.go"), "package pkg\n\n"+strings.Repeat("z", 180))
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "internal", "worker.go"), "package internal\n\n"+strings.Repeat("w", 120))
	writeRepoFile(t, filepath.Join(repoRoot, "docs", "guide.md"), strings.Repeat("g", 180))
	writeRepoFile(t, filepath.Join(repoRoot, "README.md"), strings.Repeat("r", 40))

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	service := NewTokenTreeService()
	first, err := service.Analyze(context.Background(), repoRoot, repository.TokenTreeRequest{
		MaxDepth: 1,
		MaxNodes: 3,
	})
	if err != nil {
		t.Fatalf("first Analyze() error = %v", err)
	}
	second, err := service.Analyze(context.Background(), repoRoot, repository.TokenTreeRequest{
		MaxDepth: 1,
		MaxNodes: 3,
	})
	if err != nil {
		t.Fatalf("second Analyze() error = %v", err)
	}

	if !first.Summary.Truncated || !first.Summary.DepthTruncated || !first.Summary.NodeLimitTruncated {
		t.Fatalf("summary = %+v", first.Summary)
	}
	if gotPaths := tokenTreePaths(first.Root.Children); !reflect.DeepEqual(gotPaths, []string{"pkg", "docs"}) {
		t.Fatalf("root child paths = %v", gotPaths)
	}

	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("Marshal(first) error = %v", err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("Marshal(second) error = %v", err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("token tree payload differs across reads\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}
}

func tokenTreePaths(nodes []repository.TokenTreeNode) []string {
	paths := make([]string, 0, len(nodes))
	for _, node := range nodes {
		paths = append(paths, node.Path)
	}
	return paths
}
