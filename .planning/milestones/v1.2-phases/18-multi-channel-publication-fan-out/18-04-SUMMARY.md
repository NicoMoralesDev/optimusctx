---
phase: 18-multi-channel-publication-fan-out
plan: "04"
subsystem: release
tags: [release, docs, github-actions, npm, homebrew, scoop]
requires:
  - phase: 18-02
    provides: package-manager render commands and downstream publication targets
  - phase: 18-03
    provides: canonical workflow fan-out and selective publication_channel reruns
provides:
  - truthful release prepare readiness for npm, Homebrew, and Scoop based on real workflow markers
  - operator docs for canonical fan-out and exact single-channel workflow_dispatch reruns
  - regression coverage tying workflow, prepare output, CLI review, and docs to one publication contract
affects: [phase-19, release-prepare, release-operator-docs, rollback-guidance]
tech-stack:
  added: []
  patterns: [prepare readiness derives from workflow markers plus required templates, operator docs mirror workflow_dispatch rerun contract exactly]
key-files:
  created:
    - .planning/phases/18-multi-channel-publication-fan-out/18-04-SUMMARY.md
  modified:
    - internal/release/prepare.go
    - internal/release/prepare_test.go
    - internal/cli/release_test.go
    - docs/release-checklist.md
    - docs/install-and-verify.md
    - internal/release/release_test.go
key-decisions:
  - "Prepare readiness now treats Homebrew and Scoop as ready only when the checked-in workflow actually contains their publication job names, token wiring, and render commands."
  - "The default release fixtures and docs now reflect the real automated fan-out instead of the old Phase 17 blocked-placeholder story."
  - "Operator rerun guidance is locked to workflow_dispatch with release_tag plus publication_channel so recovery remains single-channel and canonical-release rooted."
patterns-established:
  - "Truthful readiness: prepare checks infer channel status from concrete workflow markers instead of roadmap assumptions."
  - "Canonical release storytelling: docs and tests describe GitHub Release as the root plus rollback source for downstream automation."
requirements-completed: [PUB-02, PUB-03]
duration: 6m
completed: 2026-03-18
---

# Phase 18 Plan 04: Multi-Channel Publication Fan-Out Summary

**Truthful prepare readiness and operator rerun docs for npm, Homebrew, and Scoop over one canonical GitHub Release contract**

## Performance

- **Duration:** 6m
- **Started:** 2026-03-18T11:47:50Z
- **Completed:** 2026-03-18T11:54:14Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Updated release prepare so Homebrew and Scoop are marked ready only when the release workflow actually contains their publication jobs, token wiring, render commands, and template files.
- Added regression coverage for all-channels-ready prepare output and for selected-channel output staying isolated from blocked unselected channels.
- Rewrote operator docs and doc-contract tests so canonical GitHub Release fan-out, rollback, and exact `workflow_dispatch` reruns all describe the same `release_tag` and `publication_channel` contract.

## Task Commits

Each task was committed atomically:

1. **Task 1: Make prepare and CLI readiness reflect real multi-channel automation** - `9600f71` (`fix`)
2. **Task 2: Update operator docs and regression tests for automated fan-out and single-channel reruns** - `607be85` (`docs`)

## Files Created/Modified

- `internal/release/prepare.go` - Homebrew and Scoop readiness now keys off actual publication workflow markers and required templates.
- `internal/release/prepare_test.go` - Added all-channel ready coverage, exact missing-marker assertions, and a degraded fixture for selected-channel isolation.
- `internal/cli/release_test.go` - Locked CLI prepare JSON against both all-channels-ready output and selected-channel review behavior.
- `docs/release-checklist.md` - Updated release operator guidance for automated downstream fan-out and exact reruns with `release_tag` plus `publication_channel`.
- `docs/install-and-verify.md` - Clarified canonical root, rollback source, downstream package-manager fan-out, and single-channel recovery instructions.
- `internal/release/release_test.go` - Added doc-contract assertions for canonical fan-out wording and selective rerun markers in the workflow.

## Decisions Made

- Prepare no longer uses checklist wording to decide Homebrew and Scoop readiness; only workflow markers and required package-manager templates determine channel state.
- The repository’s default release fixture now models the real automated workflow, and the “unselected blockers stay isolated” case uses an explicit degraded workflow fixture.
- Operator docs describe GitHub Release as both the canonical root and rollback source so rerun and recovery guidance stays anchored to one tag.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no new external service configuration was introduced beyond the existing release publication tokens.

## Next Phase Readiness

- Phase 18 now ends with the workflow, prepare layer, CLI review output, and operator docs all describing the same multi-channel publication contract.
- Phase 19 can build on this with operator verification and recovery flows without re-explaining package-manager readiness or rerun semantics.

## Self-Check: PASSED

- Found `.planning/phases/18-multi-channel-publication-fan-out/18-04-SUMMARY.md`
- Found task commits `9600f71` and `607be85`
