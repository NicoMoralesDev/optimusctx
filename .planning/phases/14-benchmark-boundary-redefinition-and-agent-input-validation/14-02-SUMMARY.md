---
phase: 14-benchmark-boundary-redefinition-and-agent-input-validation
plan: "02"
subsystem: benchmarking
tags: [benchmarking, eval, cli, mcp, attribution, verification]
requires:
  - phase: 14-01
    provides: v2 benchmark suite boundary contract, counted-input declarations, and final-artifact schema
provides:
  - runner-side counted-input projection with raw provenance preserved separately
  - lane success gated by runtime final-artifact materialization and verification
  - report and verify wording that distinguishes counted agent inputs from provenance
affects: [14-03, 14-04, benchmark evidence, eval benchmark report]
tech-stack:
  added: []
  patterns: [boundary-tagged attribution records, lane-level normalized final-artifact materialization]
key-files:
  created: [.planning/phases/14-benchmark-boundary-redefinition-and-agent-input-validation/14-02-SUMMARY.md]
  modified: [internal/app/benchmark_runner.go, internal/app/benchmark_runner_test.go, internal/app/benchmark_service.go, internal/app/benchmark_service_test.go, internal/cli/eval.go, internal/cli/eval_integration_test.go]
key-decisions:
  - "The runner records raw CLI and MCP outputs as `system_provenance`, then projects only declared `countedInputs` into `agent_input` totals."
  - "Lane completion now requires both stop-condition progress and a materialized final artifact that satisfies the lane contract."
  - "Human-readable reports label attribution rows as counted agent inputs and leave provenance in exported evidence for auditability instead of folding it into totals."
patterns-established:
  - "Benchmark attribution uses three explicit boundaries: `agent_input`, `system_provenance`, and `final_artifact_verification`."
  - "Synthetic v2 benchmark suites in tests should model counted inputs and lane final artifacts directly instead of relying on legacy v1 fixtures."
requirements-completed: [BNCH-01, BNCH-02, BNCH-04]
duration: 22min
completed: 2026-03-16
---

# Phase 14 Plan 02: Runner Boundary Enforcement Summary

**Benchmark runtime enforcement now projects counted agent inputs from declared v2 suite contracts, preserves raw workflow payloads as provenance, and refuses lane success without comparable normalized final artifacts.**

## Performance

- **Duration:** 22 min
- **Started:** 2026-03-16T19:19:06Z
- **Completed:** 2026-03-16T19:40:48Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments
- The benchmark runner now tags raw CLI stdout/stderr, MCP payloads, and command markers as `system_provenance` and only adds estimated-token totals for suite-declared counted-input projections.
- Lane and task success now require runtime final-artifact materialization plus contract assertion checks, with explicit `final_artifact_verification` evidence persisted alongside attempts.
- Report and verify surfaces now describe counted agent-input attribution honestly and surface final-artifact failures with explicit reasons instead of generic stop-condition wording.

## Task Commits

Each task was committed atomically:

1. **Task 1: Project counted agent inputs while preserving raw provenance** - `4d3d2a8` (feat)
2. **Task 2: Gate lane completion on comparable normalized final artifacts** - `14fef8d` (feat)
3. **Task 3: Update export, report, and verify surfaces to reflect counted-boundary enforcement** - `5cf605b` (feat)

**Plan metadata:** pending docs commit

## Files Created/Modified
- `internal/app/benchmark_runner.go` - Projects v2 counted inputs, keeps provenance separate, materializes final artifacts, and persists verification records in lane results.
- `internal/app/benchmark_runner_test.go` - Reworks runner coverage around synthetic v2 suites for projection, boundary separation, final-artifact gating, and repeated runs.
- `internal/app/benchmark_service.go` - Tightens invalid-run wording and report rendering so counted agent-input attribution stays distinct from provenance.
- `internal/app/benchmark_service_test.go` - Adds service-level coverage for counted-input report wording and explicit final-artifact failure reasons.
- `internal/cli/eval.go` - Updates benchmark help text to describe counted-input evidence truthfully.
- `internal/cli/eval_integration_test.go` - Verifies CLI reports use the counted agent-input wording.

## Decisions Made
- Used the suite’s declared `countedInputs` as the only source for token totals instead of inferring billable inputs from raw transport payload sizes.
- Materialized lane final artifacts into deterministic workspace paths so report, persistence, and verification surfaces all observe the same runtime contract.
- Kept provenance visible in exported evidence rather than adding a second human summary table, which preserves auditability without corrupting counted totals.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Upgraded synthetic benchmark runner suites to the v2 contract**
- **Found during:** Task 1 (Project counted agent inputs while preserving raw provenance)
- **Issue:** Existing synthetic runner tests still generated `optimusctx/benchmark-suite@v1` fixtures, which the Phase 14 schema changes reject.
- **Fix:** Rebuilt the runner test helpers around v2 `boundary`, `countedInputs`, and `finalArtifact` declarations so projection tests exercise the real contract.
- **Files modified:** `internal/app/benchmark_runner_test.go`
- **Verification:** `go test ./internal/app -run 'TestBenchmark(AgentInput|Attribution|Boundary|Projection)'`
- **Committed in:** `4d3d2a8`

**2. [Rule 1 - Bug] Allowed zero-assertion lanes to complete through final-artifact verification**
- **Found during:** Task 3 (Update export, report, and verify surfaces to reflect counted-boundary enforcement)
- **Issue:** Discovery and context lanes with no extra assertions were skipping final-artifact verification because completion logic returned early when `len(assertions) == 0`.
- **Fix:** Changed lane completion to require stop-condition progress first, then run final-artifact verification whether or not assertion clauses exist.
- **Files modified:** `internal/app/benchmark_runner.go`, `internal/app/benchmark_runner_test.go`
- **Verification:** all three plan verification commands passed after the change
- **Committed in:** `5cf605b`

---

**Total deviations:** 2 auto-fixed (1 blocking, 1 bug)
**Impact on plan:** Both fixes were directly required to make the v2 runtime boundary executable and testable. No scope creep beyond the benchmark enforcement contract.

## Issues Encountered

- The checked-in benchmark corpus under `testdata/eval/benchmarks/` remains on the legacy v1 schema. Task 14-02 therefore used synthetic v2 suites for runner/runtime coverage while preserving the in-progress fixture edits already present in the worktree for later plan 14-03 migration.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The runtime now enforces the counted boundary and comparable final-artifact contract expected by the v2 schema.
- Phase 14-03 can migrate the frozen benchmark JSON corpus and regenerated evidence onto this runtime without needing another reporting or runner semantics change.

## Self-Check: PASSED

- Summary file exists at `.planning/phases/14-benchmark-boundary-redefinition-and-agent-input-validation/14-02-SUMMARY.md`
- Task commits `4d3d2a8`, `14fef8d`, and `5cf605b` are present in git history
