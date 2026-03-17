---
phase: 12
slug: token-attribution-and-evidence-reporting
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-16
---

# Phase 12 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./internal/repository ./internal/app ./internal/store/sqlite ./internal/cli ./internal/mcp -run 'TestBenchmark|TestToken|TestEvidence|TestReport|TestAttribution|TestVerification'` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~45 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/repository ./internal/app ./internal/store/sqlite ./internal/cli ./internal/mcp -run 'TestBenchmark|TestToken|TestEvidence|TestReport|TestAttribution|TestVerification'`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 12-01-01 | 01 | 1 | BNCH-02 | unit/integration | `go test ./internal/repository ./internal/app -run 'TestBenchmark|TestToken|TestAttribution'` | ✅ | ✅ green |
| 12-01-02 | 01 | 1 | BNCH-02 | integration | `go test ./internal/app ./internal/store/sqlite -run 'TestBenchmark|TestToken|TestAttribution'` | ✅ | ✅ green |
| 12-01-03 | 01 | 1 | BNCH-02 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmark|TestToken|TestAttribution'` | ✅ | ✅ green |
| 12-02-01 | 02 | 2 | BNCH-04 | unit/integration | `go test ./internal/repository ./internal/store/sqlite -run 'TestBenchmark|TestEvidence|TestVerification'` | ✅ | ✅ green |
| 12-02-02 | 02 | 2 | BNCH-04 | integration | `go test ./internal/app ./internal/store/sqlite -run 'TestBenchmark|TestEvidence|TestVerification'` | ✅ | ✅ green |
| 12-02-03 | 02 | 2 | BNCH-02, BNCH-04 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmark|TestEvidence|TestVerification'` | ✅ | ✅ green |
| 12-03-01 | 03 | 3 | BNCH-04 | unit | `go test ./internal/repository ./internal/app -run 'TestBenchmark|TestReport|TestAttribution'` | ✅ | ✅ green |
| 12-03-02 | 03 | 3 | BNCH-02, BNCH-04 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmark|TestReport|TestAttribution'` | ✅ | ✅ green |
| 12-03-03 | 03 | 3 | BNCH-04 | doc/integration | `go test ./internal/app ./internal/store/sqlite -run 'TestBenchmark|TestReport|TestVerification'` | ✅ | ✅ green |
| 12-04-01 | 04 | 4 | BNCH-04 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmark|TestVerification'` | ✅ | ✅ green |
| 12-04-02 | 04 | 4 | BNCH-02, BNCH-04 | integration | `go test ./internal/repository ./internal/app ./internal/store/sqlite -run 'TestBenchmark|TestAttribution|TestVerification'` | ✅ | ✅ green |
| 12-04-03 | 04 | 4 | BNCH-02, BNCH-04 | integration/doc | `go test ./...` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/repository/benchmark_test.go` — token attribution contract and artifact taxonomy coverage
- [x] `internal/app/benchmark_runner_test.go` — estimator, attribution, and report-input orchestration coverage
- [x] `internal/store/sqlite/benchmark_test.go` — derived evidence persistence and deterministic export coverage
- [x] `internal/cli/eval_integration_test.go` — end-to-end benchmark export and report verification
- [x] `internal/mcp/integration_test.go` — MCP-surface benchmark evidence and reproducibility coverage

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Final benchmark wording truthfully distinguishes estimated token savings from provider-billed tokens | BNCH-02 | Automated phrase assertions can catch forbidden wording, but final claim clarity still needs a reviewer pass on the rendered report | Generate the Phase 12 human-readable benchmark report, inspect summary language, and confirm it consistently says `bytes_div_4_ceiling` estimated tokens, includes the artifact-attribution caveats, and avoids provider-billing claims. |
| Machine-readable exports are understandable and rerunnable by an operator reading the artifact set | BNCH-04 | Artifact usability depends on documentation clarity and rendered export structure | Inspect one exported evidence bundle plus README guidance, then verify the rerun command and methodology fingerprint are visible without reading code. |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** auto-approved after retroactive audit on 2026-03-17

## Validation Audit 2026-03-17

| Metric | Count |
|--------|-------|
| Gaps found | 12 |
| Resolved | 12 |
| Escalated | 0 |

Retroactive audit confirmed the existing Phase 12 automated coverage without adding new tests.

Executed evidence:

- `go test ./internal/repository ./internal/app ./internal/store/sqlite ./internal/cli ./internal/mcp -run 'TestBenchmark|TestToken|TestEvidence|TestReport|TestAttribution|TestVerification'`
