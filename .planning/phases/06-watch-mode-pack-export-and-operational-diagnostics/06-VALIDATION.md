---
phase: 06
slug: watch-mode-pack-export-and-operational-diagnostics
status: approved
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-15
---

# Phase 06 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./... -run 'TestWatch|TestPackExport|TestDoctor'` |
| **Full suite command** | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./...` |
| **Estimated runtime** | ~20 seconds |

---

## Sampling Rate

- **After every task commit:** Run `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./... -run 'TestWatch|TestPackExport|TestDoctor'`
- **After every plan wave:** Run `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | OPS-01 | unit/integration | `... -run 'TestWatch'` | ✅ planned | ⬜ pending |
| 06-02-01 | 02 | 2 | OPS-02 | integration | `... -run 'TestWatch'` | ✅ planned | ⬜ pending |
| 06-03-01 | 03 | 3 | OPS-03 | unit/integration | `... -run 'TestPackExport'` | ✅ planned | ⬜ pending |
| 06-04-01 | 04 | 4 | OPS-04 | integration | `... -run 'TestPackExport'` | ✅ planned | ⬜ pending |
| 06-05-01 | 05 | 5 | CLI-05 | unit/integration | `... -run 'TestDoctor'` | ✅ planned | ⬜ pending |
| 06-05-02 | 05 | 5 | OPS-05 | integration | `... -run 'TestDoctor'` | ✅ planned | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/cli/watch_test.go` — watch command boundary coverage for OPS-01 and OPS-02
- [x] `internal/app/watch_test.go` — watcher runtime and refresh integration coverage
- [x] `internal/app/pack_export_test.go` — export determinism and budget-fitting coverage for OPS-03 and OPS-04
- [x] `internal/cli/doctor_test.go` — doctor command rendering coverage for CLI-05 and OPS-05

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Long-lived `watch run` operator UX remains understandable while attached to a terminal or shell backgrounding flow | OPS-01 | Process lifecycle and operator expectations are hard to fully model in automated tests | Start `optimusctx watch run` in a temp repo, mutate files, confirm it stays attached cleanly, reports status, and can be stopped without corrupting state |
| `optimusctx doctor` output remains actionable and readable as a human diagnostic entrypoint | CLI-05 | Automated tests can assert sections and values but not real operator clarity | Run `optimusctx doctor` against healthy and degraded repos; confirm the report highlights root causes and next actions without requiring database inspection |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-03-15
