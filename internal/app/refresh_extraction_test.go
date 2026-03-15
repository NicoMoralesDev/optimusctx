package app

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestNoOpRefreshKeepsArtifactsStable(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "README.md"), "hello\n")
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Alpha() {}\n")

	service := NewRefreshService()
	initial, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	})
	if err != nil {
		t.Fatalf("Refresh() initial error = %v", err)
	}

	before := loadArtifactsForRepo(t, repoRoot)
	second, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonManual,
	})
	if err != nil {
		t.Fatalf("Refresh() no-op error = %v", err)
	}
	after := loadArtifactsForRepo(t, repoRoot)

	if second.ExtractionQueued != 0 || second.ArtifactsReplaced != 0 {
		t.Fatalf("no-op extraction stats = %+v", second)
	}
	if !reflect.DeepEqual(before.extractions, after.extractions) {
		t.Fatalf("file extractions changed on no-op:\nbefore=%+v\nafter=%+v", before.extractions, after.extractions)
	}
	if !reflect.DeepEqual(before.symbols, after.symbols) {
		t.Fatalf("symbols changed on no-op:\nbefore=%+v\nafter=%+v", before.symbols, after.symbols)
	}
	if after.extractions["main.go"].SourceGeneration != initial.Generation {
		t.Fatalf("main.go source generation = %d, want %d", after.extractions["main.go"].SourceGeneration, initial.Generation)
	}
}

func TestRefreshQueuesExtractionCandidates(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "README.md"), "hello\n")
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Alpha() {}\n")

	service := NewRefreshService()
	if _, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() initial error = %v", err)
	}

	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Beta() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "helper.go"), "package pkg\n\nfunc Helper() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "docs", "guide.txt"), "guide\n")

	result, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonManual,
	})
	if err != nil {
		t.Fatalf("Refresh() incremental error = %v", err)
	}

	if result.ExtractionQueued != 3 {
		t.Fatalf("ExtractionQueued = %d, want 3", result.ExtractionQueued)
	}
	if result.ArtifactsReplaced != 3 {
		t.Fatalf("ArtifactsReplaced = %d, want 3", result.ArtifactsReplaced)
	}
	if result.UnsupportedFiles != 1 {
		t.Fatalf("UnsupportedFiles = %d, want 1", result.UnsupportedFiles)
	}

	artifacts := loadArtifactsForRepo(t, repoRoot)
	if artifacts.extractions["main.go"].SourceGeneration != result.Generation {
		t.Fatalf("main.go source generation = %d, want %d", artifacts.extractions["main.go"].SourceGeneration, result.Generation)
	}
	if artifacts.extractions["docs/guide.txt"].CoverageState != repository.ExtractionCoverageStateUnsupported {
		t.Fatalf("guide.txt coverage = %q", artifacts.extractions["docs/guide.txt"].CoverageState)
	}
	if !reflect.DeepEqual(symbolNamesForPath(artifacts.symbols, "main.go"), []string{"main", "Beta"}) {
		t.Fatalf("main.go symbols = %+v", artifacts.symbols["main.go"])
	}
}
