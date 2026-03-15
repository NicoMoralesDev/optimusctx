# Phase 5 Research: MCP Serving and Integration Contracts

## Scope and Planning Intent

Phase 5 should turn the existing local runtime into a stable MCP-first surface without reopening the core indexing or query semantics already established in Phases 1 through 4. The planner should treat this as a transport-and-contract phase first, with only the minimum new domain work needed to expose the promised MCP capabilities truthfully.

Requirements covered in this phase:
- `CLI-02`
- `MCP-01`
- `MCP-02`
- `MCP-03`
- `MCP-04`

The implementation should preserve the existing product guardrails:
- transport wrappers should sit on top of stable app services instead of bypassing them
- MCP responses should stay machine-readable, deterministic, and bounded by default
- freshness and cache-versus-refresh status should remain explicit in every relevant tool result
- install-time client registration must be opt-in only and should never silently rewrite config files
- Phase 5 should not absorb full watch-mode, doctor hardening, or broad pack/export ergonomics that belong to Phase 6

## Current Repo Context

The repo is in a strong position for Phase 5 because the underlying read and refresh surfaces already exist:
- `internal/app/repository_map.go` exposes a persisted-only repository map service.
- `internal/app/context.go` exposes L0 and L1 layered context services.
- `internal/app/lookup.go` exposes exact symbol and structure lookup services.
- `internal/app/context_block.go` exposes exact L2 context assembly.
- `internal/app/refresh.go` exposes the manual refresh service with deterministic refresh/freshness behavior.
- `internal/app/budget.go` exposes budget hotspot analysis built from persisted size metadata.
- `internal/app/snippet.go` already points at a future `optimusctx mcp serve` entrypoint, which is the natural Phase 5 CLI seam.
- `internal/cli/root.go` still only exposes `init`, `refresh`, `snippet`, and `version`, so the CLI transport surface remains intentionally small and easy to extend.

Important gaps that Phase 5 must plan around:
- there is no MCP server package, STDIO loop, tool registry, or protocol contract yet
- there are no app services yet for `token tree`, `pack`, or `health`
- there is no install-time registration workflow for supported clients
- there is no project-local `CLAUDE.md`, `.claude/skills/`, or `.agents/skills/` guidance in this repository, so planning should rely on the repo itself plus GSD workflow rules

## What Must Be True After Phase 5

- The binary can serve MCP over STDIO through a stable `optimusctx mcp serve` entrypoint.
- MCP tools expose machine-readable payloads for repository map, exact lookup, context, refresh, token tree, pack, and health flows.
- The transport carries freshness and cache-versus-refresh status consistently instead of forcing callers to infer index state from ad hoc fields.
- Tool defaults are bounded, oversized requests fail transparently, and handlers do not silently widen scope.
- Install flow can render and optionally register client MCP configuration only after explicit user consent.

## Standard Stack

- Go CLI under `internal/cli`
- Transport-neutral app services under `internal/app`
- Repository domain types under `internal/repository`
- SQLite read models under `internal/store/sqlite`
- MCP server and tool adapter code likely under a new `internal/mcp` package

## Architecture Patterns

### 1. Keep MCP as a Thin Transport Layer

Phase 5 should not duplicate business logic inside protocol handlers. The correct layering is:
- CLI command parses flags and starts STDIO server
- MCP server validates tool input and normalizes bounds
- tool handlers call transport-neutral app services
- app services resolve repo root, state layout, and store access
- result adapters shape app outputs into MCP response payloads

This preserves the exact-first architecture already established in Phase 4 and keeps Phase 6 free to add new operator entrypoints without re-implementing core behavior.

### 2. Add Shared Request/Response Envelopes Early

MCP-02 requires consistent machine-readable responses with freshness and cache-versus-refresh status metadata. That means the planner should introduce shared protocol result shapes before wiring every tool individually.

Recommended shared concepts:
- repository root
- current generation
- freshness status
- cache status such as `persisted_only`, `live_file_backed`, or `refresh_attempted`
- limit metadata for truncated or bounded results
- structured error payloads that tell callers exactly which bound or validation rule failed

The transport should not rely on free-form strings as the primary payload.

### 3. Token Tree, Pack, and Health Need Narrow v1 Semantics

Phase 5 requires MCP capabilities for `token tree`, `pack`, and `health`, but the broader CLI/operator workflows for pack export and doctor diagnostics are explicitly Phase 6 work. The planner should keep these Phase 5 versions narrow:
- `token tree`: a machine-readable hierarchical cost view derived from persisted directory/file size metadata and the existing bytes-to-token policy
- `pack`: a bounded, deterministic MCP-facing context bundle assembled from existing repository/context/lookup services, not a full export workflow or on-disk artifact pipeline
- `health`: a structured runtime-state and freshness diagnostic response for MCP callers, not the full operator-facing `doctor` command

This is the cleanest way to satisfy `MCP-03` without stealing the scope of `OPS-03` through `OPS-05`.

### 4. Explicit Bounds Must Live at the Transport Boundary

Phase 4 already enforced deterministic limits in core queries. Phase 5 should add transport-facing request validation for:
- maximum candidate counts
- maximum symbol or structure result counts
- maximum code-window line counts
- maximum token-tree depth or node count
- maximum pack size or section count

Handlers should fail with actionable errors when callers exceed allowed bounds. Silent truncation without metadata would violate `MCP-04`.

### 5. Registration Must Be Adapter-Based and Consent-Gated

`CLI-02` is not just a new command. It needs a stable representation of supported client config plus explicit write behavior. A strong v1 pattern is:
- detect supported client adapters behind a small registry
- render the exact config change before writing
- require explicit opt-in flag or interactive confirmation
- support dry-run or stdout-only rendering for manual application
- never write if target config path is unsupported, ambiguous, or already in a conflicting state without a clear error

This keeps the product non-invasive while still delivering a practical install wedge.

## Concrete Implementation Seams

### CLI

Current root command switch in `internal/cli/root.go` is simple and deterministic. Phase 5 should likely add:
- `optimusctx mcp serve`
- an install or register command for supported client config

The snippet output in `internal/app/snippet.go` should be reconciled with the real registration and serve entrypoint once those exist.

### App Services

Existing app services are already shaped for MCP exposure:
- `RepositoryMapService`
- `RepositoryContextService`
- `LookupService`
- `ContextBlockService`
- `RefreshService`
- `BudgetAnalysisService`

Likely new app services for Phase 5:
- `TokenTreeService`
- `HealthService`
- `PackService`

These should stay transport-neutral so both MCP and future Phase 6 CLI surfaces can reuse them.

### Repository / Domain Types

Existing domain types in `internal/repository` are already the right home for new transport-neutral request/result models for:
- token-tree queries and nodes
- machine-readable health summaries
- bounded pack requests and bundle summaries
- shared metadata that MCP adapters can reuse without leaking protocol structs into app code

### MCP Transport Package

Introduce a new `internal/mcp` package that owns:
- server bootstrap over STDIO
- tool registration
- input normalization and validation
- response adapters
- structured error mapping

Avoid mixing protocol code into `internal/cli` or `internal/app`.

## Likely Plan Slices

The current state file expects six plans in this phase. A strong slice order is:

### Slice 1. MCP Server Foundation and CLI Entry Point

Deliver:
- `optimusctx mcp serve`
- STDIO session bootstrap
- tool registry skeleton
- shared protocol envelope and error mapping primitives

Why first:
- every later MCP tool depends on the transport seam existing
- `CLI-02` and snippet alignment also depend on the final serve command path

### Slice 2. Read-Only MCP Query Tools Over Existing Services

Deliver:
- repository map
- L0/L1 context
- exact symbol lookup
- exact structure lookup
- targeted context block

Why early:
- these are the lowest-risk tools because the underlying app services already exist
- they prove the transport can carry freshness, cache metadata, and bounded defaults correctly

### Slice 3. Token Tree Service and MCP Tool

Deliver:
- transport-neutral token-tree request/result types
- app/service and persisted query logic built from existing budget metadata
- MCP tool surface for hierarchical cost inspection

Why separate:
- `token tree` is required by `MCP-03` but does not exist yet
- it is conceptually adjacent to Phase 4 budget work and should stay isolated from refresh or install changes

### Slice 4. Pack and Health Core Services

Deliver:
- narrow machine-readable `HealthService`
- bounded `PackService` for MCP callers
- request/result contracts and app tests

Why separate:
- these capabilities do not exist yet and need scope discipline to avoid absorbing Phase 6 CLI work

### Slice 5. MCP Mutating and Operational Tools

Deliver:
- refresh MCP tool
- token-tree, pack, and health MCP handlers
- end-to-end MCP registry covering the full promised tool surface
- transport integration tests over STDIO or direct handler execution

Why after slices 3 and 4:
- it keeps app-service creation separate from tool wiring
- it enables one focused plan for full `MCP-03` and `MCP-04` coverage

### Slice 6. Install Flow and Opt-In Client Registration

Deliver:
- supported-client registration adapters
- explicit consent flow and dry-run rendering
- CLI tests and snippet alignment

Why last:
- registration should point at a real, stable `mcp serve` contract
- the install wedge is user-facing polish on top of the now-stable transport

## Requirement-by-Requirement Notes

### `CLI-02`: Optional client registration during install

Planning implications:
- keep writes opt-in and client-specific
- support preview/dry-run output
- prefer adapter-based file rendering over hand-built string edits spread across CLI code
- do not mutate repository files; only client config files when explicitly approved

### `MCP-01`: Serve MCP over STDIO

Planning implications:
- create a dedicated server package
- keep transport bootstrap independent from app services
- add integration tests that exercise the real command path or server loop at least once

### `MCP-02`: Structured machine-readable payloads with freshness and cache metadata

Planning implications:
- define shared metadata envelope once
- standardize error payloads and truncation metadata
- ensure L2 tools can explicitly report when live file reads were required

### `MCP-03`: Expose repository map, symbol lookup, structure lookup, context block, token tree, refresh, pack, and health

Planning implications:
- existing app services cover repository map, context, lookup, context block, refresh
- new work is needed for token tree, pack, and health
- avoid pretending budget analysis alone satisfies token tree; the payload needs a hierarchical structure

### `MCP-04`: Bounded defaults and actionable failures

Planning implications:
- add explicit request validation and bound constants
- failures should name the field, allowed max, and offending value
- do not rely on implicit app-layer truncation without exposing metadata

## Main Risks and Failure Modes

- **Protocol sprawl:** mixing MCP structs into app or repository packages will make later CLI reuse harder.
- **Scope leakage into Phase 6:** pack export and doctor can easily grow beyond the narrow machine-readable surfaces Phase 5 needs.
- **Inconsistent envelopes:** if each tool invents its own freshness or limit fields, `MCP-02` will be only partially satisfied.
- **Hidden truncation:** handlers that quietly clamp large requests will violate the exact-first product promise.
- **Unsafe registration writes:** install flow that edits unsupported or ambiguous client config files will undermine trust.
- **Too little integration testing:** Phase 5 needs at least one real end-to-end STDIO or server-loop proof, not just isolated handler tests.

## Validation Architecture

Phase 5 should use the existing Go test infrastructure with both fast handler/service checks and at least one integration-oriented server/CLI path. A good validation split is:
- transport and handler unit tests for request normalization, error mapping, and registry behavior
- app/service tests for token tree, health, and pack semantics
- CLI tests for `mcp serve` command parsing and registration consent behavior
- one integrated MCP session proof covering at least repository map, lookup, refresh metadata, and one bounded failure

Recommended validation themes:
- structured payload schema stability across repeated calls
- freshness and cache metadata present on all query-style tools
- bounded failures are transparent and field-specific
- registration never writes without explicit opt-in
- token tree, health, and pack stay narrow and deterministic

## Planning Guidance

When turning this research into plans:
- keep each plan on one seam: transport foundation, read-only handlers, token tree, pack/health, operational tool wiring, registration
- derive `must_haves` from the Phase 5 goal, not from implementation convenience
- make tool-boundary and validation expectations explicit in every plan
- keep `pack` and `health` scoped to the minimal MCP contract needed now; leave richer operator UX to Phase 6
- ensure every mapped requirement ID appears in at least one plan frontmatter block

