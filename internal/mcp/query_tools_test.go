package mcp

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

func TestMCPRepositoryQueries(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "README.md"), "# Repo\n")
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\ntype Alpha struct{}\n\nfunc Beta() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "partial.go"), "package pkg\n\nfunc Healthy() {}\nfunc (\n")
	writeRepoFile(t, filepath.Join(repoRoot, "docs", "guide.go"), "package docs\n\nfunc Guide() {}\n")

	refreshRepo(t, repoRoot)
	server := NewServer(nil, nil, nil)

	repositoryResult := callTool(t, server, CallToolParams{
		Name: toolRepositoryMap,
		Arguments: map[string]any{
			"startPath":         repoRoot,
			"directoryLimit":    10,
			"filesPerDirectory": 1,
			"symbolsPerFile":    1,
		},
	})
	var repositoryEnvelope QueryEnvelope
	decodeStructuredContent(t, repositoryResult.StructuredContent, &repositoryEnvelope)

	if repositoryEnvelope.Meta.RepositoryRoot != repoRoot {
		t.Fatalf("repository root = %q, want %q", repositoryEnvelope.Meta.RepositoryRoot, repoRoot)
	}
	if repositoryEnvelope.Meta.CacheStatus != cacheStatusPersistedOnly {
		t.Fatalf("cache status = %q, want %q", repositoryEnvelope.Meta.CacheStatus, cacheStatusPersistedOnly)
	}
	if !repositoryEnvelope.Meta.Bounds.Truncated {
		t.Fatalf("repository bounds = %+v, want truncated", repositoryEnvelope.Meta.Bounds)
	}

	var repositoryMap repository.RepositoryMap
	decodeStructuredContent(t, repositoryEnvelope.Data, &repositoryMap)
	if len(repositoryMap.Directories) < 3 {
		t.Fatalf("directory count = %d, want at least 3", len(repositoryMap.Directories))
	}
	pkgDirectory := findRepositoryDirectory(t, repositoryMap.Directories, "pkg")
	if len(pkgDirectory.Files) != 1 {
		t.Fatalf("pkg file count = %d, want 1", len(pkgDirectory.Files))
	}
	if len(pkgDirectory.Files[0].Symbols) != 1 {
		t.Fatalf("pkg symbol count = %d, want 1", len(pkgDirectory.Files[0].Symbols))
	}

	contextResult := callTool(t, server, CallToolParams{
		Name: toolLayeredContextL1,
		Arguments: map[string]any{
			"startPath": repoRoot,
		},
	})
	var contextEnvelope QueryEnvelope
	decodeStructuredContent(t, contextResult.StructuredContent, &contextEnvelope)
	var layered repository.LayeredContextL1
	decodeStructuredContent(t, contextEnvelope.Data, &layered)

	if contextEnvelope.Meta.RepositoryRoot != repoRoot {
		t.Fatalf("context repository root = %q, want %q", contextEnvelope.Meta.RepositoryRoot, repoRoot)
	}
	if contextEnvelope.Meta.Freshness != string(repository.FreshnessStatusFresh) {
		t.Fatalf("context freshness = %q, want %q", contextEnvelope.Meta.Freshness, repository.FreshnessStatusFresh)
	}
	if contextEnvelope.Meta.Bounds.AppliedLimit != layered.Limits.FileLimit {
		t.Fatalf("context bounds = %+v, layered limits = %+v", contextEnvelope.Meta.Bounds, layered.Limits)
	}
	if len(layered.Candidates) == 0 {
		t.Fatal("expected layered context candidates")
	}
}

func TestMCPLookupQueries(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\ntype Alpha struct{}\n\nfunc (Alpha) Run() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "docs", "alpha.go"), "package docs\n\nfunc Alpha() {}\n")

	refreshRepo(t, repoRoot)
	server := NewServer(nil, nil, nil)

	symbolResult := callTool(t, server, CallToolParams{
		Name: toolSymbolLookup,
		Arguments: map[string]any{
			"startPath": repoRoot,
			"name":      "Alpha",
			"limit":     2,
		},
	})
	var symbolEnvelope QueryEnvelope
	decodeStructuredContent(t, symbolResult.StructuredContent, &symbolEnvelope)
	var symbolLookup repository.SymbolLookupResult
	decodeStructuredContent(t, symbolEnvelope.Data, &symbolLookup)

	if symbolEnvelope.Meta.CacheStatus != cacheStatusPersistedOnly {
		t.Fatalf("symbol cache status = %q", symbolEnvelope.Meta.CacheStatus)
	}
	if len(symbolLookup.Matches) != 2 {
		t.Fatalf("symbol matches = %d, want 2", len(symbolLookup.Matches))
	}
	if symbolLookup.Matches[0].Path != "docs/alpha.go" || symbolLookup.Matches[1].Path != "pkg/alpha.go" {
		t.Fatalf("symbol match order = %+v", symbolLookup.Matches)
	}
	if !symbolEnvelope.Meta.Bounds.LimitReached {
		t.Fatalf("symbol bounds = %+v, want limit reached", symbolEnvelope.Meta.Bounds)
	}

	structureResult := callTool(t, server, CallToolParams{
		Name: toolStructureLookup,
		Arguments: map[string]any{
			"startPath":  repoRoot,
			"kind":       "method",
			"parentName": "Alpha",
			"pathPrefix": "pkg/",
		},
	})
	var structureEnvelope QueryEnvelope
	decodeStructuredContent(t, structureResult.StructuredContent, &structureEnvelope)
	var structureLookup repository.StructureLookupResult
	decodeStructuredContent(t, structureEnvelope.Data, &structureLookup)

	if len(structureLookup.Matches) != 1 {
		t.Fatalf("structure matches = %d, want 1", len(structureLookup.Matches))
	}
	if structureLookup.Matches[0].Path != "pkg/alpha.go" {
		t.Fatalf("structure path = %q, want pkg/alpha.go", structureLookup.Matches[0].Path)
	}
	if structureEnvelope.Meta.Freshness != string(repository.FreshnessStatusFresh) {
		t.Fatalf("structure freshness = %q", structureEnvelope.Meta.Freshness)
	}
}

func TestMCPBoundedFailures(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\nfunc Alpha() {}\n")
	refreshRepo(t, repoRoot)
	server := NewServer(nil, nil, nil)

	lookupError := callToolError(t, server, CallToolParams{
		Name: toolSymbolLookup,
		Arguments: map[string]any{
			"startPath": repoRoot,
			"name":      "Alpha",
			"limit":     maxLookupLimit + 1,
		},
	})
	if lookupError.Code != errCodeBounds {
		t.Fatalf("lookup error code = %d, want %d", lookupError.Code, errCodeBounds)
	}
	if lookupError.Data["field"] != "limit" {
		t.Fatalf("lookup error field = %#v, want limit", lookupError.Data["field"])
	}
	if lookupError.Data["max"] != maxLookupLimit {
		t.Fatalf("lookup error max = %#v, want %d", lookupError.Data["max"], maxLookupLimit)
	}

	contextError := callToolError(t, server, CallToolParams{
		Name: toolTargetedContext,
		Arguments: map[string]any{
			"startPath":   repoRoot,
			"path":        "pkg/alpha.go",
			"startLine":   2,
			"endLine":     2,
			"beforeLines": maxContextWindowLines + 1,
		},
	})
	if contextError.Code != errCodeBounds {
		t.Fatalf("context error code = %d, want %d", contextError.Code, errCodeBounds)
	}
	if contextError.Data["field"] != "beforeLines" {
		t.Fatalf("context error field = %#v, want beforeLines", contextError.Data["field"])
	}
}

func TestMCPStructuredErrors(t *testing.T) {
	server := NewServer(nil, nil, nil)

	requiredName := callToolError(t, server, CallToolParams{
		Name: toolSymbolLookup,
		Arguments: map[string]any{
			"limit": 1,
		},
	})
	if requiredName.Code != errCodeValidation {
		t.Fatalf("required name code = %d, want %d", requiredName.Code, errCodeValidation)
	}
	if requiredName.Data["field"] != "name" {
		t.Fatalf("required name field = %#v, want name", requiredName.Data["field"])
	}
	if requiredName.Data["constraint"] != "required" {
		t.Fatalf("required name constraint = %#v, want required", requiredName.Data["constraint"])
	}

	conflict := callToolError(t, server, CallToolParams{
		Name: toolTargetedContext,
		Arguments: map[string]any{
			"stableKey": "sym::Alpha",
			"path":      "pkg/alpha.go",
			"startLine": 1,
			"endLine":   2,
		},
	})
	if conflict.Code != errCodeValidation {
		t.Fatalf("conflict code = %d, want %d", conflict.Code, errCodeValidation)
	}
	if conflict.Data["field"] != "stableKey" {
		t.Fatalf("conflict field = %#v, want stableKey", conflict.Data["field"])
	}
	if conflict.Data["constraint"] != "conflict" {
		t.Fatalf("conflict constraint = %#v, want conflict", conflict.Data["constraint"])
	}
}

func callTool(t *testing.T, server *Server, params CallToolParams) CallToolResult {
	t.Helper()

	response := server.handleRequest(context.Background(), Request{
		JSONRPC: jsonRPCVersion,
		ID:      1,
		Method:  "tools/call",
		Params:  params,
	})
	if response == nil {
		t.Fatal("expected response")
	}
	if response.Error != nil {
		t.Fatalf("unexpected tool error: %+v", response.Error)
	}

	var result CallToolResult
	decodeStructuredContent(t, response.Result, &result)
	return result
}

func callToolError(t *testing.T, server *Server, params CallToolParams) *ResponseError {
	t.Helper()

	response := server.handleRequest(context.Background(), Request{
		JSONRPC: jsonRPCVersion,
		ID:      1,
		Method:  "tools/call",
		Params:  params,
	})
	if response == nil {
		t.Fatal("expected response")
	}
	if response.Error == nil {
		t.Fatal("expected tool error")
	}
	return response.Error
}

func decodeStructuredContent(t *testing.T, raw any, target any) {
	t.Helper()

	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	if err := json.Unmarshal(payload, target); err != nil {
		t.Fatalf("unmarshal structured content: %v", err)
	}
}

func refreshRepo(t *testing.T, repoRoot string) {
	t.Helper()

	if _, err := app.NewRefreshService().Refresh(context.Background(), app.RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
}

func initRepo(t *testing.T) string {
	t.Helper()

	repoRoot := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, output)
	}

	if err := os.MkdirAll(filepath.Join(repoRoot, ".git", "info"), 0o755); err != nil {
		t.Fatalf("mkdir git info: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	return repoRoot
}

func writeRepoFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func findRepositoryDirectory(t *testing.T, directories []repository.RepositoryMapDirectory, path string) repository.RepositoryMapDirectory {
	t.Helper()

	for _, directory := range directories {
		if directory.Path == path {
			return directory
		}
	}

	t.Fatalf("repository directory %q not found", path)
	return repository.RepositoryMapDirectory{}
}
