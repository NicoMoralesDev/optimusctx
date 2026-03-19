---
phase: 18
slug: multi-channel-publication-fan-out
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-18
---

# Phase 18 - Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | `go test` |
| **Config file** | `.github/workflows/release.yml`, `docs/release-checklist.md`, `packaging/homebrew/optimusctx.rb.tmpl`, `packaging/scoop/optimusctx.json.tmpl` |
| **Quick run command** | `go test ./internal/release ./internal/cli -run 'Test(PlanReleaseOrchestrationCreate|PlanReleaseOrchestrationReuse|PackageManagerReleaseContract|RenderHomebrewFormula|RenderScoopManifest|NPMPublishWorkflow|NPMPublishConfig|GitHubReleaseWorkflowReuseContract|ReleasePrepareSelectedChannelsReady)$'` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~120 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/release ./internal/cli -run 'Test(PlanReleaseOrchestrationCreate|PlanReleaseOrchestrationReuse|PackageManagerReleaseContract|RenderHomebrewFormula|RenderScoopManifest|NPMPublishWorkflow|NPMPublishConfig|GitHubReleaseWorkflowReuseContract|ReleasePrepareSelectedChannelsReady)$'`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 18-01-01 | 01 | 1 | PUB-02 | unit | `go test ./internal/release -run 'Test(PlanReleaseOrchestrationCreate|TestPlanReleaseOrchestrationReuse)$'` | ✅ existing contract baseline | ⬜ pending |
| 18-01-02 | 01 | 1 | PUB-03 | unit | `go test ./internal/release -run 'TestPlanReleaseOrchestration.*|TestReleasePrepareSelectedChannelsReady'` | ✅ existing contract baseline | ⬜ pending |
| 18-02-01 | 02 | 2 | PUB-02 | contract | `go test ./internal/release -run 'Test(RenderHomebrewFormula|RenderScoopManifest|PackageManagerReleaseContract)$'` | ✅ existing render baseline | ⬜ pending |
| 18-03-01 | 03 | 2 | PUB-02 | contract | `go test ./internal/release -run 'Test(NPMPublishWorkflow|NPMPublishConfig|GitHubReleaseWorkflowReuseContract)$'` | ✅ existing workflow baseline | ⬜ pending |
| 18-03-02 | 03 | 2 | PUB-03 | contract | `go test ./internal/release ./internal/cli -run 'Test(ReleasePrepareSelectedChannelsReady|GitHubReleaseWorkflowReuseContract)'` | depends on plan outputs | ⬜ pending |
| 18-04-01 | 04 | 3 | PUB-02 | regression | `go test ./...` | depends on all prior outputs | ⬜ pending |
| 18-04-02 | 04 | 3 | PUB-03 | regression | `go test ./...` | depends on all prior outputs | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] No separate Wave 0 bootstrap is required for Phase 18.
- [x] Canonical release metadata, npm render infrastructure, and Homebrew/Scoop templates already exist from Phases 15 and 17.
- [x] Phase 18 work can build directly on the existing release contract and test harness.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Single-channel rerun publishes only the requested downstream channel against an existing tag | PUB-03 | Repository tests can prove selection and contract wiring, but not an end-to-end GitHub Actions rerun with real secrets and external repos | Dispatch the release workflow against a real tagged release with one selected downstream channel, then confirm unrelated channel outputs are untouched. |
| Homebrew and Scoop publication update the correct external repositories with the rendered payloads | PUB-02 | Real tap and bucket publication requires external repository pushes and secrets that local tests cannot perform | Run one real release against a staging or real tag, inspect the target tap/bucket commit, and confirm the rendered version, URLs, and checksums match the canonical release tag. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
