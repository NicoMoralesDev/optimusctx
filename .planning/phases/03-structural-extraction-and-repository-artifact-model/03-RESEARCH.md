# Phase 3 Research: Structural Extraction and Repository Artifact Model

## Scope and Planning Intent

Phase 3 should add deterministic parser-backed structural extraction on top of the Phase 1 and Phase 2 repository and refresh substrate. The goal is not to build a generic code intelligence layer. The goal is to persist exact structural facts that later context and lookup features can trust without reparsing on demand.

Requirements covered in this phase:
- `EXTR-01`
- `EXTR-02`
- `EXTR-03`
- `EXTR-04`
- `EXTR-05`

This phase is where OptimusCtx becomes more than a file inventory. After Phase 3, the runtime should know which files have trustworthy structure, which ones do not, which exact symbols exist in supported files, and how to build a compact repository map entirely from persisted artifacts.

## Current Repo Context

The current codebase already establishes the right substrate for Phase 3:
- `internal/repository/discovery.go` produces deterministic file records with extension-derived `LanguageHint`.
- `internal/app/refresh.go` computes a generation-scoped refresh plan from persisted and current snapshots.
- `internal/store/migrations/0001_init.sql` and `0002_refresh_state.sql` already normalize repositories, directories, files, and refresh runs.
- `files` rows already persist `content_hash`, `language`, `last_seen_generation`, and `refresh_run_id`, which are the right anchors for extraction invalidation.

Important planning constraints from the current implementation:
- The runtime is Go-first and SQLite-backed already.
- Refresh is the canonical mutation path; Phase 3 should extend it instead of creating a second indexing pipeline.
- The repository currently has no root `AGENTS.md` or `CLAUDE.md` file, so planning should not assume repo-local instruction metadata.
- The current repository is itself mostly Go, which makes Go the correct first-class adapter to plan and validate first.

## What Must Be True After Phase 3

- The runtime can determine whether an included file is structurally supported using persisted repository metadata, not ad hoc request-time guessing.
- Supported files can be parsed through Tree-sitter deterministically and yield exact symbol names, kinds, byte spans, point spans, and parent ownership.
- Unsupported and partially parsed files remain present in file inventory and are explicitly marked with structural coverage state.
- Structural artifacts are replaced per file and per refresh generation so later reads never mix old symbols with new file content.
- Repository map generation reads only persisted `directories`, `files`, and structural artifact tables. It must not re-open source files or re-run parsers.

## Supported-Language Detection Strategy

### Baseline recommendation

Use a two-step language routing model:

1. discovery-time candidate language
   Reuse the existing `files.language` value populated by `internal/repository/discovery.go`. This remains the cheap extension-based candidate.

2. extraction-time adapter resolution
   Map that candidate language to a registered extraction adapter. If no adapter exists, mark the file `unsupported` and do not attempt parsing.

This is the right Phase 3 boundary because the repository already persists extension-derived language hints, and the product only needs deterministic routing for supported grammars. Do not add content sniffing, shebang heuristics, or editor-dependent detection in this phase.

### Repository-metadata-driven behavior

Supported-language detection should use current repository metadata in two ways:

- Per file:
  Route by persisted `files.language`, `files.extension`, and `files.ignore_status`.

- Per repository:
  Build a language histogram from included `files.language` rows and initialize only adapters actually present in the repository during a refresh or extraction run.

This keeps startup and extraction work proportional to the repository instead of loading every grammar optimistically.

### Initial support set

Plan Phase 3 around an explicit v1 floor:
- `go` mandatory
- `javascript` optional if plan capacity allows
- `typescript` optional if plan capacity allows
- everything else remains explicit `unsupported`

For this repository, `go` should be the only required adapter for phase exit because it matches the runtime implementation language and gives the highest validation leverage.

### Do not do in Phase 3

- Do not infer language from file contents.
- Do not use LSPs.
- Do not promise embedded-language support such as JS in HTML or Markdown code fences.
- Do not treat Markdown, JSON, or YAML as mandatory symbol-bearing languages in this phase.

## Tree-sitter Integration Approach

## Standard Stack

- `github.com/tree-sitter/go-tree-sitter` pinned to the project’s chosen `0.25.x` family
- pinned grammar modules in the same compatibility family
- one internal adapter package per language
- parsing only from included file bytes on changed or newly included files

### Parser adapter shape

Phase 3 should introduce a transport-neutral extraction package with an interface close to:

```go
type Adapter interface {
    Language() string
    FileExtensions() []string
    NewParser() *treesitter.Parser
    Extract(req FileExtractionRequest) (FileExtractionResult, error)
}
```

Recommended internal shape:
- `internal/extract/registry`
- `internal/extract/engine`
- `internal/extract/adapter/goextract`
- shared AST-to-artifact mappers and span helpers

Keep adapter registration static and explicit. Do not load grammars dynamically from disk.

### Parser lifecycle

Use conservative parser ownership rules:
- language objects and queries are process-wide immutable singletons
- parser instances are not shared across goroutines
- each worker gets its own parser for one language at a time
- close parser/tree resources immediately after extraction

This avoids hidden concurrency bugs and keeps the first implementation easy to reason about.

### Extraction trigger model

Extraction should run as part of the canonical refresh pipeline after file inventory reconciliation determines the affected included files.

Recommended trigger set:
- newly added supported included file
- content-changed supported included file
- ignored -> included transition for a supported file
- file whose persisted extraction grammar version no longer matches the current adapter version

Recommended delete/invalidate set:
- deleted file
- included -> ignored transition
- file moved to a path whose language routing changes
- file that becomes unsupported because adapter support was removed or grammar pin changed

### Deterministic extraction boundaries

Phase 3 should explicitly limit what counts as a structural artifact:

- top-level declarations
- nested declarations with clear lexical ownership
- exact name text when the grammar exposes a stable name node
- exact byte and point spans
- parent-child ownership derived from AST containment, not semantic resolution

Do not extract:
- references
- type inference results
- semantic call graphs
- cross-file symbol resolution
- docstrings as a required artifact

For Go, the initial symbol set should include:
- package
- const
- var
- type
- struct
- interface
- field
- method
- function

### Partial parse handling

Tree-sitter is error-tolerant, but the product promise is exact-first. The correct Phase 3 rule is:

- if the file parses cleanly enough for deterministic symbol extraction, persist symbols normally
- if the tree contains errors but a symbol can be extracted from an error-free subtree with stable spans, persist it and mark the file `partial`
- if parse failure makes symbol boundaries ambiguous, persist no symbols for that file and mark it `failed`

This gives graceful degradation without pretending error-heavy files are fully indexed.

## SQLite Artifact Model

Phase 3 should add normalized structural tables and one per-file extraction-state table. Do not store opaque parser blobs.

### Keep existing Phase 2 tables as the source of file inventory truth

- `repositories`
- `directories`
- `files`
- `refresh_runs`
- `refresh_file_events`

Structural tables should reference active file rows and be replaced transactionally per file.

### Add `file_extractions`

Purpose:
- one row per active file describing structural coverage, provenance, and replacement state

Recommended columns:
- `id` integer primary key
- `repository_id` integer not null
- `file_id` integer not null unique
- `path` text not null
- `language` text not null
- `adapter_name` text not null
- `grammar_version` text not null
- `source_content_hash` text not null
- `source_generation` integer not null
- `coverage_state` text not null
- `coverage_reason` text
- `parser_error_count` integer not null default 0
- `has_error_nodes` integer not null default 0
- `symbol_count` integer not null default 0
- `top_level_symbol_count` integer not null default 0
- `max_symbol_depth` integer not null default 0
- `extracted_at` text not null
- `refresh_run_id` integer

Coverage states:
- `supported`
- `partial`
- `unsupported`
- `failed`
- `skipped`

Recommended use of `skipped`:
- reserved for future policy cases such as size limits
- optional for Phase 3 if no skip policies are introduced

### Add `symbols`

Purpose:
- exact structural inventory for supported files

Recommended columns:
- `id` integer primary key
- `repository_id` integer not null
- `file_id` integer not null
- `file_extraction_id` integer not null
- `stable_key` text not null
- `parent_symbol_id` integer
- `path` text not null
- `language` text not null
- `kind` text not null
- `name` text not null
- `qualified_name` text
- `ordinal` integer not null
- `depth` integer not null
- `start_byte` integer not null
- `end_byte` integer not null
- `start_row` integer not null
- `start_column` integer not null
- `end_row` integer not null
- `end_column` integer not null
- `name_start_byte` integer
- `name_end_byte` integer
- `signature_start_byte` integer
- `signature_end_byte` integer
- `is_exported` integer not null default 0

Rules:
- `stable_key` should be deterministic within a file version, for example a hash of `(kind, qualified_name_or_name, start_byte, end_byte)`
- `ordinal` should preserve lexical encounter order
- `parent_symbol_id` should only encode lexical ownership inside the same file

### Add `file_symbol_edges` only if import extraction is in scope

This is optional for Phase 3, but if repository map quality needs lightweight relations, keep the edge model narrow:

- `id` integer primary key
- `repository_id` integer not null
- `file_id` integer not null
- `symbol_id` integer
- `edge_kind` text not null
- `target_path_hint` text
- `target_name` text
- `ordinal` integer not null

If planning time is tight, defer this table and build Phase 3 repository maps from directories plus top-level symbols only. That is enough to satisfy `EXTR-05`.

### Add indexes

Minimum useful indexes:
- `file_extractions(repository_id, coverage_state)`
- `file_extractions(repository_id, language)`
- `file_extractions(repository_id, source_generation)`
- `symbols(repository_id, file_id, ordinal)`
- `symbols(repository_id, name)`
- `symbols(repository_id, qualified_name)`
- `symbols(repository_id, kind)`
- `symbols(repository_id, parent_symbol_id)`
- `symbols(repository_id, path, start_byte)`

### Replacement semantics

For a file being re-extracted:
- delete prior `symbols` rows for that file via `file_extraction_id` or `file_id`
- upsert the new `file_extractions` row
- insert the new symbol rows

This should happen in one SQLite transaction per extraction batch so a file never points to mixed old and new symbols.

## Repository-Map Building Blocks

Phase 3 should persist the facts needed to compose a repository map, not the final rendered map text.

Repository map generation should read only:
- `directories` for hierarchy and subtree counts
- `files` for path, language, ignore state, and size
- `file_extractions` for coverage state and symbol counts
- `symbols` for top-level declarations and lexical order

Recommended repository-map composition rules:
- include only `files.ignore_status = included`
- prefer files with `coverage_state in ('supported', 'partial')`
- show top-level symbols only
- order directories and files lexically
- within a file, order symbols by `ordinal`
- if a file is unsupported or failed, include the file entry without symbols and annotate coverage state in the returned payload

This keeps `EXTR-05` fully persisted and deterministic while avoiding a premature formatting cache.

## Degraded Behavior for Unsupported or Partially Parsed Files

Graceful degradation should be explicit and persisted.

### Unsupported files

Behavior:
- keep the `files` row
- create or update `file_extractions` with `coverage_state = 'unsupported'`
- persist zero symbols
- set `coverage_reason` to a stable code such as `unsupported_language`

### Partially parsed files

Behavior:
- keep the `files` row
- persist only symbols whose spans are stable and not derived from broken nodes
- mark `coverage_state = 'partial'`
- persist parser error metadata

### Failed extraction

Behavior:
- keep the `files` row
- persist zero symbols for that file version
- mark `coverage_state = 'failed'`
- persist a stable failure reason such as `parse_error`, `adapter_error`, or `query_error`

### Repository freshness interaction

Phase 3 should not overload the Phase 2 refresh generation model, but it does need structural truthfulness.

Recommended rule:
- file inventory freshness remains governed by Phase 2 refresh state
- structural coverage is governed by `file_extractions.coverage_state`
- repository-wide structural health is computed from persisted extraction rows, not a new hidden in-memory flag

If extraction for a changed supported file fails, the repository can still be file-inventory `fresh` while structural coverage is partially degraded. Later doctor and MCP surfaces can compute that honestly from persisted extraction rows.

## Common Pitfalls

### Pitfall 1: treating language support as binary

Do not force files into just supported or unsupported. Phase 3 needs `partial` and `failed` so diagnostics and later query layers can be truthful.

### Pitfall 2: reparsing on repository-map reads

That would violate `EXTR-05` and hide indexing drift. Repository-map generation must query persisted artifacts only.

### Pitfall 3: storing parser-specific opaque blobs

That makes migrations and cross-version compatibility harder. Store normalized spans and symbol facts instead.

### Pitfall 4: letting extraction infer semantics

Phase 3 is about lexical structure. Avoid references, inferred types, or cross-file resolution.

### Pitfall 5: sharing parser instances unsafely

Keep parser ownership per worker. Concurrency bugs here will look like nondeterministic symbol output.

## Don’t Hand-Roll

- Gitignore semantics beyond the current repository matcher
- custom parsers instead of Tree-sitter
- fuzzy symbol extraction via regex
- parser result caches outside SQLite for v1
- semantic name resolution

## Architecture Patterns

### Pattern 1: refresh-driven extraction queue

The refresh pipeline identifies affected files. The extraction engine consumes only that deterministic set.

### Pattern 2: per-file replacement artifacts

Each file’s structural rows are fully replaced for a given content hash. Do not patch symbol rows in place.

### Pattern 3: persisted coverage truth

Queries and diagnostics should answer from persisted coverage states, not from parser attempts at request time.

## Testing Strategy

Phase 3 needs three kinds of validation: parser correctness, persistence correctness, and deterministic composition correctness.

### Parser and adapter tests

- Go golden fixtures covering clean files, nested declarations, methods, receiver methods, anonymous structs, and syntax errors
- table-driven extraction assertions on exact kinds, names, parent ownership, and byte spans
- repeated-run determinism tests proving the same file yields identical symbol order and keys

### Store and migration tests

- migration tests for new extraction tables and indexes
- replacement tests proving re-extracting a file deletes old symbols and inserts the new set atomically
- delete and ignore-transition tests proving artifacts are removed when a file leaves the included set

### Refresh-to-extract integration tests

- init on a temp Git repo with Go files populates symbols
- no-op refresh keeps symbol counts and extraction state stable
- content edit replaces only affected file artifacts
- rename or move removes old-path artifacts and writes new-path artifacts
- syntax-breaking edit changes the file to `partial` or `failed` without corrupting other files

### Repository-map tests

- repository map output is generated entirely from DB state after source files are deleted or made unavailable to the query path
- unsupported and partial files appear with truthful coverage metadata
- ordering is stable across repeated runs

## Code Examples

Recommended extraction orchestration shape:

```go
type ExtractionCandidate struct {
    FileID       int64
    Path         string
    Language     string
    ContentHash  string
    Generation   int64
}

type ExtractionEngine interface {
    ExtractBatch(ctx context.Context, candidates []ExtractionCandidate) ([]FileExtractionResult, error)
}
```

Recommended persistence boundary:

```go
type ArtifactStore interface {
    ReplaceFileArtifacts(ctx context.Context, fileID int64, result FileExtractionResult) error
    DeleteFileArtifacts(ctx context.Context, fileID int64) error
}
```

## Risks and Sequencing into Executable Plans

The main Phase 3 risks are not parser choice. They are schema correctness, deterministic boundaries, and degraded-state truthfulness.

### Primary risks

- Grammar pin drift creates incompatible extraction results across machines.
- Partial parse behavior becomes inconsistent and impossible to diagnose.
- Repository map scope expands into relationships or semantic ranking too early.
- Artifact replacement is not transactional, leaving stale symbols behind.
- Planning multiple language adapters at once dilutes validation and delays a correct Go path.

### Recommended plan sequence

1. schema and repository-artifact contracts
   Add migrations, extraction-state enums, symbol table schema, indexes, and store read/write tests.

2. extraction package skeleton and adapter registry
   Introduce the registry, engine interfaces, span helpers, and deterministic symbol model types.

3. Go Tree-sitter adapter
   Implement the Go grammar integration, clean-file extraction, stable symbol ordering, and parent ownership.

4. refresh-integrated artifact invalidation and replacement
   Wire extraction candidates from refresh results, replace artifacts transactionally, and remove artifacts on delete or ignore transitions.

5. degraded coverage and repository-map composition
   Persist unsupported, partial, and failed states and build repository-map output from persisted artifacts only.

6. full validation pass
   Add temp-repo integration coverage, migration coverage, determinism coverage, and failure-path tests.

This sequence keeps the phase executable. It also avoids a common failure mode where multiple grammars are added before one adapter and one persistence path are proven correct.

## Validation Architecture

Phase validation should prove that structural artifacts are exact, persistent, and queryable without reparsing.

### 1. Migration and schema validation

- verify new tables, foreign keys, and indexes exist
- verify old databases migrate cleanly from Phase 2 to Phase 3
- verify file artifact replacement leaves no duplicate symbol rows

### 2. Adapter correctness fixtures

- use focused Go fixture files under `testdata/`
- assert symbol names, kinds, byte spans, point spans, depth, and parent relationships
- include malformed files and confirm `partial` or `failed` coverage states

### 3. End-to-end repository progression tests

Use disposable Git repositories with explicit stages:
- baseline clean Go repository
- add a new file
- edit an existing symbol
- rename or move a file
- introduce a syntax error
- repair the syntax error
- ignore and re-include a file

For each stage assert:
- `files` rows
- `file_extractions` coverage state and provenance
- `symbols` contents for affected files
- repository map payload shape and ordering

### 4. Persisted-only repository-map validation

Add a test that:
- runs init or refresh to populate artifacts
- closes the runtime
- builds repository-map output from SQLite alone

This is the direct proof for `EXTR-05`.

### 5. Determinism validation

- repeated extraction of the same repository yields identical symbol ordering and stable keys
- no-op refresh does not rewrite unaffected artifact rows
- repeated repository-map generation yields byte-for-byte equivalent serialized payloads when the DB is unchanged

### 6. Failure-path validation

- adapter error for one file does not remove artifacts for unrelated files
- failed or partial files retain truthful persisted coverage state
- delete or ignore transitions remove obsolete file artifacts

### Validation exit criteria

- `EXTR-01`: supported files are routed by persisted language metadata and unsupported files are marked explicitly
- `EXTR-02`: supported Go files extract deterministic parser-backed symbols
- `EXTR-03`: persisted symbols include exact spans, kinds, names, and parent ownership
- `EXTR-04`: unsupported and partial/failed files surface truthful coverage states
- `EXTR-05`: repository-map generation reads persisted artifacts only

## Prescriptive Recommendations

- Start with one production-quality adapter: Go.
- Route support using the existing persisted `files.language` metadata.
- Add `file_extractions` and `symbols` as the mandatory Phase 3 tables.
- Keep structural extraction lexical and deterministic; do not add semantic analysis.
- Replace file artifacts transactionally per file version.
- Build repository maps from persisted `directories`, `files`, `file_extractions`, and `symbols` only.
- Make degraded coverage explicit through persisted states, not logs or in-memory flags.

If Phase 3 exits with that model, Phase 4 can build exact lookup and layered context on top of stable artifacts instead of inventing structure at query time.
