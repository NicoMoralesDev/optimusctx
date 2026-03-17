---
phase: 11-a-b-benchmark-methodology-and-workflow-timing
plan: "04"
subsystem: benchmarking
tags: [benchmarking, sqlite, cli, mcp, verification]
requires:
  - phase: 11-03
    provides: refresh-after-change and task-completion lane execution with persisted lane evidence
provides:
  - repeated paired benchmark attempt orchestration
  - persisted benchmark attempt ordering and comparison reads
  - end-to-end methodology verification for stable reruns and drift rejection
  - truthful phase 11 rerun guidance for benchmark evidence
affects: [12-token-attribution-and-evidence-reporting, benchmarking, verification]
tech-stack:
  added: []
  patterns: [service-level repeated benchmark orchestration, sqlite-backed benchmark attempt summaries, verification-through-shipped-cli-and-mcp-surfaces]
key-files:
  created: [internal/app/benchmark_service.go]
  modified: [internal/app/benchmark_runner_test.go, internal/store/sqlite/benchmark.go, internal/store/sqlite/benchmark_test.go, internal/cli/eval_integration_test.go, internal/mcp/integration_test.go, README.md]
key-decisions:
  - "Repeated benchmark verification runs through a dedicated app-layer BenchmarkService instead of adding a new user-facing benchmark CLI command in Phase 11."
  - "Methodology drift is rejected by comparing repeated attempts against the frozen suite and lane contract, while SQLite keeps attempt ordering stable for later Phase 12 reporting."
patterns-established:
  - "Benchmark verification uses repeated paired attempts persisted per arm and attempt number before any higher-level reporting."
  - "Operator rerun guidance stays on committed tests and persisted SQLite evidence until Phase 12 adds richer report surfaces."
requirements-completed: [BNCH-01, BNCH-03]
duration: 5min
completed: 2026-03-16
---

# Phase 11 Plan 04: Repeated-run comparison and benchmark verification Summary

**Repeated paired benchmark attempts with SQLite-backed summaries, end-to-end rerun verification, and explicit methodology-drift rejection**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-16T12:50:24Z
- **Completed:** 2026-03-16T12:55:42Z
- **Tasks:** 3
- **Files modified:** 7

## Accomplishments

- Added `internal/app.BenchmarkService` to run and persist repeated paired benchmark attempts with stable suite and attempt identity.
- Added SQLite helpers for benchmark attempt numbering and ordered benchmark-run reads so comparison summaries can be built deterministically.
- Added CLI and MCP verification coverage proving stable reruns pass and methodology drift fails loudly.
- Documented the real Phase 11 rerun path without overclaiming Phase 12 token attribution or reporting.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add repeated-run orchestration and paired comparison summaries** - `cb35b90` (feat)
2. **Task 2: Verify rerun determinism and the benchmark workflow contract end to end** - `46d85fd` (test)
3. **Task 3: Document the rerun and verification path truthfully** - `0b638e1` (docs)

## Files Created/Modified

- `internal/app/benchmark_service.go` - app-layer repeated-run orchestration, verification, and comparison summary assembly
- `internal/app/benchmark_runner_test.go` - repeated-run service coverage on the benchmark runner seam
- `internal/store/sqlite/benchmark.go` - benchmark attempt numbering and ordered persisted benchmark-run reads
- `internal/store/sqlite/benchmark_test.go` - persistence coverage for repeated attempt ordering and comparison inputs
- `internal/cli/eval_integration_test.go` - end-to-end CLI verification coverage for stable reruns and drift rejection
- `internal/mcp/integration_test.go` - repeated-run MCP determinism coverage through shipped tool calls
- `README.md` - truthful operator guidance for rerunning Phase 11 benchmark verification

## Decisions Made

- Reused the existing evaluation substrate and store instead of introducing a Phase 11-only benchmark command surface.
- Kept Phase 11 summaries focused on paired attempts, lane timing aggregates, and verification failure reasons so Phase 12 can own richer reporting.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Go test initially failed in the sandbox because the default Go build cache lived outside writable roots. Verification was rerun with `GOCACHE` and `GOMODCACHE` redirected into `/tmp`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 12 can build token attribution and human-readable reporting on top of persisted repeated-run evidence instead of revisiting methodology.
- Benchmark reruns now preserve suite, arm, lane, and attempt identity strongly enough for later comparison and export work.

## Self-Check

PASSED

- Verified `.planning/phases/11-a-b-benchmark-methodology-and-workflow-timing/11-04-SUMMARY.md` exists
- Verified `internal/app/benchmark_service.go` exists
- Verified task commits `cb35b90`, `46d85fd`, and `0b638e1` exist in git history

---
*Phase: 11-a-b-benchmark-methodology-and-workflow-timing*
*Completed: 2026-03-16*
