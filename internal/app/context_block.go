package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

type ContextBlockService struct {
	Locator       repository.Locator
	OpenStore     func(context.Context, state.Layout, string) (*sqlite.Store, error)
	ResolveLayout func(string) (state.Layout, error)
	ReadFile      func(string) ([]byte, error)
}

func NewContextBlockService() ContextBlockService {
	return ContextBlockService{
		Locator: repository.NewLocator(),
		OpenStore: func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		},
		ResolveLayout: state.ResolveLayout,
		ReadFile:      os.ReadFile,
	}
}

func (s ContextBlockService) TargetedContextBlock(ctx context.Context, startPath string, request repository.TargetedContextRequest) (repository.TargetedContextResult, error) {
	root, store, repositoryID, err := s.openContextStore(ctx, startPath)
	if err != nil {
		return repository.TargetedContextResult{}, err
	}
	defer store.Close()

	request.SymbolStableKey = strings.TrimSpace(request.SymbolStableKey)
	request.Path = strings.TrimSpace(request.Path)
	if err := validateTargetedContextRequest(request); err != nil {
		return repository.TargetedContextResult{}, err
	}

	freshness, err := store.ReadRepositoryFreshness(ctx, repositoryID)
	if err != nil {
		return repository.TargetedContextResult{}, fmt.Errorf("load targeted context freshness: %w", err)
	}

	target := repository.TargetedContextTarget{}
	switch {
	case request.SymbolStableKey != "":
		match, err := LookupService{}.resolveSymbolByStableKey(ctx, store, repositoryID, request.SymbolStableKey)
		if err != nil {
			return repository.TargetedContextResult{}, err
		}
		target = repository.TargetedContextTarget{
			StableKey: match.StableKey,
			Path:      match.Path,
			Kind:      match.Kind,
			Name:      match.Name,
			StartLine: match.StartRow + 1,
			EndLine:   match.EndRow + 1,
		}
	default:
		target = repository.TargetedContextTarget{
			Path:      request.Path,
			StartLine: request.StartLine,
			EndLine:   request.EndLine,
		}
	}

	readFile := s.ReadFile
	if readFile == nil {
		readFile = os.ReadFile
	}
	contentBytes, err := readFile(filepath.Join(root.RootPath, filepath.FromSlash(target.Path)))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return repository.TargetedContextResult{}, fmt.Errorf("read targeted context file %q: file is missing from the worktree", target.Path)
		}
		return repository.TargetedContextResult{}, fmt.Errorf("read targeted context file %q: %w", target.Path, err)
	}

	lines := splitContextLines(string(contentBytes))
	if target.StartLine < 1 || target.EndLine < target.StartLine {
		return repository.TargetedContextResult{}, fmt.Errorf("load targeted context for %q: invalid target line range %d-%d", target.Path, target.StartLine, target.EndLine)
	}
	if int(target.EndLine) > len(lines) {
		return repository.TargetedContextResult{}, fmt.Errorf(
			"load targeted context for %q: live file has %d lines but indexed target needs line %d",
			target.Path,
			len(lines),
			target.EndLine,
		)
	}

	windowStart := maxInt64(1, target.StartLine-int64(request.BeforeLines))
	windowEnd := minInt64(int64(len(lines)), target.EndLine+int64(request.AfterLines))
	windowLines := lines[windowStart-1 : windowEnd]

	return repository.TargetedContextResult{
		Repository: repository.LayeredContextEnvelope{
			RepositoryRoot: freshness.RootPath,
			Generation:     freshness.LastRefreshGeneration,
			Freshness:      freshness.FreshnessStatus,
		},
		Identity: repository.LayeredContextRepositoryIdentity{
			RootPath:      freshness.RootPath,
			DetectionMode: freshness.DetectionMode,
			GitHeadRef:    freshness.GitHeadRef,
			GitHeadCommit: freshness.GitHeadCommit,
		},
		Request: request,
		Target:  target,
		Window: repository.TargetedContextWindow{
			StartLine:      windowStart,
			EndLine:        windowEnd,
			TotalLines:     int64(len(lines)),
			StartTruncated: windowStart > 1,
			EndTruncated:   windowEnd < int64(len(lines)),
		},
		Content: strings.Join(windowLines, "\n"),
	}, nil
}

func (s ContextBlockService) openContextStore(ctx context.Context, startPath string) (repository.RepositoryRoot, *sqlite.Store, int64, error) {
	root, err := s.Locator.Resolve(startPath)
	if err != nil {
		return repository.RepositoryRoot{}, nil, 0, fmt.Errorf("resolve repository root: %w", err)
	}

	layoutResolver := s.ResolveLayout
	if layoutResolver == nil {
		layoutResolver = state.ResolveLayout
	}
	layout, err := layoutResolver(root.RootPath)
	if err != nil {
		return repository.RepositoryRoot{}, nil, 0, fmt.Errorf("resolve state layout: %w", err)
	}

	openStore := s.OpenStore
	if openStore == nil {
		openStore = func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		}
	}
	store, err := openStore(ctx, layout, root.DetectionMode)
	if err != nil {
		return repository.RepositoryRoot{}, nil, 0, fmt.Errorf("open state store: %w", err)
	}

	repositoryID, err := store.LookupRepositoryID(ctx, root.RootPath)
	if err != nil {
		_ = store.Close()
		if errors.Is(err, sql.ErrNoRows) {
			return repository.RepositoryRoot{}, nil, 0, fmt.Errorf("load repository metadata: %w", err)
		}
		return repository.RepositoryRoot{}, nil, 0, fmt.Errorf("load repository metadata: %w", err)
	}

	return root, store, repositoryID, nil
}

func validateTargetedContextRequest(request repository.TargetedContextRequest) error {
	if request.BeforeLines < 0 || request.AfterLines < 0 {
		return fmt.Errorf("load targeted context: before_lines and after_lines must be non-negative")
	}

	hasSymbolTarget := request.SymbolStableKey != ""
	hasLineTarget := request.Path != "" || request.StartLine != 0 || request.EndLine != 0
	if hasSymbolTarget == hasLineTarget {
		return fmt.Errorf("load targeted context: provide exactly one of symbol stable key or explicit line range")
	}
	if hasLineTarget {
		if request.Path == "" {
			return fmt.Errorf("load targeted context: path is required for explicit line ranges")
		}
		if request.StartLine < 1 || request.EndLine < request.StartLine {
			return fmt.Errorf("load targeted context: explicit line range must be 1-based and end at or after start")
		}
	}
	return nil
}

func splitContextLines(content string) []string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.TrimSuffix(normalized, "\n")
	if normalized == "" {
		return []string{}
	}
	return strings.Split(normalized, "\n")
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
