# Phase 06 Verification: Watch Reuse, Pack Export, and Budget Fitting

## Status

`passed`

## Scope and Verification Standard

This report backfills the missing milestone-grade verification artifact for Phase 06.
It verifies the current shipped behavior for `OPS-02`, `OPS-03`, and `OPS-04` from
current implementation, current automated evidence, current Phase 06 summaries, and
current Phase 08 closure guidance.

This report is intentionally requirement-driven rather than plan-chronology-driven:

- `OPS-02`: watch mode reuses the same canonical incremental refresh pipeline as the
  manual refresh path.
- `OPS-03`: users can export a compact repository pack for offline or non-MCP
  workflows.
- `OPS-04`: pack export can fit output to a target budget while respecting explicit
  include and exclude scope rules.

This report also establishes the Phase 6 to Phase 7 ownership boundary required for
milestone closure:

- `OPS-01` is intentionally excluded here because optional-watch health semantics were
  repaired and re-verified in Phase 7.
- `CLI-05` is intentionally excluded here because the doctor command contract is owned
  by Phase 7 for closure purposes.
- `OPS-05` is intentionally excluded here because doctor diagnostics and top token-cost
  reporting were repaired and re-verified in Phase 7.

The current backfill therefore proves only the surviving Phase 6 operational scope.
It does not reopen doctor semantics, does not rewrite historical audit artifacts, and
does not treat earlier Phase 6 plan ordering as proof.

## Evidence Inputs

This report draws from:

- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `.planning/phases/08-milestone-verification-backfill-and-closure-evidence/08-RESEARCH.md`
- `.planning/phases/08-milestone-verification-backfill-and-closure-evidence/08-VALIDATION.md`
- `.planning/phases/08-milestone-verification-backfill-and-closure-evidence/08-01-SUMMARY.md`
- `.planning/phases/06-watch-mode-pack-export-and-operational-diagnostics/06-01-SUMMARY.md`
- `.planning/phases/06-watch-mode-pack-export-and-operational-diagnostics/06-02-SUMMARY.md`
- `.planning/phases/06-watch-mode-pack-export-and-operational-diagnostics/06-03-SUMMARY.md`
- `.planning/phases/06-watch-mode-pack-export-and-operational-diagnostics/06-04-SUMMARY.md`
- `.planning/phases/06-watch-mode-pack-export-and-operational-diagnostics/06-05-SUMMARY.md`
- `.planning/phases/06-watch-mode-pack-export-and-operational-diagnostics/06-VALIDATION.md`
- `.planning/phases/07-doctor-health-semantics-and-milestone-state-repair/07-01-SUMMARY.md`
- `internal/app/watch_test.go`
- `internal/app/refresh_test.go`
- `internal/app/pack_export_test.go`
- `internal/cli/watch_test.go`
- `internal/cli/pack_test.go`

## Scope Boundary Reconciliation

Phase 06 originally shipped watch mode, pack export, budget fitting, and doctor
diagnostics in one phase. That historical execution record is still valid evidence, but
it is not the right ownership boundary for current milestone closure.

Phase 7 explicitly repaired doctor health semantics and milestone state alignment after
Phase 6 completed. For closure purposes, the current source-of-truth split is:

- Phase 06 verification owns `OPS-02`, `OPS-03`, and `OPS-04`.
- Phase 07 verification owns the repaired doctor and optional-watch closure behavior for
  `CLI-05`, `OPS-01`, and `OPS-05`.

That boundary matters because it prevents a Phase 6 backfill artifact from claiming
current ownership over behaviors whose current closure truth now lives in Phase 7.
This report therefore references Phase 7 only to explain exclusions, not to re-prove
Phase 7 behavior.

## Current Command Truth

The current automated verification command path for this report is:

```sh
env GOCACHE=/tmp/optimusctx-gocache \
  GOMODCACHE=/tmp/optimusctx-gomodcache \
  GOPROXY=off \
  /usr/local/go/bin/go test ./... -run 'TestWatch|TestWatchCommand|TestWatchCommandErrors|TestWatchRunnerLifecycle|TestWatchRefreshUsesCanonicalPipeline|TestWatchDebouncesBurstEvents|TestWatchOverflowFallsBackToFullRefresh|TestWatchUncertainEventFallsBackToFullRefresh|TestWatchRefreshFailureRecovery|TestWatchStatusStaleHeartbeat|TestRefreshReasonWatch|TestPackExportManifest|TestPackExportWritesPortableArtifact|TestPackExportBudgetPolicy|TestPackExportFitsTargetBudget|TestPackExportFilterRules|TestPackExportCommand|TestPackExportCommandBudgetFlags|TestPackExportCommandErrors'
```

This is the command contract declared by Phase 08 plan `08-04`. Unlike the earlier
Phase 02 and Phase 05 backfills, this plan explicitly requires the `/tmp`
`GOMODCACHE` path, so this report records that command as the current verification
truth for this phase.

## Requirement Verification

### OPS-02: Watch mode reuses the canonical incremental refresh pipeline

Status: satisfied

Why:

- `06-02-SUMMARY.md` established that watch-triggered refreshes do not invent a second
  refresh engine. They debounce advisory file-system hints and then delegate to the
  same canonical refresh service used by manual refresh.
- `internal/app/watch_test.go::TestWatchRefreshUsesCanonicalPipeline` proves the watch
  service passes `Reason=watch`, preserves the repository root, avoids forced full
  refresh when safe hints exist, and forwards normalized changed-path hints into the
  shared refresh request.
- `internal/app/watch_test.go::TestWatchDebouncesBurstEvents` proves multiple noisy
  events collapse into one canonical refresh request with deduplicated, sorted hints.
- `internal/app/watch_test.go::TestWatchOverflowFallsBackToFullRefresh` and
  `TestWatchUncertainEventFallsBackToFullRefresh` prove correctness wins over partial
  guesswork: overflow or uncertainty clears the hints and forces a normal full
  canonical refresh.
- `internal/app/watch_test.go::TestWatchRefreshFailureRecovery` proves watch can surface
  a failed refresh attempt and then recover on a later successful canonical refresh.
- `internal/app/refresh_test.go::TestRefreshReasonWatch` proves the shared refresh
  pipeline persists `reason=watch` and sanitized `changed_hint` metadata in
  `refresh_runs`, so watch semantics are carried through the existing refresh truth
  rather than stored in a side channel.
- `internal/cli/watch_test.go::TestWatchCommand` and
  `internal/cli/watch_test.go::TestWatchCommandErrors` prove the operator-facing watch
  command stays a thin CLI boundary over that same shared runtime.

Evidence anchors:

- `TestWatchRunnerLifecycle`
- `TestWatchRefreshUsesCanonicalPipeline`
- `TestWatchDebouncesBurstEvents`
- `TestWatchOverflowFallsBackToFullRefresh`
- `TestWatchUncertainEventFallsBackToFullRefresh`
- `TestWatchRefreshFailureRecovery`
- `TestWatchStatusStaleHeartbeat`
- `TestRefreshReasonWatch`
- `TestWatchCommand`
- `TestWatchCommandErrors`

Result:

`OPS-02` is currently satisfied. Watch events remain advisory hints, and the actual
repository update path is the same canonical refresh pipeline used outside watch mode.

### OPS-03: Pack export produces a compact portable repository artifact

Status: satisfied

Why:

- `06-03-SUMMARY.md` established pack export as a transport-neutral service layered on
  top of the existing pack pipeline rather than a second retrieval engine.
- `internal/app/pack_export_test.go::TestPackExportManifest` proves export creates a
  deterministic manifest with repository metadata, freshness, export format,
  compression mode, included sections, truncation flags, and stable JSON output across
  repeated reads.
- `internal/app/pack_export_test.go::TestPackExportWritesPortableArtifact` proves the
  export service can stream JSON to stdout for shell pipelines and can also write a
  gzip-compressed artifact file without changing the underlying manifest contract.
- `internal/cli/pack_test.go::TestPackExportCommand` proves the CLI delegates to the
  export service and prints an operator-facing summary only when the artifact is written
  to an explicit output path, leaving stdout clean for raw artifact streaming.
- `internal/cli/pack_test.go::TestPackExportCommandErrors` proves unsupported formats,
  missing values, unknown flags, and downstream service errors all fail explicitly at
  the CLI boundary.

Evidence anchors:

- `TestPackExportManifest`
- `TestPackExportWritesPortableArtifact`
- `TestPackExportCommand`
- `TestPackExportCommandErrors`

Result:

`OPS-03` is currently satisfied. The current codebase still exports deterministic,
portable repository packs for offline or non-MCP workflows through one app-layer
service and one thin CLI surface.

### OPS-04: Pack export fits target budgets with include and exclude rules

Status: satisfied

Why:

- `06-04-SUMMARY.md` established budget fitting as a deterministic second pass over
  pack export output rather than a change to the underlying retrieval model.
- `internal/app/pack_export_test.go::TestPackExportBudgetPolicy` proves the policy
  contract records the shared `bytes_div_4_ceiling` estimate policy, preserves the
  requested target token budget, and filters exported content using explicit include
  and exclude paths.
- The same test also proves the artifact explains what was dropped: excluded and
  out-of-scope paths appear in omission metadata instead of silently disappearing.
- `internal/app/pack_export_test.go::TestPackExportFitsTargetBudget` proves the export
  summary can report that the artifact fits the requested budget after deterministic
  pruning.
- `internal/app/pack_export_test.go::TestPackExportFilterRules` proves include and
  exclude rules are applied consistently to structural content and manifest reporting.
- `internal/cli/pack_test.go::TestPackExportCommandBudgetFlags` proves the CLI parses
  repeated `--include`, repeated `--exclude`, and `--target-budget` flags into the same
  typed export policy consumed by the app layer.
- `internal/cli/pack_test.go::TestPackExportCommandErrors` proves invalid
  `--target-budget` input fails with an explicit positive-integer requirement rather
  than producing ambiguous behavior.

Evidence anchors:

- `TestPackExportBudgetPolicy`
- `TestPackExportFitsTargetBudget`
- `TestPackExportFilterRules`
- `TestPackExportCommandBudgetFlags`
- `TestPackExportCommandErrors`

Result:

`OPS-04` is currently satisfied. Export fitting remains explicit, deterministic, and
auditable through manifest-level omission reporting and the shared token-estimate
policy.

## Excluded Requirements

The following identifiers appear here only to document the enforced boundary:

- `OPS-01`: optional watch-mode health ownership moved to Phase 7 closure evidence.
- `CLI-05`: doctor command verification ownership moved to Phase 7 closure evidence.
- `OPS-05`: doctor diagnostics and token-cost reporting ownership moved to Phase 7
  closure evidence.

Phase 7 summary `07-01-SUMMARY.md` is the current closure source for those repairs.
They are intentionally excluded from this Phase 6 backfill artifact so the current
milestone record does not blur repaired Phase 7 behavior back into Phase 6.

## Phase Goal Verification

Phase 06 goal for current closure purposes: prove watch-refresh reuse, portable pack
export, and target-budget fitting with current automated evidence, while keeping doctor
and optional-watch closure semantics in Phase 7.

Result: satisfied

Why:

- Watch mode reuses the canonical refresh pipeline and records watch refresh metadata
  through the existing refresh persistence contract.
- Pack export still produces deterministic portable artifacts suitable for offline and
  non-MCP workflows.
- Budget fitting and include or exclude rules remain explicit, deterministic, and
  visible in exported metadata.
- The report keeps the corrected ownership boundary by excluding `CLI-05`, `OPS-01`,
  and `OPS-05` from Phase 6 proof.

## Success Criteria Verification

### `06-VERIFICATION.md` exists and stays bounded to `OPS-02`, `OPS-03`, and `OPS-04`

Satisfied. This artifact verifies only those three surviving Phase 6 requirements and
states the Phase 7 exclusion boundary explicitly.

### Current tests anchor watch reuse, pack export, and budget-fitting behavior

Satisfied. The current test anchors named above cover canonical watch refresh reuse,
debounced hint handling, uncertainty fallback, portable export artifacts, budget fit,
filter rules, and CLI command boundaries.

### The Phase 7 ownership repair is preserved instead of leaked back into Phase 6

Satisfied. `CLI-05`, `OPS-01`, and `OPS-05` appear only as exclusions tied to Phase 7.
This report does not claim current doctor ownership or re-prove repaired Phase 7
behavior.

## Automated Verification Run

Passed:

- `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./... -run 'TestWatch|TestWatchCommand|TestWatchCommandErrors|TestWatchRunnerLifecycle|TestWatchRefreshUsesCanonicalPipeline|TestWatchDebouncesBurstEvents|TestWatchOverflowFallsBackToFullRefresh|TestWatchUncertainEventFallsBackToFullRefresh|TestWatchRefreshFailureRecovery|TestWatchStatusStaleHeartbeat|TestRefreshReasonWatch|TestPackExportManifest|TestPackExportWritesPortableArtifact|TestPackExportBudgetPolicy|TestPackExportFitsTargetBudget|TestPackExportFilterRules|TestPackExportCommand|TestPackExportCommandBudgetFlags|TestPackExportCommandErrors'`

Notes:

- This backfill uses the current `/usr/local/go/bin/go` toolchain path.
- The current Phase 08 plan specifies `GOMODCACHE=/tmp/optimusctx-gomodcache`, so this
  report records that exact command as the verification truth for this artifact.
- The report intentionally avoids doctor test anchors because those belong to the Phase
  7 ownership repair, not the surviving Phase 6 backfill scope.

## Final Verdict

Phase 06 is verified as `passed` for the current closure scope of `OPS-02`, `OPS-03`,
and `OPS-04`.

The missing milestone verification artifact now exists, the requirement boundary is
corrected to respect the Phase 7 repair ownership, and the remaining Phase 6 evidence
is anchored in current automated watch and pack tests rather than historical execution
order alone.
