# Project Research Summary

**Project:** OptimusCtx
**Domain:** Local-first context optimization runtime for coding agents
**Researched:** 2026-03-14
**Confidence:** HIGH

## Executive Summary

OptimusCtx fits a clear systems-product pattern: a small local runtime that precomputes durable repository facts, serves exact-first context on demand, and avoids turning into a general-purpose RAG platform. The research strongly supports the product brief's original direction. The right v1 approach is a cross-platform single binary that keeps all durable state local, uses structural parsing instead of fuzzy retrieval, and exposes a narrow MCP-first surface for layered context access.

The recommended technical path is a Go runtime backed by SQLite and Tree-sitter, with incremental refresh as a first-class design constraint rather than a later optimization. The architecture should keep one canonical indexing pipeline that is shared by init, manual refresh, watch mode, and repair flows. This keeps correctness centered on persisted artifacts instead of ad hoc request-time analysis.

The main risks are not about choosing the wrong stack. They are product-discipline and systems-correctness risks: silent index drift, nondeterministic outputs, overscoping into semantic retrieval, and letting watch mode become the only path that feels correct. The roadmap should therefore front-load deterministic storage, invalidation, and exact query contracts before optional watch polish or richer export behaviors.

## Key Findings

### Recommended Stack

The strongest fit for v1 is Go `1.26.x` for the runtime, SQLite `3.52.x` through `modernc.org/sqlite` for local persistence, and Tree-sitter `0.25.x` for structural extraction. That combination matches the product's explicit goals: single-binary distribution, local-first operation, deterministic exact lookup, and cheap incremental refresh. `fsnotify` is appropriate for optional watch mode only when paired with reconciliation scans, and GoReleaser is the right downstream release path once the command surface stabilizes.

The main implementation discipline is to keep storage and parser contracts thin and explicit. SQLite should hold normalized repository facts and derived artifacts, not opaque blobs, and Tree-sitter support should begin with a narrow grammar set plus clear degraded behavior for unsupported files.

**Core technologies:**
- Go `1.26.x`: CLI, runtime, MCP server, refresh pipeline, and packaging-friendly single-binary delivery.
- SQLite `3.52.x` with `modernc.org/sqlite`: embedded persistent state for repository, file, symbol, budget, and health metadata.
- Tree-sitter `0.25.x`: deterministic structural parsing, symbol extraction, and exact span tracking.

### Expected Features

The research aligns closely with your original scope split. The launch product should prove persistent repository understanding, incremental cheap refresh, exact structural lookup, layered context serving, and a minimal MCP interface. Strong v1.x candidates are export packs, token-cost analysis, and optional watch mode. Hosted sync, automatic instruction rewriting, and default semantic retrieval should stay deferred because they dilute the wedge and expand trust and operational complexity.

**Must have (table stakes):**
- Persistent per-repository index — the product fails if context does not survive sessions.
- Incremental refresh — the runtime must be cheaper than repeated full rescans.
- Ignore-aware discovery — real repositories need deterministic exclusion handling.
- Deterministic structural extraction — exact-first context depends on reliable structure.
- Exact lookup by path and symbol — agents need targeted retrieval primitives.
- MCP-over-STDIO serving — portability across agent clients depends on a stable universal interface.
- Layered context outputs — context must be budget-aware and intentionally bounded.
- Health inspection and doctor diagnostics — local operation needs explainable failures and freshness reporting.
- Simple CLI lifecycle commands — install, init, serve, watch, and doctor are expected basics.

**Should have (competitive):**
- Portable context packs and export — reuse repository understanding across sessions and machines.
- Deterministic token-cost analysis — make context budget hotspots visible and actionable.
- Agent-agnostic setup discipline — keep vendor-specific instructions optional and thin.

**Defer (v2+):**
- Default semantic retrieval — keep it optional later, not core in v1.
- Hosted or shared cloud state — changes the product category too early.
- Full IDE or LSP replacement behaviors — outside the wedge and maintenance budget.

### Architecture Approach

The recommended system shape is four layers: operator surfaces (`init`, `doctor`, MCP, watch, pack`), application orchestration, a core runtime for scanning/change detection/extraction/context composition, and a persistence layer centered on SQLite plus artifact caches. The system should use one refresh pipeline regardless of whether the trigger is initialization, manual refresh, watch events, or repair. Query handlers should answer from persisted artifacts and explicit freshness metadata, not by re-parsing on demand.

**Major components:**
1. Repository scanner and ignore resolver — discovers the canonical indexable tree deterministically.
2. Change detector and refresh planner — computes stale units, invalidates affected artifacts, and schedules incremental work.
3. Extraction and composition pipeline — parses supported files, builds symbols and summaries, and produces layered context outputs.
4. Storage and migration layer — persists repository snapshots, artifact state, and compatibility boundaries.
5. MCP and CLI adapters — expose a narrow contract without leaking transport concerns into core logic.
6. Doctor and diagnostics service — reports freshness, parser coverage, configuration issues, and budget hotspots.

### Critical Pitfalls

1. **Silent index drift** — prevent it with explicit freshness state, repository-wide drift checks, and first-class rename/delete/ignore handling.
2. **Non-deterministic context output** — prevent it with stable traversal rules, explicit ordering, path normalization, and snapshot tests across repeated runs.
3. **Over-collecting instead of optimizing** — prevent it by refusing to turn v1 into a generic retrieval stack and requiring each stored artifact to justify its exact retrieval value.
4. **Watch mode becomes the source of truth** — prevent it by making manual refresh and init fully correct before watch mode exists.
5. **MCP contract leaks internal instability** — prevent it by versioning the tool surface, keeping payloads narrow, and using compatibility tests.

## Implications for Roadmap

Based on the research, the roadmap should build the product in dependency order: repository identity and storage first, refresh correctness second, structural extraction third, exact query and layered outputs fourth, MCP transport fifth, then watch/export hardening. This ordering matches both the architecture and the risk profile.

### Phase 1: Foundation and Repository State
**Rationale:** The product cannot be trusted until repository identity, traversal, storage, and schema versioning are stable.
**Delivers:** CLI skeleton, config loading, repository discovery, ignore handling, SQLite baseline, state directory creation, and doctor baseline.
**Addresses:** Persistent index, ignore-aware discovery, simple CLI lifecycle, health diagnostics.
**Avoids:** Silent index drift caused by weak repository identity and missing baseline state.

### Phase 2: Incremental Refresh and Storage Correctness
**Rationale:** Refresh behavior is core product value, and storage contracts must remain durable before higher-level features build on them.
**Delivers:** File hashing, subtree fingerprints, invalidation logic, cheap stale checks, rename/delete handling, migration discipline.
**Uses:** SQLite schema foundation and deterministic repository inventory.
**Implements:** Change detector and refresh planner.

### Phase 3: Structural Extraction and Exact Artifact Model
**Rationale:** Exact-first context depends on parser-backed structure and stable symbol/span artifacts.
**Delivers:** Language detection, initial Tree-sitter adapters, symbol extraction, relation baseline, repository map generation.
**Implements:** Extraction pipeline and normalized artifact storage.
**Avoids:** Over-collecting raw content and pretending unsupported languages are fully covered.

### Phase 4: Layered Context, Lookup, and Budget Controls
**Rationale:** Once exact artifacts exist, the product can expose the retrieval primitives agents actually need.
**Delivers:** L0/L1/L2 outputs, exact symbol and structure lookup, context block extraction, token-cost tree, bounded response shaping.
**Implements:** Context composer and budget layer.
**Avoids:** Non-deterministic or oversized responses that waste tokens.

### Phase 5: MCP Server and Stable Tool Contracts
**Rationale:** The integration layer should be built on stable core services rather than driving architecture prematurely.
**Delivers:** STDIO MCP server, machine-readable tool contracts, freshness metadata, consistent error semantics, compatibility tests.
**Implements:** MCP adapter and transport boundary.
**Avoids:** Contract churn and prompt-level client workarounds.

### Phase 6: Watch Mode, Packs, and Operator Hardening
**Rationale:** These features matter, but only after the core refresh and query path is already correct and portable.
**Delivers:** Optional watch mode, pack/export artifacts, richer diagnostics, performance profiling, and recovery flows.
**Uses:** Shared refresh pipeline and explicit storage/version contracts.
**Implements:** Watch coordinator, export packer, and hardening path.

### Phase Ordering Rationale

- Storage and refresh must precede parsing and serving because every higher-level capability depends on trustworthy persisted artifacts.
- Structural extraction should precede layered query work so the retrieval surface is built on exact spans and symbols rather than placeholder heuristics.
- MCP transport should come after query semantics are stable enough to freeze as a public contract.
- Watch mode and export belong later because they amplify correctness; they should not compensate for missing correctness.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 3:** Initial language support set and grammar versioning policy need explicit decisions.
- **Phase 4:** Token-estimation model and budget semantics need tighter specification.
- **Phase 5:** Exact MCP tool payloads and versioning policy need careful contract design.
- **Phase 6:** Export pack format and compatibility policy need explicit definition before implementation.

Phases with standard patterns:
- **Phase 1:** CLI scaffolding, SQLite baseline, repository discovery, and config/state management follow established local-tool patterns.
- **Phase 2:** Hash-based invalidation, migration testing, and stale-check flows are standard systems-work once the contracts are clear.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Strong fit to stated goals and backed by current ecosystem norms for local single-binary tools. |
| Features | HIGH | The feature split is already well defined in the project brief and reinforced by the research outputs. |
| Architecture | HIGH | The architecture follows clear dependency boundaries and matches the exact-first local runtime wedge. |
| Pitfalls | HIGH | The main failure modes are recognizable systems and product-discipline risks with clear mitigations. |

**Overall confidence:** HIGH

### Gaps to Address

- Initial language support set — decide the smallest useful v1 parser coverage before planning extraction work.
- Token estimation policy — decide whether v1 is model-agnostic only or supports pluggable tokenizer profiles.
- Relation extraction floor — define how much relationship data is mandatory beyond symbol ownership and imports.
- Pack/export format — define artifact shape, versioning, and portability guarantees before implementation.
- MCP registration scope — decide what install automates versus what remains manual and explicit.

## Sources

### Primary (HIGH confidence)
- `.planning/research/STACK.md`
- `.planning/research/FEATURES.md`
- `.planning/research/ARCHITECTURE.md`
- `.planning/research/PITFALLS.md`
- `.planning/PROJECT.md`

### Secondary (MEDIUM confidence)
- Go release and toolchain docs referenced in `STACK.md`
- SQLite release notes and driver docs referenced in `STACK.md`
- Tree-sitter parser and binding references referenced in `STACK.md`
- `fsnotify`, GoReleaser, and nFPM references referenced in `STACK.md`

---
*Research completed: 2026-03-14*
*Ready for roadmap: yes*
