package repository

import "testing"

func TestSupportedClientsRemainExplicit(t *testing.T) {
	clients := SupportedClients()
	wantIDs := []ClientID{
		ClientClaudeDesktop,
		ClientClaudeCLI,
		ClientCodexApp,
		ClientCodexCLI,
		ClientGeminiCLI,
		ClientGenericMCP,
	}

	if len(clients) != len(wantIDs) {
		t.Fatalf("supported client count = %d, want %d", len(clients), len(wantIDs))
	}

	for i, wantID := range wantIDs {
		if clients[i].ID != wantID {
			t.Fatalf("supported client %d = %q, want %q", i, clients[i].ID, wantID)
		}

		resolved, ok := LookupSupportedClient(string(wantID))
		if !ok {
			t.Fatalf("LookupSupportedClient(%q) did not resolve", wantID)
		}
		if resolved.ID != wantID {
			t.Fatalf("LookupSupportedClient(%q) = %q", wantID, resolved.ID)
		}
	}
}

func TestSupportedClientsExposeCapabilities(t *testing.T) {
	tests := []struct {
		id                   ClientID
		wantLevel            ClientSupportLevel
		wantKind             ClientConfigKind
		wantGuidance         ClientGuidanceSupport
		wantUsageEvidence    bool
		wantMixedEnvironment bool
		wantScopes           []ClientConfigScope
	}{
		{
			id:                   ClientClaudeDesktop,
			wantLevel:            ClientSupportLevelNative,
			wantKind:             ClientConfigKindJSON,
			wantGuidance:         ClientGuidanceSupportUnsupported,
			wantUsageEvidence:    true,
			wantMixedEnvironment: true,
			wantScopes:           []ClientConfigScope{ClientConfigScopeShared},
		},
		{
			id:                ClientClaudeCLI,
			wantLevel:         ClientSupportLevelNative,
			wantKind:          ClientConfigKindCommand,
			wantGuidance:      ClientGuidanceSupportManaged,
			wantUsageEvidence: true,
			wantScopes:        []ClientConfigScope{ClientConfigScopeLocal, ClientConfigScopeProject, ClientConfigScopeUser},
		},
		{
			id:                   ClientCodexApp,
			wantLevel:            ClientSupportLevelNative,
			wantKind:             ClientConfigKindTOML,
			wantGuidance:         ClientGuidanceSupportManaged,
			wantUsageEvidence:    true,
			wantMixedEnvironment: true,
			wantScopes:           []ClientConfigScope{ClientConfigScopeShared},
		},
		{
			id:                ClientCodexCLI,
			wantLevel:         ClientSupportLevelNative,
			wantKind:          ClientConfigKindTOML,
			wantGuidance:      ClientGuidanceSupportManaged,
			wantUsageEvidence: true,
			wantScopes:        []ClientConfigScope{ClientConfigScopeRepo, ClientConfigScopeShared},
		},
		{
			id:                ClientGeminiCLI,
			wantLevel:         ClientSupportLevelNative,
			wantKind:          ClientConfigKindJSON,
			wantGuidance:      ClientGuidanceSupportManaged,
			wantUsageEvidence: true,
			wantScopes:        []ClientConfigScope{ClientConfigScopeRepo, ClientConfigScopeShared},
		},
		{
			id:           ClientGenericMCP,
			wantLevel:    ClientSupportLevelManual,
			wantKind:     ClientConfigKindManual,
			wantGuidance: ClientGuidanceSupportUnsupported,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			client, ok := LookupSupportedClient(string(tt.id))
			if !ok {
				t.Fatalf("LookupSupportedClient(%q) did not resolve", tt.id)
			}
			if client.Capabilities.SupportLevel != tt.wantLevel {
				t.Fatalf("support level = %q, want %q", client.Capabilities.SupportLevel, tt.wantLevel)
			}
			if client.Capabilities.ConfigKind != tt.wantKind {
				t.Fatalf("config kind = %q, want %q", client.Capabilities.ConfigKind, tt.wantKind)
			}
			if client.Capabilities.GuidanceSupport != tt.wantGuidance {
				t.Fatalf("guidance support = %q, want %q", client.Capabilities.GuidanceSupport, tt.wantGuidance)
			}
			if client.Capabilities.UsageEvidence != tt.wantUsageEvidence {
				t.Fatalf("usage evidence = %v, want %v", client.Capabilities.UsageEvidence, tt.wantUsageEvidence)
			}
			if client.Capabilities.MixedEnvironmentAware != tt.wantMixedEnvironment {
				t.Fatalf("mixed environment = %v, want %v", client.Capabilities.MixedEnvironmentAware, tt.wantMixedEnvironment)
			}
			if len(client.Capabilities.ConfigScopes) != len(tt.wantScopes) {
				t.Fatalf("scope count = %d, want %d", len(client.Capabilities.ConfigScopes), len(tt.wantScopes))
			}
			for i, wantScope := range tt.wantScopes {
				if client.Capabilities.ConfigScopes[i] != wantScope {
					t.Fatalf("scope %d = %q, want %q", i, client.Capabilities.ConfigScopes[i], wantScope)
				}
				if !client.SupportsScope(wantScope) {
					t.Fatalf("SupportsScope(%q) = false", wantScope)
				}
			}
			if tt.wantLevel == ClientSupportLevelNative && !client.IsFirstClass() {
				t.Fatal("IsFirstClass() = false, want true")
			}
			if tt.wantLevel == ClientSupportLevelManual && client.IsFirstClass() {
				t.Fatal("IsFirstClass() = true, want false")
			}
			if got := client.CapabilitySummary(); got == "" {
				t.Fatal("CapabilitySummary() should not be empty")
			}
		})
	}
}

func TestRenderGenericClientConfig(t *testing.T) {
	document, err := MergeClientConfig(nil, DefaultMCPServerName, NewServeCommand(""))
	if err != nil {
		t.Fatalf("MergeClientConfig() error = %v", err)
	}

	got, err := RenderClientConfig(document)
	if err != nil {
		t.Fatalf("RenderClientConfig() error = %v", err)
	}

	const want = "{\n  \"mcpServers\": {\n    \"optimusctx\": {\n      \"command\": \"optimusctx\",\n      \"args\": [\n        \"run\"\n      ]\n    }\n  }\n}\n"
	if got != want {
		t.Fatalf("RenderClientConfig() = %q, want %q", got, want)
	}
}

func TestRenderGenericClientConfigSnippet(t *testing.T) {
	got, err := RenderClientConfigSnippet(DefaultMCPServerName, NewServeCommand(""))
	if err != nil {
		t.Fatalf("RenderClientConfigSnippet() error = %v", err)
	}

	const want = "{\n  \"mcpServers\": {\n    \"optimusctx\": {\n      \"command\": \"optimusctx\",\n      \"args\": [\n        \"run\"\n      ]\n    }\n  }\n}\n"
	if got != want {
		t.Fatalf("RenderClientConfigSnippet() = %q, want %q", got, want)
	}
}

func TestNormalizeClaudeCLIScope(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{name: "default local", input: "", want: ClaudeCLIScopeLocal},
		{name: "trim and lowercase", input: " Local ", want: ClaudeCLIScopeLocal},
		{name: "project", input: "project", want: ClaudeCLIScopeProject},
		{name: "user", input: "USER", want: ClaudeCLIScopeUser},
		{name: "unsupported", input: "workspace", wantErr: `unsupported Claude CLI scope "workspace"; expected local, project, or user`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeClaudeCLIScope(tt.input)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("NormalizeClaudeCLIScope() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeClaudeCLIScope() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeClaudeCLIScope() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderClaudeCLIAddCommand(t *testing.T) {
	tests := []struct {
		name       string
		serverName string
		scope      string
		command    ServeCommand
		want       string
	}{
		{
			name:    "default local scope",
			scope:   ClaudeCLIScopeLocal,
			command: NewServeCommand(""),
			want:    "claude mcp add --transport stdio --scope local optimusctx -- optimusctx run",
		},
		{
			name:       "explicit project scope and binary",
			serverName: "",
			scope:      ClaudeCLIScopeProject,
			command:    NewServeCommand("/tmp/optimusctx"),
			want:       "claude mcp add --transport stdio --scope project optimusctx -- /tmp/optimusctx run",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderClaudeCLIAddCommand(tt.serverName, tt.scope, tt.command)
			if got != tt.want {
				t.Fatalf("RenderClaudeCLIAddCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}
