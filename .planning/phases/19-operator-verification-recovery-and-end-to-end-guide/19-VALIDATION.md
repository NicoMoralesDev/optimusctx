---
phase: 19
slug: operator-verification-recovery-and-end-to-end-guide
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-18
---

# Phase 19 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go test (`go 1.26.1`) |
| **Config file** | none |
| **Quick run command** | `go test ./internal/release ./internal/cli -run 'Test(ReleasePrepareSelectedChannelsReady|ReleaseSelectedChannelsDoNotInheritUnselectedBlockers|ReleasePrepareAllChannelsReady|ReleasePrepareHomebrewAndScoopAutomationMarkers|PlanReleasePublicationFanout|PlanReleasePublicationRerun|GitHubReleaseWorkflowReuseContract|ChannelPublicationWorkflowSelectiveRerun|HomebrewPublishWorkflow|ScoopPublishWorkflow|MultiChannelPublicationDocsStayCanonical|RolloutPlanExamples|UpgradePolicy)$'` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~60 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/release ./internal/cli -run 'Test(ReleasePrepareSelectedChannelsReady|ReleaseSelectedChannelsDoNotInheritUnselectedBlockers|ReleasePrepareAllChannelsReady|ReleasePrepareHomebrewAndScoopAutomationMarkers|PlanReleasePublicationFanout|PlanReleasePublicationRerun|GitHubReleaseWorkflowReuseContract|ChannelPublicationWorkflowSelectiveRerun|HomebrewPublishWorkflow|ScoopPublishWorkflow|MultiChannelPublicationDocsStayCanonical|RolloutPlanExamples|UpgradePolicy)$'`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 19-01-01 | 01 | 1 | OPS-06 | integration/doc | `go test ./internal/release -run 'Test(ReleaseWorkflowSummaryShowsChannelStatus|ReleaseWorkflowSummaryShowsFailureGuidance)$'` | ❌ W0 | ⬜ pending |
| 19-02-01 | 02 | 2 | OPS-07 | doc | `go test ./internal/release ./internal/cli -run 'Test(OperatorReleaseGuideStaysCanonical|InstallAndVerifyGuideLinksOperatorFlow)$'` | ❌ W0 | ⬜ pending |
| 19-03-01 | 03 | 2 | OPS-08 | doc/policy | `go test ./internal/release -run 'Test(OperatorRecoveryGuideStaysCanonical|DistributionDocsStayWithinSupportedScope|ChannelPublicationWorkflowSelectiveRerun)$'` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/release/release_test.go` — add Phase 19 workflow-summary and operator-guide contract tests for `OPS-06` and `OPS-07`
- [ ] `internal/release/distribution_plan_test.go` — add rollback/support-path assertions specific to `OPS-08`
- [ ] `docs/operator-release-guide.md` — canonical guide file with verification, rerun, and rollback flow

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Real GitHub Actions run summary renders the expected per-channel status and next-step block in the hosted UI | OPS-06 | Local tests can lock summary strings and workflow structure, but the rendered GitHub Actions summary UI still needs one real-run sanity check | Trigger the release workflow against a safe tag or inspected run, open the workflow run summary, and confirm npm, Homebrew, and Scoop each show channel, tag, outcome, and rerun guidance. |
| Operator documentation is usable end to end by a human following the steps without repo-internal context | OPS-07, OPS-08 | Contract tests can lock wording, but only a real read-through validates that the steps are navigable and ordered correctly | Follow `docs/operator-release-guide.md` from `release prepare` through verification and then simulate the rerun/rollback decision tree to confirm there are no ambiguous branches. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
