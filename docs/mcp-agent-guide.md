# OptimusCtx MCP Agent Guide

Use this guide to get real value from the registered `optimusctx` MCP server instead of treating it like a generic file dumper.

## Runtime ownership

When `optimusctx init` registers a supported MCP client correctly, the host should launch `optimusctx run` automatically when it connects.

When the host supports durable agent guidance, `init --write` also installs it where the host actually reads instructions:

- Codex: the active `AGENTS.md` or `AGENTS.override.md`
- Claude CLI: `.claude/rules/optimusctx-mcp.md` or `~/.claude/rules/optimusctx-mcp.md`
- Claude Desktop: no durable guidance file is managed there, so only MCP registration is written

Manual `optimusctx run` is still useful, but mainly for:

- direct STDIO testing
- debugging MCP startup
- confirming the runtime can boot outside the host

If you run it yourself, the expected ready signal is:

```text
optimusctx mcp: ready for stdio requests
```

## What the agent gets

The registered MCP server exposes these tool families:

- `optimusctx.repository_map` for bounded repo orientation
- `optimusctx.layered_context_l0` and `optimusctx.layered_context_l1` for persisted repository context envelopes
- `optimusctx.symbol_lookup` and `optimusctx.structure_lookup` for exact lookup instead of blind grep-style exploration
- `optimusctx.targeted_context` for bounded excerpts around exact files and ranges
- `optimusctx.pack` for assembling one deterministic working set from smaller query surfaces
- `optimusctx.health` and `optimusctx.refresh` for runtime state and freshness control
- `optimusctx.token_tree` for spotting heavy paths before pulling too much context

## Recommended usage pattern

For most agent tasks, the highest-signal order is:

1. Start with `optimusctx.health` if runtime state is unclear.
2. Use `optimusctx.repository_map` to orient on directories, files, and symbols.
3. Use `optimusctx.symbol_lookup` or `optimusctx.structure_lookup` to jump to exact functions, types, or paths.
4. Use `optimusctx.targeted_context` to pull only the bounded code you need.
5. Use `optimusctx.pack` when the task needs a composed context bundle rather than one lookup.
6. Use `optimusctx.refresh` only when the repository is stale or after meaningful local changes.

That pattern is the point of OptimusCtx: smaller exact lookups first, broader context only when justified.

## Why this helps

Compared with repeated full-file scanning, the MCP surface is designed to:

- reduce blind exploration
- cut prompt size by preferring exact, bounded retrieval
- preserve repository knowledge across sessions
- improve response speed by reusing persisted maps and lookup state

## How to verify discovery and real usage

There are three useful checks, and they do not all come from the same place:

### 1. Registration check

Run `optimusctx init` in the repository and register your host there, or use the explicit `--client ... --write` path. The result should tell you where the host config landed and, when supported, where the agent-guidance file landed.

`optimusctx status` can now verify this part directly by showing:

- detected host configs that reference OptimusCtx
- detected guidance files that were installed for supported hosts

### 2. Host discovery check

In the MCP host, confirm that `optimusctx.*` tools are available. A healthy discovery usually includes names like:

- `optimusctx.repository_map`
- `optimusctx.symbol_lookup`
- `optimusctx.targeted_context`
- `optimusctx.pack`
- `optimusctx.health`

### 3. Real usage check

To prove the agent is not only discovering the server but actually using it, inspect the host transcript, tool timeline, or debug logs and look for:

- an MCP `tools/list` response that includes `optimusctx.*`
- actual tool calls such as `optimusctx.repository_map`, `optimusctx.symbol_lookup`, or `optimusctx.health`

`optimusctx status` closes part of this gap from the OptimusCtx side by persisting local evidence for:

- last MCP session start
- last `initialize`
- last `tools/list`
- recent `tools/call`

So the split is:

- the host proves what the agent UI or transcript saw
- `optimusctx status` proves what the runtime actually recorded locally

If `optimusctx status` shows registration but no initialize, the host never connected.
If it shows initialize and tools discovery but no tool calls, the host connected but the agent has not used the MCP yet.
If it shows recent tool calls, you have runtime-side evidence that the MCP is being used.

If the host is registered but no `optimusctx.*` tools appear and no tool calls are recorded, the agent is not using the MCP yet.
