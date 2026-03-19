package repository

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	DefaultMCPServerName    = "optimusctx"
	DefaultServeCommandName = "optimusctx"
)

type ClientID string

const ClientClaudeDesktop ClientID = "claude-desktop"

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
	Client     SupportedClient
	ConfigPath string
	Mode       RenderMode
	Content    string
}

func SupportedClients() []SupportedClient {
	return []SupportedClient{
		{ID: ClientClaudeDesktop, DisplayName: "Claude Desktop"},
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
