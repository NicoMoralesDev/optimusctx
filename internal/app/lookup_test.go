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

func TestSymbolLookup(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\ntype Alpha struct{}\n\nfunc Beta() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "docs", "guide.go"), "package docs\n\nfunc Alpha() {}\n")

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	got, err := NewLookupService().SymbolLookup(context.Background(), repoRoot, repository.SymbolLookupRequest{Name: "Alpha"})
	if err != nil {
		t.Fatalf("SymbolLookup() error = %v", err)
	}

	if got.Repository.RepositoryRoot != repoRoot {
		t.Fatalf("Repository.RepositoryRoot = %q, want %q", got.Repository.RepositoryRoot, repoRoot)
	}
	if got.Limit != 10 {
		t.Fatalf("Limit = %d, want 10", got.Limit)
	}
	if got.Request.Name != "Alpha" {
		t.Fatalf("Request = %+v", got.Request)
	}
	gotPaths := symbolLookupPaths(got.Matches)
	if len(gotPaths) < 2 {
		t.Fatalf("match paths = %v", gotPaths)
	}
	if gotPaths[0] != "docs/guide.go" || gotPaths[len(gotPaths)-1] != "pkg/alpha.go" {
		t.Fatalf("match paths = %v", gotPaths)
	}

	first := got.Matches[0]
	if first.StableKey == "" || first.Kind != "function" || first.Name != "Alpha" {
		t.Fatalf("first match = %+v", first)
	}
	if first.StartRow != 2 || first.EndColumn == 0 {
		t.Fatalf("first match anchors = %+v", first)
	}

	if err := os.Remove(filepath.Join(repoRoot, "pkg", "alpha.go")); err != nil {
		t.Fatalf("Remove(alpha.go) error = %v", err)
	}
	if err := os.Remove(filepath.Join(repoRoot, "docs", "guide.go")); err != nil {
		t.Fatalf("Remove(guide.go) error = %v", err)
	}

	persisted, err := NewLookupService().SymbolLookup(context.Background(), repoRoot, repository.SymbolLookupRequest{Name: "Alpha"})
	if err != nil {
		t.Fatalf("SymbolLookup() after delete error = %v", err)
	}
	gotPersistedPaths := symbolLookupPaths(persisted.Matches)
	if len(gotPersistedPaths) != len(gotPaths) {
		t.Fatalf("persisted match count = %d, want %d", len(gotPersistedPaths), len(gotPaths))
	}
	if gotPersistedPaths[0] != "docs/guide.go" || gotPersistedPaths[len(gotPersistedPaths)-1] != "pkg/alpha.go" {
		t.Fatalf("persisted match paths = %v", gotPaths)
	}
}

func TestSymbolLookupFilters(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\ntype Alpha struct{}\n\nfunc Beta() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "docs", "alpha.go"), "package docs\n\nfunc Alpha() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.py"), "class Alpha:\n    pass\n")

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	service := NewLookupService()
	byPath, err := service.SymbolLookup(context.Background(), repoRoot, repository.SymbolLookupRequest{
		Name:       "Alpha",
		PathPrefix: "pkg/",
		Limit:      1,
	})
	if err != nil {
		t.Fatalf("SymbolLookup(path filter) error = %v", err)
	}
	if got := symbolLookupPaths(byPath.Matches); !reflect.DeepEqual(got, []string{"pkg/alpha.go"}) {
		t.Fatalf("path-filtered matches = %v", got)
	}

	byKind, err := service.SymbolLookup(context.Background(), repoRoot, repository.SymbolLookupRequest{
		Name:     "Alpha",
		Kind:     "type",
		Language: "go",
	})
	if err != nil {
		t.Fatalf("SymbolLookup(kind filter) error = %v", err)
	}
	if got := symbolLookupPaths(byKind.Matches); !reflect.DeepEqual(got, []string{"pkg/alpha.go"}) {
		t.Fatalf("kind-filtered matches = %v", got)
	}

	first, err := service.SymbolLookup(context.Background(), repoRoot, repository.SymbolLookupRequest{Name: "Alpha"})
	if err != nil {
		t.Fatalf("first SymbolLookup() error = %v", err)
	}
	second, err := service.SymbolLookup(context.Background(), repoRoot, repository.SymbolLookupRequest{Name: "Alpha"})
	if err != nil {
		t.Fatalf("second SymbolLookup() error = %v", err)
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
		t.Fatalf("symbol lookup payload differs across reads\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}
}

func symbolLookupPaths(matches []repository.SymbolLookupMatch) []string {
	paths := make([]string, 0, len(matches))
	for _, match := range matches {
		paths = append(paths, match.Path)
	}
	return paths
}
