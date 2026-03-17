---
phase: 12-token-attribution-and-evidence-reporting
plan: "03"
subsystem: reporting
tags: [go, benchmark, reporting, cli, sqlite]
requires:
  - phase: 12-02
    provides: persisted benchmark evidence bundles, attribution records, and export rerun metadata
provides:
  - shared human-summary inputs derived from persisted benchmark evidence
  - deterministic human-readable benchmark report rendering with lane and attribution sections
  - shipped CLI report surface for benchmark evidence review
affects: [phase-12-verification, benchmark-review, README]
tech-stack:
  added: []
  patterns: [persisted-evidence-first report rendering, phrase-guarded benchmark wording]
key-files:
  created: [.planning/phases/12-token-attribution-and-evidence-reporting/12-03-SUMMARY.md]
  modified: [internal/app/benchmark_service.go, internal/app/benchmark_service_test.go, internal/cli/eval.go, internal/cli/eval_integration_test.go, README.md]
key-decisions:
  - "Human-readable benchmark reports render from normalized persisted evidence bundles instead of bespoke CLI aggregation."
  - "Lane token comparisons use the explicit bytes_div_4_ceiling estimator for baseline bytes-read and treatment attribution totals."
  - "Operator-facing attribution labels stay on BNCH-02 terms such as Repository Map, Exact Lookup, L2 Context, and Pack Export."
patterns-established:
  - "Benchmark reporting reuses persisted/exported evidence as the single source of truth across JSON export and human-readable output."
  - "Report wording explicitly disambiguates estimated workflow-consumed tokens from provider billing claims."
requirements-completed: [BNCH-02, BNCH-04]
duration: 9min
completed: 2026-03-16
---

# Phase 12 Plan 03: Human-Readable Benchmark Summaries and Comparison Reports Summary

**Deterministic benchmark reports with lane-by-lane timing and token comparisons, BNCH-02-facing attribution labels, and truthful estimator caveats rendered from persisted evidence bundles**

## Performance

- **Duration:** 9 min
- **Started:** 2026-03-16T14:08:30Z
- **Completed:** 2026-03-16T14:17:18Z
- **Tasks:** 3
- **Files modified:** 5

## Accomplishments
- Added a shared benchmark human-summary model that derives lane timing, estimated-token comparisons, attribution rows, caveats, and rerun guidance from persisted evidence bundles.
- Shipped `optimusctx eval benchmark report` so operators can render truthful human-readable benchmark output through the real CLI boundary.
- Updated benchmark documentation and regression coverage so report wording, attribution labels, and persisted-evidence reuse stay aligned.

## Task Commits

1. **Task 1: Build human-summary inputs from persisted benchmark evidence** - `f9eab78` (feat)
2. **Task 2: Expose human-readable comparison reports through the shipped CLI path** - `9ef2efe` (feat)
3. **Task 3: Document how to read the benchmark summaries truthfully** - `26f78e9` (docs)

## Files Created/Modified
- `internal/app/benchmark_service.go` - builds report-oriented human summary inputs and renders deterministic benchmark comparison reports from persisted evidence bundles
- `internal/app/benchmark_service_test.go` - verifies summary inputs, rendered wording, and persisted-evidence reuse
- `internal/cli/eval.go` - adds the `eval benchmark report` command path and shared benchmark report flag parsing
- `internal/cli/eval_integration_test.go` - exercises real CLI benchmark reporting, attribution labels, and wording guards across frozen suites
- `README.md` - explains export versus report surfaces, estimator caveats, attribution scope, and rerun guidance

## Decisions Made
- Human-readable reporting stays in `internal/app` so the CLI consumes the same persisted-evidence summary model as other report surfaces.
- Treatment attribution rows aggregate by lane plus BNCH-02-facing report label, while baseline token estimates are derived from recorded bytes-read with the same explicit estimator.
- The report deliberately states estimated workflow-consumed tokens and avoids claims about provider billing, universal savings, or statistical significance.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- `go test` could not write to the default Go build cache inside the sandbox, so verification ran with `GOCACHE` and `GOMODCACHE` redirected into `/tmp`.
- `gofmt` cannot format `README.md`; only the Go sources were passed to `gofmt`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 12 now has both machine-readable and human-readable benchmark evidence surfaces built on the same persisted bundle contract.
- Plan `12-04` can focus on reproducibility checks and milestone verification using the shared export and report outputs.

## Self-Check: PASSED

- Summary file created at `.planning/phases/12-token-attribution-and-evidence-reporting/12-03-SUMMARY.md`
- Task commits `f9eab78`, `9ef2efe`, and `26f78e9` exist in git history

---
*Phase: 12-token-attribution-and-evidence-reporting*
*Completed: 2026-03-16*
