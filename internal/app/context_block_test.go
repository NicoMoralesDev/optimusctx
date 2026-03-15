package app

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestTargetedContextBlock(t *testing.T) {
	repoRoot := initRepo(t)
	source := "package pkg\n\ntype Alpha struct{}\n\nfunc (Alpha) Run() {\n\tprintln(\"run\")\n}\n"
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), source)

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	lookupResult, err := NewLookupService().SymbolLookup(context.Background(), repoRoot, repository.SymbolLookupRequest{
		Name: "Run",
		Kind: "method",
	})
	if err != nil {
		t.Fatalf("SymbolLookup() error = %v", err)
	}

	service := NewContextBlockService()
	symbolBlock, err := service.TargetedContext(context.Background(), repoRoot, repository.TargetedContextRequest{
		StableKey:   lookupResult.Matches[0].StableKey,
		BeforeLines: 1,
		AfterLines:  1,
	})
	if err != nil {
		t.Fatalf("TargetedContext(stable key) error = %v", err)
	}
	if symbolBlock.Path != "pkg/alpha.go" || symbolBlock.AnchorStart != 5 || symbolBlock.AnchorEnd != 7 {
		t.Fatalf("symbol block = %+v", symbolBlock)
	}
	if !reflect.DeepEqual(symbolBlock.Source, []string{"", "func (Alpha) Run() {", "\tprintln(\"run\")", "}"}) {
		t.Fatalf("symbol block source = %v", symbolBlock.Source)
	}

	lineBlock, err := service.TargetedContext(context.Background(), repoRoot, repository.TargetedContextRequest{
		Path:        "pkg/alpha.go",
		StartLine:   3,
		EndLine:     5,
		BeforeLines: 1,
		AfterLines:  0,
	})
	if err != nil {
		t.Fatalf("TargetedContext(line range) error = %v", err)
	}
	if lineBlock.StartLine != 2 || lineBlock.EndLine != 5 {
		t.Fatalf("line block bounds = %+v", lineBlock)
	}
	if !reflect.DeepEqual(lineBlock.Source, []string{"", "type Alpha struct{}", "", "func (Alpha) Run() {"}) {
		t.Fatalf("line block source = %v", lineBlock.Source)
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

	lookupResult, err := NewLookupService().SymbolLookup(context.Background(), repoRoot, repository.SymbolLookupRequest{
		Name: "Alpha",
		Kind: "function",
	})
	if err != nil {
		t.Fatalf("SymbolLookup() error = %v", err)
	}

	service := NewContextBlockService()
	if err := os.Remove(filepath.Join(repoRoot, "pkg", "alpha.go")); err != nil {
		t.Fatalf("Remove(alpha.go) error = %v", err)
	}
	if _, err := service.TargetedContext(context.Background(), repoRoot, repository.TargetedContextRequest{
		StableKey: lookupResult.Matches[0].StableKey,
	}); err == nil {
		t.Fatal("TargetedContext() expected missing file error")
	}

	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n")
	if _, err := service.TargetedContext(context.Background(), repoRoot, repository.TargetedContextRequest{
		Path:        "pkg/alpha.go",
		StartLine:   2,
		EndLine:     3,
		BeforeLines: 0,
		AfterLines:  0,
	}); err == nil {
		t.Fatal("TargetedContext() expected stale line-range error")
	}
}
