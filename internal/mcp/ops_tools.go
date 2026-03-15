package mcp

import (
	"context"
	"strings"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

const (
	defaultTokenTreeMaxDepth = 3
	maxTokenTreeMaxDepth     = 8
	defaultTokenTreeMaxNodes = 64
	maxTokenTreeMaxNodes     = 256
	maxRefreshChangedHint    = 128
	maxPackSectionsLimit     = 8
	maxPackTargetsLimit      = 4
	maxPackLookupLimit       = 4
	maxPackContextLines      = 20
	maxPackTargetRangeLimit  = 40
)

type operationalToolServices struct {
	refresh   app.RefreshService
	tokenTree app.TokenTreeService
	pack      app.PackService
	health    app.HealthService
}

func defaultOperationalToolServices() operationalToolServices {
	return operationalToolServices{
		refresh:   app.NewRefreshService(),
		tokenTree: app.NewTokenTreeService(),
		pack:      app.NewPackService(),
		health:    app.NewHealthService(),
	}
}

func operationalToolHandlers(services operationalToolServices) []ToolHandler {
	return []ToolHandler{
		newRefreshTool(services),
		newTokenTreeTool(services),
		newPackTool(services),
		newHealthTool(services),
	}
}

func newRefreshTool(services operationalToolServices) ToolHandler {
	type request struct {
		StartPath   string   `json:"startPath"`
		ForceFull   bool     `json:"forceFull"`
		ChangedHint []string `json:"changedHint"`
	}

	return ToolHandler{
		Tool: Tool{
			Name:        toolRefresh,
			Description: "Run the manual refresh pipeline and return explicit generation and freshness state.",
			InputSchema: objectSchema(
				propertySchema("startPath", "string"),
				propertySchema("forceFull", "boolean"),
				map[string]any{"changedHint": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}},
			),
		},
		Call: func(ctx context.Context, params CallToolParams) (CallToolResult, *ResponseError) {
			var req request
			if err := decodeToolArguments(params.Arguments, &req); err != nil {
				return CallToolResult{}, newValidationError("invalid tool arguments", FieldErrorDetail{
					Field:      "arguments",
					Constraint: "invalid",
					Message:    err.Error(),
				})
			}
			if len(req.ChangedHint) > maxRefreshChangedHint {
				return CallToolResult{}, newMaximumExceededError("changedHint", maxRefreshChangedHint, len(req.ChangedHint))
			}

			result, err := services.refresh.Refresh(ctx, app.RefreshRequest{
				StartPath:   defaultStartPath(req.StartPath),
				Reason:      repository.RefreshReasonManual,
				ForceFull:   req.ForceFull,
				ChangedHint: req.ChangedHint,
			})
			if err != nil {
				return CallToolResult{}, mapOperationalToolError(err)
			}

			meta := QueryResultMetadata{
				RepositoryRoot: result.RepositoryRoot,
				Generation:     result.Generation,
				Freshness:      string(result.FreshnessStatus),
				CacheStatus:    "refresh_attempted",
			}
			return newStructuredToolResult(meta, result)
		},
	}
}

func newTokenTreeTool(services operationalToolServices) ToolHandler {
	type request struct {
		StartPath  string `json:"startPath"`
		PathPrefix string `json:"pathPrefix"`
		MaxDepth   int    `json:"maxDepth"`
		MaxNodes   int    `json:"maxNodes"`
	}

	return ToolHandler{
		Tool: Tool{
			Name:        toolTokenTree,
			Description: "Return a bounded hierarchical token tree derived from persisted repository state.",
			InputSchema: objectSchema(
				propertySchema("startPath", "string"),
				propertySchema("pathPrefix", "string"),
				propertySchema("maxDepth", "integer"),
				propertySchema("maxNodes", "integer"),
			),
		},
		Call: func(ctx context.Context, params CallToolParams) (CallToolResult, *ResponseError) {
			var req request
			if err := decodeToolArguments(params.Arguments, &req); err != nil {
				return CallToolResult{}, newValidationError("invalid tool arguments", FieldErrorDetail{
					Field:      "arguments",
					Constraint: "invalid",
					Message:    err.Error(),
				})
			}

			maxDepth, err := normalizeLimit("maxDepth", req.MaxDepth, defaultTokenTreeMaxDepth, maxTokenTreeMaxDepth)
			if err != nil {
				return CallToolResult{}, err
			}
			maxNodes, err := normalizeLimit("maxNodes", req.MaxNodes, defaultTokenTreeMaxNodes, maxTokenTreeMaxNodes)
			if err != nil {
				return CallToolResult{}, err
			}

			result, callErr := services.tokenTree.Analyze(ctx, defaultStartPath(req.StartPath), repository.TokenTreeRequest{
				PathPrefix: req.PathPrefix,
				MaxDepth:   maxDepth,
				MaxNodes:   maxNodes,
			})
			if callErr != nil {
				return CallToolResult{}, mapOperationalToolError(callErr)
			}

			meta := queryMetaFromEnvelope(result.Repository)
			meta.Bounds = QueryBoundsMetadata{
				DefaultLimit:   defaultTokenTreeMaxNodes,
				MaxLimit:       maxTokenTreeMaxNodes,
				RequestedLimit: req.MaxNodes,
				AppliedLimit:   result.Bounds.MaxNodes,
				ReturnedCount:  result.Summary.ReturnedNodeCount,
				TotalCount:     result.Summary.TotalNodeCount,
				Truncated:      result.Summary.Truncated,
			}
			return newStructuredToolResult(meta, result)
		},
	}
}

func newPackTool(services operationalToolServices) ToolHandler {
	type request struct {
		StartPath         string                              `json:"startPath"`
		IncludeRepository bool                                `json:"includeRepositoryContext"`
		IncludeStructural bool                                `json:"includeStructuralContext"`
		SymbolLookups     []repository.SymbolLookupRequest    `json:"symbolLookups"`
		StructureLookups  []repository.StructureLookupRequest `json:"structureLookups"`
		Targets           []repository.TargetedContextRequest `json:"targets"`
	}

	return ToolHandler{
		Tool: Tool{
			Name:        toolPack,
			Description: "Assemble a bounded deterministic context pack from existing query surfaces.",
			InputSchema: objectSchema(
				propertySchema("startPath", "string"),
				propertySchema("includeRepositoryContext", "boolean"),
				propertySchema("includeStructuralContext", "boolean"),
				map[string]any{"symbolLookups": map[string]any{"type": "array", "items": map[string]any{"type": "object"}}},
				map[string]any{"structureLookups": map[string]any{"type": "array", "items": map[string]any{"type": "object"}}},
				map[string]any{"targets": map[string]any{"type": "array", "items": map[string]any{"type": "object"}}},
			),
		},
		Call: func(ctx context.Context, params CallToolParams) (CallToolResult, *ResponseError) {
			var req request
			if err := decodeToolArguments(params.Arguments, &req); err != nil {
				return CallToolResult{}, newValidationError("invalid tool arguments", FieldErrorDetail{
					Field:      "arguments",
					Constraint: "invalid",
					Message:    err.Error(),
				})
			}

			result, err := services.pack.Pack(ctx, defaultStartPath(req.StartPath), repository.PackRequest{
				IncludeRepositoryContext: req.IncludeRepository,
				IncludeStructuralContext: req.IncludeStructural,
				SymbolLookups:            req.SymbolLookups,
				StructureLookups:         req.StructureLookups,
				Targets:                  req.Targets,
			})
			if err != nil {
				return CallToolResult{}, mapOperationalToolError(err)
			}

			meta := queryMetaFromEnvelope(result.Repository)
			meta.Bounds = QueryBoundsMetadata{
				DefaultLimit:   result.Bounds.MaxSections,
				MaxLimit:       result.Bounds.MaxSections,
				ReturnedCount:  result.Summary.ReturnedSectionCount,
				TotalCount:     int64(result.Summary.RequestedSectionCount),
				Truncated:      result.Summary.ReturnedSectionCount < result.Summary.RequestedSectionCount,
				RequestedLimit: result.Summary.RequestedSectionCount,
				AppliedLimit:   result.Bounds.MaxSections,
			}
			return newStructuredToolResult(meta, result)
		},
	}
}

func newHealthTool(services operationalToolServices) ToolHandler {
	type request struct {
		StartPath string `json:"startPath"`
	}

	return ToolHandler{
		Tool: Tool{
			Name:        toolHealth,
			Description: "Return structured repository state and freshness diagnostics.",
			InputSchema: objectSchema(propertySchema("startPath", "string")),
		},
		Call: func(ctx context.Context, params CallToolParams) (CallToolResult, *ResponseError) {
			var req request
			if err := decodeToolArguments(params.Arguments, &req); err != nil {
				return CallToolResult{}, newValidationError("invalid tool arguments", FieldErrorDetail{
					Field:      "arguments",
					Constraint: "invalid",
					Message:    err.Error(),
				})
			}

			result, err := services.health.Health(ctx, defaultStartPath(req.StartPath), repository.HealthRequest{})
			if err != nil {
				return CallToolResult{}, mapOperationalToolError(err)
			}

			return newStructuredToolResult(queryMetaFromEnvelope(result.Repository), result)
		},
	}
}

func mapOperationalToolError(err error) *ResponseError {
	if err == nil {
		return nil
	}

	message := err.Error()
	switch {
	case strings.Contains(message, "max sections"):
		return newMaximumExceededError("sections", maxPackSectionsLimit, message)
	case strings.Contains(message, "max targets"):
		return newMaximumExceededError("targets", maxPackTargetsLimit, message)
	case strings.Contains(message, "max matches"):
		return newMaximumExceededError("limit", maxPackLookupLimit, message)
	case strings.Contains(message, "context lines exceed max"):
		return newMaximumExceededError("contextLines", maxPackContextLines, message)
	case strings.Contains(message, "target range exceeds max lines"):
		return newMaximumExceededError("targetRangeLines", maxPackTargetRangeLimit, message)
	case strings.Contains(message, "scope") && strings.Contains(message, "not found"):
		return newValidationError("requested path is not indexed", FieldErrorDetail{
			Field:      "pathPrefix",
			Constraint: "invalid",
			Message:    message,
		})
	default:
		return mapQueryToolError(err)
	}
}
