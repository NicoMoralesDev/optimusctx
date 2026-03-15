package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestLayeredContextL0(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "README.md"), "# Repo\n")
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\ntype Alpha struct{}\n\nfunc Beta() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "partial.go"), "package pkg\n\nfunc Healthy() {}\nfunc (\n")
	writeRepoFile(t, filepath.Join(repoRoot, "docs", "guide.md"), "# Guide\n")

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	got, err := NewRepositoryContextService().LayeredContextL0(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("LayeredContextL0() error = %v", err)
	}

	if got.Repository.RepositoryRoot != repoRoot {
		t.Fatalf("Repository.RepositoryRoot = %q, want %q", got.Repository.RepositoryRoot, repoRoot)
	}
	if got.Repository.Generation == 0 {
		t.Fatal("Repository.Generation should be non-zero")
	}
	if got.Repository.Freshness != repository.FreshnessStatusFresh {
		t.Fatalf("Repository.Freshness = %q, want %q", got.Repository.Freshness, repository.FreshnessStatusFresh)
	}
	if got.Identity.RootPath != repoRoot {
		t.Fatalf("Identity.RootPath = %q, want %q", got.Identity.RootPath, repoRoot)
	}
	if got.Identity.DetectionMode != repository.DetectionModeGit {
		t.Fatalf("Identity.DetectionMode = %q, want %q", got.Identity.DetectionMode, repository.DetectionModeGit)
	}

	if got := layeredContextLanguageNames(got.Languages); !reflect.DeepEqual(got, []string{"go", "markdown"}) {
		t.Fatalf("Languages = %v", got)
	}
	if got.Languages[0].FileCount != 2 || got.Languages[0].TotalSizeBytes == 0 {
		t.Fatalf("primary language summary = %+v", got.Languages[0])
	}

	if got := layeredContextMajorAreaKeys(got.MajorAreas); !reflect.DeepEqual(got, []string{"pkg:directory", "docs:directory", ".:root_files"}) {
		t.Fatalf("MajorAreas = %v", got)
	}
}

func TestPersistedOnlyLayeredContextL0(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Persisted() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "README.md"), "# Repo\n")

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if err := os.Remove(filepath.Join(repoRoot, "main.go")); err != nil {
		t.Fatalf("Remove(main.go) error = %v", err)
	}
	if err := os.Remove(filepath.Join(repoRoot, "README.md")); err != nil {
		t.Fatalf("Remove(README.md) error = %v", err)
	}

	got, err := NewRepositoryContextService().LayeredContextL0(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("LayeredContextL0() after delete error = %v", err)
	}

	if got.Repository.RepositoryRoot != repoRoot || len(got.Languages) != 2 {
		t.Fatalf("persisted LayeredContextL0 = %+v", got)
	}
}

func TestRepositorySummaryDeterminism(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "alpha.go"), "package main\n\ntype Alpha struct{}\n\nfunc Beta() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "README.md"), "# Repo\n")

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	service := NewRepositoryContextService()
	first, err := service.LayeredContextL0(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("first LayeredContextL0() error = %v", err)
	}
	second, err := service.LayeredContextL0(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("second LayeredContextL0() error = %v", err)
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
		t.Fatalf("layered context payload differs across reads\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}
}

func layeredContextLanguageNames(summaries []repository.LayeredContextLanguageSummary) []string {
	names := make([]string, 0, len(summaries))
	for _, summary := range summaries {
		names = append(names, summary.Language)
	}
	return names
}

func layeredContextMajorAreaKeys(summaries []repository.LayeredContextMajorAreaSummary) []string {
	keys := make([]string, 0, len(summaries))
	for _, summary := range summaries {
		keys = append(keys, summary.Path+":"+string(summary.Kind))
	}
	return keys
}
