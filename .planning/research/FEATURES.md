# Feature Research

**Domain:** Local-first context optimization runtime for coding agents
**Researched:** 2026-03-14
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Persistent per-repository index | Without persistence, the product does not solve repeated repository re-understanding across sessions | MEDIUM | Requires durable local storage, repository identity, schema migrations, and rebuild/resume semantics; depends on SQLite or equivalent embedded store |
| Incremental refresh for changed files | Users expect a context runtime to be cheaper than rescanning the repo on every use | HIGH | Requires hashing, dirty detection, dependency invalidation, and partial recomputation paths; depends on traversal, metadata store, and extraction pipeline |
| Ignore-aware repository discovery | Developers expect `.gitignore`-style behavior and sane defaults for generated/vendor content | MEDIUM | Requires deterministic traversal rules, override support, and stable path normalization across platforms |
| Deterministic structural extraction | The stated product wedge is exact-first context, so symbol and file structure extraction is foundational | HIGH | Requires Tree-sitter or similar parsers, language selection strategy, fallback behavior, and normalized structural output |
| Exact lookup by path/symbol | Agents need targeted retrieval instead of broad file reads; exact lookup is the practical serving primitive | MEDIUM | Depends on persistent symbol/file index, stable identifiers, and MCP methods that expose exact queries cleanly |
| MCP-over-STDIO serving | Vendor-neutral agent integration is expected if the product claims agent portability | MEDIUM | Requires stable tool contracts, versioning, and low-latency response shaping; STDIO transport keeps installation simple |
| Layered context outputs | Agents need different context sizes for different budgets, and users expect budget-aware context assembly | HIGH | Depends on deterministic ranking/packaging rules, token estimation, and composable output tiers such as repo, module, and symbol level |
| Health inspection and doctor diagnostics | Local developer tools are expected to explain stale indexes, parser failures, and config problems | MEDIUM | Depends on observability hooks, index metadata, and actionable error reporting rather than opaque failures |
| Simple CLI lifecycle commands | Users expect install/init/build/watch/doctor/export commands from a local runtime | LOW | Mostly command-surface and orchestration work, but critical for adoption; depends on core indexing and serving flows existing underneath |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Portable context packs/export | Makes repository understanding reusable across sessions, machines, and agent clients without rescanning | MEDIUM | Depends on stable serialization format, compatibility/versioning policy, and import/export validation |
| Deterministic token-cost analysis | Gives operators predictable budget control instead of heuristic prompt growth | MEDIUM | Depends on layered outputs, tokenizer integration or approximation, and transparent accounting metadata |
| Agent-agnostic integration model | Avoids lock-in to one vendor's instruction-file conventions and keeps the runtime durable as clients change | MEDIUM | Depends on MCP-first design and thin optional wrappers rather than deep vendor-specific branching |
| Optional watch mode instead of required daemon | Preserves local-first simplicity while still supporting fast feedback for active repos | MEDIUM | Depends on incremental refresh core; should degrade cleanly to manual refresh when watchers are unreliable |
| Non-invasive setup | Differentiates from tools that rewrite repo instructions or impose workflow changes | LOW | Depends more on product discipline than hard engineering; requires explicit boundaries in CLI and docs |
| Exact-first context assembly | Competes on predictability and repository fidelity rather than fuzzy retrieval quality | HIGH | Depends on strong structural extraction, stable ranking, and disciplined refusal to substitute semantic guesses for explicit structure |
| Cross-agent cache reuse | Lets teams amortize indexing cost across Codex, Claude Code, Gemini CLI, and future MCP clients | HIGH | Depends on stable contracts, portable storage semantics, and compatibility guarantees across clients and versions |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Default semantic/vector retrieval | Sounds like a modern AI feature and promises "smarter" search | Pulls the product toward generic RAG, increases nondeterminism, complicates storage/runtime costs, and weakens the exact-first wedge | Keep v1 structural and exact-first; add optional semantic augmentation only after the deterministic core proves value |
| Hosted cloud sync/service by default | Seems attractive for team sharing and remote access | Breaks local-first/privacy positioning, adds operations burden, and expands scope into auth, sync conflicts, and tenancy | Support explicit pack/export workflows first; revisit hosted sync only after strong local adoption |
| Automatic rewriting of agent instruction files | Feels convenient during setup | Violates the project's non-invasive constraint and creates trust/review issues in user repos | Provide copy-paste snippets and doctor guidance instead of automatic modification |
| Always-on background daemon requirement | Suggests faster responses and "just works" behavior | Raises platform complexity, resource usage, and failure modes for a tool that should remain simple and explicit | Make watch mode optional and allow one-shot commands plus STDIO serving |
| Full IDE/LSP feature expansion | Users often request navigation, diagnostics, and editor-like intelligence in one tool | Dilutes the wedge and turns the runtime into a broad developer platform with much larger maintenance surface | Stay focused on repository context production and delivery; integrate with existing IDE/LSP tools rather than replacing them |
| Automatic code modification based on inferred context | Appears to close the loop from context to action | Crosses into autonomous editing/product policy, raises safety expectations, and muddies the product boundary | Serve high-fidelity context to agent clients and let those clients own action selection and editing |

## Feature Dependencies

```text
[Ignore-aware repository discovery]
    └──requires──> [Repository identity and local config]

[Persistent per-repository index]
    └──requires──> [Embedded storage schema]

[Deterministic structural extraction]
    └──requires──> [Ignore-aware repository discovery]
    └──requires──> [Language parser integration]

[Incremental refresh]
    └──requires──> [Persistent per-repository index]
    └──requires──> [File hashing and change detection]
    └──requires──> [Deterministic structural extraction]

[Exact lookup by path/symbol]
    └──requires──> [Persistent per-repository index]
    └──requires──> [Deterministic structural extraction]

[Layered context outputs]
    └──requires──> [Exact lookup by path/symbol]
    └──requires──> [Token-cost analysis]

[MCP-over-STDIO serving]
    └──requires──> [Exact lookup by path/symbol]
    └──requires──> [Layered context outputs]
    └──requires──> [Health inspection and doctor diagnostics]

[Portable context packs/export]
    └──requires──> [Persistent per-repository index]
    └──requires──> [Stable serialization/versioning]

[Optional watch mode]
    └──enhances──> [Incremental refresh]

[Default semantic/vector retrieval]
    ──conflicts──> [Exact-first context assembly]

[Hosted cloud sync/service by default]
    ──conflicts──> [Local-first runtime]
```

### Dependency Notes

- **Deterministic structural extraction requires ignore-aware repository discovery:** structural indexing quality depends on excluding generated, vendor, and ignored paths before parse work starts.
- **Incremental refresh requires persistent per-repository index:** change detection is only valuable if previous file fingerprints and extracted artifacts are stored durably.
- **Incremental refresh requires deterministic structural extraction:** partial recomputation still needs a reliable extraction pipeline for changed files.
- **Exact lookup by path/symbol requires deterministic structural extraction:** exact serving depends on stable symbol tables and normalized file metadata.
- **Layered context outputs require exact lookup by path/symbol:** context assembly needs dependable low-level retrieval primitives before it can package higher-level views.
- **Layered context outputs require token-cost analysis:** budget-aware output tiers are only trustworthy if cost is measurable and exposed.
- **MCP-over-STDIO serving requires layered context outputs:** the protocol is more useful when it can return different context depths rather than one fixed blob.
- **Portable context packs/export requires stable serialization/versioning:** portable artifacts become liabilities without explicit compatibility rules.
- **Optional watch mode enhances incremental refresh:** watchers improve freshness and ergonomics but should sit on top of a solid manual refresh path.
- **Default semantic/vector retrieval conflicts with exact-first context assembly:** fuzzy retrieval changes operator expectations and weakens determinism as the product's primary promise.
- **Hosted cloud sync/service by default conflicts with local-first runtime:** remote coordination adds security and systems concerns that materially change the product category.

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [ ] Persistent per-repository index — core proof that repository understanding survives across sessions
- [ ] Ignore-aware repository discovery — necessary for accurate, efficient indexing on real repos
- [ ] Deterministic structural extraction — core wedge for exact-first context
- [ ] Incremental refresh — validates the runtime is materially cheaper than repeated full scans
- [ ] Exact lookup by path/symbol — minimal useful retrieval primitive for coding agents
- [ ] Layered context outputs — enables practical use under prompt/token constraints
- [ ] MCP-over-STDIO serving — proves vendor-neutral integration model
- [ ] Health inspection and doctor diagnostics — required to make local operation debuggable
- [ ] Simple CLI lifecycle commands — required for installability and daily use

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] Portable context packs/export — add when users need handoff, CI artifacts, or cross-machine reuse
- [ ] Optional watch mode — add when manual refresh proves too frictionful in active editing loops
- [ ] Deterministic token-cost analysis — add when users begin tuning budgets across multiple clients
- [ ] Thin vendor-specific setup helpers — add only if MCP alone leaves onboarding gaps for major clients

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] Optional semantic augmentation — defer until the exact-first runtime is stable and well understood
- [ ] Team sync/collaboration flows — defer until local export/import usage shows where sharing truly hurts
- [ ] Hosted management or remote index serving — defer because it changes privacy, ops, and product scope significantly
- [ ] Rich policy/automation layers on top of context — defer until the runtime has a stable contract worth automating against

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Persistent per-repository index | HIGH | MEDIUM | P1 |
| Incremental refresh | HIGH | HIGH | P1 |
| Deterministic structural extraction | HIGH | HIGH | P1 |
| Exact lookup by path/symbol | HIGH | MEDIUM | P1 |
| MCP-over-STDIO serving | HIGH | MEDIUM | P1 |
| Layered context outputs | HIGH | HIGH | P1 |
| Health inspection and doctor diagnostics | MEDIUM | MEDIUM | P1 |
| Simple CLI lifecycle commands | HIGH | LOW | P1 |
| Portable context packs/export | MEDIUM | MEDIUM | P2 |
| Deterministic token-cost analysis | MEDIUM | MEDIUM | P2 |
| Optional watch mode | MEDIUM | MEDIUM | P2 |
| Optional semantic augmentation | LOW | HIGH | P3 |
| Hosted sync/service | LOW | HIGH | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | Competitor A | Competitor B | Our Approach |
|---------|--------------|--------------|--------------|
| Repository context access | General coding agents tend to rescan repos or rely on ad hoc file exploration per session | RAG-oriented code tools emphasize retrieval quality over deterministic repository state | Precompute and persist repository structure locally, then serve exact-first context on demand |
| Agent integration | Vendor-native tools are often optimized for one ecosystem and instruction format | IDE-centric tools often assume a specific editor/runtime environment | Use MCP-first integration with optional thin wrappers so the same runtime serves multiple clients |
| Freshness model | Some tools rely on broad rescans during task execution | Some systems depend on background indexing services or cloud-side processing | Support cheap incremental refresh with optional watch mode, but keep manual/local operation first-class |
| Shareability | Many tools rebuild understanding separately per client or machine | Hosted systems centralize state but trade off privacy and portability | Use portable context packs/export to move exact repository understanding without requiring a hosted service |

## Sources

- `.planning/PROJECT.md`
- `/home/nico/.codex/get-shit-done/templates/research-project/FEATURES.md`
- Stated project constraints and scope: local-first, MCP-first, deterministic structural context, non-invasive setup
- Comparative framing derived from current coding-agent tooling categories described in the project brief: vendor-native agents, ad hoc repo scanning, IDE-centric tooling, and RAG-oriented systems

---
*Feature research for: local-first context optimization runtime for coding agents*
*Researched: 2026-03-14*
