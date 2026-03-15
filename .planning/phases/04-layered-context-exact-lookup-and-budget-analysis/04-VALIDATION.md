---
phase: 04
slug: layered-context-exact-lookup-and-budget-analysis
status: draft
nyquist_compliant: false
wave_0_complete: false
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
| 04-06-01 | 06 | 6 | CTX-06 | unit | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./... -run 'TestBudgetAnalysis|TestBudgetHotspots'` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Cross-check L2 output readability for representative symbol and line-range queries | CTX-03 | Exact anchors and surrounding context may be correct mechanically but still need a human scan for usefulness | Run the future L2 query path against a temp repo fixture, inspect returned anchors, line windows, and stale-file errors for clarity |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
