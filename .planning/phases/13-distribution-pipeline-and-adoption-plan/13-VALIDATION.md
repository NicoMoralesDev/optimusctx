---
phase: 13
slug: distribution-pipeline-and-adoption-plan
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-03-16
---

# Phase 13 - Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./internal/buildinfo ./internal/cli ./internal/app ./internal/release -run 'Test(BuildInfo|Version|Install|Snippet|Doctor|Release|PackageManager|Distribution)'` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~60 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/buildinfo ./internal/cli ./internal/app ./internal/release -run 'Test(BuildInfo|Version|Install|Snippet|Doctor|Release|PackageManager|Distribution)'`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 75 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 13-01-01 | 01 | 1 | DIST-01 | unit | `go test ./internal/buildinfo ./internal/cli -run 'TestBuildInfoSummary|TestVersionCommand'` | ✅ | ⬜ pending |
| 13-01-02 | 01 | 1 | DIST-01 | integration | `go test ./internal/release -run 'TestArchiveMatrix|TestChecksumManifest'` | ❌ W0 | ⬜ pending |
| 13-01-03 | 01 | 1 | DIST-01 | integration | `go test ./internal/release ./internal/cli -run 'TestGitHubReleasePublicationConfig|TestReleaseMetadataInjection'` | ❌ W0 | ⬜ pending |
| 13-02-01 | 02 | 2 | DIST-02 | unit | `go test ./internal/release -run 'TestHomebrewFormulaRendering'` | ❌ W0 | ⬜ pending |
| 13-02-02 | 02 | 2 | DIST-02 | unit | `go test ./internal/release -run 'TestScoopManifestRendering'` | ❌ W0 | ⬜ pending |
| 13-02-03 | 02 | 2 | DIST-02 | integration | `go test ./internal/release ./internal/cli -run 'TestPackageManagerPublicationConfig|TestPackageManagerInstallDocs'` | ❌ W0 | ⬜ pending |
| 13-03-01 | 03 | 3 | DIST-03 | integration | `go test ./internal/cli ./internal/app -run 'TestInstallPreview|TestSnippetRender'` | ✅ | ⬜ pending |
| 13-03-02 | 03 | 3 | DIST-03 | integration | `go test ./internal/cli ./internal/app -run 'TestDoctorSmokeVerification|TestArchiveInstallSmoke'` | ❌ W0 | ⬜ pending |
| 13-03-03 | 03 | 3 | DIST-03 | integration | `go test ./internal/cli ./internal/app ./internal/store/sqlite -run 'TestDistributionSmokeFlow|TestReleaseVerificationCommands'` | ❌ W0 | ⬜ pending |
| 13-04-01 | 04 | 3 | DIST-04 | unit | `go test ./internal/release -run 'TestDistributionChannelPolicy'` | ❌ W0 | ⬜ pending |
| 13-04-02 | 04 | 3 | DIST-04 | integration | `go test ./internal/release ./internal/cli -run 'TestRolloutPlanExamples|TestUpgradePolicy'` | ❌ W0 | ⬜ pending |
| 13-04-03 | 04 | 3 | DIST-04 | integration/doc | `go test ./...` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/release/release_test.go` - archive matrix, checksum manifest, and workflow contract coverage
- [ ] `internal/release/package_manager_test.go` - Homebrew and Scoop rendering plus publication-path coverage
- [ ] `internal/release/distribution_plan_test.go` - release-channel policy and upgrade-path coverage
- [ ] `docs/install-and-verify.md` - canonical operator flow using archive or package-manager installs
- [ ] `docs/distribution-strategy.md` - rollout, support, and deferred-scope plan

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Final release artifacts install cleanly on real macOS/Linux and Windows environments | DIST-01, DIST-02 | CI and unit tests can validate archive names and manifest rendering, but the final confidence check still needs actual package-manager and shell behavior on target platforms | Install from one produced archive, one Homebrew path, and one Scoop path on clean environments; then run `optimusctx version`, `optimusctx doctor`, and `optimusctx snippet`. |
| Install docs are understandable to a new operator without repo-specific tribal knowledge | DIST-03 | Documentation clarity and first-run comprehension are difficult to prove with command-level assertions alone | Follow `docs/install-and-verify.md` verbatim on a clean machine and confirm the expected commands and outputs are sufficient without reading source code. |
| Rollout and support guidance is truthful about channel ownership and deferred scope | DIST-04 | Strategy quality depends on explicit human judgment about promises and omissions | Review `docs/distribution-strategy.md` and confirm it names release channels, target users, upgrade expectations, support assumptions, and v2 deferrals without claiming unsupported channels. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 75s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
