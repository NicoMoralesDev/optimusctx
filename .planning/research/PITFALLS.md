# Pitfalls Research

**Domain:** Local-first context optimization runtime for coding agents
**Researched:** 2026-03-14
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: Silent index drift

**What goes wrong:**
The persisted repository index no longer matches the working tree, but the runtime keeps serving stale symbols, stale file metadata, or stale budget estimates as if they were current.

**Why it happens:**
Teams optimize for fast incremental refresh before they have a strong invalidation model. Rename handling, ignore rule changes, generated files, branch switches, and partial watcher failures create gaps that are hard to notice.

**How to avoid:**
Use a manifest-driven indexing model with explicit content hashes, index generation IDs, and per-file freshness state. Make refresh logic handle renames, deletes, ignore changes, and branch changes as first-class cases. Add a cheap repository-wide drift check and expose staleness in diagnostics.

**Warning signs:**
Agents report symbols that no longer exist, `doctor` passes while exact lookup fails, repeated "refresh complete" logs with unchanged stale results, or watch mode misses edits after branch checkout.

**Phase to address:**
Phase 1: repository discovery, hashing, and invalidation model.

---

### Pitfall 2: Non-deterministic context output

**What goes wrong:**
The same repository state produces different summaries, symbol orderings, pack contents, or MCP responses across runs, machines, or operating systems. That makes agent behavior hard to trust and hard to debug.

**Why it happens:**
Filesystem traversal order, concurrent extraction, map iteration, SQLite query ordering, and platform-specific path normalization are left implicit.

**How to avoid:**
Define deterministic ordering rules for traversal, storage, and response rendering. Normalize paths, line endings, and timestamps. Require explicit `ORDER BY` clauses and stable serialization tests across repeated runs.

**Warning signs:**
Snapshot tests flap, exported packs differ between runs with no code changes, or users report "same repo, different answer" across Linux and macOS.

**Phase to address:**
Phase 2: deterministic storage contracts and stable output rendering.

---

### Pitfall 3: Over-collecting instead of optimizing

**What goes wrong:**
The runtime becomes a generic code search or RAG system that stores too much text, spends too many tokens, and loses the exact-first product wedge.

**Why it happens:**
It is tempting to compensate for weak structure extraction by stuffing more raw content into the index. This feels useful early but blurs the product boundary and increases cost.

**How to avoid:**
Set strict v1 boundaries around structural extraction, exact lookup, and layered outputs. Track token cost per output mode. Require every stored artifact to justify its retrieval value and determinism characteristics.

**Warning signs:**
Index size grows faster than repository size, outputs start resembling full-file dumps, or feature requests are addressed by adding fuzzier retrieval instead of better structure.

**Phase to address:**
Phase 3: layered context outputs and budget enforcement.

---

### Pitfall 4: Watch mode becomes the source of truth

**What goes wrong:**
The product only feels correct when a long-running watcher is active. One-shot refresh, CI verification, and cold-start initialization become second-class or broken.

**Why it happens:**
File watching is convenient during development, so teams accidentally hide missing rescan logic behind watcher events. Cross-platform watcher edge cases then become correctness issues.

**How to avoid:**
Design refresh correctness independently from watch mode. Treat watch mode as an optimization that schedules the same core refresh pipeline used by manual runs. Verify full rebuild, targeted refresh, and watcher-triggered refresh against the same assertions.

**Warning signs:**
Fresh clones need a watcher restart to become accurate, bug reports cannot be reproduced with manual refresh, or platform-specific watcher fixes keep appearing in unrelated phases.

**Phase to address:**
Phase 4: optional watch mode built on top of a complete refresh engine.

---

### Pitfall 5: SQLite schema ossifies too early

**What goes wrong:**
The local store works for the first few features, then every new artifact or query path requires painful migrations, slow backfills, or breaking pack/export changes.

**Why it happens:**
Early schemas are often modeled around current implementation details rather than durable domain objects such as repository snapshot, file artifact, symbol artifact, and export contract.

**How to avoid:**
Define durable storage boundaries before adding many artifact types. Add schema versioning, migration tests, and fixtures from older database versions. Keep derived tables rebuildable where possible.

**Warning signs:**
New fields are duplicated in multiple tables, migrations need custom one-off repair logic, or export/import only works for the latest build.

**Phase to address:**
Phase 2: persistent store design, migrations, and rebuild strategy.

---

### Pitfall 6: Tree-sitter coverage is treated as binary

**What goes wrong:**
Users assume structural extraction is complete, but unsupported languages, parser failures, injected languages, or malformed files silently produce partial context.

**Why it happens:**
Structural tooling is presented as "language supported" or "not supported" instead of surfacing partial extraction quality, fallback behavior, and parser health.

**How to avoid:**
Track extraction capability per file and per language with explicit confidence states. Surface parser errors and fallback modes in MCP and diagnostics. Design layered outputs to degrade gracefully when structure is incomplete.

**Warning signs:**
Repositories with mixed languages behave much worse than mono-language repos, symbol counts drop after parser upgrades, or users cannot tell whether a file was skipped, partially parsed, or fully indexed.

**Phase to address:**
Phase 5: structural extraction quality, parser diagnostics, and graceful degradation.

---

### Pitfall 7: MCP contract leaks internal instability

**What goes wrong:**
Clients integrate with tool names, payload shapes, or error messages that change as the runtime evolves, breaking portability and forcing prompt-level workarounds.

**Why it happens:**
The protocol is treated as an implementation detail instead of a product surface. Internal refactors then reshape request and response formats without compatibility discipline.

**How to avoid:**
Version the MCP surface, keep responses narrow and explicit, document error semantics, and add compatibility tests with golden transcripts. Prefer additive evolution over in-place meaning changes.

**Warning signs:**
Agent prompts need repository-specific tool instructions, integration examples stop working after minor releases, or the same failure returns inconsistent error shapes.

**Phase to address:**
Phase 6: MCP interface definition, compatibility tests, and error contracts.

---

### Pitfall 8: Privacy promises break during export and diagnostics

**What goes wrong:**
A "local-first" tool accidentally leaks repository contents through debug bundles, exported packs, verbose logs, or copied caches.

**Why it happens:**
Teams focus on avoiding network calls, but forget that local artifacts can still be redistributed, attached to tickets, or shared between machines without clear redaction boundaries.

**How to avoid:**
Classify all persisted and exported artifacts by sensitivity. Make export opt-in and explicit about included content. Provide redacted diagnostics modes and log scrubbing by default.

**Warning signs:**
Logs contain code snippets or secrets, support workflows require users to share raw databases, or pack/export semantics are described vaguely in docs.

**Phase to address:**
Phase 7: pack/export workflows, diagnostics, and privacy controls.

---

### Pitfall 9: Exact lookup is not actually exact

**What goes wrong:**
An agent asks for a path, symbol, or artifact and receives ambiguous, fuzzy, or best-effort results. This undermines trust and pushes users back to brute-force file reads.

**Why it happens:**
Developers optimize for "helpful" responses and conflate convenience search with the exact-first guarantees the product depends on.

**How to avoid:**
Separate exact lookup from discovery endpoints. Return explicit ambiguity errors with candidate lists rather than silently choosing one match. Add fixtures for duplicates, case sensitivity, and namespace collisions.

**Warning signs:**
Repeated "closest match" logic shows up in exact APIs, tests only cover unique symbol names, or users report incorrect jumps in monorepos and generated-code trees.

**Phase to address:**
Phase 3: exact lookup semantics and ambiguity handling.

---

### Pitfall 10: Monorepo scale is treated like a bigger small repo

**What goes wrong:**
The runtime appears fast on small repos but becomes unusable on monorepos because refresh cost, DB growth, and response assembly scale superlinearly.

**Why it happens:**
Early benchmarks use toy repositories. Monorepos add deep directory trees, generated assets, duplicated vendored code, and many loosely related packages that stress every stage differently.

**How to avoid:**
Benchmark against representative large repos early. Add repository partitioning, selective indexing, backpressure on watch events, and bounded output assembly. Make health inspection reveal the dominant cost centers.

**Warning signs:**
Refresh time grows faster than changed-file count, database vacuuming becomes operationally necessary, or watch mode thrashes during branch rebases or code generation.

**Phase to address:**
Phase 8: large-repository performance, health inspection, and selective indexing controls.

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Rebuild the full index on every refresh | Simplifies correctness early | Masks invalidation bugs and makes watch mode and monorepos expensive | Only for the very first spike before incremental design exists |
| Store rendered outputs instead of source artifacts | Faster demos and easier MCP responses | Cached outputs drift, storage bloats, and migrations become brittle | Acceptable only for ephemeral test fixtures |
| Treat unsupported parsers as silent no-ops | Keeps indexing pipeline simple | Users trust incomplete context and debugging becomes opaque | Never |
| Expose internal database IDs through MCP | Fast to implement | Locks protocol to storage internals and breaks compatibility | Never |
| Put all diagnostics behind verbose logs | Minimal command surface | Hard to automate, hard to redact, and noisy for agents | Only before a dedicated `doctor` command exists |
| Skip migration fixtures for local DB changes | Faster iteration on schema | Old indexes become unrecoverable and upgrades lose trust | Only in throwaway prototypes before any release |

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| MCP over STDIO | Mixing user-facing logs with protocol output | Keep protocol output clean on stdout and send diagnostics to stderr or structured log files |
| Agent instruction snippets | Assuming the snippet will always be installed | Ensure the runtime remains discoverable and useful without instruction-file changes |
| Tree-sitter grammars | Treating parser installation/versioning as implicit | Pin grammar versions, surface availability, and test parser upgrades on fixture repos |
| Git-aware refresh | Assuming file events are enough after checkout or rebase | Detect branch and HEAD changes explicitly and trigger repository-level reconciliation |
| Pack/export between machines | Assuming absolute paths remain valid | Use portable identifiers and rebuild machine-local path bindings on import |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Per-request filesystem scanning | MCP calls become slower over time and agent loops repeat disk I/O | Serve from persisted artifacts and keep refresh separate from retrieval | Breaks around medium repos with repeated interactive calls |
| N+1 symbol queries against SQLite | Exact lookup feels fine in tests but layered outputs stall on large repos | Precompute query shapes and batch artifact retrieval | Breaks in monorepos with tens of thousands of symbols |
| Re-rendering context layers on every identical request | CPU spikes on repeated agent prompts | Cache deterministic render results by snapshot and budget parameters | Breaks once agents poll the same tools in loops |
| Unbounded watch event queues | Memory grows and refresh latency lags behind edits | Coalesce events, apply backpressure, and fall back to targeted rescans | Breaks during branch switches, rebases, or mass code generation |
| Indexing vendored and generated trees by default | Huge database growth and slow cold starts | Respect ignore files, detect common generated paths, and offer explicit opt-in | Breaks early in JS, mobile, and polyglot monorepos |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Logging raw file excerpts in diagnostics | Secrets or proprietary code leak through local logs and support bundles | Default to metadata-only diagnostics and require explicit opt-in for content capture |
| Exporting packs without content classification | Sensitive repository artifacts get shared under the assumption that "local-first" means safe | Label pack contents clearly and support redacted export modes |
| Trusting symlink traversal during indexing | Index escapes repository boundaries and ingests unintended files | Enforce repository root boundaries and make symlink policy explicit |
| Executing parser or helper tooling from the indexed repo | Malicious repositories can influence runtime behavior | Use bundled or pinned tooling and treat repository contents as untrusted input |
| Treating MCP clients as trusted requesters | A prompt-influenced client can request excessive or sensitive output | Enforce output budgets, exact scopes, and explicit tool semantics regardless of caller |

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Hiding freshness state | Users cannot tell whether context is current | Show snapshot time, staleness, and last refresh cause in every health surface |
| Returning giant "helpful" payloads | Agents waste budget and humans lose trust in relevance | Offer layered outputs with explicit size and purpose |
| Ambiguous initialization flow | Users do not know whether the repo is indexed, watched, or healthy | Provide a simple init -> refresh -> doctor progression with clear status |
| Conflating search with exact lookup | Users get surprising results when they need precision | Separate tool names and behaviors for discovery vs exact retrieval |
| Making watch mode feel mandatory | Users avoid the tool when background processes are undesirable | Keep one-shot workflows first-class and clearly supported |

## "Looks Done But Isn't" Checklist

- [ ] **Incremental refresh:** Often missing rename, delete, and ignore-rule handling — verify index state matches a branch switch and mass rename fixture.
- [ ] **Exact lookup:** Often missing ambiguity handling — verify duplicate symbol names return an explicit ambiguity response rather than one arbitrary result.
- [ ] **Watch mode:** Often missing event coalescing and overflow recovery — verify correctness after large rebases or generated file bursts.
- [ ] **SQLite migrations:** Often missing upgrade coverage from prior releases — verify old fixture databases migrate and remain queryable.
- [ ] **MCP interface:** Often missing compatibility guarantees — verify golden transcript tests across repeated runs and minor versions.
- [ ] **Pack/export:** Often missing privacy review — verify what content leaves the machine and whether redaction modes work.
- [ ] **Diagnostics:** Often missing actionable drift reporting — verify `doctor` distinguishes stale index, parser failure, and unsupported language cases.
- [ ] **Cross-platform paths:** Often missing normalization and case-sensitivity checks — verify behavior on Linux and macOS fixtures.

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Silent index drift | MEDIUM | Detect drift, mark affected artifacts stale, trigger targeted or full rebuild, and expose the cause in diagnostics |
| Non-deterministic context output | MEDIUM | Capture divergent snapshots, compare ordering and normalization paths, then lock traversal and serialization rules with regression tests |
| SQLite schema ossification | HIGH | Introduce versioned migrations, add rebuildable derived artifacts, ship migration fixtures, and provide a safe reindex path |
| MCP contract instability | HIGH | Freeze a versioned contract, add compatibility adapters where needed, and publish deprecation windows before changing semantics |
| Privacy leakage in export/diagnostics | HIGH | Revoke shared artifacts where possible, scrub local logs, tighten defaults, and add redaction plus content inventory before re-enabling export |
| Monorepo performance collapse | HIGH | Profile end-to-end costs, add selective indexing and batching, backfill benchmarks, and ship a temporary scoped-index workaround |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Silent index drift | Phase 1: repository discovery, hashing, and invalidation model | Branch switch, rename, delete, and ignore-change fixtures preserve correct freshness state |
| Non-deterministic context output | Phase 2: deterministic storage contracts and stable output rendering | Repeated runs on the same fixture produce byte-stable outputs |
| Over-collecting instead of optimizing | Phase 3: layered context outputs and budget enforcement | Output size stays within budget and artifacts remain structural and exact-first |
| Watch mode becomes the source of truth | Phase 4: optional watch mode built on top of a complete refresh engine | Manual refresh and watcher-triggered refresh converge on the same index state |
| SQLite schema ossifies too early | Phase 2: persistent store design, migrations, and rebuild strategy | Old database fixtures migrate successfully and derived artifacts can be rebuilt |
| Tree-sitter coverage is treated as binary | Phase 5: structural extraction quality and diagnostics | Unsupported and partially parsed files are surfaced with explicit capability states |
| MCP contract leaks internal instability | Phase 6: MCP interface definition and compatibility tests | Golden transcript suite passes across minor releases |
| Privacy promises break during export and diagnostics | Phase 7: pack/export workflows, diagnostics, and privacy controls | Export inventory and redaction tests confirm no unintended content leaves the machine |
| Exact lookup is not actually exact | Phase 3: exact lookup semantics and ambiguity handling | Duplicate names and path collisions return deterministic exact or explicit ambiguity results |
| Monorepo scale is treated like a bigger small repo | Phase 8: large-repository performance and selective indexing | Benchmarks on representative large repos stay within refresh and query targets |

## Sources

- `.planning/PROJECT.md` product scope and constraints for OptimusCtx
- Known failure patterns from local indexing systems, code intelligence tools, and agent tooling integrations
- Common SQLite migration and watch-service failure modes in local developer tools
- Tree-sitter parser integration gotchas from multi-language repository tooling
- Practical experience with deterministic build/index pipelines and MCP-style protocol surfaces

---
*Pitfalls research for: local-first context optimization runtime for coding agents*
*Researched: 2026-03-14*
