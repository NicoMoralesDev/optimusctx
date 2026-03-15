package goextract

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/extract"
	"github.com/niccrow/optimusctx/internal/repository"
)

func TestGoAdapter(t *testing.T) {
	t.Parallel()

	source := readFixture(t, "clean.go")
	artifacts := extractFixture(t, "clean.go", source)

	if artifacts.Extraction.CoverageState != repository.ExtractionCoverageStateSupported {
		t.Fatalf("coverage state = %q", artifacts.Extraction.CoverageState)
	}
	if artifacts.Extraction.SymbolCount != int64(len(artifacts.Symbols)) {
		t.Fatalf("symbol count mismatch = %+v", artifacts.Extraction)
	}

	gotKinds := make([]string, 0, len(artifacts.Symbols))
	for _, symbol := range artifacts.Symbols {
		gotKinds = append(gotKinds, symbol.Kind)
	}
	wantKinds := []string{"package", "const", "var", "type", "struct", "field", "field", "type", "interface", "method", "method", "function"}
	if !reflect.DeepEqual(gotKinds, wantKinds) {
		t.Fatalf("symbol kinds = %v", gotKinds)
	}

	assertSymbolNameAt(t, source, artifacts.Symbols[0])
	assertSymbolNameAt(t, source, artifacts.Symbols[4])
	assertSymbolNameAt(t, source, artifacts.Symbols[11])
	if artifacts.Symbols[11].SignatureStartByte == 0 || artifacts.Symbols[11].SignatureEndByte == 0 {
		t.Fatalf("function signature bytes not recorded: %+v", artifacts.Symbols[11])
	}
}

func TestGoSymbolDeterminism(t *testing.T) {
	t.Parallel()

	source := readFixture(t, "clean.go")
	first := extractFixture(t, "clean.go", source)
	second := extractFixture(t, "clean.go", source)

	if !reflect.DeepEqual(first.Symbols, second.Symbols) {
		t.Fatalf("symbols differ across runs\nfirst=%+v\nsecond=%+v", first.Symbols, second.Symbols)
	}
	if first.Extraction != second.Extraction {
		t.Fatalf("extractions differ across runs\nfirst=%+v\nsecond=%+v", first.Extraction, second.Extraction)
	}
}

func TestGoSymbolOwnership(t *testing.T) {
	t.Parallel()

	source := readFixture(t, "clean.go")
	artifacts := extractFixture(t, "clean.go", source)

	byName := make(map[string]repository.SymbolRecord, len(artifacts.Symbols))
	for _, symbol := range artifacts.Symbols {
		byName[symbol.Kind+":"+symbol.Name] = symbol
	}

	widgetType := byName["type:Widget"]
	widgetStruct := byName["struct:Widget"]
	runMethod := byName["method:Run"]
	nameField := byName["field:Name"]
	handlerMethod := byName["method:Handle"]

	if widgetStruct.ParentStableKey != widgetType.StableKey {
		t.Fatalf("struct parent = %q, want type %q", widgetStruct.ParentStableKey, widgetType.StableKey)
	}
	if nameField.ParentStableKey != widgetStruct.StableKey {
		t.Fatalf("field parent = %q, want struct %q", nameField.ParentStableKey, widgetStruct.StableKey)
	}
	if runMethod.ParentStableKey != widgetType.StableKey || runMethod.QualifiedName != "Widget.Run" {
		t.Fatalf("method ownership = %+v", runMethod)
	}
	if handlerMethod.ParentStableKey == "" || handlerMethod.Depth != 1 {
		t.Fatalf("interface method ownership = %+v", handlerMethod)
	}
}

func TestPartialExtraction(t *testing.T) {
	t.Parallel()

	source := readFixture(t, "syntax_error.go")
	artifacts := extractFixture(t, "syntax_error.go", source)

	if artifacts.Extraction.CoverageState != repository.ExtractionCoverageStatePartial {
		t.Fatalf("coverage state = %q", artifacts.Extraction.CoverageState)
	}
	if artifacts.Extraction.CoverageReason != repository.ExtractionCoverageReasonParseError {
		t.Fatalf("coverage reason = %q", artifacts.Extraction.CoverageReason)
	}
	if artifacts.Extraction.ParserErrorCount == 0 || !artifacts.Extraction.HasErrorNodes {
		t.Fatalf("parse diagnostics = %+v", artifacts.Extraction)
	}

	names := make([]string, 0, len(artifacts.Symbols))
	for _, symbol := range artifacts.Symbols {
		names = append(names, symbol.Name)
	}
	if !slices.Contains(names, "stable") || slices.Contains(names, "broken") {
		t.Fatalf("partial symbols = %+v", artifacts.Symbols)
	}
}

func TestFailedExtraction(t *testing.T) {
	t.Parallel()

	source := []byte("package broken\nfunc (\n")
	artifacts := extractFixture(t, "broken.go", source)

	if artifacts.Extraction.CoverageState != repository.ExtractionCoverageStateFailed {
		t.Fatalf("coverage state = %q", artifacts.Extraction.CoverageState)
	}
	if artifacts.Extraction.CoverageReason != repository.ExtractionCoverageReasonParseError {
		t.Fatalf("coverage reason = %q", artifacts.Extraction.CoverageReason)
	}
	if len(artifacts.Symbols) != 0 || artifacts.Extraction.SymbolCount != 0 {
		t.Fatalf("failed extraction should not retain symbols: %+v", artifacts)
	}
}

func TestExtractionDeterminism(t *testing.T) {
	t.Parallel()

	source := readFixture(t, "syntax_error.go")
	first := extractFixture(t, "syntax_error.go", source)
	second := extractFixture(t, "syntax_error.go", source)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("partial extraction differs across runs\nfirst=%+v\nsecond=%+v", first, second)
	}
}

func extractFixture(t *testing.T, path string, source []byte) repository.FileStructuralArtifacts {
	t.Helper()

	adapter := New()
	result, err := adapter.Extract(context.Background(), extract.Request{
		Candidate: repository.ExtractionCandidate{
			RepositoryID:     1,
			FileID:           2,
			Path:             path,
			Language:         "go",
			ContentHash:      "hash-" + path,
			SourceGeneration: 3,
			RefreshRunID:     4,
		},
		Content: source,
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	return extract.BuildArtifacts(repository.ExtractionCandidate{
		RepositoryID:     1,
		FileID:           2,
		Path:             path,
		Language:         "go",
		ContentHash:      "hash-" + path,
		SourceGeneration: 3,
		RefreshRunID:     4,
	}, adapter.Name(), adapter.GrammarVersion(), result, testExtractedAt())
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	path := filepath.Join("testdata", name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return source
}

func assertSymbolNameAt(t *testing.T, source []byte, symbol repository.SymbolRecord) {
	t.Helper()

	if got := string(source[symbol.NameStartByte:symbol.NameEndByte]); got != symbol.Name {
		t.Fatalf("name span = %q, want %q for %+v", got, symbol.Name, symbol)
	}
}

func testExtractedAt() time.Time {
	return time.Date(2026, 3, 15, 1, 0, 0, 0, time.UTC)
}
