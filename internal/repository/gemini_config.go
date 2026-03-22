package repository

import (
	"encoding/json"
	"fmt"
	"strings"
)

const geminiMCPServersKey = "mcpServers"

type geminiConfigDocument map[string]any

func RenderGeminiConfig(serverName string, command ServeCommand) (string, error) {
	document := geminiConfigDocument{
		geminiMCPServersKey: map[string]any{
			normalizeGeminiServerName(serverName): renderableGeminiServeCommand(command),
		},
	}
	return renderGeminiDocument(document)
}

func MergeGeminiConfig(existing []byte, serverName string, command ServeCommand) (string, error) {
	document, err := parseGeminiConfig(existing)
	if err != nil {
		return "", err
	}
	serverTable, err := document.serverTable()
	if err != nil {
		return "", err
	}
	serverTable[normalizeGeminiServerName(serverName)] = renderableGeminiServeCommand(command)
	document[geminiMCPServersKey] = map[string]any(serverTable)
	return renderGeminiDocument(document)
}

func GeminiConfigServerNames(existing []byte) (map[string]bool, error) {
	document, err := parseGeminiConfig(existing)
	if err != nil {
		return nil, err
	}
	serverTable, err := document.serverTable()
	if err != nil {
		return nil, err
	}
	names := make(map[string]bool, len(serverTable))
	for name := range serverTable {
		names[name] = true
	}
	return names, nil
}

func parseGeminiConfig(existing []byte) (geminiConfigDocument, error) {
	document := geminiConfigDocument{}
	if len(strings.TrimSpace(string(existing))) > 0 {
		if err := json.Unmarshal(existing, &document); err != nil {
			return nil, fmt.Errorf("decode gemini config: %w", err)
		}
	}
	if document == nil {
		document = geminiConfigDocument{}
	}
	return document, nil
}

func (d geminiConfigDocument) serverTable() (map[string]any, error) {
	if d == nil {
		return map[string]any{}, nil
	}
	raw, ok := d[geminiMCPServersKey]
	if !ok || raw == nil {
		return map[string]any{}, nil
	}
	serverTable, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("decode gemini config: %q must be an object", geminiMCPServersKey)
	}
	return serverTable, nil
}

func renderGeminiDocument(document geminiConfigDocument) (string, error) {
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode gemini config: %w", err)
	}
	return string(encoded) + "\n", nil
}

func normalizeGeminiServerName(serverName string) string {
	serverName = strings.TrimSpace(serverName)
	if serverName == "" {
		return DefaultMCPServerName
	}
	return serverName
}

func renderableGeminiServeCommand(command ServeCommand) map[string]any {
	command = NormalizeServeCommand(command)

	args := make([]string, len(command.Args))
	copy(args, command.Args)

	return map[string]any{
		"command": command.Command,
		"args":    args,
	}
}
