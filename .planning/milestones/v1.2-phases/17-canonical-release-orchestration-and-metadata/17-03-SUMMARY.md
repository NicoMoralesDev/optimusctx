---
phase: 17-canonical-release-orchestration-and-metadata
plan: "03"
subsystem: infra
tags: [release, metadata, npm, homebrew, scoop, testing]
requires:
  - phase: 17-canonical-release-orchestration-and-metadata
    provides: canonical release metadata and create-versus-reuse orchestration planning
provides:
  - package-manager derivation from canonical release assets and checksum manifest data
  - npm package metadata derivation from canonical release assets, tag, and checksum manifest data
  - render-script regression coverage that locks npm output to the canonical release contract
affects: [phase-17-plan-04, phase-18, package-managers, npm-publication]
tech-stack:
  added: []
  patterns:
    - canonical-release-fed downstream consumers
    - committed npm manifest retagging instead of duplicate URL assembly
key-files:
  created: []
  modified:
    - internal/release/package_manager.go
    - internal/release/package_manager_test.go
    - internal/release/npm_package.go
    - internal/release/npm_package_test.go
    - internal/release/release_test.go
    - scripts/render-npm-package.sh
key-decisions:
  - "Homebrew, Scoop, and npm now consume canonical release asset URLs and checksum metadata instead of reconstructing tagged GitHub Release paths locally."
  - "The npm render script preserves the committed Go-rendered package.json structure and retags its canonical release URLs for the requested version instead of inventing a second release URL rule set."
patterns-established:
  - "Canonical downstream derivation: package-manager and npm helpers build consumer metadata from CanonicalRelease asset records."
  - "Render-script contract checks: release tests execute scripts/render-npm-package.sh and compare its package.json output against the Go-rendered npm manifest."
requirements-completed: [PUB-01]
duration: 7m
completed: 2026-03-17
---

# Phase 17 Plan 03: Downstream Consumer Rewiring Summary

**Canonical release assets, checksum manifests, and tagged download URLs now drive Homebrew, Scoop, and npm metadata from one shared release contract**

## Performance

- **Duration:** 7m
- **Started:** 2026-03-17T22:43:30Z
- **Completed:** 2026-03-17T22:50:41Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Rewired package-manager release derivation so Homebrew and Scoop assets now come directly from `CanonicalRelease` asset and checksum-manifest data.
- Rewired npm package metadata so release tags, per-platform archive URLs, and checksum-manifest URLs now come from the canonical release contract.
- Locked the npm render path to the canonical contract by retagging the committed manifest and comparing script output against the Go-rendered release manifest in tests.

## Task Commits

Each task was committed atomically:

1. **Task 1: Rewire package-manager and npm Go helpers to consume canonical release metadata** - `02e4887` (feat)
2. **Task 2: Keep the npm render entrypoint release-derived without duplicate asset math** - `82f351a` (feat)

**Plan metadata:** pending final docs commit

## Files Created/Modified

- `internal/release/package_manager.go` - Builds Homebrew and Scoop asset metadata from `CanonicalRelease` assets and checksum-manifest data.
- `internal/release/package_manager_test.go` - Adds package-manager contract coverage and the task-specific render test names.
- `internal/release/npm_package.go` - Builds npm package metadata from canonical release assets, tag, and checksum-manifest fields.
- `internal/release/npm_package_test.go` - Adds canonical release contract assertions for npm package metadata.
- `internal/release/release_test.go` - Verifies the publish script and downstream helpers still align to one canonical tagged release.
- `scripts/render-npm-package.sh` - Retags the committed npm manifest’s canonical URLs and archive names for a requested release version.

## Decisions Made

- Canonical release asset inventory remains the single source for downstream archive filenames and download URLs, while package-manager-specific checksums are still read from the checksum manifest payload.
- The npm render script mutates the Go-rendered development manifest into a release manifest and validates the expected archive naming contract instead of deriving repository URLs and archives from scratch.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Existing planning metadata already contained uncommitted Phase 17 plan 02 state and summary updates, so plan-state updates for 17-03 were applied on top of that newer base instead of reverting to the older checked-in 17-01 state.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 18 can fan out npm, Homebrew, and Scoop publication from one canonical release asset contract without re-deriving tagged archive URLs per channel.
- Plan `17-04` can document and regression-lock the canonical-root release contract now that downstream consumers already read from it.

---
*Phase: 17-canonical-release-orchestration-and-metadata*
*Completed: 2026-03-17*

## Self-Check: PASSED

- Found `.planning/phases/17-canonical-release-orchestration-and-metadata/17-03-SUMMARY.md`
- Found task commit `02e4887`
- Found task commit `82f351a`
