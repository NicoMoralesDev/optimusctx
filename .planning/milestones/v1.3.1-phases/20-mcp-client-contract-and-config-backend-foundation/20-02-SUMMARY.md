---
phase: 20-mcp-client-contract-and-config-backend-foundation
plan: "02"
subsystem: infra
tags: [mcp, codex, toml, install-service]
requires:
  - phase: 20-01
    provides: explicit supported-client preview adapters and truthful named-client contract routing
provides:
  - shared Codex `config.toml` parse, merge, and render helpers
  - shared Codex App and Codex CLI preview adapter targeting `~/.codex/config.toml`
  - idempotent preview coverage for Codex config preservation
affects:
  - 21-real-write-paths-and-operator-surface-integration
  - codex-client-registration
tech-stack:
  added: [github.com/pelletier/go-toml/v2]
  patterns:
    - preview-first native host config merges before write support lands
    - shared backend per host family instead of per-client copy-paste adapters
key-files:
  created:
    - internal/repository/codex_config.go
    - internal/repository/codex_config_test.go
  modified:
    - go.mod
    - go.sum
    - internal/app/install.go
    - internal/app/install_test.go
key-decisions:
  - Codex App and Codex CLI now share one `~/.codex/config.toml` preview backend so Phase 21 can add writes without changing the storage model.
  - Codex config merges use TOML parsing plus a deterministic renderer so preview output keeps the documented `[mcp_servers.<name>]` contract with double-quoted string values.
patterns-established:
  - "Shared Codex backend: repository merge/render helpers feed both App and CLI adapters."
  - "Real-home preview tests: install-service Codex coverage uses temporary HOME directories to exercise actual path resolution and merge behavior."
requirements-completed: [MCP-02, MCP-04, CDX-03]
duration: 3min
completed: 2026-03-19
---

# Phase 20 Plan 02: Codex Shared Config Backend Summary

**Shared Codex `config.toml` merge/render helpers with one native preview path for Codex App and Codex CLI**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-19T20:39:38Z
- **Completed:** 2026-03-19T20:42:08Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Added a shared Codex TOML backend that renders and merges `[mcp_servers.optimusctx]` without dropping unrelated top-level or `mcp_servers` content.
- Moved both `codex-app` and `codex-cli` previews onto the same native `~/.codex/config.toml` target and content model.
- Added deterministic tests for first render, merge preservation, repeated-merge idempotence, and install-service preview behavior for both Codex clients.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add a shared Codex `config.toml` render and merge layer** - `759185b` (`feat`)
2. **Task 2: Wire Codex App and CLI onto the shared native preview backend** - `f34aeeb` (`feat`)

## Files Created/Modified

- `go.mod` - adds `github.com/pelletier/go-toml/v2` for Codex config parsing.
- `go.sum` - records the TOML dependency checksum metadata.
- `internal/repository/codex_config.go` - implements Codex TOML render, merge, and deterministic native output helpers.
- `internal/repository/codex_config_test.go` - locks render shape, merge preservation, and repeated-merge idempotence.
- `internal/app/install.go` - adds the shared Codex preview adapter and default `~/.codex/config.toml` path resolution.
- `internal/app/install_test.go` - verifies Codex App/CLI previews and preservation of existing Codex config content.

## Decisions Made

- Codex App and Codex CLI now present distinct client labels while sharing one `config.toml` backend and default path.
- Preview stays non-mutating in Phase 20; write-backed Codex registration remains deferred to Phase 21.
- Codex preview tests now run against temporary home directories so path resolution and merge behavior are exercised through the real adapter boundary.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Replaced default `go-toml/v2` rendering for Codex preview output**
- **Found during:** Task 1 (shared Codex `config.toml` render and merge layer)
- **Issue:** The default encoder emitted literal strings with single quotes, which drifted from the documented native Codex contract shape expected for preview output.
- **Fix:** Kept TOML parsing and merge semantics in `go-toml/v2`, but replaced the final render step with a deterministic renderer that emits the native `[mcp_servers.<name>]` contract using standard double-quoted string values.
- **Files modified:** `internal/repository/codex_config.go`, `internal/repository/codex_config_test.go`
- **Verification:** `go test ./internal/repository -run 'Test(RenderCodexConfig|MergeCodexConfigPreservesExistingContent|MergeCodexConfigIsIdempotent)$'`
- **Committed in:** `759185b`

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** The fix kept the plan scope intact while aligning the rendered Codex preview with the intended native host contract.

## Issues Encountered

- `go.sum` needed a refresh after adding `github.com/pelletier/go-toml/v2`; `go mod tidy` resolved it and the targeted test suite passed afterward.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 21 can add write-backed Codex registration on top of the shared path, merge model, and preview contract without reworking the backend shape.
- Claude CLI write execution remains the next host-specific integration step, but the Codex foundation and merge semantics are now in place.

## Self-Check: PASSED

- Found `.planning/phases/20-mcp-client-contract-and-config-backend-foundation/20-02-SUMMARY.md`
- Found task commit `759185b`
- Found task commit `f34aeeb`

---
*Phase: 20-mcp-client-contract-and-config-backend-foundation*
*Completed: 2026-03-19*
