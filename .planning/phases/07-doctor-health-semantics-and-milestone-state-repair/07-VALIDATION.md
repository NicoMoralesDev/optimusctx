---
phase: 7
slug: doctor-health-semantics-and-milestone-state-repair
status: approved
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-15
---

# Phase 7 - Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./internal/app ./internal/cli` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/app ./internal/cli`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 07-01-01 | 01 | 1 | CLI-05 | unit/integration | `go test ./internal/app ./internal/cli` | ✅ | ✅ green |
| 07-01-02 | 01 | 1 | OPS-01 | unit/integration | `go test ./internal/app ./internal/cli` | ✅ | ✅ green |
| 07-01-03 | 01 | 1 | OPS-05 | unit/integration | `go test ./internal/app ./internal/cli` | ✅ | ✅ green |
| 07-02-01 | 02 | 2 | CLI-05 | doc/manual | `go test ./...` | ✅ | ✅ green |
| 07-02-02 | 02 | 2 | OPS-05 | doc/manual | `go test ./...` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Planning artifact consistency across `.planning/STATE.md`, `.planning/ROADMAP.md`, and `.planning/REQUIREMENTS.md` | CLI-05, OPS-05 | Document alignment is easier to review in diff form than through code assertions | Review the final diff and confirm all three files agree that Phase 6 executed successfully, Phase 7 owns the doctor/watch semantics repair plus state repair, and Phase 8 remains the verification backfill phase. |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-03-15
