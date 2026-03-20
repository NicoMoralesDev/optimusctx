package repository

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	DefaultMCPServerName    = "optimusctx"
	DefaultServeCommandName = "optimusctx"
	ClaudeCLIScopeLocal     = "local"
	ClaudeCLIScopeProject   = "project"
	ClaudeCLIScopeUser      = "user"
)

type ClientID string

const (
	ClientClaudeDesktop ClientID = "claude-desktop"
	ClientClaudeCLI     ClientID = "claude-cli"
	ClientCodexApp      ClientID = "codex-app"
	ClientCodexCLI      ClientID = "codex-cli"
	ClientGenericMCP    ClientID = "generic"
)

type SupportedClient struct {
	ID          ClientID
	DisplayName string
}

type ServeCommand struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

type ClientConfigDocument struct {
	MCPServers map[string]ServeCommand `json:"mcpServers"`
}

type RenderMode string

const (
	RenderModePreview RenderMode = "preview"
	RenderModeWrite   RenderMode = "write"
)

type RenderedClientConfig struct {
	Client         SupportedClient
	ConfigPath     string
	Mode           RenderMode
	Content        string
	AppliedContent string
	Notes          []string
}

func (r RenderedClientConfig) ContentForWrite() string {
	if strings.TrimSpace(r.AppliedContent) != "" {
		return r.AppliedContent
	}
	return r.Content
}

func SupportedClients() []SupportedClient {
	return []SupportedClient{
		{ID: ClientClaudeDesktop, DisplayName: "Claude Desktop"},
		{ID: ClientClaudeCLI, DisplayName: "Claude CLI"},
		{ID: ClientCodexApp, DisplayName: "Codex App"},
		{ID: ClientCodexCLI, DisplayName: "Codex CLI"},
		{ID: ClientGenericMCP, DisplayName: "Generic MCP Client"},
	}
}

func LookupSupportedClient(name string) (SupportedClient, bool) {
	normalized := strings.TrimSpace(strings.ToLower(name))
	for _, client := range SupportedClients() {
		if normalized == string(client.ID) {
			return client, true
		}
	}
	return SupportedClient{}, false
}

func NewServeCommand(binaryPath string) ServeCommand {
	binaryPath = CanonicalServeCommandPath(binaryPath)
	return ServeCommand{
		Command: binaryPath,
		Args:    []string{"run"},
	}
}

func NormalizeServeCommand(command ServeCommand) ServeCommand {
	if strings.TrimSpace(command.Command) == "" {
		return NewServeCommand("")
	}

	command.Command = CanonicalServeCommandPath(command.Command)
	if len(command.Args) == 0 {
		command.Args = []string{"run"}
	}

	return command
}

func CanonicalServeCommandPath(binaryPath string) string {
	binaryPath = strings.TrimSpace(binaryPath)
	if binaryPath == "" {
		return DefaultServeCommandName
	}
	return binaryPath
}

func MergeClientConfig(existing []byte, serverName string, command ServeCommand) (ClientConfigDocument, error) {
	document := ClientConfigDocument{
		MCPServers: map[string]ServeCommand{},
	}
	if len(strings.TrimSpace(string(existing))) > 0 {
		if err := json.Unmarshal(existing, &document); err != nil {
			return ClientConfigDocument{}, fmt.Errorf("decode client config: %w", err)
		}
	}
	if document.MCPServers == nil {
		document.MCPServers = map[string]ServeCommand{}
	}
	document.MCPServers[serverName] = command
	return document, nil
}

func RenderClientConfig(document ClientConfigDocument) (string, error) {
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode client config: %w", err)
	}
	return string(encoded) + "\n", nil
}

func RenderClientConfigSnippet(serverName string, command ServeCommand) (string, error) {
	document, err := MergeClientConfig(nil, serverName, command)
	if err != nil {
		return "", err
	}
	return RenderClientConfig(document)
}

func NormalizeClaudeCLIScope(scope string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(scope))
	if normalized == "" {
		return ClaudeCLIScopeLocal, nil
	}

	switch normalized {
	case ClaudeCLIScopeLocal, ClaudeCLIScopeProject, ClaudeCLIScopeUser:
		return normalized, nil
	default:
		return "", fmt.Errorf("unsupported Claude CLI scope %q; expected local, project, or user", scope)
	}
}

func RenderClaudeCLIAddCommand(serverName string, scope string, command ServeCommand) string {
	serverName = strings.TrimSpace(serverName)
	if serverName == "" {
		serverName = DefaultMCPServerName
	}

	scope = strings.TrimSpace(strings.ToLower(scope))
	if scope == "" {
		scope = ClaudeCLIScopeLocal
	}

	command = NormalizeServeCommand(command)
	parts := []string{"claude", "mcp", "add", "--transport", "stdio", "--scope", scope, serverName, "--", command.Command}
	parts = append(parts, command.Args...)

	return strings.Join(parts, " ")
}
