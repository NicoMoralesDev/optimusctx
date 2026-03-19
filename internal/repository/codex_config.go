package repository

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

const codexMCPServersTablePrefix = "[mcp_servers."

func RenderCodexConfig(serverName string, command ServeCommand) (string, error) {
	document := codexConfigDocument{
		codexMCPServersKey: codexServerTable{
			normalizeCodexServerName(serverName): renderableCodexServeCommand(command),
		},
	}

	return renderCodexDocument(document)
}

func MergeCodexConfig(existing []byte, serverName string, command ServeCommand) (string, error) {
	document := codexConfigDocument{}
	if len(strings.TrimSpace(string(existing))) > 0 {
		if err := toml.Unmarshal(existing, &document); err != nil {
			return "", fmt.Errorf("decode codex config: %w", err)
		}
	}
	if document == nil {
		document = codexConfigDocument{}
	}

	serverTable, err := document.serverTable()
	if err != nil {
		return "", err
	}
	serverTable[normalizeCodexServerName(serverName)] = renderableCodexServeCommand(command)
	document[codexMCPServersKey] = map[string]any(serverTable)

	return renderCodexDocument(document)
}

const codexMCPServersKey = "mcp_servers"

type codexConfigDocument map[string]any

type codexServerTable map[string]any

func (d codexConfigDocument) serverTable() (codexServerTable, error) {
	if d == nil {
		return codexServerTable{}, nil
	}

	raw, ok := d[codexMCPServersKey]
	if !ok || raw == nil {
		return codexServerTable{}, nil
	}

	serverTable, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("decode codex config: %q must be a table", codexMCPServersKey)
	}

	return codexServerTable(serverTable), nil
}

func normalizeCodexServerName(serverName string) string {
	serverName = strings.TrimSpace(serverName)
	if serverName == "" {
		return DefaultMCPServerName
	}
	return serverName
}

func renderableCodexServeCommand(command ServeCommand) map[string]any {
	command = NormalizeServeCommand(command)

	args := make([]string, len(command.Args))
	copy(args, command.Args)

	return map[string]any{
		"command": command.Command,
		"args":    args,
	}
}

func renderCodexDocument(document codexConfigDocument) (string, error) {
	var builder strings.Builder
	if err := writeCodexTable(&builder, nil, map[string]any(document)); err != nil {
		return "", fmt.Errorf("encode codex config: %w", err)
	}
	return builder.String(), nil
}

func writeCodexTable(builder *strings.Builder, path []string, table map[string]any) error {
	scalarKeys, tableKeys := splitCodexTableKeys(table)

	if len(path) > 0 {
		if builder.Len() > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString("[")
		for i, key := range path {
			if i > 0 {
				builder.WriteString(".")
			}
			builder.WriteString(renderCodexKey(key))
		}
		builder.WriteString("]\n")
	}

	for _, key := range scalarKeys {
		rendered, err := renderCodexValue(table[key])
		if err != nil {
			return fmt.Errorf("render key %q: %w", key, err)
		}
		builder.WriteString(renderCodexKey(key))
		builder.WriteString(" = ")
		builder.WriteString(rendered)
		builder.WriteString("\n")
	}

	for _, key := range tableKeys {
		child, err := codexTableValue(table[key])
		if err != nil {
			return fmt.Errorf("render table %q: %w", key, err)
		}

		nextPath := append(append([]string(nil), path...), key)
		if err := writeCodexTable(builder, nextPath, child); err != nil {
			return err
		}
	}

	return nil
}

func splitCodexTableKeys(table map[string]any) ([]string, []string) {
	scalarKeys := make([]string, 0, len(table))
	tableKeys := make([]string, 0, len(table))

	for key, value := range table {
		if _, ok := value.(map[string]any); ok {
			tableKeys = append(tableKeys, key)
			continue
		}
		if _, ok := value.(codexServerTable); ok {
			tableKeys = append(tableKeys, key)
			continue
		}
		scalarKeys = append(scalarKeys, key)
	}

	sort.Strings(scalarKeys)
	sort.Strings(tableKeys)

	return scalarKeys, tableKeys
}

func codexTableValue(value any) (map[string]any, error) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, nil
	case codexServerTable:
		return map[string]any(typed), nil
	default:
		return nil, fmt.Errorf("unexpected type %T", value)
	}
}

func renderCodexKey(key string) string {
	if key == "" {
		return `""`
	}
	for _, r := range key {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return strconv.Quote(key)
	}
	return key
}

func renderCodexValue(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return strconv.Quote(typed), nil
	case bool:
		if typed {
			return "true", nil
		}
		return "false", nil
	case int:
		return strconv.FormatInt(int64(typed), 10), nil
	case int8:
		return strconv.FormatInt(int64(typed), 10), nil
	case int16:
		return strconv.FormatInt(int64(typed), 10), nil
	case int32:
		return strconv.FormatInt(int64(typed), 10), nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case uint:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint64:
		return strconv.FormatUint(typed, 10), nil
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), nil
	case time.Time:
		return typed.Format(time.RFC3339Nano), nil
	case toml.LocalDate:
		return typed.String(), nil
	case toml.LocalTime:
		return typed.String(), nil
	case toml.LocalDateTime:
		return typed.String(), nil
	case []string:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			rendered, err := renderCodexValue(item)
			if err != nil {
				return "", err
			}
			items = append(items, rendered)
		}
		return "[" + strings.Join(items, ", ") + "]", nil
	case []any:
		return renderCodexArray(typed)
	case map[string]any:
		return renderCodexInlineTable(typed)
	default:
		return "", fmt.Errorf("unsupported value type %T", value)
	}
}

func renderCodexArray(values []any) (string, error) {
	items := make([]string, 0, len(values))
	for _, value := range values {
		rendered, err := renderCodexValue(value)
		if err != nil {
			return "", err
		}
		items = append(items, rendered)
	}
	return "[" + strings.Join(items, ", ") + "]", nil
}

func renderCodexInlineTable(table map[string]any) (string, error) {
	keys := make([]string, 0, len(table))
	for key := range table {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		rendered, err := renderCodexValue(table[key])
		if err != nil {
			return "", err
		}
		parts = append(parts, renderCodexKey(key)+" = "+rendered)
	}

	return "{ " + strings.Join(parts, ", ") + " }", nil
}
