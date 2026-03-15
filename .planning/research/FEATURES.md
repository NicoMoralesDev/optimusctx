# Feature Research

**Project:** OptimusCtx
**Milestone:** v1.1 Validation, Benchmarking, and Distribution
**Researched:** 2026-03-15
**Confidence:** HIGH

## Focus

`v1.1` is not a new runtime-capability milestone. It is a proof milestone for the shipped `v1.0` runtime: validate end-to-end usefulness, measure savings against a baseline, and turn the existing local binary into something credibly distributable.

## Feature Expectations

| Area | Table Stakes | Differentiators | Anti-Features | Complexity | Depends On Shipped v1.0 |
|------|--------------|-----------------|---------------|------------|--------------------------|
| Functional validation | Repeatable end-to-end flows that exercise `init`, `refresh`, lookup/context tools, `mcp serve`, `doctor`, and optional `watch` on realistic repos | Scenario suite that mirrors real agent workflows instead of isolated command tests; includes degraded/stale-state recovery | Inventing new serving features before proving current flows; synthetic tests that do not resemble agent usage | MEDIUM | `init`, `refresh`, freshness states, exact lookup/context layers, `mcp serve`, `doctor`, `watch` |
| A/B token-savings measurement | A baseline-vs-OptimusCtx methodology with fixed tasks, fixed repos, fixed prompts, and published measurement rules | Per-task token delta by lookup mode (`repo map`, symbol lookup, L2 context, pack) plus explanation of where savings come from | Vague “feels smaller” claims; changing prompts/tools between control and treatment; mixing semantic retrieval into the test | MEDIUM | `CTX-01..06`, `MCP-02..04`, token tree, bounded payload defaults, pack export |
| Workflow-speed measurement | Time-to-answer or time-to-task-completion comparisons between broad repo exploration and exact-first retrieval | Split metrics into discovery time, context assembly time, and recovery time after repo changes or stale state | Measuring only raw command latency; relying on anecdotal demos instead of repeatable runs | MEDIUM | incremental refresh, freshness reporting, exact symbol/structure lookup, context blocks, optional watch |
| Distribution planning | Clear install shape, release channels, packaging targets, and adoption path for individual users and teams | Distribution plan that preserves the local-first single-binary story while covering MCP registration, docs, samples, and pack-based sharing | Hosted service scope, default daemonization, auto-editing instruction files, platform-specific deep integrations as milestone blockers | LOW-MEDIUM | existing single-binary CLI surface, manual `snippet`, optional MCP registration, `doctor`, pack export |

## Table Stakes

- Functional validation must prove the current shipped flows work on realistic repositories, not just fixture-scale unit coverage.
- Token-savings measurement must use a stable A/B harness with comparable tasks, prompts, and stopping rules.
- Workflow-speed measurement must show whether exact-first retrieval reduces search and context assembly time in practice.
- Distribution planning must answer how a new user installs, verifies, integrates, updates, and shares the tool without changing the product category.

## Differentiators

- OptimusCtx can benchmark a shipped deterministic runtime rather than a speculative prototype, which makes results more credible.
- The benchmark can expose savings by artifact type because `v1.0` already ships layered context and token-budget surfaces.
- Workflow validation can include stale, degraded, and watch-assisted paths because freshness and doctor semantics already exist.
- Distribution planning can stay agent-agnostic and non-invasive because `snippet`, MCP serving, and pack export are already shipped.

## Anti-Features

- Do not turn `v1.1` into a semantic-retrieval milestone to improve benchmark optics.
- Do not add hosted sync, cloud dashboards, or remote serving as prerequisites for distribution planning.
- Do not require automatic instruction-file edits or vendor-specific wrappers to claim distribution readiness.
- Do not treat benchmark instrumentation as permission to expand the command surface unless a measurement gap is real and blocking.

## Complexity Notes

| Topic | Complexity | Why |
|------|------------|-----|
| Functional validation | MEDIUM | Mostly orchestration and fixture design, but it must cover realistic repos, MCP flows, and degraded-state behavior. |
| Token-savings A/B | MEDIUM | Measurement discipline matters more than implementation; the hard part is controlling task shape and baseline behavior. |
| Workflow-speed measurement | MEDIUM | Requires repeatable timings across discovery, lookup, refresh, and possibly watch-assisted paths. |
| Distribution planning | LOW-MEDIUM | Primarily product/release design, but it depends on accurate install, upgrade, support, and packaging assumptions. |

## Dependencies On Current Shipped Capabilities

### Functional Validation

- Needs `init` and repository discovery to establish a reproducible starting state.
- Needs `refresh`, freshness reporting, and incremental change handling to validate live-repo behavior.
- Needs `mcp serve` plus lookup/context tools to test the main agent integration contract.
- Needs `doctor` and watch status reporting to validate failure handling and operational trust.

### A/B Token-Savings Measurement

- Needs L0/L1/L2 context outputs and exact symbol/structure lookup to define the OptimusCtx treatment path.
- Needs token-cost reporting to attribute savings to bounded context selection rather than hand-waving.
- Needs MCP bounded payload behavior so benchmark responses stay comparable and reproducible.
- Can use `pack export` as an alternate treatment path for non-MCP or offline comparisons.

### Workflow-Speed Measurement

- Needs the canonical incremental refresh pipeline to compare “refresh then query” against repeated manual exploration.
- Needs exact lookup primitives to measure search-effort reduction.
- Needs optional watch mode to test whether continuous freshness reduces task interruption in active-edit loops.
- Needs `doctor` to separate true workflow regressions from broken local setup.

### Distribution Planning

- Depends on the already-shipped small CLI surface and local state model; the plan should package what exists rather than redefine it.
- Depends on manual `snippet` and optional MCP registration for adoption flows across agent clients.
- Depends on `doctor` for first-run verification and supportability.
- Depends on pack export to describe sharing and offline workflows without introducing cloud scope.

## Recommended v1.1 Feature Cut

- Must ship: realistic functional validation suite, benchmark methodology, token-savings evidence, workflow-speed evidence, and a written distribution plan.
- Should ship: a small benchmark harness or reproducible script path that others can rerun locally.
- Nice to have: one or two polished reference workflows for major agent clients using the existing MCP/snippet path.
- Should not ship: new retrieval modes, hosted infrastructure, automatic repo modification, or broad packaging expansion that outruns product proof.

## Sources

- `.planning/PROJECT.md`
- `.planning/milestones/v1.0-REQUIREMENTS.md`
- `.planning/v1.0-v1.0-MILESTONE-AUDIT.md`
- Existing shipped command and capability framing in `README.md`

---
*Milestone-specific feature research for OptimusCtx v1.1*
