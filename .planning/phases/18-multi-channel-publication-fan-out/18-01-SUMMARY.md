---
phase: 18-multi-channel-publication-fan-out
plan: "01"
subsystem: release
tags: [go, release, github-release, npm, homebrew, scoop]
requires:
  - phase: 17-canonical-release-orchestration-and-metadata
    provides: canonical release tag, release URL, checksum manifest, and selected channel handoff
provides:
  - shared downstream publication planning rooted in ReleaseOrchestrationPlan
  - rerun-safe per-channel publication metadata for npm, Homebrew, and Scoop
  - deterministic tests for downstream selection and single-channel reruns
affects: [phase-18-workflow-fanout, phase-19-operator-verification, publication-retries]
tech-stack:
  added: []
  patterns: [canonical release fanout, per-channel publication contract, deterministic release planning tests]
key-files:
  created: [internal/release/publication.go, internal/release/publication_test.go]
  modified: [internal/release/orchestration.go, internal/release/orchestration_test.go]
key-decisions:
  - "Downstream publication planning derives only from ReleaseOrchestrationPlan so npm, Homebrew, and Scoop inherit one canonical tag, release URL, and checksum manifest URL."
  - "GitHub Release archives remain the canonical root and are rejected as a downstream publication target for reruns."
patterns-established:
  - "Publication fanout follows canonical release orchestration rather than rebuilding release facts per channel."
  - "Rerun mode narrows to one exact selected downstream channel while keeping retry-safe metadata anchored to the existing tag."
requirements-completed: [PUB-02, PUB-03]
duration: 2min
completed: 2026-03-18
---

# Phase 18 Plan 01: Multi-Channel Publication Fan-Out Summary

**Shared downstream publication planning for npm, Homebrew, and Scoop from one canonical GitHub Release tag with exact per-channel rerun filtering**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-18T11:25:30Z
- **Completed:** 2026-03-18T11:27:43Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Added `ReleasePublicationPlan` and per-channel publication metadata rooted in the canonical release orchestration contract.
- Carried exact canonical tag, release URL, checksum manifest URL, credential env vars, and render commands into downstream publication units.
- Locked fanout membership, rerun filtering, and retry-safe metadata with deterministic unit tests.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add a shared downstream publication plan rooted in canonical release orchestration** - `f250a45` (feat)
2. **Task 2: Lock per-channel filtering and retry-safe metadata with deterministic tests** - `839df5b` (test)

## Files Created/Modified

- `internal/release/publication.go` - Shared downstream publication plan, channel metadata, and fanout/rerun planner.
- `internal/release/publication_test.go` - Deterministic tests for downstream channel membership, rerun narrowing, and rejection cases.
- `internal/release/orchestration.go` - Added orchestration helpers that expose selected and downstream channel handoff data.
- `internal/release/orchestration_test.go` - Added a reusable fixture for orchestration plans that select all publication channels.

## Decisions Made

- Downstream publication planning consumes `ReleaseOrchestrationPlan` directly instead of recalculating release facts from raw strings.
- `github-release-archive` is excluded from downstream publication planning and rejected explicitly for rerun requests.
- Every downstream publication unit is marked retry-safe against an existing canonical tag so later workflow fanout can retry one failed channel independently.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The initial fanout test assumed a specific channel order. The contract only requires exact downstream membership, so the test was corrected to compare deterministic set equality.
- `gsd-tools state advance-plan` stayed on the Phase 17 end marker, so `STATE.md` and `ROADMAP.md` were corrected manually after the automated updates ran.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 18 can now wire workflow fanout and channel-specific transport against one shared publication contract.
- Phase 19 can consume the per-channel metadata and retry-safe semantics for operator status and recovery guidance.

## Self-Check: PASSED
