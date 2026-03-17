---
phase: 15-add-npm-and-npx-distribution-option
plan: "03"
subsystem: distribution
tags: [npm, ci, github-actions, docs, policy, release]
requires:
  - phase: 15-add-npm-and-npx-distribution-option
    provides: committed npm package identity and runtime wrapper files from plans 15-01 and 15-02
provides:
  - npm publication wired into the tagged GitHub release workflow
  - deterministic package rendering for CI publication
  - updated distribution policy, README, and checklist that include npm while preserving existing scope boundaries
affects: [phase verification, release operations, supported-channel messaging]
tech-stack:
  added: [none]
  patterns: [post-release npm publication, release-derived package rendering, supported-channel policy enforcement]
key-files:
  created: [scripts/render-npm-package.sh]
  modified: [.github/workflows/release.yml, README.md, docs/distribution-strategy.md, docs/release-checklist.md, internal/release/distribution_plan.go, internal/release/distribution_plan_test.go, internal/release/release_test.go, internal/cli/install_test.go]
key-decisions:
  - "npm publication runs after the GitHub Release archives succeed so the npm channel remains downstream of the canonical tagged binary lifecycle."
  - "The render script copies committed package files and injects only tag-derived version and archive metadata, avoiding a second build definition in GoReleaser."
  - "Distribution policy now treats npm as an allowed wrapper channel while keeping GitHub Release archives as the rollback fallback and leaving broader installer ecosystems out of scope."
patterns-established:
  - "Publication sequencing: GitHub Release archives publish first, then npm renders from committed package files and publishes against the same tag."
  - "Policy guardrails: docs and tests explicitly allow npm while continuing to reject WinGet, Chocolatey, apt, dnf, yum, and other unsupported installer claims."
requirements-completed: [DIST-02, DIST-04]
duration: 4min
completed: 2026-03-17
---

# Phase 15 Plan 03: npm Publication and Distribution Policy Summary

**Tagged-release npm publication workflow, deterministic package render script, and supported-channel policy/docs updated for npm**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-17T13:13:28Z
- **Completed:** 2026-03-17T13:17:19Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments

- Added an `npm` publication job to the release workflow that runs after GitHub Release archive publication, verifies the npm publish contract, renders the package, and publishes with `NPM_TOKEN`.
- Added `scripts/render-npm-package.sh` to copy the committed wrapper package, inject the tag-derived version and release metadata, and fail fast if required package files are missing.
- Updated the README, distribution strategy, release checklist, distribution policy model, and guardrail tests so npm is a first-class supported wrapper channel without loosening broader installer scope.

## Task Commits

Each task was committed atomically:

1. **Task 1: Publish the npm package from the tagged release workflow** - `f10b3a2` (ci)
2. **Task 2: Update supported-channel policy, checklist, and docs tests for npm** - `bb73188` (docs)

## Files Created/Modified

- `.github/workflows/release.yml` - tagged release workflow with post-release npm publication
- `scripts/render-npm-package.sh` - deterministic npm package renderer for CI publication
- `internal/release/release_test.go` - workflow and publish-contract assertions for npm publication
- `README.md` - top-level supported-channel messaging that now includes npm and `npx`
- `docs/distribution-strategy.md` - supported-channel and rollback policy updated to include npm
- `docs/release-checklist.md` - release checklist updated with npm publication and verification steps
- `internal/release/distribution_plan.go` - distribution policy model expanded with the npm wrapper channel
- `internal/release/distribution_plan_test.go` - policy and doc guardrails updated for npm support
- `internal/cli/install_test.go` - README/package-manager doc assertions updated for npm support

## Decisions Made

- npm publication remains a downstream wrapper step after the canonical archive release completes.
- The npm channel is documented as a wrapper over the tagged binary, not as a second runtime implementation.
- GitHub Release archives remain the fallback and rollback path even after npm publication is added.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The first checklist update duplicated two existing bullets from the earlier Homebrew/Scoop-only wording; the duplicate lines were removed before the final policy verification run.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 15 now has all planned code, workflow, documentation, and policy updates required for npm support.
- The phase is ready for goal-level verification against `DIST-02`, `DIST-03`, and `DIST-04`.

## Self-Check

PASSED

---
*Phase: 15-add-npm-and-npx-distribution-option*
*Completed: 2026-03-17*
