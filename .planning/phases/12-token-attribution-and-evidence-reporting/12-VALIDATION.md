---
phase: 12
slug: token-attribution-and-evidence-reporting
status: draft
nyquist_compliant: true
wave_0_complete: false
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
| 12-01-01 | 01 | 1 | BNCH-02 | unit/integration | `go test ./internal/repository ./internal/app -run 'TestTokenAttributionContract|TestArtifactTypeAttribution'` | ❌ W0 | ⬜ pending |
| 12-01-02 | 01 | 1 | BNCH-02 | integration | `go test ./internal/app ./internal/store/sqlite -run 'TestBenchmarkTokenEstimation|TestAttributionPersistenceInputs'` | ❌ W0 | ⬜ pending |
| 12-01-03 | 01 | 1 | BNCH-02 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmarkArtifactAttribution'` | ❌ W0 | ⬜ pending |
| 12-02-01 | 02 | 2 | BNCH-04 | unit/integration | `go test ./internal/repository ./internal/store/sqlite -run 'TestBenchmarkEvidenceBundleSchema|TestBenchmarkExportDeterminism'` | ❌ W0 | ⬜ pending |
| 12-02-02 | 02 | 2 | BNCH-04 | integration | `go test ./internal/app ./internal/store/sqlite -run 'TestBenchmarkExportPersistence|TestBenchmarkComparisonExport'` | ❌ W0 | ⬜ pending |
| 12-02-03 | 02 | 2 | BNCH-02, BNCH-04 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmarkEvidenceBundleGeneration'` | ❌ W0 | ⬜ pending |
| 12-03-01 | 03 | 3 | BNCH-04 | unit | `go test ./internal/repository ./internal/app -run 'TestBenchmarkHumanSummaryInputs|TestBenchmarkComparisonReportRendering'` | ❌ W0 | ⬜ pending |
| 12-03-02 | 03 | 3 | BNCH-02, BNCH-04 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmarkHumanReadableReport|TestBenchmarkAttributionTables'` | ❌ W0 | ⬜ pending |
| 12-03-03 | 03 | 3 | BNCH-04 | doc/integration | `go test ./internal/app ./internal/store/sqlite -run 'TestBenchmarkReportReusesPersistedEvidence'` | ❌ W0 | ⬜ pending |
| 12-04-01 | 04 | 4 | BNCH-04 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmarkRerunReproducibility|TestBenchmarkMethodologyFingerprint'` | ❌ W0 | ⬜ pending |
| 12-04-02 | 04 | 4 | BNCH-02, BNCH-04 | integration | `go test ./internal/repository ./internal/app ./internal/store/sqlite -run 'TestBenchmarkRecomputedAttributionMatchesExport|TestBenchmarkMilestoneVerification'` | ❌ W0 | ⬜ pending |
| 12-04-03 | 04 | 4 | BNCH-02, BNCH-04 | integration/doc | `go test ./...` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/repository/benchmark_test.go` — token attribution contract and artifact taxonomy coverage
- [ ] `internal/app/benchmark_runner_test.go` — estimator, attribution, and report-input orchestration coverage
- [ ] `internal/store/sqlite/benchmark_test.go` — derived evidence persistence and deterministic export coverage
- [ ] `internal/cli/eval_integration_test.go` — end-to-end benchmark export and report verification
- [ ] `internal/mcp/integration_test.go` — MCP-surface benchmark evidence and reproducibility coverage

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Final benchmark wording truthfully distinguishes estimated token savings from provider-billed tokens | BNCH-02 | Automated phrase assertions can catch forbidden wording, but final claim clarity still needs a reviewer pass on the rendered report | Generate the Phase 12 human-readable benchmark report, inspect summary language, and confirm it consistently says `bytes_div_4_ceiling` estimated tokens, includes the artifact-attribution caveats, and avoids provider-billing claims. |
| Machine-readable exports are understandable and rerunnable by an operator reading the artifact set | BNCH-04 | Artifact usability depends on documentation clarity and rendered export structure | Inspect one exported evidence bundle plus README guidance, then verify the rerun command and methodology fingerprint are visible without reading code. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
