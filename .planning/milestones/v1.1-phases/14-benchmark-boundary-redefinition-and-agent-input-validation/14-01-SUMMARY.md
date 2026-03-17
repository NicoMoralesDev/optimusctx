---
phase: 14-benchmark-boundary-redefinition-and-agent-input-validation
plan: "01"
subsystem: benchmarking
tags: [go, sqlite, benchmark, methodology, evidence, cli]
requires:
  - phase: 13-distribution-pipeline-and-adoption-plan
    provides: frozen benchmark/export surface and milestone-closeout baseline reused by the phase-14 contract repair
provides:
  - benchmark-suite@v2 schema with explicit counted-input and final-artifact contracts
  - benchmark-evidence@v2 persistence with attribution boundaries and methodology snapshots
  - service and CLI export/report paths that fingerprint suite-derived methodology and reject ambiguous semantics
affects: [14-02 runtime enforcement, 14-03 suite migration, benchmark reporting]
tech-stack:
  added: [none]
  patterns: [suite-derived methodology fingerprinting, agent-input-only token aggregation, explicit evidence boundary persistence]
key-files:
  created: [internal/store/migrations/0007_benchmark_boundary_v2.sql]
  modified: [internal/repository/benchmark.go, internal/repository/benchmark_test.go, internal/store/sqlite/benchmark.go, internal/store/sqlite/benchmark_test.go, internal/app/benchmark_service.go, internal/app/benchmark_service_test.go, internal/cli/eval.go, internal/cli/eval_integration_test.go]
key-decisions:
  - "Benchmark suites now validate only as optimusctx/benchmark-suite@v2 and must declare counted agent inputs plus structured comparable final artifacts."
  - "Benchmark evidence now persists a methodology snapshot and attribution boundary so counted agent input, system provenance, and final-artifact verification cannot be conflated."
  - "Export/report fingerprinting is derived from the suite methodology contract, while human summaries aggregate only agent-input attribution rows."
patterns-established:
  - "Benchmark boundary truthfulness: raw operational outputs stay provenance unless explicitly projected into a declared counted input."
  - "Methodology identity: suite schema, counted inputs, final-artifact contracts, and lane/step surfaces are fingerprinted from the suite contract rather than inferred from raw attribution totals."
requirements-completed: [BNCH-01, BNCH-02, BNCH-04]
duration: 59min
completed: 2026-03-16
---

# Phase 14 Plan 01: Benchmark Boundary Contract and Schema Upgrade Summary

**Benchmark suite v2 with declared agent-input counting, structured final-artifact comparability, and persisted methodology-boundary evidence**

## Performance

- **Duration:** 59 min
- **Started:** 2026-03-16T18:16:00Z
- **Completed:** 2026-03-16T19:14:50Z
- **Tasks:** 3
- **Files modified:** 9

## Accomplishments
- Replaced the old `completionArtifact` hint with a required v2 benchmark contract that declares counted inputs, suite boundary policy, and comparable final artifacts explicitly.
- Added `benchmark-evidence@v2` persistence for methodology snapshots, attribution boundaries, and lane-level final-artifact verification in SQLite.
- Updated export/report logic to fingerprint suite-derived methodology, reject ambiguous persisted evidence, and sum estimated tokens from counted agent-input rows only.

## Task Commits

Each task was committed atomically:

1. **Task 1: Introduce the typed v2 benchmark boundary and final-artifact schema** - `3c990b5` (feat)
2. **Task 2: Persist v2 boundary semantics and methodology identity deterministically** - `e43f314` (feat)
3. **Task 3: Update export and summary assembly to speak the v2 contract before runtime enforcement lands** - `e357b5a` (feat)

## Files Created/Modified

- `internal/repository/benchmark.go` - benchmark-suite v2 schema, counted-input/final-artifact contracts, evidence methodology snapshot, and attribution boundary types
- `internal/repository/benchmark_test.go` - repository-level validation coverage for v2 counted-input and final-artifact rules
- `internal/store/migrations/0007_benchmark_boundary_v2.sql` - SQLite migration for suite schema version, counted-input policies, attribution boundaries, and final-artifact verification fields
- `internal/store/sqlite/benchmark.go` - v2 bundle validation and persistence for methodology snapshots, final-artifact verification, and explicit attribution boundaries
- `internal/store/sqlite/benchmark_test.go` - deterministic SQLite coverage for v2 methodology and evidence persistence
- `internal/app/benchmark_service.go` - suite-derived methodology fingerprinting, ambiguous-evidence rejection, and agent-input-only report aggregation
- `internal/app/benchmark_service_test.go` - self-contained v2 service coverage for export, report wording, persisted rebuild, and methodology drift
- `internal/cli/eval.go` - `--suite-file` handling that no longer forces `SuitesDir` into benchmark export/report/verify requests
- `internal/cli/eval_integration_test.go` - CLI coverage for v2 suite-file export/report flows and methodology identity

## Decisions Made

- Suite validation is now explicitly v2-only to prevent silent reinterpretation of v1 token-counting semantics.
- Methodology snapshots persist the suite boundary contract, counted inputs, and final-artifact contracts so drift is inspectable without reading raw logs.
- Human summaries and attribution tables now ignore provenance-only records and report only counted agent-input token rows.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed CLI benchmark commands so `--suite-file` does not also pass `SuitesDir`**
- **Found during:** Task 3 (Update export and summary assembly to speak the v2 contract before runtime enforcement lands)
- **Issue:** CLI export/report tests could not exercise v2 suite files because the request builder always injected `SuitesDir`, which made valid suite-file requests fail before the service logic ran.
- **Fix:** Conditionalized `SuitesDir` assignment in the benchmark CLI command handlers and rewrote the integration tests to seed a v2 suite-file flow.
- **Files modified:** `internal/cli/eval.go`, `internal/cli/eval_integration_test.go`
- **Verification:** `go test ./internal/app ./internal/cli -run 'TestBenchmark(Export|Report|Verify|Methodology)'`
- **Committed in:** `e357b5a`

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** The fix was required to verify the planned v2 service behavior through the shipped CLI surface. No product-scope expansion was introduced.

## Issues Encountered

- `gofmt` was initially pointed at the SQL migration file; rerunning it on Go files only resolved the formatting step cleanly.
- Sandbox writes to the default Go build cache were denied, so verification used the existing `/tmp` cache pattern already established in this repository.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 14 now has one explicit schema foundation for counted agent inputs, provenance-only system work, final-artifact comparability, and methodology identity.
- Plan `14-02` can move into runtime attribution/projection enforcement without reopening contract ambiguity in the repository, store, service, or CLI layers.

## Self-Check

PASSED

---
*Phase: 14-benchmark-boundary-redefinition-and-agent-input-validation*
*Completed: 2026-03-16*
