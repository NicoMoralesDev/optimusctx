---
phase: 05
slug: mcp-serving-and-integration-contracts
status: ready
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-15
---

# Phase 05 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./internal/...` |
| **Full suite command** | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./...` |
| **Estimated runtime** | ~60 seconds |

---

## Sampling Rate

- **After every task commit:** Run `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./internal/...`
- **After every plan wave:** Run `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 05-01-01 | 01 | 1 | MCP-01 | integration | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestMCPServeCommand|TestMCPServerStdioSession'` | ✅ | ⬜ pending |
| 05-02-01 | 02 | 2 | MCP-02, MCP-03, MCP-04 | integration | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestMCPRepositoryQueries|TestMCPBoundedFailures'` | ✅ | ⬜ pending |
| 05-03-01 | 03 | 2 | MCP-03, MCP-04 | unit | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestTokenTree|TestTokenTreeBounds'` | ✅ | ⬜ pending |
| 05-04-01 | 04 | 2 | MCP-03, MCP-04 | unit | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestHealthService|TestPackService'` | ✅ | ⬜ pending |
| 05-05-01 | 05 | 3 | MCP-01, MCP-02, MCP-03, MCP-04 | integration | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestMCPToolRegistry|TestMCPRefreshPackHealth'` | ✅ | ⬜ pending |
| 05-06-01 | 06 | 4 | CLI-02 | integration | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestInstallRegistrationDryRun|TestInstallRegistrationConsent'` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `/tmp/optimusctx-go/go/bin/go` remains the pinned toolchain for quick and full suite verification.
- [x] Existing app-service tests under `internal/app` provide reusable temp-repository fixtures for repository map, lookup, context block, refresh, and budget-backed behavior.
- [x] Existing CLI command tests under `internal/cli` provide the baseline pattern for command parsing, stdout assertions, and explicit error handling.
- [x] Existing persisted metadata from Phases 2 through 4 is sufficient to validate read-only MCP tools without requiring new indexing infrastructure.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Supported client registration preview matches the expected config shape for one real client target | CLI-02 | File-path and config-shape expectations are client-specific and benefit from one human review in addition to automated tests | Run the dry-run registration command, inspect the rendered config block, and confirm it points at `optimusctx mcp serve` with no unrelated edits |

---

## Phase-Level Integration Proof

- Verify `optimusctx mcp serve` exposes the full promised tool registry and can answer one representative read-only request over the real server boundary.
- Verify repository map, context, lookup, and refresh-style MCP results all carry freshness plus cache-versus-refresh metadata in one consistent envelope.
- Verify token tree, health, and pack tools stay deterministic and bounded, with field-specific failures for oversized requests.
- Verify registration never writes client config unless the request is explicit and consent is present.
- Recommended integrated command:
  `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestMCPServerStdioSession|TestMCPRepositoryQueries|TestMCPRefreshPackHealth|TestInstallRegistrationConsent'`

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-03-15
