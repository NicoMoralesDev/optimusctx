---
phase: 16
slug: release-versioning-and-preflight-guardrails
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-03-17
---

# Phase 16 - Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | `go test` |
| **Config file** | `.github/workflows/release.yml`, `.planning/ROADMAP.md`, `.planning/REQUIREMENTS.md` |
| **Quick run command** | `go test ./internal/release ./internal/cli -run 'Test(ReleasePreparation|ReleaseVersionProposal|ReleaseTagNormalization|ReleaseSemanticTagConflicts|ReleasePreflight|ReleasePrepareCommand|ReleasePrepareHelp)'` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~90 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/release ./internal/cli -run 'Test(ReleasePreparation|ReleaseVersionProposal|ReleaseTagNormalization|ReleaseSemanticTagConflicts|ReleasePreflight|ReleasePrepareCommand|ReleasePrepareHelp)'`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 90 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 16-01-01 | 01 | 1 | REL-01 | unit | `go test ./internal/release -run 'Test(ReleaseVersionProposal|ReleaseTagNormalization)'` | ❌ W0 | ⬜ pending |
| 16-01-02 | 01 | 1 | REL-01 | unit | `go test ./internal/release -run 'Test(ReleasePreparation|ReleaseSemanticTagConflicts)'` | ❌ W0 | ⬜ pending |
| 16-02-01 | 02 | 2 | REL-02 | unit | `go test ./internal/release -run 'Test(ReleasePreflight|ReleaseWorktreeBlockers|ReleaseRemoteTagConflicts)'` | ❌ W0 | ⬜ pending |
| 16-02-02 | 02 | 2 | REL-02, REL-03 | unit | `go test ./internal/release -run 'Test(ReleasePrerequisiteChecks|ReleasePlanJSON)'` | ❌ W0 | ⬜ pending |
| 16-03-01 | 03 | 2 | REL-01, REL-03 | integration | `go test ./internal/cli -run 'TestReleasePrepareCommand'` | ❌ W0 | ⬜ pending |
| 16-03-02 | 03 | 2 | REL-03 | integration | `go test ./internal/cli -run 'TestReleasePrepareHelp|TestReleasePrepareConfirmGate'` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/release/prepare_test.go` - version proposal, tag normalization, semantic-conflict, and preflight tests
- [ ] `internal/cli/release_test.go` - command parsing, JSON output, and confirm-gate coverage
- [ ] `internal/release/prepare.go` - shared release-preparation model and preflight logic
- [ ] `internal/cli/release.go` - operator-facing `release prepare` command

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Human review summary is understandable before publication | REL-03 | Unit tests can assert strings, but not whether the review flow is actually clear to an operator | Run `optimusctx release prepare` in a clean repo, inspect the summary, and confirm it clearly shows version, normalized tag, selected channels, blockers, and next step. |
| Dirty worktree blocks the flow before mutation | REL-02 | A real dirty-worktree run validates behavior against actual git state rather than only stubs | Create an uncommitted file, run `optimusctx release prepare`, and confirm it exits non-zero without creating a tag or starting publication. |
| Semantic tag alias detection is explicit | REL-02 | Local tests can stub tag sets, but a real repo check confirms the operator messaging is understandable | Create or simulate a legacy-style tag such as `v1.1`, run `optimusctx release prepare --version 1.1.0`, and confirm the command reports a semantic-equivalent tag conflict. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
