---
phase: "38"
name: "Cursor CLI Native Onboarding"
created: 2026-03-22
status: passed
---

# Phase 38: Cursor CLI Native Onboarding — Verification

## Goal-Backward Verification

**Phase Goal:** add truthful Cursor CLI preview, write, and verification support using Cursor's documented shared `mcp.json` contract.

## Checks

| # | Requirement | Status | Evidence |
|---|------------|--------|----------|
| 1 | `CUR-01` | Passed | `optimusctx init --client cursor-cli` now previews and writes repo-local or shared Cursor `mcp.json` using the native `mcpServers` contract and `optimusctx run`. |
| 2 | `CUR-01` | Passed | Repeated Cursor writes preserve unrelated MCP entries and avoid duplicate `optimusctx` entries through the generic JSON merge path. |
| 3 | `CUR-02` | Passed | `optimusctx status`/doctor can detect Cursor config accurately while keeping durable guidance unsupported and the shared CLI/editor config story explicit. |

## Executed Evidence

- `go test ./internal/repository ./internal/app ./internal/cli`

## Result

Phase 38 passed. Cursor CLI is now a truthful first-class host with native onboarding, precise detection, and regression coverage, without over-claiming broader Cursor editor support.
