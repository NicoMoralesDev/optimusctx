---
phase: 14-benchmark-boundary-redefinition-and-agent-input-validation
plan: "04"
subsystem: benchmarking
tags: [benchmarking, evaluation, cli, mcp, evidence, reproducibility]
requires:
  - phase: 14-03
    provides: committed v2 benchmark suites, refreshed fairness evidence, and passing export/verify/report flows on the repaired boundary
provides:
  - reproducibility verification that reports counted-boundary and final-artifact drift as benchmark failures instead of export crashes
  - documentation and fairness reporting that explicitly supersede the pre-Phase-14 attribution-first benchmark answer
  - full-suite verification and planning closeout for the repaired benchmark milestone
affects: [phase-14-closeout, benchmark-reporting, milestone-verification, planning-state]
tech-stack:
  added: []
  patterns:
    - reproducibility exports stay buildable even when reruns drift so verification can report the failure cleanly
    - benchmark integration tests assert active schema and label presence instead of brittle pre-repair ordering assumptions
key-files:
  created:
    - .planning/phases/14-benchmark-boundary-redefinition-and-agent-input-validation/14-04-SUMMARY.md
  modified:
    - internal/app/benchmark_service.go
    - internal/app/benchmark_runner.go
    - internal/cli/eval_integration_test.go
    - internal/mcp/integration_test.go
    - internal/store/migrations/runner_test.go
    - README.md
    - .planning/benchmark-fairness-report.md
    - .planning/STATE.md
    - .planning/ROADMAP.md
key-decisions:
  - "Repeated-run benchmark exports must preserve invalid attempts inside verification metadata instead of aborting on missing or drifted final-artifact records."
  - "The active benchmark narrative must state that pre-Phase-14 attribution-first evidence is superseded by the repaired v2 counted-input methodology."
  - "Full-suite verification should lock onto the active v2 evidence schema and required attribution labels without depending on stale ordering assumptions."
patterns-established:
  - "Benchmark reproducibility checks should treat contract drift as evidence, not as a serialization error path."
  - "Benchmark-facing docs should separate counted-input wins from raw provenance size so milestone claims stay narrow and honest."
requirements-completed: [BNCH-01, BNCH-02, BNCH-04]
duration: 13min
completed: 2026-03-16
---

# Phase 14 Plan 04 Summary

**Phase 14 now closes on reproducible v2 benchmark evidence, explicit supersession of the old attribution-first reading, and a fully green repository verification run.**

## Performance

- **Duration:** 13 min
- **Started:** 2026-03-16T20:10:48Z
- **Completed:** 2026-03-16T20:23:00Z
- **Tasks:** 3
- **Files modified:** 10

## Accomplishments

- Finished the v2 reproducibility path so persisted rerun evidence, methodology drift, and final-artifact drift are reported as benchmark verification failures rather than export-time errors.
- Tightened README and fairness-report wording so the active milestone story consistently describes declared agent inputs, provenance-only system work, comparable final-artifact gating, and the fact that provider billing is still out of scope.
- Revalidated the full repository test suite and aligned stale benchmark and migration assertions with the active v2 evidence schema before closing the planning state.

## Task Commits

Each task was committed atomically:

1. **Task 1: Finalize reproducibility and persistence verification for the v2 benchmark boundary** - `c9c3dc6` (`fix`)
2. **Task 2: Lock wording and human-summary honesty around the repaired methodology** - `01e8976` (`docs`)
3. **Task 3: Reclose Phase 14 in planning state after full benchmark verification** - `82c579e` (`test`)

## Files Created/Modified

- `internal/app/benchmark_service.go` - preserves drifted reruns as verification evidence and records explicit missing or mismatched final-artifact reasons
- `internal/app/benchmark_runner.go` - exposes a loaded-suite execution helper so single-run persistence reuses one resolved suite contract
- `README.md` - clarifies the active Phase 14 benchmark claim and removes stale benchmark-proof wording
- `.planning/benchmark-fairness-report.md` - explicitly supersedes the old attribution-first story and states the repaired v2 outcome narrowly
- `internal/cli/eval_integration_test.go` - updates benchmark export expectations to the active v2 evidence schema
- `internal/mcp/integration_test.go` - verifies required discovery attribution labels by presence rather than brittle ordering
- `internal/store/migrations/runner_test.go` - advances migration expectations to include the current benchmark migration set
- `.planning/STATE.md` - records Phase 14 completion and the repaired benchmark milestone state
- `.planning/ROADMAP.md` - marks the reopened Phase 14 milestone work complete
- `.planning/phases/14-benchmark-boundary-redefinition-and-agent-input-validation/14-04-SUMMARY.md` - documents the closeout evidence and decisions

## Decisions Made

- Routed missing final-artifact verification and contract mismatch cases into benchmark verification output because those states are exactly what reproducibility checks need to report.
- Kept the documentation explicit that counted-input improvements do not imply small raw provenance payloads or provider-billed savings.
- Updated the full-suite blockers as contract-drift fixes, not product-surface changes, so the closeout remained focused on benchmark truthfulness and milestone state.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Repeated-run export failed before verification could report final-artifact drift**
- **Found during:** Task 1 (Finalize reproducibility and persistence verification for the v2 benchmark boundary)
- **Issue:** Persisted rerun bundles aborted on missing or drifted final-artifact records, so `verify` surfaced an error instead of a reproducibility failure.
- **Fix:** Moved those cases into summary invalid-attempt reporting and kept the export path serializable so verification can compare persisted and regenerated evidence.
- **Files modified:** `internal/app/benchmark_service.go`, `internal/app/benchmark_runner.go`
- **Verification:** `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off go test ./internal/app ./internal/cli ./internal/store/sqlite -run 'TestBenchmark(Reproducibility|Persistence|Verification)'`
- **Committed in:** `c9c3dc6`

**2. [Rule 3 - Blocking] Full-suite assertions still targeted pre-closeout benchmark contracts**
- **Found during:** Task 3 (Reclose Phase 14 in planning state after full benchmark verification)
- **Issue:** The full repository suite still expected the old benchmark evidence schema version, brittle MCP attribution ordering, and a pre-benchmark migration count.
- **Fix:** Updated the CLI, MCP, and migration tests to assert the active v2 schema, required labels, and current migration set.
- **Files modified:** `internal/cli/eval_integration_test.go`, `internal/mcp/integration_test.go`, `internal/store/migrations/runner_test.go`
- **Verification:** `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off go test ./...`
- **Committed in:** `82c579e`

---

**Total deviations:** 2 auto-fixed (1 bug fix, 1 blocking test-alignment fix)
**Impact on plan:** Both deviations were required to finish Phase 14 honestly: the first kept reproducibility drift observable, and the second brought the repository verification surface up to the active benchmark contract without widening product scope.

## Issues Encountered

- Go sandbox writes to the default build cache were denied; rerouting `GOCACHE` and `GOMODCACHE` into `/tmp` kept all verification inside the permitted workspace.
- The full-suite run surfaced stale benchmark and migration assertions that had not been advanced alongside the repaired benchmark methodology.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 14 is ready to close with verified export, report, verify, and full-suite coverage on the repaired counted-input methodology.
- The fairness report, README, roadmap, and planning state can now treat the v2 rerun evidence as the active milestone source of truth.
- Any future work on shrinking `repository_map`, `health`, or other raw provenance payloads stays as backlog/product optimization, not hidden inside benchmark-closeout claims.

## Self-Check

PASSED

- Summary file created at `.planning/phases/14-benchmark-boundary-redefinition-and-agent-input-validation/14-04-SUMMARY.md`
- Task commits present: `c9c3dc6`, `01e8976`, `82c579e`
- Planning state reflects Phase 14 completion and roadmap progress is `4/4 plans complete`

---
*Phase: 14-benchmark-boundary-redefinition-and-agent-input-validation*
*Completed: 2026-03-16*
