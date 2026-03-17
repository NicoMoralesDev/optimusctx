---
phase: 10-functional-runtime-validation
plan: "03"
subsystem: testing
tags: [go, eval, cli, mcp, sqlite, runtime-validation]
requires:
  - phase: 09-evaluation-harness-and-fixture-foundation
    provides: shared fixture repositories, scenario loading, and persisted eval evidence storage
  - phase: 10-functional-runtime-validation
    provides: CLI and MCP eval runner support from plans 10-01 and 10-02
provides:
  - typed stale, watch-state, and refresh-failure scenario mutations
  - CLI stale diagnostics and MCP degraded/recovery scenarios
  - persisted eval evidence for stale, partially degraded, and fresh recovery transitions
affects: [10-04 milestone reporting, EVAL-03 evidence, future eval scenario additions]
tech-stack:
  added: []
  patterns: [typed eval state mutations, step-scoped refresh failure hooks, persisted eval artifact verification]
key-files:
  created:
    - testdata/eval/scenarios/05-cli-go-stale-v1.json
    - testdata/eval/scenarios/06-mcp-go-degraded-v1.json
    - testdata/eval/scenarios/07-mcp-go-recovery-v1.json
  modified:
    - internal/repository/eval.go
    - internal/repository/eval_test.go
    - internal/app/eval_runner.go
    - internal/app/eval_runner_test.go
    - internal/app/refresh.go
    - internal/cli/eval_integration_test.go
    - internal/cli/doctor_test.go
    - internal/cli/watch_test.go
key-decisions:
  - "Eval state transitions stay declarative through typed setup actions instead of shell hooks or direct test-only SQL."
  - "Stale proof uses doctor and seeded watch diagnostics, while degraded and recovery proof uses shipped MCP refresh, health, and repository_map responses."
  - "Recovery evidence must show both fresh generation advancement and updated symbol visibility, not only a successful refresh call."
patterns-established:
  - "Eval scenarios can seed repository freshness, watch status, and refresh-failure seams before a shipped CLI or MCP step."
  - "Failure-path integration tests read persisted eval artifacts under .optimusctx/eval/run-<id>/ instead of relying on console summaries alone."
requirements-completed: [EVAL-03]
duration: 19min
completed: 2026-03-16
---

# Phase 10 Plan 03: Stale, Degraded, and Recovery Scenario Coverage Summary

**Typed stale freshness seeding, watch-heartbeat diagnostics, and MCP degraded-to-recovery flows with persisted eval evidence on the shared harness**

## Performance

- **Duration:** 19 min
- **Started:** 2026-03-16T10:49:42Z
- **Completed:** 2026-03-16T11:08:40Z
- **Tasks:** 3
- **Files modified:** 11

## Accomplishments
- Added typed eval setup support for seeded watch status, persisted repository freshness transitions, and step-scoped refresh failure injection.
- Added fixture-backed CLI and MCP scenarios that prove stale, partially degraded, and recovered runtime states through shipped `doctor`, `refresh`, `health`, and `repository_map` surfaces.
- Locked the phase proof into persisted eval-artifact assertions plus stale doctor/watch formatter coverage so milestone reporting can reuse stored evidence directly.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add deterministic mutation support for stale and watch-state scenarios** - `2b1b900` (feat)
2. **Task 2: Add degraded and recovery scenario execution on shipped boundaries** - `aebdf45` (feat)
3. **Task 3: Persist state-transition evidence and lock it into tests** - `d486310` (test)

## Files Created/Modified
- `internal/repository/eval.go` - Extended the eval schema with repository-state seeding, watch-status seeding, and refresh-failure controls.
- `internal/app/eval_runner.go` - Applied typed state mutations in the shared runner and threaded step-scoped refresh failure hooks into real refresh execution.
- `internal/app/refresh.go` - Chained eval-controlled failure injection into the normal refresh service path.
- `internal/cli/eval_integration_test.go` - Added stale, degraded, recovery, and persisted-evidence integration coverage on repo-local eval artifacts.
- `internal/cli/doctor_test.go` - Locked stale freshness reporting into the CLI doctor formatter.
- `internal/cli/watch_test.go` - Locked stale watch-status rendering into CLI watch formatter coverage.
- `testdata/eval/scenarios/05-cli-go-stale-v1.json` - Declared the stale doctor/watch scenario on the shared eval harness.
- `testdata/eval/scenarios/06-mcp-go-degraded-v1.json` - Declared the degraded MCP refresh and last-good repository-map scenario.
- `testdata/eval/scenarios/07-mcp-go-recovery-v1.json` - Declared the degraded-to-recovery MCP flow with generation advancement and updated symbol visibility.

## Decisions Made
- Typed setup actions remained the only way to induce stale or degraded runtime state, so scenarios stay auditable and replayable.
- Repository freshness is seeded in persisted state only when the shipped runtime has no surface that can infer worktree drift on its own.
- Recovery proof stays on MCP `refresh`, `health`, and `repository_map`; no nonexistent MCP `doctor` or `watch` tools were introduced.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added repository-freshness state seeding to express stale runtime state**
- **Found during:** Task 2 (Add degraded and recovery scenario execution on shipped boundaries)
- **Issue:** `doctor` and `health` do not infer stale repository freshness from filesystem drift alone, so the plan's stale-path scenario could not be expressed truthfully with file mutations only.
- **Fix:** Added a typed `set_repository_state` eval action that updates persisted freshness metadata through the shared runner before the shipped surface is exercised.
- **Files modified:** `internal/repository/eval.go`, `internal/app/eval_runner.go`, `internal/repository/eval_test.go`, `testdata/eval/scenarios/05-cli-go-stale-v1.json`
- **Verification:** `go test ./internal/app ./internal/cli ./internal/mcp -run 'TestEvalStaleAndDegradedScenarios|TestEvalRecoveryScenarios'`
- **Committed in:** `aebdf45`

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** The added state seed kept the stale path declarative and reproducible without widening product scope.

## Issues Encountered
- Go verification initially failed under the sandbox because the default build cache path was not writable and the module cache was cold. Redirecting `GOCACHE` and `GOMODCACHE` into `/tmp` resolved local writes, and one approved dependency download completed the missing module fetches.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 10 now has persisted scenario evidence for `stale`, `partially_degraded`, and `fresh` recovery transitions on real CLI and MCP surfaces.
- Plan 10-04 can consume repo-local eval artifacts directly to build milestone reporting and requirement traceability for `EVAL-03`.

## Self-Check: PASSED

- Found summary path `.planning/phases/10-functional-runtime-validation/10-03-SUMMARY.md`
- Verified task commits `2b1b900`, `aebdf45`, and `d486310` exist in git history

---
*Phase: 10-functional-runtime-validation*
*Completed: 2026-03-16*
