package mcp

import (
	"path/filepath"
	"testing"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

func TestMCPRefreshPackHealth(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\ntype Alpha struct{}\n\nfunc (Alpha) Run() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "docs", "guide.go"), "package docs\n\nfunc Guide() {}\n")
	refreshRepo(t, repoRoot)

	server := NewServer(nil, nil, nil)

	refreshResult := callTool(t, server, CallToolParams{
		Name: toolRefresh,
		Arguments: map[string]any{
			"startPath": repoRoot,
		},
	})
	var refreshEnvelope QueryEnvelope
	decodeStructuredContent(t, refreshResult.StructuredContent, &refreshEnvelope)
	var refreshPayload app.RefreshResult
	decodeStructuredContent(t, refreshEnvelope.Data, &refreshPayload)
	if refreshEnvelope.Meta.CacheStatus != "refresh_attempted" {
		t.Fatalf("refresh cache status = %q, want refresh_attempted", refreshEnvelope.Meta.CacheStatus)
	}
	if refreshEnvelope.Meta.Generation == 0 || refreshPayload.Generation == 0 {
		t.Fatalf("refresh envelope = %+v payload = %+v", refreshEnvelope.Meta, refreshPayload)
	}

	tokenTreeResult := callTool(t, server, CallToolParams{
		Name: toolTokenTree,
		Arguments: map[string]any{
			"startPath": repoRoot,
			"maxDepth":  2,
			"maxNodes":  8,
		},
	})
	var tokenTreeEnvelope QueryEnvelope
	decodeStructuredContent(t, tokenTreeResult.StructuredContent, &tokenTreeEnvelope)
	var tokenTreePayload repository.TokenTreeResult
	decodeStructuredContent(t, tokenTreeEnvelope.Data, &tokenTreePayload)
	if tokenTreePayload.Root.Path != "." {
		t.Fatalf("token tree root path = %q, want .", tokenTreePayload.Root.Path)
	}

	healthResult := callTool(t, server, CallToolParams{
		Name: toolHealth,
		Arguments: map[string]any{
			"startPath": repoRoot,
		},
	})
	var healthEnvelope QueryEnvelope
	decodeStructuredContent(t, healthResult.StructuredContent, &healthEnvelope)
	var healthPayload repository.HealthResult
	decodeStructuredContent(t, healthEnvelope.Data, &healthPayload)
	if !healthPayload.Summary.Initialized {
		t.Fatalf("health summary = %+v", healthPayload.Summary)
	}

	packResult := callTool(t, server, CallToolParams{
		Name: toolPack,
		Arguments: map[string]any{
			"startPath":                repoRoot,
			"includeRepositoryContext": true,
			"includeStructuralContext": true,
		},
	})
	var packEnvelope QueryEnvelope
	decodeStructuredContent(t, packResult.StructuredContent, &packEnvelope)
	var packPayload repository.PackResult
	decodeStructuredContent(t, packEnvelope.Data, &packPayload)
	if packPayload.Summary.ReturnedSectionCount != 2 {
		t.Fatalf("pack summary = %+v", packPayload.Summary)
	}
}
