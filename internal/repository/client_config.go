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
	ClientGeminiCLI     ClientID = "gemini-cli"
	ClientGenericMCP    ClientID = "generic"
)

type ClientSupportLevel string

const (
	ClientSupportLevelNative ClientSupportLevel = "native"
	ClientSupportLevelManual ClientSupportLevel = "manual"
)

type ClientConfigKind string

const (
	ClientConfigKindJSON    ClientConfigKind = "json"
	ClientConfigKindTOML    ClientConfigKind = "toml"
	ClientConfigKindCommand ClientConfigKind = "command"
	ClientConfigKindManual  ClientConfigKind = "manual"
)

type ClientConfigScope string

const (
	ClientConfigScopeRepo    ClientConfigScope = "repo"
	ClientConfigScopeShared  ClientConfigScope = "shared"
	ClientConfigScopeProject ClientConfigScope = "project"
	ClientConfigScopeLocal   ClientConfigScope = "local"
	ClientConfigScopeUser    ClientConfigScope = "user"
)

type ClientGuidanceSupport string

const (
	ClientGuidanceSupportManaged     ClientGuidanceSupport = "managed"
	ClientGuidanceSupportUnsupported ClientGuidanceSupport = "unsupported"
)

type ClientCapabilities struct {
	SupportLevel          ClientSupportLevel
	ConfigKind            ClientConfigKind
	ConfigScopes          []ClientConfigScope
	GuidanceSupport       ClientGuidanceSupport
	UsageEvidence         bool
	MixedEnvironmentAware bool
}

type SupportedClient struct {
	ID           ClientID
	DisplayName  string
	Capabilities ClientCapabilities
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
		{
			ID:          ClientClaudeDesktop,
			DisplayName: "Claude Desktop",
			Capabilities: ClientCapabilities{
				SupportLevel:          ClientSupportLevelNative,
				ConfigKind:            ClientConfigKindJSON,
				ConfigScopes:          []ClientConfigScope{ClientConfigScopeShared},
				GuidanceSupport:       ClientGuidanceSupportUnsupported,
				UsageEvidence:         true,
				MixedEnvironmentAware: true,
			},
		},
		{
			ID:          ClientClaudeCLI,
			DisplayName: "Claude CLI",
			Capabilities: ClientCapabilities{
				SupportLevel:    ClientSupportLevelNative,
				ConfigKind:      ClientConfigKindCommand,
				ConfigScopes:    []ClientConfigScope{ClientConfigScopeLocal, ClientConfigScopeProject, ClientConfigScopeUser},
				GuidanceSupport: ClientGuidanceSupportManaged,
				UsageEvidence:   true,
			},
		},
		{
			ID:          ClientCodexApp,
			DisplayName: "Codex App",
			Capabilities: ClientCapabilities{
				SupportLevel:          ClientSupportLevelNative,
				ConfigKind:            ClientConfigKindTOML,
				ConfigScopes:          []ClientConfigScope{ClientConfigScopeShared},
				GuidanceSupport:       ClientGuidanceSupportManaged,
				UsageEvidence:         true,
				MixedEnvironmentAware: true,
			},
		},
		{
			ID:          ClientCodexCLI,
			DisplayName: "Codex CLI",
			Capabilities: ClientCapabilities{
				SupportLevel:    ClientSupportLevelNative,
				ConfigKind:      ClientConfigKindTOML,
				ConfigScopes:    []ClientConfigScope{ClientConfigScopeRepo, ClientConfigScopeShared},
				GuidanceSupport: ClientGuidanceSupportManaged,
				UsageEvidence:   true,
			},
		},
		{
			ID:          ClientGeminiCLI,
			DisplayName: "Gemini CLI",
			Capabilities: ClientCapabilities{
				SupportLevel:    ClientSupportLevelNative,
				ConfigKind:      ClientConfigKindJSON,
				ConfigScopes:    []ClientConfigScope{ClientConfigScopeRepo, ClientConfigScopeShared},
				GuidanceSupport: ClientGuidanceSupportManaged,
				UsageEvidence:   true,
			},
		},
		{
			ID:          ClientGenericMCP,
			DisplayName: "Generic MCP Client",
			Capabilities: ClientCapabilities{
				SupportLevel:    ClientSupportLevelManual,
				ConfigKind:      ClientConfigKindManual,
				GuidanceSupport: ClientGuidanceSupportUnsupported,
			},
		},
	}
}

func (c SupportedClient) IsFirstClass() bool {
	return c.Capabilities.SupportLevel == ClientSupportLevelNative
}

func (c SupportedClient) SupportsScope(scope ClientConfigScope) bool {
	for _, candidate := range c.Capabilities.ConfigScopes {
		if candidate == scope {
			return true
		}
	}
	return false
}

func (c SupportedClient) CapabilitySummary() string {
	parts := []string{
		fmt.Sprintf("support=%s", c.Capabilities.SupportLevel),
		fmt.Sprintf("config=%s", c.Capabilities.ConfigKind),
	}
	if len(c.Capabilities.ConfigScopes) > 0 {
		scopes := make([]string, 0, len(c.Capabilities.ConfigScopes))
		for _, scope := range c.Capabilities.ConfigScopes {
			scopes = append(scopes, string(scope))
		}
		parts = append(parts, fmt.Sprintf("scopes=%s", strings.Join(scopes, "/")))
	}
	parts = append(parts, fmt.Sprintf("guidance=%s", c.Capabilities.GuidanceSupport))
	if c.Capabilities.UsageEvidence {
		parts = append(parts, "usage_evidence=repo-local")
	} else {
		parts = append(parts, "usage_evidence=none")
	}
	if c.Capabilities.MixedEnvironmentAware {
		parts = append(parts, "env=mixed-aware")
	} else {
		parts = append(parts, "env=current")
	}
	return strings.Join(parts, " ")
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
	document, err := ParseClientConfig(existing)
	if err != nil {
		return ClientConfigDocument{}, err
	}
	if document.MCPServers == nil {
		document.MCPServers = map[string]ServeCommand{}
	}
	document.MCPServers[serverName] = command
	return document, nil
}

func ParseClientConfig(existing []byte) (ClientConfigDocument, error) {
	document := ClientConfigDocument{
		MCPServers: map[string]ServeCommand{},
	}
	if len(strings.TrimSpace(string(existing))) == 0 {
		return document, nil
	}
	if err := json.Unmarshal(existing, &document); err != nil {
		return ClientConfigDocument{}, fmt.Errorf("decode client config: %w", err)
	}
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
