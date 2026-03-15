package mcp

const (
	jsonRPCVersion  = "2.0"
	protocolVersion = "2026-03-15"
)

type Request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id,omitempty"`
	Result  any            `json:"result,omitempty"`
	Error   *ResponseError `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

type InitializeParams struct {
	ClientInfo      ClientInfo `json:"clientInfo,omitempty"`
	ProtocolVersion string     `json:"protocolVersion,omitempty"`
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

type ClientInfo struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerCapabilities struct {
	Tools ToolsCapabilities `json:"tools"`
}

type ToolsCapabilities struct {
	ListChanged bool `json:"listChanged"`
}

type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema,omitempty"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type CallToolParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type CallToolResult struct {
	Content           []ToolContent `json:"content"`
	StructuredContent any           `json:"structuredContent,omitempty"`
	IsError           bool          `json:"isError,omitempty"`
}

type ToolContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	MIMEType string `json:"mimeType,omitempty"`
}

const cacheStatusPersistedOnly = "persisted_only"

type QueryEnvelope struct {
	Meta QueryResultMetadata `json:"meta"`
	Data any                 `json:"data"`
}

type QueryResultMetadata struct {
	RepositoryRoot string              `json:"repositoryRoot"`
	Generation     int64               `json:"generation"`
	Freshness      string              `json:"freshness"`
	CacheStatus    string              `json:"cacheStatus"`
	Bounds         QueryBoundsMetadata `json:"bounds,omitempty"`
}

type QueryBoundsMetadata struct {
	DefaultLimit   int   `json:"defaultLimit,omitempty"`
	MaxLimit       int   `json:"maxLimit,omitempty"`
	RequestedLimit int   `json:"requestedLimit,omitempty"`
	AppliedLimit   int   `json:"appliedLimit,omitempty"`
	ReturnedCount  int   `json:"returnedCount,omitempty"`
	TotalCount     int64 `json:"totalCount,omitempty"`
	Truncated      bool  `json:"truncated,omitempty"`
	LimitReached   bool  `json:"limitReached,omitempty"`
	BeforeLines    int   `json:"beforeLines,omitempty"`
	AfterLines     int   `json:"afterLines,omitempty"`
}
