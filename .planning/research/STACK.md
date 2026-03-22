# Stack Research

**Domain:** Next first-class MCP hosts after Claude and Codex
**Researched:** 2026-03-21
**Confidence:** HIGH

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.26.x | Core runtime, CLI surface, host integration logic | Already shipped stack; this milestone is host integration depth, not a platform rewrite |
| `encoding/json` | stdlib | Gemini CLI `settings.json` and Cursor `mcp.json` merge/render flows | Both candidate hosts use documented JSON contracts, so the existing JSON path should be extended rather than replaced |
| Host capability metadata | existing app/repository layers | Keep preview/write semantics, scope truth, and diagnostics explicit per supported client | The real gap is contract accuracy per host, so the capability/adapter boundary should stay the integration seam |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Existing filesystem/path helpers | current repo | Reuse repo-local/shared/default path resolution | Needed because host correctness is partly about writing to the right file in mixed environments |
| Targeted JSON schema wrappers | current repo patterns | Keep host-specific render/write logic explicit even when both use JSON | Needed so Gemini CLI and Cursor CLI can diverge cleanly on path and notes without forking the whole install service |
| Existing Go test harnesses | current repo | Regression coverage for adapter and CLI behavior | Needed to lock host-specific preview/write output without broad E2E fragility |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| Gemini CLI docs | Verify `settings.json`, `mcpServers`, and repo/shared config scope | Official source for Gemini CLI MCP setup |
| Cursor docs | Verify `mcp.json`, CLI commands, and shared editor/CLI config contract | Official source for Cursor CLI MCP support |
| Existing Go test suite | Validate new adapters without changing runtime core behavior | Extend targeted install/status/init coverage first, then run full suite |

## Installation

```bash
# Core verification
go test ./internal/app ./internal/cli ./internal/repository

# Full regression after implementation
go test ./...
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| JSON-backed host adapters with shared helpers | One universal JSON writer with only note-text differences | Only if host paths, scopes, and diagnostics turn out to be identical, which research suggests they are not |
| Existing JSON merge flow with host-specific wrappers | New parser or schema layer | Only if Gemini or Cursor config shapes prove materially more complex than the current MCP server object model |
| Structured merge/write | Manual string concatenation | Acceptable for preview text only, not for persisted writes |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Treating Gemini CLI and Cursor CLI as "just another generic JSON host" | Both use JSON, but the path, scope, and support wording differ materially | Host-specific preview and write adapters with shared helpers |
| Assuming Linux shared-config paths for desktop/editor-backed hosts from WSL | Recent fixes showed this silently targets the wrong file | Resolve the real target path or require explicit `--config` |
| Making `--write` implicit | Violates the shipped explicit-consent boundary around client config writes | Keep preview-first and write-only on operator request |

## Stack Patterns by Variant

**If the host stores JSON MCP config with repo/shared variants:**
- Use the current typed JSON merge-and-render path with host-specific path resolvers and notes.
- Because Gemini CLI and Cursor CLI both fit the local JSON merge pattern but differ on path semantics.

**If the host shares CLI and editor config:**
- Keep the adapter explicit about the verified entrypoint while reusing the shared path/backend.
- Because Cursor docs say CLI and editor share MCP config, but support claims still need a precise boundary.

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| `optimusctx run` | Gemini CLI `mcpServers` JSON entry | Official docs define local MCP servers via `settings.json` |
| `optimusctx run` | Cursor `mcp.json` stdio server entry | Official docs define local stdio MCP servers through `mcp.json` |
| `optimusctx run` | Existing Claude/Codex contracts | Must remain stable while the host matrix expands |

## Sources

- https://geminicli.com/docs/tools/mcp-server — verified `mcpServers`, `settings.json`, and global MCP settings
- https://geminicli.com/docs/cli/tutorials/mcp-setup/ — verified repo-local and shared `settings.json` targeting
- https://docs.cursor.com/cli/mcp — verified Cursor CLI MCP support and management commands
- https://docs.cursor.com/advanced/model-context-protocol — verified `mcp.json` JSON contract and stdio server fields
- Local code: `internal/app/install.go`, `internal/repository/client_config.go` — verified current supported-host registry and install-service seams

---
*Stack research for: next first-class MCP hosts after Claude and Codex*
*Researched: 2026-03-21*
