---
phase: 11-a-b-benchmark-methodology-and-workflow-timing
plan: "03"
subsystem: testing
tags: [benchmarking, cli, mcp, sqlite, eval]
requires:
  - phase: 11-02
    provides: discovery/context benchmark timing, persisted benchmark run identities, and real-surface CLI/MCP benchmark execution seams
provides:
  - deterministic refresh-after-change benchmark lane setup and assertions
  - end-to-end task-completion benchmark execution through CLI and MCP surfaces
  - persisted benchmark lane metadata for mutations, assertions, and evidence paths
affects: [11-04, benchmarking, reporting]
tech-stack:
  added: []
  patterns: [lane-scoped setup via eval actions, per-arm isolated benchmark workspaces, sqlite metadata persistence for lane evidence]
key-files:
  created: [.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-03-SUMMARY.md]
  modified: [internal/repository/benchmark.go, internal/app/benchmark_runner.go, internal/app/benchmark_runner_test.go, internal/cli/eval_integration_test.go, internal/mcp/integration_test.go, internal/store/sqlite/benchmark.go, internal/store/sqlite/benchmark_test.go, testdata/eval/benchmarks/go-benchmark-refresh-v1.json]
key-decisions:
  - "Refresh-after-change and task-completion lanes reuse eval setup actions so both arms apply the same committed repository mutation before timing begins."
  - "Each benchmark arm now runs in its own copied workspace so baseline and OptimusCtx timings cannot contaminate each other through shared state."
  - "Mutation-lane persistence stores setup, assertions, and evidence paths in SQLite metadata instead of introducing report-specific tables before Phase 12."
patterns-established:
  - "Benchmark lanes carry deterministic setup and machine-checkable assertions at the suite-contract layer."
  - "CLI benchmark steps precreate declared output directories before command execution, matching eval artifact handling."
requirements-completed: [BNCH-01, BNCH-03]
duration: 14min
completed: 2026-03-16
---

# Phase 11 Plan 03: Refresh/Completion Benchmark Summary

**Deterministic refresh-after-change and task-completion benchmark lanes with shared mutations, real CLI/MCP execution, and persisted lane evidence metadata**

## Performance

- **Duration:** 14 min
- **Started:** 2026-03-16T12:26:00Z
- **Completed:** 2026-03-16T12:40:11Z
- **Tasks:** 3
- **Files modified:** 9

## Accomplishments

- Extended benchmark suite contracts and runner behavior to apply fixture-backed repository mutations before timing refresh-after-change lanes and to enforce machine-checkable lane assertions.
- Executed the refresh and task-completion comparison workflow on the shipped CLI and MCP surfaces using the committed benchmark corpus.
- Persisted mutation-lane setup, assertion, and evidence metadata in SQLite so later rerun/reporting work can compare runs without rediscovering intent from raw timings alone.

## Task Commits

Each task was committed atomically:

1. **Task 1: Define refresh-after-change and task-completion lane contracts** - `bcae913` (feat)
2. **Task 2: Execute post-change and completion lanes through baseline and OptimusCtx arms** - `4937f6b` (feat)
3. **Task 3: Persist mutation-lane and completion-lane evidence** - `8df624a` (feat)

## Files Created/Modified

- `internal/repository/benchmark.go` - Added lane setup/assertion contracts and richer arm/lane result metadata.
- `internal/repository/benchmark_test.go` - Validated mutation-bearing benchmark suite contracts.
- `internal/app/benchmark_runner.go` - Isolated arm workspaces, applied lane setup before timing, enforced assertions, and prepared CLI artifact paths.
- `internal/app/benchmark_runner_test.go` - Added focused refresh-after-change and task-completion lane coverage.
- `testdata/eval/benchmarks/go-benchmark-refresh-v1.json` - Declared deterministic docs mutation, shared answer contract, and CLI/MCP completion steps.
- `internal/cli/eval_integration_test.go` - Verified refresh-after-change comparison through the shipped CLI and MCP harness.
- `internal/mcp/integration_test.go` - Verified task-completion comparison through the shipped MCP server surface.
- `internal/store/sqlite/benchmark.go` - Serialized workspace, setup, assertion, and evidence metadata into persisted benchmark records.
- `internal/store/sqlite/benchmark_test.go` - Proved mutation-lane metadata round-trips into SQLite benchmark persistence.

## Decisions Made

- Refresh and completion benchmark semantics now live in the benchmark-suite contract instead of being inferred from individual steps.
- Treatment-arm bootstrapping and execution happen in per-arm workspaces so the benchmark methodology stays fair across baseline and OptimusCtx lanes.
- Phase 11 persistence stays metadata-based inside existing benchmark tables; richer report shaping remains deferred to Phase 12 as planned.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Isolated benchmark arms into separate workspaces**
- **Found during:** Task 1 (Define refresh-after-change and task-completion lane contracts)
- **Issue:** Both arms shared one materialized workspace, allowing treatment bootstrap state to leak into baseline measurements.
- **Fix:** Copied a fresh workspace per arm and bootstrapped only the OptimusCtx arm before execution.
- **Files modified:** `internal/app/benchmark_runner.go`
- **Verification:** `go test ./internal/repository ./internal/app -run 'TestBenchmarkRefreshAfterChangeLane|TestBenchmarkTaskCompletionLane'`
- **Committed in:** `bcae913`

**2. [Rule 2 - Missing Critical] Precreated benchmark CLI artifact directories**
- **Found during:** Task 2 (Execute post-change and completion lanes through baseline and OptimusCtx arms)
- **Issue:** `pack export --output artifacts/pack.json` failed because benchmark execution did not prepare parent directories the way eval execution already does.
- **Fix:** Added benchmark CLI artifact-path preparation and richer nonzero-exit diagnostics before command execution.
- **Files modified:** `internal/app/benchmark_runner.go`
- **Verification:** `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmarkRefreshAfterChangeComparison|TestBenchmarkTaskCompletionComparison'`
- **Committed in:** `4937f6b`

---

**Total deviations:** 2 auto-fixed (1 bug, 1 missing critical)
**Impact on plan:** Both fixes were required for methodological correctness and for the real CLI completion lane to execute. No scope creep beyond benchmark correctness.

## Issues Encountered

- Default sandbox Go settings attempted to write to the user cache and produced unreliable stdlib/cache errors. Verification used `/usr/local/go/bin/go` with `GOCACHE` and `GOMODCACHE` redirected into `/tmp`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 11 now has all four benchmark lanes with deterministic setup and persisted evidence, so 11-04 can compare repeated runs without redefining lane semantics.
- SQLite benchmark metadata now includes the mutation and evidence context required for rerun comparison and later reporting.

## Self-Check: PASSED

---
*Phase: 11-a-b-benchmark-methodology-and-workflow-timing*
*Completed: 2026-03-16*
