---
phase: 04
slug: layered-context-exact-lookup-and-budget-analysis
status: ready
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-15
---

# Phase 04 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./internal/...` |
| **Full suite command** | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./...` |
| **Estimated runtime** | ~45 seconds |

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
| 04-01-01 | 01 | 1 | CTX-01 | unit | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestLayeredContextL0|TestRepositorySummary'` | ✅ | ⬜ pending |
| 04-02-01 | 02 | 2 | CTX-02 | unit | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestLayeredContextL1|TestLayeredContextOrdering'` | ✅ | ⬜ pending |
| 04-03-01 | 03 | 3 | CTX-04 | unit | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestSymbolLookup|TestSymbolLookupFilters'` | ✅ | ⬜ pending |
| 04-04-01 | 04 | 4 | CTX-05 | unit | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestStructureLookup|TestStructureLookupValidation'` | ✅ | ⬜ pending |
| 04-05-01 | 05 | 5 | CTX-03 | integration | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestTargetedContextBlock|TestTargetedContextBlockFileAvailability'` | ✅ | ⬜ pending |
| 04-06-01 | 06 | 3 | CTX-06 | unit | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestBudgetAnalysis|TestBudgetHotspots'` | ✅ | ⬜ pending |
| 04-05-02 | 05 | 5 | CTX-03, CTX-04 | integration | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestLookupDrivenContextBlock|TestLayeredContextFreshnessContract'` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `/tmp/optimusctx-go/go/bin/go` is the pinned Phase 1-3 toolchain used for all Phase 4 quick and full suite commands.
- [x] Existing temp-repository helpers in `internal/app/refresh_test.go` and `internal/app/repository_map_test.go` are reused for persisted-read and live-file L2 fixtures.
- [x] Existing SQLite-backed store test patterns in `internal/store/sqlite/store_test.go` cover deterministic ordering, filter semantics, and persisted-only query assertions.
- [x] Existing app-service patterns for repository resolution and injected file reads in `internal/app/repository_map.go` and `internal/app/refresh.go` are the required shared contracts for Phase 4 services.

---

## Manual-Only Verifications

All Phase 4 core behaviors are expected to have automated verification. Optional operator review can inspect one representative L2 block after execution, but Phase 4 planning does not require a manual checkpoint.

---

## Phase-Level Integration Proof

- Verify L0, L1, and L2 all carry the same repository identity and freshness contract across one indexed temp repository.
- Verify exact symbol lookup returns stable identity that can drive L2 symbol-targeted context assembly without reparsing or fuzzy matching.
- Verify the end-to-end layered surface remains persisted-first: L0, L1, exact lookup, and structure lookup succeed from SQLite alone, while only L2 performs live file reads for final code windows.
- Verify budget analysis reports deterministic file and directory hotspots from persisted size metadata and stays aligned with the same repository/freshness contract used by L0 and L1 outputs.
- Recommended integrated command:
  `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestLayeredContextFreshnessContract|TestLookupDrivenContextBlock|TestPersistedFirstQuerySurface|TestBudgetHotspots'`

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-03-15
