package mcp

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

const (
	toolRepositoryMap    = "optimusctx.repository_map"
	toolLayeredContextL0 = "optimusctx.layered_context_l0"
	toolLayeredContextL1 = "optimusctx.layered_context_l1"
	toolSymbolLookup     = "optimusctx.symbol_lookup"
	toolStructureLookup  = "optimusctx.structure_lookup"
	toolTargetedContext  = "optimusctx.targeted_context"

	defaultRepositoryDirectoryLimit = 24
	maxRepositoryDirectoryLimit     = 64
	defaultRepositoryFileLimit      = 24
	maxRepositoryFileLimit          = 64
	defaultRepositorySymbolLimit    = 12
	maxRepositorySymbolLimit        = 32
	maxLookupLimit                  = 25
	maxContextWindowLines           = 80
)

type queryToolServices struct {
	repositoryMap app.RepositoryMapService
	context       app.RepositoryContextService
	lookup        app.LookupService
	contextBlock  app.ContextBlockService
}

func defaultQueryToolServices() queryToolServices {
	return queryToolServices{
		repositoryMap: app.NewRepositoryMapService(),
		context:       app.NewRepositoryContextService(),
		lookup:        app.NewLookupService(),
		contextBlock:  app.NewContextBlockService(),
	}
}

func registerDefaultQueryTools(server *Server) {
	registerQueryTools(server, defaultQueryToolServices())
}

func registerQueryTools(server *Server, services queryToolServices) {
	tools := []ToolHandler{
		newRepositoryMapTool(services),
		newLayeredContextL0Tool(services),
		newLayeredContextL1Tool(services),
		newSymbolLookupTool(services),
		newStructureLookupTool(services),
		newTargetedContextTool(services),
	}
	for _, tool := range tools {
		server.RegisterTool(tool)
	}
}

func newRepositoryMapTool(services queryToolServices) ToolHandler {
	type request struct {
		StartPath         string `json:"startPath"`
		DirectoryLimit    int    `json:"directoryLimit"`
		FilesPerDirectory int    `json:"filesPerDirectory"`
		SymbolsPerFile    int    `json:"symbolsPerFile"`
	}

	return ToolHandler{
		Tool: Tool{
			Name:        toolRepositoryMap,
			Description: "Return the persisted repository map with explicit directory, file, and symbol bounds.",
			InputSchema: objectSchema(
				propertySchema("startPath", "string"),
				propertySchema("directoryLimit", "integer"),
				propertySchema("filesPerDirectory", "integer"),
				propertySchema("symbolsPerFile", "integer"),
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

			dirLimit, err := normalizeLimit("directoryLimit", req.DirectoryLimit, defaultRepositoryDirectoryLimit, maxRepositoryDirectoryLimit)
			if err != nil {
				return CallToolResult{}, err
			}
			fileLimit, err := normalizeLimit("filesPerDirectory", req.FilesPerDirectory, defaultRepositoryFileLimit, maxRepositoryFileLimit)
			if err != nil {
				return CallToolResult{}, err
			}
			symbolLimit, err := normalizeLimit("symbolsPerFile", req.SymbolsPerFile, defaultRepositorySymbolLimit, maxRepositorySymbolLimit)
			if err != nil {
				return CallToolResult{}, err
			}

			result, callErr := services.repositoryMap.RepositoryMap(ctx, defaultStartPath(req.StartPath))
			if callErr != nil {
				return CallToolResult{}, mapQueryToolError(callErr)
			}

			bounded, truncated := applyRepositoryMapBounds(result, dirLimit, fileLimit, symbolLimit)
			meta := QueryResultMetadata{
				RepositoryRoot: result.RepositoryRoot,
				Generation:     result.Generation,
				Freshness:      string(result.Freshness),
				CacheStatus:    cacheStatusPersistedOnly,
				Bounds: QueryBoundsMetadata{
					DefaultLimit:   defaultRepositoryDirectoryLimit,
					MaxLimit:       maxRepositoryDirectoryLimit,
					RequestedLimit: req.DirectoryLimit,
					AppliedLimit:   dirLimit,
					ReturnedCount:  len(bounded.Directories),
					TotalCount:     int64(len(result.Directories)),
					Truncated:      truncated,
				},
			}

			return newStructuredToolResult(meta, bounded)
		},
	}
}

func newLayeredContextL0Tool(services queryToolServices) ToolHandler {
	type request struct {
		StartPath string `json:"startPath"`
	}

	return ToolHandler{
		Tool: Tool{
			Name:        toolLayeredContextL0,
			Description: "Return the persisted L0 repository context envelope.",
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

			result, err := services.context.LayeredContextL0(ctx, defaultStartPath(req.StartPath))
			if err != nil {
				return CallToolResult{}, mapQueryToolError(err)
			}

			return newStructuredToolResult(queryMetaFromEnvelope(result.Repository), result)
		},
	}
}

func newLayeredContextL1Tool(services queryToolServices) ToolHandler {
	type request struct {
		StartPath string `json:"startPath"`
	}

	return ToolHandler{
		Tool: Tool{
			Name:        toolLayeredContextL1,
			Description: "Return the persisted L1 repository context with shared freshness metadata.",
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

			result, err := services.context.LayeredContextL1(ctx, defaultStartPath(req.StartPath))
			if err != nil {
				return CallToolResult{}, mapQueryToolError(err)
			}

			meta := queryMetaFromEnvelope(result.Repository)
			meta.Bounds = QueryBoundsMetadata{
				DefaultLimit:  result.Limits.FileLimit,
				AppliedLimit:  result.Limits.FileLimit,
				ReturnedCount: len(result.Candidates),
				TotalCount:    result.Limits.TotalCandidateCount,
				Truncated:     result.Limits.FileTruncated,
				MaxLimit:      result.Limits.FileLimit,
			}
			return newStructuredToolResult(meta, result)
		},
	}
}

func newSymbolLookupTool(services queryToolServices) ToolHandler {
	type request struct {
		StartPath  string `json:"startPath"`
		Name       string `json:"name"`
		PathPrefix string `json:"pathPrefix"`
		Language   string `json:"language"`
		Kind       string `json:"kind"`
		Limit      int    `json:"limit"`
	}

	return ToolHandler{
		Tool: Tool{
			Name:        toolSymbolLookup,
			Description: "Run an exact symbol-name lookup over persisted symbols with explicit result limits.",
			InputSchema: objectSchema(
				propertySchema("startPath", "string"),
				propertySchema("name", "string"),
				propertySchema("pathPrefix", "string"),
				propertySchema("language", "string"),
				propertySchema("kind", "string"),
				propertySchema("limit", "integer"),
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
			if strings.TrimSpace(req.Name) == "" {
				return CallToolResult{}, newRequiredFieldError("name")
			}
			limit, err := normalizeLimit("limit", req.Limit, maxLookupLimit, maxLookupLimit)
			if err != nil {
				return CallToolResult{}, err
			}

			result, callErr := services.lookup.SymbolLookup(ctx, defaultStartPath(req.StartPath), repository.SymbolLookupRequest{
				Name:       req.Name,
				PathPrefix: req.PathPrefix,
				Language:   req.Language,
				Kind:       req.Kind,
				Limit:      limit,
			})
			if callErr != nil {
				return CallToolResult{}, mapQueryToolError(callErr)
			}

			meta := queryMetaFromEnvelope(result.Repository)
			meta.Bounds = QueryBoundsMetadata{
				DefaultLimit:   maxLookupLimit,
				MaxLimit:       maxLookupLimit,
				RequestedLimit: req.Limit,
				AppliedLimit:   result.Limit,
				ReturnedCount:  len(result.Matches),
				LimitReached:   len(result.Matches) == result.Limit,
			}
			return newStructuredToolResult(meta, result)
		},
	}
}

func newStructureLookupTool(services queryToolServices) ToolHandler {
	type request struct {
		StartPath  string `json:"startPath"`
		Kind       string `json:"kind"`
		ParentName string `json:"parentName"`
		Name       string `json:"name"`
		PathPrefix string `json:"pathPrefix"`
		Language   string `json:"language"`
		Limit      int    `json:"limit"`
	}

	return ToolHandler{
		Tool: Tool{
			Name:        toolStructureLookup,
			Description: "Run an exact structural lookup over persisted symbols with explicit result limits.",
			InputSchema: objectSchema(
				propertySchema("startPath", "string"),
				propertySchema("kind", "string"),
				propertySchema("parentName", "string"),
				propertySchema("name", "string"),
				propertySchema("pathPrefix", "string"),
				propertySchema("language", "string"),
				propertySchema("limit", "integer"),
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
			if strings.TrimSpace(req.Kind) == "" {
				return CallToolResult{}, newRequiredFieldError("kind")
			}
			if strings.TrimSpace(req.Name) == "" && strings.TrimSpace(req.ParentName) == "" && strings.TrimSpace(req.PathPrefix) == "" {
				return CallToolResult{}, newValidationError("structure lookup requires at least one narrowing selector", FieldErrorDetail{
					Field:      "selectors",
					Constraint: "required",
					Message:    "provide at least one of name, parentName, or pathPrefix",
				})
			}
			limit, err := normalizeLimit("limit", req.Limit, maxLookupLimit, maxLookupLimit)
			if err != nil {
				return CallToolResult{}, err
			}

			result, callErr := services.lookup.StructureLookup(ctx, defaultStartPath(req.StartPath), repository.StructureLookupRequest{
				Kind:       req.Kind,
				ParentName: req.ParentName,
				Name:       req.Name,
				PathPrefix: req.PathPrefix,
				Language:   req.Language,
				Limit:      limit,
			})
			if callErr != nil {
				return CallToolResult{}, mapQueryToolError(callErr)
			}

			meta := queryMetaFromEnvelope(result.Repository)
			meta.Bounds = QueryBoundsMetadata{
				DefaultLimit:   maxLookupLimit,
				MaxLimit:       maxLookupLimit,
				RequestedLimit: req.Limit,
				AppliedLimit:   result.Limit,
				ReturnedCount:  len(result.Matches),
				LimitReached:   len(result.Matches) == result.Limit,
			}
			return newStructuredToolResult(meta, result)
		},
	}
}

func newTargetedContextTool(services queryToolServices) ToolHandler {
	type request struct {
		StartPath   string `json:"startPath"`
		StableKey   string `json:"stableKey"`
		Path        string `json:"path"`
		StartLine   int    `json:"startLine"`
		EndLine     int    `json:"endLine"`
		BeforeLines int    `json:"beforeLines"`
		AfterLines  int    `json:"afterLines"`
	}

	return ToolHandler{
		Tool: Tool{
			Name:        toolTargetedContext,
			Description: "Return a targeted code window by stable symbol key or explicit line range.",
			InputSchema: objectSchema(
				propertySchema("startPath", "string"),
				propertySchema("stableKey", "string"),
				propertySchema("path", "string"),
				propertySchema("startLine", "integer"),
				propertySchema("endLine", "integer"),
				propertySchema("beforeLines", "integer"),
				propertySchema("afterLines", "integer"),
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
			if req.BeforeLines < 0 {
				return CallToolResult{}, newMinimumViolationError("beforeLines", 0, req.BeforeLines)
			}
			if req.AfterLines < 0 {
				return CallToolResult{}, newMinimumViolationError("afterLines", 0, req.AfterLines)
			}
			if req.BeforeLines > maxContextWindowLines {
				return CallToolResult{}, newMaximumExceededError("beforeLines", maxContextWindowLines, req.BeforeLines)
			}
			if req.AfterLines > maxContextWindowLines {
				return CallToolResult{}, newMaximumExceededError("afterLines", maxContextWindowLines, req.AfterLines)
			}
			if strings.TrimSpace(req.StableKey) != "" && (strings.TrimSpace(req.Path) != "" || req.StartLine != 0 || req.EndLine != 0) {
				return CallToolResult{}, newConflictFieldError("stableKey", req.StableKey, "stableKey requests cannot also specify path or line bounds")
			}
			if strings.TrimSpace(req.StableKey) == "" {
				if strings.TrimSpace(req.Path) == "" {
					return CallToolResult{}, newRequiredFieldError("path")
				}
				if req.StartLine <= 0 {
					return CallToolResult{}, newMinimumViolationError("startLine", 1, req.StartLine)
				}
				if req.EndLine < req.StartLine {
					return CallToolResult{}, newValidationError("endLine must be greater than or equal to startLine", FieldErrorDetail{
						Field:      "endLine",
						Constraint: "invalid",
						Received:   req.EndLine,
					})
				}
			}

			result, err := services.contextBlock.TargetedContext(ctx, defaultStartPath(req.StartPath), repository.TargetedContextRequest{
				StableKey:   req.StableKey,
				Path:        req.Path,
				StartLine:   req.StartLine,
				EndLine:     req.EndLine,
				BeforeLines: req.BeforeLines,
				AfterLines:  req.AfterLines,
			})
			if err != nil {
				return CallToolResult{}, mapQueryToolError(err)
			}

			meta := queryMetaFromEnvelope(result.Repository)
			meta.Bounds = QueryBoundsMetadata{
				MaxLimit:    maxContextWindowLines,
				BeforeLines: result.BeforeLines,
				AfterLines:  result.AfterLines,
				Truncated:   result.TruncatedStart || result.TruncatedEnd,
			}
			return newStructuredToolResult(meta, result)
		},
	}
}

func decodeToolArguments(arguments map[string]any, target any) error {
	if len(arguments) == 0 {
		return nil
	}

	payload, err := json.Marshal(arguments)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, target)
}

func normalizeLimit(field string, requested int, defaultLimit int, maxLimit int) (int, *ResponseError) {
	if requested < 0 {
		return 0, newMinimumViolationError(field, 0, requested)
	}
	if requested == 0 {
		return defaultLimit, nil
	}
	if requested > maxLimit {
		return 0, newMaximumExceededError(field, maxLimit, requested)
	}
	return requested, nil
}

func defaultStartPath(startPath string) string {
	if strings.TrimSpace(startPath) == "" {
		return "."
	}
	return startPath
}

func newStructuredToolResult(meta QueryResultMetadata, data any) (CallToolResult, *ResponseError) {
	envelope := QueryEnvelope{
		Meta: meta,
		Data: data,
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		return CallToolResult{}, &ResponseError{
			Code:    errCodeInternal,
			Message: "failed to encode tool result",
			Data:    map[string]any{"details": err.Error()},
		}
	}

	return CallToolResult{
		Content: []ToolContent{{
			Type:     "text",
			Text:     string(payload),
			MIMEType: "application/json",
		}},
		StructuredContent: envelope,
	}, nil
}

func queryMetaFromEnvelope(envelope repository.LayeredContextEnvelope) QueryResultMetadata {
	return QueryResultMetadata{
		RepositoryRoot: envelope.RepositoryRoot,
		Generation:     envelope.Generation,
		Freshness:      string(envelope.Freshness),
		CacheStatus:    cacheStatusPersistedOnly,
	}
}

func mapQueryToolError(err error) *ResponseError {
	if err == nil {
		return nil
	}

	message := err.Error()
	switch {
	case strings.Contains(message, "name is required"):
		return newRequiredFieldError("name")
	case strings.Contains(message, "kind is required"):
		return newRequiredFieldError("kind")
	case strings.Contains(message, "at least one of name, parent name, or path prefix is required"):
		return newValidationError("structure lookup requires at least one narrowing selector", FieldErrorDetail{
			Field:      "selectors",
			Constraint: "required",
			Message:    "provide at least one of name, parentName, or pathPrefix",
		})
	default:
		return &ResponseError{
			Code:    errCodeInvalidRequest,
			Message: "tool request failed",
			Data:    map[string]any{"details": message},
		}
	}
}

func applyRepositoryMapBounds(result repository.RepositoryMap, directoryLimit int, fileLimit int, symbolLimit int) (repository.RepositoryMap, bool) {
	bounded := repository.RepositoryMap{
		RepositoryRoot: result.RepositoryRoot,
		Generation:     result.Generation,
		Freshness:      result.Freshness,
	}

	limit := len(result.Directories)
	truncated := false
	if limit > directoryLimit {
		limit = directoryLimit
		truncated = true
	}

	bounded.Directories = make([]repository.RepositoryMapDirectory, 0, limit)
	for _, directory := range result.Directories[:limit] {
		copyDirectory := directory
		if len(copyDirectory.Files) > fileLimit {
			copyDirectory.Files = append([]repository.RepositoryMapFile(nil), copyDirectory.Files[:fileLimit]...)
			truncated = true
		} else {
			copyDirectory.Files = append([]repository.RepositoryMapFile(nil), copyDirectory.Files...)
		}

		for fileIndex := range copyDirectory.Files {
			if len(copyDirectory.Files[fileIndex].Symbols) > symbolLimit {
				copyDirectory.Files[fileIndex].Symbols = append([]repository.RepositoryMapSymbol(nil), copyDirectory.Files[fileIndex].Symbols[:symbolLimit]...)
				truncated = true
				continue
			}
			copyDirectory.Files[fileIndex].Symbols = append([]repository.RepositoryMapSymbol(nil), copyDirectory.Files[fileIndex].Symbols...)
		}

		bounded.Directories = append(bounded.Directories, copyDirectory)
	}

	return bounded, truncated
}

func propertySchema(name string, fieldType string) map[string]any {
	return map[string]any{
		name: map[string]any{
			"type": fieldType,
		},
	}
}

func objectSchema(properties ...map[string]any) map[string]any {
	merged := map[string]any{}
	for _, property := range properties {
		for key, value := range property {
			merged[key] = value
		}
	}
	return map[string]any{
		"type":       "object",
		"properties": merged,
	}
}
