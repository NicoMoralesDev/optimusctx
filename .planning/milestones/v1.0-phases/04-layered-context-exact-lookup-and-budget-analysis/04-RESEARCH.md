# Phase 4 Research: Layered Context, Exact Lookup, and Budget Analysis

## Scope and Planning Intent

Phase 4 should expose the first exact retrieval surface built on the persisted repository facts from Phases 1 through 3. The planner should treat this as a query-and-composition phase, not another indexing phase. The main question is not how to extract more data. The main question is how to shape deterministic, bounded, agent-usable outputs from the data already stored.

Requirements covered in this phase:
- `CTX-01`
- `CTX-02`
- `CTX-03`
- `CTX-04`
- `CTX-05`
- `CTX-06`

The implementation should stay exact-first and local-first:
- answer from persisted metadata and symbols wherever possible
- read repository files only when exact code context is required
- return explicit freshness and coverage metadata
- keep ordering and bounds deterministic
- do not introduce semantic retrieval, fuzzy ranking, or an open-ended query DSL

## Current Repo Context

Phase 3 left the repo in a good position for Phase 4:
- `internal/app/repository_map.go` already proves the basic service shape: resolve repo root, open SQLite, load persisted read models, return deterministic output.
- `internal/store/sqlite/store.go` already exposes the core persisted facts Phase 4 needs: repository freshness, file inventory, directory aggregates, file extraction coverage, and symbol rows with exact spans.
- `internal/repository/metadata.go` already defines repository freshness, extraction, symbol, and repository-map domain types, so the next step is extending read-side types rather than inventing a separate data model.
- `internal/app/refresh.go` already exposes a reusable `ReadFile` seam, which is the obvious pattern for L2 context assembly because file contents are not persisted in SQLite.
- `internal/store/migrations/0003_structural_artifacts.sql` already added the most important query indexes for symbols by `name`, `qualified_name`, `path`, and ordinal ordering.

Important constraint: Phase 3 persists exact symbol spans, but it does not persist source text. That means `CTX-03` cannot be served purely from SQLite. The planner should make this explicit instead of pretending the runtime has a content cache.

Also important: there is no repo-local `CLAUDE.md`, `.claude/skills/`, or `.agents/skills/` guidance in this repository, so planning should rely on the root docs and current code only.

## What Must Be True After Phase 4

- The runtime can return a compact repository-level `L0` view from persisted metadata with freshness, language mix, and major-area summaries.
- The runtime can return an `L1` structural view from persisted artifacts with bounded file and symbol payloads.
- The runtime can return an `L2` targeted context block for a symbol or explicit line range with exact anchors and bounded surrounding lines.
- Exact symbol lookup can resolve by symbol name with optional path and language scoping using persisted symbol rows.
- Exact structural lookup can resolve through a narrow normalized query contract over persisted structure, not through reparsing or free-text search.
- Budget analysis can rank large files and directories deterministically from persisted size metadata and an explicit token-estimation policy.

## Standard Stack

- Go application services under `internal/app`
- SQLite read models under `internal/store/sqlite`
- Existing repository domain types under `internal/repository`
- Existing persisted artifacts from Phase 3 as the canonical lookup source
- Model-agnostic token estimation derived from persisted byte counts in v1

## Architecture Patterns

### 1. Query Services Mirror `RepositoryMapService`

Phase 4 should follow the same pattern as `internal/app/repository_map.go`:
- resolve repository root
- resolve state layout
- open SQLite
- load repository ID and freshness
- execute deterministic read queries
- shape the app-facing result

Recommended new services:
- `RepositoryContextService` for `L0` and `L1`
- `LookupService` for symbol and structure lookup
- `ContextBlockService` for `L2`
- `BudgetAnalysisService` for token-cost summaries and hotspots

Keep services transport-neutral. MCP-specific tool contracts belong to Phase 5.

### 2. Persisted Facts First, Live File Reads Only for L2

Use persisted SQLite facts for:
- repository identity and freshness
- file and directory ordering
- dominant language summaries
- candidate files and symbols
- symbol exact spans and qualified names
- directory and file size rollups

Use live file reads only for:
- extracting the exact code block around a symbol or line range
- computing the final bounded line window for `CTX-03`

This keeps the runtime truthful: metadata and lookup remain available from persisted state even if files are temporarily unavailable, while code-context assembly can fail explicitly when the working tree no longer matches the indexed generation.

### 3. Narrow, Typed Query Contracts

Do not plan a stringly typed mini-language in Phase 4. Use normalized request structs.

Recommended lookup contracts:
- symbol lookup: `name`, optional `path_prefix`, optional `language`, optional `kind`, `limit`
- structural lookup: `kind`, optional `parent_name`, optional `name`, optional `path_prefix`, optional `language`, `limit`
- context block: `path`, one of `symbol_stable_key` or `line_range`, plus `before_lines` and `after_lines`
- budget analysis: optional `path_prefix`, `limit`, `group_by`

`CTX-05` should be implemented as exact structural filtering over persisted symbol facts. It should not become pattern matching over raw source text.

### 4. Deterministic Ordering and Bounds Are Part of the Product

Phase 4 outputs should be sorted and bounded at the query layer, not left to callers.

Recommended ordering defaults:
- directories by path
- files by path
- symbols by path, then ordinal
- lookup matches by exactness, then path, then ordinal
- budget hotspots by estimated tokens descending, then path ascending

Recommended bounded defaults:
- `L1` file count limit
- per-file symbol cap
- lookup result limit
- `L2` max line budget and max byte budget
- budget hotspot limit

## Concrete Implementation Seams

### Domain Types: `internal/repository/metadata.go`

This file is the natural place for new result types:
- `LayeredContextL0`
- `LayeredContextL1`
- `TargetedContextBlock`
- `SymbolLookupMatch`
- `StructureLookupMatch`
- `BudgetHotspot`
- `BudgetAnalysis`

It already owns the Phase 3 query-facing types, so extending it preserves the existing layering.

### App Services: `internal/app`

Existing seam:
- `repository_map.go` is the template for new query services

Likely new files:
- `context.go`
- `lookup.go`
- `budget.go`

Possible reuse:
- share the locator, layout, and store-opening pattern already used by `RepositoryMapService`
- reuse the `ReadFile` injection style from `RefreshService` for `L2`

### SQLite Read Models: `internal/store/sqlite/store.go`

Phase 4 will likely need new read methods such as:
- `LoadL0Summary`
- `LoadL1Files`
- `LookupSymbols`
- `LookupStructures`
- `LoadSymbolByStableKey`
- `LoadBudgetHotspotsByFile`
- `LoadBudgetHotspotsByDirectory`

The most important current assets:
- `ReadRepositoryFreshness`
- `LoadRepositoryMapDirectories`
- `LoadRepositoryMapRecords`
- `ListSymbols`
- file snapshot queries with `size_bytes`, `language`, and path metadata

The store layer should own SQL filtering and ordering, not the app layer.

### Refresh / Extraction Reuse

Likely no extraction pipeline changes are required for baseline `CTX-01` through `CTX-06`.

However, the planner should evaluate whether Phase 4 needs:
- one extra SQLite migration for query indexes if lookup tests show poor query ergonomics
- a stored line-count column only if bounded `L2` or budget reporting becomes awkward without it

Current evidence suggests Phase 4 can start without schema changes because:
- symbol spans already exist
- file sizes already exist
- directory aggregate sizes already exist
- key lookup indexes already exist

## Likely Plan Slices

The state file already expects six Phase 4 plans. A strong slice order is:

### Slice 1. Query Domain Types and `L0` Repository Snapshot

Deliver:
- typed Phase 4 result structs
- persisted `L0` summary service
- dominant language and major-area summaries from existing file/directory metadata
- freshness metadata carried through the result

Why first:
- it establishes the output vocabulary for the rest of the phase
- it is low risk and requires no live file reads

### Slice 2. `L1` Structural Map and Bounded Candidate Views

Deliver:
- bounded structural view richer than Phase 3 repository map
- candidate files, top-level symbols, coverage flags, and relevance-limiting metadata
- deterministic limits and ordering

Why next:
- it reuses almost all Phase 3 persisted artifacts
- it clarifies what “layered context” means before exact lookup and L2 code blocks land

### Slice 3. Exact Symbol Lookup

Deliver:
- symbol lookup by name with optional `path_prefix`, `language`, and `kind`
- exact results carrying file path, kind, qualified name, and span anchors
- clear match-order rules

Why separate:
- `CTX-04` is a clean persisted-query problem and should be isolated from the harder `L2` file-reading behavior

### Slice 4. Exact Structural Lookup

Deliver:
- normalized structural query contract over persisted symbols
- narrow exact filters such as kind, parent, name, path scope, language
- explicit “unsupported query shape” failures rather than vague matching

Why separate:
- this is the requirement most likely to expand uncontrollably if not fenced

### Slice 5. Targeted `L2` Context Blocks

Deliver:
- context block assembly from symbol spans or explicit line ranges
- bounded surrounding lines
- exact file path and line anchors
- clear stale/missing-file behavior

Main dependency:
- exact lookup contracts should exist first so L2 can target by stable symbol identity rather than re-deriving matches

### Slice 6. Budget Analysis and Hotspot Ranking

Deliver:
- estimated token cost by file and directory
- ranked hotspots
- budget-aware shaping helpers usable by later MCP tools

Why last:
- it is largely orthogonal to lookup correctness
- it can reuse the same query-service scaffolding once Phase 4 result conventions are settled

## Requirement-by-Requirement Notes

### `CTX-01`: L0 repository snapshot

Good fit for persisted-only reads:
- repository root and freshness from `repositories`
- dominant languages from `files.language`
- major areas from top directories in `directories`
- size and included-file counts from `directories`

Planning note: define “major areas” deterministically, likely top-level included directories plus root-level file counts.

### `CTX-02`: L1 structural map

This is probably an evolution of the existing repository map, not a replacement.

Planner should decide whether to:
- extend `RepositoryMap`
- or create a distinct `L1` type that can later back MCP payloads directly

Recommendation: create a distinct L1 type. The repository map is already a compact artifact-oriented view; L1 needs explicit payload bounds and query-specific metadata.

### `CTX-03`: L2 targeted context block

This is the most important behavior boundary in the phase.

Because source text is not stored, the planner must define:
- whether L2 requires the file to exist on disk at request time
- how to behave when indexed spans refer to a file that changed after the last refresh
- whether freshness `stale` still allows best-effort reads or returns an actionable failure

Recommendation:
- use indexed symbol spans as the authoritative anchor
- read the current file contents from disk
- if the file is missing, return a targeted failure
- if the anchor range no longer fits the file cleanly, return a stale/drift failure that points the caller to refresh

Do not silently fall back to fuzzy text search for the symbol name.

### `CTX-04`: exact symbol lookup

This fits the current schema well:
- `symbols.name`
- `symbols.qualified_name`
- `symbols.path`
- `symbols.language`
- `symbols.kind`
- exact spans from row/column fields

Likely no migration required for baseline correctness.

### `CTX-05`: exact structural lookup

This does not need a full AST query engine in v1.

Recommendation:
- define a normalized structural query object over persisted symbol facts
- support exact filters over symbol kind, name, qualified name, path scope, language, and lexical parent
- treat unsupported structural intents as explicit validation failures

Do not plan:
- Tree-sitter query execution at request time
- raw source scanning
- CSS-selector-like structure syntax
- fuzzy relationship search

### `CTX-06`: budget analysis

This is a good fit for current persisted size data.

Recommendation:
- use a simple documented token estimate derived from file or directory byte counts
- rank files and directories separately
- expose both absolute estimated tokens and percent-of-total when useful

Do not pull in a model-specific tokenizer in this phase unless requirements tighten later. The requirement says estimate, not exact tokenize.

## Dependencies and Risks

### Main Dependencies

- Phase 1 repository identity and state layout
- Phase 2 freshness and generation model
- Phase 3 persisted extraction and symbol spans

### Main Risks

#### 1. L2 requires live file content but the index does not store content

This is the sharpest implementation seam in the phase. If not planned explicitly, L2 will drift into ad hoc behavior.

Mitigation:
- define L2 as persisted-anchor plus live-file read
- surface drift and missing-file failures explicitly
- test stale and deleted-file scenarios

#### 2. Structural lookup scope explosion

`CTX-05` can easily turn into “build a query language.”

Mitigation:
- keep the structural query contract narrow and typed
- support only exact filters over persisted symbol facts
- reject unsupported query combinations clearly

#### 3. Budget analysis becomes pseudo-precision

If the plan implies exact model-token accounting, the phase will get stuck on tokenizer policy.

Mitigation:
- document the estimator as model-agnostic
- keep the algorithm simple and deterministic
- use size-based ranking, not false precision

#### 4. Layered context duplicates repository-map behavior

If the planner does not separate repository map from L0/L1 semantics, Phase 4 may produce overlapping products with unclear callers.

Mitigation:
- define L0/L1 as agent-facing bounded context products
- leave repository map as the lower-level artifact summary introduced in Phase 3

## Don’t Hand-Roll

- Do not build a free-text search engine for symbol or structure lookup.
- Do not add semantic retrieval, embeddings, or fuzzy ranking.
- Do not introduce request-time Tree-sitter parsing for lookups that persisted symbols can already answer.
- Do not build model-specific tokenizer integrations in v1 unless a later requirement forces it.
- Do not silently compensate for stale or drifted files in L2 by guessing symbol boundaries from source text.

## Common Pitfalls

- Treating `CTX-05` as “support arbitrary structure patterns.”
- Returning oversized L1 or L2 payloads without deterministic caps.
- Mixing persisted metadata reads with ad hoc filesystem scans in L0/L1.
- Forgetting that Phase 3 symbols are exact only for indexed generations, not for arbitrary later working tree edits.
- Returning partial coverage files as if they had full structural fidelity.
- Ranking budget hotspots in a nondeterministic order when estimates tie.

## Recommended Validation Strategy

Phase 4 should stay heavily automated. The planner should assume store-level query tests plus app-level temp-repo integration tests, following the existing pattern from Phases 2 and 3.

Validation emphasis:
- deterministic ordering
- bound enforcement
- stale and drift behavior
- exact span anchoring
- persisted-only lookup behavior where appropriate

## Validation Architecture

### Test Infrastructure

- Framework: `go test`
- Main layers:
  - `internal/store/sqlite/*_test.go` for SQL filtering, ordering, limits, and edge cases
  - `internal/app/*_test.go` for end-to-end service behavior on temp repositories
- Existing refresh and extraction fixtures are sufficient for most setup; Phase 4 should add targeted helpers for symbol lookup and context-block assertions rather than new infrastructure

### Core Automated Test Groups

#### 1. L0 summary tests

Verify:
- repository root and freshness are present
- dominant languages are derived deterministically
- major areas are ordered deterministically
- persisted-only reads work after source files are deleted

Likely files:
- `internal/app/context_test.go`
- `internal/store/sqlite/context_test.go`

#### 2. L1 structural view tests

Verify:
- bounded file counts and per-file symbol caps
- deterministic directory/file/symbol ordering
- coverage-gap metadata for partial, failed, unsupported, and skipped files
- persisted-only behavior after worktree deletion

#### 3. Exact symbol lookup tests

Verify:
- exact name matches
- optional path and language scoping
- kind filtering
- qualified-name behavior where available
- deterministic tie ordering
- explicit empty result behavior

#### 4. Exact structural lookup tests

Verify:
- normalized structural query validation
- exact kind/name/parent/path filters
- unsupported query-shape failures
- no request-time parsing or filesystem dependency for persisted lookups

#### 5. L2 targeted context tests

Verify:
- symbol-targeted block extraction from exact spans
- line-range-targeted extraction
- bounded surrounding lines
- exact anchor metadata in the result
- missing-file failure
- stale/drift failure when file contents no longer match indexed anchors

These tests should use temp repositories and mutate files after refresh to prove the error contract.

#### 6. Budget analysis tests

Verify:
- deterministic token-estimate calculation
- file hotspot ranking
- directory hotspot ranking
- tie ordering
- path-prefix filtering
- percent-of-total or summary calculations if included

### Suggested Wave 0 Test Additions

- `internal/store/sqlite/context_test.go`
- `internal/store/sqlite/lookup_test.go`
- `internal/store/sqlite/budget_test.go`
- `internal/app/context_test.go`
- `internal/app/lookup_test.go`
- `internal/app/budget_test.go`

If the planner keeps Phase 4 in six slices, each slice should map to at least one focused automated group so Nyquist can sample continuously.

### High-Value Regression Cases

- lookup remains deterministic across repeated reads
- L1 bounds do not change output shape unexpectedly when repositories grow
- partial and unsupported files never present fabricated symbols
- L2 does not return code when the indexed anchor cannot be trusted
- budget analysis stays available even when source files are deleted, because it should rely on persisted sizes

## Planning Guidance

The planner should bias toward a no-migration Phase 4 unless profiling or query complexity proves otherwise. Most of the value is already unlocked by the current Phase 3 schema. The phase should therefore spend its budget on clear contracts, bounded output shaping, and exact failure semantics.

The best planning question is not “what more data do we need?” It is:

“Which Phase 4 behaviors can be answered from persisted artifacts alone, and where must the runtime explicitly cross into live file reads?”

That boundary should drive the final plan.
