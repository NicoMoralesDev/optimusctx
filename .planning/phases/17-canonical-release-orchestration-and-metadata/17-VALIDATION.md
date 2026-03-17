---
phase: 17
slug: canonical-release-orchestration-and-metadata
status: draft
nyquist_compliant: true
wave_0_complete: false
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
| **Quick run command** | `go test ./internal/release ./internal/cli -run 'Test(ArchiveMatrix|ChecksumManifest|GitHubReleasePublicationConfig|NPMPublishWorkflow|NPMPublishConfig|ReleasePrerequisiteChecks|ReleasePrepareSelectedChannelsReady)'` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~120 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/release ./internal/cli -run 'Test(ArchiveMatrix|ChecksumManifest|GitHubReleasePublicationConfig|NPMPublishWorkflow|NPMPublishConfig|ReleasePrerequisiteChecks|ReleasePrepareSelectedChannelsReady)'`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 17-01-01 | 01 | 1 | PUB-01 | unit | `go test ./internal/release -run 'Test(ArchiveMatrix|ChecksumManifest)'` | ❌ W0 | ⬜ pending |
| 17-01-02 | 01 | 1 | PUB-01 | unit | `go test ./internal/release -run 'TestGitHubReleasePublicationConfig'` | ❌ W0 | ⬜ pending |
| 17-02-01 | 02 | 2 | PUB-01 | unit | `go test ./internal/release -run 'TestReleasePrerequisiteChecks|TestReleasePrepareSelectedChannelsReady'` | ❌ W0 | ⬜ pending |
| 17-02-02 | 02 | 2 | PUB-01 | integration | `go test ./internal/cli -run 'TestReleasePrepareCommand|TestReleasePrepareConfirmGate|TestReleasePrepareSelectedChannelsReady'` | ❌ W0 | ⬜ pending |
| 17-03-01 | 03 | 2 | PUB-01 | unit | `go test ./internal/release -run 'TestNPMPublishWorkflow|TestNPMPublishConfig'` | ❌ W0 | ⬜ pending |
| 17-03-02 | 03 | 2 | PUB-01 | unit | `go test ./internal/release -run 'Test(ArchiveMatrix|ChecksumManifest|NPMPublishWorkflow|NPMPublishConfig)'` | ❌ W0 | ⬜ pending |
| 17-04-01 | 04 | 3 | PUB-01 | integration | `go test ./internal/release ./internal/cli -run 'Test(ArchiveMatrix|ChecksumManifest|GitHubReleasePublicationConfig|NPMPublishWorkflow|NPMPublishConfig|ReleasePrepareSelectedChannelsReady)'` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/release/release_test.go` - canonical archive, checksum, workflow, and npm publish contract tests
- [ ] `internal/release/package_manager_test.go` - deterministic package-manager asset derivation coverage
- [ ] `internal/release/npm_package_test.go` - npm wrapper metadata derivation coverage
- [ ] `internal/release/prepare_test.go` - selected-channel prepare contract coverage carried forward from Phase 16
- [ ] `internal/cli/release_test.go` - operator-facing prepare command contract coverage

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Canonical tagged release remains the root source of truth | PUB-01 | Unit tests can prove derivation logic, but a human check confirms the operator story stays understandable across GitHub Release and downstream consumers | Inspect the release workflow and docs, then confirm the same tag and archive/checksum contract are described consistently for GitHub Release and downstream consumers. |
| Rerun path reuses an existing tagged release rather than rebuilding unrelated assets | PUB-01 | The roadmap requires rerun semantics, and a human review of the contract is the fastest way to catch hidden duplication | Review the final orchestration docs or CLI output and confirm an existing tagged release can be reused as the root artifact set for downstream publication. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
