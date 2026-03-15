---
phase: 05-mcp-serving-and-integration-contracts
plan: "08"
subsystem: cli
tags: [mcp, cli, claude-desktop, install, snippet]
requires:
  - phase: 05-mcp-serving-and-integration-contracts
    provides: "Preview-first client registration flow and shared MCP config rendering from plan 05-06"
provides:
  - "Canonical executable-path policy for rendered `optimusctx mcp serve` client config"
  - "CLI normalization that prevents transient runtime paths from leaking into preview or write output"
  - "Regression coverage for snippet, install preview, explicit overrides, and go-run path handling"
affects: [phase-05, client-registration, mcp-integration]
tech-stack:
  added: []
  patterns: ["Shared repository-level serve-command rendering", "CLI boundary normalization for reusable config output"]
key-files:
  created: [.planning/phases/05-mcp-serving-and-integration-contracts/05-08-SUMMARY.md]
  modified:
    - internal/repository/client_config.go
    - internal/app/snippet.go
    - internal/app/snippet_test.go
    - internal/app/install.go
    - internal/cli/install.go
    - internal/cli/install_test.go
key-decisions:
  - "Omitted `--binary` now always renders the canonical `optimusctx` command name instead of any runtime-resolved executable path."
  - "Snippet and install continue to share `repository.NewServeCommand`, with explicit binary overrides preserved by passing the operator-supplied path through unchanged."
patterns-established:
  - "Reusable client config output must never depend on the transient location of the current process image."
  - "Snippet guidance and install preview/write paths must assert both command and args alignment in tests."
requirements-completed: [CLI-02, MCP-01]
duration: 1 min
completed: 2026-03-15
---

# Phase 5 Plan 08: Executable-Path Contract Summary

**Canonical `optimusctx mcp serve` rendering now uses the stable `optimusctx` command name by default across snippet and install flows, while preserving explicit binary overrides**

## Performance

- **Duration:** 1 min
- **Started:** 2026-03-15T16:49:21Z
- **Completed:** 2026-03-15T16:50:13Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments
- Centralized the default serve-command path policy so reusable client config renders `optimusctx mcp serve` instead of a placeholder or runtime-derived absolute path.
- Updated install request handling so omitted `--binary` flows normalize to the canonical command contract before preview or write rendering.
- Added deterministic command-boundary coverage for snippet/install alignment, explicit overrides, and ephemeral `go run` path normalization.

## Task Commits

Each task was committed atomically:

1. **Task 1: Define one canonical executable-path policy for Phase 5 client config rendering** - `f532f10` (feat)
2. **Task 2: Normalize unstable runtime-derived executable paths at the CLI boundary** - `1ce5eb3` (fix)
3. **Task 3: Add end-to-end alignment coverage for snippet, preview, and explicit overrides** - `031ced4` (test)

## Files Created/Modified
- `internal/repository/client_config.go` - Centralizes the canonical `optimusctx` default for rendered MCP serve commands.
- `internal/app/snippet.go` - Renders snippet JSON through the shared canonical serve-command helper.
- `internal/app/install.go` - Allows install rendering to reuse the shared default command policy when no explicit binary path is supplied.
- `internal/app/snippet_test.go` - Verifies snippet output no longer emits the placeholder executable path.
- `internal/cli/install.go` - Normalizes omitted-`--binary` flows onto the reusable default contract before registration rendering.
- `internal/cli/install_test.go` - Covers stable preview semantics, explicit overrides, and `go run` path normalization for preview and write modes.

## Decisions Made

- Omitted `--binary` uses `optimusctx` as the reusable registration command even if the current process path is stable, because client config must outlive the current invocation.
- Explicit operator-provided `--binary` values remain transparent and are written exactly as supplied.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Initial `go test` runs failed inside the sandbox when Go tried to write under `/home/nico/.cache/go-build`; verification was rerun successfully with `GOCACHE` and `GOMODCACHE` redirected into `/tmp`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 5 preview and write guidance now share one reusable executable-path contract.
- Remaining Phase 5 work can rely on snippet/install command-path alignment as a stable integration invariant.

## Self-Check: PASSED

- Verified `.planning/phases/05-mcp-serving-and-integration-contracts/05-08-SUMMARY.md` exists.
- Verified task commits `f532f10`, `1ce5eb3`, and `031ced4` exist in git history.

---
*Phase: 05-mcp-serving-and-integration-contracts*
*Completed: 2026-03-15*
