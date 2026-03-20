---
phase: 20-mcp-client-contract-and-config-backend-foundation
plan: "03"
subsystem: api
tags: [go, mcp, claude-desktop, json-config, idempotence]
requires:
  - phase: 20-mcp-client-contract-and-config-backend-foundation
    provides: shared Codex config backend and supported-client adapter refactor from plan 20-02
provides:
  - Claude Desktop path resolution routed through a platform-input helper with explicit override precedence
  - Service-level Claude Desktop preview and write regression coverage for config preservation and idempotence
affects: [phase-21-write-paths, claude-desktop-registration, install-service]
tech-stack:
  added: []
  patterns: [platform-input path resolution helpers, service-level persisted config regression tests]
key-files:
  created: []
  modified: [internal/app/install.go, internal/app/install_test.go]
key-decisions:
  - "Claude Desktop path resolution now delegates runtime values into a platform-input helper so default locations stay explicit and directly testable."
  - "Claude Desktop JSON-file safety is locked through InstallService regression tests instead of lower-level merge-only coverage."
patterns-established:
  - "Path-resolution wrappers should gather environment inputs once and delegate to pure helpers for cross-platform coverage."
  - "Write-backed MCP host adapters need service-level tests that prove unrelated entries survive repeated writes."
requirements-completed: [MCP-04, CLD-01]
duration: 2min
completed: 2026-03-19
---

# Phase 20 Plan 03: Claude Desktop parity and JSON-file safety Summary

**Claude Desktop now resolves default config paths through a pure helper and keeps unrelated `mcpServers` entries intact across preview and repeated writes**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-19T17:48:40-03:00
- **Completed:** 2026-03-19T20:50:36Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Extracted Claude Desktop path resolution into `resolveClaudeDesktopConfigPathForPlatform` while keeping existing default locations and explicit `--config` precedence unchanged.
- Added direct path-resolution tests for override, Linux, macOS, and missing Windows `%AppData%`.
- Added service-level Claude Desktop preview/write regressions that prove unrelated JSON config entries survive and repeated writes remain idempotent.

## Task Commits

Each task was committed atomically:

1. **Task 1: Make Claude Desktop path resolution explicit and testable** - `cc8cbd4` (fix)
2. **Task 2: Lock Claude Desktop preview and write idempotence after the adapter refactor** - `366f41e` (test)

## Files Created/Modified
- `internal/app/install.go` - routes runtime path inputs into a pure Claude Desktop path-resolution helper.
- `internal/app/install_test.go` - adds direct path-resolution coverage plus Claude Desktop preview/write preservation and idempotence regressions.

## Decisions Made
- Claude Desktop kept its existing default config paths on macOS, Linux, and Windows instead of introducing any adapter-specific path changes during the refactor.
- Regression coverage for persisted Claude Desktop behavior runs through `InstallService.Register` so later adapter changes are tested at the real integration boundary.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Corrected the idempotence assertion to count the JSON key instead of the command value**
- **Found during:** Task 2 (Lock Claude Desktop preview and write idempotence after the adapter refactor)
- **Issue:** The first assertion counted `"optimusctx"` anywhere in the file, which also matched the command value and produced a false duplicate failure.
- **Fix:** Narrowed the assertion to count the `"optimusctx":` JSON key and kept the parsed server-count assertion.
- **Files modified:** internal/app/install_test.go
- **Verification:** `go test ./internal/app -run 'Test(InstallServiceClaudeDesktopPreviewUsesResolvedPath|InstallServiceClaudeDesktopWritePreservesExistingServers|InstallServiceClaudeDesktopWriteIsIdempotent)$'`
- **Committed in:** `366f41e` (part of task commit)

---

**Total deviations:** 1 auto-fixed (1 Rule 1 bug)
**Impact on plan:** The fix corrected a test defect in the new regression coverage. No scope creep and no product-behavior change.

## Issues Encountered
- The first idempotence assertion matched both the server key and the `"command": "optimusctx"` value; tightening the assertion resolved the false negative.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 20 now has explicit, testable Claude Desktop path resolution plus persisted JSON safety coverage, closing the Phase 20 requirement set.
- Phase 21 can add real write-backed flows for Claude CLI and Codex clients without destabilizing the already-shipped Claude Desktop integration.

## Self-Check: PASSED
- Found summary file: `.planning/phases/20-mcp-client-contract-and-config-backend-foundation/20-03-SUMMARY.md`
- Found task commit: `cc8cbd4`
- Found task commit: `366f41e`
