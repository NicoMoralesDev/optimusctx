---
phase: 15
slug: add-npm-and-npx-distribution-option
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-17
---

# Phase 15 - Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test + npm pack |
| **Config file** | `.goreleaser.yml`, `.github/workflows/release.yml`, `packaging/npm/package.json` |
| **Quick run command** | `go test ./internal/release ./internal/cli -run 'Test(NPMPackage|PackageManager|Distribution|ReleaseVerificationCommands)'` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~75 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/release ./internal/cli -run 'Test(NPMPackage|PackageManager|Distribution|ReleaseVerificationCommands)'`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 90 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 15-01-01 | 01 | 1 | DIST-02 | unit | `go test ./internal/release -run 'Test(NPMPackageMetadata|NPMPackageRendering|NPMPackageArchiveSelection)'` | ✅ | ✅ green |
| 15-01-02 | 01 | 1 | DIST-02 | unit | `go test ./internal/release -run 'Test(NPMPackageChecksums|NPMPackageCommands)'` | ✅ | ✅ green |
| 15-02-01 | 02 | 2 | DIST-02 | integration | `go test ./internal/release -run 'Test(NPMInstaller|NPMLauncher|NPMSupportedPlatforms)'` | ✅ | ✅ green |
| 15-02-02 | 02 | 2 | DIST-03 | integration/doc | `go test ./internal/cli ./internal/release -run 'Test(NPMInstallGuide|ReleaseVerificationCommands|PackageManagerInstallDocs)'` | ✅ | ✅ green |
| 15-03-01 | 03 | 2 | DIST-02 | integration | `go test ./internal/release -run 'Test(NPMPublishWorkflow|NPMPublishConfig)'` | ✅ | ✅ green |
| 15-03-02 | 03 | 2 | DIST-04 | integration/doc | `go test ./internal/release ./internal/cli -run 'Test(DistributionChannelPolicy|DistributionDocsStayWithinSupportedScope|PackageManagerInstallDocs|ReleaseVerificationCommands)'` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/release/npm_package_test.go` - package metadata, archive selection, checksum, and publish-workflow coverage
- [x] `packaging/npm/package.json` - publishable npm package manifest with `bin` and `postinstall`
- [x] `packaging/npm/bin/optimusctx.js` - launcher script for the real binary
- [x] `packaging/npm/lib/install.js` - package-local binary acquisition and verification logic

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| npm global install resolves to the tagged Go binary and the real CLI surface works | DIST-02, DIST-03 | Registry publication and shell/path behavior are hard to prove from repository-only tests | On a clean machine run `npm install -g @niccrow/optimusctx`, then run `optimusctx version`, `optimusctx doctor`, and `optimusctx snippet`. |
| `npx` ephemeral execution resolves the package bin correctly | DIST-02, DIST-03 | `npx` execution path depends on npm client behavior outside the Go test harness | On a clean machine run `npx @niccrow/optimusctx version` and `npx @niccrow/optimusctx doctor`, confirm the wrapper downloads or reuses the tagged binary and prints the expected outputs. |
| Support and rollback guidance stay truthful after npm is added | DIST-04 | Human judgment is required to confirm the policy still describes supported channels and fallbacks honestly | Review `README.md`, `docs/distribution-strategy.md`, and `docs/release-checklist.md` and confirm they still name GitHub Release archives as fallback while adding npm without claiming broader installer scope. |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 90s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** auto-approved after retroactive audit on 2026-03-17

## Validation Audit 2026-03-17

| Metric | Count |
|--------|-------|
| Gaps found | 6 |
| Resolved | 6 |
| Escalated | 0 |

Retroactive audit confirmed the existing Phase 15 automated coverage without adding new tests.

Executed evidence:

- `go test ./internal/release ./internal/cli -run 'Test(NPMPackageMetadata|NPMPackageRendering|NPMPackageArchiveSelection|NPMPackageChecksums|NPMPackageCommands|NPMInstaller|NPMLauncher|NPMSupportedPlatforms|NPMInstallGuide|ReleaseVerificationCommands|PackageManagerInstallDocs|NPMPublishWorkflow|NPMPublishConfig|GitHubReleasePublicationConfig|DistributionChannelPolicy|DistributionDocsStayWithinSupportedScope|ArchiveMatrix|ChecksumManifest)'`
