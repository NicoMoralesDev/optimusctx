---
phase: 06-watch-mode-pack-export-and-operational-diagnostics
plan: 01
subsystem: infra
tags: [watch, cli, heartbeat, status, diagnostics, polling]
requires:
  - phase: 02-incremental-refresh-and-freshness-model
    provides: canonical refresh pipeline and freshness truth reused by watch mode
  - phase: 01-bootstrap-repository-discovery-and-persistent-state
    provides: repository-local .optimusctx layout with tmp and logs directories
provides:
  - optional `optimusctx watch run` and `optimusctx watch status` command surface
  - transport-neutral watch runner with debounced refresh triggering
  - repo-scoped ephemeral watch status contract under `.optimusctx/tmp/`
affects: [phase-06-plan-02, phase-06-plan-05, watch-mode, diagnostics]
tech-stack:
  added: []
  patterns: [thin-cli-to-app-service, repo-local-ephemeral-status-json, polling-watch-observer]
key-files:
  created: [internal/cli/watch.go, internal/app/watch.go, internal/repository/watch.go, internal/cli/watch_test.go, internal/app/watch_test.go]
  modified: [internal/cli/root.go]
key-decisions:
  - "Watch liveness stays in a repo-local JSON file under `.optimusctx/tmp/` while refresh freshness remains in SQLite."
  - "The initial watch runtime uses a transport-neutral polling observer seam and debounced refresh calls rather than adding daemon management."
patterns-established:
  - "Optional operator tooling follows the existing thin CLI pattern and delegates lifecycle behavior to `internal/app`."
  - "Graceful watch shutdown removes the status file, while crashed or stalled processes degrade to stale via heartbeat age."
requirements-completed: [OPS-01]
duration: 22min
completed: 2026-03-15
---

# Phase 06 Plan 01: Watch Command Foundation Summary

**Optional watch-mode CLI with repo-local heartbeat status, debounced refresh triggering, and stale-versus-absent liveness reporting**

## Performance

- **Duration:** 22 min
- **Started:** 2026-03-15T18:00:00Z
- **Completed:** 2026-03-15T18:22:04Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments
- Added `optimusctx watch run` and `optimusctx watch status` as real root-command entrypoints with explicit unsupported-argument handling.
- Implemented a transport-neutral watch runner that writes heartbeat status under `.optimusctx/tmp/watch-status.json`, debounces event-triggered refreshes, and classifies stale versus absent status.
- Added deterministic tests covering CLI invocation, runtime lifecycle, graceful cleanup, and stale heartbeat detection.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add the Phase 6 `watch` CLI surface** - `f468812` (feat)
2. **Task 2: Implement repo-scoped watch runtime and ephemeral status tracking** - `3cd756c` (feat)
3. **Task 3: Lock the watch CLI and status contract into tests** - `65f6edb` (test)

**Plan metadata:** pending until state updates and final docs commit

## Files Created/Modified

- `internal/cli/root.go` - registers `watch` in the root command help and dispatch table
- `internal/cli/watch.go` - parses `watch run` and `watch status`, delegates execution, and renders status output
- `internal/app/watch.go` - owns watch lifecycle, status-file persistence, polling observation, and refresh triggering
- `internal/repository/watch.go` - defines watch request, event, heartbeat, and status contracts shared across layers
- `internal/cli/watch_test.go` - covers root exposure, subcommand behavior, rendering, and invalid invocation paths
- `internal/app/watch_test.go` - covers lifecycle status updates, graceful cleanup, and stale heartbeat classification

## Decisions Made

- Watch process liveness is ephemeral operator state in `.optimusctx/tmp/`, not a new persistent freshness source of truth.
- `watch run` stays foreground-oriented and signal-driven; shell backgrounding or later supervisors can reuse the same app-level runner.
- The default runtime uses a polling observer seam so the CLI and status contract exist before any platform-specific watcher backend is introduced.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Initial compile errors in `internal/app/watch.go` came from mismatched repository-root and time-helper usage; these were corrected before any task commit and covered by the targeted watch suite.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 6 plan 02 can reuse the established watch runner and status contract to couple watch events into the canonical refresh pipeline more deeply if needed.
- Phase 6 plan 05 can consume the repo-local watch status file alongside health diagnostics without adding new storage.

## Self-Check

PASSED

- Verified created and modified watch files exist on disk.
- Verified task commits `f468812`, `3cd756c`, and `65f6edb` exist in git history.

---
*Phase: 06-watch-mode-pack-export-and-operational-diagnostics*
*Completed: 2026-03-15*
