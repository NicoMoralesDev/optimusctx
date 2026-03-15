---
phase: 06-watch-mode-pack-export-and-operational-diagnostics
plan: "05"
subsystem: api
tags: [go, cli, diagnostics, sqlite, watch, mcp]
requires:
  - phase: 06-watch-mode-pack-export-and-operational-diagnostics
    provides: "health reads, watch status contracts, and budget analysis seams"
provides:
  - "Typed doctor report sections for install, state, refresh, watch, structural coverage, budget, and MCP readiness"
  - "Read-only `optimusctx doctor` CLI rendering with actionable healthy, degraded, and missing-state output"
  - "Healthy and degraded doctor coverage across app aggregation and CLI formatting"
affects: [phase-06-operations, cli, diagnostics, watch, budget, mcp]
tech-stack:
  added: []
  patterns: [typed diagnostic aggregation, read-only operator reporting, actionable CLI issue synthesis]
key-files:
  created: [internal/repository/doctor.go, internal/app/doctor.go, internal/app/doctor_test.go, internal/cli/doctor.go, internal/cli/doctor_test.go]
  modified: [internal/cli/root.go]
key-decisions:
  - "Doctor reuses existing health, watch, and budget seams and adds only minimal read-only SQL for latest refresh-run and structural coverage details."
  - "Operator output reports section-specific root causes and next actions instead of exposing raw database jargon."
  - "MCP readiness is derived from the existing snippet and serve-command contract so manual and automated guidance stay aligned."
patterns-established:
  - "Diagnostic services should aggregate existing transport-neutral app seams before introducing new store contracts."
  - "Operator-facing CLI diagnostics should always pair a degraded or missing status with a concrete next action."
requirements-completed: [CLI-05, OPS-05, OPS-01]
duration: 9min
completed: 2026-03-15
---

# Phase 6 Plan 05: Doctor Command Summary

**Repository-wide doctor diagnostics with typed status sections, actionable CLI rendering, and coverage for healthy, stale, partial, and failed operating states**

## Performance

- **Duration:** 9 min
- **Started:** 2026-03-15T19:05:03Z
- **Completed:** 2026-03-15T19:13:52Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments
- Added a typed doctor report model and aggregation service that reuses health, watch, budget, and repository freshness signals.
- Added `optimusctx doctor` as a read-only operator command that renders install, repository, state, refresh, watch, parsing, budget, and MCP readiness clearly.
- Added healthy and degraded coverage for stale watch state, failed refresh history, coverage gaps, missing state, and token-cost hotspots.

## Task Commits

Each task was committed atomically:

1. **Task 1: Define the typed doctor report and aggregation seams** - `a0cfa8e` (feat)
2. **Task 2: Add the `optimusctx doctor` CLI and actionable rendering** - `08404c3` (feat)
3. **Task 3: Add coverage for healthy and degraded diagnostics, including watch and budget signals** - `9b95ae8` (test)

**Plan metadata:** pending

## Files Created/Modified
- `internal/repository/doctor.go` - Typed doctor report sections and status models.
- `internal/app/doctor.go` - Read-only aggregation service for health, refresh, watch, structural coverage, budget, and MCP readiness.
- `internal/cli/doctor.go` - Doctor command wiring and actionable report formatting.
- `internal/cli/root.go` - Root-command dispatch and help entry for `doctor`.
- `internal/app/doctor_test.go` - App-level healthy and degraded doctor integration coverage.
- `internal/cli/doctor_test.go` - CLI rendering coverage for healthy and missing or degraded output.

## Decisions Made
- Doctor reads latest refresh-run failure details and structural coverage examples directly from the existing SQLite state in read-only mode instead of opening the mutating store bootstrap path.
- Structural coverage diagnostics treat partial and failed extraction states as degraded while unsupported and skipped files remain visible as examples for operators.
- MCP readiness reuses the existing snippet generator and serve-command contract so `doctor`, `snippet`, and `install` cannot drift.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Corrected structural coverage summary query to use the existing files and file_extractions schema**
- **Found during:** Task 1 (Define the typed doctor report and aggregation seams)
- **Issue:** The first implementation queried a non-existent derived table instead of the current persisted schema.
- **Fix:** Replaced that query with the same read-only files and file_extractions aggregation pattern the store already uses.
- **Files modified:** `internal/app/doctor.go`
- **Verification:** `go test ./... -run 'TestDoctorReportSections|TestDoctorDetectsDegradedState'`
- **Committed in:** `a0cfa8e` (part of task commit)

**2. [Rule 1 - Bug] Normalized ambiguous and nullable structural example columns**
- **Found during:** Task 1 (Define the typed doctor report and aggregation seams)
- **Issue:** Joined structural-example reads failed on ambiguous column names and NULL language values for unsupported files.
- **Fix:** Qualified the joined SQL columns explicitly and scanned language through `sql.NullString`.
- **Files modified:** `internal/app/doctor.go`
- **Verification:** `go test ./... -run 'TestDoctorReportSections|TestDoctorDetectsDegradedState'`
- **Committed in:** `a0cfa8e` (part of task commit)

---

**Total deviations:** 2 auto-fixed (2 bug fixes)
**Impact on plan:** Both fixes were required for correct read-only diagnostics. No scope creep.

## Issues Encountered
- The previously referenced `/tmp/optimusctx-go` toolchain path no longer existed in this session, so verification used the current `/usr/local/go/bin/go` and `/usr/local/go/bin/gofmt` binaries with the same isolated `/tmp` build caches.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 6 now has a single operator-facing diagnostics entrypoint that surfaces install, state, refresh, watch, extraction, budget, and MCP readiness in one command.
- No blockers remain for milestone wrap-up; the doctor path is covered by both targeted and full-suite verification.

## Self-Check: PASSED
- Verified `.planning/phases/06-watch-mode-pack-export-and-operational-diagnostics/06-05-SUMMARY.md` exists.
- Verified task commits `a0cfa8e`, `08404c3`, and `9b95ae8` exist in git history.
