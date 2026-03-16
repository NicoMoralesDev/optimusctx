---
phase: 11-a-b-benchmark-methodology-and-workflow-timing
research_date: 2026-03-16
objective: "What do I need to know to PLAN this phase well?"
status: complete
---

# Phase 11 Research: A/B Benchmark Methodology and Workflow Timing

## Executive Summary

Phase 11 should not build a generic benchmark platform. It should extend the shipped Phase 9 and Phase 10 evaluation harness with one narrow benchmark layer that can answer a specific milestone question:

- does OptimusCtx reduce discovery and context-assembly work on the same repository tasks
- does it reduce time to recover after a controlled repository change
- does it reduce end-to-end task completion time under fixed stop conditions

The main planning risk is not missing timers. The current eval system already records run and step start and finish timestamps, persists repo-local evidence, materializes fixtures deterministically, and supports controlled setup mutations. The real Phase 11 work is methodological:

1. define benchmark suites that pair one baseline workflow and one OptimusCtx workflow for the same task
2. freeze fixtures, mutations, stop conditions, and allowed baseline actions before timing anything
3. add benchmark-specific lane and metric contracts instead of overloading functional validation semantics
4. run repeated comparisons on both arms and verify the methodology itself before making claims

The cleanest phase cut is:

1. benchmark suite and baseline contract
2. discovery and context-assembly timing capture
3. refresh-after-change and end-to-end completion lanes
4. repeated-run comparison and benchmark verification

## Repository Reality Check

### Planning inputs reviewed

- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `.planning/ROADMAP.md`
- `.planning/research/SUMMARY.md`
- `.planning/research/STACK.md`
- `.planning/research/ARCHITECTURE.md`
- `.planning/research/PITFALLS.md`
- `.planning/phases/09-evaluation-harness-and-fixture-foundation/09-RESEARCH.md`
- `.planning/phases/10-functional-runtime-validation/10-RESEARCH.md`
- `.planning/phases/10-functional-runtime-validation/10-VALIDATION.md`
- `README.md`

### Local guidance files

- `CLAUDE.md`: not present
- `.claude/skills/`: not present
- `.agents/skills/`: not present

### Existing evaluation seams Phase 11 should reuse

- `internal/repository/eval.go` already defines versioned fixture, step, setup, assertion, artifact, and run-result contracts.
- `internal/app/eval_runner.go` already materializes fresh workspaces, applies deterministic setup mutations, records per-step timing, and executes shipped CLI or MCP flows.
- `internal/app/eval_service.go` already persists repo-local evaluation evidence under `.optimusctx/eval/run-<id>/`.
- `internal/store/sqlite/eval.go` already stores eval runs, steps, and artifacts with run and step timestamps.
- `internal/state/layout.go` already gives a durable eval artifact root under `.optimusctx/eval/`.

### Existing constraints that matter

- The current eval contract is built for functional proof, not paired A/B comparison.
- The current scenario model has no first-class concept of benchmark arm, lane, stop condition, attempt number, or comparison group.
- The current fixture repositories are good for functional validation but are very small for credible workflow-speed claims.
- Phase 12 owns machine-readable benchmark result storage and human-readable reporting depth, so Phase 11 should add only the persistence needed to make Phase 11 comparisons and Phase 12 reporting possible.

## What Phase 11 Must Decide Before Planning

## 1. What exactly counts as the baseline

This is the most important planning decision.

Phase 11 must not treat the baseline as an open-ended human workflow. That would make repeatability too weak for `BNCH-01`. The baseline needs a fixed, replayable contract.

Recommended rule:

- baseline workflows are encoded as typed, replayable repository-exploration actions
- the allowed baseline action set is narrow and explicit
- the allowed action set should cover only credible non-OptimusCtx exploration primitives such as:
  - list files or directories
  - run exact text search
  - open bounded file slices
  - inspect Git-tracked paths
- no arbitrary shell scripts as the primary baseline engine
- no baseline path that uses OptimusCtx artifacts, caches, MCP tools, or pack export outputs

If the planner does not lock this down in `11-01`, later timings will be hard to defend.

## 2. What exactly counts as the OptimusCtx treatment

The treatment path also needs to be fixed, not improvised per run.

Recommended rule:

- treatment workflows must use the shipped OptimusCtx surfaces only
- for workflow benchmarking, treatment steps should be drawn from the shipped CLI and MCP surfaces already exercised in Phase 10
- treatment steps may consume repository map, layered context, exact lookup, targeted context, refresh, token tree, pack export, and health only if the scenario contract says so
- no benchmark-only private helpers that bypass the shipped product boundary

## 3. What “search effort” means in Phase 11

Phase 11 is not the token-attribution phase. It still needs a concrete metric vocabulary for search effort so `BNCH-01` does not collapse into vague claims.

Recommended minimum search-effort metrics:

- broad search action count
- targeted lookup action count
- file-open action count
- total bytes read into the workflow context
- number of artifacts or files consulted before the stop condition is satisfied

These should be benchmark metrics in Phase 11. Token attribution by artifact type stays in Phase 12.

## 4. What benchmark lanes exist and how they stop

Phase 11 should benchmark lanes, not just one monolithic stopwatch.

Recommended lanes:

- discovery lane
  - start: first exploration action
  - stop: target file, symbol, or path is identified
- context-assembly lane
  - start: after discovery completes
  - stop: the bounded context package required by the task is assembled
- refresh-after-change lane
  - start: after a deterministic repository mutation is applied
  - stop: repository state is usable again for the same task
- task-completion lane
  - start: first workflow action for the task
  - stop: the task-defined completion artifact is produced

Every lane needs a stop condition that is machine-checkable. If the stop condition is subjective, the benchmark is not ready to plan.

## 5. Whether the current fixtures are enough

The current `go-basic` and `go-worktree` fixtures are sufficient for harness reuse and initial wiring. They are probably too small to support credible workflow-speed claims because discovery noise will dominate and baseline workflows may finish almost instantly.

Recommended planning assumption:

- keep the existing fixtures for smoke-level benchmark contract tests
- add one or two benchmark-oriented fixture versions with enough directory depth, symbol variety, and plausible ambiguity to make discovery and context assembly non-trivial
- keep the corpus small and frozen early

Phase 11 should freeze a small benchmark corpus rather than trying to benchmark many repositories.

## Phase Boundary

### In scope

- `BNCH-01`: fixed A/B methodology comparing a baseline workflow with an OptimusCtx-assisted workflow on the same tasks and repositories
- `BNCH-03`: repeatable timing capture for discovery, context assembly, refresh-after-change, and end-to-end completion
- benchmark suite contracts
- narrow baseline action vocabulary
- lane timing capture and repeated-run comparison
- benchmark verification that proves the method is consistent and rerunnable

### Explicitly out of scope

- token attribution by OptimusCtx artifact type beyond the minimum metadata needed for Phase 12
- final benchmark export bundle and milestone-ready report packaging
- hosted telemetry, dashboards, or remote benchmark collection
- watch-assisted benchmark lanes
- semantic retrieval additions to improve benchmark optics
- arbitrary scripting as the main benchmark engine

## Standard Stack

- Go benchmark orchestration in the existing repo and package structure
- the current eval fixture and workspace materialization pipeline from `internal/app/eval_runner.go`
- repo-local persistence under `.optimusctx/eval/` and SQLite for run evidence
- existing shipped CLI and MCP surfaces for the treatment arm
- a narrow benchmark domain adjacent to the eval domain for suites, lanes, and measurements
- `hyperfine` for repeated outer-loop timing samples where command-level repetition matters
- `benchstat` for repeated-run comparison summaries

## Architecture Patterns

## 1. One substrate, separate domains

Reuse the Phase 9 and Phase 10 evaluation substrate, but do not force benchmark semantics into the functional scenario model.

Recommended shape:

- keep fixture, setup-action, artifact, and workspace materialization helpers shared
- add benchmark-specific contracts for:
  - suite identity
  - baseline arm
  - treatment arm
  - lane definitions
  - stop conditions
  - measured metrics
  - repeated-run samples

The current eval model is centered on pass or fail functional steps. Phase 11 needs paired comparison semantics.

## 2. Paired scenarios, not free-form workflows

Each benchmark should be a paired suite:

- one baseline workflow
- one OptimusCtx workflow
- same fixture
- same task statement
- same mutation script if the lane includes repository change
- same stop condition

This is the core benchmark architecture. If the suite does not encode the pair, later comparison logic will drift.

## 3. Typed baseline actions, not shell freedom

Baseline actions should be typed and measurable, for example:

- `list_tree`
- `search_text`
- `read_file_slice`
- `git_list_files`
- `git_grep`
- `mark_target_found`
- `mark_context_ready`
- `mark_task_complete`

Each action should record:

- wall-clock start and finish
- paths touched
- bytes read
- whether it is a broad search or a targeted read

This makes baseline effort measurable without turning the benchmark runner into a shell runtime.

## 4. Treatment workflows should stay on the shipped product boundary

Treatment actions should reuse the same command and MCP execution adapters already used in functional validation. Phase 11 should measure actual product usage, not idealized internal-service calls.

## 5. Keep measurement at orchestration boundaries

Timing and effort metrics should be collected in the benchmark runner or service layer, not deep in storage or retrieval internals.

That keeps the runtime deterministic and keeps benchmark logic easy to verify.

## Recommended Domain Cut

### New benchmark contracts

Create benchmark-specific contracts in a dedicated file, preferably adjacent to the existing eval contracts, for example:

- `internal/repository/benchmark.go`
- `internal/repository/benchmark_test.go`

The contract should define:

- benchmark suite definition
- benchmark arm definition
- benchmark lane definition
- baseline action definition
- stop-condition definition
- benchmark measurement result
- repeated-run sample set

Reuse existing eval types where sensible:

- fixture references
- setup actions
- artifact references

Do not duplicate those contracts without need.

### Benchmark orchestration

Add benchmark orchestration beside the existing eval runner, for example:

- `internal/app/benchmark_runner.go`
- `internal/app/benchmark_service.go`

This layer should:

- materialize fixtures using the same patterns as eval
- execute baseline or treatment steps
- capture lane metrics
- persist enough run metadata for repeated-run comparison
- verify stop conditions

### Persistence guidance

Phase 11 should add only the persistence required for benchmark pairing and repeated-run comparison. The easiest planning split is:

- Phase 11 stores benchmark sample identity and lane metrics
- Phase 12 expands storage and export into machine-readable benchmark artifacts and human-readable summaries

Recommended persistence shape:

- add benchmark-specific tables or clearly separated benchmark records, not overloaded functional coverage rows
- allow each benchmark sample to reference:
  - fixture ID and version
  - suite ID
  - arm ID
  - lane ID
  - attempt number
  - run timestamps
  - summary metrics
  - linked eval run IDs or artifact roots when treatment execution reuses the eval runner

If the team wants to minimize schema churn, the smallest acceptable compromise is:

- keep detailed per-step evidence in existing eval artifacts
- add benchmark-indexed rows for pairing and comparison

Do not rely on `metadata_json` alone for all benchmark identity. Repeated-run comparison will need queryable suite, arm, lane, and attempt fields.

## Recommended Benchmark Lanes

## Discovery lane

Purpose:

- measure time and effort needed to identify the relevant file, path, symbol, or subsystem

Required stop condition examples:

- exact target file path found
- exact symbol key found
- exact directory or subsystem root identified

Recommended metrics:

- duration
- broad search count
- targeted lookup count
- file-open count
- bytes read

## Context-assembly lane

Purpose:

- measure time and effort needed to gather a bounded, sufficient context package for the task once the target is known

Required stop condition examples:

- required file set assembled
- required symbol excerpts assembled
- required pack or context artifact produced

Recommended metrics:

- duration
- bytes assembled
- number of context fragments
- files consulted after discovery

## Refresh-after-change lane

Purpose:

- measure how quickly each arm recovers after the same deterministic repository mutation

Required stop condition examples:

- fresh usable OptimusCtx state restored after `refresh`
- baseline arm re-discovers and reassembles task context after the mutation

Recommended metrics:

- mutation-to-ready duration
- actions needed after change
- bytes reread after change
- generation or freshness transitions for the treatment arm

## Task-completion lane

Purpose:

- measure the end-to-end time for the whole task, not just one sub-step

Required stop condition examples:

- benchmark answer artifact written
- target patch, report, or structured response produced
- task-specific verification assertion passes

Recommended metrics:

- total duration
- total action count
- total bytes read
- lane breakdown for discovery, context assembly, and post-change work

## Scenario Selection Guidance

Phase 11 should not start with many tasks. A small frozen corpus is better.

Recommended benchmark suite selection rules:

- at least one discovery-heavy task
- at least one context-assembly-heavy task
- at least one refresh-after-change task
- at least one end-to-end completion task
- every task must have an objective completion artifact or assertion
- every task must be runnable against a versioned fixture snapshot

Recommended task characteristics:

- task requires crossing multiple directories or files
- task has a clear “target found” condition
- task has a clear “context ready” condition
- task is realistic for a coding agent and not a synthetic micro-benchmark
- task does not require semantic retrieval to succeed

Recommended corpus size:

- 2 or 3 fixtures
- 1 or 2 benchmark tasks per fixture

Anything larger is likely too much for Phase 11.

## Plan Split Recommendation

### Plan 11-01: benchmark scenario selection and baseline rules

Purpose:

- define the benchmark suite contract
- freeze the benchmark corpus
- define the allowed baseline action vocabulary
- define lane and stop-condition rules

This plan should own methodology. Do not start timing before this is complete.

### Plan 11-02: workflow timing capture for discovery and context assembly

Purpose:

- implement benchmark runner timing for paired baseline and treatment arms
- capture the minimum search-effort metrics
- prove discovery and context-assembly lanes on the frozen corpus

This plan should focus on lane timing and metric capture, not comparison reporting depth.

### Plan 11-03: refresh-after-change and task-completion benchmark lanes

Purpose:

- add deterministic mutation support for benchmark suites
- benchmark recovery and end-to-end completion lanes
- ensure the same mutation and stop conditions apply to both arms

This plan should reuse the setup or mutation patterns already present in the eval harness.

### Plan 11-04: repeated-run comparison and benchmark verification

Purpose:

- run repeated samples for both arms
- compare repeated timings consistently
- verify the methodology catches mismatched repo state, mismatched steps, or stop-condition drift

This is where `hyperfine`, `benchstat`, or equivalent repeated-run comparison glue belongs.

## Validation Architecture

Phase 11 validation should prove both implementation correctness and methodological discipline. The validation stack should have four layers:

1. Contract validation
   - benchmark suite definitions validate required pairing fields, lane rules, stop conditions, and baseline action restrictions
   - invalid suites fail before execution

2. Runner invariants
   - tests prove both arms run against the same fixture snapshot and mutation script
   - tests prove lane timers start and stop at the correct orchestration boundaries
   - tests prove baseline metrics and treatment metrics are recorded under the same metric schema

3. Repeated-run consistency
   - tests prove attempt numbering, suite identity, arm identity, and lane identity remain queryable
   - tests prove repeated runs can be compared without guessing from free-form metadata

4. Benchmark verification
   - tests prove the methodology rejects mismatched stop conditions, mismatched fixtures, or treatment workflows that bypass the shipped surface
   - comparison logic should fail closed when the pairing rules are broken

Recommended validation targets:

- `internal/repository`
  - benchmark contract validation
- `internal/app`
  - benchmark runner timing boundaries
  - paired-arm invariants
  - mutation and stop-condition handling
- `internal/store/sqlite`
  - benchmark sample persistence and comparison queries
- `internal/cli`
  - rerunnable benchmark command or subcommand integration if Phase 11 adds one

Recommended sampling rhythm:

- after each benchmark-domain task: targeted package tests
- after each plan wave: `go test ./internal/repository ./internal/app ./internal/store/sqlite ./internal/cli`
- before Phase 11 verification: `go test ./...`

## Don’t Hand-Roll

- do not build a generic shell-script benchmark engine
- do not measure treatment workflows through private services instead of the shipped CLI or MCP surface
- do not use free-form human judgment as the stop condition
- do not rely only on `metadata_json` blobs for benchmark identity
- do not turn Phase 11 into token-attribution reporting or export packaging work
- do not benchmark watch-assisted flows yet
- do not expand fixtures endlessly; freeze a small corpus early

## Common Pitfalls

## 1. Baseline drift

If scenario authors can add arbitrary commands or vary behavior between attempts, the benchmark becomes a script contest instead of a product comparison.

## 2. Tiny repositories producing noisy speed claims

The current functional fixtures are likely too small for meaningful workflow-speed comparisons. Plan for at least one slightly richer benchmark fixture.

## 3. Subjective “done” states

“Enough context gathered” is not a valid stop condition unless the suite defines the exact context artifact or assertion that marks readiness.

## 4. Comparing warm and cold states unfairly

Phase 11 must define whether each lane starts from:

- a fresh materialized workspace
- an already initialized OptimusCtx repository
- a post-mutation state

The chosen state must be the same every run.

## 5. Mixing phase responsibilities

Phase 11 should establish the method and repeated timings. Phase 12 should turn that evidence into token attribution, machine-readable exports, and milestone-facing reports.

## 6. Hiding methodological assumptions in code

Allowed baseline actions, lane rules, and stop conditions should live in typed scenario contracts, not in undocumented runner behavior.

## Code Examples

### Recommended benchmark suite shape

```json
{
  "schemaVersion": "optimusctx/benchmark-suite@v1",
  "id": "go-worktree-discovery-v1",
  "fixture": {
    "id": "go-worktree",
    "version": "v2",
    "path": "go-worktree/v2/repository",
    "materialize": "copy_tree",
    "workspaceDir": "workspace"
  },
  "task": {
    "prompt": "Identify where runtime naming is defined and assemble the minimal context needed to explain the runtime/config relationship."
  },
  "lanes": [
    {
      "id": "discovery",
      "stopCondition": {
        "kind": "target_found",
        "targetPath": "internal/core/runtime.go"
      }
    },
    {
      "id": "context_assembly",
      "stopCondition": {
        "kind": "context_ready",
        "requiredPaths": [
          "internal/core/runtime.go",
          "pkg/config/config.go"
        ]
      }
    }
  ],
  "arms": [
    {
      "id": "baseline",
      "surface": "baseline",
      "steps": [
        {"kind": "git_list_files"},
        {"kind": "search_text", "query": "runtime"},
        {"kind": "read_file_slice", "path": "internal/core/runtime.go", "startLine": 1, "endLine": 40}
      ]
    },
    {
      "id": "optimusctx",
      "surface": "mcp",
      "steps": [
        {"kind": "mcp_tool", "tool": "optimusctx.repository_map"},
        {"kind": "mcp_tool", "tool": "optimusctx.symbol_lookup"},
        {"kind": "mcp_tool", "tool": "optimusctx.targeted_context"}
      ]
    }
  ]
}
```

### Recommended benchmark sample summary shape

```json
{
  "suiteId": "go-worktree-discovery-v1",
  "attempt": 3,
  "lane": "context_assembly",
  "arm": "optimusctx",
  "startedAt": "2026-03-16T12:00:00Z",
  "finishedAt": "2026-03-16T12:00:02Z",
  "durationMs": 1820,
  "broadSearchCount": 1,
  "targetedLookupCount": 2,
  "fileOpenCount": 1,
  "bytesRead": 1180
}
```

## Planning Implications

If Phase 11 is planned well, the work should assume:

- shared fixture and artifact plumbing from Phase 9 and Phase 10 stays in place
- benchmark methodology gets its own typed domain instead of stretching functional validation types too far
- at least minimal benchmark-specific persistence is needed in Phase 11, even though Phase 12 owns richer reporting and export
- the baseline must be replayable and constrained
- current tiny fixtures are not enough by themselves for credible workflow-speed evidence

That is the planning posture most likely to satisfy `BNCH-01` and `BNCH-03` without turning the milestone into a telemetry product.

## RESEARCH COMPLETE
