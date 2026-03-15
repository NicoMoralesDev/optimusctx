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

func TestBudgetAnalysis(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\n"+strings.Repeat("a", 200))
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "zeta.go"), "package pkg\n\n"+strings.Repeat("z", 200))
	writeRepoFile(t, filepath.Join(repoRoot, "docs", "guide.md"), strings.Repeat("g", 120))

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	got, err := NewBudgetAnalysisService().Analyze(context.Background(), repoRoot, repository.BudgetAnalysisRequest{
		GroupBy: repository.BudgetGroupByFile,
		Limit:   2,
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if got.Policy.BytesPerToken != 4 || got.Summary.TotalEstimatedTokens == 0 {
		t.Fatalf("policy/summary = %+v %+v", got.Policy, got.Summary)
	}
	if gotPaths := budgetAnalysisPaths(got.Hotspots); !reflect.DeepEqual(gotPaths, []string{"pkg/alpha.go", "pkg/zeta.go"}) {
		t.Fatalf("hotspots = %v", gotPaths)
	}

	if err := os.Remove(filepath.Join(repoRoot, "pkg", "alpha.go")); err != nil {
		t.Fatalf("Remove(alpha.go) error = %v", err)
	}
	if err := os.Remove(filepath.Join(repoRoot, "pkg", "zeta.go")); err != nil {
		t.Fatalf("Remove(zeta.go) error = %v", err)
	}
	if err := os.Remove(filepath.Join(repoRoot, "docs", "guide.md")); err != nil {
		t.Fatalf("Remove(guide.md) error = %v", err)
	}

	persisted, err := NewBudgetAnalysisService().Analyze(context.Background(), repoRoot, repository.BudgetAnalysisRequest{
		GroupBy: repository.BudgetGroupByFile,
		Limit:   2,
	})
	if err != nil {
		t.Fatalf("Analyze() after delete error = %v", err)
	}
	if gotPaths := budgetAnalysisPaths(persisted.Hotspots); !reflect.DeepEqual(gotPaths, []string{"pkg/alpha.go", "pkg/zeta.go"}) {
		t.Fatalf("persisted hotspots = %v", gotPaths)
	}
}

func TestBudgetHotspots(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\n"+strings.Repeat("a", 200))
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "zeta.go"), "package pkg\n\n"+strings.Repeat("z", 200))
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "internal", "worker.go"), "package internal\n\n"+strings.Repeat("w", 120))
	writeRepoFile(t, filepath.Join(repoRoot, "docs", "guide.md"), strings.Repeat("g", 200))

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	service := NewBudgetAnalysisService()
	first, err := service.Analyze(context.Background(), repoRoot, repository.BudgetAnalysisRequest{
		GroupBy:    repository.BudgetGroupByDirectory,
		PathPrefix: "pkg",
		Limit:      5,
	})
	if err != nil {
		t.Fatalf("first Analyze() error = %v", err)
	}
	second, err := service.Analyze(context.Background(), repoRoot, repository.BudgetAnalysisRequest{
		GroupBy:    repository.BudgetGroupByDirectory,
		PathPrefix: "pkg",
		Limit:      5,
	})
	if err != nil {
		t.Fatalf("second Analyze() error = %v", err)
	}

	if gotPaths := budgetAnalysisPaths(first.Hotspots); !reflect.DeepEqual(gotPaths, []string{"pkg", "pkg/internal"}) {
		t.Fatalf("hotspots = %v", gotPaths)
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
		t.Fatalf("budget payload differs across reads\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}
}

func budgetAnalysisPaths(hotspots []repository.BudgetHotspot) []string {
	paths := make([]string, 0, len(hotspots))
	for _, hotspot := range hotspots {
		paths = append(paths, hotspot.Path)
	}
	return paths
}
