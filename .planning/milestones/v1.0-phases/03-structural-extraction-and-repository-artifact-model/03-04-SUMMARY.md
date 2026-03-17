---
phase: 03-structural-extraction-and-repository-artifact-model
plan: "04"
subsystem: api
tags: [go, sqlite, repository-map, structural-artifacts, deterministic-read-model]
requires:
  - phase: 03-03
    provides: persisted directories, files, file_extractions, and symbols aligned to refresh generations
provides:
  - persisted-only repository map service built from sqlite read models
  - explicit repository map coverage-gap metadata for supported, partial, failed, unsupported, and skipped files
  - persisted-only and deterministic repository map integration coverage after worktree files disappear
affects: [phase-04-query-surface-and-context-assembly, mcp, repository-map]
tech-stack:
  added: []
  patterns: [persisted-only query composition, lexical ordering from sqlite read models, coverage-gap flags for degraded structure]
key-files:
  created: [internal/app/repository_map.go, internal/app/repository_map_test.go]
  modified: [internal/repository/metadata.go, internal/store/sqlite/store.go, internal/store/sqlite/store_test.go, internal/app/refresh_test.go]
key-decisions:
  - "Repository-map reads resolve repository identity from persisted sqlite metadata instead of mutating repository rows during query time."
  - "Only supported and partial files surface top-level symbols; unsupported, failed, and skipped files remain visible with explicit coverage metadata."
  - "Repository-map payloads stay compact by returning directory-grouped files with lexical ordering and reduced symbol fields."
patterns-established:
  - "Persisted-only query services open the repository-local sqlite store, look up repository state, and compose read models without touching worktree files."
  - "Coverage truthfulness is explicit in API payloads through coverage state plus a derived coverage-gap flag."
requirements-completed: [EXTR-04, EXTR-05]
duration: 10min
completed: 2026-03-15
---

# Phase 3 Plan 04: Repository Map Summary

**Deterministic repository-map reads from persisted sqlite directories, files, extraction rows, and top-level symbols without reopening the worktree**

## Performance

- **Duration:** 10 min
- **Started:** 2026-03-15T00:20:09Z
- **Completed:** 2026-03-15T00:30:08Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments
- Added a persisted-only repository-map service that groups lexically ordered files under persisted directories and returns compact top-level symbol payloads.
- Exposed truthful structural coverage for supported, partial, unsupported, failed, and skipped files, including an explicit coverage-gap flag for downstream consumers.
- Proved repository-map reads remain stable from SQLite alone after source files are deleted and across repeated reads from unchanged database state.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add persisted repository-map read models and composition service** - `f4c3914` (feat)
2. **Task 2: Surface truthful coverage states for unsupported and degraded files in repository maps** - `146b5dd` (feat)
3. **Task 3: Prove repository-map reads are persisted-only and deterministic** - `43334df` (test)

**Plan metadata:** pending

## Files Created/Modified
- `internal/app/repository_map.go` - Repository-map service that resolves a repository, opens SQLite, and composes deterministic directory-grouped payloads.
- `internal/app/repository_map_test.go` - Repository-map coverage for ordering, coverage states, persisted-only reads, and deterministic output.
- `internal/repository/metadata.go` - Shared repository-map payload and directory/file metadata types.
- `internal/store/sqlite/store.go` - Repository lookup plus persisted repository-map directory and file read models with skipped-state normalization.
- `internal/store/sqlite/store_test.go` - SQLite repository-map read-model coverage for directory ordering and top-level symbol filtering.
- `internal/app/refresh_test.go` - Integration check that refresh-persisted artifacts remain queryable through repository maps after worktree files are removed.

## Decisions Made
- Repository-map services now perform read-only repository identity lookup through SQLite instead of using `UpsertRepository` on query paths.
- Top-level symbol output stays tied to persisted `ordinal` ordering, and the returned symbol payload is intentionally compact rather than reusing full span-heavy `SymbolRecord` structs.
- Files without persisted extraction rows normalize to `skipped` so repository-map consumers see an explicit structural truth instead of an empty state.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 4 can consume a deterministic persisted repository map without reparsing repositories on demand.
- Coverage-aware query layers can branch on `CoverageState` and `HasCoverageGap` instead of inferring degraded structure from missing symbols.

## Self-Check: PASSED

---
*Phase: 03-structural-extraction-and-repository-artifact-model*
*Completed: 2026-03-15*
