---
phase: 21-real-write-paths-and-operator-surface-integration
plan: "01"
subsystem: cli
tags: [mcp, claude-cli, status, install-service, go]
requires:
  - phase: "20-03"
    provides: "Claude Desktop parity protections and the existing install/status registration surface this plan extends for Claude CLI writes."
provides:
  - "Scope-aware Claude CLI command rendering with explicit local, project, and user support."
  - "A real Claude CLI write path that executes the host-native claude mcp add contract through a test seam."
  - "Status command support for --scope and truthful preview or write output for Claude CLI command-mode registration."
affects: [21-02, 21-03, status, install, claude-cli]
tech-stack:
  added: []
  patterns: ["shared preview/write command contract", "exec-backed host-native client adapter", "status-to-install request forwarding"]
key-files:
  created: []
  modified: [internal/repository/client_config.go, internal/repository/client_config_test.go, internal/app/install.go, internal/app/install_test.go, internal/cli/status.go, internal/cli/status_test.go]
key-decisions:
  - "Claude CLI preview and write now reuse one scope-aware rendered command contract instead of a preview-only fallback."
  - "Claude CLI write failures surface host-native remediation by distinguishing missing claude installs from command execution output."
  - "The status command forwards --scope directly into InstallRequest so preview text matches the real write path."
patterns-established:
  - "Command-mode clients should render one canonical registration command and reuse it for both preview and execution."
  - "CLI status output should stay truthful about whether a named-client registration was previewed or written even when ConfigPath is command."
requirements-completed: [MCP-03, CLD-02, CLD-03]
duration: 5min
completed: 2026-03-20
---

# Phase 21 Plan 01: Claude CLI write path Summary

**Claude CLI now renders and executes the same scope-aware `claude mcp add --transport stdio --scope ... -- optimusctx run` contract through install and status**

## Performance

- **Duration:** 5min
- **Started:** 2026-03-20T01:21:13Z
- **Completed:** 2026-03-20T01:26:28Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Added repository-owned Claude CLI scope constants, normalization, and command rendering with explicit `local`, `project`, and `user` support.
- Replaced the preview-only Claude CLI adapter with an exec-backed install adapter that previews and writes through the same rendered command contract.
- Extended `optimusctx status` with `--scope <local|project|user>` so the preview surface matches the real write flow and status messaging stays truthful for command-mode clients.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add scope-aware Claude CLI preview and real command execution** - `f0b73e1` (`test`), `a14f5db` (`feat`)
2. **Task 2: Surface Claude CLI scope selection on the status command** - `b4b6bac` (`test`), `39f8e9c` (`feat`)

## Files Created/Modified
- `internal/repository/client_config.go` - Added Claude CLI scope constants, scope normalization, and `--scope` command rendering.
- `internal/repository/client_config_test.go` - Locked scoped Claude CLI rendering and normalization behavior.
- `internal/app/install.go` - Added `InstallRequest.Scope`, a Claude CLI exec seam, and preview/write parity for command-mode registration.
- `internal/app/install_test.go` - Added regression coverage for scoped previewing, real write execution, unsupported scopes, missing Claude installs, and non-zero command failures.
- `internal/cli/status.go` - Added `--scope` flag parsing and the documented status help contract.
- `internal/cli/status_test.go` - Added status coverage for Claude CLI scope forwarding, command-mode write status, and help output.

## Decisions Made
- Used repository normalization plus one rendered command contract as the single source of truth so preview output and write execution cannot drift.
- Treated Claude CLI registration as a host-native command operation rather than a config-file merge because the documented host contract is `claude mcp add`.
- Kept `status` output behavior unchanged apart from scope forwarding so preview still reports `status: preview only` and real writes still report `status: wrote config` when `ConfigPath` is `command`.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase `21-02` can reuse the install-service write seam and status forwarding pattern for Codex App and Codex CLI persisted writes.
- Phase `21-03` can build operator guidance on top of the now-truthful Claude CLI preview and write contract.

## Self-Check
PASSED

- Verified summary file exists at `.planning/phases/21-real-write-paths-and-operator-surface-integration/21-01-SUMMARY.md`.
- Verified task commits `f0b73e1`, `a14f5db`, `b4b6bac`, and `39f8e9c` exist in git history.

---
*Phase: 21-real-write-paths-and-operator-surface-integration*
*Completed: 2026-03-20*
