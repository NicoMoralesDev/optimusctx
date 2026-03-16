---
phase: 12-token-attribution-and-evidence-reporting
plan: "01"
subsystem: benchmarking
tags: [go, sqlite, benchmarking, token-attribution, evidence-reporting]
requires:
  - phase: 11-a-b-benchmark-methodology-and-workflow-timing
    provides: paired benchmark runs, frozen suite contracts, and persisted lane evidence
provides:
  - canonical benchmark token estimator contract and wording guardrails
  - step-scoped artifact attribution records for baseline and OptimusCtx lanes
  - persisted attribution-ready benchmark metadata across sqlite, CLI, and MCP runs
affects: [phase-12-02, phase-12-03, phase-12-04, bnch-02, bnch-04]
tech-stack:
  added: []
  patterns: [shared bytes_div_4_ceiling benchmark estimator, typed artifact attribution in lane metadata]
key-files:
  created: []
  modified: [internal/repository/benchmark.go, internal/repository/benchmark_test.go, internal/app/benchmark_runner.go, internal/app/benchmark_runner_test.go, internal/store/sqlite/benchmark.go, internal/store/sqlite/benchmark_test.go, internal/cli/eval_integration_test.go, internal/mcp/integration_test.go]
key-decisions:
  - "Phase 12 benchmark token math stays on the shared bytes_div_4_ceiling policy and is labeled as estimated workflow-consumed tokens, not provider billing."
  - "Treatment artifact attribution is step-scoped and typed in repository contracts so runner, store, exports, and summaries reuse one canonical vocabulary."
  - "Attribution evidence persists inside existing benchmark run and lane metadata JSON so reporting can read canonical inputs without recomputing from logs."
patterns-established:
  - "Benchmark attribution records carry step id, lane, source kind, report label, and estimated token inputs."
  - "CLI pack export attribution reuses manifest section estimates instead of re-estimating bytes from exported files."
requirements-completed: [BNCH-02, BNCH-04]
duration: 6m
completed: 2026-03-16
---

# Phase 12 Plan 01: Token Accounting Contract and Artifact Attribution Summary

**Canonical benchmark token attribution with one bytes_div_4_ceiling estimator, step-scoped artifact evidence, and persisted SQLite/CLI/MCP report inputs**

## Performance

- **Duration:** 6 min
- **Started:** 2026-03-16T13:44:11Z
- **Completed:** 2026-03-16T13:50:15Z
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments

- Added one repository-level benchmark attribution contract that defines estimator identity, truthful wording, artifact taxonomy, and BNCH-02 report-label mapping.
- Instrumented the benchmark runner to emit step-scoped attribution records for bounded baseline reads, MCP payloads, and pack-export section estimates.
- Persisted canonical attribution inputs through benchmark metadata and verified the real SQLite, CLI, and MCP execution paths round-trip the evidence.

## Task Commits

Each task was committed atomically:

1. **Task 1: Define the canonical token-attribution taxonomy and estimator contract** - `f327094` (feat)
2. **Task 2: Capture step-level attribution evidence in the benchmark runner** - `7b3f86c` (feat)
3. **Task 3: Persist attribution-ready benchmark evidence through the real execution boundary** - `99337a2` (feat)

## Files Created/Modified

- `internal/repository/benchmark.go` - Added the benchmark token estimate contract, artifact taxonomy, report-label mapping, and typed attribution record.
- `internal/repository/benchmark_test.go` - Added contract tests for estimator policy, truthful wording, and artifact-type/report-label mappings.
- `internal/app/benchmark_runner.go` - Emitted step-scoped attribution records for baseline bounded reads, MCP payloads, and pack-export section estimates.
- `internal/app/benchmark_runner_test.go` - Verified runner token estimation and step attribution behavior.
- `internal/store/sqlite/benchmark.go` - Persisted estimator contract and lane attribution into benchmark metadata JSON.
- `internal/store/sqlite/benchmark_test.go` - Verified attribution inputs survive SQLite persistence without lossy parsing.
- `internal/cli/eval_integration_test.go` - Verified the shipped CLI benchmark path persists L2-context attribution evidence.
- `internal/mcp/integration_test.go` - Verified the MCP benchmark path persists repository-map and exact-lookup attribution evidence.

## Decisions Made

- Kept benchmark token estimation on the shared `bytes_div_4_ceiling` policy so Phase 12 reporting does not introduce a second estimator.
- Modeled baseline token usage as bounded file-content estimates via `sourceKind=bounded_file_content` instead of inventing baseline artifact types.
- Persisted attribution inside existing benchmark metadata records to keep reports and exports downstream readers of canonical evidence rather than new sources of truth.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The default sandbox Go environment pointed at an unusable cache/stdlib setup, so verification used `/usr/local/go/bin/go` with `GOCACHE` and `GOMODCACHE` redirected into `/tmp`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 12 plan 02 can build machine-readable exports on persisted estimator and attribution inputs instead of reconstructing token math from benchmark logs.
- Phase 12 plan 03 can render human summaries from canonical report labels already stored with benchmark lane metadata.

## Self-Check: PASSED

- Verified `.planning/phases/12-token-attribution-and-evidence-reporting/12-01-SUMMARY.md` exists.
- Verified task commits `f327094`, `7b3f86c`, and `99337a2` exist in git history.
