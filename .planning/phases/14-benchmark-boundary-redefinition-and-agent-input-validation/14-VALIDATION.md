---
phase: 14
slug: benchmark-boundary-redefinition-and-agent-input-validation
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-16
---

# Phase 14 - Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./internal/repository ./internal/app ./internal/store/sqlite ./internal/cli ./internal/mcp -run 'TestBenchmark'` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~75 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/repository ./internal/app ./internal/store/sqlite ./internal/cli ./internal/mcp -run 'TestBenchmark'`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 90 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 14-01-01 | 01 | 1 | BNCH-01, BNCH-02 | unit | `go test ./internal/repository -run 'TestBenchmark(Suite|Task|Schema|Validation)'` | ✅ | ⬜ pending |
| 14-01-02 | 01 | 1 | BNCH-02, BNCH-04 | unit | `go test ./internal/repository ./internal/store/sqlite -run 'TestBenchmark(Evidence|Bundle|Methodology|Fingerprint)'` | ✅ | ⬜ pending |
| 14-01-03 | 01 | 1 | BNCH-01, BNCH-04 | integration | `go test ./internal/app ./internal/cli -run 'TestBenchmark(Export|Report|Verify|Methodology)'` | ✅ | ⬜ pending |
| 14-02-01 | 02 | 2 | BNCH-01, BNCH-02 | unit | `go test ./internal/app -run 'TestBenchmark(AgentInput|Attribution|Boundary|Projection)'` | ✅ | ⬜ pending |
| 14-02-02 | 02 | 2 | BNCH-01, BNCH-04 | unit | `go test ./internal/app -run 'TestBenchmark(FinalArtifact|LaneCompletion|RepeatedRuns)'` | ✅ | ⬜ pending |
| 14-02-03 | 02 | 2 | BNCH-02, BNCH-04 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmark(Report|HumanSummary|Integration)'` | ✅ | ⬜ pending |
| 14-03-01 | 03 | 3 | BNCH-01, BNCH-02 | integration | `go test ./internal/repository ./internal/app -run 'TestBenchmark(Migration|SuiteV2|Corpus)'` | ✅ | ⬜ pending |
| 14-03-02 | 03 | 3 | BNCH-04 | integration | `go test ./internal/cli -run 'TestBenchmark(Export|Verify|Report)'` | ✅ | ⬜ pending |
| 14-03-03 | 03 | 3 | BNCH-01, BNCH-02, BNCH-04 | integration | `go run ./cmd/optimusctx eval benchmark export --suite go-benchmark-discovery-v1 --attempts 2 && go run ./cmd/optimusctx eval benchmark verify --suite go-benchmark-discovery-v1 --attempts 2 && go run ./cmd/optimusctx eval benchmark report --suite go-benchmark-discovery-v1 --attempts 2 && go run ./cmd/optimusctx eval benchmark export --suite go-benchmark-refresh-v1 --attempts 2 && go run ./cmd/optimusctx eval benchmark verify --suite go-benchmark-refresh-v1 --attempts 2 && go run ./cmd/optimusctx eval benchmark report --suite go-benchmark-refresh-v1 --attempts 2` | ✅ | ⬜ pending |
| 14-04-01 | 04 | 3 | BNCH-04 | integration | `go test ./internal/app ./internal/cli ./internal/store/sqlite -run 'TestBenchmark(Reproducibility|Persistence|Verification)'` | ✅ | ⬜ pending |
| 14-04-02 | 04 | 3 | BNCH-02, BNCH-04 | integration/doc | `go test ./internal/app ./internal/cli -run 'TestBenchmark(Report|Wording|HumanSummary)'` | ✅ | ⬜ pending |
| 14-04-03 | 04 | 3 | BNCH-01, BNCH-02, BNCH-04 | integration/doc | `go test ./...` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing benchmark infrastructure covers this phase. Add dedicated `*_v2_test.go` files only if the existing benchmark test suites become too overloaded during execution.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| The new benchmark boundary matches the product mental model of "system work vs agent input" | BNCH-01, BNCH-02 | This is a methodological truthfulness question, not just a code-path question | Review the v2 suite schema, runner docs, and refreshed report; confirm `refresh`, `health`, and internal discovery are only counted when explicitly promoted to agent-facing inputs. |
| Final-artifact contracts are strict enough to prevent bogus wins but not so brittle that deterministic reruns fail | BNCH-01, BNCH-04 | This is a tradeoff that requires human judgment about benchmark quality | Inspect the migrated suite contracts and rerun reports; confirm both arms are validated against comparable normalized outputs and that the rules do not depend on incidental transport payload shape. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
