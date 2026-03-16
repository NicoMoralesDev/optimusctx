---
phase: 12-token-attribution-and-evidence-reporting
plan: "02"
subsystem: benchmarking
tags: [benchmarking, sqlite, cli, evidence-export, token-attribution]
requires:
  - phase: 12-01
    provides: canonical token attribution records and persistence inputs for benchmark lanes
provides:
  - deterministic schema-versioned benchmark evidence bundles
  - persisted derived benchmark evidence metadata and attribution rows
  - shipped `eval benchmark export` CLI path with rerun contract
affects: [benchmark-reporting, evidence-review, rerunability]
tech-stack:
  added: []
  patterns: [persist-derived-evidence-bundles, cli-export-backed-by-stored-runs]
key-files:
  created: [internal/store/migrations/0006_benchmark_evidence.sql, internal/app/benchmark_service_test.go]
  modified: [internal/repository/benchmark.go, internal/store/sqlite/benchmark.go, internal/store/sqlite/benchmark_test.go, internal/app/benchmark_service.go, internal/cli/eval.go, internal/cli/eval_test.go, internal/cli/eval_integration_test.go]
key-decisions:
  - "Benchmark evidence exports are rebuilt from persisted benchmark runs and attribution metadata rather than ad hoc in-memory terminal output."
  - "The rerun contract is a real shipped CLI path: `optimusctx eval benchmark export --suite|--suite-file ... --attempts N`."
  - "Derived evidence persistence stores the canonical bundle JSON plus queryable lane-summary and attribution rows on top of raw benchmark run tables."
patterns-established:
  - "Benchmark exports use schema-versioned repository contracts with deterministic normalization before persistence or JSON emission."
  - "CLI benchmark export commands reuse the shipped CLI and MCP execution seams so integration tests exercise the real operator path."
requirements-completed: [BNCH-02, BNCH-04]
duration: 33min
completed: 2026-03-16
---

# Phase 12 Plan 02: Benchmark Evidence Export Summary

**Schema-versioned benchmark evidence bundles with persisted attribution summaries and a shipped `eval benchmark export` rerun path**

## Performance

- **Duration:** 33 min
- **Started:** 2026-03-16T13:33:00Z
- **Completed:** 2026-03-16T14:06:11Z
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments
- Added a canonical benchmark evidence bundle contract in `internal/repository` with deterministic normalization, methodology identity, estimator policy, rerun metadata, per-attempt provenance, and comparison summaries.
- Added SQLite migration `0006_benchmark_evidence.sql` plus store save/load support for derived evidence bundles, lane summaries, and per-attempt attribution rows without changing the raw benchmark run tables.
- Shipped `optimusctx eval benchmark export` with app-layer export generation from persisted runs and end-to-end CLI tests proving deterministic JSON shape, methodology identity, report labels, and rerun-command correctness.

## Task Commits

Each task was committed atomically:

1. **Task 1: Define the benchmark evidence bundle schema and derived persistence layer** - `4f57729` (feat)
2. **Task 2: Generate and persist evidence bundles from repeated benchmark runs** - `13451ee` (feat)
3. **Task 3: Prove the export format through the shipped benchmark boundary** - `899fda0` (test)

**Plan metadata:** pending

## Files Created/Modified
- `internal/repository/benchmark.go` - Defines the benchmark evidence bundle schema, normalization helpers, and deterministic JSON export contract.
- `internal/store/migrations/0006_benchmark_evidence.sql` - Adds derived evidence, lane-summary, and attribution persistence tables plus supporting indexes.
- `internal/store/sqlite/benchmark.go` - Saves and loads benchmark evidence bundles and persists queryable derived export records.
- `internal/store/sqlite/benchmark_test.go` - Verifies migration coverage, evidence bundle persistence, and deterministic export reload behavior.
- `internal/app/benchmark_service.go` - Rebuilds evidence bundles from persisted benchmark attempts and emits a real rerun command contract.
- `internal/app/benchmark_service_test.go` - Covers persisted export generation, comparison export shape, and rerun command construction.
- `internal/cli/eval.go` - Adds `eval benchmark export` parsing, output writing, and real CLI/MCP runner wiring.
- `internal/cli/eval_test.go` - Covers benchmark export request routing through the CLI boundary.
- `internal/cli/eval_integration_test.go` - Proves benchmark evidence generation through the shipped CLI path and validates methodology identity plus report labels.

## Decisions Made
- Persisted benchmark evidence is the portable source of truth and is rebuilt from stored benchmark rows plus attribution metadata rather than assembled from transient run output.
- The benchmark rerun contract now points at a real shipped export path instead of a placeholder test command, making reruns operable from the CLI surface.
- Derived evidence records are stored both as canonical bundle JSON and as flattened lane-summary and attribution tables so future reporting and diffing do not need bespoke re-aggregation.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Wired benchmark export through the real CLI and MCP executors**
- **Found during:** Task 3 (Prove the export format through the shipped benchmark boundary)
- **Issue:** `eval benchmark export` instantiated `BenchmarkService` without the CLI and MCP execution hooks, so the real export path failed with `benchmark command executor is not configured`.
- **Fix:** Configured the command-path benchmark service to reuse `executeEvalCLICommand` and `executeEvalCLIMCPSession` before rerunning the end-to-end export tests.
- **Files modified:** `internal/cli/eval.go`
- **Verification:** `go test ./internal/app ./internal/cli -run 'TestBenchmarkEvidenceBundleGeneration|TestBenchmarkExportContainsMethodologyIdentity|TestBenchmarkExportCLIPath'`
- **Committed in:** `899fda0` (part of task commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** The fix was required for the shipped export surface to work through the real CLI boundary. No scope creep.

## Issues Encountered
- Initial sandbox Go settings pointed at a non-writable build cache, so verification used `/usr/local/go/bin/go` with `/tmp` `GOCACHE` and `GOMODCACHE`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Benchmark exports are now deterministic, schema-versioned, and persisted, so later report-rendering work can read a stable evidence bundle instead of rebuilding aggregates ad hoc.
- The CLI rerun contract and derived SQLite records are in place for later comparison, archiving, and human-readable reporting work.

## Self-Check

PASSED

---
*Phase: 12-token-attribution-and-evidence-reporting*
*Completed: 2026-03-16*
