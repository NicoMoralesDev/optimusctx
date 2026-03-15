package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestTargetedContextBlock(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), strings.Join([]string{
		"package pkg",
		"",
		"type Alpha struct{}",
		"",
		"func (Alpha) Run() {",
		`\tprintln("run")`,
		"}",
		"",
		"func Tail() {}",
		"",
	}, "\n"))

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	lookup, err := NewLookupService().SymbolLookup(context.Background(), repoRoot, repository.SymbolLookupRequest{
		Name: "Tail",
		Kind: "function",
	})
	if err != nil {
		t.Fatalf("SymbolLookup() error = %v", err)
	}
	if len(lookup.Matches) != 1 {
		t.Fatalf("SymbolLookup() matches = %d, want 1", len(lookup.Matches))
	}

	service := NewContextBlockService()
	symbolBlock, err := service.TargetedContextBlock(context.Background(), repoRoot, repository.TargetedContextRequest{
		SymbolStableKey: lookup.Matches[0].StableKey,
		BeforeLines:     1,
		AfterLines:      2,
	})
	if err != nil {
		t.Fatalf("TargetedContextBlock(symbol) error = %v", err)
	}

	if symbolBlock.Repository.RepositoryRoot != repoRoot {
		t.Fatalf("Repository.RepositoryRoot = %q, want %q", symbolBlock.Repository.RepositoryRoot, repoRoot)
	}
	if symbolBlock.Target.Path != "pkg/alpha.go" || symbolBlock.Target.Kind != "function" || symbolBlock.Target.Name != "Tail" {
		t.Fatalf("symbol target = %+v", symbolBlock.Target)
	}
	if symbolBlock.Target.StartLine != 9 || symbolBlock.Target.EndLine != 9 {
		t.Fatalf("symbol target anchors = %+v", symbolBlock.Target)
	}
	if symbolBlock.Window.StartLine != 8 || symbolBlock.Window.EndLine != 9 {
		t.Fatalf("symbol window = %+v", symbolBlock.Window)
	}
	if !strings.Contains(symbolBlock.Content, "func Tail() {}") {
		t.Fatalf("symbol block content = %q", symbolBlock.Content)
	}

	lineBlock, err := service.TargetedContextBlock(context.Background(), repoRoot, repository.TargetedContextRequest{
		Path:        "pkg/alpha.go",
		StartLine:   5,
		EndLine:     7,
		BeforeLines: 0,
		AfterLines:  1,
	})
	if err != nil {
		t.Fatalf("TargetedContextBlock(line range) error = %v", err)
	}
	if lineBlock.Target.StableKey != "" || lineBlock.Target.Path != "pkg/alpha.go" {
		t.Fatalf("line target = %+v", lineBlock.Target)
	}
	if lineBlock.Window.StartLine != 5 || lineBlock.Window.EndLine != 8 {
		t.Fatalf("line window = %+v", lineBlock.Window)
	}
	if !strings.Contains(lineBlock.Content, "func (Alpha) Run() {") || strings.Contains(lineBlock.Content, "func Tail() {}") {
		t.Fatalf("line block content = %q", lineBlock.Content)
	}
}

func TestTargetedContextBlockFileAvailability(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\nfunc Alpha() {}\n")

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	lookup, err := NewLookupService().SymbolLookup(context.Background(), repoRoot, repository.SymbolLookupRequest{Name: "Alpha"})
	if err != nil {
		t.Fatalf("SymbolLookup() error = %v", err)
	}
	if len(lookup.Matches) != 1 {
		t.Fatalf("SymbolLookup() matches = %d, want 1", len(lookup.Matches))
	}

	service := NewContextBlockService()
	if err := os.Remove(filepath.Join(repoRoot, "pkg", "alpha.go")); err != nil {
		t.Fatalf("Remove(alpha.go) error = %v", err)
	}
	if _, err := service.TargetedContextBlock(context.Background(), repoRoot, repository.TargetedContextRequest{
		SymbolStableKey: lookup.Matches[0].StableKey,
	}); err == nil || !strings.Contains(err.Error(), "missing from the worktree") {
		t.Fatalf("TargetedContextBlock(missing file) error = %v", err)
	}

	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n")
	if _, err := service.TargetedContextBlock(context.Background(), repoRoot, repository.TargetedContextRequest{
		SymbolStableKey: lookup.Matches[0].StableKey,
	}); err == nil || !strings.Contains(err.Error(), "indexed target needs line") {
		t.Fatalf("TargetedContextBlock(stale anchor) error = %v", err)
	}

	if _, err := service.TargetedContextBlock(context.Background(), repoRoot, repository.TargetedContextRequest{
		Path:      "pkg/alpha.go",
		StartLine: 0,
		EndLine:   1,
	}); err == nil || !strings.Contains(err.Error(), "explicit line range") {
		t.Fatalf("TargetedContextBlock(invalid line range) error = %v", err)
	}
}
