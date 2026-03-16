---
phase: 12-token-attribution-and-evidence-reporting
research_date: 2026-03-16
objective: "What do I need to know to PLAN this phase well?"
status: complete
---

# Phase 12 Research: Token Attribution and Evidence Reporting

## Executive Summary

Phase 12 should not invent a new benchmark runner. Phase 11 already established the hard part:

- frozen benchmark suites under `testdata/eval/benchmarks/`
- paired baseline vs OptimusCtx runs
- lane-scoped timing and effort metrics
- repeated-attempt orchestration in `internal/app/benchmark_service.go`
- persisted benchmark evidence in SQLite via `benchmark_runs`, `benchmark_lane_samples`, and `benchmark_lane_metrics`

The planning problem for Phase 12 is turning those persisted repeated runs into defensible evidence for `BNCH-02` and `BNCH-04`:

1. define one explicit token-accounting contract that attributes treatment-side savings to concrete OptimusCtx artifact types
2. persist derived benchmark evidence in a machine-readable format that can be regenerated from frozen suites and inputs
3. render truthful human-readable reports that explain both absolute results and attribution methodology
4. verify reproducibility so milestone claims are based on rerunnable artifacts, not one-off summaries

The main risk is attribution drift. If Phase 12 computes token savings from ad hoc report logic instead of a shared contract, machine exports, human summaries, and verification will disagree. The first plan must therefore define a canonical attribution model and the exact provenance each result must carry.

## Repository Reality Check

### Inputs reviewed

- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `.planning/ROADMAP.md`
- `README.md`
- `.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-RESEARCH.md`
- `.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-02-SUMMARY.md`
- `.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-03-SUMMARY.md`
- `.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-04-SUMMARY.md`
- `.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-VALIDATION.md`
- `internal/repository/benchmark.go`
- `internal/app/benchmark_runner.go`
- `internal/app/benchmark_service.go`
- `internal/store/sqlite/benchmark.go`
- `internal/store/migrations/0005_benchmark_runs.sql`
- `internal/app/pack_export.go`
- `internal/repository/pack_export.go`
- `internal/repository/token_tree.go`
- `internal/app/lookup.go`
- `testdata/eval/benchmarks/go-benchmark-discovery-v1.json`
- `testdata/eval/benchmarks/go-benchmark-refresh-v1.json`

### Local guidance files

- `CLAUDE.md`: not present
- `.claude/skills/`: not present
- `.agents/skills/`: not present

## What Already Exists From Phase 11

### Benchmark substrate that Phase 12 should reuse

- `internal/repository/benchmark.go` already defines the benchmark suite, arm, lane, metric, stop-condition, and run-result contracts.
- `internal/app/benchmark_runner.go` already records lane elapsed time, action counts, bytes read, and consulted artifact paths per arm.
- `internal/app/benchmark_service.go` already runs repeated attempts, summarizes per-lane stats, rejects methodology drift, and exposes a rerun command.
- `internal/store/sqlite/benchmark.go` already persists one row per suite arm attempt plus separate lane sample and lane metric rows.
- `README.md` already truthfully states that token attribution and richer reporting belong to Phase 12, not Phase 11.

### Existing signals Phase 12 can build on

- Treatment workflows already use concrete OptimusCtx surfaces in frozen suites:
  - `optimusctx.repository_map`
  - `optimusctx.symbol_lookup`
  - `optimusctx.targeted_context`
  - `optimusctx.health`
  - `optimusctx.pack`
  - `pack export`
  - `refresh`
- The repository already has a documented, shared token estimate policy: `bytes_div_4_ceiling`.
- Pack export already models deterministic machine-readable manifests with section records, estimates, omission reasons, and truncation metadata. That is the closest existing pattern for Phase 12 reporting artifacts.

### Constraint that matters most

`BNCH-02` requires one explicit milestone estimator and attribution by artifact type. That means Phase 12 must stay on the existing v1 token estimate policy unless the requirements change. Do not switch to model-specific tokenizers in this phase.

## What Phase 12 Must Decide Before Planning

## 1. What exactly is being attributed

The benchmark runner currently records:

- elapsed time
- action counts
- bytes read
- consulted artifact paths

It does not yet record canonical artifact-type attribution. Phase 12 needs a first-class model that answers:

- which OptimusCtx artifact type was used
- how many estimated tokens that artifact contributed
- which lane and attempt it belongs to
- whether the estimate reflects a direct artifact payload, a file-path-derived estimate, or an exported bundle section

Recommended artifact taxonomy for v1:

- `repository_map`
- `symbol_lookup`
- `structure_lookup`
- `targeted_context`
- `structural_context`
- `repository_context`
- `pack`
- `pack_export`
- `health`
- `refresh`

Not every type will carry direct token cost. `refresh` and `health` may be operational steps with zero direct token payload. Keep them in provenance if they matter to the workflow, but do not force fake token numbers onto non-content operations.

## 2. What “token savings by artifact type” means

The requirement wording is easy to overread. The phase does not need to prove literal model-billed tokens. It needs a deterministic estimate using one explicit estimator and attributable savings.

Recommended v1 definition:

- baseline token cost = estimated tokens for the bounded file content the baseline workflow had to read before satisfying the lane stop condition
- treatment token cost = estimated tokens of the OptimusCtx artifacts consumed before the same stop condition
- savings = `baseline_estimated_tokens - treatment_estimated_tokens`
- artifact-type attribution = treatment estimated tokens grouped by canonical OptimusCtx artifact type, plus optional derived percentage of total treatment token cost

This keeps the attribution claim honest:

- it compares workflow-consumed information volume, not provider invoices
- it uses the same documented estimator on both arms
- it explains where treatment-side token usage came from without pretending the baseline has typed artifacts

## 3. Where the token estimate should be computed

Do not duplicate token math in report generation.

Recommended contract:

- one shared benchmark-attribution layer computes estimated token usage
- SQLite persists the derived attribution records and report inputs
- machine exports and human reports both read the same derived records

The existing `bytes_div_4_ceiling` policy already appears in budget and token-tree services. Phase 12 should reuse one shared estimator source rather than re-encoding `ceil(bytes/4)` in multiple packages.

## 4. How consulted artifacts become typed evidence

Phase 11 stores artifact paths and suite step intent, but not typed attribution rows. Phase 12 should derive attribution from benchmark steps plus known result shapes, not from free-text parsing.

Recommended approach:

- extend benchmark result models so treatment lane results can carry structured artifact-consumption records
- each treatment step should emit zero or more artifact-consumption entries
- each entry should capture:
  - `artifact_type`
  - `surface`
  - `tool_or_command`
  - `path`
  - `estimated_tokens`
  - `source_kind`
  - `lane`
  - `step_id`

`source_kind` should distinguish:

- direct response payload estimate
- path-derived estimate from consulted repository files
- pack-export manifest-derived section estimate

That avoids later ambiguity when a report needs to explain whether `pack_export` tokens came from the exported section manifest or from a fallback path estimate.

## 5. What machine-readable artifact should be the source of truth

SQLite is necessary for local queryability, but `BNCH-04` also calls for reproducible artifacts. The source of truth for exchange and inspection should be a deterministic exported JSON artifact built from persisted runs and frozen suites.

Recommended output shape:

- one benchmark evidence bundle per repeated-run comparison
- top-level metadata:
  - schema version
  - generated-at timestamp
  - generator version
  - suite ID and version
  - fixture ID and version
  - repository root or fixture materialization identity
  - methodology fingerprint
  - rerun command
  - estimator policy
- sections:
  - repeated attempts
  - per-arm per-lane aggregates
  - per-attempt attribution records
  - reproducibility verification result
  - human-summary inputs

Use the pack export manifest style as the reference pattern: explicit policy, identity, summary, omission, and truncation fields.

## 6. What the human-readable report must and must not do

Human-readable reporting must explain results without inventing stronger claims than the data supports.

Recommended summary contents:

- suite and fixture identity
- attempt count and methodology fingerprint
- lane-by-lane timing comparison
- lane-by-lane token comparison
- treatment artifact attribution table
- invalid attempt or drift reasons
- rerun instructions
- caveats:
  - estimated tokens use `bytes_div_4_ceiling`
  - results reflect frozen benchmark suites only
  - artifact attribution is treatment-side attribution, not model billing telemetry

Do not add prose that implies:

- universal repository-level token savings
- exact provider token accounting
- statistically significant claims beyond the repeated attempts actually run

## Standard Stack

- Go only, inside the existing `internal/repository`, `internal/app`, and `internal/store/sqlite` layers
- existing benchmark runner and benchmark service as the orchestration entrypoint
- existing SQLite persistence model extended with derived attribution and report rows
- deterministic JSON export artifacts written under the repo-local `.optimusctx/eval/` evidence tree
- existing `bytes_div_4_ceiling` budget policy as the sole v1 estimator
- pack export manifest conventions as the model for explicit policy and summary metadata

## Architecture Patterns

## 1. Separate raw run evidence from derived reporting evidence

Keep the current Phase 11 raw benchmark evidence intact. Add a second layer for derived attribution and reporting outputs.

Recommended shape:

- raw evidence:
  - `benchmark_runs`
  - `benchmark_lane_samples`
  - `benchmark_lane_metrics`
- derived evidence:
  - benchmark artifact-attribution rows
  - benchmark comparison/export records
  - benchmark report bundles on disk

This avoids mixing runner concerns with report-shaping concerns.

## 2. Attribution should be step-scoped, then aggregated

Do not try to infer all attribution from lane totals. The planner should preserve enough detail to explain why a lane used tokens.

Recommended flow:

1. step emits artifact-consumption records
2. lane aggregates step records
3. attempt summary aggregates lane records
4. comparison summary aggregates attempts

This is the minimum granularity that keeps machine-readable exports and human summaries explainable.

## 3. Use explicit provenance fields everywhere

Every derived benchmark record should carry:

- suite ID
- suite version
- fixture ID
- arm kind
- attempt
- lane
- step ID if applicable
- estimator policy name
- methodology fingerprint

Without those fields, later reproducibility checks will become brittle joins over implicit assumptions.

## 4. Prefer deterministic export builders over ad hoc templates

The pack export code already demonstrates the right style:

- stable ordering
- explicit omission reasons
- explicit truncation reasons
- policy attached to the artifact

Phase 12 exports and summaries should follow the same pattern.

## 5. Reproducibility checks should compare normalized evidence, not timestamps

Repeated-run verification already rejects methodology drift. Phase 12 must add artifact reproducibility checks, but those checks should normalize unstable fields.

Recommended normalization:

- ignore generated-at timestamps
- ignore temp workspace paths
- keep suite, arm, attempt, lane, marker, metric, attribution, and estimator fields
- compare stable JSON payloads or a derived fingerprint

That matches the existing repository decision from Phase 9 to avoid byte-identical comparisons when artifacts contain temp-root paths.

## Don’t Hand-Roll

- Do not introduce model-specific tokenizer integrations in Phase 12.
- Do not create a brand-new benchmark CLI if the app-layer service can drive reporting.
- Do not parse human-readable logs to recover attribution; emit structured records directly.
- Do not compute token estimates in multiple report paths.
- Do not store only rendered Markdown and treat it as the canonical evidence.
- Do not redesign the benchmark suite schema more than needed for attribution provenance.

## Common Pitfalls

- Computing treatment attribution only from consulted file paths will miss tools like `repository_map` or `health` whose value is not a simple file slice.
- Attributing full pack export size to every contributing section will double count. Section-level accounting must deduplicate or clearly distinguish aggregate vs component totals.
- Mixing baseline effort metrics with treatment artifact attribution in one metric table without a type contract will make queries ambiguous.
- Comparing repeated export files byte-for-byte will fail on timestamps or temp workspace roots unless the export format is normalized.
- Letting report builders calculate their own totals independently of persisted derived records will create drift between JSON and Markdown outputs.
- Treating `refresh` as token-bearing content instead of an operational step will distort attribution.
- Expanding the taxonomy too far in v1 will create empty buckets and weak reports. Start with the artifact types already exercised in the frozen suites.

## Recommended Decomposition Into 4 Plans

## 12-01 token accounting contract and artifact attribution

Purpose:
Define the canonical attribution model and instrument the benchmark execution/reporting path so every treatment-side artifact consumption can be estimated and grouped by artifact type.

Scope:

- add typed repository models for benchmark attribution records and summaries
- extend benchmark runner and/or service outputs to carry structured artifact-consumption provenance
- reuse one shared token estimate policy
- persist derived attribution rows in SQLite

Likely files/modules:

- `internal/repository/benchmark.go`
- `internal/repository/budget.go`
- `internal/app/benchmark_runner.go`
- `internal/app/benchmark_service.go`
- `internal/store/sqlite/benchmark.go`
- `internal/store/sqlite/benchmark_test.go`
- new migration after `0005_benchmark_runs.sql`

Exit criteria:

- treatment benchmark runs emit typed artifact attribution records
- per-attempt and per-lane token totals can be derived deterministically
- attribution policy is explicit and shared

## 12-02 benchmark result storage and export format

Purpose:
Create a durable derived-evidence store and deterministic machine-readable export contract that packages repeated-run benchmark results for inspection and rerun.

Scope:

- define export schema
- add app-layer export builder
- write benchmark evidence bundles to the repo-local artifact tree
- ensure exports include estimator, methodology fingerprint, rerun command, and attribution details

Likely files/modules:

- `internal/repository/benchmark.go`
- possibly new `internal/repository/benchmark_report.go` or adjacent typed export file
- `internal/app/benchmark_service.go`
- possibly new `internal/app/benchmark_report.go`
- `internal/store/sqlite/benchmark.go`
- `internal/store/sqlite/benchmark_test.go`
- `README.md` only if rerun/report guidance changes in this phase

Exit criteria:

- one deterministic JSON export can be produced from persisted benchmark evidence
- export contains enough data to regenerate the human-readable summary without rereading raw logs
- export path and naming are deterministic under `.optimusctx/eval/`

## 12-03 human-readable benchmark summaries and comparison reports

Purpose:
Render milestone-grade human-readable summaries from the same derived evidence contract used by machine exports.

Scope:

- render Markdown or similarly reviewable text reports
- add clear lane comparisons and attribution tables
- include truthful caveats and rerun instructions
- avoid separate calculation logic from the machine export path

Likely files/modules:

- `internal/app/benchmark_service.go`
- new report builder file in `internal/app`
- benchmark repository/report model file in `internal/repository`
- tests for summary rendering and ordering
- `README.md` if operator guidance changes

Exit criteria:

- human-readable summary is generated from the same derived evidence as JSON
- report clearly explains attribution methodology and reproducibility status
- comparison output stays deterministic in ordering and wording

## 12-04 reproducibility checks and milestone verification

Purpose:
Prove the exported benchmark evidence is rerunnable, normalized, and trustworthy enough for milestone verification.

Scope:

- add normalized evidence comparison checks
- verify repeated exports from the same frozen suite stay equivalent after normalization
- add milestone verification/tests that cover `BNCH-02` and `BNCH-04`
- document truthful rerun commands and residual caveats

Likely files/modules:

- `internal/app/benchmark_service.go`
- new or expanded benchmark verification tests in `internal/app/benchmark_runner_test.go`
- `internal/cli/eval_integration_test.go`
- `internal/mcp/integration_test.go` only if shipped surfaces are part of reproducibility proof
- `internal/store/sqlite/benchmark_test.go`
- `.planning/phases/12-token-attribution-and-evidence-reporting/12-VALIDATION.md`
- `README.md`

Exit criteria:

- normalized repeated exports compare equal across reruns
- methodology drift and export drift fail loudly with actionable reasons
- milestone evidence maps directly to `BNCH-02` and `BNCH-04`

## Likely Files and Modules

High-confidence seams:

- `internal/repository/benchmark.go`
- `internal/app/benchmark_runner.go`
- `internal/app/benchmark_service.go`
- `internal/store/sqlite/benchmark.go`
- `internal/store/sqlite/benchmark_test.go`
- `internal/store/migrations/0005_benchmark_runs.sql` as the pattern reference for the next migration
- `testdata/eval/benchmarks/*.json`

Reference patterns to reuse rather than copy blindly:

- `internal/repository/pack_export.go`
- `internal/app/pack_export.go`
- `internal/repository/budget.go`
- `internal/repository/token_tree.go`

Potential new files:

- `internal/repository/benchmark_report.go`
- `internal/app/benchmark_report.go`
- new SQLite migration for benchmark attribution/report tables

## Dependencies

Hard dependencies:

- Phase 11 benchmark suites, arm rules, lane markers, and repeated-run service
- existing v1 estimator policy from budget/token-tree work
- repo-local `.optimusctx` evidence layout from the eval harness

Soft dependencies:

- pack export manifest/report conventions for deterministic machine-readable artifacts
- README benchmark rerun guidance as the truthful operator-facing contract

## Implementation Constraints

- Stay within `BNCH-02` and `BNCH-04`; do not reopen Phase 11 methodology design.
- Use exactly one explicit milestone estimator. For v1, that should remain `bytes_div_4_ceiling`.
- Keep benchmark evidence reproducible from frozen suites and fixture inputs.
- Avoid any dependence on hosted telemetry, external dashboards, or provider billing APIs.
- Keep report generation deterministic in ordering and schema.
- Preserve Phase 11 truth: no new public benchmark command is required unless the existing service boundary becomes clearly insufficient.
- Do not contaminate benchmark evidence with mutable working-tree state outside the fixture materialization and per-arm workspaces Phase 11 already controls.

## Validation Architecture

Phase 12 should stay Nyquist-friendly by validating each new seam at the layer where it can fail first, then proving end-to-end artifact reproducibility.

### Wave 0 files that should exist by the end of the phase

- `internal/repository/benchmark_test.go`
- `internal/app/benchmark_runner_test.go`
- `internal/store/sqlite/benchmark_test.go`
- tests covering report/export builders in `internal/app`
- `.planning/phases/12-token-attribution-and-evidence-reporting/12-VALIDATION.md`

### Recommended verification map by plan

`12-01`

- repository/app tests for attribution contracts and token-estimate derivation
- store tests for attribution row persistence and stable ordering

`12-02`

- app/store tests for deterministic JSON export shape
- tests proving exported evidence includes methodology fingerprint, estimator, and rerun command

`12-03`

- report rendering tests for deterministic section ordering and truthful caveat text
- comparison-summary tests proving JSON and human report read from the same derived source

`12-04`

- repeated-run integration tests proving normalized evidence equivalence across reruns
- methodology-drift and export-drift rejection tests
- README or docs verification if rerun/report instructions change

### Manual-only checks worth planning

- confirm the human-readable report never blurs estimated tokens with exact provider token billing
- confirm artifact-type labels map to real shipped OptimusCtx artifacts users can recognize
- confirm invalid attempts and caveats remain visible in rendered summaries

## Requirement Coverage

### `BNCH-02`: explicit milestone estimator and artifact attribution

Phase 12 must prove:

- one explicit estimator policy is named in evidence artifacts
- treatment-side token usage is attributable to concrete artifact types
- savings can be compared against baseline workflow token estimates

### `BNCH-04`: machine-readable and human-readable reproducible artifacts

Phase 12 must prove:

- repeated-run evidence can be exported deterministically
- summaries can be regenerated from the same persisted evidence
- rerun instructions and methodology fingerprints are included so others can inspect and reproduce the claim

## Planning Recommendations

- Make `12-01` the contract plan. If attribution types and provenance are weak, the rest of the phase will drift.
- Keep SQLite as the local query store, but make deterministic JSON exports the canonical portable evidence.
- Reuse pack export’s manifest style for benchmark evidence bundles.
- Treat human-readable reporting as a pure rendering layer over derived evidence, not a second calculator.
- Define normalization rules explicitly in `12-04` before writing reproducibility assertions.

## Code Examples

Use these files as implementation anchors when planning:

- `internal/app/benchmark_service.go`: repeated-run summary and methodology fingerprint patterns
- `internal/store/sqlite/benchmark.go`: current raw benchmark evidence persistence and stable attempt ordering
- `internal/repository/pack_export.go`: deterministic machine-readable artifact contracts with policy metadata
- `internal/app/pack_export.go`: explicit summary/manifest builder style with omission and truncation reasons
- `testdata/eval/benchmarks/go-benchmark-discovery-v1.json`: discovery and context-assembly treatment artifacts
- `testdata/eval/benchmarks/go-benchmark-refresh-v1.json`: refresh, pack, and export treatment artifacts
