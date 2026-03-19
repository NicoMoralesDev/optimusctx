---
phase: 18-multi-channel-publication-fan-out
verified: 2026-03-19T19:24:43.237Z
status: passed
score: 4/4 observable truths verified
gaps: []
human_verification:
  - test: "Hosted single-channel rerun against a real tagged release"
    command: "Dispatch `.github/workflows/release.yml` with `release_tag=<existing tag>` and `publication_channel=<npm|homebrew|scoop>`, then confirm unrelated channels were untouched."
    status: deferred
    observed: "Repository contracts, UAT evidence, and the full Go suite prove the rerun surface and channel gating on the current branch; one real hosted rerun is still useful as operator confirmation and is covered by the broader Phase 19 operator workflow."
  - test: "Homebrew and Scoop external repository update check"
    command: "Run a real release or staged tag, inspect the resulting `niccrow/homebrew-tap` and `niccrow/scoop-bucket` updates, and confirm rendered versions, URLs, and checksums match the canonical GitHub Release."
    status: deferred
    observed: "Local publication-plan, render, workflow, and docs contracts all pass on the repository; the remaining external push check is a live-environment validation concern rather than a branch implementation gap."
---

# Phase 18: Multi-Channel Publication Fan-Out Verification Report

**Phase Goal:** Automate publication of npm, Homebrew, and Scoop from the canonical release tag, with selective rerun support per channel.
**Verified:** 2026-03-19T19:24:43.237Z
**Status:** passed
**Re-verification:** No

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | One shared downstream publication contract fans out npm, Homebrew, and Scoop from the canonical GitHub Release tag and rejects unsupported downstream channels. | ✓ VERIFIED | `18-UAT.md` records passing publication-plan coverage for `TestPlanReleasePublicationFanout`, `TestPlanReleasePublicationRerun`, `TestPlanReleasePublicationRejectsUnknownChannel`, and `TestPlanReleasePublicationRejectsGitHubArchiveChannel`; the same release-layer contract remains present on the branch under `internal/release/publication.go` and related tests. |
| 2 | Homebrew and Scoop payloads render deterministically from the canonical release tag and checksum manifest, and the shell wrappers stay transport-only. | ✓ VERIFIED | `18-UAT.md` records passing render coverage for `TestRenderHomebrewFormulaForTag`, `TestRenderScoopManifestForTag`, `TestRenderHomebrewFormulaScript`, and `TestRenderScoopManifestScript`; on 2026-03-19 `go test ./internal/release -run 'Test(PlanReleasePublicationFanout|PlanReleasePublicationRerun|PlanReleasePublicationRejectsUnknownChannel|PlanReleasePublicationRejectsGitHubArchiveChannel|RenderHomebrewFormulaForTag|RenderScoopManifestForTag|RenderHomebrewFormulaScript|RenderScoopManifestScript|GitHubReleaseWorkflowReuseContract|NPMPublishWorkflow|ChannelPublicationWorkflowFanout|ChannelPublicationWorkflowSelectiveRerun|HomebrewPublishWorkflow|ScoopPublishWorkflow|MultiChannelPublicationDocsStayCanonical)$'` passed. |
| 3 | The release workflow can fan out to npm, Homebrew, and Scoop and can rerun exactly one requested downstream channel with `workflow_dispatch`, `release_tag`, and `publication_channel` without rebuilding unrelated release assets. | ✓ VERIFIED | `18-UAT.md` records passing workflow coverage for `TestGitHubReleaseWorkflowReuseContract`, `TestNPMPublishWorkflow`, `TestChannelPublicationWorkflowFanout`, `TestChannelPublicationWorkflowSelectiveRerun`, `TestHomebrewPublishWorkflow`, and `TestScoopPublishWorkflow`; the same grouped command passed again on 2026-03-19. |
| 4 | Prepare output, CLI review behavior, and operator docs describe the same canonical-rooted multi-channel publication contract and exact rerun semantics. | ✓ VERIFIED | `18-UAT.md` records passing prepare/CLI/doc coverage for `TestReleasePrepareSelectedChannelsReady`, `TestReleaseSelectedChannelsDoNotInheritUnselectedBlockers`, `TestReleasePrepareAllChannelsReady`, `TestReleasePrepareHomebrewAndScoopAutomationMarkers`, and `TestMultiChannelPublicationDocsStayCanonical`; on 2026-03-19 `go test ./internal/release ./internal/cli -run 'Test(ReleasePrepareSelectedChannelsReady|ReleaseSelectedChannelsDoNotInheritUnselectedBlockers|ReleasePrepareAllChannelsReady|ReleasePrepareHomebrewAndScoopAutomationMarkers)$'` passed, and `go test ./...` passed on the full repository. |

**Score:** 4/4 observable truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `internal/release/publication.go` | Shared downstream publication planning contract for npm, Homebrew, and Scoop | ✓ VERIFIED | Exists and is exercised by the publication-plan tests recorded in `18-UAT.md`. |
| `internal/release/package_manager_publication.go` | Homebrew and Scoop render entry points rooted in canonical release metadata | ✓ VERIFIED | Exists and is exercised by the deterministic render tests recorded in `18-UAT.md`. |
| `.github/workflows/release.yml` | Workflow fan-out plus exact single-channel rerun contract | ✓ VERIFIED | Exists and is exercised by the workflow contract tests re-run on 2026-03-19. |
| `docs/release-checklist.md` | Operator-facing canonical fan-out and rerun guidance | ✓ VERIFIED | Exists and is checked by the multi-channel publication documentation tests. |
| `docs/install-and-verify.md` | Verification and rerun guidance that stays aligned with the workflow contract | ✓ VERIFIED | Exists and is checked by the multi-channel publication documentation tests. |
| `internal/release/release_test.go` | Workflow and documentation regressions for fan-out and selective rerun behavior | ✓ VERIFIED | Exists and is exercised by the grouped release-layer tests re-run on 2026-03-19. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `internal/release/publication.go` | `.github/workflows/release.yml` | Publication planning and workflow inputs share the same channel IDs and rerun model | WIRED | The workflow exposes `publication_channel` values `all`, `npm`, `homebrew`, and `scoop`, and the release-layer rerun tests passed. |
| `.github/workflows/release.yml` | `internal/release/release_test.go` | Workflow fan-out and rerun wording are contract-tested | WIRED | `ChannelPublicationWorkflowFanout`, `ChannelPublicationWorkflowSelectiveRerun`, `HomebrewPublishWorkflow`, and `ScoopPublishWorkflow` remained green on 2026-03-19. |
| `internal/release/prepare.go` | `internal/cli/release_test.go` | Prepare readiness and selected-channel review output stay aligned with the real workflow markers | WIRED | `ReleasePrepareSelectedChannelsReady`, `ReleaseSelectedChannelsDoNotInheritUnselectedBlockers`, `ReleasePrepareAllChannelsReady`, and `ReleasePrepareHomebrewAndScoopAutomationMarkers` passed on 2026-03-19. |
| `docs/release-checklist.md` | `docs/install-and-verify.md` | Operator docs mirror the canonical release root and exact rerun contract | WIRED | The documentation contract tests in `internal/release/release_test.go` remained green on 2026-03-19. |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| `PUB-02` | `18-01`, `18-02`, `18-03`, `18-04` | Operator can publish npm, Homebrew, and Scoop from the same canonical release tag after GitHub Release assets are available. | ✓ SATISFIED | Publication-plan tests, deterministic Homebrew/Scoop rendering tests, workflow fan-out tests, documentation contract tests, and the full `go test ./...` suite all passed. |
| `PUB-03` | `18-01`, `18-03`, `18-04` | Operator can rerun publication for one specific channel against an existing release tag without rebuilding or republishing unrelated channels. | ✓ SATISFIED | `TestPlanReleasePublicationRerun`, `TestChannelPublicationWorkflowSelectiveRerun`, and the prepare/CLI selected-channel tests all passed and remain aligned with the documented `release_tag` plus `publication_channel` rerun contract. |

Orphaned phase requirements from `REQUIREMENTS.md`: none. All declared Phase 18 requirement IDs in summary frontmatter are accounted for: `PUB-02` and `PUB-03`.

### Anti-Patterns Found

None in the verified Phase 18 workflow, publication contract, or operator documentation surfaces.

### Automated Verification Executed

- `go test ./internal/release -run 'Test(PlanReleasePublicationFanout|PlanReleasePublicationRerun|PlanReleasePublicationRejectsUnknownChannel|PlanReleasePublicationRejectsGitHubArchiveChannel|RenderHomebrewFormulaForTag|RenderScoopManifestForTag|RenderHomebrewFormulaScript|RenderScoopManifestScript|GitHubReleaseWorkflowReuseContract|NPMPublishWorkflow|ChannelPublicationWorkflowFanout|ChannelPublicationWorkflowSelectiveRerun|HomebrewPublishWorkflow|ScoopPublishWorkflow|MultiChannelPublicationDocsStayCanonical)$'`
- `go test ./internal/release ./internal/cli -run 'Test(ReleasePrepareSelectedChannelsReady|ReleaseSelectedChannelsDoNotInheritUnselectedBlockers|ReleasePrepareAllChannelsReady|ReleasePrepareHomebrewAndScoopAutomationMarkers)$'`
- `go test ./...`

### Human Verification Notes

The remaining live-environment checks for Phase 18 are operational confirmations against GitHub Actions and external package-manager repositories, not repository-level implementation blockers. Phase 19 already owns the end-to-end operator verification surface that exercises the same rerun and recovery contract from the operator perspective.

### Gaps Summary

Phase 18 is complete at the repository level. The canonical publication contract, deterministic Homebrew and Scoop rendering, selective single-channel rerun model, prepare readiness, CLI review output, and operator documentation all align on the current branch, and the grouped release tests plus the full Go suite passed again on 2026-03-19.

No repository implementation gaps remain for `PUB-02` or `PUB-03`. Live hosted rerun and external repository publication checks are deferred as operator-environment confirmation and do not block milestone archival.

---

_Verified: 2026-03-19T19:24:43.237Z_  
_Verifier: Codex_
