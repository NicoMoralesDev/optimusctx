# Roadmap: OptimusCtx

**Created:** 2026-03-14
**Project:** OptimusCtx
**Granularity:** Standard
**Scope covered:** v1 only

## Roadmap Principles

- Build in dependency order for a greenfield Go/SQLite/Tree-sitter/MCP runtime.
- Keep one canonical local indexing pipeline that is reused by init, refresh, watch, export, and MCP-triggered operations.
- Favor deterministic persisted artifacts over request-time recomputation.
- Keep v1 exact-first and local-first; do not let semantic or hosted concerns leak into roadmap scope.

## Phase 1: Bootstrap, Repository Discovery, and Persistent State

**Goal:** Establish the local runtime, repository identity, ignore-aware discovery, and durable project-local state so the system has a trustworthy base to build on.

**Why now:** Nothing else is credible until repository detection, state layout, and persistent metadata contracts are stable.

**Mapped v1 requirements:** CLI-01, CLI-03, CLI-04, REPO-01, REPO-02, REPO-03, REPO-04, REPO-05

**Success criteria:**
- A user can install the binary, run `optimusctx init`, and get a project-local state directory without modifying repository instruction files.
- The runtime resolves the repository root from nested working directories and persists repository identity and runtime metadata in SQLite.
- File discovery respects ignore rules plus common generated/vendor exclusions and records per-file metadata including language hint, size, hash slot, last indexed time, and ignore status.
- Schema versioning and migrations are exercised by tests so an empty database and an upgraded database both initialize cleanly.
- `optimusctx snippet` returns a manual-copy integration snippet and performs no file writes outside the local state directory.

**Plan progress:**
- Completed: `01-01`, `01-02`, `01-03`, `01-04`
- Remaining: none
- Summary coverage: 4 of 4 Phase 1 plans completed

## Phase 2: Incremental Refresh and Freshness Model

**Goal:** Make refresh cheap, correct, and explicit by implementing hash-driven change detection, subtree fingerprints, and freshness state.

**Why now:** Incremental correctness is a core product promise and must exist before parser-backed artifacts or serving layers rely on stored state.

**Mapped v1 requirements:** REFR-01, REFR-02, REFR-03, REFR-04, REFR-05

**Success criteria:**
- Refresh computes per-file hashes and subtree fingerprints and uses them to distinguish unchanged and changed repository areas.
- Manual refresh correctly identifies added, changed, deleted, and moved files and invalidates only the affected records.
- Incremental refresh updates changed files and dependent aggregates without rebuilding the entire repository index.
- The runtime exposes explicit freshness states of `fresh`, `stale`, or `partially degraded` before context is served.

**Plan progress:**
- Completed: `02-01`, `02-02`, `02-03`
- Remaining: `02-04`
- Summary coverage: 3 of 4 Phase 2 plans completed

## Phase 3: Structural Extraction and Repository Artifact Model

**Goal:** Add deterministic parser-backed structural extraction and persist exact symbols, spans, and repository-map building blocks.

**Why now:** Exact-first context depends on stable structural artifacts rather than ad hoc file scanning.

**Mapped v1 requirements:** EXTR-01, EXTR-02, EXTR-03, EXTR-04, EXTR-05

**Success criteria:**
- The runtime detects supported languages for indexed files and routes supported files through Tree-sitter-based extraction.
- Supported files persist exact symbol names, kinds, spans, and parent relationships in normalized SQLite tables.
- Unsupported or partially parsed files degrade gracefully, remain queryable as files, and surface coverage gaps in diagnostics metadata.
- A compact repository map can be generated entirely from persisted structural artifacts without reparsing the repository on demand.

## Phase 4: Layered Context, Exact Lookup, and Budget Analysis

**Goal:** Expose the exact retrieval primitives agents need: layered context views, exact symbol/structure lookup, and budget-aware context shaping.

**Why now:** Once persisted structural artifacts exist, the runtime can safely expose bounded context outputs without inventing heuristics.

**Mapped v1 requirements:** CTX-01, CTX-02, CTX-03, CTX-04, CTX-05, CTX-06

**Success criteria:**
- The runtime returns L0, L1, and L2 outputs with deterministic ordering, freshness metadata, and bounded payload sizes.
- Exact lookup resolves symbols by name with optional path/language scope and resolves structure queries through normalized structural patterns.
- Targeted context blocks include exact file paths and symbol or line-range anchors with bounded surrounding code context.
- Token-cost analysis ranks expensive files and directories and exposes actionable budget hotspots from persisted metadata.

## Phase 5: MCP Serving and Integration Contracts

**Goal:** Turn the core runtime into a stable MCP-first product surface with machine-readable payloads, bounded defaults, and explicit client registration.

**Why now:** The transport contract should sit on top of already-stable indexing and query semantics rather than driving them.

**Mapped v1 requirements:** CLI-02, MCP-01, MCP-02, MCP-03, MCP-04

**Success criteria:**
- The runtime serves MCP over STDIO and exposes repository map, symbol lookup, structure lookup, context block, token tree, refresh, pack, and health capabilities.
- MCP responses are structured, machine-readable, and consistently include freshness and cache-versus-refresh status metadata.
- Payload defaults are bounded and oversized requests fail with transparent actionable errors instead of silent truncation or unstable output.
- Install flow can optionally register supported client MCP configuration only when the user explicitly opts in.

## Phase 6: Watch Mode, Pack Export, and Operational Diagnostics

**Goal:** Harden the operator experience with optional background freshness, portable export workflows, and comprehensive diagnostics.

**Why now:** These features amplify the core runtime; they should land after the indexing and serving pipeline is already correct.

**Mapped v1 requirements:** CLI-05, OPS-01, OPS-02, OPS-03, OPS-04, OPS-05

**Success criteria:**
- Watch mode is optional, uses the same incremental refresh pipeline as manual refresh, and never becomes the only correct path.
- Users can export a compact repository pack for offline or non-MCP workflows with explicit include/exclude controls and target-budget fitting.
- `optimusctx doctor` reports installation, repository detection, freshness, watch status, storage health, parsing failures, and top token-cost paths in one actionable output.
- Recovery and diagnostics flows make degraded parser, storage, and refresh states visible without requiring database inspection.

## Requirement Coverage

| Phase | Requirement count | Requirements |
|-------|-------------------|--------------|
| Phase 1 | 8 | CLI-01, CLI-03, CLI-04, REPO-01, REPO-02, REPO-03, REPO-04, REPO-05 |
| Phase 2 | 5 | REFR-01, REFR-02, REFR-03, REFR-04, REFR-05 |
| Phase 3 | 5 | EXTR-01, EXTR-02, EXTR-03, EXTR-04, EXTR-05 |
| Phase 4 | 6 | CTX-01, CTX-02, CTX-03, CTX-04, CTX-05, CTX-06 |
| Phase 5 | 5 | CLI-02, MCP-01, MCP-02, MCP-03, MCP-04 |
| Phase 6 | 6 | CLI-05, OPS-01, OPS-02, OPS-03, OPS-04, OPS-05 |

**Coverage validation**
- v1 requirements total: 35
- v1 requirements mapped: 35
- Unmapped v1 requirements: 0
- Multi-mapped v1 requirements: 0

## Phase Order Rationale

1. Phase 1 establishes repository identity, state layout, and storage contracts.
2. Phase 2 makes refresh and freshness trustworthy before higher-level artifacts depend on them.
3. Phase 3 adds deterministic structural facts on top of stable persisted inventory.
4. Phase 4 exposes exact-first query and context products from those persisted artifacts.
5. Phase 5 freezes the MCP-facing contract only after the underlying semantics are stable.
6. Phase 6 adds watch, export, and operator hardening without using them to compensate for missing correctness.

---
*Last updated: 2026-03-14 after completing plan 02-03*
