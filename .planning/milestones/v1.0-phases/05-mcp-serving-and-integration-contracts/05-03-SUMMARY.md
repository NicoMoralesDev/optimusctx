---
phase: 05-mcp-serving-and-integration-contracts
plan: "03"
subsystem: api
tags: [go, sqlite, mcp, token-tree, persisted-metadata]
requires:
  - phase: 04-layered-context-exact-lookup-and-budget-analysis
    provides: shared bytes-to-token policy and layered context repository envelopes
provides:
  - transport-neutral token tree contracts with explicit depth and node bounds
  - persisted sqlite token tree reads assembled from directory and file size metadata
  - app-layer token tree service with deterministic path-scoped results
affects: [phase-05-mcp-serving-and-integration-contracts, mcp-tools, query-surface]
tech-stack:
  added: []
  patterns: [persisted hierarchical query assembly, bounded transport-neutral service outputs]
key-files:
  created: [internal/repository/token_tree.go, internal/store/sqlite/token_tree.go, internal/store/sqlite/token_tree_test.go, internal/app/token_tree.go, internal/app/token_tree_test.go]
  modified: []
key-decisions:
  - "Token tree estimation reuses the existing bytes_div_4_ceiling policy instead of introducing a model-specific tokenizer."
  - "Hierarchical token tree results order directories before files, then sort each group deterministically by size and path."
patterns-established:
  - "Persisted metadata only: token tree queries read from directories and files tables without request-time repository scanning."
  - "Bound summaries distinguish depth truncation from global node-limit truncation so later MCP handlers can surface actionable limits."
requirements-completed: [MCP-03, MCP-04]
duration: 6min
completed: 2026-03-15
---

# Phase 5 Plan 03: Token Tree Summary

**Hierarchical token-tree contracts and persisted SQLite reads exposed through a transport-neutral app service using the shared bytes-per-token policy**

## Performance

- **Duration:** 6 min
- **Started:** 2026-03-15T15:00:06Z
- **Completed:** 2026-03-15T15:05:50Z
- **Tasks:** 3
- **Files modified:** 5

## Accomplishments
- Added explicit token-tree request, bounds, summary, and node contracts under `internal/repository`.
- Implemented deterministic SQLite token-tree reads over persisted directory and file size metadata with scoped depth and node limits.
- Added an app-layer token-tree service plus tests covering path scoping, truncation behavior, persisted reads, and deterministic repeated results.

## Task Commits

Each task was committed atomically:

1. **Task 1: Define token-tree request and result contracts** - `aaaa9b8` (feat)
2. **Task 2: Add persisted token-tree reads over directory and file size metadata** - `251077b` (feat)
3. **Task 3: Implement the app-layer token-tree service** - `50193b2` (feat)

## Files Created/Modified
- `internal/repository/token_tree.go` - Transport-neutral token-tree request, result, bounds, summary, and node models.
- `internal/store/sqlite/token_tree.go` - Persisted scope resolution, subtree counting, deterministic child ordering, and bounded tree assembly.
- `internal/store/sqlite/token_tree_test.go` - Store coverage for full tree shaping and explicit depth/node truncation metadata.
- `internal/app/token_tree.go` - TokenTreeService that resolves repository state and reuses the shared bytes-to-token policy.
- `internal/app/token_tree_test.go` - Service-level coverage for path scoping, truncation, persisted reads, and repeated-read determinism.

## Decisions Made
- Reused `bytes_div_4_ceiling` from Phase 4 budget analysis so token-tree estimates stay consistent with the existing v1 cost model.
- Represented the result as a rooted hierarchy with explicit summary bounds and per-node child truncation metadata so later MCP handlers do not need extra semantics.
- Ordered directory children before file children, then sorted deterministically by size and path to keep repeated reads stable.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Corrected depth-limited summary counting**
- **Found during:** Task 2 (Add persisted token-tree reads over directory and file size metadata)
- **Issue:** The initial depth-limited directory count filtered valid descendants when the scope was nested, under-reporting `DepthLimitedNodeCount`.
- **Fix:** Removed the incorrect scope-depth subtraction guard and kept depth filtering relative to the scoped root only.
- **Files modified:** `internal/store/sqlite/token_tree.go`
- **Verification:** `GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/home/nico/go/pkg/mod GOPROXY=off go test ./... -run 'TestTokenTree|TestTokenTreeBounds'`
- **Committed in:** `251077b` (part of task commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Required for correct bounds reporting. No scope creep.

## Issues Encountered
- Full `go test ./...` is currently blocked by unrelated in-progress MCP changes already present in the worktree. `internal/mcp/server_test.go` fails because `TestMCPServerStdioSession` now sees 8 tools instead of 2. This was documented in `.planning/phases/05-mcp-serving-and-integration-contracts/deferred-items.md`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Token-tree semantics are now available for later MCP tool wrapping without adding new estimation or truncation rules.
- The full-suite MCP failure should be resolved in the separate in-progress MCP tool work before relying on a clean `go test ./...` signal.

## Self-Check
PASSED

---
*Phase: 05-mcp-serving-and-integration-contracts*
*Completed: 2026-03-15*
