---
phase: 17-canonical-release-orchestration-and-metadata
plan: "01"
subsystem: infra
tags: [release, goreleaser, github-releases, metadata, testing]
requires:
  - phase: 16-release-versioning-and-preflight-guardrails
    provides: canonical semver tags, release preparation inputs, and channel readiness checks
provides:
  - shared canonical GitHub release metadata rooted in one normalized tag
  - deterministic archive and checksum manifest derivation for the shipped platform matrix
  - contract tests that lock CanonicalRelease to the GoReleaser and workflow release contract
affects: [phase-17-plan-02, phase-17-plan-03, release-orchestration, package-managers]
tech-stack:
  added: []
  patterns: [shared release metadata model, config-backed release contract tests]
key-files:
  created:
    - internal/release/canonical_release.go
    - internal/release/canonical_release_test.go
  modified:
    - internal/release/release_test.go
key-decisions:
  - "Canonical release metadata reuses existing archiveName, archiveFormat, and checksumManifestName helpers instead of introducing a second filename contract."
  - "Release contract tests compare CanonicalRelease against .goreleaser.yml and the GitHub release workflow before any downstream consumer rewiring."
patterns-established:
  - "Canonical release root: version, tag, checksum manifest, and archive URLs are derived once in internal/release and consumed downstream."
  - "Release contract coverage: Go tests lock shared helpers to repository config and workflow semantics instead of duplicating release facts in code."
requirements-completed: [PUB-01]
duration: 1m
completed: 2026-03-17
---

# Phase 17 Plan 01: Canonical Release Metadata Summary

**Canonical GitHub release metadata with shared archive, checksum, and asset URLs derived from one normalized tag**

## Performance

- **Duration:** 1m
- **Started:** 2026-03-17T22:35:50Z
- **Completed:** 2026-03-17T22:37:09Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Added `CanonicalRelease`, `CanonicalReleaseAsset`, and `CanonicalChecksumManifest` so one helper owns the release tag, release URL, checksum manifest, and archive asset inventory.
- Added deterministic unit tests for canonical tag normalization, checksum manifest URLs, per-platform assets, Windows zip handling, and invalid version rejection.
- Extended release contract coverage so the shared canonical model is checked against `.goreleaser.yml` and the GitHub release workflow before later plans rewire downstream consumers.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add shared canonical release metadata and asset derivation** - `8db6b3f` (feat)
2. **Task 2: Lock the canonical release model to the existing archive and workflow contract** - `cd4a8b2` (test)

**Plan metadata:** pending final docs commit

## Files Created/Modified

- `internal/release/canonical_release.go` - Shared canonical release model, checksum manifest metadata, asset derivation, and per-platform lookup.
- `internal/release/canonical_release_test.go` - Deterministic coverage for canonical tags, checksum URLs, asset filenames, archive formats, and invalid versions.
- `internal/release/release_test.go` - Contract test that proves `CanonicalRelease` still matches GoReleaser archive rules and GitHub Release workflow semantics.

## Decisions Made

- Reused the established release helper functions for archive names, archive formats, and checksum manifest names so `.goreleaser.yml` remains the only naming contract.
- Kept Task 2 focused on test-backed contract locking rather than adoption work, preserving Phase 17 plan boundaries for later downstream rewiring.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Repaired stale Phase 17 counters in `.planning/STATE.md`**
- **Found during:** Post-task state updates
- **Issue:** `state advance-plan` could not parse `Current Plan` and `Total Plans in Phase` because the state file still said `Not started` and `0` after Phase 17 planning had already produced four plans.
- **Fix:** Updated the stale Phase 17 counters in `STATE.md`, reran the standard `gsd-tools` state update flow, and then normalized the resulting planning metadata to the actual v1.2 milestone state.
- **Files modified:** `.planning/STATE.md`
- **Verification:** `state advance-plan`, `state update-progress`, `state record-metric`, `state add-decision`, `state record-session`
- **Committed in:** pending final docs commit

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Planning metadata was stale, but the fix was limited to state bookkeeping and did not change the implementation scope.

## Issues Encountered

- `state advance-plan` initially failed on stale Phase 17 counters in `.planning/STATE.md`; fixing the counters restored the normal planning update flow.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 17 now has one canonical release metadata surface for fresh-release orchestration and downstream consumer adoption.
- Plan `17-02` can build fresh-versus-existing release orchestration on top of `CanonicalRelease` without re-deriving tag or asset facts.
- Plan `17-03` can rewire npm and package-manager consumers onto the shared model without reopening the release contract.

---
*Phase: 17-canonical-release-orchestration-and-metadata*
*Completed: 2026-03-17*

## Self-Check: PASSED

- Found `.planning/phases/17-canonical-release-orchestration-and-metadata/17-01-SUMMARY.md`
- Found task commit `8db6b3f`
- Found task commit `cd4a8b2`
