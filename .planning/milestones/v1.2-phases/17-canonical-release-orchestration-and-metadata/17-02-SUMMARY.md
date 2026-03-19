---
phase: 17-canonical-release-orchestration-and-metadata
plan: "02"
subsystem: infra
tags: [release, orchestration, github-releases, metadata, testing]
requires:
  - phase: 17-canonical-release-orchestration-and-metadata
    provides: canonical release metadata rooted in one normalized tag
provides:
  - fresh-versus-reuse orchestration planning for canonical GitHub releases
  - prepare-layer canonical release handoff without re-deriving version or tag facts
  - regression coverage for selected-channel intent across prepare and orchestration boundaries
affects: [phase-17-plan-03, phase-17-plan-04, release-orchestration, release-prepare]
tech-stack:
  added: []
  patterns: [prepare-owned canonical release handoff, mode-based release orchestration planning]
key-files:
  created:
    - internal/release/orchestration.go
    - internal/release/orchestration_test.go
  modified:
    - internal/release/prepare.go
    - internal/release/prepare_test.go
key-decisions:
  - "Release orchestration resolves a CanonicalRelease from ReleasePreparation so Phase 16 remains the source of truth for version, tag, and selected channel intent."
  - "Reuse mode requires an explicit release_tag and rejects any mismatch against the prepared canonical tag before downstream publication starts."
patterns-established:
  - "Mode-explicit orchestration: create and reuse flows share one plan type with booleans that make GitHub Release creation versus reuse explicit."
  - "Prepare boundary ownership: orchestration consumes prepared canonical release data and selected channels instead of reconstructing release facts from raw strings."
requirements-completed: [PUB-01]
duration: 7m
completed: 2026-03-17
---

# Phase 17 Plan 02: Release Orchestration Summary

**Canonical release orchestration plan types for fresh creation and tagged-release reuse built directly on prepared version, tag, and selected channel data**

## Performance

- **Duration:** 7m
- **Started:** 2026-03-17T22:41:31Z
- **Completed:** 2026-03-17T22:48:49Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Added one `PlanReleaseOrchestration` contract that distinguishes fresh canonical release creation from reusing an existing tagged release while targeting the same `CanonicalRelease` facts.
- Added deterministic orchestration tests for create mode, reuse mode, invalid mode rejection, and explicit reuse-tag mismatch failures.
- Extended `ReleasePreparation` with a canonical-release handoff helper and added regressions proving selected-channel intent survives the prepare-to-orchestration boundary.

## Task Commits

Each task was committed atomically:

1. **Task 1: Define fresh-versus-reuse orchestration types and planning service** - `4930a75` (feat)
2. **Task 2: Wire orchestration to the prepare contract and lock rerun semantics** - `ca2764b` (fix)

**Plan metadata:** pending final docs commit

## Files Created/Modified

- `internal/release/orchestration.go` - Shared orchestration mode, request, and plan types plus the canonical create-versus-reuse planner.
- `internal/release/orchestration_test.go` - Coverage for create and reuse planning, invalid mode rejection, and canonical tag mismatch handling.
- `internal/release/prepare.go` - `ReleasePreparation.CanonicalRelease()` helper so orchestration consumes the prepared canonical version and tag directly.
- `internal/release/prepare_test.go` - Selected-channel and handoff regressions proving prepare output stays authoritative when orchestration planning begins.

## Decisions Made

- Kept create and reuse in one orchestration plan model so later publication phases can switch behavior without changing the canonical release metadata contract.
- Moved canonical release resolution onto `ReleasePreparation` instead of letting orchestration call low-level constructors directly, preserving Phase 16 as the upstream boundary.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Normalized stale planning metadata after automated state updates**
- **Found during:** Final planning artifact updates
- **Issue:** The standard `gsd-tools` state and roadmap update flow advanced counters, but left stale milestone text in `STATE.md` and still pointed `ROADMAP.md` at executing plan `17-02`.
- **Fix:** Manually updated the stale milestone, progress text, completed plan checkbox, and next-step guidance to reflect completion of `17-02`.
- **Files modified:** `.planning/STATE.md`, `.planning/ROADMAP.md`
- **Verification:** Re-read both planning files after patching and confirmed they now point to plan `17-03`.
- **Committed in:** pending final docs commit

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** The fix was limited to planning metadata consistency; implementation scope and task outputs were unchanged.

## Issues Encountered

- `git commit` required elevated filesystem permission to write `.git/index.lock`; after approval, both task commits completed normally.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 17 now has one orchestration boundary that downstream publication consumers can call whether they are creating a new GitHub Release or reusing an existing tag.
- Plan `17-03` can rewire npm and package-manager consumers onto the orchestration-plus-canonical-release contract without reopening Phase 16 selection logic.

---
*Phase: 17-canonical-release-orchestration-and-metadata*
*Completed: 2026-03-17*

## Self-Check: PASSED

- Found `.planning/phases/17-canonical-release-orchestration-and-metadata/17-02-SUMMARY.md`
- Found task commit `4930a75`
- Found task commit `ca2764b`
