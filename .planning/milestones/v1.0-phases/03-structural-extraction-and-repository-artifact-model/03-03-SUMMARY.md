---
phase: 03-structural-extraction-and-repository-artifact-model
plan: "03"
subsystem: database
tags: [sqlite, refresh, tree-sitter, extraction, testing]
requires:
  - phase: 03-01
    provides: structural artifact tables and store contracts
  - phase: 03-02
    provides: deterministic extraction engine and Go adapter
provides:
  - refresh-scoped structural artifact replacement and deletion
  - canonical refresh-driven extraction candidate orchestration
  - end-to-end mutation coverage for edit, move, ignore, and syntax-break flows
affects: [repository-map, refresh, diagnostics]
tech-stack:
  added: []
  patterns: [refresh transaction callback for structural persistence, per-file artifact replacement by path and generation]
key-files:
  created: [internal/store/sqlite/extraction.go, internal/store/sqlite/extraction_test.go, internal/app/refresh_extraction_test.go]
  modified: [internal/store/sqlite/refresh.go, internal/app/refresh.go, internal/app/refresh_test.go, internal/extract/types.go]
key-decisions:
  - "Structural artifact writes now run inside ApplyRefreshPlan through a SQLite callback instead of a second post-refresh transaction."
  - "Refresh derives extraction work strictly from diff-affected included paths and leaves unchanged artifact rows untouched on no-op runs."
  - "Files with no persisted language hint normalize to `unknown` when persisted as unsupported artifacts so coverage remains explicit."
patterns-established:
  - "Canonical refresh owns structural extraction: snapshot reconciliation, candidate queueing, artifact replacement, and artifact deletion advance together."
  - "Per-file extraction failures degrade to failed coverage rows and do not poison unrelated file artifacts."
requirements-completed: [EXTR-02, EXTR-03, EXTR-04]
duration: 12min
completed: 2026-03-15
---

# Phase 03 Plan 03: Structural Refresh Integration Summary

**Refresh now persists deterministic structural artifacts inline, replacing or deleting per-file symbols only for changed repository paths while keeping degraded coverage truthful.**

## Performance

- **Duration:** 12 min
- **Started:** 2026-03-15T00:07:23Z
- **Completed:** 2026-03-15T00:19:15Z
- **Tasks:** 3
- **Files modified:** 7

## Accomplishments
- Added transaction-scoped SQLite helpers to replace, query, and delete structural artifacts by affected refresh paths.
- Extended the canonical refresh service to queue supported and unsupported extraction candidates from diff results and persist them inside the refresh transaction.
- Added temp-repository progression tests covering no-op stability, edits, moves, ignore and re-include transitions, and syntax-break recovery.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add transactional structural-artifact persistence helpers** - `a066f88` (feat)
2. **Task 2: Queue extraction candidates from refresh and persist results** - `0b7c8a6` (feat)
3. **Task 3: Add end-to-end progression tests for edit, move, ignore, and syntax-break flows** - `3f65bc2` (test)

## Files Created/Modified
- `internal/store/sqlite/extraction.go` - refresh transaction helpers for candidate lookup plus path-based artifact replacement and deletion
- `internal/store/sqlite/extraction_test.go` - focused store coverage for delete transitions and failed-file isolation
- `internal/store/sqlite/refresh.go` - refresh transaction callback hook for structural persistence
- `internal/app/refresh.go` - extraction candidate planning, inline persistence, and extraction result accounting
- `internal/app/refresh_extraction_test.go` - no-op stability and candidate queue coverage
- `internal/app/refresh_test.go` - mutation progression coverage across edit, move, ignore, and syntax-break flows
- `internal/extract/types.go` - stable `unknown` fallback for unsupported files lacking a language hint

## Decisions Made
- Structural persistence runs inside `ApplyRefreshPlan` so file inventory generation and structural artifacts cannot diverge on success.
- Unsupported affected files still get explicit `file_extractions` rows, even when discovery leaves the language hint empty.
- Refresh reports extraction counters but preserves the existing refresh result fields and semantics.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Normalize empty language hints for unsupported artifacts**
- **Found during:** Task 2 (Queue extraction candidates from refresh and persist results)
- **Issue:** Included files such as `.gitignore` or `.txt` can have an empty persisted language hint, which blocked explicit unsupported-row persistence.
- **Fix:** Normalized empty extraction languages to `unknown` at artifact build time so refresh can persist truthful unsupported coverage without inventing parser work.
- **Files modified:** `internal/extract/types.go`
- **Verification:** `go test ./... -run 'TestRefreshService|TestNoOpRefreshKeepsArtifactsStable|TestRefreshQueuesExtractionCandidates|TestExtractionPersistence|TestMoveReplacesArtifacts|TestIgnoreTransitionRemovesArtifacts|TestSyntaxBreakExtractionRecovery'`
- **Committed in:** `0b7c8a6`

---

**Total deviations:** 1 auto-fixed (1 missing critical)
**Impact on plan:** Required for correctness of unsupported-file persistence. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Refresh, coverage-state persistence, and mutation truthfulness are in place for repository-map work in `03-04`.
- No blockers identified for persisted-only repository map generation.

## Self-Check: PASSED
- Verified `.planning/phases/03-structural-extraction-and-repository-artifact-model/03-03-SUMMARY.md` exists on disk.
- Verified task commits `a066f88`, `0b7c8a6`, and `3f65bc2` exist in git history.
