---
phase: 19-operator-verification-recovery-and-end-to-end-guide
plan: "02"
subsystem: docs
tags: [release, docs, operator-guidance, testing]
requires:
  - phase: 19-01
    provides: workflow summary status and safe rerun guidance for GitHub Release, npm, Homebrew, and Scoop
provides:
  - canonical operator release guide covering prepare, publish, verify, rerun, and rollback from the GitHub Release root
  - checklist and install docs that point operators to the same guide instead of restating divergent release flows
  - regression tests that lock rerun and rollback wording across release and install docs
affects: [phase-19, operator-docs, release-operations]
tech-stack:
  added: []
  patterns: [one canonical operator guide for release flow, docs locked to guide with exact-string contract tests]
key-files:
  created:
    - docs/operator-release-guide.md
    - .planning/phases/19-operator-verification-recovery-and-end-to-end-guide/19-02-SUMMARY.md
  modified:
    - docs/release-checklist.md
    - docs/install-and-verify.md
    - internal/release/release_test.go
    - internal/cli/install_test.go
key-decisions:
  - "The operator guide is the single canonical release flow, and supporting docs link to it instead of duplicating recovery instructions."
  - "Rerun guidance stays rooted in workflow_dispatch with release_tag and publication_channel, while rollback remains anchored to prior GitHub Release archives."
patterns-established:
  - "Operator docs start from the canonical GitHub Release root before any package-manager verification."
  - "Release-facing docs and tests share exact rerun and rollback wording so operator guidance cannot drift silently."
requirements-completed: [OPS-07, OPS-08]
duration: 2m
completed: 2026-03-18
---

# Phase 19 Plan 02: Operator Verification Recovery And End-To-End Guide Summary

**One canonical operator guide now covers release preparation, publish verification, targeted reruns, and rollback, with checklist and install docs routed to the same GitHub-Release-rooted flow**

## Performance

- **Duration:** 2m
- **Started:** 2026-03-18T18:08:46Z
- **Completed:** 2026-03-18T18:10:15Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Added `docs/operator-release-guide.md` as the canonical operator workflow from `optimusctx release prepare` through GitHub Release verification, downstream channel checks, rerun, and rollback.
- Routed `docs/release-checklist.md` and `docs/install-and-verify.md` to the operator guide while preserving the exact `workflow_dispatch`, `release_tag`, and `publication_channel` rerun contract.
- Added release and CLI tests that lock the guide linkage, canonical-root wording, and shipped verification command references.

## Task Commits

Each task was committed atomically:

1. **Task 1: Write the canonical end-to-end operator release guide** - `42fd461` (`feat`)
2. **Task 2: Point checklist and install docs at the canonical guide and lock the wording with tests** - `274c07a` (`fix`)

## Files Created/Modified

- `docs/operator-release-guide.md` - Added the canonical operator release flow across prepare, publish, verify, rerun, and rollback.
- `docs/release-checklist.md` - Pointed release operators at the canonical guide and kept exact rerun input names in checklist form.
- `docs/install-and-verify.md` - Added a release-operator pointer to the canonical guide while preserving shipped verification commands and rerun wording.
- `internal/release/release_test.go` - Added contract coverage for canonical operator guide wording and operator-doc linkage.
- `internal/cli/install_test.go` - Added install-guide linkage coverage for the operator flow and canonical rollback wording.
- `.planning/phases/19-operator-verification-recovery-and-end-to-end-guide/19-02-SUMMARY.md` - Recorded execution details and verification evidence for this plan.

## Decisions Made

- The new operator guide owns the end-to-end release story so the checklist and install docs can stay concise and avoid drifting into parallel recovery guides.
- Recovery instructions remain asymmetric: rerun one downstream channel with the existing tag when GitHub Release is healthy, and treat GitHub Release archives as the rollback source.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no new external service configuration required.

## Next Phase Readiness

- The repository now has one canonical operator document that Wave 3 can reference when locking recovery and rollback policy wording.
- The remaining Phase 19 work is narrowed to distribution policy alignment and regression tests for rerun-versus-rollback semantics.

## Self-Check: PASSED

- Found `docs/operator-release-guide.md`
- Found task commit `42fd461`
- Found task commit `274c07a`
