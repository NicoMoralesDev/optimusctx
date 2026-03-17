---
phase: 06-watch-mode-pack-export-and-operational-diagnostics
plan: "02"
subsystem: runtime
tags: [watch, refresh, cli, sqlite, diagnostics]
requires:
  - phase: 06-01
    provides: repo-scoped watch lifecycle, heartbeat status, and foreground watch CLI entrypoints
provides:
  - watch-triggered refreshes routed through RefreshService with watch reason metadata
  - debounced watch hint coalescing with full-refresh fallback on overflow or uncertainty
  - CLI-visible watch refresh reports and parity tests for degraded recovery
affects: [watch, refresh, doctor, diagnostics]
tech-stack:
  added: []
  patterns: [watch-events-as-hints, canonical-refresh-reuse, cli-refresh-reporting]
key-files:
  created: []
  modified:
    - internal/app/watch.go
    - internal/app/refresh.go
    - internal/repository/watch.go
    - internal/app/watch_test.go
    - internal/app/refresh_test.go
    - internal/cli/watch.go
    - internal/cli/watch_test.go
key-decisions:
  - "Watch events remain advisory: safe relative paths are forwarded as ChangedHint metadata, while overflow or uncertainty forces a normal refresh."
  - "Canonical refresh results, not watcher-local counters, now drive operator-facing watch output."
patterns-established:
  - "Watch-to-refresh orchestration: debounce filesystem noise, then call RefreshService with Reason=watch."
  - "Fallback safety: uncertain watcher state clears path hints and forces a full refresh instead of partial truth."
requirements-completed: [OPS-02]
duration: 10min
completed: 2026-03-15
---

# Phase 6 Plan 02: Watch-to-refresh reuse with debounced hints and safe fallback

**Watch mode now reuses the canonical refresh pipeline with normalized hint metadata, full-refresh fallback on uncertain events, and CLI reporting tied to refresh generations and freshness.**

## Performance

- **Duration:** 10 min
- **Started:** 2026-03-15T18:26:00Z
- **Completed:** 2026-03-15T18:36:17Z
- **Tasks:** 3
- **Files modified:** 7

## Accomplishments
- Routed watch-triggered refreshes through `RefreshService.Refresh(...)` with `Reason: watch`, `ForceFull`, and sanitized `ChangedHint` metadata.
- Coalesced bursty watch events into one refresh window and degraded overflow or uncertain watcher state into a full canonical refresh.
- Surfaced canonical watch refresh outcomes in the CLI and added parity coverage for watch failure, degraded freshness, and recovery.

## Task Commits

Each task was committed atomically:

1. **Task 1: Route watch-triggered refreshes through the canonical refresh service** - `8e902e9` (feat)
2. **Task 2: Add noisy-event and uncertain-state fallback behavior** - `526199f` (test)
3. **Task 3: Prove watch/manual refresh parity in integration-oriented tests** - `215721d` (feat)

## Files Created/Modified
- `internal/app/watch.go` - Debounces event windows, accumulates safe hints, falls back to full refresh, and reports refresh outcomes.
- `internal/app/refresh.go` - Normalizes advisory changed hints and persists watch refresh metadata through the existing refresh transaction.
- `internal/repository/watch.go` - Defines uncertainty and refresh-report contracts shared by app and CLI layers.
- `internal/app/watch_test.go` - Covers canonical refresh reuse, debounce behavior, uncertainty fallback, and degraded recovery.
- `internal/app/refresh_test.go` - Verifies watch refresh metadata persists `reason=watch` plus sanitized hints.
- `internal/cli/watch.go` - Prints canonical watch refresh summaries as the watch loop runs.
- `internal/cli/watch_test.go` - Verifies operator-facing watch output includes refresh outcomes.

## Decisions Made
- Watch-originated hints are normalized and recorded for metadata only; discovery plus diffing remain the correctness boundary.
- Watch CLI output now reports canonical refresh generation and freshness so operators see the same contract manual refresh produces.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- A missing `strings` import in the new hint sanitizer caused one compile failure during verification and was fixed before the first task commit.
- The recovery test initially waited on ephemeral status-file text; it was tightened to wait on canonical SQLite freshness state instead.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Watch mode now satisfies `OPS-02` and exposes stable refresh semantics for doctor diagnostics and future operator surfaces.
- Phase 6 export and doctor plans can rely on watch status for liveness while continuing to treat SQLite freshness as the source of truth.

## Self-Check: PASSED
- Found `.planning/phases/06-watch-mode-pack-export-and-operational-diagnostics/06-02-SUMMARY.md`
- Found commit `8e902e9`
- Found commit `526199f`
- Found commit `215721d`

---
*Phase: 06-watch-mode-pack-export-and-operational-diagnostics*
*Completed: 2026-03-15*
