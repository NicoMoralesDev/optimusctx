package extract

import (
	"context"
	"sort"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
)

const (
	unsupportedAdapterName    = "none"
	unsupportedGrammarVersion = "none"
)

type Request struct {
	Candidate   repository.ExtractionCandidate
	Content     []byte
	ExtractedAt time.Time
}

type Result struct {
	CoverageState    repository.ExtractionCoverageState
	CoverageReason   repository.ExtractionCoverageReason
	ParserErrorCount int64
	HasErrorNodes    bool
	Symbols          []repository.SymbolRecord
}

type Adapter interface {
	Name() string
	Language() string
	GrammarVersion() string
	Extract(context.Context, Request) (Result, error)
}

func NormalizeSymbols(path string, language string, symbols []repository.SymbolRecord) []repository.SymbolRecord {
	normalized := make([]repository.SymbolRecord, len(symbols))
	copy(normalized, symbols)

	sort.SliceStable(normalized, func(i, j int) bool {
		left := normalized[i]
		right := normalized[j]
		switch {
		case left.StartByte != right.StartByte:
			return left.StartByte < right.StartByte
		case left.Depth != right.Depth:
			return left.Depth < right.Depth
		case left.EndByte != right.EndByte:
			return left.EndByte < right.EndByte
		case left.Name != right.Name:
			return left.Name < right.Name
		case left.Kind != right.Kind:
			return left.Kind < right.Kind
		default:
			return left.StableKey < right.StableKey
		}
	})

	for i := range normalized {
		normalized[i].Path = path
		normalized[i].Language = language
		normalized[i].Ordinal = int64(i)
	}

	return normalized
}

func BuildArtifacts(candidate repository.ExtractionCandidate, adapterName string, grammarVersion string, result Result, extractedAt time.Time) repository.FileStructuralArtifacts {
	language := candidate.Language
	if language == "" {
		language = "unknown"
	}

	symbols := NormalizeSymbols(candidate.Path, language, result.Symbols)

	artifacts := repository.FileStructuralArtifacts{
		Extraction: repository.FileExtractionRecord{
			RepositoryID:        candidate.RepositoryID,
			FileID:              candidate.FileID,
			Path:                candidate.Path,
			Language:            language,
			AdapterName:         adapterName,
			GrammarVersion:      grammarVersion,
			SourceContentHash:   candidate.ContentHash,
			SourceGeneration:    candidate.SourceGeneration,
			CoverageState:       result.CoverageState,
			CoverageReason:      result.CoverageReason,
			ParserErrorCount:    result.ParserErrorCount,
			HasErrorNodes:       result.HasErrorNodes,
			SymbolCount:         int64(len(symbols)),
			TopLevelSymbolCount: countTopLevelSymbols(symbols),
			MaxSymbolDepth:      maxSymbolDepth(symbols),
			ExtractedAt:         extractedAt.UTC(),
			RefreshRunID:        candidate.RefreshRunID,
		},
		Symbols: symbols,
	}

	return artifacts
}

func UnsupportedArtifacts(candidate repository.ExtractionCandidate, extractedAt time.Time) repository.FileStructuralArtifacts {
	return BuildArtifacts(candidate, unsupportedAdapterName, unsupportedGrammarVersion, Result{
		CoverageState:  repository.ExtractionCoverageStateUnsupported,
		CoverageReason: repository.ExtractionCoverageReasonUnsupportedLanguage,
	}, extractedAt)
}

func countTopLevelSymbols(symbols []repository.SymbolRecord) int64 {
	var count int64
	for _, symbol := range symbols {
		if symbol.Depth == 0 {
			count++
		}
	}
	return count
}

func maxSymbolDepth(symbols []repository.SymbolRecord) int64 {
	var maxDepth int64
	for _, symbol := range symbols {
		if symbol.Depth > maxDepth {
			maxDepth = symbol.Depth
		}
	}
	return maxDepth
}
