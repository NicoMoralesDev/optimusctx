---
phase: 09-evaluation-harness-and-fixture-foundation
plan: 01
subsystem: testing
tags: [go, evaluation, fixtures, scenarios, cli]
requires:
  - phase: 08-milestone-verification-backfill-and-closure-evidence
    provides: verification baselines and shipped CLI surfaces that the evaluation schema now targets
provides:
  - typed evaluation scenario, fixture, artifact, and run-result contracts
  - committed versioned fixture repositories for deterministic eval setup
  - committed scenario definitions and repository tests that lock schema loading
affects: [phase-09-02, phase-09-03, phase-09-04, phase-10]
tech-stack:
  added: []
  patterns: [versioned eval schema files under testdata/eval, transport-neutral repository contracts for evaluation]
key-files:
  created:
    - internal/repository/eval.go
    - internal/repository/eval_test.go
    - testdata/eval/scenarios/01-cli-go-basic-v1.json
    - testdata/eval/scenarios/02-cli-go-worktree-v1.json
  modified: []
key-decisions:
  - "Phase 9 scenario definitions use JSON files with an explicit schemaVersion so later CLI and persistence layers can load one canonical contract."
  - "The initial command surface is intentionally narrow: init, refresh, doctor, and pack_export with CLI-only sequencing rules."
  - "Fixture references must encode stable id and version segments in their paths so reruns stay deterministic from committed inputs."
patterns-established:
  - "Evaluation contracts live in internal/repository as transport-neutral types shared by runner and persistence work."
  - "Committed fixture repos and scenario definitions live under testdata/eval/{fixtures,scenarios} with versioned directories."
requirements-completed: [EVAL-04]
duration: 18min
completed: 2026-03-15
---

# Phase 9 Plan 01: Evaluation Harness and Fixture Foundation Summary

**Typed evaluation contracts plus versioned fixture repositories and scenario JSON definitions for deterministic CLI reruns**

## Performance

- **Duration:** 18 min
- **Started:** 2026-03-15T23:40:00Z
- **Completed:** 2026-03-15T23:58:19Z
- **Tasks:** 3
- **Files modified:** 13

## Accomplishments
- Added one canonical evaluation schema in `internal/repository` for fixtures, scenarios, ordered steps, artifacts, and run results.
- Added two committed fixture repositories and two stable scenario definitions covering the shipped `init`, `refresh`, `doctor`, and `pack export` CLI flows.
- Locked schema loading, fixture references, and malformed sequence failures into repository tests so downstream runner and persistence plans can distinguish contract drift from execution failures.

## Task Commits

Each task was committed atomically:

1. **Task 1: Define transport-neutral evaluation contracts** - `0a924a1` (feat)
2. **Task 2: Add committed fixture repositories and scenario definitions** - `03dbf86` (feat)
3. **Task 3: Lock schema and fixture loading into tests** - `2ee9f0a` (test)

## Files Created/Modified
- `internal/repository/eval.go` - Canonical evaluation contracts, scenario loading, and validation helpers.
- `internal/repository/eval_test.go` - Schema, ordering, and committed fixture-reference regression tests.
- `testdata/eval/fixtures/go-basic/v1/repository` - Small Go fixture repository for basic CLI validation flows.
- `testdata/eval/fixtures/go-worktree/v1/repository` - Slightly richer nested fixture repository for deterministic refresh and pack export paths.
- `testdata/eval/scenarios/01-cli-go-basic-v1.json` - Stable basic CLI scenario definition.
- `testdata/eval/scenarios/02-cli-go-worktree-v1.json` - Stable nested-worktree CLI scenario definition.

## Decisions Made
- Used JSON scenario files with explicit `schemaVersion` instead of ad hoc test structs so later CLI and SQLite layers can load the same contract.
- Kept the first schema intentionally narrow and CLI-focused rather than modeling benchmark metrics or MCP-specific details before those phases exist.
- Enforced `init -> refresh -> doctor/pack_export` sequencing in repository validation so malformed scenario files fail early and explicitly.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 09-02 can load committed scenarios directly without inventing new fixture or step types.
- Phase 09-03 can persist run and artifact records against stable scenario IDs and artifact references.
- No blockers identified for downstream Phase 9 plans.

## Self-Check

PASSED

Verified summary claims:
- `internal/repository/eval.go` exists.
- `internal/repository/eval_test.go` exists.
- `testdata/eval/fixtures` exists.
- `testdata/eval/scenarios` exists.
- Task commits `0a924a1`, `03dbf86`, and `2ee9f0a` exist in git history.

---
*Phase: 09-evaluation-harness-and-fixture-foundation*
*Completed: 2026-03-15*
