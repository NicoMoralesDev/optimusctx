package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
)

func TestDeleteFileArtifacts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	run := createTestRefreshRun(t, ctx, store, repoID, 5)
	removedFileID := insertTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{
		Path:           "pkg/remove.go",
		DirectoryPath:  "pkg",
		Extension:      ".go",
		Language:       "go",
		ContentHash:    "hash-remove-1",
		LastGeneration: 5,
	})
	stableFileID := insertTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{
		Path:           "pkg/stable.go",
		DirectoryPath:  "pkg",
		Extension:      ".go",
		Language:       "go",
		ContentHash:    "hash-stable-1",
		LastGeneration: 5,
	})

	if _, err := store.ReplaceFileArtifacts(ctx, artifactSeed(repoID, removedFileID, run.ID, 5, "pkg/remove.go", "hash-remove-1", "Remove")); err != nil {
		t.Fatalf("ReplaceFileArtifacts() removed error = %v", err)
	}
	if _, err := store.ReplaceFileArtifacts(ctx, artifactSeed(repoID, stableFileID, run.ID, 5, "pkg/stable.go", "hash-stable-1", "Stable")); err != nil {
		t.Fatalf("ReplaceFileArtifacts() stable error = %v", err)
	}

	if err := deleteFileArtifactsByPath(ctx, store.DB(), repoID, []string{"pkg/remove.go"}); err != nil {
		t.Fatalf("deleteFileArtifactsByPath() error = %v", err)
	}

	extractions, err := store.ListFileExtractions(ctx, repoID)
	if err != nil {
		t.Fatalf("ListFileExtractions() error = %v", err)
	}
	if len(extractions) != 1 || extractions[0].Path != "pkg/stable.go" {
		t.Fatalf("remaining extractions = %+v", extractions)
	}

	symbols, err := store.ListSymbols(ctx, repoID)
	if err != nil {
		t.Fatalf("ListSymbols() error = %v", err)
	}
	if len(symbols) != 1 || symbols[0].Path != "pkg/stable.go" {
		t.Fatalf("remaining symbols = %+v", symbols)
	}
}

func TestExtractionFailureIsolation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	run := createTestRefreshRun(t, ctx, store, repoID, 9)
	failedFileID := insertTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{
		Path:           "pkg/failing.go",
		DirectoryPath:  "pkg",
		Extension:      ".go",
		Language:       "go",
		ContentHash:    "hash-failing-1",
		LastGeneration: 9,
	})
	stableFileID := insertTestFileRecord(t, ctx, store, repoID, run.ID, testFileSeed{
		Path:           "pkg/stable.go",
		DirectoryPath:  "pkg",
		Extension:      ".go",
		Language:       "go",
		ContentHash:    "hash-stable-1",
		LastGeneration: 9,
	})

	if _, err := store.ReplaceFileArtifacts(ctx, artifactSeed(repoID, failedFileID, run.ID, 9, "pkg/failing.go", "hash-failing-1", "Failing")); err != nil {
		t.Fatalf("ReplaceFileArtifacts() failing baseline error = %v", err)
	}
	if _, err := store.ReplaceFileArtifacts(ctx, artifactSeed(repoID, stableFileID, run.ID, 9, "pkg/stable.go", "hash-stable-1", "Stable")); err != nil {
		t.Fatalf("ReplaceFileArtifacts() stable baseline error = %v", err)
	}

	if _, err := store.ReplaceFileArtifacts(ctx, repository.FileStructuralArtifacts{
		Extraction: repository.FileExtractionRecord{
			RepositoryID:      repoID,
			FileID:            failedFileID,
			Path:              "pkg/failing.go",
			Language:          "go",
			AdapterName:       "tree-sitter-go",
			GrammarVersion:    "v0.25.0",
			SourceContentHash: "hash-failing-2",
			SourceGeneration:  10,
			CoverageState:     repository.ExtractionCoverageStateFailed,
			CoverageReason:    repository.ExtractionCoverageReasonAdapterError,
			ExtractedAt:       time.Date(2026, 3, 15, 2, 0, 0, 0, time.UTC),
			RefreshRunID:      run.ID,
		},
	}); err != nil {
		t.Fatalf("ReplaceFileArtifacts() failed replacement error = %v", err)
	}

	extractions, err := store.ListFileExtractions(ctx, repoID)
	if err != nil {
		t.Fatalf("ListFileExtractions() error = %v", err)
	}
	if len(extractions) != 2 {
		t.Fatalf("extraction count = %d, want 2", len(extractions))
	}

	symbols, err := store.ListSymbols(ctx, repoID)
	if err != nil {
		t.Fatalf("ListSymbols() error = %v", err)
	}
	if len(symbols) != 1 || symbols[0].Path != "pkg/stable.go" || symbols[0].Name != "Stable" {
		t.Fatalf("symbols after failed replacement = %+v", symbols)
	}
}

func artifactSeed(repoID, fileID, refreshRunID, generation int64, path, contentHash, name string) repository.FileStructuralArtifacts {
	return repository.FileStructuralArtifacts{
		Extraction: repository.FileExtractionRecord{
			RepositoryID:        repoID,
			FileID:              fileID,
			Path:                path,
			Language:            "go",
			AdapterName:         "tree-sitter-go",
			GrammarVersion:      "v0.25.0",
			SourceContentHash:   contentHash,
			SourceGeneration:    generation,
			CoverageState:       repository.ExtractionCoverageStateSupported,
			TopLevelSymbolCount: 1,
			SymbolCount:         1,
			ExtractedAt:         time.Date(2026, 3, 15, 1, 30, 0, 0, time.UTC),
			RefreshRunID:        refreshRunID,
		},
		Symbols: []repository.SymbolRecord{{
			StableKey:   path + ":" + name,
			Path:        path,
			Language:    "go",
			Kind:        "function",
			Name:        name,
			Ordinal:     0,
			StartByte:   0,
			EndByte:     10,
			StartRow:    0,
			StartColumn: 0,
			EndRow:      0,
			EndColumn:   10,
		}},
	}
}
