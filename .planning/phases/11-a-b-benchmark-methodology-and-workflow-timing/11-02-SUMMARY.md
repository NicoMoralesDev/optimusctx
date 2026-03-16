---
phase: 11-a-b-benchmark-methodology-and-workflow-timing
plan: "02"
subsystem: benchmarking
tags: [benchmarking, sqlite, mcp, cli, timing]
requires:
  - phase: 11-a-b-benchmark-methodology-and-workflow-timing
    provides: frozen benchmark suites, baseline action rules, and fixture-backed suite selection from 11-01
provides:
  - lane-level benchmark timing with explicit start, success, and stop markers
  - real-surface benchmark execution through shipped CLI and MCP boundaries
  - persisted benchmark runs, lane samples, and effort metrics for later repeated-run comparison
affects: [phase-11, phase-12, benchmark-reporting, repeated-run-comparison]
tech-stack:
  added: []
  patterns: [lane-first benchmark timing, stable-key context anchoring, per-arm benchmark persistence]
key-files:
  created:
    - internal/store/migrations/0005_benchmark_runs.sql
    - internal/store/sqlite/benchmark.go
    - internal/store/sqlite/benchmark_test.go
  modified:
    - internal/repository/benchmark.go
    - internal/app/benchmark_runner.go
    - internal/app/benchmark_runner_test.go
    - internal/cli/eval_integration_test.go
    - internal/mcp/integration_test.go
    - internal/store/migrations/runner_test.go
key-decisions:
  - "Treatment context-assembly requests reuse the stable symbol key returned by symbol lookup when available so MCP timing stays anchored to indexed product behavior instead of synthetic line guesses."
  - "Benchmark persistence stores one run row per suite arm attempt and separates lane samples from metric rows so repeated-run comparison stays queryable."
  - "The treatment workspace is bootstrapped with shipped init and refresh commands before timed lanes begin so discovery and context-assembly measurements reflect retrieval work rather than repository setup noise."
patterns-established:
  - "Benchmark lane results always carry explicit start-marker, success-marker, and stop-marker names alongside elapsed time and effort counters."
  - "Real-surface benchmark tests execute treatment actions through the same CLI and MCP seams already validated in Phase 10."
requirements-completed: [BNCH-01, BNCH-03]
duration: 10min
completed: 2026-03-16
---

# Phase 11 Plan 02: Benchmark Methodology and Workflow Timing Summary

**Discovery and context-assembly benchmark lanes now run with explicit timing boundaries, shipped-surface treatment execution, and persisted SQLite evidence per arm and lane**

## Performance

- **Duration:** 10 min
- **Started:** 2026-03-16T12:14:40Z
- **Completed:** 2026-03-16T12:24:37Z
- **Tasks:** 3
- **Files modified:** 9

## Accomplishments
- Extended the benchmark domain and runner to capture discovery and context-assembly lanes as separate timed activities with explicit start, success, and stop markers plus effort counters.
- Verified treatment execution through the shipped `init`, `refresh`, and `mcp serve` boundaries and added integration coverage that proves lane completion on real fixture workspaces.
- Added benchmark persistence tables and store code for benchmark runs, lane samples, and lane metrics so later repeated-run comparison can query structured evidence instead of logs.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add discovery and context-assembly lane definitions with measurable effort counters** - `ed3f405` (feat)
2. **Task 2: Run the discovery and context lanes through real shipped surfaces** - `4c34fc1` (feat)
3. **Task 3: Persist lane timing and effort evidence for later comparison** - `6ba7199` (feat)

## Files Created/Modified
- `internal/repository/benchmark.go` - Benchmark lane result contracts, marker helpers, and effort types shared across runner and persistence layers.
- `internal/app/benchmark_runner.go` - Workspace preparation, per-lane timing, baseline action accounting, treatment execution, and stable-key anchored context retrieval.
- `internal/app/benchmark_runner_test.go` - Coverage for lane markers and timing/effort capture.
- `internal/cli/eval_integration_test.go` - Real CLI plus `mcp serve` benchmark discovery-lane coverage.
- `internal/mcp/integration_test.go` - Real MCP server benchmark context-assembly coverage.
- `internal/store/migrations/0005_benchmark_runs.sql` - SQLite schema for benchmark runs, lane samples, and lane metrics.
- `internal/store/sqlite/benchmark.go` - Benchmark run persistence and runner-to-store translation helpers.
- `internal/store/sqlite/benchmark_test.go` - Persistence coverage for lane samples and effort metrics.
- `internal/store/migrations/runner_test.go` - Migration expectations updated for schema version 5 and benchmark indexes.

## Decisions Made

- Reused the stable symbol key from `symbol_lookup` for `targeted_context` whenever available so treatment timing stops on machine-checkable indexed anchors.
- Stored benchmark identity at the run and lane tables instead of only in `metadata_json`, which keeps suite, arm, attempt, and lane comparison queryable.
- Kept workspace bootstrap outside timed lanes so the benchmark measures discovery and context assembly rather than one-time repository setup.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Replaced synthetic targeted-context line guesses with stable-key anchoring**
- **Found during:** Task 2 (Run the discovery and context lanes through real shipped surfaces)
- **Issue:** The real MCP `targeted_context` call rejected the synthetic `1-40` line window because it did not satisfy the indexed anchor on the live fixture file.
- **Fix:** Threaded the stable symbol key captured from `symbol_lookup` into later treatment context requests and kept a minimal explicit-line fallback only for synthetic test doubles.
- **Files modified:** `internal/app/benchmark_runner.go`
- **Verification:** `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmarkDiscoveryLane|TestBenchmarkContextAssemblyLane'`
- **Committed in:** `4c34fc1`

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** The fix kept the runner aligned with the shipped MCP contract and prevented benchmark-only retrieval shortcuts.

## Issues Encountered

- Direct `go test` invocations hit the same sandboxed cache and stdlib resolution issues documented in earlier phases. Verification succeeded with `/usr/local/go/bin/go` plus `GOCACHE` and `GOMODCACHE` rooted in `/tmp`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 11 now has persisted evidence for the first two workflow-speed lanes, so Plan 11-03 can add refresh-after-change and task-completion timing on the same runner and store contracts.
- Phase 12 can build comparison and reporting layers on top of queryable benchmark run, lane, and metric tables instead of parsing ad hoc logs.

## Self-Check: PASSED

- Verified key files exist on disk: `internal/app/benchmark_runner.go`, `internal/store/migrations/0005_benchmark_runs.sql`, `internal/store/sqlite/benchmark.go`, `.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-02-SUMMARY.md`
- Verified task commits exist in Git history: `ed3f405`, `4c34fc1`, `6ba7199`
