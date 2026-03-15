---
phase: 08
slug: milestone-verification-backfill-and-closure-evidence
status: ready
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-15
---

# Phase 08 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./... -run 'TestApplyRefreshPlan|TestRefreshService|TestMCPServerBasicSession|TestMCPBoundedFailures|TestWatchRefreshUsesCanonicalPipeline|TestPackExportFitsTargetBudget'` |
| **Full suite command** | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./...` |
| **Estimated runtime** | ~25 seconds |

---

## Sampling Rate

- **After every task commit:** Run the targeted evidence subset for the phase area being documented (`go test`, `rg`, or `test -f` as listed below).
- **After every plan wave:** Run `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 08-01-01 | 01 | 1 | REFR-01, REFR-02, REFR-03, REFR-04, REFR-05, CLI-02, MCP-01, MCP-02, MCP-03, MCP-04, OPS-02, OPS-03, OPS-04 | doc | `rg -n 'REFR-0[1-5]|CLI-02|MCP-0[1-4]|OPS-0[2-4]|CLI-05|OPS-01|OPS-05' .planning/v1.0-v1.0-MILESTONE-AUDIT.md .planning/REQUIREMENTS.md .planning/phases/08-milestone-verification-backfill-and-closure-evidence/08-RESEARCH.md` | ✅ | ⬜ pending |
| 08-01-02 | 01 | 1 | REFR-01, REFR-02, REFR-03, REFR-04, REFR-05, CLI-02, MCP-01, MCP-02, MCP-03, MCP-04, OPS-02, OPS-03, OPS-04 | doc | `rg -n '/usr/local/go/bin/go|GOCACHE=/tmp/optimusctx-gocache|Status|Inputs Reviewed|Requirement Verification|Final Verdict' .planning/phases/08-milestone-verification-backfill-and-closure-evidence/08-RESEARCH.md .planning/phases/08-milestone-verification-backfill-and-closure-evidence/08-VALIDATION.md .planning/phases/03-structural-extraction-and-repository-artifact-model/03-VERIFICATION.md .planning/phases/04-layered-context-exact-lookup-and-budget-analysis/04-VERIFICATION.md` | ✅ | ⬜ pending |
| 08-01-03 | 01 | 1 | REFR-01, REFR-02, REFR-03, REFR-04, REFR-05, CLI-02, MCP-01, MCP-02, MCP-03, MCP-04, OPS-02, OPS-03, OPS-04 | doc | `test -f .planning/phases/08-milestone-verification-backfill-and-closure-evidence/08-01-SUMMARY.md` | ✅ | ⬜ pending |
| 08-02-01 | 02 | 2 | REFR-01, REFR-02, REFR-03, REFR-04, REFR-05 | doc | `rg -n 'REFR-0[1-5]' .planning/phases/02-incremental-refresh-and-freshness-model/02-0*-PLAN.md .planning/phases/02-incremental-refresh-and-freshness-model/02-0*-SUMMARY.md .planning/phases/02-incremental-refresh-and-freshness-model/02-VALIDATION.md` | ✅ | ⬜ pending |
| 08-02-02 | 02 | 2 | REFR-01, REFR-02, REFR-03, REFR-04, REFR-05 | unit/integration/doc | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./... -run 'TestMigrationRunner|TestApplyMigrations|TestOpenOrCreateStore|TestRefreshSchemaContracts|TestSnapshotReadModel|TestRepositoryFreshnessState|TestRefreshRunPersistence|TestDegradedRefreshMetadata|TestDiscovery|TestConditionalHashing|TestStreamingHashing|TestRefreshDiff|TestMoveDetection|TestIgnoreTransitions|TestSubtreeFingerprint|TestFingerprintPropagation|TestAffectedDirectories|TestApplyRefreshPlan|TestIncrementalRefreshTransaction|TestDeletedFilesAreRemoved|TestDegradedRefreshState|TestRefreshService|TestNoOpRefresh|TestSnapshotEquivalence|TestInitService|TestInitUsesRefreshBaseline|TestRefreshCommand|TestRefreshCommandErrors|TestInitCommand|TestInitIntegration|TestRefreshIntegration|TestFreshnessStateCLI|TestDegradedRefreshRecovery'` | ✅ | ⬜ pending |
| 08-02-03 | 02 | 2 | REFR-01, REFR-02, REFR-03, REFR-04, REFR-05 | doc | `rg -n 'hermetic|temp repo|historical|UAT|not a blocker' .planning/phases/02-incremental-refresh-and-freshness-model/02-VERIFICATION.md` | ✅ | ⬜ pending |
| 08-03-01 | 03 | 3 | CLI-02, MCP-01, MCP-02, MCP-03, MCP-04 | doc | `rg -n 'CLI-02|MCP-01|MCP-02|MCP-03|MCP-04' .planning/phases/05-mcp-serving-and-integration-contracts/05-0*-PLAN.md .planning/phases/05-mcp-serving-and-integration-contracts/05-0*-SUMMARY.md .planning/phases/05-mcp-serving-and-integration-contracts/05-VALIDATION.md` | ✅ | ⬜ pending |
| 08-03-02 | 03 | 3 | CLI-02, MCP-01, MCP-02, MCP-03, MCP-04 | unit/integration/doc | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./... -run 'TestMCPServeCommand|TestMCPServeReadinessSignalUsesStderr|TestMCPServerBasicSession|TestMCPServerRejectsUnknownTool|TestMCPServerRejectsUnimplementedTool|TestMCPRepositoryQueries|TestMCPLookupQueries|TestMCPBoundedFailures|TestMCPStructuredErrors|TestTokenTree|TestTokenTreeBounds|TestHealthService|TestPackService|TestMCPToolRegistry|TestMCPRefreshPackHealth|TestMCPServerStdioSession|TestSnippetGeneratorRender|TestSnippetInstallCommandAlignment|TestInstallRegistrationDryRun|TestInstallRegistrationConsent|TestInstallNormalizesEphemeralExecutablePath|TestInstallWriteNormalizesEphemeralExecutablePath|TestInstallCommandRejectsUnsupportedClient'` | ✅ | ⬜ pending |
| 08-03-03 | 03 | 3 | CLI-02, MCP-01, MCP-02, MCP-03, MCP-04 | doc | `rg -n 'CLI-02|MCP-01|MCP-02|MCP-03|MCP-04|passed|satisfied|Final Verdict' .planning/phases/05-mcp-serving-and-integration-contracts/05-VERIFICATION.md` | ✅ | ⬜ pending |
| 08-04-01 | 04 | 4 | OPS-02, OPS-03, OPS-04 | doc | `rg -n 'OPS-02|OPS-03|OPS-04|CLI-05|OPS-01|OPS-05|Phase 7' .planning/phases/06-watch-mode-pack-export-and-operational-diagnostics/06-VERIFICATION.md` | ✅ | ⬜ pending |
| 08-04-02 | 04 | 4 | OPS-02, OPS-03, OPS-04 | unit/integration/doc | `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./... -run 'TestWatch|TestWatchCommand|TestWatchCommandErrors|TestWatchRunnerLifecycle|TestWatchRefreshUsesCanonicalPipeline|TestWatchDebouncesBurstEvents|TestWatchOverflowFallsBackToFullRefresh|TestWatchUncertainEventFallsBackToFullRefresh|TestWatchRefreshFailureRecovery|TestWatchStatusStaleHeartbeat|TestRefreshReasonWatch|TestPackExportManifest|TestPackExportWritesPortableArtifact|TestPackExportBudgetPolicy|TestPackExportFitsTargetBudget|TestPackExportFilterRules|TestPackExportCommand|TestPackExportCommandBudgetFlags|TestPackExportCommandErrors'` | ✅ | ⬜ pending |
| 08-04-03 | 04 | 4 | REFR-01, REFR-02, REFR-03, REFR-04, REFR-05, CLI-02, MCP-01, MCP-02, MCP-03, MCP-04, OPS-02, OPS-03, OPS-04 | doc | `test -f .planning/phases/02-incremental-refresh-and-freshness-model/02-VERIFICATION.md && test -f .planning/phases/05-mcp-serving-and-integration-contracts/05-VERIFICATION.md && test -f .planning/phases/06-watch-mode-pack-export-and-operational-diagnostics/06-VERIFICATION.md` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Verification write-ups correctly reconcile historical summaries, current tests, and requirement ownership | REFR-01..05, CLI-02, MCP-01..04, OPS-02..04 | Evidence quality and requirement traceability are document judgments, not just executable assertions | Review each generated `VERIFICATION.md` against the corresponding summaries, validation file, and current tests to confirm every in-scope requirement has explicit evidence and no out-of-scope Phase 7 doctor ownership leaks back into Phase 6. |
| Milestone blocker closure is reflected without rewriting historical audit evidence | REFR-01..05, CLI-02, MCP-01..04, OPS-02..04 | Closure depends on the relationship between current source-of-truth docs and preserved audit artifacts | Confirm the final Phase 8 outputs remove the missing-verification blocker by adding current `VERIFICATION.md` files for Phases 02, 05, and 06 while leaving `.planning/v1.0-v1.0-MILESTONE-AUDIT.md` unchanged as historical evidence. |

---

## Phase-Level Integration Proof

- Verify Phase 02, Phase 05, and Phase 06 each gain a current `VERIFICATION.md` tied to present code and present test results.
- Verify requirement traceability remains aligned with Phase 7 ownership of `CLI-05`, `OPS-01`, and `OPS-05`, while Phase 8 covers only `REFR-01..05`, `CLI-02`, `MCP-01..04`, and `OPS-02..04`.
- Verify the final evidence set is sufficient for a follow-up milestone audit to clear the missing-verification blockers without requiring new feature work.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-03-15
