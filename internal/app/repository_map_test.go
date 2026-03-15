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

func TestRepositoryMap(t *testing.T) {
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

	service := NewRepositoryMapService()
	got, err := service.RepositoryMap(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("RepositoryMap() error = %v", err)
	}

	if got.RepositoryRoot != repoRoot {
		t.Fatalf("RepositoryRoot = %q, want %q", got.RepositoryRoot, repoRoot)
	}
	if got.Generation == 0 {
		t.Fatal("Generation should be non-zero")
	}

	pkgDir := repositoryMapDirectoryByPath(t, got, "pkg")
	if got := fileNames(pkgDir.Files); !reflect.DeepEqual(got, []string{"pkg/alpha.go", "pkg/partial.go"}) {
		t.Fatalf("pkg files = %v", got)
	}

	alphaFile := repositoryMapFileByPath(t, got, "pkg/alpha.go")
	if alphaFile.CoverageState != repository.ExtractionCoverageStateSupported {
		t.Fatalf("alpha coverage = %q", alphaFile.CoverageState)
	}
	if alphaFile.TopLevelSymbolCount != 3 {
		t.Fatalf("alpha top-level count = %d", alphaFile.TopLevelSymbolCount)
	}
	if got := repositoryMapSymbolNames(alphaFile.Symbols); !reflect.DeepEqual(got, []string{"pkg", "Alpha", "Beta"}) {
		t.Fatalf("alpha symbols = %v", got)
	}

	rootDir := repositoryMapDirectoryByPath(t, got, ".")
	readmeFile := repositoryMapFileByPath(t, got, "README.md")
	if !containsFile(rootDir.Files, "README.md") || readmeFile.CoverageState != repository.ExtractionCoverageStateUnsupported {
		t.Fatalf("root README map = %+v", readmeFile)
	}
}

func TestRepositoryMapCoverageStates(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "healthy.go"), "package main\n\nfunc Healthy() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "partial.go"), "package main\n\nfunc HealthyPartial() {}\nfunc (\n")
	writeRepoFile(t, filepath.Join(repoRoot, "failed.go"), "package main\nfunc (\n")
	writeRepoFile(t, filepath.Join(repoRoot, "notes.txt"), "plain text\n")

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	mapResult, err := NewRepositoryMapService().RepositoryMap(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("RepositoryMap() error = %v", err)
	}

	healthy := repositoryMapFileByPath(t, mapResult, "healthy.go")
	if healthy.CoverageState != repository.ExtractionCoverageStateSupported || len(healthy.Symbols) == 0 {
		t.Fatalf("healthy map file = %+v", healthy)
	}

	partial := repositoryMapFileByPath(t, mapResult, "partial.go")
	if partial.CoverageState != repository.ExtractionCoverageStatePartial || len(partial.Symbols) == 0 {
		t.Fatalf("partial map file = %+v", partial)
	}

	failed := repositoryMapFileByPath(t, mapResult, "failed.go")
	if failed.CoverageState != repository.ExtractionCoverageStateFailed || len(failed.Symbols) != 0 {
		t.Fatalf("failed map file = %+v", failed)
	}

	unsupported := repositoryMapFileByPath(t, mapResult, "notes.txt")
	if unsupported.CoverageState != repository.ExtractionCoverageStateUnsupported || len(unsupported.Symbols) != 0 {
		t.Fatalf("unsupported map file = %+v", unsupported)
	}
}

func TestRepositoryMapOrdering(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "z-last.go"), "package main\n\nfunc Zeta() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "a-first.go"), "package main\n\ntype Alpha struct{}\n\nfunc Beta() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "m-middle.go"), "package pkg\n\nfunc Middle() {}\n")

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	mapResult, err := NewRepositoryMapService().RepositoryMap(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("RepositoryMap() error = %v", err)
	}

	if got := repositoryMapDirectoryPaths(mapResult); !reflect.DeepEqual(got, []string{".", "nested", "pkg"}) {
		t.Fatalf("directory order = %v", got)
	}
	if got := fileNames(repositoryMapDirectoryByPath(t, mapResult, ".").Files); !reflect.DeepEqual(got, []string{"a-first.go", "z-last.go"}) {
		t.Fatalf("root file order = %v", got)
	}
	if got := repositoryMapSymbolNames(repositoryMapFileByPath(t, mapResult, "a-first.go").Symbols); !reflect.DeepEqual(got, []string{"main", "Alpha", "Beta"}) {
		t.Fatalf("symbol order = %v", got)
	}
}

func TestPersistedOnlyRepositoryMap(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Persisted() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "notes.txt"), "plain text\n")

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
	if err := os.Remove(filepath.Join(repoRoot, "notes.txt")); err != nil {
		t.Fatalf("Remove(notes.txt) error = %v", err)
	}

	mapResult, err := NewRepositoryMapService().RepositoryMap(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("RepositoryMap() after delete error = %v", err)
	}

	mainFile := repositoryMapFileByPath(t, mapResult, "main.go")
	if mainFile.CoverageState != repository.ExtractionCoverageStateSupported || len(mainFile.Symbols) != 1 {
		t.Fatalf("main persisted map file = %+v", mainFile)
	}
	notesFile := repositoryMapFileByPath(t, mapResult, "notes.txt")
	if notesFile.CoverageState != repository.ExtractionCoverageStateUnsupported {
		t.Fatalf("notes persisted map file = %+v", notesFile)
	}
}

func TestRepositoryMapDeterminism(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "alpha.go"), "package main\n\ntype Alpha struct{}\n\nfunc Beta() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "notes.txt"), "plain text\n")

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	service := NewRepositoryMapService()
	first, err := service.RepositoryMap(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("first RepositoryMap() error = %v", err)
	}
	second, err := service.RepositoryMap(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("second RepositoryMap() error = %v", err)
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
		t.Fatalf("repository map payload differs across reads\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}
}

func repositoryMapDirectoryByPath(t *testing.T, result repository.RepositoryMap, path string) repository.RepositoryMapDirectory {
	t.Helper()

	for _, directory := range result.Directories {
		if directory.Path == path {
			return directory
		}
	}

	t.Fatalf("directory %q not found", path)
	return repository.RepositoryMapDirectory{}
}

func repositoryMapFileByPath(t *testing.T, result repository.RepositoryMap, path string) repository.RepositoryMapFile {
	t.Helper()

	for _, directory := range result.Directories {
		for _, file := range directory.Files {
			if file.Path == path {
				return file
			}
		}
	}

	t.Fatalf("file %q not found", path)
	return repository.RepositoryMapFile{}
}

func repositoryMapDirectoryPaths(result repository.RepositoryMap) []string {
	paths := make([]string, 0, len(result.Directories))
	for _, directory := range result.Directories {
		paths = append(paths, directory.Path)
	}
	return paths
}

func repositoryMapSymbolNames(symbols []repository.RepositoryMapSymbol) []string {
	names := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		names = append(names, symbol.Name)
	}
	return names
}

func fileNames(files []repository.RepositoryMapFile) []string {
	names := make([]string, 0, len(files))
	for _, file := range files {
		names = append(names, file.Path)
	}
	return names
}

func containsFile(files []repository.RepositoryMapFile, path string) bool {
	for _, file := range files {
		if file.Path == path {
			return true
		}
	}
	return false
}
