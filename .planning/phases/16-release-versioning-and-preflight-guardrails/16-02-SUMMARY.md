---
phase: 16-release-versioning-and-preflight-guardrails
plan: "02"
subsystem: release
tags: [release, git, github-actions, json, preflight]
requires:
  - phase: 16-01
    provides: canonical release version, tag normalization, and base release preparation model
provides:
  - git-backed release preflight checks for dirty worktrees and local or remote tag conflicts
  - repository prerequisite validation and per-channel readiness states for current release contracts
  - deterministic JSON output for version, tag, channels, checks, blockers, and warnings
affects: [phase-17-release-orchestration, phase-18-publication-fan-out, phase-19-operator-guidance]
tech-stack:
  added: []
  patterns: [injectable git probe for release checks, machine-readable release preparation output, workflow-contract assertions]
key-files:
  created: []
  modified: [internal/release/prepare.go, internal/release/prepare_test.go, internal/release/release_test.go]
key-decisions:
  - "Remote tag verification failures stay explicit blockers instead of silently downgrading to warnings."
  - "GitHub Release and npm reflect the current workflow as ready, while Homebrew and Scoop stay blocked until publication wiring exists."
patterns-established:
  - "Release preparation composes deterministic git, filesystem, and workflow probes into one reusable contract."
  - "Top-level release JSON always emits channels, checks, blockers, and warnings arrays for downstream automation."
requirements-completed: [REL-02, REL-03]
duration: 10min
completed: 2026-03-17
---

# Phase 16 Plan 02: Git And Prerequisite Preflight Summary

**Git-backed release preflight now blocks dirty worktrees and tag hazards, validates repository release prerequisites, and emits a stable JSON review payload for downstream automation**

## Performance

- **Duration:** 10 min
- **Started:** 2026-03-17T20:10:30Z
- **Completed:** 2026-03-17T20:20:40Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Added an injectable git probe layer that checks worktree cleanliness plus local and remote tag conflicts before release mutation.
- Added repository prerequisite checks and per-channel readiness results aligned to the current workflow and operator checklist.
- Added deterministic JSON serialization and contract tests so later phases can consume one shared preparation payload.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add git-backed release preflight probes and blockers** - `fd1ef6d` (feat)
2. **Task 2: Add prerequisite checks and machine-readable review output** - `053b6ef` (feat)

**Plan metadata:** pending final docs commit

## Files Created/Modified

- `internal/release/prepare.go` - release preparation model, git probe abstraction, prerequisite checks, channel readiness, and JSON output
- `internal/release/prepare_test.go` - coverage for dirty worktrees, remote tag failures, prerequisite checks, and JSON payload shape
- `internal/release/release_test.go` - repo-contract assertions for prerequisite files and publication credentials

## Decisions Made

- Remote tag lookup failures are blockers because publication-oriented preparation cannot safely proceed without proving the target tag is free.
- Channel readiness stays truthful to current repository wiring: GitHub Release and npm are ready from the workflow contract, while Homebrew and Scoop remain blocked until future automation exists.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Treat the operator checklist as a required prerequisite**
- **Found during:** Task 2 (Add prerequisite checks and machine-readable review output)
- **Issue:** The readiness logic already consumed `docs/release-checklist.md`, but the file was not treated as a required prerequisite, which could have let a missing checklist silently degrade channel evaluation.
- **Fix:** Added `docs/release-checklist.md` to the required prerequisite set and covered it in prerequisite tests.
- **Files modified:** `internal/release/prepare.go`, `internal/release/prepare_test.go`
- **Verification:** `go test ./internal/release`
- **Committed in:** `053b6ef`

---

**Total deviations:** 1 auto-fixed (1 missing critical)
**Impact on plan:** No scope creep. The extra prerequisite check keeps the readiness model honest.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 17 can consume one deterministic preparation contract instead of rediscovering release facts from git and workflow files.
- Homebrew and Scoop publication remain intentionally blocked until their automation is implemented in later phases.

## Self-Check: PASSED

- Verified `.planning/phases/16-release-versioning-and-preflight-guardrails/16-02-SUMMARY.md` exists.
- Verified task commits `fd1ef6d` and `053b6ef` exist in git history.

---
*Phase: 16-release-versioning-and-preflight-guardrails*
*Completed: 2026-03-17*
