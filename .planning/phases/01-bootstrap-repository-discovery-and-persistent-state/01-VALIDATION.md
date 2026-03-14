---
phase: 1
slug: bootstrap-repository-discovery-and-persistent-state
status: ready
nyquist_compliant: true
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
| 1-01-01 | 01 | 1 | CLI-01 | integration | `go test ./... && TMP_GOBIN="$(mktemp -d)" && GOBIN="$TMP_GOBIN" go install ./cmd/optimusctx && "$TMP_GOBIN/optimusctx" version` | ❌ planned | ⬜ pending |
| 1-02-01 | 02 | 2 | REPO-01 | unit | `go test ./... -run 'TestRepositoryLocator|TestResolveRepositoryRoot'` | ❌ planned | ⬜ pending |
| 1-02-02 | 02 | 2 | REPO-02 | unit | `go test ./... -run 'TestDiscovery|TestIgnoreMatcher'` | ❌ planned | ⬜ pending |
| 1-02-03 | 02 | 2 | REPO-05 | unit | `go test ./... -run 'TestDiscoveryMetadata|TestFileRecord'` | ❌ planned | ⬜ pending |
| 1-03-01 | 03 | 2 | REPO-03 | integration | `go test ./... -run 'TestStateLayout|TestEnsureStateLayout|TestSQLiteStore|TestOpenOrCreateStore'` | ❌ planned | ⬜ pending |
| 1-03-02 | 03 | 2 | REPO-04 | integration | `go test ./... -run 'TestMigrationRunner|TestApplyMigrations|TestSQLiteStore|TestOpenOrCreateStore'` | ❌ planned | ⬜ pending |
| 1-04-01 | 04 | 3 | CLI-03 | integration | `go test ./... -run 'TestInitService|TestInitWorkflow|TestInitCommand|TestInitIntegration'` | ❌ planned | ⬜ pending |
| 1-04-02 | 04 | 3 | CLI-04 | integration | `go test ./... -run 'TestSnippetCommand|TestSnippetGenerator'` | ❌ planned | ⬜ pending |
| 1-04-03 | 04 | 3 | REPO-05 | integration | `go test ./... -run 'TestInitService|TestInitWorkflow|TestInitIntegration'` | ❌ planned | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `go.mod` — initialize module and toolchain baseline
- [ ] `cmd/optimusctx/main.go` — CLI entrypoint for command tests
- [ ] `internal/...` package skeleton — runtime services under test
- [ ] `*_test.go` files for CLI, repository discovery, and SQLite initialization — coverage stubs for phase requirements
- [ ] fixture repositories for nested-root discovery, ignore matching, and init/snippet integration flows
- [ ] assertions for `state.json` fields and SQLite schema indexes in migration/store tests

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Printed snippet is clear and copyable | CLI-04 | Quality of snippet wording is easier to review manually than assert exhaustively | Run `optimusctx snippet` and confirm it prints only stdout content, no file writes, and no misleading MCP claims |
| Init output is understandable to a first-time operator | CLI-03 | Human-facing output quality matters beyond exit code assertions | Run `optimusctx init` in a fixture repo and confirm root path, state path, schema version, and file count are clearly reported |
| Ignored-path reporting is understandable | REPO-05 | Humans should confirm ignored files are surfaced with reasons instead of disappearing silently | Run `optimusctx init` against fixtures using both `.gitignore` and built-in exclusions and confirm ignored counts/reasons are inspectable through test outputs or fixtures |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-03-14
