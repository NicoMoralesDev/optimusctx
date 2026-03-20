# OptimusCtx

## What This Is

OptimusCtx is a shipped local-first context runtime for coding agents. It maintains a persistent, incrementally refreshed representation of a repository and exposes that context through a vendor-neutral interface, with MCP as the primary integration layer. Shipped versions now cover repository discovery, persistent state, incremental refresh, structural extraction, layered exact-first context, MCP serving, optional watch mode, pack export, diagnostics, evaluation flows, benchmark evidence, narrow release channels, supported-client onboarding, and release automation.

## Core Value

Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## Current State

- Shipped version: `v1.3.3`
- Runtime stack: Go, SQLite, Tree-sitter, MCP-over-STDIO
- Delivered surface: `init`, `refresh`, `snippet`, `mcp serve`, `watch`, `pack export`, `status`, `doctor`, `eval`, `install`, `version`, and `release prepare`, with first-class supported-client onboarding for Claude and Codex hosts plus canonical GitHub Release-rooted downstream publication
- Product state: `v1.0`, `v1.1`, `v1.2`, `v1.3.1`, `v1.3.2`, and `v1.3.3` are published; `v1.3.4` is complete on the branch but intentionally skipped for release while `v1.3.5` folds in the missing MCP observability and status unification work

## Current Milestone

`v1.3.5` MCP observability and status unification

**Goal:** Make `status` the canonical operational surface, prove whether a registered MCP host actually discovered and used OptimusCtx, and register agent-usable guidance where host integrations support it.

**Target features:**
- Persist local MCP session evidence for `initialize`, `tools/list`, and `tools/call` so the product can distinguish registered vs discovered vs used
- Unify the useful `doctor` diagnostics into `status` and make `status` answer the real operator question: "is this actually working?"
- Register durable OptimusCtx usage guidance for the agent where host config formats support it, with explicit fallback truth where they do not

## Next Milestone Goals

- Deliver MCP observability that lets operators prove real host discovery and tool usage from OptimusCtx itself
- Collapse duplicated operational diagnostics into one authoritative `status` command
- Tighten supported-host integrations so agent guidance is actually consumable by the host when possible

## Requirements

### Validated

- ✓ Local-first persistent repository context runtime — `v1.0`
- ✓ MCP-first exact context delivery and bounded operational tooling — `v1.0`
- ✓ Developer workflow support for install, init, watch, doctor, pack export, and versioned release artifacts — `v1.0-v1.1`
- ✓ Fixture-backed CLI and MCP evaluation flows with persisted rerun evidence — `v1.1`
- ✓ Counted-input benchmark methodology, reproducible exports, and human-readable evidence — `v1.1`
- ✓ Supported distribution channels through GitHub Releases, Homebrew, Scoop, npm, and `npx` without implicit config writes — `v1.1`
- ✓ Guided release preparation with canonical version/tag proposal and preflight gating — `v1.2`
- ✓ Canonical GitHub Release-rooted downstream publication and selective rerun control — `v1.2`
- ✓ Canonical operator workflow for release, verification, rerun, and rollback — `v1.2`
- ✓ First-class supported-client onboarding and write-backed Claude/Codex integration with init-led command ownership — `v1.3.1`
- ✓ Smooth same-command `init` onboarding with focused client previews and aligned docs — `v1.3.2`
- ✓ Intent-led onboarding conversation, destination-first targeting, and outcome-oriented supported-client output — `v1.3.3`
- ✓ Release preflight secret verification, downstream publication-status truth, and better MCP runtime-handoff guidance — `v1.3.4`

### Active

- [ ] `status` becomes the authoritative command for runtime readiness, MCP discovery, and MCP usage evidence
- [ ] MCP session evidence proves whether a supported host only registered OptimusCtx, actually discovered it, and actually used its tools
- [ ] Supported-host onboarding registers durable agent-usable OptimusCtx guidance where the host supports it, with explicit fallback truth where it does not

### Out of Scope

- Hosted telemetry, dashboards, or managed rollout services — the product remains local-first and operator-driven.
- Default semantic retrieval or general-purpose RAG behavior — the wedge is still deterministic exact-first context optimization.
- Automatic modification of repository instruction files or client configs during install beyond explicit supported-host onboarding — installation and integration remain explicit.
- Additional first-class MCP hosts beyond `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` — `v1.3.5` fixes observability and guidance quality for the current host set rather than expanding coverage.
- New distribution channels beyond the currently supported set — `.deb`, `.rpm`, WinGet, Chocolatey, signing, and SBOMs stay deferred until the current channels are fully truthful and operator-safe.

## Context

v1.0 proved the core runtime wedge. v1.1 then proved the shipped product works end to end on fixture-backed CLI and MCP workflows, tightened benchmark claims around declared agent-facing inputs and comparable final artifacts, and expanded distribution through a narrow set of verifiable release channels.

v1.2 closed the operator loop around the release surface. v1.3.1 then finished the supported Claude and Codex onboarding story by delivering host-native preview/write behavior, correcting command ownership around `init`, and updating the docs/evidence to match the shipped contract. v1.3.2 tightened that operator experience further by collapsing the common bootstrap and onboarding path into one smooth interactive `init` flow while preserving explicit scripting and direct-flag usage. v1.3.3 refined that same onboarding path again by making the conversation intent-led and destination-first, while trimming avoidable noise from the result output and docs.

`v1.3.4` improved release truthfulness and clarified runtime handoff, but it still left the core adoption question unresolved inside the product: OptimusCtx itself still cannot prove whether the host discovered or actually used the MCP server, `status` and `doctor` still overlap heavily, and the new guidance mostly landed as human docs rather than durable agent-facing instructions consumed by the host. `v1.3.5` exists to correct that gap directly before any further release cut.

## Constraints

- **Architecture**: Agent-agnostic, MCP-first runtime — The core product must work across mixed agent clients without separate vendor-specific implementations.
- **Privacy**: Local-first operation with no hosted dependency for core indexing or observability — Repository contents and MCP evidence should stay on the machine unless export is explicitly requested.
- **Behavior**: Deterministic and exact-first retrieval — Structural extraction, hashes, symbol maps, explicit bounds, and concrete session evidence take priority over speculative or semantic approaches in v1.
- **Performance**: Incremental refresh and MCP observability must stay cheap — instrumentation cannot turn normal host usage into a heavy tracing system.
- **Integration**: Non-invasive installation and setup — The tool must not silently rewrite repository instruction files or unsupported host settings.
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
| Release automation must fail before publication when version, tag, or prerequisite checks are invalid | The operator workflow should be safe to start and cheap to abort before touching release channels | ✓ Shipped in v1.2 |
| Every downstream channel should derive from the same tag and release metadata contract | Multi-channel automation is only trustworthy if there is one source of truth for archives, checksums, and package metadata | ✓ Shipped in v1.2 |
| `init` is the onboarding front door for supported clients | Repository bootstrap and host onboarding should feel like one coherent operator flow, while explicit flags remain available for automation and direct control | ✓ Shipped in v1.3.2 |
| Onboarding prompts should speak in terms of user intention and destination, not backend implementation jargon | The init flow should optimize for operator comprehension first, while still preserving precise direct-control escape hatches | ✓ Shipped in v1.3.3 |
| Downstream release channels must be operator-truthful even when they do not publish | A release flow that silently degrades into `skipped` publication is too easy to misread; channel readiness and outcomes need first-class visibility | ✓ Completed in v1.3.4 |
| Runtime handoff guidance alone is not enough; OptimusCtx must expose proof of host discovery and use | Human docs cannot substitute for product-visible observability and host-consumable guidance | — Active in v1.3.5 |

---
*Last updated: 2026-03-20 after starting milestone v1.3.5 MCP observability and status unification*
