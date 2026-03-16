package mcp

import (
	"bytes"
	"context"
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

func TestMCPServerStdioSession(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\nfunc Alpha() {}\n")
	refreshRepo(t, repoRoot)

	var input bytes.Buffer
	writeTestFrame(t, &input, Request{
		JSONRPC: jsonRPCVersion,
		ID:      1,
		Method:  "initialize",
		Params: InitializeParams{
			ClientInfo:      ClientInfo{Name: "test-client", Version: "1.0.0"},
			ProtocolVersion: protocolVersion,
		},
	})
	writeTestFrame(t, &input, Request{JSONRPC: jsonRPCVersion, Method: "notifications/initialized"})
	writeTestFrame(t, &input, Request{
		JSONRPC: jsonRPCVersion,
		ID:      2,
		Method:  "tools/list",
	})
	writeTestFrame(t, &input, Request{
		JSONRPC: jsonRPCVersion,
		ID:      3,
		Method:  "tools/call",
		Params: CallToolParams{
			Name: toolRepositoryMap,
			Arguments: map[string]any{
				"startPath": repoRoot,
			},
		},
	})
	writeTestFrame(t, &input, Request{
		JSONRPC: jsonRPCVersion,
		ID:      4,
		Method:  "tools/call",
		Params: CallToolParams{
			Name: toolRefresh,
			Arguments: map[string]any{
				"startPath": repoRoot,
			},
		},
	})
	writeTestFrame(t, &input, Request{
		JSONRPC: jsonRPCVersion,
		ID:      5,
		Method:  "tools/call",
		Params: CallToolParams{
			Name: toolTokenTree,
			Arguments: map[string]any{
				"startPath": repoRoot,
				"maxNodes":  maxTokenTreeMaxNodes + 1,
			},
		},
	})

	var output bytes.Buffer
	var stderr bytes.Buffer
	if err := ServeStdio(context.Background(), &input, &output, &stderr); err != nil {
		t.Fatalf("ServeStdio() error = %v", err)
	}
	if stderr.String() != readinessMessage+"\n" {
		t.Fatalf("stderr = %q, want readiness signal", stderr.String())
	}
	if bytes.Contains(output.Bytes(), []byte(readinessMessage)) {
		t.Fatal("stdout included readiness signal")
	}

	responses := readTestResponses(t, &output)
	if len(responses) != 5 {
		t.Fatalf("response count = %d, want 5", len(responses))
	}

	var listResult ToolsListResult
	mustDecodeResult(t, responses[1].Result, &listResult)
	if !containsTool(listResult.Tools, toolRefresh) || !containsTool(listResult.Tools, toolPack) || !containsTool(listResult.Tools, toolHealth) {
		t.Fatalf("tools/list missing operational tools: %+v", listResult.Tools)
	}

	var repoCall CallToolResult
	mustDecodeResult(t, responses[2].Result, &repoCall)
	var repoEnvelope QueryEnvelope
	decodeStructuredContent(t, repoCall.StructuredContent, &repoEnvelope)
	if repoEnvelope.Meta.RepositoryRoot != repoRoot {
		t.Fatalf("repository map root = %q, want %q", repoEnvelope.Meta.RepositoryRoot, repoRoot)
	}
	if repoEnvelope.Meta.CacheStatus != cacheStatusPersistedOnly {
		t.Fatalf("repository map cache status = %q, want %q", repoEnvelope.Meta.CacheStatus, cacheStatusPersistedOnly)
	}

	var refreshCall CallToolResult
	mustDecodeResult(t, responses[3].Result, &refreshCall)
	var refreshEnvelope QueryEnvelope
	decodeStructuredContent(t, refreshCall.StructuredContent, &refreshEnvelope)
	if refreshEnvelope.Meta.CacheStatus != "refresh_attempted" {
		t.Fatalf("refresh cache status = %q, want refresh_attempted", refreshEnvelope.Meta.CacheStatus)
	}
	if refreshEnvelope.Meta.Generation == 0 {
		t.Fatalf("refresh generation = %d, want non-zero", refreshEnvelope.Meta.Generation)
	}
	if refreshCall.IsError {
		t.Fatal("refresh call unexpectedly marked as error")
	}

	if responses[4].Error == nil {
		t.Fatal("expected bounded token tree failure")
	}
	if responses[4].Error.Code != errCodeBounds {
		t.Fatalf("bounded failure code = %d, want %d", responses[4].Error.Code, errCodeBounds)
	}
	if responses[4].Error.Data["field"] != "maxNodes" {
		t.Fatalf("bounded failure field = %#v, want maxNodes", responses[4].Error.Data["field"])
	}
}
