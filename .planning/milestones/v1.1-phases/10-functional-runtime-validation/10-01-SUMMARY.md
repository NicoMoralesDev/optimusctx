---
phase: 10-functional-runtime-validation
plan: 01
subsystem: testing
tags: [go, cli, eval, fixtures]
requires:
  - phase: 09-evaluation-harness-and-fixture-foundation
    provides: rerunnable eval scenarios, persisted eval run storage, and repository-local evidence layout
provides:
  - typed eval setup and assertion contracts for CLI workflow scenarios
  - runner-enforced stdout, stderr, and JSON artifact assertions across real CLI steps
  - committed healthy-path CLI scenarios with persisted `.optimusctx/eval/run-<id>/` evidence
affects: [phase-10-runtime-validation, eval-harness, cli-integration]
tech-stack:
  added: []
  patterns: [typed eval assertions, bounded step setup actions, committed fixture-backed CLI validation]
key-files:
  created:
    - testdata/eval/fixtures/go-basic/v1/repository/go.mod
    - testdata/eval/fixtures/go-basic/v1/repository/main.go
    - testdata/eval/fixtures/go-worktree/v1/repository/go.mod
    - testdata/eval/fixtures/go-worktree/v1/repository/cmd/app/main.go
    - testdata/eval/fixtures/go-worktree/v1/repository/internal/core/runtime.go
  modified:
    - internal/repository/eval.go
    - internal/repository/eval_test.go
    - internal/app/eval_runner.go
    - internal/app/eval_runner_test.go
    - internal/cli/eval_integration_test.go
    - testdata/eval/scenarios/01-cli-go-basic-v1.json
    - testdata/eval/scenarios/02-cli-go-worktree-v1.json
key-decisions:
  - "Eval assertions stay contract-focused: bounded stdout/stderr substring checks plus JSON field presence or equality on captured artifacts."
  - "Scenario setup stays workspace-bounded with only write, overwrite, and delete actions so future degraded-path plans can reuse one transport-neutral contract."
  - "CLI integration coverage copies committed eval fixtures into temp repositories, keeping end-to-end eval runs isolated while still exercising shipped scenario files."
patterns-established:
  - "Eval workflow scenarios express proof inline through `assert` blocks rather than artifact-exists-only checks."
  - "Persisted eval evidence is verified by scenario id and artifact id, not by SQLite row order."
requirements-completed: []
duration: 10m
completed: 2026-03-16
---

# Phase 10 Plan 01: Executable CLI Runtime Validation Summary

**Typed eval assertions and bounded setup actions now drive committed `init -> refresh -> doctor -> pack export` CLI scenarios with persisted repository-local evidence.**

## Performance

- **Duration:** 10 min
- **Started:** 2026-03-16T10:16:04Z
- **Completed:** 2026-03-16T10:26:01Z
- **Tasks:** 3
- **Files modified:** 12

## Accomplishments
- Extended the shared eval schema with transport-neutral step `setup` actions and targeted `assert` primitives for stdout, stderr, and JSON artifacts.
- Upgraded the eval runner to execute bounded workspace setup and fail workflows on missing contract signals instead of only exit-code or file-presence mismatches.
- Added committed Go fixture repositories plus end-to-end CLI scenario coverage that persists evidence under repository-local `.optimusctx/eval/run-<id>/`.

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend eval scenario contracts for assertions and bounded setup** - `4ecdefd` (feat)
2. **Task 2: Upgrade the CLI runner to execute assertions against real command output** - `d1742a9` (feat)
3. **Task 3: Add healthy-path CLI scenarios for the shipped workflow** - `39770a1` (feat)

**Verification fix:** `8fc8ad6` (fix: stabilize persisted artifact assertions after full-suite verification)

## Files Created/Modified
- `internal/repository/eval.go` - Adds typed eval setup and assertion schema primitives plus validation helpers.
- `internal/repository/eval_test.go` - Covers schema round-tripping, assertion contracts, and invalid setup or artifact references.
- `internal/app/eval_runner.go` - Applies per-step setup actions, evaluates stdout/stderr or JSON artifact assertions, and surfaces step-scoped failures.
- `internal/app/eval_runner_test.go` - Verifies CLI workflow setup, assertion evaluation, and artifact persistence compatibility.
- `internal/cli/eval_integration_test.go` - Runs committed CLI scenarios end to end in temp repositories and checks persisted `.optimusctx/eval` evidence.
- `testdata/eval/scenarios/01-cli-go-basic-v1.json` - Asserts healthy init, refresh, doctor, and pack-export signals for the basic fixture.
- `testdata/eval/scenarios/02-cli-go-worktree-v1.json` - Asserts the same shipped workflow on a nested worktree fixture.
- `testdata/eval/fixtures/go-basic/v1/repository/main.go` - Provides real tracked source content for the basic CLI scenario.
- `testdata/eval/fixtures/go-worktree/v1/repository/cmd/app/main.go` - Provides nested entrypoint content for the richer worktree scenario.

## Decisions Made
- Used step-local assertions rather than full transcript snapshots so eval scenarios prove stable operator-facing contract lines and structured manifest fields.
- Kept setup actions file-system-only and workspace-bounded instead of introducing shell hooks or transport-specific runner behavior.
- Validated persisted artifacts by artifact id in integration tests because SQLite row ordering is not part of the product contract.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Removed artifact-order assumptions from CLI persistence checks**
- **Found during:** Task 3 (Add healthy-path CLI scenarios for the shipped workflow)
- **Issue:** The new end-to-end persistence test assumed the pack artifact would be the last SQLite artifact row, which is not guaranteed.
- **Fix:** Updated the integration assertion to locate persisted pack artifacts by `artifact_id` and verify stored paths from that lookup.
- **Files modified:** `internal/cli/eval_integration_test.go`
- **Verification:** `go test ./internal/repository ./internal/app ./internal/cli`
- **Committed in:** `8fc8ad6`

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** The fix only hardened verification; shipped runtime behavior and scope remained unchanged.

## Issues Encountered
- Planning traceability mismatch: `10-01-PLAN.md` references `EVAL-03`, but `REQUIREMENTS.md` defines `EVAL-03` as healthy, stale, degraded, and recovery validation. This plan delivered the healthy-path foundation only, so requirement completion remains deferred to later Phase 10 plans.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Healthy-path CLI runtime validation is now reusable for stale, degraded, recovery, and milestone-reporting scenarios without changing the harness contract again.
- Persisted eval evidence remains under repository-local `.optimusctx/eval/run-<id>/`, ready for later plans to mine for richer reporting.
- `EVAL-03` should stay pending until Phase 10 covers stale, degraded, and recovery paths or the requirement mapping is corrected.

## Self-Check: PASSED
