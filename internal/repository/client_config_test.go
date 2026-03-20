package repository

import "testing"

func TestSupportedClientsRemainExplicit(t *testing.T) {
	clients := SupportedClients()
	wantIDs := []ClientID{
		ClientClaudeDesktop,
		ClientClaudeCLI,
		ClientCodexApp,
		ClientCodexCLI,
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
