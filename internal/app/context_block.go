package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

type ContextBlockService struct {
	Lookup   LookupService
	ReadFile func(string) ([]byte, error)
}

func NewContextBlockService() ContextBlockService {
	return ContextBlockService{
		Lookup:   NewLookupService(),
		ReadFile: os.ReadFile,
	}
}

func (s ContextBlockService) TargetedContext(ctx context.Context, startPath string, request repository.TargetedContextRequest) (repository.TargetedContextResult, error) {
	lookup := s.Lookup
	if lookup.Locator == (repository.Locator{}) {
		lookup = NewLookupService()
	}

	store, repositoryID, err := lookup.openLookupStore(ctx, startPath)
	if err != nil {
		return repository.TargetedContextResult{}, err
	}
	defer store.Close()

	freshness, err := store.ReadRepositoryFreshness(ctx, repositoryID)
	if err != nil {
		return repository.TargetedContextResult{}, fmt.Errorf("load targeted context freshness: %w", err)
	}

	path, anchorStart, anchorEnd, err := resolveTargetRequest(ctx, lookup, store, repositoryID, request)
	if err != nil {
		return repository.TargetedContextResult{}, err
	}

	readFile := s.ReadFile
	if readFile == nil {
		readFile = os.ReadFile
	}
	data, err := readFile(filepath.Join(freshness.RootPath, path))
	if err != nil {
		return repository.TargetedContextResult{}, fmt.Errorf("read targeted context file %q: %w", path, err)
	}

	lines := splitLines(string(data))
	if anchorStart < 1 || anchorEnd < anchorStart {
		return repository.TargetedContextResult{}, fmt.Errorf("load targeted context: invalid anchor range %d-%d", anchorStart, anchorEnd)
	}
	if anchorEnd > len(lines) {
		return repository.TargetedContextResult{}, fmt.Errorf("load targeted context: live file %q no longer satisfies indexed anchor %d-%d", path, anchorStart, anchorEnd)
	}

	startLine := maxInt(1, anchorStart-request.BeforeLines)
	endLine := minInt(len(lines), anchorEnd+request.AfterLines)
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
		Request:        request,
		Path:           path,
		AnchorStart:    anchorStart,
		AnchorEnd:      anchorEnd,
		StartLine:      startLine,
		EndLine:        endLine,
		BeforeLines:    request.BeforeLines,
		AfterLines:     request.AfterLines,
		TruncatedStart: startLine > 1,
		TruncatedEnd:   endLine < len(lines),
		Source:         append([]string(nil), lines[startLine-1:endLine]...),
	}, nil
}

func resolveTargetRequest(ctx context.Context, lookup LookupService, store *sqlite.Store, repositoryID int64, request repository.TargetedContextRequest) (string, int, int, error) {
	if request.BeforeLines < 0 || request.AfterLines < 0 {
		return "", 0, 0, fmt.Errorf("load targeted context: before/after lines must be non-negative")
	}

	if request.StableKey != "" {
		if request.Path != "" || request.StartLine != 0 || request.EndLine != 0 {
			return "", 0, 0, fmt.Errorf("load targeted context: stable key requests cannot also specify an explicit line range")
		}
		anchor, err := lookup.loadSymbolAnchor(ctx, store, repositoryID, strings.TrimSpace(request.StableKey))
		if err != nil {
			return "", 0, 0, fmt.Errorf("load targeted context: %w", err)
		}
		return anchor.Path, int(anchor.StartRow) + 1, int(anchor.EndRow) + 1, nil
	}

	request.Path = strings.TrimSpace(request.Path)
	if request.Path == "" || request.StartLine <= 0 || request.EndLine < request.StartLine {
		return "", 0, 0, fmt.Errorf("load targeted context: explicit path and valid start/end lines are required")
	}
	return request.Path, request.StartLine, request.EndLine, nil
}

func splitLines(contents string) []string {
	normalized := strings.ReplaceAll(contents, "\r\n", "\n")
	if normalized == "" {
		return []string{""}
	}
	lines := strings.Split(normalized, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
