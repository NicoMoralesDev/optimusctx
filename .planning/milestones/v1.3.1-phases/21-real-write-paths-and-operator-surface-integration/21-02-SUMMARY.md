---
phase: 21-real-write-paths-and-operator-surface-integration
plan: "02"
subsystem: infra
tags: [mcp, codex, toml, config, install-service, testing]
requires:
  - phase: 20-02
    provides: shared Codex `config.toml` merge/render helpers and preview adapters for Codex App and Codex CLI
provides:
  - write-backed Codex App registration through the shared `config.toml` backend
  - write-backed Codex CLI registration through the shared `config.toml` backend
  - persisted-write regression coverage for explicit paths, preservation, and idempotence
affects:
  - 21-real-write-paths-and-operator-surface-integration
  - codex-client-registration
tech-stack:
  added: []
  patterns:
    - shared preview and write rendering for file-backed Codex registration
    - explicit-path Codex writes over the same merge backend as the default home config
key-files:
  created: []
  modified:
    - internal/app/install.go
    - internal/app/install_test.go
key-decisions:
  - Codex App and Codex CLI continue to share one `config.toml` merge backend for both preview and persisted writes.
  - Codex write support stays file-backed and uses `--config` for repo-local targets instead of introducing a separate `codex mcp add` flow.
patterns-established:
  - "Preview/write parity: Codex write calls the same shared preview merge path with `request.Write = true` before mutating the filesystem."
  - "Preservation-first writes: repeated Codex writes merge into existing TOML and keep unrelated top-level tables and MCP entries intact."
requirements-completed: [MCP-03, CDX-01, CDX-02]
duration: 3min
completed: 2026-03-20
---

# Phase 21 Plan 02: Codex Native Write Path Summary

**Write-backed Codex App and CLI `config.toml` registration with shared TOML merge preservation and idempotence coverage**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-20T01:32:04Z
- **Completed:** 2026-03-20T01:35:05Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Upgraded the shared Codex adapter in `InstallService` from preview-only rendering to real persisted writes for both Codex labels.
- Kept preview and write behavior on one shared `MergeCodexConfig` path so explicit `--config` targets and default home-directory writes stay aligned.
- Added persisted-write regressions for preservation of unrelated TOML content, explicit Codex CLI paths, and repeated-write idempotence.

## Task Commits

Each task was committed atomically:

1. **Task 1: Enable write-backed Codex App and Codex CLI registration** - `1ce2d98` (`test`), `1eb517c` (`feat`)
2. **Task 2: Lock Codex preservation, explicit-path, and idempotence regressions** - `fe76280` (`test`)

## Files Created/Modified

- `internal/app/install.go` - adds the write-backed Codex adapter, shared preview/write mode selection, and updated operator notes for native Codex config targets.
- `internal/app/install_test.go` - locks persisted-write coverage for Codex App and Codex CLI, including preservation of unrelated TOML content and idempotent rewrites.

## Decisions Made

- Codex App and Codex CLI keep sharing one backend so preview output and persisted writes cannot drift by client label.
- Repo-local `.codex/config.toml` support remains driven by the existing `--config` override instead of adding new scope or host-command abstractions in this plan.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 21 now has real Codex persisted writes behind the existing install-service surface.
- Plan `21-03` can update the remaining operator-facing guidance against truthful Codex and Claude write-backed behavior.

## Self-Check: PASSED

- Found `.planning/phases/21-real-write-paths-and-operator-surface-integration/21-02-SUMMARY.md`
- Found task commit `1ce2d98`
- Found task commit `1eb517c`
- Found task commit `fe76280`
