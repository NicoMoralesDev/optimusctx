---
phase: 05-mcp-serving-and-integration-contracts
plan: "05"
subsystem: api
tags: [mcp, stdio, go, refresh, pack, health, token-tree, testing]
requires:
  - phase: 05-02
    provides: Read-only MCP query handlers and shared structured envelopes
  - phase: 05-03
    provides: Transport-neutral token tree service and bounded hierarchical results
  - phase: 05-04
    provides: Transport-neutral health and pack services for MCP exposure
provides:
  - Canonical MCP registry covering repository map, layered context, lookup, targeted context, refresh, token tree, pack, and health
  - Operational MCP handlers that reuse app services and preserve shared freshness and cache metadata
  - End-to-end stdio session coverage across read-only calls, refresh calls, and bounded failures
affects: [phase-06, mcp-clients, cli]
tech-stack:
  added: []
  patterns: [canonical-mcp-registry, shared-query-envelope, stdio-session-verification]
key-files:
  created: [internal/mcp/registry.go, internal/mcp/ops_tools.go, internal/mcp/integration_test.go]
  modified: [internal/mcp/server.go, internal/mcp/query_tools.go, internal/mcp/query_tools_test.go, internal/mcp/server_test.go, internal/cli/mcp.go]
key-decisions:
  - "The MCP server now builds its tool surface from one canonical registry so tools/list and tool invocation cannot drift."
  - "Refresh, token tree, pack, and health reuse the same QueryEnvelope metadata contract as read-only tools; refresh reports cache status as refresh_attempted."
  - "Server-boundary verification uses real stdio framing plus repository-backed tool calls instead of synthetic handler-only assertions."
patterns-established:
  - "Canonical registry: New MCP capabilities register through internal/mcp/registry.go and are exposed automatically by NewServer."
  - "Operational handler envelope: mutating and diagnostic tools return structuredContent via newStructuredToolResult instead of bespoke transport payloads."
requirements-completed: [MCP-01, MCP-02, MCP-03, MCP-04]
duration: 12min
completed: 2026-03-15
---

# Phase 05 Plan 05: MCP Serving and Integration Contracts Summary

**Unified MCP registry with refresh, token tree, pack, and health handlers plus real stdio session proof across the full Phase 5 tool surface**

## Performance

- **Duration:** 12 min
- **Started:** 2026-03-15T15:10:00Z
- **Completed:** 2026-03-15T15:21:52Z
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments

- Centralized the full Phase 5 MCP tool surface in one registry so `tools/list` reflects the actual server contract.
- Added operational MCP handlers for refresh, token tree, pack, and health on top of existing app services with shared structured envelopes and bounded validation.
- Proved the server boundary over stdio with one real session covering initialize, tools/list, a read-only query, a refresh call, and a structured bounded failure.

## Task Commits

Each task was committed atomically:

1. **Task 1: Register the complete Phase 5 MCP tool surface** - `4933aeb` (feat)
2. **Task 2: Implement operational MCP handlers for refresh, token tree, pack, and health** - `32fe211` (test)
3. **Task 3: Add end-to-end MCP integration coverage** - `42b5766` (test)

## Files Created/Modified

- `internal/mcp/registry.go` - Canonical registry for all Phase 5 MCP tool handlers.
- `internal/mcp/ops_tools.go` - MCP adapters for refresh, token tree, pack, and health with shared envelope mapping and bounded errors.
- `internal/mcp/integration_test.go` - Handler-level operational coverage and real stdio session proof across the server boundary.
- `internal/mcp/server.go` - Default server wiring now delegates to the canonical registry and exposes a shared stdio entrypoint.
- `internal/cli/mcp.go` - CLI serve command now delegates to the shared stdio server helper.

## Decisions Made

- The server now registers tools through one registry function instead of per-file eager registration so future CLI or documentation updates can follow one source of truth.
- Operational tools return the same `QueryEnvelope` shape as earlier query tools, keeping freshness and cache status machine-readable across read-only and mutating flows.
- The new integration test exercises actual frame parsing and request routing so Phase 5 verification is not limited to direct handler invocation.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Default sandbox Go cache paths under `/home/nico/.cache` were not writable, so verification used `GOCACHE=/tmp/optimusctx-gocache`.
- Network access for module downloads was blocked, so verification reused the existing module cache with `GOMODCACHE=/home/nico/go/pkg/mod GOPROXY=off`.
- Restoring `internal/mcp/server_test.go` exposed a duplicate `TestMCPServerStdioSession`; the older narrow transport test was renamed to keep both unit and integration coverage.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 5 now has one coherent MCP registry and verified end-to-end transport coverage for the full promised tool surface.
- Phase 06 can build install registration and operator workflows on top of a stable stdio server and complete tool contract.

## Self-Check

PASSED

- Found `.planning/phases/05-mcp-serving-and-integration-contracts/05-05-SUMMARY.md`
- Verified task commits `4933aeb`, `32fe211`, and `42b5766` in `git log`

---
*Phase: 05-mcp-serving-and-integration-contracts*
*Completed: 2026-03-15*
