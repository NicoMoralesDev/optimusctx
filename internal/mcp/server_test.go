package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
)

func TestMCPServerStdioSession(t *testing.T) {
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
	writeTestFrame(t, &input, Request{
		JSONRPC: jsonRPCVersion,
		Method:  "notifications/initialized",
	})
	writeTestFrame(t, &input, Request{
		JSONRPC: jsonRPCVersion,
		ID:      2,
		Method:  "tools/list",
	})

	var output bytes.Buffer
	server := NewServer(&input, &output, io.Discard)
	server.RegisterTool(ToolHandler{
		Tool: Tool{
			Name:        "optimusctx.z-last",
			Description: "Registered out of order to prove deterministic listing",
		},
		Call: func(ctx context.Context, params CallToolParams) (CallToolResult, *ResponseError) {
			return CallToolResult{
				Content: []ToolContent{{Type: "text", Text: "z"}},
			}, nil
		},
	})
	server.RegisterTool(ToolHandler{
		Tool: Tool{
			Name:        "optimusctx.echo",
			Description: "Echo input during MCP session tests",
			InputSchema: map[string]any{
				"type": "object",
			},
		},
		Call: func(ctx context.Context, params CallToolParams) (CallToolResult, *ResponseError) {
			return CallToolResult{
				Content: []ToolContent{{Type: "text", Text: "ok"}},
			}, nil
		},
	})

	if err := server.Serve(context.Background()); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}

	responses := readTestResponses(t, &output)
	if len(responses) != 2 {
		t.Fatalf("response count = %d, want 2", len(responses))
	}

	var initResult InitializeResult
	mustDecodeResult(t, responses[0].Result, &initResult)
	if initResult.ProtocolVersion != protocolVersion {
		t.Fatalf("initialize protocol version = %q, want %q", initResult.ProtocolVersion, protocolVersion)
	}
	if initResult.ServerInfo.Name != "optimusctx" {
		t.Fatalf("server name = %q, want optimusctx", initResult.ServerInfo.Name)
	}
	if initResult.Capabilities.Tools.ListChanged {
		t.Fatal("tools.listChanged = true, want false")
	}

	var toolsResult ToolsListResult
	mustDecodeResult(t, responses[1].Result, &toolsResult)
	if len(toolsResult.Tools) != 2 {
		t.Fatalf("tool count = %d, want 2", len(toolsResult.Tools))
	}
	if toolsResult.Tools[0].Name != "optimusctx.echo" {
		t.Fatalf("tool name = %q, want optimusctx.echo", toolsResult.Tools[0].Name)
	}
	if toolsResult.Tools[1].Name != "optimusctx.z-last" {
		t.Fatalf("tool name = %q, want optimusctx.z-last", toolsResult.Tools[1].Name)
	}
}

func TestMCPServerRejectsUnknownTool(t *testing.T) {
	var input bytes.Buffer
	writeTestFrame(t, &input, Request{
		JSONRPC: jsonRPCVersion,
		ID:      7,
		Method:  "tools/call",
		Params: CallToolParams{
			Name: "optimusctx.missing",
		},
	})

	var output bytes.Buffer
	server := NewServer(&input, &output, io.Discard)
	if err := server.Serve(context.Background()); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}

	responses := readTestResponses(t, &output)
	if len(responses) != 1 {
		t.Fatalf("response count = %d, want 1", len(responses))
	}
	if responses[0].Error == nil {
		t.Fatal("expected structured error response")
	}
	if responses[0].Error.Code != errCodeUnknownTool {
		t.Fatalf("error code = %d, want %d", responses[0].Error.Code, errCodeUnknownTool)
	}
	if responses[0].Error.Data["tool"] != "optimusctx.missing" {
		t.Fatalf("error data tool = %#v, want optimusctx.missing", responses[0].Error.Data["tool"])
	}
}

func TestMCPServerRejectsUnimplementedTool(t *testing.T) {
	var input bytes.Buffer
	writeTestFrame(t, &input, Request{
		JSONRPC: jsonRPCVersion,
		ID:      8,
		Method:  "tools/call",
		Params: CallToolParams{
			Name: "optimusctx.pending",
		},
	})

	var output bytes.Buffer
	server := NewServer(&input, &output, io.Discard)
	server.RegisterTool(ToolHandler{
		Tool: Tool{
			Name:        "optimusctx.pending",
			Description: "Placeholder slot for future MCP plan slices",
		},
	})

	if err := server.Serve(context.Background()); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}

	responses := readTestResponses(t, &output)
	if len(responses) != 1 {
		t.Fatalf("response count = %d, want 1", len(responses))
	}
	if responses[0].Error == nil {
		t.Fatal("expected structured error response")
	}
	if responses[0].Error.Code != errCodeMethodNotFound {
		t.Fatalf("error code = %d, want %d", responses[0].Error.Code, errCodeMethodNotFound)
	}
	if responses[0].Error.Data["tool"] != "optimusctx.pending" {
		t.Fatalf("error data tool = %#v, want optimusctx.pending", responses[0].Error.Data["tool"])
	}
}

func writeTestFrame(t *testing.T, writer io.Writer, request Request) {
	t.Helper()

	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("json.Marshal(request) error = %v", err)
	}
	if err := writeFrame(writer, payload); err != nil {
		t.Fatalf("writeFrame() error = %v", err)
	}
}

func readTestResponses(t *testing.T, reader io.Reader) []Response {
	t.Helper()

	buffered := bufioReader(reader)
	var responses []Response
	for {
		payload, err := readFrame(buffered)
		if err != nil {
			if err == io.EOF {
				return responses
			}
			t.Fatalf("readFrame() error = %v", err)
		}

		var response Response
		if err := json.Unmarshal(payload, &response); err != nil {
			t.Fatalf("json.Unmarshal(response) error = %v", err)
		}
		responses = append(responses, response)
	}
}

func mustDecodeResult(t *testing.T, raw any, target any) {
	t.Helper()

	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("json.Marshal(result) error = %v", err)
	}
	if err := json.Unmarshal(payload, target); err != nil {
		t.Fatalf("json.Unmarshal(result) error = %v", err)
	}
}

func bufioReader(reader io.Reader) *bufio.Reader {
	if buffered, ok := reader.(*bufio.Reader); ok {
		return buffered
	}
	return bufio.NewReader(reader)
}
