package extract

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
)

type stubAdapter struct {
	name           string
	language       string
	grammarVersion string
	result         Result
	calls          int
}

func (s *stubAdapter) Name() string           { return s.name }
func (s *stubAdapter) Language() string       { return s.language }
func (s *stubAdapter) GrammarVersion() string { return s.grammarVersion }
func (s *stubAdapter) Extract(context.Context, Request) (Result, error) {
	s.calls++
	return s.result, nil
}

func TestExtractionRegistry(t *testing.T) {
	t.Parallel()

	goAdapter := &stubAdapter{name: "tree-sitter-go", language: "go", grammarVersion: "v0.25.0"}
	tsAdapter := &stubAdapter{name: "tree-sitter-typescript", language: "typescript", grammarVersion: "v0.25.0"}

	registry, err := NewRegistry(goAdapter, tsAdapter)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	if got := registry.SupportedLanguages(); !reflect.DeepEqual(got, []string{"go", "typescript"}) {
		t.Fatalf("SupportedLanguages() = %v", got)
	}

	adapter, ok := registry.Resolve(repository.ExtractionCandidate{Language: "go"})
	if !ok || adapter.Name() != "tree-sitter-go" {
		t.Fatalf("Resolve(go) = (%v, %t)", adapter, ok)
	}
}

func TestUnsupportedLanguageRouting(t *testing.T) {
	t.Parallel()

	adapter := &stubAdapter{name: "tree-sitter-go", language: "go", grammarVersion: "v0.25.0"}
	registry, err := NewRegistry(adapter)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	engine := NewEngine(registry)
	extractedAt := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	artifacts, err := engine.Extract(context.Background(), Request{
		Candidate: repository.ExtractionCandidate{
			RepositoryID:     1,
			FileID:           2,
			Path:             "scripts/tool.py",
			Language:         "python",
			ContentHash:      "hash-py",
			SourceGeneration: 4,
			RefreshRunID:     8,
		},
		ExtractedAt: extractedAt,
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if adapter.calls != 0 {
		t.Fatalf("unsupported file should not invoke adapter, calls = %d", adapter.calls)
	}
	if artifacts.Extraction.CoverageState != repository.ExtractionCoverageStateUnsupported {
		t.Fatalf("coverage state = %q", artifacts.Extraction.CoverageState)
	}
	if artifacts.Extraction.CoverageReason != repository.ExtractionCoverageReasonUnsupportedLanguage {
		t.Fatalf("coverage reason = %q", artifacts.Extraction.CoverageReason)
	}
	if artifacts.Extraction.AdapterName != unsupportedAdapterName || artifacts.Extraction.GrammarVersion != unsupportedGrammarVersion {
		t.Fatalf("unsupported adapter metadata = %+v", artifacts.Extraction)
	}
	if len(artifacts.Symbols) != 0 || artifacts.Extraction.SymbolCount != 0 {
		t.Fatalf("unsupported symbols = %+v", artifacts.Symbols)
	}
}

func TestExtractionEngine(t *testing.T) {
	t.Parallel()

	adapter := &stubAdapter{
		name:           "tree-sitter-go",
		language:       "go",
		grammarVersion: "v0.25.0",
		result: Result{
			CoverageState: repository.ExtractionCoverageStateSupported,
			Symbols: []repository.SymbolRecord{
				{
					StableKey:  "method:widget.run",
					Kind:       "method",
					Name:       "Run",
					Depth:      1,
					StartByte:  34,
					EndByte:    78,
					StartRow:   4,
					EndRow:     6,
					IsExported: true,
				},
				{
					StableKey:          "type:widget",
					Kind:               "type",
					Name:               "Widget",
					QualifiedName:      "Widget",
					Depth:              0,
					StartByte:          10,
					EndByte:            33,
					StartRow:           1,
					EndRow:             3,
					NameStartByte:      15,
					NameEndByte:        21,
					SignatureStartByte: 10,
					SignatureEndByte:   22,
					IsExported:         true,
				},
			},
		},
	}
	registry, err := NewRegistry(adapter)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	engine := NewEngine(registry)
	artifacts, err := engine.Extract(context.Background(), Request{
		Candidate: repository.ExtractionCandidate{
			RepositoryID:     1,
			FileID:           10,
			Path:             "pkg/widget.go",
			Language:         "go",
			ContentHash:      "hash-go",
			SourceGeneration: 9,
			RefreshRunID:     3,
		},
		ExtractedAt: time.Date(2026, 3, 15, 0, 5, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if adapter.calls != 1 {
		t.Fatalf("adapter calls = %d, want 1", adapter.calls)
	}
	if artifacts.Extraction.AdapterName != "tree-sitter-go" || artifacts.Extraction.GrammarVersion != "v0.25.0" {
		t.Fatalf("adapter metadata = %+v", artifacts.Extraction)
	}
	if artifacts.Extraction.SymbolCount != 2 || artifacts.Extraction.TopLevelSymbolCount != 1 || artifacts.Extraction.MaxSymbolDepth != 1 {
		t.Fatalf("extraction summary = %+v", artifacts.Extraction)
	}
	if artifacts.Symbols[0].Name != "Widget" || artifacts.Symbols[0].Ordinal != 0 {
		t.Fatalf("first symbol = %+v", artifacts.Symbols[0])
	}
	if artifacts.Symbols[1].Name != "Run" || artifacts.Symbols[1].Ordinal != 1 {
		t.Fatalf("second symbol = %+v", artifacts.Symbols[1])
	}
	if artifacts.Symbols[0].Path != "pkg/widget.go" || artifacts.Symbols[1].Language != "go" {
		t.Fatalf("normalized symbols = %+v", artifacts.Symbols)
	}
}
