---
phase: 14-benchmark-boundary-redefinition-and-agent-input-validation
research_date: 2026-03-16
objective: "What do I need to know to PLAN this phase well?"
status: complete
---

# Phase 14 Research: Benchmark Boundary Redefinition and Agent-Input Validation

## Executive Summary

Phase 14 is a benchmark-contract correction phase, not a product-expansion phase.

The current benchmark stack is useful for diagnosis, but it still measures the wrong boundary for the reopened milestone goal:

- treatment token totals are computed from raw treatment attribution records, including operational `refresh` and `health` payloads when they are recorded
- the suite schema validates `completionArtifact`, but the runner never enforces a comparable final artifact at runtime
- lane success is still mostly driven by stop markers plus lane assertions, not by proving both arms delivered the same final agent-usable artifact

The planning target should therefore be:

1. define an explicit benchmark boundary around counted agent-facing inputs
2. preserve internal system work as provenance and timing evidence, but exclude it from token totals unless the suite explicitly promotes a projected output into an agent input
3. require every benchmarked lane or task to validate one comparable final artifact contract
4. migrate the two frozen suites to the new contract and rerun all benchmark evidence

This phase should repair `BNCH-01`, `BNCH-02`, and `BNCH-04` methodology truthfulness. It should not absorb the separate product backlog from the fairness report, such as redesigning `repository_map` or shrinking `health` payloads.

## Inputs Reviewed

- `.planning/STATE.md`
- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
- `.planning/benchmark-fairness-report.md`
- `internal/app/benchmark_runner.go`
- `internal/app/benchmark_service.go`
- `internal/repository/benchmark.go`
- `internal/repository/benchmark_test.go`
- `internal/app/benchmark_runner_test.go`
- `internal/app/benchmark_service_test.go`
- `testdata/eval/benchmarks/go-benchmark-discovery-v1.json`
- `testdata/eval/benchmarks/go-benchmark-refresh-v1.json`
- `CLAUDE.md`: not present

## Current-State Diagnosis

### 1. The current token boundary is attribution-first, not agent-input-first

Today the runner records treatment attribution directly from raw CLI and MCP outputs:

- CLI output is counted in `benchmarkArmState.addCLIOutputAttribution`
- MCP payloads are counted in `benchmarkArmState.addToolAttribution`
- command and tool labels are rolled into human summaries by `benchmarkLaneEstimatedTokens` and `BuildBenchmarkHumanSummary`

That is why the current methodology can still charge treatment lanes for operational output:

- `refresh` is mapped to `BenchmarkArtifactTypeRefresh`
- `health` is mapped to `BenchmarkArtifactTypeHealth`
- both roll up under the human report label `operational`

This is consistent with the current code, but it is not the boundary the user reopened the milestone for. Phase 14 needs a distinction between:

- system work or provenance: what the runtime had to do
- counted agent inputs: what is actually considered benchmarked token cost

### 2. Comparable final artifacts are not enforced

`BenchmarkTaskDefinition` already has `CompletionArtifact`, and validation checks that it is a safe relative path. But runtime execution never consumes it:

- `BenchmarkTaskDefinition.validate` validates the path only
- `markLaneComplete`, `tryCompleteLane`, and `applyToolResult` do not require the artifact
- evidence comparison and methodology fingerprints do not include final-artifact equivalence

The result is a framework that can prove repeated stop-marker success, but not equal end-user deliverables.

### 3. Current suites rely on lane-specific heuristics rather than explicit comparable outputs

The two frozen suites illustrate the gap:

- `go-benchmark-discovery-v1` stops discovery on target identification and context assembly on `targeted_context` returning the target path
- `go-benchmark-refresh-v1` treats readiness as marker-plus-assertion success, then compares bounded context retrieval separately

Neither suite currently defines one normalized artifact per lane that both arms must produce or validate.

### 4. Existing tests encode the old boundary

Current tests lock in the behavior that Phase 14 intends to change:

- `internal/app/benchmark_runner_test.go` expects refresh-lane treatment attribution to include refresh output, a refresh marker record, and `health`
- `internal/app/benchmark_service_test.go` expects operational attribution rows in the exported comparison
- `internal/repository/benchmark_test.go` treats `CompletionArtifact` as a contract field, but only at schema-validation level

Planning must account for rewriting those expectations, not just adding new tests beside them.

## Target Benchmark Model

## 1. Counted token cost must be based on declared agent inputs

Recommended rule:

- a benchmark step may emit raw provenance
- only declared `agent input` projections contribute estimated tokens
- provenance remains persisted for auditability and timing interpretation
- the report must sum counted agent inputs only

That means a treatment step such as `refresh` or `health` can still exist in the workflow, but its full raw payload should not automatically count as token cost. It only counts if the suite declares a projected readiness summary as an agent input.

## 2. Every benchmark lane needs one comparable final artifact contract

Recommended rule:

- each benchmarked lane declares a normalized final artifact or completion contract
- both arms must either produce that artifact class or satisfy the same canonical expected artifact checks
- lane success requires both the stop condition and final-artifact validation

The contract should stay deterministic and machine-checkable. This phase does not need semantic free-form answer grading.

## 3. Normalize before comparing

Do not require raw baseline output and raw treatment output to look the same. Require them to normalize to the same artifact class.

Examples:

- discovery lane: normalized `{path, symbol}` target record
- context assembly lane: normalized bounded context bundle
- refresh-ready lane: normalized readiness summary `{freshness, generation, target-ready}`
- task-completion lane: normalized final bounded context artifact

This keeps the benchmark focused on comparable deliverables instead of identical transport payloads.

## 4. Keep product optimization separate from benchmark correction

The fairness report correctly identifies `repository_map` and `health` as expensive surfaces. That is product backlog. Phase 14 should only fix the benchmark contract so those costs are measured at the right boundary.

Do not turn this phase into:

- a `repository_map` redesign
- a `health_summary` product feature phase
- a tokenizer/provider-billing phase

## Schema and Runner Implications

## 1. Benchmark suite schema needs a contract bump

This is a semantic change, not a small interpretation tweak. Recommended direction:

- introduce `optimusctx/benchmark-suite@v2`
- add a corresponding evidence/export version bump if the persisted meaning of attribution changes

Why a version bump is warranted:

- v1 means "count raw treatment attribution"
- Phase 14 needs "count only declared agent-input projections"

Silently reinterpreting v1 would make old evidence misleading.

## 2. Add explicit agent-input declarations

The suite contract needs a typed way to say what counts.

Recommended minimum shape:

- step-level or lane-level `agentInput` declarations
- each declaration identifies:
  - source: CLI stdout, CLI artifact file, MCP payload projection, bounded file content, or other typed extractor
  - normalized artifact path or logical name
  - whether it counts toward token totals
  - artifact label or provenance label for reporting

Prefer typed extractors over arbitrary scripts or free-form expressions. The benchmark framework has intentionally avoided arbitrary shell freedom so far, and Phase 14 should preserve that discipline.

## 3. Promote `completionArtifact` into a real runtime contract

The current string field is too weak. Recommended direction:

- replace or extend it with a structured final-artifact contract
- allow canonical assertions against the normalized artifact
- require every benchmarked lane or the task as a whole to declare one

The runner should fail the lane if:

- the normalized final artifact is missing
- the artifact does not satisfy the canonical assertions
- one arm succeeds on stop markers but fails comparable-artifact validation

## 4. Separate provenance from counted attribution in the runner

The runner currently has one attribution stream. Phase 14 needs two concepts:

- provenance: raw workflow evidence, including operational steps and validation-only outputs
- counted agent inputs: projections that feed token totals

This can be modeled either as:

- separate collections, or
- one attribution type with an explicit boundary field such as `system_provenance`, `agent_input`, `final_artifact_validation`

The second option is probably the smaller code change because it preserves the existing storage flow while making the counting rule explicit.

## 5. Summary and evidence generation must aggregate counted inputs only

`benchmarkLaneEstimatedTokens`, human summaries, attribution tables, and reproducibility snapshots must switch from:

- "sum all treatment attribution"

to:

- "sum attribution records whose boundary is `agent_input`"

The methodology fingerprint and evidence comparison logic should also include the new boundary contract so reruns detect drift in what is counted.

## Migration Strategy For Existing Suites

## 1. Migrate the frozen corpus in-place after the new contract exists

There are only two benchmark suites. The cleanest path is:

1. implement the v2 contract
2. rewrite both JSON suites to v2
3. update tests and evidence exports
4. rerun the benchmark milestone on the new methodology

This is a better fit than carrying long-term dual-semantics support.

## 2. Discovery suite migration

`go-benchmark-discovery-v1` should gain explicit normalized outputs:

- discovery lane final artifact:
  - normalized target locator record containing the resolved file path and symbol
- context assembly lane final artifact:
  - normalized bounded context artifact for the required files or slices

Counted inputs should be projections, not raw transport blobs:

- baseline:
  - matched path or symbol lines
  - bounded file slices actually used
- treatment:
  - projected symbol match record
  - projected context bundle
  - `repository_map` only if the suite explicitly says the agent consumed a projected subset from it

## 3. Refresh suite migration

`go-benchmark-refresh-v1` should stop treating raw operational payloads as implicitly billable.

Recommended migration:

- refresh-ready lane final artifact:
  - normalized readiness summary or readiness marker artifact
- task-completion lane final artifact:
  - normalized updated-notes context artifact

The likely planning choice is:

- keep `refresh` and `health` as provenance and timing evidence
- count zero tokens for them unless the suite declares a small projected readiness summary as agent input
- continue counting the bounded context retrieval artifact in task completion

## 4. Regenerate persisted benchmark evidence

Because the boundary meaning changes, persisted bundles from the old methodology should not remain the active source of truth for the reopened milestone. Plan for full benchmark evidence regeneration after migration.

## Verification Needs

## Contract coverage

- suite validation rejects benchmark lanes without a final-artifact contract
- suite validation rejects counted agent-input declarations that have no typed extractor or no normalized target
- methodology fingerprints include the new boundary fields

## Runner coverage

- operational provenance can be recorded without affecting token totals
- counted agent-input projections do affect token totals
- lane success requires final-artifact validation, not only stop markers
- projected treatment artifacts remain deterministic across repeated runs

## Evidence/report coverage

- exported evidence distinguishes provenance from counted agent inputs
- human summaries and attribution tables aggregate counted inputs only
- reproducibility checks fail if counted boundary definitions drift

## Corpus/integration coverage

- both benchmark suites load and execute under the new schema
- the migrated suites produce the expected normalized final artifacts
- benchmark export, report, and verify flows remain truthful after the migration

## Likely Risks

- undercounting by projecting too little and accidentally excluding real agent-visible input
- overfitting the new typed extractors to the two current suites instead of defining reusable benchmark contracts
- mixing benchmark correction with product optimization work from the fairness report
- keeping old persisted evidence around without clearly separating pre-Phase-14 and post-Phase-14 methodology
- making final-artifact contracts too strict and brittle for deterministic reruns

## Recommended Plan Slices

## 14-01 Boundary contract and schema upgrade

Purpose:

- define the new benchmark boundary formally
- introduce suite and evidence schema updates
- add validation rules for agent-input declarations and final-artifact contracts

Outputs:

- v2 benchmark suite contract
- explicit attribution boundary model
- methodology fingerprint updates

## 14-02 Runner and service boundary enforcement

Purpose:

- teach the runner to project counted agent inputs
- enforce comparable final artifacts at runtime
- update service summaries and report generation to aggregate counted inputs only

Outputs:

- runner boundary enforcement
- updated export/report/verify behavior
- rewritten benchmark tests for the new counting rule

## 14-03 Corpus migration and benchmark evidence refresh

Purpose:

- migrate the two frozen suites to the new schema
- regenerate benchmark evidence under the corrected methodology

Outputs:

- migrated benchmark JSON corpus
- refreshed machine-readable evidence
- updated comparison expectations

## 14-04 Verification and milestone re-close

Purpose:

- prove the corrected benchmark contract is reproducible and truthful
- close the reopened benchmark milestone on the new methodology

Outputs:

- phase validation or verification artifact updates
- rerun benchmark reports with corrected wording and counts
- explicit note that old pre-Phase-14 evidence is superseded

## Planning Recommendation

Plan Phase 14 as a benchmark-framework repair across existing benchmark requirements, with explicit mapping to:

- `BNCH-01`: fixed A/B methodology on the same task
- `BNCH-02`: truthful token accounting boundary
- `BNCH-04`: reproducible evidence on the corrected contract

The key planning question is not "how do we make OptimusCtx look cheaper?" It is "what exact normalized inputs and outputs define a fair benchmark boundary?" The phase plan should answer that first, then let the rerun results say whatever they say.
