package app

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestHealthService(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Healthy() {}\n")

	refresh := NewRefreshService()
	initial, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	})
	if err != nil {
		t.Fatalf("Refresh() baseline error = %v", err)
	}

	service := NewHealthService()
	healthy, err := service.Health(context.Background(), repoRoot, repository.HealthRequest{})
	if err != nil {
		t.Fatalf("Health() healthy error = %v", err)
	}

	if !healthy.Summary.Initialized || !healthy.Summary.RepositoryRegistered {
		t.Fatalf("healthy summary = %+v", healthy.Summary)
	}
	if healthy.Summary.StateStatus != repository.HealthStateStatusReady {
		t.Fatalf("healthy state status = %q", healthy.Summary.StateStatus)
	}
	if !healthy.Metadata.Present || !healthy.State.DatabaseFile.Exists || !healthy.State.MetadataFile.Exists {
		t.Fatalf("healthy state = %+v metadata = %+v", healthy.State, healthy.Metadata)
	}
	if healthy.Repository.RepositoryRoot != repoRoot {
		t.Fatalf("RepositoryRoot = %q, want %q", healthy.Repository.RepositoryRoot, repoRoot)
	}
	if healthy.Repository.Generation != initial.Generation || healthy.Repository.Freshness != repository.FreshnessStatusFresh {
		t.Fatalf("healthy repository envelope = %+v", healthy.Repository)
	}
	if healthy.Refresh.LastRefreshStatus != repository.RefreshRunStatusSuccess || healthy.Refresh.LastRefreshGeneration != initial.Generation {
		t.Fatalf("healthy refresh = %+v", healthy.Refresh)
	}

	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Degraded() {}\n")
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonManual,
		InjectFailure: func(stage string) error {
			if stage == "after_files" {
				return errors.New("forced degradation")
			}
			return nil
		},
	}); err == nil {
		t.Fatal("Refresh() expected injected failure")
	}

	degraded, err := service.Health(context.Background(), repoRoot, repository.HealthRequest{})
	if err != nil {
		t.Fatalf("Health() degraded error = %v", err)
	}
	if degraded.Repository.Freshness != repository.FreshnessStatusPartiallyDegraded {
		t.Fatalf("degraded freshness = %q", degraded.Repository.Freshness)
	}
	if degraded.Refresh.LastRefreshStatus != repository.RefreshRunStatusFailed {
		t.Fatalf("degraded last refresh status = %q", degraded.Refresh.LastRefreshStatus)
	}
	if degraded.Refresh.CurrentGeneration != initial.Generation+1 || degraded.Refresh.LastRefreshGeneration != initial.Generation {
		t.Fatalf("degraded generations = %+v", degraded.Refresh)
	}
	if degraded.Refresh.FreshnessReason != "forced degradation" {
		t.Fatalf("degraded freshness reason = %q", degraded.Refresh.FreshnessReason)
	}
}

func TestHealthServiceForRootPath(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Healthy() {}\n")

	refresh := NewRefreshService()
	initial, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	})
	if err != nil {
		t.Fatalf("Refresh() baseline error = %v", err)
	}

	service := NewHealthService()
	healthy, err := service.HealthForRootPath(context.Background(), repoRoot, repository.HealthRequest{})
	if err != nil {
		t.Fatalf("HealthForRootPath() error = %v", err)
	}

	if healthy.Repository.RepositoryRoot != repoRoot {
		t.Fatalf("RepositoryRoot = %q, want %q", healthy.Repository.RepositoryRoot, repoRoot)
	}
	if healthy.Repository.Generation != initial.Generation {
		t.Fatalf("Generation = %d, want %d", healthy.Repository.Generation, initial.Generation)
	}
	if healthy.Repository.Freshness != repository.FreshnessStatusFresh {
		t.Fatalf("Freshness = %q, want %q", healthy.Repository.Freshness, repository.FreshnessStatusFresh)
	}
}

func TestPackService(t *testing.T) {
	repoRoot := initRepo(t)
	source := "package pkg\n\ntype Alpha struct{}\n\nfunc (Alpha) Run() {\n\tprintln(\"run\")\n}\n"
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), source)
	writeRepoFile(t, filepath.Join(repoRoot, "docs", "guide.go"), "package docs\n\nfunc Alpha() {}\n")

	refresh := NewRefreshService()
	if _, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	lookup, err := NewLookupService().SymbolLookup(context.Background(), repoRoot, repository.SymbolLookupRequest{
		Name:  "Run",
		Kind:  "method",
		Limit: 1,
	})
	if err != nil {
		t.Fatalf("SymbolLookup() error = %v", err)
	}

	request := repository.PackRequest{
		IncludeRepositoryContext: true,
		IncludeStructuralContext: true,
		SymbolLookups: []repository.SymbolLookupRequest{{
			Name:  "Alpha",
			Limit: 2,
		}},
		StructureLookups: []repository.StructureLookupRequest{{
			Kind:       "method",
			ParentName: "Alpha",
			Limit:      2,
		}},
		Targets: []repository.TargetedContextRequest{{
			StableKey:   lookup.Matches[0].StableKey,
			BeforeLines: 1,
			AfterLines:  1,
		}},
	}

	service := NewPackService()
	first, err := service.Pack(context.Background(), repoRoot, request)
	if err != nil {
		t.Fatalf("Pack() first error = %v", err)
	}
	second, err := service.Pack(context.Background(), repoRoot, request)
	if err != nil {
		t.Fatalf("Pack() second error = %v", err)
	}

	if first.Repository.RepositoryRoot != repoRoot || first.Repository.Freshness != repository.FreshnessStatusFresh {
		t.Fatalf("pack repository envelope = %+v", first.Repository)
	}
	if first.Summary.RequestedSectionCount != 5 || first.Summary.ReturnedSectionCount != 5 {
		t.Fatalf("pack summary = %+v", first.Summary)
	}
	if first.Bundle.RepositoryContext == nil || first.Bundle.StructuralContext == nil {
		t.Fatalf("pack bundle missing shared context = %+v", first.Bundle)
	}
	if len(first.Bundle.Symbols) != 1 || len(first.Bundle.Structures) != 1 || len(first.Bundle.Targets) != 1 {
		t.Fatalf("pack bundle counts = %+v", first.Bundle)
	}
	if first.Bundle.Targets[0].Path != "pkg/alpha.go" || len(first.Bundle.Targets[0].Source) == 0 {
		t.Fatalf("pack target = %+v", first.Bundle.Targets[0])
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
		t.Fatalf("pack payload differs across reads\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}

	_, err = service.Pack(context.Background(), repoRoot, repository.PackRequest{
		IncludeRepositoryContext: true,
		IncludeStructuralContext: true,
		SymbolLookups: []repository.SymbolLookupRequest{
			{Name: "Alpha", Limit: 1},
			{Name: "Alpha", Limit: 1},
			{Name: "Alpha", Limit: 1},
			{Name: "Alpha", Limit: 1},
			{Name: "Alpha", Limit: 1},
			{Name: "Alpha", Limit: 1},
			{Name: "Alpha", Limit: 1},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "max sections") {
		t.Fatalf("Pack() oversized error = %v, want max sections failure", err)
	}
}
