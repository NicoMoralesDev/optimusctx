---
phase: 16-release-versioning-and-preflight-guardrails
plan: "04"
subsystem: release
tags: [release, cli, testing, versioning]
requires:
  - phase: 16-02
    provides: shared release-preparation checks, per-channel readiness, and JSON-safe release review state
  - phase: 16-03
    provides: operator-facing release prepare command and review-only confirmation gate
provides:
  - selected-channel-aware blocker scoping in the shared release-preparation model
  - regression coverage for selected channel subsets in release preparation and CLI flows
  - proof that Phase 16 confirmation remains review-only for ready selected-channel plans
affects: [17-release orchestration, 18-multi-channel publication, REL-03]
tech-stack:
  added: []
  patterns:
    - shared release-preparation model remains the single source of truth for blocker scope
    - CLI tests assert operator-visible output from prepared channel state without duplicating selection logic
key-files:
  created:
    - .planning/phases/16-release-versioning-and-preflight-guardrails/16-04-SUMMARY.md
  modified:
    - internal/release/prepare.go
    - internal/release/prepare_test.go
    - internal/cli/release_test.go
key-decisions:
  - "Blocked readiness only becomes a blocker when the matching ReleaseChannelPlan is selected."
  - "Homebrew and Scoop stay visible as blocked channels even when they are informational for a narrower selected release plan."
patterns-established:
  - "Selected-channel blocker scope lives in internal/release/prepare.go so text and JSON stay aligned."
  - "CLI regression tests should stub full ReleasePreparation results and assert output contracts rather than reimplementing readiness rules."
requirements-completed: [REL-03]
duration: 2min
completed: 2026-03-17
---

# Phase 16 Plan 04: Release blocker scope follows the exact selected channel set

**Selected-channel release preparation now stays ready for GitHub Release archives plus npm while leaving Homebrew and Scoop truthfully blocked but informational**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-17T21:20:37Z
- **Completed:** 2026-03-17T21:22:31Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Shared release preparation now appends channel blockers only for selected blocked channels while keeping readiness checks for every supported channel.
- Release unit coverage proves `github-release-archive` plus `npm` can be ready even when Homebrew and Scoop remain blocked and unselected.
- CLI regression coverage proves `release prepare --channel github-release-archive --channel npm` stays ready in JSON and review-only on `--confirm`.

## Task Commits

Each task was committed atomically:

1. **Task 1: Make blocker scope follow the selected channel set in shared preparation** - `7aba993` (fix)
2. **Task 2: Lock the selected-channel operator contract in CLI regression tests** - `8ef8c87` (test)

## Files Created/Modified
- `internal/release/prepare.go` - gates channel blockers on the selected plan while preserving per-channel readiness and checks
- `internal/release/prepare_test.go` - covers selected GitHub Release archive plus npm preparation without inherited Homebrew or Scoop blockers
- `internal/cli/release_test.go` - proves JSON and confirm flows stay ready and review-only for the exact selected subset

## Decisions Made
- Kept blocker scoping inside `setChannelReadiness` so the shared `ReleasePreparation` model remains authoritative for both CLI text and JSON output.
- Preserved blocked readiness for Homebrew and Scoop when unselected so the operator still sees truthful distribution-policy status without false plan failures.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- `gsd-tools state advance-plan` and `roadmap update-plan-progress` updated progress counts, but left stale Phase 16 milestone and plan text in `STATE.md` and the 16-04 checklist entry in `ROADMAP.md`. I corrected those planning artifacts manually so the final metadata matches the executed plan.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 16 now satisfies the remaining selected-channel blocker-scope truth for `REL-03`.
- Later orchestration and publication phases can consume the shared preparation model without adding separate channel-selection logic in the CLI layer.

## Self-Check

PASSED

- Found `.planning/phases/16-release-versioning-and-preflight-guardrails/16-04-SUMMARY.md`
- Found commit `7aba993`
- Found commit `8ef8c87`

---
*Phase: 16-release-versioning-and-preflight-guardrails*
*Completed: 2026-03-17*
