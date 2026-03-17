# Phase 03 Verification

## Status

`passed`

## Scope

- Phase: `03-structural-extraction-and-repository-artifact-model`
- Goal: Add deterministic parser-backed structural extraction and persist exact symbols, spans, and repository-map building blocks.
- Requirements: `EXTR-01`, `EXTR-02`, `EXTR-03`, `EXTR-04`, `EXTR-05`
- Verified against: current codebase, persisted read models, and test suite behavior

## Inputs Reviewed

- `AGENTS.md`: not present on disk at repo root during verification; used the user-supplied AGENTS instructions in this task instead
- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `.planning/phases/03-structural-extraction-and-repository-artifact-model/03-03-PLAN.md`
- `.planning/phases/03-structural-extraction-and-repository-artifact-model/03-03-SUMMARY.md`
- Relevant implementation and tests under `internal/repository`, `internal/extract`, `internal/store/sqlite`, and `internal/app`

## Verification Summary

Phase 03 goal achievement is verified in the actual implementation. The codebase now:

- detects persisted file languages and routes supported files through a static Tree-sitter adapter registry
- extracts deterministic structural symbols from supported Go files
- persists exact extraction metadata and normalized symbol rows in SQLite, including spans and parent relationships
- degrades unsupported, partial, failed, and skipped files truthfully while keeping them queryable as files
- builds a deterministic compact repository map entirely from persisted SQLite artifacts without reparsing source files on read

I also ran the full Go test suite successfully with the Phase 03 dependencies available:

```bash
env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./...
```

## Requirement Verification

### EXTR-01: Runtime can detect supported languages for indexed files

Status: satisfied

Why:

- Discovery persists per-file language hints during indexing via `languageHint(path)` and `FileRecord.LanguageHint` in [internal/repository/discovery.go](/home/nico/projects/optimusctx/internal/repository/discovery.go#L149) and [internal/repository/discovery.go](/home/nico/projects/optimusctx/internal/repository/discovery.go#L226).
- Extraction routing resolves adapter support from the persisted `ExtractionCandidate.Language` through the registry in [internal/extract/registry.go](/home/nico/projects/optimusctx/internal/extract/registry.go#L10).
- The canonical refresh service registers the Go Tree-sitter adapter in [internal/app/refresh.go](/home/nico/projects/optimusctx/internal/app/refresh.go#L62).

Evidence:

- [internal/repository/discovery_test.go](/home/nico/projects/optimusctx/internal/repository/discovery_test.go#L162) verifies `.go` files persist `LanguageHint == "go"`.
- [internal/extract/engine_test.go](/home/nico/projects/optimusctx/internal/extract/engine_test.go#L28) verifies deterministic registry support and resolution.
- [internal/extract/engine_test.go](/home/nico/projects/optimusctx/internal/extract/engine_test.go#L49) verifies unsupported languages are identified without invoking an adapter.

### EXTR-02: Runtime can extract structural blocks and symbols from supported languages using deterministic parser-backed analysis

Status: satisfied

Why:

- The extraction engine routes supported files through adapter-backed parsing in [internal/extract/engine.go](/home/nico/projects/optimusctx/internal/extract/engine.go#L23).
- The Go adapter uses Tree-sitter parsing and AST walking to emit structural symbols in [internal/extract/adapter/goextract/adapter.go](/home/nico/projects/optimusctx/internal/extract/adapter/goextract/adapter.go#L42).
- Symbol output is normalized into deterministic lexical order with stable ordinals in [internal/extract/types.go](/home/nico/projects/optimusctx/internal/extract/types.go#L37).
- Refresh integrates extraction into the main indexing path and only reprocesses affected files in [internal/app/refresh.go](/home/nico/projects/optimusctx/internal/app/refresh.go#L255).

Evidence:

- [internal/extract/adapter/goextract/adapter_test.go](/home/nico/projects/optimusctx/internal/extract/adapter/goextract/adapter_test.go#L16) verifies parser-backed extraction across package, const, var, type, struct, field, interface, method, and function symbols.
- [internal/extract/adapter/goextract/adapter_test.go](/home/nico/projects/optimusctx/internal/extract/adapter/goextract/adapter_test.go#L46) verifies deterministic extraction across repeated runs.
- [internal/app/refresh_test.go](/home/nico/projects/optimusctx/internal/app/refresh_test.go#L288) verifies extraction is persisted via refresh and only changed files are updated.

### EXTR-03: Runtime stores exact symbol spans, kinds, names, and parent relationships for supported files

Status: satisfied

Why:

- Phase 03 introduced normalized structural storage in `file_extractions` and `symbols` via [internal/store/migrations/0003_structural_artifacts.sql](/home/nico/projects/optimusctx/internal/store/migrations/0003_structural_artifacts.sql#L1).
- The persisted symbol model includes byte spans, row/column spans, name spans, signature spans, kind, name, qualified name, ordinal, depth, and parent linkage in [internal/repository/metadata.go](/home/nico/projects/optimusctx/internal/repository/metadata.go#L164).
- The Go adapter populates exact span fields and parent stable keys in [internal/extract/adapter/goextract/adapter.go](/home/nico/projects/optimusctx/internal/extract/adapter/goextract/adapter.go#L164).
- SQLite artifact replacement persists the extraction row and symbol rows atomically, then resolves `parent_symbol_id` from `ParentStableKey` in [internal/store/sqlite/extraction.go](/home/nico/projects/optimusctx/internal/store/sqlite/extraction.go#L106).

Evidence:

- [internal/extract/adapter/goextract/adapter_test.go](/home/nico/projects/optimusctx/internal/extract/adapter/goextract/adapter_test.go#L38) verifies name and signature spans point at the exact source text.
- [internal/extract/adapter/goextract/adapter_test.go](/home/nico/projects/optimusctx/internal/extract/adapter/goextract/adapter_test.go#L61) verifies parent ownership and qualified names.
- [internal/store/sqlite/store_test.go](/home/nico/projects/optimusctx/internal/store/sqlite/store_test.go#L526) verifies persisted extraction records, parent symbol IDs, and repository structural coverage reads.
- [internal/store/sqlite/store_test.go](/home/nico/projects/optimusctx/internal/store/sqlite/store_test.go#L703) verifies stale symbols are removed on replacement and the current file artifacts remain exact.

### EXTR-04: Runtime degrades gracefully for unsupported or partially parsed files and surfaces coverage state in diagnostics

Status: satisfied

Why:

- Unsupported files are persisted explicitly with `unsupported` coverage and zero symbols via [internal/extract/types.go](/home/nico/projects/optimusctx/internal/extract/types.go#L103).
- The Go adapter distinguishes `partial` from `failed` based on whether trustworthy symbols remain after parse errors in [internal/extract/adapter/goextract/adapter.go](/home/nico/projects/optimusctx/internal/extract/adapter/goextract/adapter.go#L77).
- Refresh turns read or adapter failures into truthful failed artifacts without corrupting unrelated files in [internal/app/refresh.go](/home/nico/projects/optimusctx/internal/app/refresh.go#L357).
- Repository-wide coverage diagnostics are computed from persisted rows in [internal/store/sqlite/store.go](/home/nico/projects/optimusctx/internal/store/sqlite/store.go#L726).
- Repository-map reads surface `partial`, `unsupported`, `failed`, and `skipped` states as coverage gaps in [internal/store/sqlite/store.go](/home/nico/projects/optimusctx/internal/store/sqlite/store.go#L766) and [internal/app/repository_map.go](/home/nico/projects/optimusctx/internal/app/repository_map.go#L136).

Evidence:

- [internal/extract/adapter/goextract/adapter_test.go](/home/nico/projects/optimusctx/internal/extract/adapter/goextract/adapter_test.go#L92) verifies partial extraction keeps only safe symbols and records parser diagnostics.
- [internal/extract/adapter/goextract/adapter_test.go](/home/nico/projects/optimusctx/internal/extract/adapter/goextract/adapter_test.go#L117) verifies failed extraction persists zero symbols.
- [internal/store/sqlite/store_test.go](/home/nico/projects/optimusctx/internal/store/sqlite/store_test.go#L1132) verifies partial and failed replacements clear stale symbols and update structural coverage summaries.
- [internal/app/refresh_test.go](/home/nico/projects/optimusctx/internal/app/refresh_test.go#L406) verifies syntax-break degradation and recovery without poisoning unrelated files.
- [internal/app/repository_map_test.go](/home/nico/projects/optimusctx/internal/app/repository_map_test.go#L72) verifies coverage states and gaps remain visible in the repository-map read model.

### EXTR-05: Runtime can generate a compact repository map from persisted structural artifacts

Status: satisfied

Why:

- The repository-map service reads repository metadata, freshness, directories, and file map records from SQLite only in [internal/app/repository_map.go](/home/nico/projects/optimusctx/internal/app/repository_map.go#L30).
- Repository-map file records come from persisted `files` plus `file_extractions`, with only top-level symbols included in [internal/store/sqlite/store.go](/home/nico/projects/optimusctx/internal/store/sqlite/store.go#L766).
- Nested symbols are deliberately excluded from repository-map payloads, keeping the map compact and deterministic.
- The resulting payload compacts symbols into a lightweight representation in [internal/app/repository_map.go](/home/nico/projects/optimusctx/internal/app/repository_map.go#L119).

Evidence:

- [internal/store/sqlite/store_test.go](/home/nico/projects/optimusctx/internal/store/sqlite/store_test.go#L840) verifies repository-map directory ordering, skipped/partial/supported/unsupported states, and top-level-only symbol compaction.
- [internal/app/repository_map_test.go](/home/nico/projects/optimusctx/internal/app/repository_map_test.go#L154) verifies repository-map generation still works after source files are removed from disk, proving it reads persisted artifacts instead of reparsing.
- [internal/app/repository_map_test.go](/home/nico/projects/optimusctx/internal/app/repository_map_test.go#L190) verifies deterministic repeated repository-map reads.

## Phase Goal Verification

Phase 03 goal: Add deterministic parser-backed structural extraction and persist exact symbols, spans, and repository-map building blocks.

Result: satisfied

Why:

- Deterministic parser-backed extraction exists for supported files through the Go Tree-sitter adapter and symbol normalization.
- Exact symbols and spans are persisted in normalized SQLite tables with parent relationships and coverage metadata.
- Structural persistence is integrated into the canonical refresh transaction, so file inventory and structural artifacts advance together.
- Repository-map building blocks are persisted and queryable independent of the working tree.

## Success Criteria Verification

### The runtime detects supported languages for indexed files and routes supported files through Tree-sitter-based extraction

Satisfied. Language hints are persisted during discovery and consumed by the extraction registry and refresh service routing.

### Supported files persist exact symbol names, kinds, spans, and parent relationships in normalized SQLite tables

Satisfied. `file_extractions` and `symbols` persist exact extraction metadata, symbol spans, and parent links, with tests covering span correctness and parent ownership.

### Unsupported or partially parsed files degrade gracefully, remain queryable as files, and surface coverage gaps in diagnostics metadata

Satisfied. Unsupported, partial, failed, and skipped states are persisted explicitly, remain visible in repository-map output, and contribute to structural coverage diagnostics.

### A compact repository map can be generated entirely from persisted structural artifacts without reparsing the repository on demand

Satisfied. Repository-map generation reads SQLite state only, and persisted-only tests continue to pass after worktree file deletion.

## Test Outcome

Passed:

- full suite: `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache /tmp/optimusctx-go/go/bin/go test ./...`

Notes:

- The default shell environment did not have `go` on `PATH`, so verification used `/tmp/optimusctx-go/go/bin/go`.
- The repo had an unrelated modified file at verification time: `.planning/config.json`. It was not touched.

## Final Verdict

Phase 03 is verified as `passed`.
