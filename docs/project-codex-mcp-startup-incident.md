# Project Incident: Codex MCP Startup Failure

This document records the project-side investigation, root cause, fix, and verification trail for the Codex-specific MCP startup failure found on `2026-03-21`.

This is intentionally project and operator documentation, not product-facing usage guidance.

## Failure summary

Codex could discover an `optimusctx` MCP registration but still leave the server stuck in startup instead of reaching a healthy connected state.

Typical Codex UI symptoms:

- `Booting MCP server: optimusctx`
- `MCP startup interrupted. The following servers were not initialized: optimusctx`
- `MCP client for optimusctx failed to start: MCP startup failed: handshaking with MCP server failed: connection closed: initialize response`

Typical runtime-side symptoms:

- `optimusctx run` printed `optimusctx mcp: ready for stdio requests`
- `optimusctx status` or `.optimusctx/mcp-activity.json` showed `last_session_start_at`
- but `last_initialize_at` and `last_tools_list_at` did not move forward

That pattern meant the host launched the process but did not finish the MCP handshake.

## Root cause

The Codex stdio transport used during this failure mode did not speak the same framing dialect as the original OptimusCtx MCP runtime.

Observed behavior during investigation on `2026-03-21`:

- Codex launched `optimusctx run` from the expected repository directory.
- Codex wrote line-delimited JSON requests to stdin, for example:

```json
{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"codex-mcp-client","title":"Codex","version":"0.116.0"}}}
```

- The original OptimusCtx runtime only accepted framed input using `Content-Length: ...`.
- After adding line-delimited input support, Codex still failed because the runtime replied using `Content-Length` framing while Codex expected line-delimited JSON responses too.

So the actual failure was a bidirectional stdio transport mismatch:

1. Codex sent line-delimited JSON.
2. OptimusCtx originally expected `Content-Length` frames only.
3. After parsing input, OptimusCtx still answered in `Content-Length` frames.
4. Codex then closed the connection waiting for an `initialize` response in its own transport format.

## Primary fix

Source fix:

- commit `a8e3f43` `fix: support codex line-delimited mcp transport`

What changed:

- the MCP server now autodetects whether the client is using:
  - `Content-Length` framed stdio, or
  - line-delimited JSON stdio
- the server responds in the same format it received

This preserves the existing framed transport while allowing Codex's line-delimited startup flow to complete.

## Related fixes from the same investigation

These were separate issues found during the same debugging pass. They improved startup behavior or prevented adjacent regressions, but they did not replace the transport fix above.

- `db1ccb2` preserves user content when managed guidance updates `AGENTS.md`
- `0392e33` keeps watch refresh reports off MCP stdout
- `7e6d8f2` starts the MCP server before init and refresh bootstrap work
- `f354531` starts the server before background `health` completes
- `96b0a62` batches git ignore checks during discovery
- `2a8c8fd` uses shallower repository resolution on startup
- `97a744b` reuses the resolved root for bootstrap health
- `866c8b3` reuses the resolved root for watch startup

Later follow-up from the same operator debugging thread:

- `5b8c162` stops treating WSL `~/.codex/config.toml` as the shared `Codex App` default when the app actually lives on Windows.
- In mixed WSL plus Windows setups, `Codex App` may read `C:\Users\<user>\.codex\config.toml` while `optimusctx init` running inside WSL would otherwise write `/home/<user>/.codex/config.toml`.
- The fix now requires an explicit Windows-backed `--config` path when the `Codex App` shared config cannot be inferred safely from WSL.

## Verification trail

The important verification outcome was not only that Codex could list the server in config, but that it could actually complete MCP startup and tools discovery.

Runtime evidence used during investigation:

- `last_session_start_at` proved Codex launched the process
- `last_initialize_at` proved the handshake completed
- `last_tools_list_at` proved Codex completed discovery

After the transport fix and installed-runtime replacement, `.optimusctx/mcp-activity.json` advanced all three timestamps during a real Codex startup from the repository.

## Installed runtime note

If you use the npm-installed wrapper, source changes alone are not enough.

The wrapper delegates into a package-local runtime binary, so the installed runtime must be rebuilt or replaced too. Otherwise Codex will continue launching the old broken binary even when the repository source already contains the fix.

## Distinguishing this failure from other startup issues

Do not collapse every Codex startup issue into the same bucket.

Useful separation:

- `last_session_start_at` changes, but `last_initialize_at` does not:
  suspect transport incompatibility or host-side startup failure before handshake completion
- `last_initialize_at` changes, but `last_tools_list_at` does not:
  the handshake started but discovery did not finish
- all three advance:
  Codex startup and discovery completed successfully

## Manual debugging checklist

Use this order:

1. Run `codex mcp get optimusctx`.
2. Run `optimusctx status`.
3. Inspect `.optimusctx/mcp-activity.json`.
4. If `last_session_start_at` changes without `last_initialize_at`, suspect transport incompatibility or host-side startup failure.
5. If startup still fails after upgrading, confirm the binary used by Codex is the rebuilt runtime, not an older installed runtime.
