---
phase: 03-structural-extraction-and-repository-artifact-model
plan: 02
subsystem: api
tags: [tree-sitter, go, structural-extraction, symbols, parser, testing]
requires:
  - phase: 03-structural-extraction-and-repository-artifact-model
    provides: persisted file extraction rows, symbol storage, and extraction candidate metadata
provides:
  - deterministic extraction engine types, routing, and unsupported-language outcomes
  - pinned Tree-sitter Go adapter with exact symbol spans and ownership
  - explicit partial and failed extraction coverage for malformed Go files
affects: [phase-03-refresh-integration, phase-03-repository-map, diagnostics, extraction-runtime]
tech-stack:
  added: [github.com/tree-sitter/go-tree-sitter, github.com/tree-sitter/tree-sitter-go]
  patterns: [adapter-owned parsers, persisted-language routing, deterministic lexical symbol normalization]
key-files:
  created: [.planning/phases/03-structural-extraction-and-repository-artifact-model/03-02-SUMMARY.md, internal/extract/types.go, internal/extract/registry.go, internal/extract/engine.go, internal/extract/engine_test.go, internal/extract/adapter/goextract/adapter.go, internal/extract/adapter/goextract/adapter_test.go, internal/extract/adapter/goextract/testdata/clean.go, internal/extract/adapter/goextract/testdata/syntax_error.go]
  modified: [go.mod, go.sum]
key-decisions:
  - "Resolve extraction support from persisted files.language values plus a static registry and emit explicit unsupported artifacts without parser work."
  - "Keep Tree-sitter parsers adapter-owned and short-lived so concurrency stays simple and deterministic."
  - "Treat malformed Go files as partial only when at least one non-package symbol comes from an error-free subtree; otherwise fail with zero symbols."
patterns-established:
  - "Engine normalization: adapters emit lexical symbols and the extraction core assigns stable ordering, counts, and coverage metadata."
  - "Lexical ownership first: struct and interface bodies become nested containers, and receiver methods attach to same-file type symbols when the receiver type is known."
requirements-completed: [EXTR-01, EXTR-02, EXTR-03, EXTR-04]
duration: 13min
completed: 2026-03-15
---

# Phase 3 Plan 2: Extraction Engine and Go Adapter Summary

**Deterministic extraction routing with a pinned Tree-sitter Go adapter that emits exact lexical symbols, spans, and truthful degraded coverage states**

## Performance

- **Duration:** 13 min
- **Started:** 2026-03-14T23:53:00Z
- **Completed:** 2026-03-15T00:06:09Z
- **Tasks:** 3
- **Files modified:** 10

## Accomplishments
- Added the `internal/extract` boundary with typed requests/results, static adapter registration, unsupported-language handling, and stable symbol normalization.
- Implemented the first production adapter for Go using pinned Tree-sitter packages and fixture-backed assertions for names, spans, determinism, and ownership.
- Made malformed Go extraction truthful by separating `partial` from `failed` outcomes and proving repeated runs keep symbols and coverage metadata stable.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add deterministic extraction types and the adapter registry** - `17632ca` (feat)
2. **Task 2: Implement the Go Tree-sitter adapter with exact symbol extraction** - `1975764` (feat)
3. **Task 3: Add malformed-fixture coverage for partial and failed extraction** - `5f2ee38` (feat)

## Files Created/Modified
- `internal/extract/types.go` - Defines extraction requests/results plus artifact construction and deterministic symbol normalization.
- `internal/extract/registry.go` - Registers adapters statically and resolves supported languages from persisted metadata hints.
- `internal/extract/engine.go` - Routes extraction requests through adapters and emits explicit unsupported artifacts without parser work.
- `internal/extract/engine_test.go` - Covers registry support, engine normalization, and unsupported-language routing.
- `internal/extract/adapter/goextract/adapter.go` - Parses Go files with Tree-sitter and emits package, const, var, type, struct, interface, field, method, and function symbols.
- `internal/extract/adapter/goextract/adapter_test.go` - Verifies clean extraction, stable output, lexical ownership, and malformed partial/failed outcomes.
- `internal/extract/adapter/goextract/testdata/clean.go` - Clean declaration fixture for exact symbol and ownership assertions.
- `internal/extract/adapter/goextract/testdata/syntax_error.go` - Malformed fixture proving partial extraction keeps only trustworthy symbols.
- `go.mod` - Pins the Tree-sitter parser and Go grammar modules used by the adapter.
- `go.sum` - Records checksums for the pinned parser dependencies.

## Decisions Made
- Reused the persisted `files.language` hint as the only routing input so adapter support stays aligned with repository inventory instead of request-time sniffing.
- Kept the engine transport-neutral and limited to ordering, counts, and coverage semantics while adapters stay responsible for parser lifecycle and raw symbol discovery.
- Modeled named `type` declarations separately from nested `struct` and `interface` bodies so fields and interface methods have a precise lexical parent.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Installed pinned Tree-sitter Go modules**
- **Found during:** Task 2 (Implement the Go Tree-sitter adapter with exact symbol extraction)
- **Issue:** The workspace did not have the required Tree-sitter Go parser modules in the local module cache, so the adapter could not compile or run against the real parser API.
- **Fix:** Added `github.com/tree-sitter/go-tree-sitter` and `github.com/tree-sitter/tree-sitter-go` at `v0.25.0` to `go.mod` and `go.sum`.
- **Files modified:** `go.mod`, `go.sum`
- **Verification:** `go test ./... -run 'TestGoAdapter|TestGoSymbolDeterminism|TestGoSymbolOwnership'`
- **Committed in:** `1975764`

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** The added dependencies were required to satisfy the planned Tree-sitter adapter implementation. No scope creep beyond the pinned parser modules.

## Issues Encountered

- Receiver-method ownership initially resolved to the receiver parameter name instead of the receiver type; the adapter now walks the receiver subtree to attach methods to the same-file type symbol deterministically.
- Tree-sitter error counting initially missed anonymous error nodes; malformed coverage now scans full child lists so partial and failed diagnostics remain truthful.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 3 can now wire refresh-driven extraction persistence against a stable engine boundary and one production adapter.
- Repository-map work can assume persisted Go symbols have deterministic lexical ordering, exact spans, and explicit coverage states for malformed files.

## Self-Check: PASSED

- Confirmed `.planning/phases/03-structural-extraction-and-repository-artifact-model/03-02-SUMMARY.md` exists.
- Confirmed task commits `17632ca`, `1975764`, and `5f2ee38` exist in Git history.
