---
phase: 03-structural-extraction-and-repository-artifact-model
plan: 01
subsystem: database
tags: [sqlite, migrations, structural-extraction, symbols, repository-map, testing]
requires:
  - phase: 02-incremental-refresh-and-freshness-model
    provides: refresh generations, persisted file metadata, and canonical SQLite migration flow
provides:
  - normalized file extraction rows with explicit coverage states
  - exact symbol persistence with deterministic lexical ordering
  - repository map and coverage read models sourced entirely from SQLite
affects: [phase-03-extraction-engine, phase-03-refresh-integration, repository-map, diagnostics]
tech-stack:
  added: [sqlite schema additions]
  patterns: [per-file artifact replacement, persisted structural coverage truth, transport-neutral store models]
key-files:
  created: [internal/store/migrations/0003_structural_artifacts.sql, .planning/phases/03-structural-extraction-and-repository-artifact-model/03-01-SUMMARY.md]
  modified: [internal/store/migrations/runner_test.go, internal/store/sqlite/store.go, internal/store/sqlite/store_test.go, internal/repository/metadata.go]
key-decisions:
  - "Persist structural coverage in file_extractions while keeping files.language as the routing hint and single file-inventory source of truth."
  - "Replace per-file symbols transactionally inside SQLite so later generations cannot mix stale and current artifacts."
  - "Build repository-map inputs from top-level persisted symbols and explicit coverage states instead of parser-owned blobs."
patterns-established:
  - "Per-file artifact replacement: delete prior symbols, upsert the extraction row, then insert the new symbol set in one transaction."
  - "Persisted degradation truth: unsupported, partial, and failed states stay queryable even when a file has zero symbols."
requirements-completed: [EXTR-01, EXTR-03, EXTR-04]
duration: 8min
completed: 2026-03-14
---

# Phase 3 Plan 1: Structural Artifact Contract Summary

**SQLite-backed structural extraction rows, exact symbol persistence, and repository-map read models using persisted coverage states and lexical symbol order**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-14T23:42:51Z
- **Completed:** 2026-03-14T23:51:07Z
- **Tasks:** 3
- **Files modified:** 5

## Accomplishments
- Added the Phase 3 `file_extractions` and `symbols` schema with coverage-state checks and deterministic query indexes.
- Introduced shared structural metadata types and SQLite store methods for candidates, file artifacts, coverage summaries, and repository-map inputs.
- Proved replacement semantics and degraded persistence states so later generations remove stale symbols instead of appending mixed artifacts.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add Phase 3 structural-artifact tables and indexes** - `ef09ec7` (feat)
2. **Task 2: Introduce shared structural metadata types and store contracts** - `048b63f` (feat)
3. **Task 3: Prove replacement and degradation persistence semantics** - `20d6255` (test)

## Files Created/Modified
- `internal/store/migrations/0003_structural_artifacts.sql` - Adds `file_extractions`, `symbols`, constraints, and deterministic lookup indexes.
- `internal/store/migrations/runner_test.go` - Extends migration coverage for Phase 3 tables, columns, and indexes.
- `internal/repository/metadata.go` - Defines transport-neutral structural artifact, coverage, and repository-map models.
- `internal/store/sqlite/store.go` - Adds candidate listing, artifact replacement, deletion, coverage summary, and repository-map read methods.
- `internal/store/sqlite/store_test.go` - Covers read models, per-file replacement, unsupported files, and failed-generation cleanup.

## Decisions Made
- Reused persisted `files.language` as the routing hint instead of introducing a second language-detection source.
- Kept repository-map inputs limited to persisted files, extraction rows, and top-level symbols so later reads never reparse source files.
- Used `ParentStableKey` in transport models to preserve lexical parent relationships while SQLite assigns symbol IDs.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 3 can now add the extraction engine and Go adapter against one canonical SQLite artifact contract.
- Refresh integration can replace or delete per-file structural rows without inventing ad hoc parser-owned storage.

## Self-Check: PASSED

- Confirmed `.planning/phases/03-structural-extraction-and-repository-artifact-model/03-01-SUMMARY.md` exists.
- Confirmed task commits `ef09ec7`, `048b63f`, and `20d6255` exist in Git history.
