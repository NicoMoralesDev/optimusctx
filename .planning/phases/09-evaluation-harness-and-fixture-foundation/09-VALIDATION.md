---
phase: 9
slug: evaluation-harness-and-fixture-foundation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-15
---

# Phase 9 - Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./internal/repository ./internal/app ./internal/cli ./internal/state ./internal/store/sqlite` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~45 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/repository ./internal/app ./internal/cli ./internal/state ./internal/store/sqlite`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 45 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 09-01-01 | 01 | 1 | EVAL-04 | unit | `go test ./internal/repository` | ✅ | ⬜ pending |
| 09-01-02 | 01 | 1 | EVAL-04 | unit/integration | `go test ./internal/repository` | ✅ | ⬜ pending |
| 09-01-03 | 01 | 1 | EVAL-01 | unit | `go test ./internal/repository` | ✅ | ⬜ pending |
| 09-02-01 | 02 | 2 | EVAL-01 | unit | `go test ./internal/app ./internal/cli` | ✅ | ⬜ pending |
| 09-02-02 | 02 | 2 | EVAL-01 | integration | `go test ./internal/app ./internal/cli` | ✅ | ⬜ pending |
| 09-02-03 | 02 | 2 | EVAL-01 | unit/integration | `go test ./internal/app ./internal/cli` | ✅ | ⬜ pending |
| 09-03-01 | 03 | 2 | EVAL-04 | unit | `go test ./internal/state ./internal/store/sqlite` | ✅ | ⬜ pending |
| 09-03-02 | 03 | 2 | EVAL-04 | unit/integration | `go test ./internal/state ./internal/store/sqlite` | ✅ | ⬜ pending |
| 09-03-03 | 03 | 2 | EVAL-04 | unit | `go test ./internal/state ./internal/store/sqlite` | ✅ | ⬜ pending |
| 09-04-01 | 04 | 3 | EVAL-01 | integration | `go test ./internal/app ./internal/cli` | ✅ | ⬜ pending |
| 09-04-02 | 04 | 3 | EVAL-04 | integration | `go test ./internal/app ./internal/cli ./internal/state ./internal/store/sqlite` | ✅ | ⬜ pending |
| 09-04-03 | 04 | 3 | EVAL-01, EVAL-04 | integration/doc | `go test ./...` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| README eval workflow reads truthfully against the shipped command surface | EVAL-01, EVAL-04 | The final operator-facing wording is easiest to confirm in rendered markdown, not with unit assertions alone | Review the Phase 9 README diff and confirm the documented `eval` flow uses real commands, names versioned fixture inputs, and does not imply hidden manual setup. |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 45s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
