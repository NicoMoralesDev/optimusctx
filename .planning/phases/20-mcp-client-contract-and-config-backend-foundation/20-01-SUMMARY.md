---
phase: 20-mcp-client-contract-and-config-backend-foundation
plan: "01"
subsystem: backend
tags: [mcp, claude-cli, install-service, preview-contract, go]
requires:
  - phase: "19-03"
    provides: "The status/install MCP registration surface that Phase 20 refines into explicit supported-client contracts."
provides:
  - "Explicit supported-client identity tests and repository preview helpers for named MCP clients."
  - "A Claude CLI command-style preview contract that keeps `optimusctx run` canonical."
  - "An install-service adapter split between named-client previews and the generic manual fallback."
affects: [20-02, 20-03, 21-01, 21-02, status, install]
tech-stack:
  added: []
  patterns: ["host-specific preview adapters", "canonical runtime handoff via optimusctx run"]
key-files:
  created: [internal/repository/client_config_test.go]
  modified: [internal/repository/client_config.go, internal/app/install.go, internal/app/install_test.go, internal/cli/install.go, internal/cli/status.go]
key-decisions:
  - "Claude CLI preview uses a `command` config path and renders the native `claude mcp add --transport stdio ...` contract instead of JSON/manual output."
  - "The `generic` adapter remains the only `manual` fallback while named clients move onto explicit preview adapter types."
patterns-established:
  - "Repository helpers own canonical preview rendering so app and CLI layers reuse one source of truth."
  - "InstallService routes supported clients through explicit adapters instead of treating named hosts as aliases of the generic/manual path."
requirements-completed: [MCP-01, MCP-02]
duration: 7min
completed: 2026-03-19
---

# Phase 20 Plan 01: Supported-client contract and preview model refactor Summary

**Explicit supported-client preview helpers with a native Claude CLI command contract and a clean adapter split between named hosts and the generic manual fallback**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-19T20:22:00Z
- **Completed:** 2026-03-19T20:29:19Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Added repository-level preview helpers that lock the supported client IDs and the canonical `optimusctx run` handoff.
- Introduced an explicit Claude CLI command preview contract: `claude mcp add --transport stdio optimusctx -- optimusctx run`.
- Refactored the install-service registry so named clients no longer share the generic/manual adapter surface.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add explicit preview render helpers for named MCP clients** - `fcc23d1` (`feat`)
2. **Task 2: Refactor install adapters so only the generic client remains manual** - `efa2719` (`feat`)

## Files Created/Modified
- `internal/repository/client_config.go` - Added serve-command normalization and the Claude CLI command preview renderer.
- `internal/repository/client_config_test.go` - Locked explicit supported-client IDs, generic JSON rendering, and the Claude CLI command preview string.
- `internal/app/install.go` - Split named preview adapters from the generic/manual fallback and moved Claude CLI onto `ConfigPath: "command"`.
- `internal/app/install_test.go` - Added named-client preview coverage and asserted that `generic` is the only `manual` preview path.
- `internal/cli/install.go` - Ensured command-style preview output is printed cleanly in the deprecated install surface.
- `internal/cli/status.go` - Ensured status preview output is printed cleanly for non-JSON preview contracts.

## Decisions Made
- Used repository-owned preview helpers as the source of truth so app-layer adapters and later write paths can reuse the same runtime handoff contract.
- Reserved `ConfigPath: "manual"` for the generic fallback only; Claude CLI now exposes `ConfigPath: "command"` to reflect the host-native registration model.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Added newline handling for command-style preview output**
- **Found during:** Task 2 (Refactor install adapters so only the generic client remains manual)
- **Issue:** The new Claude CLI command preview would have caused `status` and `install` output to concatenate notes and status text onto the command line.
- **Fix:** Added CLI-side trailing-newline normalization before rendering preview content.
- **Files modified:** `internal/cli/install.go`, `internal/cli/status.go`
- **Verification:** `go test ./internal/cli`
- **Committed in:** `efa2719` (part of Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** The auto-fix was required to keep the new preview contract readable at the existing CLI surfaces. No scope creep.

## Issues Encountered
None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase `20-02` can reuse the explicit adapter seam to move Codex App and Codex CLI onto one native shared `config.toml` preview backend.
- Phase `21-01` can add real Claude CLI `--write` execution without reopening client identity, preview shape, or the canonical `optimusctx run` handoff.

## Self-Check
PASSED

- Verified summary file exists at `.planning/phases/20-mcp-client-contract-and-config-backend-foundation/20-01-SUMMARY.md`.
- Verified task commits `fcc23d1` and `efa2719` exist in git history.

---
*Phase: 20-mcp-client-contract-and-config-backend-foundation*
*Completed: 2026-03-19*
