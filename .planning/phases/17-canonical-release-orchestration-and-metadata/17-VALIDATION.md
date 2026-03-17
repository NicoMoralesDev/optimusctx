---
phase: 17
slug: canonical-release-orchestration-and-metadata
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-17
---

# Phase 17 - Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | `go test` |
| **Config file** | `.goreleaser.yml`, `.github/workflows/release.yml`, `.planning/ROADMAP.md`, `.planning/REQUIREMENTS.md` |
| **Quick run command** | `go test ./internal/release -run 'Test(CanonicalReleaseMetadata|CanonicalReleaseAssets|CanonicalReleaseMatchesGoReleaserContract|PlanReleaseOrchestrationCreate|PlanReleaseOrchestrationReuse|ReleasePrepareSelectedChannelsReady|PackageManagerReleaseContract|NPMPackageReleaseContract|GitHubReleaseWorkflowReuseContract|GitHubReleaseDocsStayCanonical)$'` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~120 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/release -run 'Test(CanonicalReleaseMetadata|CanonicalReleaseAssets|CanonicalReleaseMatchesGoReleaserContract|PlanReleaseOrchestrationCreate|PlanReleaseOrchestrationReuse|ReleasePrepareSelectedChannelsReady|PackageManagerReleaseContract|NPMPackageReleaseContract|GitHubReleaseWorkflowReuseContract|GitHubReleaseDocsStayCanonical)$'`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 17-01-01 | 01 | 1 | PUB-01 | unit | `go test ./internal/release -run 'Test(CanonicalReleaseMetadata|CanonicalReleaseAssets|CanonicalReleaseRejectsInvalidVersion)$'` | planned creation in task | ⬜ pending |
| 17-01-02 | 01 | 1 | PUB-01 | contract | `go test ./internal/release -run 'Test(ArchiveMatrix|ChecksumManifest|CanonicalReleaseMatchesGoReleaserContract)$'` | depends on 17-01-01 output | ⬜ pending |
| 17-02-01 | 02 | 2 | PUB-01 | unit | `go test ./internal/release -run 'Test(PlanReleaseOrchestrationCreate|PlanReleaseOrchestrationReuse|PlanReleaseOrchestrationRejectsInvalidMode|PlanReleaseOrchestrationRejectsTagMismatch)$'` | planned creation in task | ⬜ pending |
| 17-02-02 | 02 | 2 | PUB-01 | unit | `go test ./internal/release -run 'Test(PlanReleaseOrchestrationCreate|PlanReleaseOrchestrationReuse|PlanReleaseOrchestrationRejectsTagMismatch|ReleasePrepareSelectedChannelsReady|ReleaseSelectedChannelsDoNotInheritUnselectedBlockers)$'` | depends on 17-02-01 output | ⬜ pending |
| 17-03-01 | 03 | 2 | PUB-01 | unit | `go test ./internal/release -run 'Test(NPMPackageReleaseContract|PackageManagerReleaseContract|RenderHomebrewFormula|RenderScoopManifest)$'` | ✅ existing plus 17-01 output | ⬜ pending |
| 17-03-02 | 03 | 2 | PUB-01 | contract | `go test ./internal/release -run 'Test(NPMPublishConfig|NPMPublishWorkflow|CanonicalReleaseFeedsDownstreamConsumers)$'` | ✅ existing | ⬜ pending |
| 17-04-01 | 04 | 3 | PUB-01 | contract | `go test ./internal/release -run 'Test(GitHubReleasePublicationConfig|GitHubReleaseWorkflowReuseContract|CanonicalReleaseMatchesGoReleaserContract|PlanReleaseOrchestrationCreate|PlanReleaseOrchestrationReuse)$'` | ✅ existing plus 17-02 output | ⬜ pending |
| 17-04-02 | 04 | 3 | PUB-01 | contract | `go test ./internal/release -run 'Test(ReleaseChecklistPublicationCredentials|GitHubReleaseDocsStayCanonical|GitHubReleaseWorkflowReuseContract)$'` | ✅ existing plus 17-02/17-03 output | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] No separate Wave 0 bootstrap is required for Phase 17.
- [x] Missing files are created directly by `17-01-01` and `17-02-01`, each with an explicit automated verify command.
- [x] Later tasks depend on those plan outputs explicitly instead of relying on an unstated pre-phase scaffold.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Canonical tagged release remains the root source of truth | PUB-01 | Unit tests can prove derivation logic, but a human check confirms the operator story stays understandable across GitHub Release and downstream consumers | Inspect the release workflow and docs, then confirm the same tag and archive/checksum contract are described consistently for GitHub Release and downstream consumers. |
| Rerun path reuses an existing tagged release rather than rebuilding unrelated assets | PUB-01 | The roadmap requires rerun semantics, and a human review of the contract is the fastest way to catch hidden duplication | Review the final workflow and docs wording around `release_tag`, then confirm an existing tagged release can be reused as the root artifact set for downstream publication. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
