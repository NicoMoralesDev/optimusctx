---
phase: 10
slug: functional-runtime-validation
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-16
---

# Phase 10 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./internal/repository ./internal/app ./internal/cli ./internal/mcp ./internal/state ./internal/store/sqlite` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~60 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/repository ./internal/app ./internal/cli ./internal/mcp ./internal/state ./internal/store/sqlite`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 10-01-01 | 01 | 1 | EVAL-03 | unit | `go test ./internal/repository -run 'TestEvalScenarioContracts|TestEvalAssertions|TestEvalScenarioValidation'` | ✅ | ✅ green |
| 10-01-02 | 01 | 1 | EVAL-03 | integration | `go test ./internal/app -run 'TestEvalRunnerExecutesCLIWorkflow|TestEvalCLIWorkflowAssertions|TestEvalRunnerPersistsCLIArtifacts'` | ✅ | ✅ green |
| 10-01-03 | 01 | 1 | EVAL-03 | integration | `go test ./internal/app ./internal/cli -run 'TestEvalRunnerPersistsCLIArtifacts|TestEvalCLIScenariosRerun'` | ✅ | ✅ green |
| 10-02-01 | 02 | 2 | EVAL-02 | unit/integration | `go test ./internal/repository ./internal/app ./internal/mcp -run 'TestEvalMCPStepContracts|TestEvalMCPSessionExecution'` | ✅ | ✅ green |
| 10-02-02 | 02 | 2 | EVAL-02 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestEvalMCPInitializeAndToolsList|TestEvalMCPToolFlows'` | ✅ | ✅ green |
| 10-02-03 | 02 | 2 | EVAL-02 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestEvalMCPArtifactsPersist|TestEvalMCPScenariosRerun'` | ✅ | ✅ green |
| 10-03-01 | 03 | 3 | EVAL-03 | unit | `go test ./internal/repository ./internal/app -run 'TestEvalWorkspaceMutations|TestEvalStateMutationValidation'` | ✅ | ✅ green |
| 10-03-02 | 03 | 3 | EVAL-03 | integration | `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestEvalStaleAndDegradedScenarios|TestEvalRecoveryScenarios'` | ✅ | ✅ green |
| 10-03-03 | 03 | 3 | EVAL-03 | integration | `go test ./internal/app ./internal/cli ./internal/mcp ./internal/store/sqlite -run 'TestEvalStateTransitionsPersistEvidence|TestEvalRecoveryAdvancesGeneration'` | ✅ | ✅ green |
| 10-04-01 | 04 | 4 | EVAL-02 | integration/doc | `go test ./internal/app ./internal/cli ./internal/store/sqlite -run 'TestEvalReportSummaries|TestEvalRequirementCoverageReport'` | ✅ | ✅ green |
| 10-04-02 | 04 | 4 | EVAL-03 | integration/doc | `go test ./internal/app ./internal/cli ./internal/store/sqlite -run 'TestEvalReportSummaries|TestEvalRequirementCoverageReport'` | ✅ | ✅ green |
| 10-04-03 | 04 | 4 | EVAL-02, EVAL-03 | integration/doc | `go test ./...` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

No extra Wave 0 setup is required.

Phase 10 starts from the shipped Phase 9 eval harness, but the phase-critical MCP session support, typed mutation support, and milestone-grade reporting are delivered by Waves 1 through 4 rather than by pre-phase bootstrap work.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Functional report wording truthfully maps scenario IDs, rerun commands, and artifact roots to `EVAL-02` and `EVAL-03` | EVAL-02, EVAL-03 | The final milestone-facing wording is easiest to confirm in rendered markdown and generated report output, not only through unit assertions | Review the Phase 10 report artifacts and confirm they list real scenario IDs, real rerun commands, actual artifact paths under `.optimusctx/eval/`, and do not claim MCP tools or operational behaviors the shipped contract does not provide. |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 is complete because no additional pre-phase setup is required
- [x] No watch-mode flags
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** retroactively confirmed by Phase 10 verification on 2026-03-17

## Validation Audit 2026-03-17

| Metric | Count |
|--------|-------|
| Gaps found | 12 |
| Resolved | 12 |
| Escalated | 0 |

Validation state aligned to the existing Phase 10 verification evidence.
