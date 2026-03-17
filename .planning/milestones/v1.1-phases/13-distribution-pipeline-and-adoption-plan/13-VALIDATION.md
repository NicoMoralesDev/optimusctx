---
phase: 13
slug: distribution-pipeline-and-adoption-plan
status: complete
nyquist_compliant: true
wave_0_complete: true
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
| 13-01-01 | 01 | 1 | DIST-01 | unit | `go test ./internal/buildinfo ./internal/cli -run 'Test(BuildInfo|Version)'` | ✅ | ✅ green |
| 13-01-02 | 01 | 1 | DIST-01 | integration | `go test ./internal/release -run 'Test(ArchiveMatrix|ChecksumManifest|Release)'` | ✅ | ✅ green |
| 13-01-03 | 01 | 1 | DIST-01 | integration | `go test ./internal/release ./internal/cli -run 'Test(Release|Version)'` | ✅ | ✅ green |
| 13-02-01 | 02 | 2 | DIST-02 | unit | `go test ./internal/release -run 'Test(PackageManager|Distribution)'` | ✅ | ✅ green |
| 13-02-02 | 02 | 2 | DIST-02 | unit | `go test ./internal/release -run 'Test(PackageManager|Distribution)'` | ✅ | ✅ green |
| 13-02-03 | 02 | 2 | DIST-02 | integration | `go test ./internal/release ./internal/cli -run 'Test(PackageManager|Install|Distribution)'` | ✅ | ✅ green |
| 13-03-01 | 03 | 3 | DIST-03 | integration | `go test ./internal/cli ./internal/app -run 'Test(Install|Snippet|Doctor)'` | ✅ | ✅ green |
| 13-03-02 | 03 | 3 | DIST-03 | integration | `go test ./internal/cli ./internal/app -run 'Test(Install|Snippet|Doctor)'` | ✅ | ✅ green |
| 13-03-03 | 03 | 3 | DIST-03 | integration | `go test ./internal/cli ./internal/app ./internal/store/sqlite -run 'Test(Install|Snippet|Doctor|ReleaseVerificationCommands)'` | ✅ | ✅ green |
| 13-04-01 | 04 | 3 | DIST-04 | unit | `go test ./internal/release -run 'Test(Distribution|PackageManager)'` | ✅ | ✅ green |
| 13-04-02 | 04 | 3 | DIST-04 | integration | `go test ./internal/release ./internal/cli -run 'Test(Distribution|PackageManager|Install)'` | ✅ | ✅ green |
| 13-04-03 | 04 | 3 | DIST-04 | integration/doc | `go test ./...` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/release/release_test.go` - archive matrix, checksum manifest, and workflow contract coverage
- [x] `internal/release/package_manager_test.go` - Homebrew and Scoop rendering plus publication-path coverage
- [x] `internal/release/distribution_plan_test.go` - release-channel policy and upgrade-path coverage
- [x] `docs/install-and-verify.md` - canonical operator flow using archive or package-manager installs
- [x] `docs/distribution-strategy.md` - rollout, support, and deferred-scope plan

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Final release artifacts install cleanly on real macOS/Linux and Windows environments | DIST-01, DIST-02 | CI and unit tests can validate archive names and manifest rendering, but the final confidence check still needs actual package-manager and shell behavior on target platforms | Install from one produced archive, one Homebrew path, and one Scoop path on clean environments; then run `optimusctx version`, `optimusctx doctor`, and `optimusctx snippet`. |
| Install docs are understandable to a new operator without repo-specific tribal knowledge | DIST-03 | Documentation clarity and first-run comprehension are difficult to prove with command-level assertions alone | Follow `docs/install-and-verify.md` verbatim on a clean machine and confirm the expected commands and outputs are sufficient without reading source code. |
| Rollout and support guidance is truthful about channel ownership and deferred scope | DIST-04 | Strategy quality depends on explicit human judgment about promises and omissions | Review `docs/distribution-strategy.md` and confirm it names release channels, target users, upgrade expectations, support assumptions, and v2 deferrals without claiming unsupported channels. |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 75s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** auto-approved after retroactive audit on 2026-03-17

## Validation Audit 2026-03-17

| Metric | Count |
|--------|-------|
| Gaps found | 12 |
| Resolved | 12 |
| Escalated | 0 |

Retroactive audit confirmed the existing Phase 13 automated coverage without adding new tests.

Executed evidence:

- `go test ./internal/buildinfo ./internal/cli ./internal/app ./internal/release -run 'Test(BuildInfo|Version|Install|Snippet|Doctor|Release|PackageManager|Distribution)'`
