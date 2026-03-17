---
phase: 16-release-versioning-and-preflight-guardrails
plan: 01
subsystem: release
tags: [go, semver, git-tag, release-preparation]
requires:
  - phase: 15-03
    provides: npm wrapper distribution policy and release-derived package contract
provides:
  - canonical MAJOR.MINOR.PATCH release version normalization
  - canonical vMAJOR.MINOR.PATCH release tag normalization and legacy tag alias detection
  - shared release preparation model with policy-derived channel ordering
affects: [16-02, 16-03, 17-canonical-release-orchestration-and-metadata]
tech-stack:
  added: []
  patterns:
    - canonical semver release inputs
    - policy-derived release channel planning
key-files:
  created:
    - internal/release/prepare.go
    - internal/release/prepare_test.go
  modified: []
key-decisions:
  - Release preparation accepts only MAJOR.MINOR.PATCH versions and emits only vMAJOR.MINOR.PATCH tags.
  - Legacy tags such as v1.1 canonicalize to v1.1.0 and block the same semantic release lane.
patterns-established:
  - "Release version helpers live in internal/release so later CLI and preflight work reuse one contract."
  - "Default release channels are assembled from CurrentDistributionPolicy() instead of a second hardcoded list."
requirements-completed: [REL-01]
duration: 2min
completed: 2026-03-17
---

# Phase 16 Plan 01: Canonical Semver and Release-Preparation Contract Summary

**Canonical release semver helpers with legacy tag alias detection and a shared policy-derived release preparation model**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-17T17:04:17-03:00
- **Completed:** 2026-03-17T20:05:58Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Added strict normalization helpers that accept only canonical `MAJOR.MINOR.PATCH` versions and emit only `vMAJOR.MINOR.PATCH` tags.
- Added milestone-based version proposal logic that treats legacy tags like `v1.2` as the same semantic lane as `v1.2.0`.
- Added a shared `ReleasePreparation` model with default channel ordering derived from `CurrentDistributionPolicy()` plus exact and semantic tag-conflict reporting.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add canonical release version and tag normalization helpers** - `1b8ae5d` (`feat`)
2. **Task 2: Define the shared release-preparation model and channel plan** - `a1b4bee` (`feat`)

Plan metadata is captured in the separate docs commit for this summary and planning-state update.

## Files Created/Modified

- `internal/release/prepare.go` - Canonical semver normalization, legacy tag canonicalization, version proposal, and release preparation assembly.
- `internal/release/prepare_test.go` - Coverage for normalized tags, milestone-based version proposal, channel planning, and semantic tag conflicts.

## Decisions Made

- Release preparation now treats `v1.1` and `v1.1.0` as the same semantic release lane so duplicate human releases cannot slip through under different tag spellings.
- The preparation model derives its default channel order from the structured distribution policy, keeping later CLI and preflight work aligned with the documented supported channels.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 16 now has one reusable semver and tag contract for preflight probes, JSON output, and CLI rendering.
- Plan `16-02` can add git and prerequisite checks without reopening version, tag, or channel-order decisions.

## Self-Check

PASSED

- FOUND: `.planning/phases/16-release-versioning-and-preflight-guardrails/16-01-SUMMARY.md`
- FOUND: `1b8ae5d`
- FOUND: `a1b4bee`

---
*Phase: 16-release-versioning-and-preflight-guardrails*
*Completed: 2026-03-17*
