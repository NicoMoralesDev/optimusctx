---
phase: 09-evaluation-harness-and-fixture-foundation
research_date: 2026-03-15
objective: "What do I need to know to PLAN this phase well?"
status: complete
---

# Phase 9 Research: Evaluation Harness and Fixture Foundation

## Executive Summary

Phase 9 is a foundation phase for the whole `v1.1` milestone. It should not try to prove every functional flow yet. Its job is to create the stable substrate that later validation and benchmark phases can reuse without rebuilding scenario state by hand.

The cleanest cut is four workstreams:

1. Define typed evaluation contracts plus versioned fixture repositories and scenario definitions.
2. Add a narrow `eval` runner surface that executes scenarios through the real shipped command boundary.
3. Persist evaluation runs and artifacts separately from refresh history.
4. Add one deterministic rerun path and documentation so later phases can extend the same harness instead of inventing new scripts.

The most important scope rule is separation: Phase 9 builds the harness, fixtures, and evidence plumbing, while Phase 10 fills that harness with milestone-grade functional scenarios.

## Repository Reality Check

### Planning inputs reviewed

- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `.planning/research/SUMMARY.md`
- `.planning/research/STACK.md`
- `.planning/research/FEATURES.md`
- `.planning/research/ARCHITECTURE.md`
- `.planning/research/PITFALLS.md`

### Current code seams that matter

- `internal/cli` already exposes the real operator surface: `init`, `refresh`, `doctor`, `snippet`, `pack`, `mcp`, `watch`.
- `internal/app` already owns orchestration services such as `RefreshService`, `WatchService`, `PackExportService`, `HealthService`, `BudgetAnalysisService`, and `TokenTreeService`.
- `internal/state/layout.go` already defines durable state locations under `.optimusctx/` with `logs/` and `tmp/`.
- `internal/store/sqlite` and `internal/store/migrations` are the correct place for new evaluation persistence.
- Existing integration tests already use temp repositories and command-boundary assertions; there is no existing benchmark or eval domain yet.

### Local guidance files

- `CLAUDE.md`: not present
- `.claude/skills/`: not present
- `.agents/skills/`: not present

## Phase Boundary

### In scope

- `EVAL-01`: lay the foundation for repeatable end-to-end CLI scenarios.
- `EVAL-04`: make scenarios rerunnable from versioned fixtures and scenario definitions without manual reconstruction.
- evaluation domain contracts
- fixture repository layout and scenario schema
- evaluation runner foundation
- evaluation artifact layout and persistence
- one rerunnable command and documentation path

### Explicitly out of scope for Phase 9

- full MCP scenario coverage for `EVAL-02`
- degraded and recovery scenario depth for `EVAL-03`
- benchmark methodology or token claims
- release or distribution work
- adding a second execution stack that bypasses the shipped CLI

## Recommended Technical Cut

## 1. Add an evaluation domain instead of ad hoc test-only structs

The codebase currently lacks a shared representation for:

- fixture repositories
- scenario definitions
- scenario steps
- run results
- artifact references

Phase 9 should add those contracts in `internal/repository` or another transport-neutral package so:

- CLI can load and validate scenarios
- app orchestration can execute runs
- SQLite can persist runs and step results
- later benchmark phases can reuse the same types

This should stay intentionally narrow. The initial schema only needs enough to express:

- fixture identity and version
- repository setup mode
- ordered steps
- expected command/tool target
- artifact capture metadata
- pass/fail plus timing/result summaries

## 2. Keep the runner CLI-first and deterministic

Research for `v1.1` already converged on a Go-first harness. For Phase 9, that means:

- add a dedicated evaluation command surface in `internal/cli`
- load scenarios from versioned files under repository-controlled `testdata`
- execute steps through the same command boundary the shipped product uses
- capture stdout, stderr, exit status, and run metadata deterministically

Avoid a shell-script-based harness as the primary implementation. Shell wrappers are useful later for docs or smoke runs, but the main milestone harness needs typed results and stable reuse.

## 3. Separate evaluation evidence from refresh history

Do not overload:

- `refresh_runs`
- `refresh_file_events`
- existing state metadata

Those tables prove repository maintenance behavior, not milestone evaluation evidence. Phase 9 should introduce dedicated evaluation persistence with:

- run table
- step table
- artifact table or artifact-path references

The state layout should also gain a stable eval artifact location. The cleanest options are:

- `.optimusctx/logs/eval/`
- `.optimusctx/eval/`

Given the current layout already has `logs/` and `tmp/`, a dedicated `logs/eval/` subdirectory is the lower-risk choice for this phase.

## 4. Version fixtures inside the repo

The current repo has parser-specific fixture files, but no evaluation fixture repository set. Phase 9 should introduce a milestone-owned fixture tree, for example:

```text
testdata/eval/fixtures/
testdata/eval/scenarios/
```

The fixture repos should be small but realistic enough to support later phases:

- one small Go application with multiple files and a clean baseline
- one repo state that can exercise `init`, `refresh`, `doctor`, and `pack export`
- deterministic commit or file-state setup instructions stored alongside the scenario definition

Do not start with many fixtures. One or two carefully chosen repos is enough for the foundation phase.

## 5. Plan for reruns from day one

`EVAL-04` is the key planning constraint. The runner should never assume manual prep done outside the repo. Phase 9 should therefore support:

- copying or materializing fixture repos into temp directories
- deterministic naming for output artifacts
- scenario selection by stable ID
- rerun against the same fixture input without hand-editing state

That can be achieved without full “resume” complexity. A simple `eval run --scenario <id>` flow that always reconstructs the fixture workspace is enough foundation for this phase.

## Likely File and Package Targets

### Most likely new files

- `internal/repository/eval.go`
- `internal/repository/eval_test.go`
- `internal/app/eval_runner.go`
- `internal/app/eval_runner_test.go`
- `internal/cli/eval.go`
- `internal/cli/eval_test.go`
- `internal/store/migrations/0004_eval_runs.sql`
- `internal/store/sqlite/eval.go`
- `internal/store/sqlite/eval_test.go`
- `testdata/eval/fixtures/...`
- `testdata/eval/scenarios/...`

### Most likely modified files

- `internal/cli/root.go`
- `internal/state/layout.go`
- `internal/state/layout_test.go`
- `README.md`

## Recommended Plan Split

### Plan 09-01: fixture repository set and scenario schema

Purpose:

- define the evaluation contracts
- add the first fixture repo set
- add versioned scenario definitions

This plan should own the domain shape so later plans do not compete over schema decisions.

### Plan 09-02: CLI evaluation runner foundation

Purpose:

- add `optimusctx eval` or equivalent narrow command surface
- load scenarios
- run ordered steps through the real command boundary

This is the harness entrypoint, not the place for persistence internals.

### Plan 09-03: evaluation artifact layout and persistence

Purpose:

- add eval artifact location to state layout
- add SQLite migrations and store APIs for runs, steps, and artifacts

This should stay independent enough from the runner to execute in parallel once the domain contracts exist.

### Plan 09-04: rerunnable scenario orchestration and docs

Purpose:

- connect fixture materialization, runner, and persistence into one rerunnable workflow
- document how later phases use it

This should depend on the runner and persistence foundations rather than redefining them.

## Risks and Guardrails

## 1. Building a benchmark product instead of a phase foundation

Risk:

- adding advanced reporting, dashboards, tokenizer variants, or benchmark UX before the foundation exists

Guardrail:

- Phase 9 only builds the substrate later phases need

## 2. Measuring private helpers instead of the shipped surface

Risk:

- the runner calls internal services directly and later claims CLI proof

Guardrail:

- scenario execution must route through the real command boundary for CLI flows

## 3. Mixing eval evidence with operational refresh state

Risk:

- evaluation data pollutes the product’s operational history tables

Guardrail:

- use separate persistence and artifact paths

## 4. Fixture sprawl

Risk:

- too many repos and too many scenario variants in the first phase

Guardrail:

- start with one or two realistic fixtures and stable scenario IDs

## 5. Hidden manual setup

Risk:

- scenarios only work because the developer prepared a repo out of band

Guardrail:

- each scenario must declare enough fixture setup to rebuild the workspace automatically

## Verification Guidance

Phase 9 should bias toward automated coverage around:

- scenario schema parsing and validation
- fixture materialization into temp repos
- CLI eval runner argument handling and step execution
- layout and migration correctness for eval artifacts
- repeatable reruns against the same fixture definition

Manual verification should be minimal. If any manual check remains, it should only confirm the documented command path is understandable, not the core logic.

## Validation Architecture

Phase 9 is Nyquist-friendly because it is mostly foundation and orchestration work with clear automated seams.

Recommended validation approach:

- quick loop:
  - `go test ./internal/repository ./internal/app ./internal/cli ./internal/state ./internal/store/sqlite`
- full loop:
  - `go test ./...`

Per-plan validation emphasis:

- 09-01: contract parsing and fixture-shape tests
- 09-02: runner execution and command-boundary tests
- 09-03: layout, migration, and SQLite persistence tests
- 09-04: rerun integration tests plus docs truth checks

Wave 0 is not needed if the phase stays on the existing Go test stack.

## Final Planning Guidance

The phase should finish with one thing that did not exist before:

- a committed, typed, rerunnable evaluation substrate that later phases can extend for real functional proof and benchmark evidence

If the plan starts drifting into “prove all CLI flows now,” it is too large. If it only adds fixture files without a runnable harness, it is too small.
