# Requirements: OptimusCtx

**Defined:** 2026-03-14
**Core Value:** Make repository understanding persistent, compact, incremental, and reusable across coding agents.

## v1 Requirements

### Installation and CLI

- [x] **CLI-01**: User can install the OptimusCtx runtime locally through a bootstrap path without modifying repository contents.
- [x] **CLI-02**: User can optionally register MCP configuration for supported clients during install, with explicit consent.
- [x] **CLI-03**: User can run `optimusctx init` in a repository and create the local project state directory successfully.
- [x] **CLI-04**: User can run `optimusctx snippet` and receive a manual-copy integration snippet without any file being modified automatically.
- [x] **CLI-05**: User can run `optimusctx doctor` and receive actionable installation, repository, state, and MCP diagnostics.

### Repository Discovery and State

- [x] **REPO-01**: Runtime can detect the repository root for the current working directory.
- [x] **REPO-02**: Runtime can collect indexable files while respecting ignore rules and common generated/vendor exclusions.
- [x] **REPO-03**: Runtime stores repository configuration, runtime metadata, and persistent index state in a project-local state directory.
- [x] **REPO-04**: Runtime persists repository, file, and directory metadata in SQLite with explicit schema versioning and migration support.
- [x] **REPO-05**: Runtime tracks per-file language, size, hash, last indexed time, and ignore status in persistent state.

### Refresh and Change Detection

- [x] **REFR-01**: Runtime can compute file hashes for indexed files and use them to detect changed content.
- [x] **REFR-02**: Runtime can compute directory or subtree fingerprints to support cheap stale checks.
- [x] **REFR-03**: Runtime can detect added, changed, deleted, and moved files during refresh.
- [x] **REFR-04**: Runtime can refresh only changed files and affected aggregates without rebuilding the full index.
- [x] **REFR-05**: Runtime can report whether project state is fresh, stale, or partially degraded before serving context.

### Structural Extraction and Symbols

- [x] **EXTR-01**: Runtime can detect supported languages for indexed files.
- [x] **EXTR-02**: Runtime can extract structural blocks and symbols from supported languages using deterministic parser-backed analysis.
- [x] **EXTR-03**: Runtime stores exact symbol spans, kinds, names, and parent relationships for supported files.
- [x] **EXTR-04**: Runtime degrades gracefully for unsupported or partially parsed files and surfaces coverage state in diagnostics.
- [x] **EXTR-05**: Runtime can generate a compact repository map from persisted structural artifacts.

### Context and Query Surface

- [x] **CTX-01**: Runtime can return an L0 repository snapshot with repository identity, dominant languages, major areas, and freshness metadata.
- [x] **CTX-02**: Runtime can return an L1 structural map with candidate files, symbols, concise summaries, and relevance-limiting metadata.
- [x] **CTX-03**: Runtime can return an L2 targeted context block with exact file paths, symbol or line-range targeting, and bounded surrounding code context.
- [x] **CTX-04**: Runtime can resolve exact symbol lookups by symbol name with optional path and language scoping.
- [x] **CTX-05**: Runtime can resolve exact structural lookups by supported pattern or normalized structural query.
- [x] **CTX-06**: Runtime can estimate token cost by file and directory and expose ranked context-budget hotspots.

### MCP Integration

- [x] **MCP-01**: Runtime can serve MCP requests over STDIO as the primary integration mode.
- [x] **MCP-02**: MCP tools return structured machine-readable payloads with freshness metadata and cache-versus-refresh status.
- [x] **MCP-03**: MCP surface includes repository map, symbol lookup, structure lookup, context block, token tree, refresh, pack, and health capabilities.
- [x] **MCP-04**: MCP handlers enforce bounded payload defaults and return transparent actionable failures.

### Watch, Export, and Operations

- [x] **OPS-01**: User can optionally run watch mode to keep the index fresh in the background without it being required for normal use.
- [x] **OPS-02**: Watch mode uses the same incremental refresh pipeline as manual refresh paths.
- [x] **OPS-03**: User can export a compact repository pack for offline or non-MCP workflows.
- [x] **OPS-04**: Pack export can fit output to a target budget using include/exclude rules.
- [x] **OPS-05**: Doctor output reports repository root detection, index freshness, watch status, storage health, parsing failures, and top token-cost paths.

## v1.1 Requirements

### Output Optimization

- **OUT-01**: Runtime can compact verbose tool outputs through deduplication, truncation, and line-focused extraction.
- **OUT-02**: Runtime can produce richer repository-level context-budget analysis and exclusion suggestions.
- **OUT-03**: Runtime can offer improved pack/export presets optimized for target token budgets.
- **OUT-04**: Runtime can expose richer symbol and reference relationships where extraction support is feasible.

## v2 Requirements

### Advanced Context Intelligence

- **ADV-01**: Runtime can support deeper language-specific graph extraction beyond baseline symbol ownership and imports.
- **ADV-02**: Runtime can support optional LSP-backed enrichment as a plugin or secondary layer.
- **ADV-03**: Runtime can assemble multi-hop context bundles based on query intent.
- **ADV-04**: Runtime can provide CI helpers and shared-workspace workflows for teams.
- **ADV-05**: Runtime can support an optional semantic layer as a plugin without making it a core dependency.

## Out of Scope

| Feature | Reason |
|---------|--------|
| Hosted service or cloud sync in v1 | Violates the local-first wedge and expands scope into infra, auth, and trust concerns too early |
| Default semantic or vector retrieval in v1 | Weakens the exact-first deterministic product promise and turns the product into generic RAG infrastructure |
| Automatic rewriting of repository instruction or policy files | Conflicts with the explicit non-invasive integration principle |
| Full IDE or LSP replacement behavior | Outside the product boundary; OptimusCtx should complement agents and existing dev tools |
| Multi-user shared index store in v1 | Adds distributed-state complexity before the single-user local runtime is proven |
| Analytics dashboard as a first-class v1 feature | Not required to validate the core runtime and integration wedge |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| CLI-01 | Phase 1 - Bootstrap, Repository Discovery, and Persistent State | Complete |
| CLI-02 | Phase 8 - Milestone Verification Backfill and Closure Evidence | Complete |
| CLI-03 | Phase 1 - Bootstrap, Repository Discovery, and Persistent State | Complete |
| CLI-04 | Phase 1 - Bootstrap, Repository Discovery, and Persistent State | Complete |
| CLI-05 | Phase 7 - Doctor Health Semantics and Milestone State Repair | Complete |
| REPO-01 | Phase 1 - Bootstrap, Repository Discovery, and Persistent State | Complete |
| REPO-02 | Phase 1 - Bootstrap, Repository Discovery, and Persistent State | Complete |
| REPO-03 | Phase 1 - Bootstrap, Repository Discovery, and Persistent State | Complete |
| REPO-04 | Phase 1 - Bootstrap, Repository Discovery, and Persistent State | Complete |
| REPO-05 | Phase 1 - Bootstrap, Repository Discovery, and Persistent State | Complete |
| REFR-01 | Phase 8 - Milestone Verification Backfill and Closure Evidence | Complete |
| REFR-02 | Phase 8 - Milestone Verification Backfill and Closure Evidence | Complete |
| REFR-03 | Phase 8 - Milestone Verification Backfill and Closure Evidence | Complete |
| REFR-04 | Phase 8 - Milestone Verification Backfill and Closure Evidence | Complete |
| REFR-05 | Phase 8 - Milestone Verification Backfill and Closure Evidence | Complete |
| EXTR-01 | Phase 3 - Structural Extraction and Repository Artifact Model | Complete |
| EXTR-02 | Phase 3 - Structural Extraction and Repository Artifact Model | Complete |
| EXTR-03 | Phase 3 - Structural Extraction and Repository Artifact Model | Complete |
| EXTR-04 | Phase 3 - Structural Extraction and Repository Artifact Model | Complete |
| EXTR-05 | Phase 3 - Structural Extraction and Repository Artifact Model | Complete |
| CTX-01 | Phase 4 - Layered Context, Exact Lookup, and Budget Analysis | Complete |
| CTX-02 | Phase 4 - Layered Context, Exact Lookup, and Budget Analysis | Complete |
| CTX-03 | Phase 4 - Layered Context, Exact Lookup, and Budget Analysis | Complete |
| CTX-04 | Phase 4 - Layered Context, Exact Lookup, and Budget Analysis | Complete |
| CTX-05 | Phase 4 - Layered Context, Exact Lookup, and Budget Analysis | Complete |
| CTX-06 | Phase 4 - Layered Context, Exact Lookup, and Budget Analysis | Complete |
| MCP-01 | Phase 8 - Milestone Verification Backfill and Closure Evidence | Complete |
| MCP-02 | Phase 8 - Milestone Verification Backfill and Closure Evidence | Complete |
| MCP-03 | Phase 8 - Milestone Verification Backfill and Closure Evidence | Complete |
| MCP-04 | Phase 8 - Milestone Verification Backfill and Closure Evidence | Complete |
| OPS-01 | Phase 7 - Doctor Health Semantics and Milestone State Repair | Complete |
| OPS-02 | Phase 8 - Milestone Verification Backfill and Closure Evidence | Complete |
| OPS-03 | Phase 8 - Milestone Verification Backfill and Closure Evidence | Complete |
| OPS-04 | Phase 8 - Milestone Verification Backfill and Closure Evidence | Complete |
| OPS-05 | Phase 7 - Doctor Health Semantics and Milestone State Repair | Complete |

**Coverage:**
- v1 requirements: 35 total
- Mapped to phases: 35
- Unmapped: 0
- Multi-mapped: 0

---
*Requirements defined: 2026-03-14*
*Last updated: 2026-03-15 after completing Phase 08 plan 04*
