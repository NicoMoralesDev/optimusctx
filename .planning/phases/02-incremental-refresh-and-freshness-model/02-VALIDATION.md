---
phase: 02
slug: incremental-refresh-and-freshness-model
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-14
---

# Phase 02 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./... -run 'TestRefresh|TestDiff|TestFingerprint|TestFreshness'` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -run 'TestRefresh|TestDiff|TestFingerprint|TestFreshness'`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 02-xx-01 | TBD | TBD | REFR-01 | unit/integration | `go test ./... -run 'TestRefreshHashes|TestNoOpRefresh'` | ❌ W0 | ⬜ pending |
| 02-xx-02 | TBD | TBD | REFR-02 | unit/integration | `go test ./... -run 'TestSubtreeFingerprint|TestFingerprintPropagation'` | ❌ W0 | ⬜ pending |
| 02-xx-03 | TBD | TBD | REFR-03 | unit/integration | `go test ./... -run 'TestRefreshDiff|TestMoveDetection'` | ❌ W0 | ⬜ pending |
| 02-xx-04 | TBD | TBD | REFR-04 | integration | `go test ./... -run 'TestIncrementalRefresh|TestSnapshotEquivalence'` | ❌ W0 | ⬜ pending |
| 02-xx-05 | TBD | TBD | REFR-05 | integration | `go test ./... -run 'TestFreshnessState|TestDegradedRefreshState'` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/refresh/diff_test.go` — diff classification coverage for add/change/delete/move/ignore transitions
- [ ] `internal/refresh/fingerprint_test.go` — bottom-up subtree fingerprint propagation tests
- [ ] `internal/app/refresh_integration_test.go` or equivalent — fixture progression and snapshot-equivalence coverage
- [ ] `internal/store/sqlite/refresh_store_test.go` or equivalent — transactional refresh application and degraded-state assertions

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Medium-repo no-op refresh is observably cheaper than a full rebuild | REFR-04 | A meaningful “cheap” check needs a human sanity pass on runtime and touched rows, not just functional assertions | Run refresh twice on a medium fixture or real repo and compare logged hash counts, touched rows, and elapsed time |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
