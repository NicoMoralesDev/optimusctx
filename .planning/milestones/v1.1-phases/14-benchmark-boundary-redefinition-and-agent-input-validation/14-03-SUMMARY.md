---
phase: 14-benchmark-boundary-redefinition-and-agent-input-validation
plan: "03"
subsystem: benchmarking
tags: [benchmarking, evaluation, cli, mcp, evidence, reproducibility]
requires:
  - phase: 14-02
    provides: runner-side counted-input boundaries, final-artifact validation, and report semantics for benchmark v2 suites
provides:
  - committed discovery and refresh benchmark suites migrated onto the v2 counted-input boundary
  - passing export, verify, and report reruns for the active benchmark corpus
  - refreshed fairness diagnosis grounded in Phase 14 counted-input evidence
affects: [phase-14-closeout, benchmark-reporting, milestone-verification]
tech-stack:
  added: []
  patterns:
    - concatenated lane-text final artifacts for multi-step baseline context assembly
    - readiness-summary rendering that tolerates real MCP field casing while keeping normalized output stable
    - reproducibility fingerprints that ignore path-sensitive system-provenance token magnitudes
key-files:
  created:
    - .planning/phases/14-benchmark-boundary-redefinition-and-agent-input-validation/14-03-SUMMARY.md
  modified:
    - testdata/eval/benchmarks/go-benchmark-discovery-v1.json
    - testdata/eval/benchmarks/go-benchmark-refresh-v1.json
    - internal/app/benchmark_runner.go
    - internal/app/benchmark_runner_test.go
    - internal/app/benchmark_service.go
    - internal/app/benchmark_service_test.go
    - internal/cli/eval_integration_test.go
    - README.md
    - .planning/benchmark-fairness-report.md
key-decisions:
  - "The frozen benchmark selectors stay on go-benchmark-*-v1 ids while schemaVersion moves to optimusctx/benchmark-suite@v2."
  - "Counted benchmark totals now come only from declared agent-input projections; raw CLI and MCP provenance stays exported but does not drive counted deltas."
  - "Refresh readiness final artifacts normalize the shared targetReady signal, while treatment-only freshness and generation remain counted operational projections."
  - "Repeated-run fingerprints ignore system-provenance token magnitudes because temp-workspace roots make those raw payload bytes path-sensitive."
patterns-established:
  - "Benchmark suite migrations should lock committed corpus ids in tests before evidence refresh begins."
  - "Lane-level text final artifacts should combine all observed bounded reads in sorted step order."
  - "Benchmark reproducibility checks should fingerprint methodology-significant records, not temp-path-sensitive provenance payload sizes."
requirements-completed: [BNCH-01, BNCH-02, BNCH-04]
duration: 25min
completed: 2026-03-16
---

# Phase 14 Plan 03 Summary

**Committed discovery and refresh suites now run as v2 counted-input benchmarks with passing export/verify/report evidence and a refreshed fairness diagnosis that separates counted cost from raw provenance.**

## Performance

- **Duration:** 25 min
- **Started:** 2026-03-16T19:44:00Z
- **Completed:** 2026-03-16T20:09:03Z
- **Tasks:** 3
- **Files modified:** 9

## Accomplishments

- Migrated `go-benchmark-discovery-v1` and `go-benchmark-refresh-v1` onto the Phase 14 v2 boundary while preserving their committed suite ids.
- Aligned service, CLI, and README guidance with the active committed corpus and the counted-input plus final-artifact semantics.
- Re-ran both suites through the shipped `export -> verify -> report` path and refreshed the fairness report from the passing evidence.

## Task Commits

Each task was committed atomically:

1. **Task 1: Migrate the frozen discovery and refresh suites to the v2 benchmark contract** - `48b838e` (`feat`)
2. **Task 2: Refresh export, verify, and report expectations around the migrated suites** - `fd9896d` (`feat`)
3. **Task 3: Rerun both suites and refresh the benchmark evidence narrative** - `4ca45fc` (`fix`)

## Files Created/Modified

- `testdata/eval/benchmarks/go-benchmark-discovery-v1.json` - migrated discovery suite with declared counted inputs and per-lane final-artifact contracts
- `testdata/eval/benchmarks/go-benchmark-refresh-v1.json` - migrated refresh suite with readiness and task-completion contracts that the shipped workflow can satisfy
- `internal/app/benchmark_runner.go` - combined multi-step lane text, normalized readiness summaries, and removed path-sensitive provenance sizes from reproducibility fingerprints
- `internal/app/benchmark_runner_test.go` - locked the frozen corpus to v2 and added regression coverage for combined lane text and capitalized readiness fields
- `internal/app/benchmark_service.go` - clarified report wording around final-artifact validation and tuned reproducibility comparison semantics
- `internal/app/benchmark_service_test.go` - aligned service-level benchmark fixtures with the active committed refresh corpus and reproducibility expectations
- `internal/cli/eval_integration_test.go` - added CLI coverage for the committed corpus export and verify path
- `README.md` - documented the active v2 benchmark corpus, counted-input semantics, and the current rerun outcome at a high level
- `.planning/benchmark-fairness-report.md` - replaced the stale negative diagnosis with the corrected counted-input rerun story and separate product backlog notes

## Decisions Made

- Kept the committed suite ids stable as `go-benchmark-*-v1` so the user-facing selection contract did not change while the schema semantics did.
- Treated counted-input projections as the benchmark answer and raw provenance as audit-only evidence, matching the Phase 14 requirement intent.
- Reduced the refresh readiness final-artifact contract to the shared `targetReady` signal because both arms can prove readiness, while only treatment exposes operational freshness details.
- Treated raw provenance token magnitudes as non-fingerprintable because temp-workspace roots make those bytes nondeterministic across otherwise equivalent reruns.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Multi-step lane artifacts only kept the first bounded read**
- **Found during:** Task 3 (Rerun both suites and refresh the benchmark evidence narrative)
- **Issue:** The migrated discovery suite used two baseline context reads, but lane final-artifact rendering only emitted the first observed text block.
- **Fix:** Changed lane-text artifact rendering to concatenate all observed bounded reads in sorted step order and added a regression test.
- **Files modified:** `internal/app/benchmark_runner.go`, `internal/app/benchmark_runner_test.go`
- **Verification:** `go test ./internal/app -run 'TestBenchmark(TextFromLaneStepsCombinesObservations|LaneCompletionRequiresFinalArtifact|FinalArtifactValidation)'`
- **Committed in:** `4ca45fc`

**2. [Rule 1 - Bug] Readiness summary rendering assumed lower-case health fields**
- **Found during:** Task 3 (Rerun both suites and refresh the benchmark evidence narrative)
- **Issue:** The real MCP `health` payload exposed capitalized field names, so the migrated refresh suite could not materialize a readiness summary consistently.
- **Fix:** Added capitalized and nested field fallbacks for readiness-summary rendering and covered the real payload shape in tests.
- **Files modified:** `internal/app/benchmark_runner.go`, `internal/app/benchmark_runner_test.go`
- **Verification:** `go test ./internal/app -run 'TestBenchmark(ReadinessSummarySupportsCapitalizedHealthFields|AttributionBoundary)'`
- **Committed in:** `4ca45fc`

**3. [Rule 1 - Bug] Reproducibility fingerprint drifted on temp-path-sensitive provenance sizes**
- **Found during:** Task 3 (Rerun both suites and refresh the benchmark evidence narrative)
- **Issue:** Discovery `verify` failed even when the suite contract held because raw provenance payload sizes varied with temp workspace roots across attempts.
- **Fix:** Excluded system-provenance token magnitudes from the attempt fingerprint while still fingerprinting counted inputs, final-artifact verification, and workflow structure.
- **Files modified:** `internal/app/benchmark_service.go`, `internal/app/benchmark_service_test.go`
- **Verification:** `go test ./internal/app -run 'TestBenchmark(MethodologyFingerprint|AttemptFingerprintIgnoresSystemProvenanceTokenDrift)'` and `go run ./cmd/optimusctx eval benchmark verify --suite go-benchmark-discovery-v1 --attempts 2`
- **Committed in:** `4ca45fc`

**4. [Rule 2 - Missing Critical] Refresh readiness artifact required treatment-only detail**
- **Found during:** Task 3 (Rerun both suites and refresh the benchmark evidence narrative)
- **Issue:** The migrated refresh suite initially required a `freshness` field in the readiness final artifact, but the baseline arm cannot produce that operational detail.
- **Fix:** Narrowed the readiness final-artifact normalization to the comparable `targetReady` field while keeping treatment freshness and generation as counted operational projections.
- **Files modified:** `testdata/eval/benchmarks/go-benchmark-refresh-v1.json`, `.planning/benchmark-fairness-report.md`
- **Verification:** `go run ./cmd/optimusctx eval benchmark verify --suite go-benchmark-refresh-v1 --attempts 2`
- **Committed in:** `4ca45fc`

---

**Total deviations:** 4 auto-fixed (3 bug fixes, 1 missing-critical contract correction)
**Impact on plan:** All deviations were required to make the migrated committed suites executable and reproducible on the shipped CLI path. No benchmark-product scope creep was introduced.

## Issues Encountered

- Go build cache writes initially failed under the sandbox because the default cache path pointed at `/home/nico/.cache`; rerouting `GOCACHE` and `GOMODCACHE` into `/tmp` resolved the verification path cleanly.
- The first discovery rerun exposed the lane-artifact concatenation bug immediately, which was faster to fix in the runner than to weaken the suite contract around multi-step context assembly.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The committed corpus now passes `export`, `verify`, and `report` on both suites under the v2 counted-input contract.
- The fairness report now reflects the corrected counted-input outcome: discovery and refresh readiness favor OptimusCtx, while bounded task completion is a tie.
- Phase 14 closeout can focus on reproducibility sign-off and milestone re-close rather than unfinished suite migration work.

## Self-Check

PASSED

- Summary file created at `.planning/phases/14-benchmark-boundary-redefinition-and-agent-input-validation/14-03-SUMMARY.md`
- Task commits present: `48b838e`, `fd9896d`, `4ca45fc`

---
*Phase: 14-benchmark-boundary-redefinition-and-agent-input-validation*
*Completed: 2026-03-16*
