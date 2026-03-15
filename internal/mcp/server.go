package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
)

const (
	errCodeParseInvalidJSON = -32700
	errCodeInvalidRequest   = -32600
	errCodeMethodNotFound   = -32601
	errCodeInternal         = -32603
	errCodeUnknownTool      = -32004
)

type ToolHandler struct {
	Tool Tool
	Call func(context.Context, CallToolParams) (CallToolResult, *ResponseError)
}

type Server struct {
	input   *bufio.Reader
	output  io.Writer
	errout  io.Writer
	tools   map[string]ToolHandler
	order   []string
	version string
}

func NewServer(stdin io.Reader, stdout io.Writer, stderr io.Writer) *Server {
	server := &Server{
		input:   bufio.NewReader(stdin),
		output:  stdout,
		errout:  stderr,
		tools:   map[string]ToolHandler{},
		version: "0.1.0",
	}
	registerDefaultTools(server)
	return server
}

func ServeStdio(ctx context.Context, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	return NewServer(stdin, stdout, stderr).Serve(ctx)
}

func (s *Server) RegisterTool(handler ToolHandler) {
	if _, exists := s.tools[handler.Tool.Name]; !exists {
		s.order = append(s.order, handler.Tool.Name)
	}
	s.tools[handler.Tool.Name] = handler
}

func (s *Server) Serve(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		payload, err := readFrame(s.input)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			if s.errout != nil {
				log.New(s.errout, "optimusctx mcp: ", 0).Printf("session ended with transport error: %v", err)
			}
			return err
		}

		response, ok := s.handlePayload(ctx, payload)
		if !ok {
			continue
		}
		if err := writeFrame(s.output, response); err != nil {
			return err
		}
	}
}

func (s *Server) handlePayload(ctx context.Context, payload []byte) ([]byte, bool) {
	var request Request
	if err := json.Unmarshal(payload, &request); err != nil {
		return mustMarshal(Response{
			JSONRPC: jsonRPCVersion,
			Error: &ResponseError{
				Code:    errCodeParseInvalidJSON,
				Message: "invalid JSON payload",
			},
		}), true
	}

	if request.JSONRPC != "" && request.JSONRPC != jsonRPCVersion {
		return mustMarshal(Response{
			JSONRPC: jsonRPCVersion,
			ID:      request.ID,
			Error: &ResponseError{
				Code:    errCodeInvalidRequest,
				Message: "unsupported jsonrpc version",
				Data:    map[string]any{"expected": jsonRPCVersion},
			},
		}), true
	}

	response := s.handleRequest(ctx, request)
	if response == nil {
		return nil, false
	}

	return mustMarshal(*response), true
}

func (s *Server) handleRequest(ctx context.Context, request Request) *Response {
	switch request.Method {
	case "notifications/initialized":
		return nil
	case "initialize":
		return &Response{
			JSONRPC: jsonRPCVersion,
			ID:      request.ID,
			Result: InitializeResult{
				ProtocolVersion: protocolVersion,
				Capabilities: ServerCapabilities{
					Tools: ToolsCapabilities{ListChanged: false},
				},
				ServerInfo: ServerInfo{
					Name:    "optimusctx",
					Version: s.version,
				},
			},
		}
	case "tools/list":
		tools := make([]Tool, 0, len(s.tools))
		names := append([]string(nil), s.order...)
		sort.Strings(names)
		for _, name := range names {
			tools = append(tools, s.tools[name].Tool)
		}
		return &Response{
			JSONRPC: jsonRPCVersion,
			ID:      request.ID,
			Result:  ToolsListResult{Tools: tools},
		}
	case "tools/call":
		params, err := decodeCallToolParams(request.Params)
		if err != nil {
			return &Response{
				JSONRPC: jsonRPCVersion,
				ID:      request.ID,
				Error: &ResponseError{
					Code:    errCodeInvalidRequest,
					Message: "invalid tool call parameters",
					Data:    map[string]any{"details": err.Error()},
				},
			}
		}

		handler, ok := s.tools[params.Name]
		if !ok {
			return &Response{
				JSONRPC: jsonRPCVersion,
				ID:      request.ID,
				Error: &ResponseError{
					Code:    errCodeUnknownTool,
					Message: "tool is not registered",
					Data:    map[string]any{"tool": params.Name},
				},
			}
		}
		if handler.Call == nil {
			return &Response{
				JSONRPC: jsonRPCVersion,
				ID:      request.ID,
				Error: &ResponseError{
					Code:    errCodeMethodNotFound,
					Message: "tool is registered but not implemented",
					Data:    map[string]any{"tool": params.Name},
				},
			}
		}

		result, callErr := handler.Call(ctx, params)
		if callErr != nil {
			return &Response{
				JSONRPC: jsonRPCVersion,
				ID:      request.ID,
				Error:   callErr,
			}
		}
		return &Response{
			JSONRPC: jsonRPCVersion,
			ID:      request.ID,
			Result:  result,
		}
	default:
		return &Response{
			JSONRPC: jsonRPCVersion,
			ID:      request.ID,
			Error: &ResponseError{
				Code:    errCodeMethodNotFound,
				Message: "method not found",
				Data:    map[string]any{"method": request.Method},
			},
		}
	}
}

func decodeCallToolParams(raw any) (CallToolParams, error) {
	if raw == nil {
		return CallToolParams{}, errors.New("missing params")
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return CallToolParams{}, err
	}

	var params CallToolParams
	if err := json.Unmarshal(data, &params); err != nil {
		return CallToolParams{}, err
	}
	if params.Name == "" {
		return CallToolParams{}, errors.New("name is required")
	}
	return params, nil
}

func readFrame(reader *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("invalid header line %q", line)
		}
		if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			contentLength, err = strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, fmt.Errorf("parse content length: %w", err)
			}
		}
	}

	if contentLength < 0 {
		return nil, errors.New("missing Content-Length header")
	}

	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func writeFrame(writer io.Writer, payload []byte) error {
	if _, err := fmt.Fprintf(writer, "Content-Length: %d\r\n\r\n", len(payload)); err != nil {
		return err
	}
	_, err := io.Copy(writer, bytes.NewReader(payload))
	return err
}

func mustMarshal(response Response) []byte {
	payload, err := json.Marshal(response)
	if err != nil {
		panic(fmt.Sprintf("marshal MCP response: %v", err))
	}
	return payload
}
