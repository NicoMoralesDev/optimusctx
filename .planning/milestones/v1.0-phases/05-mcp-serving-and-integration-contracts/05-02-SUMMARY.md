---
phase: 05-mcp-serving-and-integration-contracts
plan: "02"
subsystem: api
tags: [mcp, json-rpc, repository-map, layered-context, lookup]
requires:
  - phase: 05-01
    provides: MCP STDIO transport, tool registry, and JSON-RPC session handling
provides:
  - Read-only MCP tools for repository map, layered context L0/L1, symbol lookup, structure lookup, and targeted context
  - Shared machine-readable metadata envelopes with freshness, cache status, and bounds metadata
  - Structured field-specific validation and bounds errors for read-only MCP requests
affects: [05-03, 05-04, 05-05, mcp-tool-contracts]
tech-stack:
  added: []
  patterns: [thin MCP handlers over app services, shared query envelopes, field-specific validation errors]
key-files:
  created: [internal/mcp/errors.go, internal/mcp/query_tools.go, internal/mcp/query_tools_test.go]
  modified: [internal/mcp/protocol.go, internal/mcp/server.go, internal/mcp/server_test.go]
key-decisions:
  - "Read-only MCP tools return one shared structuredContent envelope that wraps existing app-layer result structs instead of flattening transport-specific payloads."
  - "Query handlers enforce transport bounds at the MCP edge and report field-specific maximum, minimum, required, and conflict errors."
  - "The default MCP server registers the read-only query tools eagerly so clients discover a truthful tool surface from tools/list."
patterns-established:
  - "Thin MCP adapter: decode args, normalize bounds, delegate to app service, then wrap the service result in shared metadata."
  - "Structured MCP errors: validation failures identify the field, constraint, and received value instead of returning opaque strings."
requirements-completed: [MCP-02, MCP-03, MCP-04]
duration: 2min
completed: 2026-03-15
---

# Phase 5 Plan 02: Read-Only MCP Query Tools Summary

**Read-only MCP query tools now expose persisted repository map, layered context, exact lookup, and targeted context results with one shared metadata envelope and explicit bounds errors**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-15T15:06:49Z
- **Completed:** 2026-03-15T15:08:50Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments
- Added shared MCP query-result metadata and structured error helpers for freshness, cache status, and field-specific validation failures.
- Implemented read-only MCP handlers for repository map, L0/L1 context, symbol lookup, structure lookup, and targeted context by delegating to the existing app-layer services.
- Added contract coverage for repository queries, lookup queries, and transparent bounds failures across the new MCP surface.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add shared query-tool metadata and structured error contracts** - `80b5877` (feat)
2. **Task 2: Implement read-only MCP query handlers over existing app services** - `72474e6` (feat)
3. **Task 3: Add coverage for payload structure, metadata, and oversized requests** - `ab5346f` (test)

## Files Created/Modified

- `internal/mcp/protocol.go` - Extends MCP call results with structured content and shared query metadata types.
- `internal/mcp/errors.go` - Defines field-specific validation and bounds error helpers for MCP tools.
- `internal/mcp/query_tools.go` - Implements the read-only MCP tool registry, request normalization, service delegation, and payload adaptation.
- `internal/mcp/query_tools_test.go` - Verifies metadata consistency, lookup behavior, and structured bound failures through the MCP surface.
- `internal/mcp/server.go` - Registers the read-only query tools on the default MCP server.
- `internal/mcp/server_test.go` - Updates registry assertions for the expanded default tool surface.

## Decisions Made

- Read-only MCP tools return `structuredContent` plus a JSON content item so callers get machine-readable payloads without duplicating transport-specific result structs.
- Repository-map bounds are applied during MCP adaptation, while layered context and lookup tools preserve the existing app-service semantics and surface the applied limits in metadata.
- Validation failures are normalized into explicit `required`, `minimum`, `maximum`, and `conflict` contracts so MCP callers can respond programmatically.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The first verification run used an empty Go module cache with `GOPROXY=off`; rerunning against the populated `/home/nico/go/pkg/mod` cache resolved the environment issue without changing code.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 5 now has a stable read-only MCP query surface that later token-tree, health, pack, and mutating tools can extend.
- The next MCP plans can reuse the shared query envelope and structured error helpers instead of inventing new per-tool response contracts.

## Self-Check: PASSED

---
*Phase: 05-mcp-serving-and-integration-contracts*
*Completed: 2026-03-15*
