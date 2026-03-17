# OptimusCtx

## What This Is

OptimusCtx is a shipped local-first context runtime for coding agents. It maintains a persistent, incrementally refreshed representation of a repository and exposes that context through a vendor-neutral interface, with MCP as the primary integration layer. Shipped versions now cover repository discovery, persistent state, incremental refresh, structural extraction, layered exact-first context, MCP serving, optional watch mode, pack export, doctor diagnostics, evaluation flows, benchmark evidence, and narrow release channels.

## Core Value

Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## Current State

- Shipped version: `v1.1`
- Runtime stack: Go, SQLite, Tree-sitter, MCP-over-STDIO
- Delivered surface: `init`, `refresh`, `snippet`, `mcp serve`, `watch`, `pack export`, `doctor`, `eval`, `install`, `version`
- Product state: `v1.0` and `v1.1` are shipped, the `v1.1` audit passed, and functional, benchmark, and distribution evidence is archived

## Next Milestone Goals

- Extend benchmark depth with secondary tokenizer views and a watch-assisted edit-loop lane
- Broaden trustworthy distribution with native Linux packages, signed artifacts, and SBOM coverage
- Decide whether the next milestone stays on validation and distribution expansion or reopens core runtime capability work

## Requirements

### Validated

- ✓ Local-first persistent repository context runtime — `v1.0`
- ✓ MCP-first exact context delivery and bounded operational tooling — `v1.0`
- ✓ Developer workflow support for install, init, watch, doctor, pack export, and versioned release artifacts — `v1.0-v1.1`
- ✓ Fixture-backed CLI and MCP evaluation flows with persisted rerun evidence — `v1.1`
- ✓ Counted-input benchmark methodology, reproducible exports, and human-readable evidence — `v1.1`
- ✓ Supported distribution channels through GitHub Releases, Homebrew, Scoop, npm, and `npx` without implicit config writes — `v1.1`

### Active

- [ ] Secondary token metrics for named model tokenizers alongside the milestone-default estimator
- [ ] Watch-assisted edit-loop benchmarking once the non-watch baseline remains stable
- [ ] Native Linux package formats, signed artifacts, and SBOM metadata for released builds

### Out of Scope

- Hosted telemetry, dashboards, or managed services — the product remains local-first and proof-oriented.
- Default semantic retrieval or general-purpose RAG behavior — the wedge is still deterministic exact-first context optimization.
- Automatic modification of repository instruction files or client configs — installation and integration remain explicit.
- Broad installer sprawl without verification discipline — every new channel must stay truthful and test-backed.

## Context

v1.0 proved the core runtime wedge. v1.1 then proved the shipped product works end to end on fixture-backed CLI and MCP workflows, tightened benchmark claims around declared agent-facing inputs and comparable final artifacts, and expanded distribution through a narrow set of verifiable release channels. The codebase is now both product-bearing and evidence-bearing: functionality, benchmark methodology, and release policy are all backed by committed artifacts and verification files.

No milestone is currently active. The next planning cycle should start from fresh requirements rather than carrying forward the archived v1.1 scope file.

<details>
<summary>Archived v1.1 planning context</summary>

## Current Milestone: v1.1 Validation, Benchmarking, and Distribution

**Goal:** Prove the real-world value of OptimusCtx with functional evaluation, measurable A/B token and workflow savings, and a credible distribution plan.

**Target features:**
- Functional test flows that validate the runtime end to end in realistic agent workflows
- A/B benchmarking for token savings and work-speed improvements versus baseline repository exploration
- A solid technical distribution plan for adoption beyond the current local development setup

## Next Milestone Goals

- Prove functional correctness in end-to-end user and agent flows
- Quantify token savings and search-time reduction with repeatable benchmarks
- Define the distribution strategy, packaging shape, and rollout path for the tool

</details>

<details>
<summary>Archived pre-v1.0 project context</summary>

OptimusCtx exists to fix a repeated failure mode in agent-driven development: repository understanding gets rebuilt from scratch across sessions, context compressions, and broad exploratory tool calls. The intended wedge is a lightweight local runtime that precomputes and serves structured repository context so agents can avoid repeated scans, full-file reads, and broad traversal before doing targeted work.

The product is deliberately agent-agnostic. MCP is the universal contract, and vendor-specific instruction-file differences are treated as thin optional wrappers rather than separate product implementations. The product must remain useful even when users never paste the optional snippet into agent instructions.

The v1 scope centers on deterministic local indexing and delivery: repository discovery, ignore-aware traversal, hashing, structural extraction, symbol indexing, layered context outputs, token-cost analysis, exact lookup, MCP-over-STDIO serving, optional watch mode, pack/export, and operator diagnostics. The recommended implementation stack is Go for a cross-platform single-binary runtime, SQLite for persistent state, and Tree-sitter for structural extraction.

This repository is greenfield. The development process is expected to be heavily agent-driven under human supervision, with strong emphasis on stable command surfaces, stable storage and MCP contracts, incremental execution, and tests accompanying every meaningful feature.

</details>

## Constraints

- **Architecture**: Agent-agnostic, MCP-first runtime — The core product must work across mixed agent clients without separate vendor-specific implementations.
- **Privacy**: Local-first operation with no network requirement for core indexing — Repository contents should stay on the machine unless export is explicitly requested.
- **Behavior**: Deterministic and exact-first retrieval — Structural extraction, hashes, symbol maps, and explicit budgets take priority over speculative or semantic approaches in v1.
- **Performance**: Incremental refresh must be cheap — The runtime should recompute only what changed and keep stale checks much cheaper than full rebuilds.
- **Integration**: Non-invasive installation and setup — The tool must not automatically rewrite repository instruction files or policy files.
- **Distribution**: Simple cross-platform local runtime — The product should ship as a small command-surface binary that is easy for individuals and teams to install.
- **Scope**: v1 remains intentionally narrow — Advanced semantic, graph, enterprise, hosted, and multi-user capabilities are deferred to later versions.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use MCP as the primary integration layer | A single protocol keeps the runtime portable across coding-agent ecosystems | ✓ Shipped in v1.0 |
| Favor deterministic structural context over semantic retrieval in v1 | The initial wedge is precision, budget control, and predictable behavior | ✓ Shipped in v1.0 |
| Build the core runtime as a local-first single binary | Low-friction installation and cross-platform operation are core product requirements | ✓ Shipped in v1.0 |
| Use SQLite as the primary persistent store | The system needs structured local state, migrations, and queryable artifacts without external services | ✓ Shipped in v1.0 |
| Keep command surface intentionally small | The product should be easy to adopt and reason about for both humans and agents | ✓ Shipped in v1.0 |
| Make watch mode optional, not required | Daily usability should not depend on background processes or platform-specific watcher reliability | ✓ Shipped in v1.0 |
| Never auto-modify agent instruction files | Integration must remain explicit and non-invasive to preserve user control | ✓ Shipped in v1.0 |
| Keep evaluation and benchmark evidence fixture-backed and repo-local | Milestone claims need rerunnable evidence anchored in committed inputs and persisted outputs | ✓ Shipped in v1.1 |
| Count only declared agent-facing inputs in benchmark claims | Token savings must measure user-visible OptimusCtx value, not hidden system provenance | ✓ Shipped in v1.1 |
| Keep GitHub Releases as the canonical binary source and package managers as wrappers | Distribution breadth is useful only if every channel stays truthful to the same shipped runtime | ✓ Shipped in v1.1 |

---
*Last updated: 2026-03-17 after archiving v1.1*
