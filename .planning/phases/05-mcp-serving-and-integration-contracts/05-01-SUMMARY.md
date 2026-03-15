---
phase: 05-mcp-serving-and-integration-contracts
plan: "01"
subsystem: api
tags: [mcp, stdio, json-rpc, cli, testing]
requires:
  - phase: 04-layered-context-exact-lookup-and-budget-analysis
    provides: persisted repository query and refresh services that remain transport-neutral under internal/app
provides:
  - stable `optimusctx mcp serve` CLI entrypoint
  - dedicated `internal/mcp` stdio server foundation with shared protocol types
  - deterministic tool-registry and structured transport-error coverage
affects: [phase-05-read-tools, phase-05-pack-health, phase-05-registration]
tech-stack:
  added: []
  patterns: [thin CLI transport bootstrap, framed json-rpc stdio session handling, structured MCP tool error envelopes]
key-files:
  created:
    - internal/cli/mcp.go
    - internal/cli/mcp_test.go
    - internal/mcp/protocol.go
    - internal/mcp/server.go
    - internal/mcp/server_test.go
  modified:
    - internal/cli/root.go
key-decisions:
  - "The `mcp serve` command stays a thin CLI shim and delegates STDIO session lifecycle to `internal/mcp`."
  - "The Phase 5 MCP transport uses header-framed JSON-RPC responses with shared initialize, tools/list, and tools/call payload primitives."
  - "Tool discovery is deterministic and unavailable or unimplemented tool slots fail with structured error payloads instead of silent success."
patterns-established:
  - "CLI-to-transport delegation: nested CLI commands may wire transport seams but should not embed protocol handlers."
  - "MCP registry pattern: register tool definitions in `internal/mcp`, keep ordering stable, and surface failure metadata through `ResponseError`."
requirements-completed: [MCP-01]
duration: 3min
completed: 2026-03-15
---

# Phase 5 Plan 01: MCP Server Foundation Summary

**STDIO MCP serving with a real `optimusctx mcp serve` command, shared JSON-RPC protocol envelopes, and deterministic tool-registry failure handling**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-15T14:50:08Z
- **Completed:** 2026-03-15T14:53:27Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments

- Added a real nested `optimusctx mcp serve` entrypoint with explicit subcommand and flag rejection behavior.
- Introduced `internal/mcp` as the dedicated STDIO transport layer with shared request, response, tool, and error primitives.
- Added deterministic server-loop coverage for initialize, tools/list, unknown tool failures, and registered-but-unimplemented tool failures.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add the CLI `mcp serve` entrypoint** - `c157cf8` (feat)
2. **Task 2: Implement the stdio MCP server foundation and shared protocol primitives** - `ed0a009` (feat)
3. **Task 3: Add serve-path and stdio-session coverage** - `cb09166` (test)

## Files Created/Modified

- `internal/cli/root.go` - Adds the root-level `mcp` command branch and help text.
- `internal/cli/mcp.go` - Implements nested `mcp` command parsing and delegates serving to `internal/mcp`.
- `internal/cli/mcp_test.go` - Verifies the real root command reaches the serve seam and rejects unsupported arguments.
- `internal/mcp/protocol.go` - Defines shared JSON-RPC request, response, tool, initialize, and error payload types.
- `internal/mcp/server.go` - Implements framed STDIO session handling, initialize/tools/list/tools/call routing, and structured transport errors.
- `internal/mcp/server_test.go` - Covers deterministic tool listing and structured failures for missing or unimplemented tools.

## Decisions Made

- The command contract is now fixed at `optimusctx mcp serve`; later Phase 5 plans can add tools without changing CLI parsing.
- The transport layer uses shared protocol structs rather than ad hoc maps so later tools can reuse one response contract.
- Empty or future tool slots must fail explicitly with structured metadata instead of returning placeholder success results.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed response marshaling after initial serve-command implementation**
- **Found during:** Task 1 (Add the CLI `mcp serve` entrypoint)
- **Issue:** The new server loop passed a `*Response` into the marshal helper, which caused the first targeted verification compile to fail.
- **Fix:** Dereferenced the response before marshaling and reran the targeted serve-command test.
- **Files modified:** `internal/mcp/server.go`
- **Verification:** `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/home/nico/go/pkg/mod GOPROXY=off go test ./... -run 'TestMCPServeCommand'`
- **Committed in:** `c157cf8`

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** The fix was required for correctness and did not expand scope beyond the planned MCP foundation.

## Issues Encountered

- The recorded `/tmp/optimusctx-go/go/bin/gofmt` toolchain path from earlier summaries was not present in this environment, so formatting and verification used the local Go toolchain under `/usr/local/go/bin`.
- Default `go test` attempted network fetches despite local module cache availability; verification succeeded offline with `GOMODCACHE=/home/nico/go/pkg/mod` and `GOPROXY=off`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The binary now exposes a stable MCP STDIO entrypoint and a dedicated transport package for tool registration.
- Later Phase 5 plans can add read-only tools, pack/health services, and registration workflows without changing the `mcp serve` command contract.

## Self-Check: PASSED

- Verified summary file exists on disk.
- Verified task commits `c157cf8`, `ed0a009`, and `cb09166` exist in `git log --oneline --all`.

---
*Phase: 05-mcp-serving-and-integration-contracts*
*Completed: 2026-03-15*
