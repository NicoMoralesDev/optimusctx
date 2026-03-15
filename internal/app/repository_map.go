package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

type RepositoryMapService struct {
	Locator       repository.Locator
	OpenStore     func(context.Context, state.Layout, string) (*sqlite.Store, error)
	ResolveLayout func(string) (state.Layout, error)
}

func NewRepositoryMapService() RepositoryMapService {
	return RepositoryMapService{
		Locator: repository.NewLocator(),
		OpenStore: func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		},
		ResolveLayout: state.ResolveLayout,
	}
}

func (s RepositoryMapService) RepositoryMap(ctx context.Context, startPath string) (repository.RepositoryMap, error) {
	root, err := s.Locator.Resolve(startPath)
	if err != nil {
		if errors.Is(err, repository.ErrRepositoryNotFound) {
			return repository.RepositoryMap{}, fmt.Errorf("resolve repository root: %w", err)
		}
		return repository.RepositoryMap{}, fmt.Errorf("resolve repository root: %w", err)
	}

	layoutResolver := s.ResolveLayout
	if layoutResolver == nil {
		layoutResolver = state.ResolveLayout
	}
	layout, err := layoutResolver(root.RootPath)
	if err != nil {
		return repository.RepositoryMap{}, fmt.Errorf("resolve state layout: %w", err)
	}

	openStore := s.OpenStore
	if openStore == nil {
		openStore = func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		}
	}
	store, err := openStore(ctx, layout, root.DetectionMode)
	if err != nil {
		return repository.RepositoryMap{}, fmt.Errorf("open state store: %w", err)
	}
	defer store.Close()

	repositoryID, err := store.LookupRepositoryID(ctx, root.RootPath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.RepositoryMap{}, fmt.Errorf("load repository metadata: %w", err)
		}
		return repository.RepositoryMap{}, fmt.Errorf("load repository metadata: %w", err)
	}

	freshness, err := store.ReadRepositoryFreshness(ctx, repositoryID)
	if err != nil {
		return repository.RepositoryMap{}, fmt.Errorf("load repository freshness: %w", err)
	}
	directories, err := store.LoadRepositoryMapDirectories(ctx, repositoryID)
	if err != nil {
		return repository.RepositoryMap{}, fmt.Errorf("load repository map directories: %w", err)
	}
	files, err := store.LoadRepositoryMapRecords(ctx, repositoryID)
	if err != nil {
		return repository.RepositoryMap{}, fmt.Errorf("load repository map files: %w", err)
	}

	filesByDirectory := make(map[string][]repository.RepositoryMapFile, len(directories))
	for _, record := range files {
		filesByDirectory[record.DirectoryPath] = append(filesByDirectory[record.DirectoryPath], repository.RepositoryMapFile{
			Path:                record.Path,
			DirectoryPath:       record.DirectoryPath,
			Language:            record.Language,
			CoverageState:       record.CoverageState,
			CoverageReason:      record.CoverageReason,
			SymbolCount:         record.SymbolCount,
			TopLevelSymbolCount: record.TopLevelSymbolCount,
			MaxSymbolDepth:      record.MaxSymbolDepth,
			SourceGeneration:    record.SourceGeneration,
			Symbols:             compactRepositoryMapSymbols(record.Symbols),
		})
	}

	result := repository.RepositoryMap{
		RepositoryRoot: root.RootPath,
		Generation:     freshness.LastRefreshGeneration,
		Freshness:      freshness.FreshnessStatus,
		Directories:    make([]repository.RepositoryMapDirectory, 0, len(directories)),
	}
	for _, directory := range directories {
		result.Directories = append(result.Directories, repository.RepositoryMapDirectory{
			Path:                   directory.Path,
			ParentPath:             directory.ParentPath,
			IncludedFileCount:      directory.IncludedFileCount,
			IncludedDirectoryCount: directory.IncludedDirectoryCount,
			TotalSizeBytes:         directory.TotalSizeBytes,
			LastRefreshGeneration:  directory.LastRefreshGeneration,
			Files:                  filesByDirectory[directory.Path],
		})
	}

	return result, nil
}

func compactRepositoryMapSymbols(symbols []repository.SymbolRecord) []repository.RepositoryMapSymbol {
	if len(symbols) == 0 {
		return nil
	}

	compact := make([]repository.RepositoryMapSymbol, 0, len(symbols))
	for _, symbol := range symbols {
		compact = append(compact, repository.RepositoryMapSymbol{
			Kind:          symbol.Kind,
			Name:          symbol.Name,
			QualifiedName: symbol.QualifiedName,
			Ordinal:       symbol.Ordinal,
		})
	}
	return compact
}
