# Architecture Research

**Domain:** Local-first context optimization runtime for coding agents
**Researched:** 2026-03-14
**Confidence:** HIGH

## Standard Architecture

### System Overview

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Operator Surfaces                               │
├─────────────────────────────────────────────────────────────────────────────┤
│  CLI commands        MCP server         Watch mode          Pack/export     │
│  init/index/doctor   exact context      optional refresh    portable bundle │
└──────────────┬──────────────────┬──────────────────┬───────────────────────┘
               │                  │                  │
┌──────────────┴──────────────────┴──────────────────┴───────────────────────┐
│                            Application Layer                               │
├─────────────────────────────────────────────────────────────────────────────┤
│  Command handlers   MCP handlers   Health/diagnostics   Budget formatting   │
│  Job orchestration  Query routing  Error reporting      Response shaping    │
└──────────────┬───────────────────────────────┬──────────────────────────────┘
               │                               │
┌──────────────┴───────────────────────────────┴──────────────────────────────┐
│                               Core Runtime                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│  Repository scanner  Change detector  Extractors  Symbol/context builder    │
│  Ignore rules        Hash planner     Tree-sitter Layer composer            │
│  File metadata       Incremental DAG  Language adapters Budget calculator   │
└──────────────┬───────────────────────────────┬──────────────────────────────┘
               │                               │
┌──────────────┴───────────────────────────────┴──────────────────────────────┐
│                              Persistence Layer                              │
├─────────────────────────────────────────────────────────────────────────────┤
│  SQLite catalog      Blob/cache store      Export packer      Migration set │
│  Repo state          Derived artifacts      Portable archives  Schema guard  │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| CLI surface | Human-facing commands for init, index, doctor, serve, pack | Go `cobra`-style command package or stdlib command tree |
| MCP adapter | Stable agent-facing API over STDIO with exact-first methods | Thin protocol layer that maps requests to query services |
| Repository scanner | Walk files deterministically with ignore awareness | Filesystem traversal plus `.gitignore`/tool ignore matcher |
| Change detector | Identify stale files cheaply and schedule recomputation | Content hash + mtime/size short-circuit + dependency invalidation |
| Extractor pipeline | Produce structural artifacts from source files | Tree-sitter parsers behind language-specific adapters |
| Context composer | Build layered outputs for summaries, symbols, exact lookup, packs | Pure services over persisted artifacts and budget rules |
| Storage catalog | Persist repository metadata, artifacts, health, and migrations | SQLite with normalized tables and explicit schema versions |
| Watch coordinator | Optional background refresh loop for changed files | Debounced file watcher feeding the same incremental pipeline |
| Doctor/health service | Inspect corruption, parser gaps, stale state, and config problems | Read-only checks with clear remediation actions |

## Recommended Project Structure

```text
cmd/
├── optimusctx/              # main entrypoint wiring commands and process startup

internal/
├── app/                     # command handlers and MCP request orchestration
│   ├── cli/                 # init/index/doctor/serve/pack subcommands
│   ├── mcp/                 # protocol transport, method registration, DTOs
│   └── jobs/                # shared job runner, progress, cancellation
├── core/                    # domain logic with no transport/storage assumptions
│   ├── repo/                # repository model, ignore rules, traversal contracts
│   ├── detect/              # change detection, invalidation, dependency planning
│   ├── extract/             # parser adapters, symbol extraction, file analyzers
│   ├── compose/             # layered context outputs and budget-aware formatting
│   └── health/              # consistency checks and diagnostics rules
├── store/                   # persistence implementations
│   ├── sqlite/              # schema, queries, migrations, transactions
│   ├── cache/               # optional artifact/blob cache
│   └── pack/                # import/export archive format
├── platform/                # OS-facing adapters
│   ├── fs/                  # file IO, path normalization, watcher abstraction
│   └── log/                 # structured logging and tracing hooks
└── testutil/                # fixtures, fake repositories, parser stubs

schemas/                     # MCP schemas, pack manifest examples, sample configs
testdata/                    # snapshot fixtures and representative repositories
docs/                        # protocol notes and operator docs
```

### Structure Rationale

- **`cmd/`:** Keeps the binary entrypoint thin and avoids leaking startup concerns into domain code.
- **`internal/app/`:** Owns transport-specific orchestration so CLI and MCP can share the same services without coupling.
- **`internal/core/`:** Preserves deterministic business rules in a testable layer independent of SQLite, STDIO, or file watchers.
- **`internal/store/`:** Contains all persistence details, making schema changes and pack format evolution explicit.
- **`internal/platform/`:** Isolates OS and filesystem behavior that tends to vary across platforms.
- **`testdata/` and `internal/testutil/`:** Matter early because incremental indexing and context rendering need snapshot-heavy tests from the start.

## Architectural Patterns

### Pattern 1: Hexagonal Core With Thin Adapters

**What:** Keep repository analysis and context composition in a pure core, with CLI, MCP, SQLite, and watcher logic as adapters around it.
**When to use:** From the first implementation; this project has multiple operator surfaces but one domain.
**Trade-offs:** Slightly more upfront interface design, but it prevents transport code from contaminating indexing logic.

**Example:**
```go
type ContextQueryService interface {
	GetFileContext(ctx context.Context, repoID string, path string, budget Tokens) (LayeredContext, error)
}

type MCPServer struct {
	query ContextQueryService
}
```

### Pattern 2: Incremental Materialized Artifacts

**What:** Persist normalized repository facts and derived artifacts so queries read precomputed state instead of re-scanning files.
**When to use:** For symbols, file summaries, structural maps, dependency edges, and budget metadata.
**Trade-offs:** Requires invalidation discipline and migrations, but keeps query latency predictable and cheap.

**Example:**
```go
type FileRecord struct {
	Path        string
	ContentHash string
	ParseHash   string
	Language    string
}

type ArtifactWriter interface {
	UpsertFile(tx Tx, file FileRecord) error
	ReplaceSymbols(tx Tx, path string, symbols []Symbol) error
}
```

### Pattern 3: Single Pipeline, Multiple Triggers

**What:** Run the same indexing pipeline whether invoked by `index`, `watch`, import, or doctor repair.
**When to use:** Always; change sources differ, but recomputation rules should not.
**Trade-offs:** The pipeline contract must be explicit, but the payoff is consistent behavior and fewer edge cases.

**Example:**
```go
type RefreshRequest struct {
	RepoRoot      string
	ChangedPaths  []string
	Reason        string // manual, watch, import, repair
	FullRescan    bool
}

func (s *Indexer) Refresh(ctx context.Context, req RefreshRequest) (RefreshResult, error)
```

### Pattern 4: Exact-First Layered Query Responses

**What:** Return structured layers such as repository summary, file metadata, symbols, and exact snippets instead of fuzzy blended answers.
**When to use:** In every MCP method and pack/export path in v1.
**Trade-offs:** Less magical than semantic retrieval, but aligned with determinism, cost control, and agent portability.

## Data Flow

### Request Flow

```text
[CLI command or MCP request]
    ↓
[Handler validates input and budgets]
    ↓
[Query or refresh service]
    ↓
[SQLite catalog + artifact store]
    ↓
[Layer composer formats deterministic response]
    ↓
[CLI output / MCP response / pack archive]
```

### State Management

```text
[Filesystem state]
    ↓ scan/watch
[Change detector]
    ↓ invalidation plan
[Extractor pipeline]
    ↓ derived artifacts
[SQLite catalog]
    ↓ query
[Context composer]
```

### Key Data Flows

1. **Repository initialization:** `init` records repo identity, effective ignore rules, schema version, and baseline scan state before any extraction work.
2. **Incremental refresh:** Scanner discovers candidate changes, detector computes stale units, extractors rebuild only affected artifacts, and the store swaps them transactionally.
3. **Exact context lookup:** MCP request resolves repository, fetches persisted facts and snippets, applies token budget shaping, and returns layered structured data.
4. **Health inspection:** Doctor walks metadata, parser coverage, and artifact freshness to report corruption, stale indexes, or unsupported-language gaps.
5. **Pack/export:** Selected repository artifacts are serialized into a portable archive with manifest, schema version, and integrity metadata for transfer between machines.

## Build Order

1. **Repository identity and storage foundation:** schema, migrations, repo registration, config loading, ignore resolution, and deterministic filesystem traversal.
2. **Incremental refresh engine:** file inventory, hashing, stale detection, transactional upsert flow, and snapshot tests for unchanged vs changed repositories.
3. **Structural extraction:** Tree-sitter integration, language adapter contract, symbol tables, and exact snippet metadata.
4. **Context composition and budgets:** repository summary, file summary, symbol lookup, layered output structs, and budget-aware formatting rules.
5. **CLI command surface:** `init`, `index`, `status`, `doctor`, and `pack` built on the same services.
6. **MCP server adapter:** STDIO transport, method registration, request validation, and deterministic response schemas.
7. **Optional watch mode:** debounced watcher trigger feeding the existing refresh engine, with platform abstraction and failure recovery.
8. **Hardening:** corruption recovery, import/export verification, large-repo performance tuning, and compatibility tests across sample repositories.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| Single developer, small repo | Single process, single SQLite file, synchronous refresh jobs are sufficient |
| Large monorepo on one machine | Add chunked traversal, bounded worker pools, parser concurrency limits, and artifact compaction |
| Team sharing portable packs | Keep runtime local-first, but harden manifest/version checks and make export/import incremental |

### Scaling Priorities

1. **First bottleneck:** File traversal and reparsing on large repos. Fix with aggressive stale detection, parser caching, and work scheduling before introducing more complexity.
2. **Second bottleneck:** SQLite write contention during watch bursts. Fix with batched transactions, debounce windows, and append-friendly artifact replacement patterns.

## Anti-Patterns

### Anti-Pattern 1: Mixing Query-Time Analysis With Serve Logic

**What people do:** Re-read files or re-run parsers inside MCP handlers to answer requests.
**Why it's wrong:** It destroys determinism, makes latency depend on repo size, and creates drift between CLI and MCP behavior.
**Do this instead:** Keep handlers thin and answer from persisted artifacts, with explicit refresh commands or watch-driven updates.

### Anti-Pattern 2: Treating the Index as an Opaque Blob

**What people do:** Store one large serialized context payload per repository.
**Why it's wrong:** Small changes force full rewrites, diagnostics are poor, and exact lookup becomes awkward.
**Do this instead:** Persist normalized file facts plus derived artifacts with clear ownership and invalidation rules.

### Anti-Pattern 3: Letting Transport Semantics Leak Into the Core

**What people do:** Shape domain models around MCP method names or CLI output strings.
**Why it's wrong:** It locks the architecture to one surface and makes testing and reuse harder.
**Do this instead:** Define transport-neutral services and map them to CLI/MCP DTOs at the edge.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Model Context Protocol clients | STDIO MCP server | Primary agent integration surface; keep method set intentionally small |
| Tree-sitter grammars | Embedded or linked parser adapters | Version pinning matters because parse output stability affects artifact determinism |
| Local filesystem | Direct OS access through adapter layer | Must normalize paths, symlinks, and ignore semantics consistently across platforms |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| `app` ↔ `core` | Service interfaces and DTO mapping | `app` orchestrates; `core` owns rules |
| `core` ↔ `store` | Repository and artifact interfaces | Keep SQL details out of domain logic |
| `core.extract` ↔ language adapters | Parser contract | Unsupported languages should degrade cleanly, not fail the whole refresh |
| `platform.fs` ↔ refresh engine | File and watch events | Watch mode must be optional and reuse the same refresh request type |

## Sources

- `.planning/PROJECT.md`
- MCP-first and local-first product constraints from project brief
- Recommended implementation choices already declared in project context: Go, SQLite, Tree-sitter
- Common architecture practice for deterministic local indexing runtimes and adapter-based CLIs

---
*Architecture research for: OptimusCtx*
*Researched: 2026-03-14*
