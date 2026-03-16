---
phase: 09-evaluation-harness-and-fixture-foundation
plan: "04"
subsystem: testing
tags: [eval, cli, sqlite, fixtures, harness]
requires:
  - phase: 09-02
    provides: CLI eval scenario loading and command-boundary execution
  - phase: 09-03
    provides: repo-local eval artifact layout and sqlite persistence tables
provides:
  - fixture-backed rerun coverage for eval scenarios
  - persisted eval run metadata and copied artifacts under deterministic run directories
  - truthful README guidance for the shipped eval rerun workflow
affects: [phase-10-functional-runtime-validation, eval, docs]
tech-stack:
  added: []
  patterns: [fixture-backed temp workspace reruns, two-phase eval run persistence, repo-local eval evidence storage]
key-files:
  created: [internal/app/eval_service.go, internal/cli/eval_integration_test.go]
  modified: [internal/app/eval_runner.go, internal/app/eval_runner_test.go, internal/cli/eval.go, internal/repository/eval.go, README.md]
key-decisions:
  - "The default `optimusctx eval` path now persists runs in the source repository's `.optimusctx/eval/run-<id>` tree instead of leaving evidence in ephemeral temp workspaces."
  - "Eval runner pre-creates file-artifact parent directories so real CLI commands can write deterministic outputs without hidden manual setup."
  - "Rerun validation asserts deterministic repo-local paths and persisted metadata rather than byte-identical exported payloads, because pack artifacts include temp workspace roots."
patterns-established:
  - "EvalService opens the source repository store, persists a run header to obtain the run ID, then copies step outputs and scenario artifacts into the final run directory before updating sqlite rows."
  - "CLI integration tests verify reruns against real fixture repositories and inspect persisted eval evidence through the sqlite store."
requirements-completed: [EVAL-01, EVAL-04]
duration: 10min
completed: 2026-03-16
---

# Phase 9 Plan 04: Rerunnable Eval Workflow Summary

**Fixture-backed CLI eval reruns with persisted repo-local evidence and truthful operator documentation**

## Performance

- **Duration:** 10 min
- **Started:** 2026-03-16T00:14:30Z
- **Completed:** 2026-03-16T00:24:25Z
- **Tasks:** 3
- **Files modified:** 7

## Accomplishments

- Added end-to-end rerun coverage proving scenarios materialize fresh workspaces from committed fixtures and can be executed repeatedly by stable scenario ID.
- Connected the default `eval` command to repo-local sqlite persistence and copied artifact storage under `.optimusctx/eval/run-<id>/`.
- Documented the shipped `optimusctx eval --scenario ...` and `--scenario-file ...` workflow with real repository paths and rerun behavior.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add end-to-end rerun coverage for fixture-backed scenarios** - `8e4c57f` (test)
2. **Task 2: Wire artifact output and scenario rerun behavior together** - `d4897d9` (feat)
3. **Task 3: Document the eval workflow truthfully** - `eb8955e` (docs)

## Files Created/Modified

- `internal/app/eval_service.go` - Opens the source repository state store, persists eval runs, and copies evidence into deterministic run directories.
- `internal/app/eval_runner.go` - Records workspace paths in results and prepares file-artifact directories before executing scenario steps.
- `internal/app/eval_runner_test.go` - Covers fixture materialization and fresh-workspace reruns at the runner boundary.
- `internal/cli/eval.go` - Routes default eval execution through the persistence-backed app service.
- `internal/cli/eval_integration_test.go` - Verifies reruns and persisted artifact metadata through the real CLI boundary.
- `internal/repository/eval.go` - Extends eval run results with persisted run metadata and workspace context.
- `README.md` - Documents the versioned fixture/scenario workflow and repo-local eval evidence path.

## Decisions Made

- Persisted eval evidence in the source repository state tree so later phases can inspect prior runs even though scenarios execute in disposable temp workspaces.
- Used a two-phase save for eval runs so sqlite can assign the run ID before artifacts are copied into the final deterministic `run-<id>` directory.
- Kept rerun assertions focused on deterministic paths and persisted metadata because pack exports include temp workspace roots that make payload bytes vary across reruns.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Prepared file-artifact parent directories before command execution**
- **Found during:** Task 2 (Wire artifact output and scenario rerun behavior together)
- **Issue:** Real `pack export --output artifacts/pack.json` runs depended on a missing `artifacts/` directory and failed the rerun contract without manual setup.
- **Fix:** Added runner-side artifact directory preparation before executing each step that captures file artifacts.
- **Files modified:** `internal/app/eval_runner.go`
- **Verification:** `go test ./internal/app ./internal/cli -run 'TestEvalScenarioRerun|TestEvalScenarioMaterializesFixture|TestEvalArtifactsPersistAcrossRun'`
- **Committed in:** `d4897d9`

---

**Total deviations:** 1 auto-fixed (1 missing critical)
**Impact on plan:** The auto-fix was necessary for the real CLI rerun path to match the documented fixture-backed contract. No scope creep.

## Issues Encountered

- The sandbox blocked Go from using the default build cache under `/home/nico/.cache/go-build`, so verification used `GOCACHE=/tmp/optimusctx-gocache` and `GOMODCACHE=/tmp/optimusctx-gomodcache`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 10 can add broader CLI validation scenarios on top of one real rerunnable path with persisted evidence instead of inventing new harness plumbing.
- Eval reruns now leave inspectable sqlite rows and copied artifacts in deterministic run directories, which later report and benchmark work can query directly.

## Self-Check

PASSED

---
*Phase: 09-evaluation-harness-and-fixture-foundation*
*Completed: 2026-03-16*
