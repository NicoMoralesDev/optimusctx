# Stack Research

**Domain:** MCP client compatibility for local coding-agent hosts
**Researched:** 2026-03-19
**Confidence:** HIGH

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.26.x | Core runtime, CLI surface, host integration logic | Already shipped stack; this milestone is integration depth, not a platform rewrite |
| `encoding/json` | stdlib | Claude Desktop config merge/render and JSON preview flows | Existing shipped path already proves the JSON merge model for Claude Desktop |
| Host-specific registration adapters | existing app/repository layers | Keep preview/write semantics explicit per supported client | The real gap is contract accuracy per host, so the adapter boundary should stay the integration seam |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| One maintained Go TOML library | current v2 line | Read, merge, and write Codex `config.toml` safely | Needed for `codex-app` and `codex-cli`, whose official MCP config lives in `config.toml` |
| `os/exec` | stdlib | Invoke Claude CLI's official registration command when `--write` is requested | Preferred if direct mutation of Claude user config remains underdocumented |
| Existing Go test harnesses | current repo | Regression coverage for adapter and CLI behavior | Needed to lock host-specific preview/write output without broad E2E fragility |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| Anthropic Claude Code docs | Verify Claude CLI scope and registration semantics | Official source for `claude mcp add` and `claude mcp add-json` |
| OpenAI Codex docs | Verify Codex config path, schema, and shared app/CLI behavior | Official source for `config.toml` and `[mcp_servers.<name>]` |
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
| Shared Codex `config.toml` adapter for app and CLI | Separate `codex-app` and `codex-cli` writers | Only if implementation proves a real host divergence the docs do not show today |
| Claude CLI command-driven write path | Direct edits to `~/.claude.json` | Only if the exact user-scope file schema is validated during implementation and is safer than invoking Claude |
| Structured TOML merge/write | Manual string concatenation | Acceptable for preview text only, not for persisted writes |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Treating all supported hosts as one JSON shape | Claude and Codex do not use the same persisted config format | Host-specific preview and write adapters |
| Blind edits to undocumented Claude CLI user config | Official docs emphasize CLI registration commands and scopes, not hand-editing the user file | Use Claude's supported registration mechanism unless implementation proves otherwise |
| Making `--write` implicit | Violates the shipped explicit-consent boundary around client config writes | Keep preview-first and write-only on operator request |

## Stack Patterns by Variant

**If the host stores JSON MCP config:**
- Use the current typed JSON merge-and-render path.
- Because Claude Desktop already has a working deterministic implementation.

**If the host stores TOML MCP config:**
- Use a typed TOML adapter with merge semantics and shared path resolution.
- Because Codex App and Codex CLI officially share `config.toml`.

**If the host documents command-first registration:**
- Wrap that host command behind `optimusctx ... --write`.
- Because the host CLI is the authoritative writer for its own config surface.

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| `optimusctx run` | Claude Desktop `mcpServers` JSON entry | Existing shipped path; must remain stable |
| `optimusctx run` | Claude CLI stdio registration via `claude mcp add-json` | Official docs support JSON-based stdio server registration |
| `optimusctx run` | Codex `[mcp_servers.<server-name>]` TOML table | Official docs define this as the native Codex MCP contract |

## Sources

- https://code.claude.com/docs/en/mcp — verified `claude mcp add`, `claude mcp add-json`, scopes, `.mcp.json`, and `~/.claude.json` storage notes
- https://developers.openai.com/codex/mcp — verified `config.toml`, `~/.codex/config.toml`, project-scoped `.codex/config.toml`, CLI commands, and `[mcp_servers.<name>]`
- https://developers.openai.com/codex/app/settings — verified Codex App shares `config.toml`-backed agent and MCP settings with Codex CLI and the IDE extension
- Local code: `internal/app/install.go`, `internal/repository/client_config.go` — verified current write support exists only for Claude Desktop

---
*Stack research for: MCP client compatibility for local coding-agent hosts*
*Researched: 2026-03-19*
