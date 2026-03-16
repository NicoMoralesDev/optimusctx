package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
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

func TestBenchmarkContextAssemblyLane(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "internal", "http", "handler", "rollout.go"), "package handler\n\nfunc LoadRolloutConfig() string {\n\treturn \"prod\"\n}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "internal", "config", "loader.go"), "package config\n\nfunc Load() string {\n\treturn \"loader\"\n}\n")
	refreshRepo(t, repoRoot)

	runner := app.NewBenchmarkRunner()
	runner.RunCommand = nil
	bootstrapped := map[string]bool{}
	runner.RunTool = func(_ context.Context, invocation app.BenchmarkToolInvocation) (app.BenchmarkToolExecutionResult, error) {
		server := NewServer(nil, nil, nil)
		if !bootstrapped[invocation.WorkingDir] {
			refresh := callTool(t, server, CallToolParams{
				Name: toolRefresh,
				Arguments: map[string]any{
					"startPath": invocation.WorkingDir,
				},
			})
			if refresh.IsError {
				t.Fatalf("refresh bootstrap failed: %+v", refresh)
			}
			bootstrapped[invocation.WorkingDir] = true
		}
		call := callTool(t, server, CallToolParams{
			Name:      invocation.Name,
			Arguments: invocation.Arguments,
		})
		payload, err := decodeMCPBenchmarkPayload(call.StructuredContent)
		if err != nil {
			return app.BenchmarkToolExecutionResult{}, err
		}
		return app.BenchmarkToolExecutionResult{Payload: payload}, nil
	}
	runner.MkdirTemp = func(string, string) (string, error) {
		return t.TempDir(), nil
	}
	runner.CopyTree = func(src string, dst string) error {
		return copyMCPTree(t, src, dst)
	}

	benchmarksRoot := t.TempDir()
	copyMCPTree(t, filepath.Join("..", "..", "testdata", "eval", "fixtures"), filepath.Join(benchmarksRoot, "fixtures"))
	copyMCPTree(t, filepath.Join("..", "..", "testdata", "eval", "benchmarks"), filepath.Join(benchmarksRoot, "benchmarks"))

	result, err := runner.Run(context.Background(), app.BenchmarkRunRequest{
		SuiteID:      "go-benchmark-discovery-v1",
		SuitesDir:    filepath.Join(benchmarksRoot, "benchmarks"),
		FixturesRoot: filepath.Join(benchmarksRoot, "fixtures"),
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	contextLane := result.Arms[1].LaneResults[1]
	if contextLane.StopMarker != "context_ready" || !contextLane.Success {
		t.Fatalf("context lane = %+v", contextLane)
	}
	if contextLane.Effort.FileReadActions == 0 || contextLane.Effort.BytesRead == 0 {
		t.Fatalf("context effort = %+v", contextLane.Effort)
	}
}

func TestBenchmarkRerunsDeterministic(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "internal", "http", "handler", "rollout.go"), "package handler\n\nfunc LoadRolloutConfig() string {\n\treturn \"prod\"\n}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "internal", "config", "loader.go"), "package config\n\nfunc Load() string {\n\treturn \"loader\"\n}\n")
	refreshRepo(t, repoRoot)

	service := app.NewBenchmarkService()
	service.Runner = app.NewBenchmarkRunner()
	service.Runner.RunCommand = nil
	bootstrapped := map[string]bool{}
	service.Runner.RunTool = func(_ context.Context, invocation app.BenchmarkToolInvocation) (app.BenchmarkToolExecutionResult, error) {
		server := NewServer(nil, nil, nil)
		if !bootstrapped[invocation.WorkingDir] {
			refresh := callTool(t, server, CallToolParams{
				Name: toolRefresh,
				Arguments: map[string]any{
					"startPath": invocation.WorkingDir,
				},
			})
			if refresh.IsError {
				t.Fatalf("refresh bootstrap failed: %+v", refresh)
			}
			bootstrapped[invocation.WorkingDir] = true
		}
		call := callTool(t, server, CallToolParams{
			Name:      invocation.Name,
			Arguments: invocation.Arguments,
		})
		payload, err := decodeMCPBenchmarkPayload(call.StructuredContent)
		if err != nil {
			return app.BenchmarkToolExecutionResult{}, err
		}
		return app.BenchmarkToolExecutionResult{Payload: payload}, nil
	}
	service.Runner.MkdirTemp = func(string, string) (string, error) {
		return t.TempDir(), nil
	}
	service.Runner.CopyTree = func(src string, dst string) error {
		return copyMCPTree(t, src, dst)
	}

	benchmarksRoot := t.TempDir()
	copyMCPTree(t, filepath.Join("..", "..", "testdata", "eval", "fixtures"), filepath.Join(benchmarksRoot, "fixtures"))
	copyMCPTree(t, filepath.Join("..", "..", "testdata", "eval", "benchmarks"), filepath.Join(benchmarksRoot, "benchmarks"))

	result, err := service.RunRepeated(context.Background(), app.BenchmarkRepeatedRunRequest{
		StartPath:    repoRoot,
		SuiteID:      "go-benchmark-discovery-v1",
		SuitesDir:    filepath.Join(benchmarksRoot, "benchmarks"),
		FixturesRoot: filepath.Join(benchmarksRoot, "fixtures"),
		Attempts:     2,
	})
	if err != nil {
		t.Fatalf("RunRepeated() error = %v", err)
	}
	if !result.Summary.Verification.Passed {
		t.Fatalf("verification = %+v", result.Summary.Verification)
	}
	if got, want := result.Summary.AttemptCount, 2; got != want {
		t.Fatalf("summary attempts = %d, want %d", got, want)
	}
	if got, want := result.Attempts[0].Result.Arms[1].LaneResults[0].Lane, repository.BenchmarkLaneDiscovery; got != want {
		t.Fatalf("lane identity = %q, want %q", got, want)
	}
}

func TestBenchmarkArtifactAttribution(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "internal", "http", "handler", "rollout.go"), "package handler\n\nfunc LoadRolloutConfig() string {\n\treturn \"prod\"\n}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "internal", "config", "loader.go"), "package config\n\nfunc Load() string {\n\treturn \"loader\"\n}\n")
	refreshRepo(t, repoRoot)

	service := app.NewBenchmarkService()
	service.Runner = app.NewBenchmarkRunner()
	service.Runner.RunCommand = nil
	bootstrapped := map[string]bool{}
	service.Runner.RunTool = func(_ context.Context, invocation app.BenchmarkToolInvocation) (app.BenchmarkToolExecutionResult, error) {
		server := NewServer(nil, nil, nil)
		if !bootstrapped[invocation.WorkingDir] {
			refresh := callTool(t, server, CallToolParams{
				Name: toolRefresh,
				Arguments: map[string]any{
					"startPath": invocation.WorkingDir,
				},
			})
			if refresh.IsError {
				t.Fatalf("refresh bootstrap failed: %+v", refresh)
			}
			bootstrapped[invocation.WorkingDir] = true
		}
		call := callTool(t, server, CallToolParams{
			Name:      invocation.Name,
			Arguments: invocation.Arguments,
		})
		payload, err := decodeMCPBenchmarkPayload(call.StructuredContent)
		if err != nil {
			return app.BenchmarkToolExecutionResult{}, err
		}
		return app.BenchmarkToolExecutionResult{Payload: payload}, nil
	}
	service.Runner.MkdirTemp = func(string, string) (string, error) { return t.TempDir(), nil }
	service.Runner.CopyTree = func(src string, dst string) error { return copyMCPTree(t, src, dst) }

	benchmarksRoot := t.TempDir()
	copyMCPTree(t, filepath.Join("..", "..", "testdata", "eval", "fixtures"), filepath.Join(benchmarksRoot, "fixtures"))
	copyMCPTree(t, filepath.Join("..", "..", "testdata", "eval", "benchmarks"), filepath.Join(benchmarksRoot, "benchmarks"))

	if _, err := service.RunRepeated(context.Background(), app.BenchmarkRepeatedRunRequest{
		StartPath:    repoRoot,
		SuiteID:      "go-benchmark-discovery-v1",
		SuitesDir:    filepath.Join(benchmarksRoot, "benchmarks"),
		FixturesRoot: filepath.Join(benchmarksRoot, "fixtures"),
		Attempts:     1,
	}); err != nil {
		t.Fatalf("RunRepeated() error = %v", err)
	}

	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}
	store, err := sqlite.OpenOrCreateStore(context.Background(), layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("OpenOrCreateStore() error = %v", err)
	}
	defer store.Close()

	repositoryID, err := store.LookupRepositoryID(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("LookupRepositoryID() error = %v", err)
	}
	runs, err := store.ListBenchmarkRuns(context.Background(), repositoryID, "go-benchmark-discovery-v1", "v1")
	if err != nil {
		t.Fatalf("ListBenchmarkRuns() error = %v", err)
	}

	var found bool
	for _, run := range runs {
		if run.Run.ArmKind != repository.BenchmarkArmKindOptimusCtx {
			continue
		}
		for _, sample := range run.Samples {
			if sample.Sample.Lane != repository.BenchmarkLaneDiscovery {
				continue
			}
			var metadata struct {
				Attribution []repository.BenchmarkArtifactConsumption `json:"attribution"`
			}
			if err := json.Unmarshal([]byte(sample.Sample.MetadataJSON), &metadata); err != nil {
				t.Fatalf("sample metadata json: %v", err)
			}
			if len(metadata.Attribution) < 2 {
				t.Fatalf("discovery attribution = %+v, want repository_map and exact_lookup", metadata.Attribution)
			}
			if metadata.Attribution[0].ReportLabel != repository.BenchmarkReportArtifactLabelRepositoryMap {
				t.Fatalf("first report label = %q", metadata.Attribution[0].ReportLabel)
			}
			if metadata.Attribution[1].ReportLabel != repository.BenchmarkReportArtifactLabelExactLookup {
				t.Fatalf("second report label = %q", metadata.Attribution[1].ReportLabel)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("did not find persisted mcp discovery attribution")
	}
}

func TestBenchmarkTaskCompletionComparison(t *testing.T) {
	runner := app.NewBenchmarkRunner()
	runner.RunCommand = func(_ context.Context, invocation app.BenchmarkCommandInvocation) (app.BenchmarkCommandExecutionResult, error) {
		switch {
		case len(invocation.Args) > 0 && invocation.Args[0] == "refresh":
			return app.BenchmarkCommandExecutionResult{ExitCode: 0}, nil
		case len(invocation.Args) >= 5 && invocation.Args[0] == "pack" && invocation.Args[1] == "export":
			writeRepoFile(t, filepath.Join(invocation.WorkingDir, "artifacts", "pack.json"), "{\"documents\":[\"docs/notes.txt\"],\"status\":\"ok\"}\n")
			return app.BenchmarkCommandExecutionResult{ExitCode: 0}, nil
		default:
			return app.BenchmarkCommandExecutionResult{ExitCode: 0}, nil
		}
	}
	bootstrapped := map[string]bool{}
	runner.RunTool = func(_ context.Context, invocation app.BenchmarkToolInvocation) (app.BenchmarkToolExecutionResult, error) {
		server := NewServer(nil, nil, nil)
		if !bootstrapped[invocation.WorkingDir] {
			refresh := callTool(t, server, CallToolParams{
				Name: toolRefresh,
				Arguments: map[string]any{
					"startPath": invocation.WorkingDir,
				},
			})
			if refresh.IsError {
				t.Fatalf("refresh bootstrap failed: %+v", refresh)
			}
			bootstrapped[invocation.WorkingDir] = true
		}
		call := callTool(t, server, CallToolParams{
			Name:      invocation.Name,
			Arguments: invocation.Arguments,
		})
		payload, err := decodeMCPBenchmarkPayload(call.StructuredContent)
		if err != nil {
			return app.BenchmarkToolExecutionResult{}, err
		}
		return app.BenchmarkToolExecutionResult{Payload: payload}, nil
	}
	runner.MkdirTemp = func(string, string) (string, error) {
		return t.TempDir(), nil
	}
	runner.CopyTree = func(src string, dst string) error {
		return copyMCPTree(t, src, dst)
	}

	benchmarksRoot := t.TempDir()
	copyMCPTree(t, filepath.Join("..", "..", "testdata", "eval", "fixtures"), filepath.Join(benchmarksRoot, "fixtures"))
	copyMCPTree(t, filepath.Join("..", "..", "testdata", "eval", "benchmarks"), filepath.Join(benchmarksRoot, "benchmarks"))

	result, err := runner.Run(context.Background(), app.BenchmarkRunRequest{
		SuiteID:      "go-benchmark-refresh-v1",
		SuitesDir:    filepath.Join(benchmarksRoot, "benchmarks"),
		FixturesRoot: filepath.Join(benchmarksRoot, "fixtures"),
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	taskLane := result.Arms[1].LaneResults[1]
	if !taskLane.Success || taskLane.StopMarker != "task_complete" {
		t.Fatalf("task lane = %+v", taskLane)
	}
	if taskLane.Effort.ActionCount < 2 {
		t.Fatalf("task effort = %+v, want MCP preview plus pack export", taskLane.Effort)
	}
	if !slices.Contains(taskLane.EvidencePaths, "artifacts/pack.json") {
		t.Fatalf("task evidence = %+v", taskLane.EvidencePaths)
	}
}

func copyMCPTree(t *testing.T, src string, dst string) error {
	t.Helper()

	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyMCPTree(t, srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dstPath, content, info.Mode().Perm()); err != nil {
			return err
		}
	}
	return nil
}

func decodeMCPBenchmarkPayload(value any) (any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var envelope QueryEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}
