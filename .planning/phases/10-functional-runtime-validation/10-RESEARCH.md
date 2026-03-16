---
phase: 10-functional-runtime-validation
research_date: 2026-03-16
objective: "What do I need to know to PLAN this phase well?"
status: complete
---

# Phase 10 Research: Functional Runtime Validation

## Executive Summary

Phase 10 should treat Phase 9 as complete infrastructure and spend its effort on scenario depth, not a second harness. The current eval system already gives the project versioned fixtures, scenario loading, workspace materialization, real CLI execution, repo-local evidence persistence, and deterministic reruns. The main planning question is therefore not how to build another test runner, but how to express MCP sessions and stale/degraded/recovery state transitions through the same evaluation contracts.

The cleanest cut is:

1. Extend the Phase 9 eval contract only where Phase 10 is blocked: MCP session steps, bounded output assertions, and controlled workspace or state mutations.
2. Fill the harness with end-to-end healthy-path CLI and MCP scenarios against the committed fixtures.
3. Reuse existing runtime seams for stale, degraded, and recovery validation instead of inventing synthetic behaviors.
4. Finish with milestone-grade reporting that maps scenario evidence back to `EVAL-02` and `EVAL-03`.

The most important scope rule is that Phase 10 proves the shipped runtime contract. It should not add new product capabilities, a second persistence system, or benchmark-oriented instrumentation.

## Repository Reality Check

### Planning inputs reviewed

- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `.planning/ROADMAP.md`
- `.planning/phases/09-evaluation-harness-and-fixture-foundation/09-RESEARCH.md`
- `.planning/phases/09-evaluation-harness-and-fixture-foundation/09-VALIDATION.md`
- `.planning/phases/09-evaluation-harness-and-fixture-foundation/09-01-PLAN.md`
- `.planning/phases/09-evaluation-harness-and-fixture-foundation/09-02-PLAN.md`
- `.planning/phases/09-evaluation-harness-and-fixture-foundation/09-03-PLAN.md`
- `.planning/phases/09-evaluation-harness-and-fixture-foundation/09-04-PLAN.md`

### Local guidance files

- `CLAUDE.md`: not present
- `.claude/skills/`: not present
- `.agents/skills/`: not present

### Existing harness capabilities from Phase 9

- Eval scenarios are versioned JSON under `testdata/eval/scenarios/`.
- Fixtures are committed under `testdata/eval/fixtures/<fixture>/<version>/repository`.
- The runner materializes a fresh temp workspace, copies the fixture tree, and runs `git init`.
- `optimusctx eval` executes scenario steps through the shipped CLI boundary, not private app calls.
- Evidence is persisted under `.optimusctx/eval/run-<id>/` and mirrored in `eval_runs`, `eval_steps`, and `eval_artifacts`.

### Current shipped MCP and operational contract

- `optimusctx mcp serve` is a stdio JSON-RPC server with readiness signaled on `stderr`.
- The shipped MCP tool surface is ten tools: `repository_map`, `layered_context_l0`, `layered_context_l1`, `symbol_lookup`, `structure_lookup`, `targeted_context`, `refresh`, `token_tree`, `pack`, and `health`.
- There is no MCP `doctor` tool and no MCP `watch` tool in the shipped contract.
- Operational runtime state already models repository freshness `fresh`, `stale`, and `partially_degraded`, plus watch status `absent`, `running`, and `stale`.

### Current contract limits that matter for planning

- Scenario steps are still CLI-only.
- Supported eval commands are only `init`, `refresh`, `doctor`, and `pack_export`.
- Scenario verification is mostly exit-code and artifact-presence based today.
- There is no first-class way to express an MCP stdio session.
- There is no first-class way to express controlled repository or state mutations between steps.

## Phase Boundary

### In scope

- `EVAL-02`: repeatable end-to-end MCP scenarios for `mcp serve` and the shipped query and ops tool surface
- `EVAL-03`: healthy, stale, degraded, and recovery scenario coverage
- deeper CLI functional proof on the existing harness
- controlled scenario-state transitions needed to create stale, degraded, and recovery conditions
- milestone-grade reporting that summarizes persisted eval evidence

### Explicitly out of scope

- new runtime features outside the shipped CLI, MCP, and operational contract
- a new benchmark harness or benchmark metrics
- distribution or packaging work
- arbitrary shell scripting as the primary validation engine
- adding ad hoc one-off integration tests that bypass the eval harness

## What Phase 10 Needs To Solve

## 1. Express MCP traffic without replacing the eval runner

The current eval runner is intentionally CLI-first, but `EVAL-02` requires proof that `optimusctx mcp serve` works over real stdio framing, including initialize, tool discovery, and tool calls. Phase 10 therefore needs one narrow harness extension for long-lived MCP sessions.

The recommended design is:

- keep the scenario loader, fixture materialization, run persistence, and artifact pipeline from Phase 9
- add an MCP-oriented step surface rather than a parallel runner subsystem
- execute `optimusctx mcp serve` through the shipped command boundary, but drive it with in-memory stdio pipes and the existing MCP framing helpers already used in integration tests
- persist the session transcript and selected tool responses as normal eval artifacts

This keeps MCP validation inside the shared eval harness instead of creating a separate integration framework.

## 2. Express runtime state changes as typed scenario actions

Phase 10 cannot prove stale, degraded, or recovery paths from static fixtures alone. It needs a narrow way to mutate repository or state inputs between steps.

The recommended design is to add constrained typed actions, not arbitrary scripts. The minimal useful action set is:

- write, overwrite, or delete a workspace file
- seed or modify `.optimusctx/tmp/watch-status.json` for watch-status and doctor-state scenarios
- optionally inject a known internal refresh failure only from the harness, using the existing shared failure seam, with the observed behavior still validated through the real CLI or MCP surface

This is the smallest credible way to express `stale`, `partially_degraded`, and `recovery` scenarios while preserving determinism.

## 3. Move from “captured output exists” to “captured output proves the contract”

Phase 9 proves that outputs and files can be captured and persisted. Phase 10 needs scenario assertions that prove the captured evidence means the runtime behaved correctly.

The most useful assertion types are:

- stdout or stderr contains a required substring
- JSON artifact contains a required field or exact value
- MCP response envelope metadata includes required freshness, cache, generation, or bounds fields
- repo-local eval evidence was persisted where the docs say it should be

Avoid snapshotting full human-readable output. Validate only the operator-facing lines and structured fields that are part of the shipped contract.

## 4. Treat operational validation as state diagnostics, not daemon orchestration

Operational flows already exist across `doctor`, `health`, `refresh`, and watch-status semantics. Phase 10 should validate those contracts directly.

It should prefer:

- `doctor` and `watch status` for watch liveness states
- `doctor` and `health` output or MCP tool results for repository freshness states
- `refresh` and MCP `optimusctx.refresh` for recovery proof

It should avoid turning Phase 10 into full background-process orchestration for `watch run` unless a concrete requirement is blocked without it.

## Standard Stack

- Go `go test` remains the execution and verification framework.
- The shared eval domain stays in `internal/repository/eval.go`.
- Eval orchestration stays in `internal/app/eval_runner.go` and `internal/app/eval_service.go`.
- Repo-local evidence stays under `state.Layout.EvalDir`, persisted by `internal/store/sqlite/eval.go`.
- MCP session validation should reuse `internal/mcp` protocol framing and server helpers, especially the patterns already exercised by `internal/mcp/integration_test.go`.
- CLI command-boundary execution should continue to route through `internal/cli` root command execution instead of bespoke subprocess wrappers where possible.

## Architecture Patterns

## 1. One harness, multiple surfaces

Keep one eval scenario model and one persistence model. Expand the surface vocabulary inside that model instead of introducing separate CLI and MCP validation systems.

## 2. Fixture plus mutation, not fixture explosion

Keep the committed base fixtures small and stable. Express stale and degraded states by applying deterministic typed mutations after workspace materialization, not by cloning many near-duplicate fixture repositories.

## 3. Structured evidence over golden transcripts

Store session transcripts and command outputs as artifacts, but assert against explicit fields, substrings, and status metadata. This keeps the suite robust while still proving user-visible behavior.

## 4. Observe failure paths through shipped boundaries

If a harness-only fault injection is needed, inject it behind the scenes and validate the resulting behavior through `refresh`, `doctor`, `watch status`, or MCP tool calls. Do not add user-facing fault-injection flags.

## Validation Architecture

Phase 10 should keep Phase 9’s repo-local eval layout and SQLite evidence schema as the single source of truth for validation runs. The recommended architecture has four layers:

1. Scenario contracts
   - extend the existing eval schema with minimal new primitives for MCP session steps, typed workspace or state mutations, and explicit assertions
2. Execution adapters
   - keep the current CLI step executor
   - add one MCP stdio session executor that drives `optimusctx mcp serve` with framed requests and captures transcript artifacts
3. Evidence capture
   - reuse `eval_runs`, `eval_steps`, and `eval_artifacts`
   - store MCP transcripts and structured tool payloads as file artifacts under `.optimusctx/eval/run-<id>/`
4. Reporting
   - aggregate persisted run results into milestone-facing summaries for `EVAL-02` and `EVAL-03`

The planning implication is important: Phase 10 probably does not need a new migration. `eval_steps.surface` and `eval_steps.command` are already plain text fields, and MCP transcripts can be stored as normal file artifacts. Only add schema if reporting truly needs new queryable dimensions.

## Don’t Hand-Roll

- Do not build a shell-script-driven MCP client when `internal/mcp` already has framing and integration-test patterns.
- Do not create a second artifact store outside `.optimusctx/eval/`.
- Do not encode large golden outputs for doctor, pack export, or MCP responses when targeted assertions will do.
- Do not turn scenario setup into arbitrary script execution.
- Do not add product-facing debug flags purely to make validation easier.
- Do not broaden Phase 10 into benchmark capture or distribution smoke tests.

## Common Pitfalls

## 1. Overfitting to current pretty-printed output

`doctor` and CLI summaries are operator-facing text. Assert on stable lines and statuses, not full rendered blocks.

## 2. Building two different validation systems

If MCP validation lands as standalone integration tests with different fixtures, persistence, and reporting, the milestone will lose the Phase 9 reuse it just built.

## 3. Creating degraded states that are not part of the shipped contract

Degraded scenarios should correspond to real runtime semantics already present in the codebase:

- repository freshness `stale`
- repository freshness `partially_degraded`
- watch status `absent` or `stale`
- doctor degraded summaries and recommended fixes

## 4. Requiring background-process management too early

Phase 10 can prove operational semantics through `doctor`, `watch status`, and persisted watch-state files without making `watch run` a core evaluation dependency.

## 5. Reporting only test names instead of requirement evidence

The milestone needs requirement proof, not just passing tests. Reporting should answer which scenarios prove `EVAL-02` and which prove `EVAL-03`, with artifact roots and rerun commands.

## Scenario Matrix To Plan Around

### Healthy CLI baseline

- `init -> refresh -> doctor -> pack export` on both committed fixtures
- validate pack artifact exists and basic JSON structure is sane
- validate doctor reports healthy or optional-watch semantics truthfully

### Healthy MCP baseline

- start `mcp serve`
- send `initialize`
- send `tools/list`
- call representative query tools
- call representative ops tools such as `optimusctx.refresh`, `optimusctx.pack`, and `optimusctx.health`
- assert readiness signal stays on stderr and tool payloads return the expected structured metadata
- do not plan scenarios around nonexistent MCP `doctor` or `watch` tools

### Stale paths

- mutate repository files after a successful refresh and prove stale semantics through doctor or MCP responses
- seed a stale watch-status record and prove `watch status` and `doctor` report the stale heartbeat correctly

### Degraded paths

- induce a partial refresh failure through the existing shared failure seam and verify the runtime reports `partially_degraded`
- validate last-good snapshot behavior by inspecting post-failure command or tool behavior rather than only internal state

### Recovery paths

- rerun `refresh` or `optimusctx.refresh` after the degraded condition is removed
- prove freshness returns to `fresh`
- prove the new generation advances and updated content becomes visible

## Code Examples

### Minimal MCP scenario shape to add

```json
{
  "schemaVersion": "optimusctx/eval-scenario@v1",
  "id": "mcp-go-basic-v1",
  "version": "v1",
  "fixture": {
    "id": "go-basic",
    "version": "v1",
    "path": "go-basic/v1/repository",
    "materialize": "copy_tree"
  },
  "steps": [
    {
      "id": "bootstrap",
      "kind": "command",
      "expect": { "surface": "cli", "command": "init", "exitCode": 0 }
    },
    {
      "id": "mcp-session",
      "kind": "mcp_session",
      "expect": { "surface": "mcp_stdio", "command": "serve", "exitCode": 0 },
      "captureArtifact": ["mcp-transcript", "health-response"]
    }
  ]
}
```

### Minimal typed mutation pattern to add

```json
{
  "id": "make-watch-stale",
  "kind": "workspace_mutation",
  "mutation": {
    "action": "write_file",
    "path": ".optimusctx/tmp/watch-status.json",
    "source": "fixtures/watch-status/stale.json"
  }
}
```

The point is not the exact field names. The point is that Phase 10 should add a small typed vocabulary for session execution and state mutation rather than arbitrary scripts.

## Recommended 4-Plan Split

### 10-01: end-to-end CLI workflow scenarios

Purpose:

- deepen the existing CLI eval scenarios into milestone-grade healthy-path proof
- add explicit assertions for stdout lines, persisted artifacts, and exported pack structure
- introduce only the smallest shared schema additions that both CLI and MCP plans need, such as assertion blocks

Why first:

- it keeps the first plan close to the current harness
- it establishes the common assertion model before MCP and degraded-state work depend on it

### 10-02: MCP serve and tool-flow scenarios

Purpose:

- add one MCP stdio execution adapter inside the eval harness
- cover `initialize`, `tools/list`, representative query tools, and representative ops tools on committed fixtures
- persist session transcripts and selected tool payloads as eval artifacts

Why second:

- `EVAL-02` is Phase 10’s new surface-area requirement
- this plan should reuse the same fixtures, artifact paths, and reporting format as the CLI scenarios

### 10-03: stale, degraded, and recovery scenario coverage

Purpose:

- add constrained typed scenario mutations for stale-watch, stale-repository, degraded-refresh, and recovery flows
- reuse existing freshness and watch semantics plus the existing shared refresh failure seam where deterministic degraded coverage needs it
- prove both CLI and MCP surfaces reflect these states correctly

Why third:

- it depends on the healthy-path surfaces already being executable
- it is the right place to concentrate state-transition complexity instead of scattering it across earlier plans

### 10-04: milestone-grade functional reports and verification

Purpose:

- aggregate persisted eval evidence into requirement-facing summaries for `EVAL-02` and `EVAL-03`
- document the rerun commands, scenario IDs, artifact roots, and current pass matrix
- produce the phase validation artifact from executed evidence rather than only test names

Why last:

- it should consume the actual runs from Plans `10-01` through `10-03`
- it keeps reporting truthful and tied to real persisted evidence

## Bottom Line

Phase 10 is mostly a scenario-authoring and validation-contract phase, not a harness-foundation phase. Plan it around one careful extension of the Phase 9 eval model for MCP sessions, assertions, and controlled mutations, then spend the remaining work proving healthy, stale, degraded, and recovery behavior through the real shipped CLI, MCP, and operational surfaces.
