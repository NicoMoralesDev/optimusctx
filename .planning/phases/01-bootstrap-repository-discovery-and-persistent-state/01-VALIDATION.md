---
phase: 1
slug: bootstrap-repository-discovery-and-persistent-state
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-14
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~20 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 1-01-01 | 01 | 1 | CLI-01 | integration | `go test ./...` | ❌ W0 | ⬜ pending |
| 1-01-02 | 01 | 1 | CLI-03 | integration | `go test ./...` | ❌ W0 | ⬜ pending |
| 1-01-03 | 01 | 1 | CLI-04 | integration | `go test ./...` | ❌ W0 | ⬜ pending |
| 1-02-01 | 02 | 1 | REPO-01 | unit | `go test ./...` | ❌ W0 | ⬜ pending |
| 1-02-02 | 02 | 1 | REPO-02 | unit | `go test ./...` | ❌ W0 | ⬜ pending |
| 1-02-03 | 02 | 1 | REPO-03 | integration | `go test ./...` | ❌ W0 | ⬜ pending |
| 1-03-01 | 03 | 2 | REPO-04 | integration | `go test ./...` | ❌ W0 | ⬜ pending |
| 1-03-02 | 03 | 2 | REPO-05 | integration | `go test ./...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `go.mod` — initialize module and toolchain baseline
- [ ] `cmd/optimusctx/main.go` — CLI entrypoint for command tests
- [ ] `internal/...` package skeleton — runtime services under test
- [ ] `*_test.go` files for CLI, repository discovery, and SQLite initialization — coverage stubs for phase requirements

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Printed snippet is clear and copyable | CLI-04 | Quality of snippet wording is easier to review manually than assert exhaustively | Run `optimusctx snippet` and confirm it prints only stdout content, no file writes, and no misleading MCP claims |
| Init output is understandable to a first-time operator | CLI-03 | Human-facing output quality matters beyond exit code assertions | Run `optimusctx init` in a fixture repo and confirm root path, state path, schema version, and file count are clearly reported |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
