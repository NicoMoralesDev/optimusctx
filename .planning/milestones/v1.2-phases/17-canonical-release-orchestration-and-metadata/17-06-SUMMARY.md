---
phase: 17-canonical-release-orchestration-and-metadata
plan: "06"
subsystem: infra
tags: [release-orchestration, github-release, go-test, release-preparation]
requires:
  - phase: 16-release-versioning-and-preflight-guardrails
    provides: normalized release tags, selected-channel preparation, and canonical release inputs
  - phase: 17-canonical-release-orchestration-and-metadata
    provides: canonical release metadata and create-versus-reuse orchestration semantics from plans 17-01 through 17-05
provides:
  - Explicit GitHub Release action metadata for create versus reuse orchestration
  - Prepare-owned orchestration handoff with validated canonical release and selected channel plans
  - Regression coverage for reuse tag normalization, selected channel preservation, and invalid handoff rejection
affects: [18-multi-channel-publication-fan-out, 19-operator-verification-recovery-and-end-to-end-guide, release-workflow]
tech-stack:
  added: []
  patterns: [prepare-owned orchestration handoff, explicit GitHub Release action metadata, selected-channel contract tests]
key-files:
  created: []
  modified:
    - internal/release/orchestration.go
    - internal/release/orchestration_test.go
    - internal/release/prepare.go
    - internal/release/prepare_test.go
key-decisions:
  - "ReleasePreparation owns an OrchestrationHandoff object so orchestration consumes already-validated version, tag, canonical release, and selected channel plans instead of rebuilding them piecemeal."
  - "ReleaseOrchestrationPlan now carries explicit GitHub Release action metadata, while the old booleans remain derived compatibility fields rather than the primary contract."
patterns-established:
  - "Release handoff boundaries should validate tag agreement between prepared and canonical release data before orchestration continues."
  - "Selected channel IDs and selected channel plan structs must stay aligned through prepare-to-orchestration handoff tests."
requirements-completed: [PUB-01]
duration: 16min
completed: 2026-03-18
---

# Phase 17 Plan 06: Canonical Release Handoff Summary

**Prepare-owned orchestration handoff with explicit GitHub Release action metadata, normalized reuse tags, and preserved selected channel plans**

## Performance

- **Duration:** 16 min
- **Started:** 2026-03-17T23:44:59Z
- **Completed:** 2026-03-18T00:01:01Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Expanded `ReleaseOrchestrationPlan` to carry explicit GitHub Release action metadata, canonical release URL, requested reuse tag, and selected channel plans instead of relying on thin create/reuse booleans alone.
- Added `ReleasePreparation.OrchestrationHandoff()` so the prepare layer now owns the validated handoff of version, tag, canonical release metadata, selected channel IDs, and selected channel plans.
- Added regressions that prove reuse tag normalization, selected `github-release-archive` plus `npm` channel preservation, and invalid handoff rejection against the shared prepare/orchestration contract.

## Task Commits

Each task was committed atomically:

1. **Task 1: Expand the orchestration plan to carry explicit GitHub Release action metadata** - `8e54faa` (feat)
2. **Task 2: Add a prepare-layer orchestration handoff helper and deepen regressions** - `65d7930` (feat)

## Files Created/Modified

- `internal/release/orchestration.go` - Release orchestration contract, explicit GitHub Release action metadata, selected channel plan carry-through, and handoff-driven planning
- `internal/release/orchestration_test.go` - Create/reuse orchestration regressions for action metadata, normalized reuse tags, selected channel plan preservation, and invalid handoff failures
- `internal/release/prepare.go` - Prepare-owned orchestration handoff type and validation for canonical release plus selected channel agreement
- `internal/release/prepare_test.go` - Handoff regression proving version, tag, canonical release tag, and selected `github-release-archive` and `npm` channels remain unchanged

## Decisions Made

- Kept `PlanReleaseOrchestration(preparation, request)` as the external entrypoint and moved the richer contract behind a prepare-owned handoff so later phases do not need to change callers to gain deeper orchestration metadata.
- Retained `CreateGitHubRelease` and `ReuseExistingRelease` as derived compatibility fields while making `GitHubReleaseAction` the primary contract for explicit create-versus-reuse behavior.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- An unrelated staged index state caused the first Task 1 commit attempt to include extra files; the task commit was corrected by committing later work with explicit path-limited staging and commit scope.
- Existing later Phase 17 artifacts on the branch caused GSD state helpers to point back at stale 17-06 metadata reconciliation text, so `STATE.md` and `ROADMAP.md` were reconciled manually after the summary was written.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 18 can fan out publication from a single handoff object that already preserves canonical release metadata and selected channel intent.
- Recovery and verification work can now branch on explicit GitHub Release action metadata instead of inferring create-versus-reuse behavior from booleans alone.

## Self-Check

PASSED

- Found `.planning/phases/17-canonical-release-orchestration-and-metadata/17-06-SUMMARY.md`
- Found commit `8e54faa`
- Found commit `65d7930`
- Verified targeted release tests and depth thresholds:
  - `go test ./internal/release -run 'Test(PlanReleaseOrchestrationCreate|PlanReleaseOrchestrationReuse|PlanReleaseOrchestrationRejectsInvalidMode|PlanReleaseOrchestrationRejectsTagMismatch|PlanReleaseOrchestrationCarriesSelectedChannelPlans|PlanReleaseOrchestrationNormalizesReuseTag|ReleasePrepareSelectedChannelsReady|ReleasePreparationOrchestrationHandoff|ReleaseSelectedChannelsDoNotInheritUnselectedBlockers)$'`
  - `internal/release/orchestration.go`: 187 lines
  - `internal/release/orchestration_test.go`: 250 lines
  - `internal/release/prepare.go`: 928 lines

---
*Phase: 17-canonical-release-orchestration-and-metadata*
*Completed: 2026-03-18*
