---
phase: 16-release-versioning-and-preflight-guardrails
verified: 2026-03-17T21:28:28Z
status: human_needed
score: 6/6 must-haves verified
re_verification:
  previous_status: gaps_found
  previous_score: 5/6
  gaps_closed:
    - "Operator can review and confirm the exact release plan for the selected target channels before any mutation or publication."
  gaps_remaining: []
  regressions: []
human_verification:
  - test: "Selected-channel release-prepare smoke test"
    expected: "`optimusctx release prepare --channel github-release-archive --channel npm --confirm` succeeds from a clean worktree, confirms only `github-release-archive, npm`, and states that no tag was created and publication was not started."
    why_human: "The real command depends on the current repository worktree and live `git ls-remote` behavior, which the automated tests stub."
  - test: "Real-repo invalid-state preflight smoke test"
    expected: "A dirty worktree or real tag conflict causes `optimusctx release prepare` to exit non-zero before any publication begins, with the blocker surfaced in text or JSON output."
    why_human: "Programmatic coverage verifies the logic with fakes, but only a repo-level smoke test proves the operator experience against actual git state."
---

# Phase 16: Release Versioning and Preflight Guardrails Verification Report

**Phase Goal:** Create a guided release-preparation flow that proposes a version and tag, validates release prerequisites, and stops cleanly before publication when the release state is invalid.
**Verified:** 2026-03-17T21:28:28Z
**Status:** human_needed
**Re-verification:** Yes - after gap closure

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | Operator can launch one release-preparation entrypoint that proposes the next release version and normalized tag. | ✓ VERIFIED | `release` is registered on the root CLI in `internal/cli/root.go:46`; `runReleasePrepareCommand` resolves the repo and milestone, then calls the shared preparation service in `internal/cli/release.go:81-133`; proposal and normalization remain covered by `TestReleaseVersionProposal`, `TestReleaseTagNormalization`, and `TestReleasePrepareCommand` in `internal/release/prepare_test.go:12-93` and `internal/cli/release_test.go:15-180`. |
| 2 | Phase 16 enforces one canonical `MAJOR.MINOR.PATCH` version format and one canonical `vMAJOR.MINOR.PATCH` tag format, with semantic-equivalent tags treated as conflicts. | ✓ VERIFIED | `NormalizeReleaseVersion`, `NormalizeReleaseTag`, and `CanonicalizeExistingTag` remain implemented in `internal/release/prepare.go`; regression coverage passed in `TestReleaseTagNormalization`, `TestReleaseVersionProposal`, and `TestReleaseSemanticTagConflicts` (`internal/release/prepare_test.go:12-148`). |
| 3 | Invalid release state blocks the flow before publication starts. | ✓ VERIFIED | `applyGitPreflight` blocks dirty worktrees, exact or semantic tag conflicts, remote tag conflicts, and remote lookup failures before returning a ready plan; `applyPrerequisiteChecks` blocks missing required release files (`internal/release/prepare.go:353-528`). Regression tests passed for `TestReleasePreflight`, `TestReleaseWorktreeBlockers`, `TestReleaseRemoteTagConflicts`, and `TestReleasePrerequisiteChecks` (`internal/release/prepare_test.go:150-288`). |
| 4 | The preparation flow emits machine-readable output that later automation can reuse directly. | ✓ VERIFIED | `ReleasePreparation.MarshalJSON` emits non-nil `channels`, `checks`, `warnings`, and `blockers` in `internal/release/prepare.go:282-290`; CLI JSON output wraps the shared model in `internal/cli/release.go:349-366`. `TestReleasePlanJSON`, `TestReleasePrepareCommand`, and `TestReleasePrepareSelectedChannelsReady` cover the contract (`internal/release/prepare_test.go:330-360`, `internal/cli/release_test.go:125-169`, `internal/cli/release_test.go:300-392`). |
| 5 | Confirmation in this phase is review-only and does not create a tag or start publication. | ✓ VERIFIED | The text renderer prints `release plan confirmed`, `confirmed tag`, `confirmed channels`, and the explicit boundary `no tag created; publication not started` only when blockers are empty (`internal/cli/release.go:307-331`). Confirm behavior remains covered by `TestReleasePrepareConfirmGate` and the selected-channel confirm regression in `internal/cli/release_test.go:203-298` and `internal/cli/release_test.go:394-453`. |
| 6 | The exact selected target channels determine whether the plan is blocked. | ✓ VERIFIED | `defaultReleaseChannels` preserves full channel visibility while marking only requested channels as selected (`internal/release/prepare.go:303-319`), and `setChannelReadiness` now appends a blocker only when a blocked channel is also selected (`internal/release/prepare.go:651-680`). `TestReleaseSelectedChannelsDoNotInheritUnselectedBlockers` proves Homebrew and Scoop stay blocked but non-blocking for a GitHub+npm subset (`internal/release/prepare_test.go:290-328`), and `TestReleasePrepareSelectedChannelsReady` verifies the JSON and `--confirm` CLI flows stay ready and review-only for that same subset (`internal/cli/release_test.go:300-453`). |

**Score:** 6/6 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `internal/release/prepare.go` | Shared release-preparation contract, canonical semver and tag logic, git and prerequisite preflight, selected-channel-aware blocker scope | ✓ VERIFIED | Exists, substantive (868 lines), and wired from the CLI. The prior blocker bug is closed by the `selected` gate in `setChannelReadiness`. |
| `internal/release/prepare_test.go` | Coverage for normalization, conflicts, preflight, prerequisites, JSON output, and selected-channel blocker scope | ✓ VERIFIED | Exists, substantive (468 lines), and wired by `go test ./internal/release`. Includes the gap-closure test `TestReleaseSelectedChannelsDoNotInheritUnselectedBlockers`. |
| `internal/cli/release.go` | Operator-facing `optimusctx release prepare` command with text, JSON, and review-only confirm behavior | ✓ VERIFIED | Exists, substantive (430 lines), and delegates to the shared release-preparation model instead of duplicating readiness logic. |
| `internal/cli/release_test.go` | CLI regression coverage for guided prepare, JSON output, blocker exits, selected-channel readiness, and review-only confirmation | ✓ VERIFIED | Exists, substantive (497 lines), and wired by `go test ./internal/cli`. Includes selected-channel JSON and confirm regressions for the exact Phase 16 gap closure. |
| `internal/cli/root.go` | Top-level CLI registration and help discoverability for `release` | ✓ VERIFIED | Exists, substantive (96 lines), and dispatches the `release` command from the root CLI help and execution path. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `internal/cli/root.go` | `internal/cli/release.go` | Root command dispatches `release` and exposes it in help text | WIRED | `Execute` routes `release` to `newReleaseCommand().Run(...)`, and help prints the release summary. |
| `internal/cli/release.go` | `internal/release/prepare.go` | `defaultReleasePrepareCommandService()` calls `release.PrepareRelease()` | WIRED | The CLI remains a thin wrapper over the shared preparation model for both text and JSON output. |
| `internal/release/prepare.go` | `internal/release/prepare_test.go` | Shared selected-channel blocker logic is exercised by unit tests | WIRED | `TestReleaseSelectedChannelsDoNotInheritUnselectedBlockers` proves blocked unselected channels stay informational. |
| `internal/cli/release.go` | `internal/cli/release_test.go` | Operator-facing JSON and `--confirm` flows are checked against shared selected-channel state | WIRED | `TestReleasePrepareSelectedChannelsReady` verifies `status`, channel selection, and review-only confirm text for the GitHub+npm subset. |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| `REL-01` | `16-01`, `16-03` | Operator can start a release from an interactive or guided configuration flow that proposes the next version and normalized git tag before publication begins. | ✓ SATISFIED | Phase 16 exposes one `release prepare` entrypoint, proposes the next version from milestone and tag state, and normalizes the derived tag through the shared preparation model and CLI wrapper. |
| `REL-02` | `16-02` | Operator gets preflight validation for duplicate tags, worktree state, and required release prerequisites before any tag creation or channel publication runs. | ✓ SATISFIED | Dirty worktree, local and remote tag conflicts, remote tag lookup failures, and missing prerequisite files become blockers inside `PrepareRelease`, with passing release-layer regression coverage. |
| `REL-03` | `16-02`, `16-03`, `16-04` | Operator can review and confirm the exact release plan, derived tag, and target channels before the release process mutates git state or publishes artifacts. | ✓ SATISFIED | The selected-channel blocker gap is closed in `setChannelReadiness`, and the CLI now has passing regressions proving `--channel github-release-archive --channel npm` stays ready in JSON and review-only on `--confirm` while Homebrew and Scoop remain visible but unselected. |

Orphaned phase requirements from `REQUIREMENTS.md`: none. All declared Phase 16 requirement IDs in plan frontmatter are accounted for: `REL-01`, `REL-02`, and `REL-03`.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| --- | --- | --- | --- | --- |
| None | - | No TODO, placeholder, empty-implementation, or console-only stub patterns found in the modified Phase 16 gap-closure files. | - | No automated blocker or warning surfaced from the anti-pattern scan. |

### Human Verification Required

### 1. Selected-Channel Release-Prepare Smoke Test

**Test:** Run `optimusctx release prepare --channel github-release-archive --channel npm --confirm` from a clean repository checkout with remote tag lookup available.
**Expected:** The command succeeds, confirms only `github-release-archive, npm`, and states that no tag was created and publication was not started.
**Why human:** The real operator flow depends on the live repository worktree and `git ls-remote` behavior, which the automated tests stub out.

### 2. Invalid-State Preflight Smoke Test

**Test:** Introduce a dirty worktree or a real tag conflict, then run `optimusctx release prepare` and `optimusctx release prepare --json`.
**Expected:** The command exits non-zero before any publication begins and surfaces the blocker clearly in text and JSON output.
**Why human:** The preflight logic is covered by unit tests, but a real-repo smoke test is still the most direct proof that the operator-facing failure mode is correct.

### Gaps Summary

The previous selected-channel blocker-scope gap is closed. The shared release-preparation model now keeps Homebrew and Scoop visible as blocked channels while restricting blocker propagation to the exact selected plan, which restores the intended `REL-03` contract for targeted prepare flows.

Automated verification found no remaining code-level gaps for Phase 16. What remains is real-repository smoke validation against live git state and remote-tag lookup, so the phase is recorded as `human_needed` rather than another implementation failure.

---

_Verified: 2026-03-17T21:28:28Z_
_Verifier: Claude (gsd-verifier)_
