---
phase: 09-evaluation-harness-and-fixture-foundation
plan: "02"
subsystem: testing
tags: [eval, cli, fixtures, go, verification]
requires:
  - phase: 09-01
    provides: canonical evaluation scenario and fixture contracts
provides:
  - stable `optimusctx eval` command surface
  - app-layer scenario runner with fixture materialization and step capture
  - CLI and app tests for scenario selection and execution failures
affects: [09-03, 09-04, phase-10-functional-validation]
tech-stack:
  added: []
  patterns: [thin CLI to app runner delegation, fixture copy plus git init materialization, in-process CLI command execution for eval steps]
key-files:
  created: [internal/app/eval_runner.go, internal/app/eval_runner_test.go, internal/cli/eval.go, internal/cli/eval_test.go]
  modified: [internal/cli/root.go]
key-decisions:
  - "Evaluation steps execute through the existing root command surface in-process so scenario results match shipped CLI behavior without shell-script glue."
  - "Fixture repositories are materialized by copying committed trees and initializing a fresh Git repository before scenario execution."
  - "EvalRunner applies NewEvalRunner defaults to partial configurations so tests and future extensions can override only the seams they need."
patterns-established:
  - "Eval CLI pattern: parse one scenario selector, resolve repo-owned scenario and fixture roots, then delegate execution to internal/app."
  - "Eval runner pattern: capture stdout, stderr, exit code, timing, and declared artifacts per step while failing with step-specific context."
requirements-completed: [EVAL-01]
duration: 12min
completed: 2026-03-16
---

# Phase 09 Plan 02: CLI Evaluation Runner Summary

**CLI-first evaluation runs now load versioned scenarios, materialize fixture repositories, and execute ordered steps through the shipped command boundary with captured step results.**

## Performance

- **Duration:** 12 min
- **Started:** 2026-03-16T00:00:00Z
- **Completed:** 2026-03-16T00:12:13Z
- **Tasks:** 3
- **Files modified:** 5

## Accomplishments
- Added `optimusctx eval` as a narrow command surface with `--scenario` and `--scenario-file` selectors.
- Implemented an app-layer runner that copies fixture trees, initializes Git state, and captures stdout, stderr, exit codes, timing, and declared artifacts per step.
- Locked the runner behavior into CLI and app tests covering real command-boundary execution, unknown scenario IDs, and step-level runtime failures.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add the `eval` CLI entrypoint** - `dda268d` (feat)
2. **Task 2: Implement app-layer scenario execution** - `968cbf6` (feat)
3. **Task 3: Lock runner argument and execution behavior into tests** - `986a98a` (test)

## Files Created/Modified
- `internal/cli/eval.go` - Parses the narrow eval command surface and delegates runs to the app layer.
- `internal/cli/root.go` - Registers `eval` in the shipped command list and help output.
- `internal/app/eval_runner.go` - Loads scenarios, materializes fixtures, runs CLI steps, and collects per-step result data.
- `internal/cli/eval_test.go` - Verifies selector handling, real command-boundary execution, and user-facing failure messages.
- `internal/app/eval_runner_test.go` - Verifies ordered scenario execution, result capture, default seam behavior, and runtime failure context.

## Decisions Made
- Ran eval steps through `NewRootCommand().Execute(...)` instead of an external shell process so the harness stays on the same shipped command boundary as operators.
- Materialized fixture repositories as copied working trees plus a fresh `git init`, which keeps committed fixtures small while still satisfying repository-root detection.
- Preserved a seam-based runner API in `internal/app` so later persistence and broader scenario orchestration can reuse the same core execution logic.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Sandbox writes to `/home/nico/.cache/go-build` were denied during verification, so all Go test commands were rerun with `GOCACHE` and `GOMODCACHE` redirected into `/tmp`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 09 now has a reusable CLI-first scenario runner that later plans can extend with persisted artifacts and broader scenario coverage.
- The command surface stayed intentionally narrow, so persistence and rerun orchestration can build on this foundation without redesigning the eval UX.

## Self-Check

PASSED
