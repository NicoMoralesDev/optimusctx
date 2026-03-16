---
phase: 12-token-attribution-and-evidence-reporting
plan: "04"
subsystem: testing
tags: [benchmarking, reproducibility, cli, sqlite, evidence]
requires:
  - phase: 12-token-attribution-and-evidence-reporting
    provides: persisted benchmark evidence bundles, human-readable benchmark reports, and rerunnable export/report command surfaces
provides:
  - milestone benchmark verification through `optimusctx eval benchmark verify`
  - reproducibility checks that compare regenerated evidence to persisted bundles and latest persisted rerun records
  - reviewer guidance for rerun, report, and verification workflow
affects: [phase-13-distribution-pipeline-and-adoption-plan, benchmark-verification, milestone-proof]
tech-stack:
  added: []
  patterns: [persisted rerun windowing, methodology-and-wording milestone gates, tolerant path-variance attribution comparison]
key-files:
  created: [.planning/phases/12-token-attribution-and-evidence-reporting/12-04-SUMMARY.md]
  modified:
    - internal/app/benchmark_service.go
    - internal/app/benchmark_service_test.go
    - internal/cli/eval.go
    - internal/cli/eval_integration_test.go
    - internal/store/sqlite/benchmark.go
    - internal/store/sqlite/benchmark_test.go
    - README.md
    - internal/app/benchmark_runner_test.go
    - internal/store/migrations/runner_test.go
key-decisions:
  - "Repeated benchmark reruns now persist from the next available attempt number, while milestone reproducibility compares normalized evidence instead of raw historical attempt IDs."
  - "Milestone verification checks methodology fingerprint, estimator contract, deterministic lane summaries, and guarded report wording, while tolerating path-sensitive token variance where prior benchmark contracts already allowed it."
patterns-established:
  - "Benchmark reproducibility reads the latest persisted rerun window instead of all historical attempts."
  - "CLI milestone verification reuses the same service/export/report pipeline and prints one operator-facing pass/fail summary."
requirements-completed: [BNCH-02, BNCH-04]
duration: 26min
completed: 2026-03-16
---

# Phase 12 Plan 04: Reproducibility Checks and Milestone Verification Summary

**Milestone-grade benchmark verification now reruns frozen suites, compares regenerated evidence against persisted benchmark artifacts, and exposes pass/fail review output through `eval benchmark verify`**

## Performance

- **Duration:** 26 min
- **Started:** 2026-03-16T14:18:08Z
- **Completed:** 2026-03-16T14:44:04Z
- **Tasks:** 3
- **Files modified:** 9

## Accomplishments
- Added reproducibility verification in the benchmark service, including persisted rerun-window queries and evidence comparison against stored bundles.
- Exposed milestone verification through `optimusctx eval benchmark verify` with methodology fingerprint, rerun command, wording guards, and drift reporting.
- Updated benchmark guidance and aligned verification tests so the documented rerun/export/report flow matches the shipped command surface.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add reproducibility checks for exported attribution and report evidence** - `ce1ecd1` (feat)
2. **Task 2: Expose milestone verification and drift failures through the benchmark command path** - `9818cb1` (feat)
3. **Task 3: Finalize rerun and milestone-verification guidance** - `d7b754a` (chore)

## Files Created/Modified
- `internal/app/benchmark_service.go` - milestone verification flow, persisted rerun windowing, and normalized reproducibility comparison helpers
- `internal/app/benchmark_service_test.go` - reproducibility and methodology fingerprint coverage for regenerated evidence
- `internal/store/sqlite/benchmark.go` - latest persisted rerun window lookup plus evidence-bundle upsert timestamp fix
- `internal/store/sqlite/benchmark_test.go` - persisted rerun-window coverage for attribution recomputation inputs
- `internal/cli/eval.go` - `eval benchmark verify` command surface and operator summary rendering
- `internal/cli/eval_integration_test.go` - milestone verification and wording-guard coverage through the real CLI path
- `README.md` - rerun/export/report/verify guidance for milestone reviewers
- `internal/app/benchmark_runner_test.go` - rerun-command expectation aligned with shipped benchmark export surface
- `internal/store/migrations/runner_test.go` - migration runner expectations aligned with the existing benchmark evidence migration set

## Decisions Made

- Reproducibility compares the latest persisted rerun slice for the requested attempt window instead of rebuilding from every historical benchmark run.
- Milestone verification treats exact path-sensitive token bytes as tolerable variance while still failing on methodology, label, estimator, or wording drift.
- The benchmark CLI keeps `export` as the machine-readable rerun contract and adds `verify` as the milestone-facing gate rather than inventing a separate verification pipeline.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Repeated benchmark reruns collided on persisted attempt uniqueness**
- **Found during:** Task 1
- **Issue:** `RunRepeated` always persisted attempts as `1..N`, which failed as soon as milestone verification reran a suite with prior benchmark evidence in SQLite.
- **Fix:** Repeated runs now start from `NextBenchmarkAttempt` and milestone bundle reconstruction reads the latest persisted rerun window instead of colliding with historical attempts.
- **Files modified:** `internal/app/benchmark_service.go`, `internal/store/sqlite/benchmark.go`
- **Verification:** `go test ./internal/app ./internal/store/sqlite -run 'TestBenchmarkRerunReproducibility|TestBenchmarkRecomputedAttributionMatchesExport'`
- **Committed in:** `ce1ecd1`

**2. [Rule 1 - Bug] Evidence-bundle upsert fallback decoded `created_at` incorrectly**
- **Found during:** Task 2
- **Issue:** Re-saving a bundle with the same methodology fingerprint hit the SQLite upsert fallback path and failed scanning `created_at` into `time.Time`.
- **Fix:** The store now scans the timestamp as text and parses RFC3339Nano before returning the existing bundle row ID.
- **Files modified:** `internal/store/sqlite/benchmark.go`
- **Verification:** `go test ./internal/app ./internal/cli -run 'TestBenchmarkMilestoneVerification|TestBenchmarkMethodologyFingerprint|TestBenchmarkVerificationWordingGuards'`
- **Committed in:** `ce1ecd1`

**3. [Rule 1 - Bug] Full-suite regressions surfaced stale benchmark and migration expectations**
- **Found during:** Task 3
- **Issue:** The benchmark runner test still expected an obsolete rerun-command string, and the migration runner test still assumed the schema stopped at migration `0005`.
- **Fix:** Updated those expectations to the shipped benchmark export contract and the existing `0006_benchmark_evidence.sql` migration set.
- **Files modified:** `internal/app/benchmark_runner_test.go`, `internal/store/migrations/runner_test.go`
- **Verification:** `go test ./...`
- **Committed in:** `d7b754a`

---

**Total deviations:** 3 auto-fixed (3 rule-1 bugs)
**Impact on plan:** All three fixes were required for milestone verification to work reliably against persisted evidence and for the full verification suite to pass.

## Issues Encountered

- Full-suite verification initially failed on an unrelated flaky watch recovery test once, but the targeted rerun passed and no code changes were needed there.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 12 now closes with a shipped benchmark milestone gate, persisted reproducibility checks, and reviewer-facing documentation.
- Phase 13 can treat benchmark evidence as finalized milestone proof and build distribution guidance on top of a stable verification story.

## Self-Check

PASSED
