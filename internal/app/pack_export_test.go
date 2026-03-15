package app

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestPackExportManifest(t *testing.T) {
	repoRoot := initRepo(t)
	source := "package pkg\n\ntype Alpha struct{}\n\ntype Beta struct{}\n\nfunc (Alpha) Run() {\n\tprintln(\"run\")\n}\n\nfunc (Beta) Run() {\n\tprintln(\"run\")\n}\n"
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

	request := repository.PackExportRequest{
		PackRequest: repository.PackRequest{
			IncludeRepositoryContext: true,
			IncludeStructuralContext: true,
			SymbolLookups: []repository.SymbolLookupRequest{{
				Name:  "Alpha",
				Limit: 1,
			}},
			StructureLookups: []repository.StructureLookupRequest{{
				Kind:       "method",
				ParentName: "Alpha",
				Limit:      1,
			}},
			Targets: []repository.TargetedContextRequest{{
				StableKey:   lookup.Matches[0].StableKey,
				BeforeLines: 1,
				AfterLines:  1,
			}},
		},
		GeneratedAt: "2026-03-15T18:40:00Z",
		Generator:   "optimusctx test",
	}

	service := NewPackExportService()
	first, err := service.Export(context.Background(), repoRoot, request)
	if err != nil {
		t.Fatalf("Export() first error = %v", err)
	}
	second, err := service.Export(context.Background(), repoRoot, request)
	if err != nil {
		t.Fatalf("Export() second error = %v", err)
	}

	manifest := first.Artifact.Manifest
	if manifest.Format != repository.PackExportFormatJSON {
		t.Fatalf("Format = %q, want json", manifest.Format)
	}
	if manifest.Compression != repository.PackExportCompressionNone {
		t.Fatalf("Compression = %q, want none", manifest.Compression)
	}
	if manifest.Repository.RepositoryRoot != repoRoot || manifest.Freshness != repository.FreshnessStatusFresh {
		t.Fatalf("Repository = %+v", manifest.Repository)
	}
	if manifest.ExportSummary.RequestedSectionCount != 5 || manifest.ExportSummary.IncludedSectionCount != 5 {
		t.Fatalf("ExportSummary = %+v", manifest.ExportSummary)
	}
	if manifest.ExportSummary.TruncatedSectionCount != 3 {
		t.Fatalf("TruncatedSectionCount = %d, want 3", manifest.ExportSummary.TruncatedSectionCount)
	}
	if len(manifest.IncludedSections) != 5 || len(manifest.OmittedSections) != 0 {
		t.Fatalf("section counts = %d included / %d omitted", len(manifest.IncludedSections), len(manifest.OmittedSections))
	}

	if manifest.IncludedSections[0].Kind != repository.PackExportSectionRepositoryContext {
		t.Fatalf("section[0] kind = %q", manifest.IncludedSections[0].Kind)
	}
	if manifest.IncludedSections[1].Kind != repository.PackExportSectionStructuralContext {
		t.Fatalf("section[1] kind = %q", manifest.IncludedSections[1].Kind)
	}
	if manifest.IncludedSections[2].Kind != repository.PackExportSectionSymbolLookup || !manifest.IncludedSections[2].Truncated {
		t.Fatalf("section[2] = %+v, want truncated symbol lookup", manifest.IncludedSections[2])
	}
	if manifest.IncludedSections[3].Kind != repository.PackExportSectionStructureLookup || !manifest.IncludedSections[3].Truncated {
		t.Fatalf("section[3] = %+v, want truncated structure lookup", manifest.IncludedSections[3])
	}
	if manifest.IncludedSections[4].Kind != repository.PackExportSectionTargetContext || !manifest.IncludedSections[4].Truncated {
		t.Fatalf("section[4] = %+v, want truncated target context", manifest.IncludedSections[4])
	}

	firstJSON, err := json.Marshal(first.Artifact)
	if err != nil {
		t.Fatalf("Marshal(first.Artifact) error = %v", err)
	}
	secondJSON, err := json.Marshal(second.Artifact)
	if err != nil {
		t.Fatalf("Marshal(second.Artifact) error = %v", err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("artifact differs across reads\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}

	indented, err := marshalPackExportArtifact(first.Artifact, repository.PackExportFormatJSON)
	if err != nil {
		t.Fatalf("marshalPackExportArtifact() error = %v", err)
	}
	output := string(indented)
	for _, fragment := range []string{
		"\"format\": \"json\"",
		"\"compression\": \"none\"",
		"\"generatedAt\": \"2026-03-15T18:40:00Z\"",
		"\"includedSections\": [",
		"\"label\": \"repository_context\"",
		"\"label\": \"symbol_lookup[0]\"",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("output = %s, want fragment %q", output, fragment)
		}
	}
}
