---
phase: 19-operator-verification-recovery-and-end-to-end-guide
verified: 2026-03-18T18:25:57Z
status: human_needed
score: 3/3 observable truths verified
gaps: []
human_verification:
  - test: "GitHub Actions summary render check"
    command: "Trigger `.github/workflows/release.yml` against a safe tag or existing rerun target, then inspect the hosted run summary UI."
    status: pending
    observed: "Repository-only verification locks summary wording and job wiring, but a live GitHub Actions run is still needed to confirm the rendered summary surface."
  - test: "Operator guide end-to-end read-through"
    command: "Follow `docs/operator-release-guide.md` from `optimusctx release prepare` through verification and the rerun/rollback decision tree."
    status: pending
    observed: "Docs and contract tests are aligned, but a human operator still needs to confirm the guide reads cleanly without repo-internal context."
---

# Phase 19: Operator Verification, Recovery, and End-to-End Guide Verification Report

**Phase Goal:** Document and verify the complete operator workflow for release, republish, verification, and rollback across all supported channels.
**Verified:** 2026-03-18T18:25:57Z
**Status:** human_needed
**Re-verification:** No

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | One release workflow run exposes GitHub Release, npm, Homebrew, and Scoop status with failure reason and next-step guidance from the same operator-facing surface. | ✓ VERIFIED | `.github/workflows/release.yml` now emits `### GitHub Release publication`, `### npm publication`, `### Homebrew publication`, and `### Scoop publication` blocks with `channel`, `tag`, `outcome`, `failure_reason`, and `next_step`, and `go test ./internal/release -run 'Test(ReleaseWorkflowSummaryShowsChannelStatus|ReleaseWorkflowSummaryShowsFailureGuidance)$'` passed. |
| 2 | The repository has one canonical operator guide that starts with `optimusctx release prepare`, verifies the canonical GitHub Release first, then covers downstream verification, selective rerun, and rollback. | ✓ VERIFIED | `docs/operator-release-guide.md` now anchors the release workflow, `docs/release-checklist.md` and `docs/install-and-verify.md` link to it explicitly, and `go test ./internal/release ./internal/cli -run 'Test(OperatorReleaseGuideStaysCanonical|InstallAndVerifyGuideLinksOperatorFlow)$'` passed. |
| 3 | Recovery policy consistently says to fix GitHub Release first, rerun only the affected downstream channel with exact workflow inputs, and treat a prior tagged GitHub Release archive as the rollback source. | ✓ VERIFIED | `docs/distribution-strategy.md` now points to `docs/operator-release-guide.md`, documents `gh workflow run release.yml -f release_tag=... -f publication_channel=...`, and rejects channel-native rollback as the canonical path; `go test ./internal/release -run 'Test(OperatorRecoveryGuideStaysCanonical|DistributionDocsStayWithinSupportedScope|ChannelPublicationWorkflowSelectiveRerun)$'` passed. |

**Score:** 3/3 observable truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `.github/workflows/release.yml` | Per-channel operator summary surface with failure and next-step guidance | ✓ VERIFIED | Summary blocks exist for GitHub Release, npm, Homebrew, and Scoop and are locked by release tests. |
| `docs/operator-release-guide.md` | Canonical prepare, publish, verify, rerun, and rollback procedure | ✓ VERIFIED | Exists, substantive, and linked from checklist and install docs. |
| `docs/release-checklist.md` | Checklist entry point that routes operators to the canonical guide | ✓ VERIFIED | Exists and references `./operator-release-guide.md` plus exact rerun input names. |
| `docs/install-and-verify.md` | Install guide that points release operators to the canonical guide while preserving shipped verification commands | ✓ VERIFIED | Exists and references `./operator-release-guide.md`, `workflow_dispatch`, `release_tag`, and `publication_channel`. |
| `docs/distribution-strategy.md` | Recovery-policy layer aligned to the canonical operator guide | ✓ VERIFIED | Exists and documents fix-first, selective rerun, and archive-root rollback. |
| `internal/release/release_test.go` | Workflow and operator-doc contract tests | ✓ VERIFIED | Contains summary-surface and operator-guide contract coverage. |
| `internal/cli/install_test.go` | Install-guide linkage tests | ✓ VERIFIED | Contains the operator-flow linkage regression. |
| `internal/release/distribution_plan_test.go` | Recovery-policy tests that reject unsupported guidance | ✓ VERIFIED | Contains canonical rerun/rollback assertions and rejects `npm unpublish` or unsupported channels. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `.github/workflows/release.yml` | `internal/release/release_test.go` | Exact summary headings, keys, and rerun wording are contract-tested | WIRED | Release tests prove the workflow keeps the canonical-root summary surface and forbids `publication_channel=github-release`. |
| `docs/operator-release-guide.md` | `.github/workflows/release.yml` | The guide reuses the exact `workflow_dispatch`, `release_tag`, and `publication_channel` rerun contract | WIRED | The guide documents `gh workflow run release.yml -f release_tag=... -f publication_channel=...` and the exact allowed channel values. |
| `docs/release-checklist.md` | `docs/operator-release-guide.md` | The checklist defers end-to-end release operation to the canonical guide | WIRED | Checklist entries now point at `./operator-release-guide.md` before and after tag push. |
| `docs/install-and-verify.md` | `docs/operator-release-guide.md` | The install guide routes release operators to the canonical release flow without duplicating it | WIRED | The install guide references `./operator-release-guide.md` and preserves shipped verification commands. |
| `docs/distribution-strategy.md` | `docs/operator-release-guide.md` | Recovery policy points to the canonical guide and mirrors its rerun/rollback split | WIRED | Distribution strategy now references `docs/operator-release-guide.md` and shares the same exact rerun markers. |
| `docs/distribution-strategy.md` | `internal/release/distribution_plan_test.go` | Policy wording is locked and unsupported recovery claims fail tests | WIRED | `TestOperatorRecoveryGuideStaysCanonical` and `TestDistributionDocsStayWithinSupportedScope` enforce the supported recovery boundary. |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| `OPS-06` | `19-01-PLAN` | Operator can see per-channel release status, failure reason, and next-step guidance from one release workflow. | ✓ SATISFIED | Workflow summary blocks and release tests are present and passing. |
| `OPS-07` | `19-02-PLAN` | Operator can follow one documented verification flow that checks the published archive, npm, Homebrew, and Scoop outputs after release. | ✓ SATISFIED (repo evidence) | Canonical operator guide exists, linked docs stay aligned, and linkage tests pass. A final human read-through is still pending before full phase closeout. |
| `OPS-08` | `19-02-PLAN`, `19-03-PLAN` | Operator can follow one documented recovery or rollback path when a channel publish or post-release verification step fails. | ✓ SATISFIED | Operator guide and distribution strategy share exact rerun-versus-rollback wording, and recovery-policy tests pass. |

Orphaned phase requirements from `REQUIREMENTS.md`: none. All declared Phase 19 requirement IDs in plan frontmatter are accounted for: `OPS-06`, `OPS-07`, and `OPS-08`.

### Anti-Patterns Found

None in the verified Phase 19 workflow, docs, or release-policy tests.

### Automated Verification Executed

- `go test ./internal/release -run 'Test(ReleaseWorkflowSummaryShowsChannelStatus|ReleaseWorkflowSummaryShowsFailureGuidance)$'`
- `go test ./internal/release ./internal/cli -run 'Test(OperatorReleaseGuideStaysCanonical|InstallAndVerifyGuideLinksOperatorFlow)$'`
- `go test ./internal/release -run 'Test(OperatorRecoveryGuideStaysCanonical|DistributionDocsStayWithinSupportedScope|ChannelPublicationWorkflowSelectiveRerun)$'`
- `go test ./...`

### Human Verification Required

#### 1. GitHub Actions Summary Render Check

**Test:** Trigger `.github/workflows/release.yml` for a safe tag or inspect an equivalent existing run.
**Expected:** The hosted summary UI shows GitHub Release, npm, Homebrew, and Scoop with channel, tag, outcome, failure reason, and next-step guidance.
**Why human:** Repository tests lock the workflow contract but cannot verify the rendered GitHub Actions UI.

#### 2. Operator Guide Read-Through

**Test:** Follow `docs/operator-release-guide.md` from `optimusctx release prepare` through verification, targeted rerun, and rollback decision points.
**Expected:** The guide is navigable without repo-internal context and the rerun-versus-rollback branches are unambiguous.
**Why human:** Contract tests prove wording alignment, but only a human operator can validate document usability end to end.

### Gaps Summary

Phase 19 is complete at the repository level: workflow summaries, operator docs, recovery policy, and release tests all align with the GitHub-Release-rooted operator contract, and the full Go test suite passed on 2026-03-18. No implementation gaps remain on the current branch.

The only remaining work is human verification of the hosted GitHub Actions summary rendering and the end-to-end operator guide usability. Once those two checks are approved, Phase 19 can be fully closed and `REQUIREMENTS.md` can mark `OPS-07` complete.

---

_Verified: 2026-03-18T18:25:57Z_  
_Verifier: Codex_
