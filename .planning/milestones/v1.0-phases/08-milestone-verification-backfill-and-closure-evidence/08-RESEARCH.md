# Phase 8 Research: Milestone Verification Backfill and Closure Evidence

## Scope and Planning Intent

Phase 8 is a milestone-closure evidence phase, not a feature phase. The implementation for Phases 2, 5, and 6 already exists, the milestone audit already identifies the missing proof, and the current full suite is green with:

```bash
env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./...
```

The planner should treat this phase as:

- backfilling milestone-grade `VERIFICATION.md` artifacts for Phases 2, 5, and 6
- reconciling those verification files against current code, current tests, and the executed plan summaries
- closing the audit blocker that says those phases are only evidenced by summaries

Requirements in scope for this phase:

- `REFR-01`, `REFR-02`, `REFR-03`, `REFR-04`, `REFR-05`
- `CLI-02`
- `MCP-01`, `MCP-02`, `MCP-03`, `MCP-04`
- `OPS-02`, `OPS-03`, `OPS-04`

Project-local planning guidance review:

- `CLAUDE.md`: not present
- `.claude/skills/`: not present
- `.agents/skills/`: not present

## Exact Evidence Gaps

The gap is narrow and explicit in `.planning/v1.0-v1.0-MILESTONE-AUDIT.md`:

- Phase 02 is missing `.planning/phases/02-incremental-refresh-and-freshness-model/02-VERIFICATION.md`
- Phase 05 is missing `.planning/phases/05-mcp-serving-and-integration-contracts/05-VERIFICATION.md`
- Phase 06 is missing `.planning/phases/06-watch-mode-pack-export-and-operational-diagnostics/06-VERIFICATION.md`

Requirement-level blockers caused by those missing files:

- Phase 02: `REFR-01..05`
- Phase 05: `CLI-02`, `MCP-01..04`
- Phase 06: `OPS-02..04`

Important scope boundary from the current planning sources:

- `CLI-05`, `OPS-01`, and `OPS-05` were moved into Phase 7 gap closure after the audit found a real doctor/watch regression.
- Phase 8 should not reopen that repair. It should only backfill Phase 6 evidence for `OPS-02`, `OPS-03`, and `OPS-04`, while treating Phase 7 as the owner of the doctor/watch fix.

## Current Repo Context

What the planner can rely on now:

- The roadmap and requirements already assign Phase 8 to verification backfill and closure evidence.
- The milestone audit already names the missing verification artifacts and the exact requirement IDs blocked by them.
- Phase 02, Phase 05, and Phase 06 each already have:
  - complete plan summaries
  - `VALIDATION.md`
  - enough implementation and test coverage to support a proper `VERIFICATION.md`
- The current codebase passes the full Go suite with the `/usr/local/go/bin/go` toolchain and `/tmp` caches.

Existing verification-document format to copy:

- concise top-level status and scope
- inputs reviewed
- requirement-by-requirement verification
- explicit evidence sources with file references and test names
- phase-goal and success-criteria verification
- test outcome and residual notes

Good examples:

- `.planning/phases/03-structural-extraction-and-repository-artifact-model/03-VERIFICATION.md`
- `.planning/phases/04-layered-context-exact-lookup-and-budget-analysis/04-VERIFICATION.md`

## Phase-Specific Research Notes

### Phase 02 Backfill

Source-of-truth artifacts:

- `02-01-SUMMARY.md` through `02-06-SUMMARY.md`
- `02-VALIDATION.md`
- `02-UAT.md`
- `README.md`
- implementation and tests under `internal/repository`, `internal/refresh`, `internal/store/sqlite`, `internal/app`, and `internal/cli`

What matters:

- Phase 02 already has strong automated evidence across store, diff engine, app refresh orchestration, and CLI flows.
- The historical `02-UAT.md` includes early failures from running count-based checks in the mutable project worktree. Later Phase 2 plans explicitly fixed those gaps with hermetic temp-repo fixtures and README guidance.
- The Phase 02 verification write-up should explain that the current truth comes from the post-gap-closure implementation and tests, not from the earlier failed manual UAT snapshot.

Likely evidence anchors:

- hash and diff behavior: `internal/repository/discovery_test.go`, `internal/refresh/diff_test.go`
- subtree fingerprints: `internal/refresh/fingerprint_test.go`
- transactional incremental refresh: `internal/store/sqlite/refresh_test.go`, `internal/app/refresh_test.go`
- CLI-visible freshness and degraded recovery: `internal/cli/refresh_test.go`, `internal/cli/refresh_integration_test.go`
- supported manual smoke path: `README.md`

### Phase 05 Backfill

Source-of-truth artifacts:

- `05-01-SUMMARY.md` through `05-08-SUMMARY.md`
- `05-VALIDATION.md`
- `05-UAT.md`
- implementation and tests under `internal/mcp`, `internal/app`, `internal/cli`, and `internal/repository`

What matters:

- Phase 05 has the cleanest evidence story: UAT passed, validation is approved, and the MCP suite already covers transport, registry, structured results, bounded failures, and consent-gated install behavior.
- The verification file needs to synthesize eight plans into one requirement-level argument instead of repeating every summary.
- The current environment differs from some summaries: the stable toolchain is now `/usr/local/go/bin/go`, not `/tmp/optimusctx-go/go/bin/go`. The verification doc should record current commands truthfully rather than copying old paths.

Likely evidence anchors:

- stdio serving and registry: `internal/mcp/server.go`, `internal/mcp/server_test.go`, `internal/mcp/integration_test.go`
- structured query tools and bounded failures: `internal/mcp/query_tools.go`, `internal/mcp/query_tools_test.go`
- token tree, health, and pack services: `internal/app/token_tree_test.go`, `internal/app/health_pack_test.go`
- install and snippet alignment: `internal/app/snippet.go`, `internal/cli/install_test.go`, `internal/app/snippet_test.go`
- readiness signaling and canonical command rendering: `internal/cli/mcp_test.go`, `internal/repository/client_config.go`

### Phase 06 Backfill

Source-of-truth artifacts:

- `06-01-SUMMARY.md` through `06-05-SUMMARY.md`
- `06-VALIDATION.md`
- implementation and tests under `internal/app`, `internal/cli`, and `internal/repository`

What matters:

- Phase 06 mixed valid Phase 6 operational work with a doctor/watch regression later repaired in Phase 7.
- Phase 8 must keep the verification boundary tight:
  - verify `OPS-02` through current watch-refresh reuse behavior
  - verify `OPS-03` through pack export artifact generation
  - verify `OPS-04` through budget fitting and include/exclude policy
- Do not make Phase 06 verification depend on `CLI-05`, `OPS-01`, or `OPS-05`. Those belong to Phase 7 now.

Likely evidence anchors:

- watch reuse and fallback behavior: `internal/app/watch_test.go`, `internal/app/refresh_test.go`, `internal/cli/watch_test.go`
- export manifest and artifact writing: `internal/app/pack_export_test.go`, `internal/cli/pack_test.go`
- budget fit and filter rules: `internal/app/pack_export_test.go`

## Likely Plan Partitioning

The cleanest split is four plans, with one cross-phase inventory pass and three write-up or closure passes:

### Plan A. Evidence Inventory and Verification Template

Purpose:

- collect exact requirement-to-evidence mappings for Phases 2, 5, and 6
- decide one consistent `VERIFICATION.md` format based on Phases 3 and 4
- confirm current verification commands and toolchain paths

Expected outputs:

- a reusable verification outline
- a requirement evidence matrix for all in-scope IDs
- a current command set for later write-ups

### Plan B. Phase 02 Verification Backfill

Purpose:

- write `02-VERIFICATION.md`
- reconcile earlier UAT failures against later hermetic fixture work and current code

Expected outputs:

- current requirement coverage for `REFR-01..05`
- explicit explanation of why the earlier UAT gaps are no longer milestone blockers

### Plan C. Phase 05 Verification Backfill

Purpose:

- write `05-VERIFICATION.md`
- prove `CLI-02` and `MCP-01..04` from current code and current MCP tests

Expected outputs:

- one coherent verification artifact covering transport, structured payloads, bounded failures, and install registration

### Plan D. Phase 06 Verification Backfill and Milestone Closure Check

Purpose:

- write `06-VERIFICATION.md` limited to `OPS-02..04`
- re-check milestone traceability after all three verification files exist

Expected outputs:

- Phase 06 verification bounded to the right requirement IDs
- closure note or updated evidence that the audit blocker about missing verification files is removed

If the planner prefers five plans, split the final closure check into its own plan so the last plan is explicitly milestone-facing rather than phase-facing.

## Dependencies

Hard dependencies:

- Phase 7 artifact and code state must remain the source of truth for doctor/watch semantics. Phase 8 should not overwrite that ownership.
- All plan summaries for Phases 02, 05, and 06 must be treated as historical execution evidence.
- Existing validation files should inform command selection and evidence organization.

Soft dependencies:

- none on new feature implementation if the suite remains green
- possible need for very small docs edits if an evidence file reveals naming drift or stale command text

## Verification Commands and Evidence Sources

Recommended base command for current verification runs:

```bash
env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./...
```

Recommended targeted command groups:

### Phase 02

```bash
env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./... -run 'TestMigrationRunner|TestApplyMigrations|TestOpenOrCreateStore|TestRefreshSchemaContracts|TestSnapshotReadModel|TestRepositoryFreshnessState|TestRefreshRunPersistence|TestDegradedRefreshMetadata|TestDiscovery|TestConditionalHashing|TestStreamingHashing|TestRefreshDiff|TestMoveDetection|TestIgnoreTransitions|TestSubtreeFingerprint|TestFingerprintPropagation|TestAffectedDirectories|TestApplyRefreshPlan|TestIncrementalRefreshTransaction|TestDeletedFilesAreRemoved|TestDegradedRefreshState|TestRefreshService|TestNoOpRefresh|TestSnapshotEquivalence|TestInitService|TestInitUsesRefreshBaseline|TestRefreshCommand|TestRefreshCommandErrors|TestInitCommand|TestInitIntegration|TestRefreshIntegration|TestFreshnessStateCLI|TestDegradedRefreshRecovery'
```

Manual evidence source:

- `README.md` temp-repo smoke flow for `optimusctx init` and `optimusctx refresh`

### Phase 05

```bash
env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./... -run 'TestMCPServeCommand|TestMCPServeReadinessSignalUsesStderr|TestMCPServerBasicSession|TestMCPServerRejectsUnknownTool|TestMCPServerRejectsUnimplementedTool|TestMCPRepositoryQueries|TestMCPLookupQueries|TestMCPBoundedFailures|TestMCPStructuredErrors|TestTokenTree|TestTokenTreeBounds|TestHealthService|TestPackService|TestMCPToolRegistry|TestMCPRefreshPackHealth|TestMCPServerStdioSession|TestSnippetGeneratorRender|TestSnippetInstallCommandAlignment|TestInstallRegistrationDryRun|TestInstallRegistrationConsent|TestInstallNormalizesEphemeralExecutablePath|TestInstallWriteNormalizesEphemeralExecutablePath|TestInstallCommandRejectsUnsupportedClient'
```

Manual evidence sources:

- `05-UAT.md`
- dry-run install preview versus snippet output

### Phase 06

```bash
env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./... -run 'TestWatch|TestWatchCommand|TestWatchCommandErrors|TestWatchRunnerLifecycle|TestWatchRefreshUsesCanonicalPipeline|TestWatchDebouncesBurstEvents|TestWatchOverflowFallsBackToFullRefresh|TestWatchUncertainEventFallsBackToFullRefresh|TestWatchRefreshFailureRecovery|TestWatchStatusStaleHeartbeat|TestRefreshReasonWatch|TestPackExportManifest|TestPackExportWritesPortableArtifact|TestPackExportBudgetPolicy|TestPackExportFitsTargetBudget|TestPackExportFilterRules|TestPackExportCommand|TestPackExportCommandBudgetFlags|TestPackExportCommandErrors'
```

Manual evidence sources:

- `06-VALIDATION.md` manual-only notes for long-lived watch UX
- pack export CLI behavior against stdout/file modes if the planner wants one human confirmation

## Validation Architecture

Phase 8 should produce a validation artifact because the phase is fundamentally about evidence quality.

Recommended `08-VALIDATION.md` shape:

- framework: `go test`
- quick run: one combined targeted command covering the three backfill phases
- full run: the current full-suite command
- per-plan verification map tied to the four recommended plans above
- manual-only checks:
  - verify each new `VERIFICATION.md` matches current requirement ownership
  - verify Phase 06 verification does not reclaim `CLI-05`, `OPS-01`, or `OPS-05`
  - verify milestone audit blockers about missing phase verification are closed by artifact presence

Recommended combined quick run:

```bash
env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./... -run 'TestRefresh|TestDiscovery|TestFingerprint|TestMCP|TestTokenTree|TestHealthService|TestPackService|TestInstall|TestSnippet|TestWatch|TestPackExport'
```

Validation focus:

- artifact correctness over feature invention
- current command truth over copied historical command text
- requirement ownership boundaries preserved after Phase 7

## Risks

### 1. Turning a docs phase back into a product phase

If the planner assumes missing verification files imply missing implementation, Phase 8 will expand unnecessarily. Current evidence says the implementation exists and the suite is green.

### 2. Pulling Phase 7 doctor/watch semantics back into Phase 6 verification

This would reintroduce traceability confusion. Phase 06 verification should stay limited to `OPS-02..04`.

### 3. Copying stale toolchain commands from old summaries

Older summaries mention `/tmp/optimusctx-go/go/bin/go`. The current successful full-suite run used `/usr/local/go/bin/go` with `/tmp` caches. Verification artifacts should report what works now.

### 4. Repeating summary prose instead of writing requirement evidence

The phase goal is not to restate execution history. The write-ups need direct requirement-to-code-to-test arguments.

### 5. Losing the distinction between historical evidence and current verification

Plan summaries and UAT files are valuable inputs, but milestone verification should ultimately describe the current codebase and current test outcomes.

## Recommended Scope Boundaries

In scope:

- create `02-VERIFICATION.md`, `05-VERIFICATION.md`, and `06-VERIFICATION.md`
- run current targeted and full verification commands as needed
- cite current code, tests, summaries, validation files, and UAT where relevant
- add a Phase 8 validation artifact if the planner wants structured execution control
- perform a final traceability and closure pass against the milestone audit

Out of scope unless verification fails:

- new runtime behavior
- watch/doctor semantics work already assigned to Phase 7
- changes to requirement ownership in `REQUIREMENTS.md`
- rewriting historical audit findings instead of superseding them with current evidence

## Planner Takeaways

- Treat Phase 8 as evidence synthesis and closure, not implementation.
- Start with one cross-phase evidence matrix so the three verification docs stay consistent.
- Keep Phase 06 verification scoped to `OPS-02..04`; let Phase 7 own doctor/watch correctness.
- Prefer current test commands and current code references over historical environment details.
- End with an explicit closure check against `.planning/v1.0-v1.0-MILESTONE-AUDIT.md` so the milestone blocker removal is demonstrable.
