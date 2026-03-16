---
phase: 11-a-b-benchmark-methodology-and-workflow-timing
plan: "01"
subsystem: benchmarking
tags: [benchmarking, eval, fixtures, methodology, mcp]
requires:
  - phase: 10-functional-runtime-validation
    provides: persisted eval fixtures, runner seams, and repository-local evidence patterns reused by the benchmark contract
provides:
  - canonical benchmark suite schema with paired baseline and OptimusCtx arms
  - app-layer benchmark suite loading with fixture-backed validation
  - frozen benchmark suite JSON definitions and a non-trivial benchmark fixture
affects: [phase-11, phase-12, benchmark-reporting, reproducibility]
tech-stack:
  added: []
  patterns: [json-backed benchmark suites, typed baseline action vocabulary, fixture-backed benchmark corpus]
key-files:
  created:
    - internal/repository/benchmark.go
    - internal/repository/benchmark_test.go
    - internal/app/benchmark_runner.go
    - internal/app/benchmark_runner_test.go
    - testdata/eval/benchmarks/go-benchmark-discovery-v1.json
    - testdata/eval/benchmarks/go-benchmark-refresh-v1.json
  modified: []
key-decisions:
  - "Baseline workflows are restricted to typed listing, exact-search, bounded-read, git-file, and explicit lane-complete actions rather than arbitrary shell steps."
  - "Benchmark suites stay JSON-backed and fixture-referenced so later timing and reporting plans reuse committed corpus definitions instead of inline test data."
  - "Treatment workflows are limited to shipped CLI commands and MCP tools, preventing benchmark-only shortcuts from contaminating later A/B claims."
patterns-established:
  - "Benchmark suites are validated in repository code and selected in app code through shared fixture-reference enforcement."
  - "Each suite defines paired baseline and OptimusCtx arms on the same fixture, task prompt, and lane stop markers."
requirements-completed: [BNCH-01, BNCH-03]
duration: 1m
completed: 2026-03-16
---

# Phase 11 Plan 01: Benchmark Methodology and Workflow Timing Summary

**Frozen benchmark suite contracts and fixture-backed paired baseline-versus-OptimusCtx corpus for discovery, context assembly, refresh, and task-completion lanes**

## Performance

- **Duration:** 1 min
- **Started:** 2026-03-16T12:02:30Z
- **Completed:** 2026-03-16T12:03:38Z
- **Tasks:** 3
- **Files modified:** 14

## Accomplishments
- Added one canonical benchmark schema covering suite identity, paired arms, lane contracts, stop markers, metrics, baseline actions, and shipped treatment surfaces.
- Added an app-layer benchmark runner that loads suites by path or ID and enforces fixture-backed selection rules before timing work begins.
- Committed a frozen benchmark corpus under `testdata/eval/benchmarks` plus a new benchmark-oriented fixture repository for non-trivial discovery and context assembly tasks.

## Task Commits

Each task was committed atomically:

1. **Task 1: Define canonical benchmark suite contracts and baseline action vocabulary** - `ff940ea` (feat)
2. **Task 2: Enforce suite selection and baseline-vs-treatment rules in the benchmark runner** - `36609c9` (feat)
3. **Task 3: Commit the benchmark corpus and suite definitions** - `0446170` (feat)

## Files Created/Modified
- `internal/repository/benchmark.go` - Canonical benchmark suite, lane, arm, and action contracts with validation.
- `internal/repository/benchmark_test.go` - Contract tests for benchmark suites and baseline action enforcement.
- `internal/app/benchmark_runner.go` - Fixture-backed suite loading and benchmark selection validation.
- `internal/app/benchmark_runner_test.go` - App-layer validation for benchmark suite loading and committed corpus checks.
- `testdata/eval/benchmarks/go-benchmark-discovery-v1.json` - Frozen paired benchmark definition for discovery and context assembly.
- `testdata/eval/benchmarks/go-benchmark-refresh-v1.json` - Frozen paired benchmark definition for refresh-after-change and task completion.
- `testdata/eval/fixtures/go-benchmark/v1/repository` - Non-trivial benchmark fixture repository used by the committed corpus.

## Decisions Made

- Baseline workflows are intentionally narrow and transport-neutral so Phase 11 timings compare a replayable non-OptimusCtx exploration path rather than ad hoc operator behavior.
- Treatment workflows point only at shipped CLI or MCP product surfaces, preserving the product boundary for later benchmark evidence.
- Benchmark suites live alongside eval fixtures as committed JSON so repeated runs can reuse the exact same fixture, task, and lane stop conditions.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Initial `go test` invocation failed under the sandboxed default Go environment because the build cache path was blocked and the default binary did not resolve the stdlib correctly. Verification was rerun successfully with `/usr/local/go/bin/go` plus `/tmp` cache roots, matching prior phase verification practice.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Phase 11 now has a fixed A/B benchmark substrate. The next plans can add lane timing capture, repeated comparison, and artifact persistence without redefining what the benchmark suites, baseline actions, or frozen corpus are allowed to do.

## Self-Check: PASSED

- Verified key files exist on disk: `internal/repository/benchmark.go`, `internal/app/benchmark_runner.go`, `testdata/eval/benchmarks/go-benchmark-discovery-v1.json`, `.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-01-SUMMARY.md`
- Verified task commits exist in Git history: `ff940ea`, `36609c9`, `0446170`
