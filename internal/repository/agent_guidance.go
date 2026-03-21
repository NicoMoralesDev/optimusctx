package repository

import (
	"fmt"
	"strings"
)

const (
	managedGuidanceBegin = "<!-- optimusctx:guidance-begin -->"
	managedGuidanceEnd   = "<!-- optimusctx:guidance-end -->"
	ClaudeRulesFilename  = "optimusctx-mcp.md"
)

type RenderedGuidance struct {
	Label          string
	Path           string
	Mode           RenderMode
	Content        string
	AppliedContent string
	Notes          []string
}

func (r RenderedGuidance) ContentForWrite() string {
	if strings.TrimSpace(r.AppliedContent) != "" {
		return r.AppliedContent
	}
	return r.Content
}

func RenderCodexGuidanceBlock() string {
	lines := []string{
		managedGuidanceBegin,
		"## OptimusCtx MCP guidance",
		"",
		"- When `optimusctx.*` tools are available, prefer them over broad repository exploration.",
		"- Start with `optimusctx.symbol_lookup` or `optimusctx.structure_lookup` when you know the symbol, file, or path you need.",
		"- Use `optimusctx.repository_map` only to orient yourself when the exact target is still unclear.",
		"- After narrowing the target, use `optimusctx.targeted_context` to fetch bounded context instead of large reads.",
		"- Use `optimusctx.layered_context_l0` or `optimusctx.layered_context_l1` for high-level exploration, not as the default when exact lookup is possible.",
		"- Use `optimusctx.pack` only after you already know which files or artifacts need to be bundled together.",
		"- If the runtime looks stale or a call fails unexpectedly, check `optimusctx.health` first and call `optimusctx.refresh` only when needed.",
		managedGuidanceEnd,
	}
	return strings.Join(lines, "\n") + "\n"
}

func RenderClaudeGuidanceDocument() string {
	lines := []string{
		"# OptimusCtx MCP guidance",
		"",
		"- When `optimusctx.*` tools are available, prefer them over broad repository exploration.",
		"- Start with `optimusctx.symbol_lookup` or `optimusctx.structure_lookup` when you know the exact symbol, file, or path you need.",
		"- Use `optimusctx.repository_map` only to orient yourself when the exact target is still unclear.",
		"- Once you have the target, use `optimusctx.targeted_context` to fetch bounded context instead of large reads.",
		"- Use `optimusctx.layered_context_l0` or `optimusctx.layered_context_l1` for high-level exploration only.",
		"- Use `optimusctx.pack` after narrowing scope and only when you need a composed evidence bundle.",
		"- If the runtime looks stale or a call fails unexpectedly, check `optimusctx.health` first and call `optimusctx.refresh` only when necessary.",
	}
	return strings.Join(lines, "\n") + "\n"
}

func MergeManagedGuidance(existing []byte, block string) (string, error) {
	content := string(existing)
	block = strings.TrimSpace(block)
	if block == "" {
		return "", fmt.Errorf("merge managed guidance: block is required")
	}

	if len(existing) == 0 {
		return block + "\n", nil
	}

	start := strings.Index(content, managedGuidanceBegin)
	end := strings.Index(content, managedGuidanceEnd)
	if start >= 0 && end >= start {
		end += len(managedGuidanceEnd)
		replaced := content[:start] + block + content[end:]
		if !strings.HasSuffix(replaced, "\n") {
			replaced += "\n"
		}
		return replaced, nil
	}

	var builder strings.Builder
	builder.WriteString(content)
	switch {
	case strings.HasSuffix(content, "\n\n"):
	case strings.HasSuffix(content, "\n"):
		builder.WriteString("\n")
	default:
		builder.WriteString("\n\n")
	}
	builder.WriteString(block)
	builder.WriteString("\n")
	return builder.String(), nil
}
