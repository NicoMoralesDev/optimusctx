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

func TestRenderClaudeCLIAddCommand(t *testing.T) {
	got := RenderClaudeCLIAddCommand("", NewServeCommand(""))
	const want = "claude mcp add --transport stdio optimusctx -- optimusctx run"
	if got != want {
		t.Fatalf("RenderClaudeCLIAddCommand() = %q, want %q", got, want)
	}
}
