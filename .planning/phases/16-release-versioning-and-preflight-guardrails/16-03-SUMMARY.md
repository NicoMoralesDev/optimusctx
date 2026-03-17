---
phase: 16-release-versioning-and-preflight-guardrails
plan: "03"
subsystem: cli
tags: [release, cli, json, git, preflight]
requires:
  - phase: 16-release-versioning-and-preflight-guardrails
    provides: canonical semver normalization, tag conflict detection, and shared release preparation data
provides:
  - operator-facing `optimusctx release prepare` entrypoint
  - human-readable and JSON release-plan rendering from shared preparation results
  - explicit Phase 16 review gate that never creates a tag or starts publication
affects: [phase-17-release-orchestration, release-operator-workflow, cli-help, preflight-automation]
tech-stack:
  added: []
  patterns: [thin-cli-over-shared-release-model, repo-root-aware-git-probe, review-only-confirm-gate]
key-files:
  created: [internal/cli/release.go, internal/cli/release_test.go]
  modified: [internal/cli/root.go, .planning/phases/16-release-versioning-and-preflight-guardrails/16-03-SUMMARY.md]
key-decisions:
  - "The CLI resolves repository root and active milestone before calling `internal/release.PrepareRelease`, so default version proposals stay tied to the current release lane."
  - "Text and JSON output both render the shared `ReleasePreparation` result instead of rebuilding version, tag, or channel logic in the CLI."
  - "Phase 16 confirmation is explicit but review-only: the command can acknowledge the plan and return blocker-driven exit codes without mutating git state or publishing."
patterns-established:
  - "Release commands should stay thin wrappers over `internal/release` contracts and only add presentation or operator-flow behavior."
  - "Repo-root-sensitive CLI release flows should run git probes from the resolved repository root, not the caller's cwd."
requirements-completed: [REL-01, REL-03]
duration: 13min
completed: 2026-03-17
---

# Phase 16 Plan 03: Release Prepare CLI Summary

**`optimusctx release prepare` now proposes milestone-scoped versions and canonical tags, renders channel readiness in text or JSON, and enforces a review-only confirm gate that stops before tag creation or publication.**

## Performance

- **Duration:** 13min
- **Started:** 2026-03-17T20:25:35Z
- **Completed:** 2026-03-17T20:38:22Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Added a top-level `release` command with a discoverable `prepare` subcommand in the root help surface.
- Wired the CLI to resolve repo root and milestone, then call the shared release-preparation logic for version, tag, channel, blocker, and warning data.
- Added the explicit Phase 16 review gate: `--json`, repeated `--channel`, `--confirm`, and blocker-driven non-zero exits without tag creation or publication.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add the `optimusctx release prepare` command surface** - `860335f` (feat)
2. **Task 2: Add the review and confirm gate without tag creation** - `1484163` (feat)
3. **Auto-fix: Align blocker semantics in Task 1 JSON coverage** - `5187e25` (fix)

## Files Created/Modified

- `internal/cli/release.go` - Implements the `release prepare` command, repo-root-aware git probing, milestone loading, text rendering, JSON output, and review-only confirmation.
- `internal/cli/release_test.go` - Covers root help discoverability, version override, channel selection, JSON output, help text, blocker handling, and confirm-gate behavior.
- `internal/cli/root.go` - Registers `release` as a first-class top-level command in the CLI dispatcher and help output.

## Decisions Made

- The command reads the active milestone from `.planning/STATE.md` after resolving repository root, which keeps the default proposal aligned with the current release series without duplicating semver logic.
- CLI JSON output includes operator-flow metadata such as `status`, `confirmed`, `nextStep`, and `phaseBoundary`, while the underlying version/tag/channel/check data still comes directly from `ReleasePreparation`.
- Unknown `--channel` values are rejected in the CLI before invoking release preparation, preventing silent empty selections.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Aligned the Task 1 JSON-path test with the new blocker gate**
- **Found during:** Final verification after Task 2
- **Issue:** `TestReleasePrepareCommand` still created blocker fixtures but expected a zero exit after Task 2 made blocker results return non-zero.
- **Fix:** Kept the JSON smoke test focused on machine-readable output and left blocker exit-code assertions in `TestReleasePrepareConfirmGate`.
- **Files modified:** `internal/cli/release_test.go`
- **Verification:** `go test ./internal/cli -run 'TestReleasePrepareCommand'` and `go test ./internal/cli -run 'Test(ReleasePrepareHelp|ReleasePrepareConfirmGate)'`
- **Committed in:** `5187e25`

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** The fix was required to keep the final test suite aligned with the intended blocker-gate behavior. No scope creep.

## Issues Encountered

- `git apply --3way` could not be used to replay the Task 2 delta because the sandbox denied `.git/index.lock` writes. The Task 2 delta was reapplied with direct file edits instead.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 17 can build release orchestration on top of one stable front door that already exposes proposed version, canonical tag, selected channels, blocker reasons, and a review-only confirmation step.
- The real-repo smoke check still reports expected blockers from the dirty `.planning/config.json` worktree state, unavailable remote-tag DNS resolution in the sandbox, and still-blocked Homebrew/Scoop publication wiring.

## Self-Check

PASSED

- Found `.planning/phases/16-release-versioning-and-preflight-guardrails/16-03-SUMMARY.md`
- Found commit `860335f`
- Found commit `1484163`
- Found commit `5187e25`

---
*Phase: 16-release-versioning-and-preflight-guardrails*
*Completed: 2026-03-17*
