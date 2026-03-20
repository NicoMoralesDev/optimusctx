---
phase: 21
slug: real-write-paths-and-operator-surface-integration
status: ready
nyquist_compliant: true
wave_0_complete: false
created: 2026-03-19
---

# Phase 21 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | `go test` |
| **Config file** | none |
| **Quick run command** | `go test ./internal/app ./internal/cli ./internal/repository` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/app ./internal/cli ./internal/repository`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 21-01-01 | 01 | 1 | CLD-02 | unit/service | `go test ./internal/repository ./internal/app -run 'Test(RenderClaudeCLIAddCommand|InstallServiceSupportsClaudeCLIPreview)'` | ✅ | ⬜ pending |
| 21-01-02 | 01 | 1 | CLD-03, MCP-03 | service | `go test ./internal/app -run 'TestInstallServiceClaudeCLI'` | ❌ W0 | ⬜ pending |
| 21-02-01 | 02 | 2 | CDX-01, CDX-02 | service | `go test ./internal/repository ./internal/app -run 'Test(MergeCodexConfig|InstallService.*Codex(App|CLI))'` | ✅ | ⬜ pending |
| 21-02-02 | 02 | 2 | MCP-03 | service/integration | `go test ./internal/app ./internal/cli -run 'Test(InstallService.*Codex|StatusCommand.*Codex)'` | ❌ W0 | ⬜ pending |
| 21-03-01 | 03 | 3 | OPS-01 | CLI/unit | `go test ./internal/cli ./internal/app -run 'Test(Init|Status|Snippet|Doctor)'` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/app/install_test.go` — add fake-exec Claude CLI write tests for success, missing `claude`, non-zero exit, and scope selection.
- [ ] `internal/app/install_test.go` — add Codex App and Codex CLI persisted write tests mirroring the existing preview and idempotence coverage.
- [ ] `internal/cli/status_test.go` — add supported-client guidance and write-backed status rendering coverage.
- [ ] `internal/cli/init_onboarding_test.go` — replace Claude Desktop-only onboarding expectations with supported-client-aware guidance.
- [ ] `internal/app/snippet_test.go` and `internal/app/doctor_test.go` — lock the updated operator guidance if those surfaces change in Phase 21.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Claude CLI registration against a real host install | CLD-03 | Automated tests can stub `claude`, but cannot prove the live host binary is installed and accepts the command on the operator machine. | Run `optimusctx status --client claude-cli --write` in a disposable repo with Claude installed, then confirm `claude mcp list` shows `optimusctx` under the chosen scope. |
| Codex project-scoped write with explicit config path | CDX-01, CDX-02 | The repo should automate merge behavior, but a real host smoke check is still useful because Phase 21 may keep project-scoped Codex on explicit `--config`. | Create a temp repo-local `.codex/config.toml`, run the write flow with `--config`, then verify the file contains one `[mcp_servers.optimusctx]` entry and unrelated TOML remains intact. |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
