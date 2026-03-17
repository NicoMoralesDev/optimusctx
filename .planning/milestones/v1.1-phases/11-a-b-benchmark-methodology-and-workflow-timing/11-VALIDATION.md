---
phase: 11
slug: a-b-benchmark-methodology-and-workflow-timing
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-16
---

# Phase 11 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `Run the task-specific automated command from the verification map` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~25 seconds |

---

## Sampling Rate

- **After every task commit:** Run that task's `Automated Command` from the verification map. If the referenced tests do not exist yet, complete the corresponding Wave 0 prerequisite in the same task before declaring it done.
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 11-01-01 | 01 | 1 | BNCH-01 | unit/integration | `go test ./internal/repository ./internal/app -run 'TestBenchmark'` | ✅ | ✅ green |
| 11-01-02 | 01 | 1 | BNCH-01 | integration | `go test ./internal/app ./internal/store/sqlite -run 'TestBenchmark'` | ✅ | ✅ green |
| 11-01-03 | 01 | 1 | BNCH-01 | integration | `go test ./internal/app ./internal/store/sqlite -run 'TestBenchmark'` | ✅ | ✅ green |
| 11-02-01 | 02 | 2 | BNCH-03 | unit | `go test ./internal/repository ./internal/app -run 'TestBenchmark'` | ✅ | ✅ green |
| 11-02-02 | 02 | 2 | BNCH-03 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmark'` | ✅ | ✅ green |
| 11-02-03 | 02 | 2 | BNCH-01 | integration | `go test ./internal/app ./internal/store/sqlite -run 'TestBenchmark'` | ✅ | ✅ green |
| 11-03-01 | 03 | 3 | BNCH-03 | unit/integration | `go test ./internal/repository ./internal/app -run 'TestBenchmark'` | ✅ | ✅ green |
| 11-03-02 | 03 | 3 | BNCH-03 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmark'` | ✅ | ✅ green |
| 11-03-03 | 03 | 3 | BNCH-01 | integration | `go test ./internal/app ./internal/store/sqlite -run 'TestBenchmark'` | ✅ | ✅ green |
| 11-04-01 | 04 | 4 | BNCH-01 | unit/integration | `go test ./internal/repository ./internal/app ./internal/store/sqlite -run 'TestBenchmark'` | ✅ | ✅ green |
| 11-04-02 | 04 | 4 | BNCH-03 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmark'` | ✅ | ✅ green |
| 11-04-03 | 04 | 4 | BNCH-01, BNCH-03 | integration/doc | `go test ./...` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/repository/benchmark_test.go` — contract coverage for suites, arms, lanes, stop conditions, and baseline actions
- [x] `internal/app/benchmark_runner_test.go` — orchestration coverage for timing capture and paired workflow execution
- [x] `internal/store/sqlite/benchmark_test.go` — persistence coverage for repeated-run samples and comparison summaries
- [x] `internal/app/benchmark_service.go` — repeated-run orchestration path for benchmark verification
- [x] `internal/store/migrations/0005_benchmark_runs.sql` — schema for persisted benchmark runs and lane samples
- [x] `testdata/eval/benchmarks` — committed benchmark suites and frozen corpus definitions

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Benchmark wording truthfully distinguishes baseline actions from OptimusCtx treatment actions | BNCH-01 | Reviewer must confirm no benchmark report or docs blur the product boundary | Inspect benchmark scenario docs and generated comparison summaries; verify baseline uses only allowed non-OptimusCtx actions and treatment uses shipped CLI or MCP surfaces only. |
| Benchmark stop conditions match the intended workflow milestones | BNCH-03 | The correctness of lane boundaries is easier to judge in rendered artifacts than in unit assertions alone | Review the Phase 11 benchmark artifacts and confirm discovery, context assembly, refresh-after-change, and task-completion lanes each stop on explicit machine-checkable milestones. |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** auto-approved after retroactive audit on 2026-03-17

## Validation Audit 2026-03-17

| Metric | Count |
|--------|-------|
| Gaps found | 12 |
| Resolved | 12 |
| Escalated | 0 |

Retroactive audit confirmed the existing Phase 11 automated coverage without adding new tests.

Executed evidence:

- `go test ./internal/repository ./internal/app ./internal/store/sqlite ./internal/cli ./internal/mcp -run 'TestBenchmark'`
