---
phase: 19
slug: operator-verification-recovery-and-end-to-end-guide
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-18
---

# Phase 19 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | `go test` plus shell/`rg` task-level file checks |
| **Config file** | `.github/workflows/release.yml`, `docs/operator-release-guide.md`, `docs/release-checklist.md`, `docs/install-and-verify.md`, `docs/distribution-strategy.md` |
| **Quick run command** | `go test ./internal/release -run 'Test(ReleaseWorkflowSummaryShowsChannelStatus|ReleaseWorkflowSummaryShowsFailureGuidance)$'`<br>`go test ./internal/release ./internal/cli -run 'Test(OperatorReleaseGuideStaysCanonical|InstallAndVerifyGuideLinksOperatorFlow)$'`<br>`go test ./internal/release -run 'Test(OperatorRecoveryGuideStaysCanonical|DistributionDocsStayWithinSupportedScope|ChannelPublicationWorkflowSelectiveRerun)$'` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~15-30 seconds per targeted quick command; ~60-90 seconds for `go test ./...` |

---

## Sampling Rate

- **After every task commit:** Run that task's `<automated>` command from the per-task map so verification remains runnable in scheduled order.
- **After 19-01 Task 2:** Run `go test ./internal/release -run 'Test(ReleaseWorkflowSummaryShowsChannelStatus|ReleaseWorkflowSummaryShowsFailureGuidance)$'`
- **After 19-02 Task 2:** Run `go test ./internal/release ./internal/cli -run 'Test(OperatorReleaseGuideStaysCanonical|InstallAndVerifyGuideLinksOperatorFlow)$'`
- **After 19-03 Task 2:** Run `go test ./internal/release -run 'Test(OperatorRecoveryGuideStaysCanonical|DistributionDocsStayWithinSupportedScope|ChannelPublicationWorkflowSelectiveRerun)$'`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 19-01-01 | 01 | 1 | OPS-06 | workflow file | `bash -lc "test -f .github/workflows/release.yml && rg -n --fixed-strings -- '### GitHub Release publication' .github/workflows/release.yml >/dev/null && rg -n --fixed-strings -- '- failure_reason:' .github/workflows/release.yml >/dev/null && rg -n --fixed-strings -- '- next_step:' .github/workflows/release.yml >/dev/null && rg -n 'GitHub Release archive publication failed\|publication_channel=npm\|publication_channel=homebrew\|publication_channel=scoop' .github/workflows/release.yml >/dev/null && rg -n 'GitHub Release remains the canonical root' .github/workflows/release.yml >/dev/null"` | ✅ existing workflow file | ⬜ pending |
| 19-01-02 | 01 | 1 | OPS-06 | workflow contract | `go test ./internal/release -run 'Test(ReleaseWorkflowSummaryShowsChannelStatus|ReleaseWorkflowSummaryShowsFailureGuidance)$'` | ✅ task extends existing test file | ⬜ pending |
| 19-02-01 | 02 | 2 | OPS-07 | doc file | `bash -lc "test -f docs/operator-release-guide.md && rg -n 'optimusctx release prepare|gh release view|gh release download|workflow_dispatch|release_tag|publication_channel|optimusctx version|optimusctx doctor|optimusctx snippet' docs/operator-release-guide.md >/dev/null && rg -n 'GitHub Release remains the canonical root and rollback source' docs/operator-release-guide.md >/dev/null"` | ✅ task creates required guide | ⬜ pending |
| 19-02-02 | 02 | 2 | OPS-07, OPS-08 | doc linkage contract | `go test ./internal/release ./internal/cli -run 'Test(OperatorReleaseGuideStaysCanonical|InstallAndVerifyGuideLinksOperatorFlow)$'` | ✅ task extends existing test files | ⬜ pending |
| 19-03-01 | 03 | 3 | OPS-08 | policy doc | `bash -lc "test -f docs/distribution-strategy.md && rg -n 'operator-release-guide.md|gh workflow run release.yml|release_tag=|publication_channel=npm|publication_channel=homebrew|publication_channel=scoop|prior tagged GitHub Release archive|publish a new fixed version' docs/distribution-strategy.md >/dev/null && ! rg -n 'npm unpublish|winget install|choco install|apt install|dnf install|yum install' docs/distribution-strategy.md >/dev/null"` | ✅ existing policy doc | ⬜ pending |
| 19-03-02 | 03 | 3 | OPS-08 | policy contract | `go test ./internal/release -run 'Test(OperatorRecoveryGuideStaysCanonical|DistributionDocsStayWithinSupportedScope|ChannelPublicationWorkflowSelectiveRerun)$'` | ✅ task extends existing test file | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] No separate Wave 0 bootstrap is required for Phase 19 after revision.
- [x] Each Task 1 now has a self-contained shell/`rg` verification command that can run immediately after its file edit.
- [x] Each Task 2 now owns the Phase 19 Go tests it introduces, so Nyquist verification is runnable in execution order.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Real GitHub Actions run summary renders the expected per-channel status and next-step block in the hosted UI | OPS-06 | Local tests can lock summary strings and workflow structure, but the rendered GitHub Actions summary UI still needs one real-run sanity check | Trigger the release workflow against a safe tag or inspected run, open the workflow run summary, and confirm GitHub Release, npm, Homebrew, and Scoop each show channel, tag, outcome, failure reason, and next-step guidance. |
| Operator documentation is usable end to end by a human following the steps without repo-internal context | OPS-07, OPS-08 | Contract tests can lock wording, but only a real read-through validates that the steps are navigable and ordered correctly | Follow `docs/operator-release-guide.md` from `release prepare` through verification and then simulate the rerun/rollback decision tree to confirm there are no ambiguous branches. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
