---
phase: 05-mcp-serving-and-integration-contracts
plan: "06"
subsystem: cli
tags: [go, mcp, cli, install, claude-desktop, json]
requires:
  - phase: 05-01
    provides: stable `optimusctx mcp serve` CLI contract
  - phase: 05-05
    provides: completed MCP server surface that registration now targets
provides:
  - supported-client config models and JSON rendering for MCP registration
  - opt-in install command with preview-first and explicit write behavior
  - command-boundary coverage for dry-run, write consent, and unsupported clients
affects: [phase-06-ops, install-flow, snippet-guidance, supported-clients]
tech-stack:
  added: []
  patterns: [adapter-based client registration, preview-first install flow, shared snippet and install contract]
key-files:
  created: [internal/repository/client_config.go, internal/app/install.go, internal/cli/install.go, internal/cli/install_test.go]
  modified: [internal/app/snippet.go, internal/app/snippet_test.go, internal/cli/root.go, internal/cli/init_integration_test.go]
key-decisions:
  - "Supported-client registration stays adapter-based and preview-first, with writes allowed only behind `--write`."
  - "Snippet and install registration share the same rendered MCP JSON contract so the manual and automated paths cannot drift."
  - "Claude Desktop is the initial supported client, with explicit `--config` override support to keep tests hermetic and platform behavior transparent."
patterns-established:
  - "Client registration adapters own path resolution and JSON merge behavior while app and CLI layers stay transport-neutral."
  - "Preview output prints the exact merged config document before any write path is taken."
requirements-completed: [CLI-02]
duration: 13 min
completed: 2026-03-15
---

# Phase 5 Plan 06: Install Registration Summary

**Opt-in Claude Desktop MCP registration with shared JSON rendering for snippet previews and explicit config writes**

## Performance

- **Duration:** 13 min
- **Started:** 2026-03-15T15:24:00Z
- **Completed:** 2026-03-15T15:36:31Z
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments
- Added transport-neutral supported-client models and JSON rendering for MCP registration targets.
- Introduced an `install` CLI flow that previews by default and writes only with explicit consent.
- Added command-boundary coverage proving dry-run behavior, consented writes, unsupported-client failures, and snippet alignment.

## Task Commits

Each task was committed atomically:

1. **Task 1: Define supported-client registration models and rendering contracts** - `1e1a589` (feat)
2. **Task 2: Implement the opt-in registration service and CLI command** - `872682c` (feat)
3. **Task 3: Add real command-boundary coverage for dry-run, consent, and failure cases** - `090de3f` (test)

## Files Created/Modified
- `internal/repository/client_config.go` - shared supported-client, serve-command, merge, and JSON render contracts
- `internal/app/install.go` - adapter-based registration service with preview and explicit write paths
- `internal/cli/install.go` - first-class install command with manual flag parsing and consent-gated writes
- `internal/cli/install_test.go` - command-boundary tests for preview, write, unsupported client, and snippet alignment
- `internal/app/snippet.go` - snippet output aligned to the real `optimusctx mcp serve` contract
- `internal/app/snippet_test.go` - snippet rendering assertions against the shared JSON contract
- `internal/cli/root.go` - root command wiring and help text for `install`
- `internal/cli/init_integration_test.go` - updated integration assertions for the Phase 5 snippet contract

## Decisions Made
- Used a shared JSON rendering path in `internal/repository` so snippet and install preview output are generated from the same MCP server definition.
- Kept registration preview-first instead of interactive confirmation to preserve non-invasive installs while still satisfying explicit-consent writes.
- Scoped supported-client automation to Claude Desktop first, but structured the app layer around adapters so more clients can be added without rewriting the CLI contract.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- Initial Go test runs inside the sandbox could not populate missing module artifacts. Verification completed after allowing the necessary `go test` download path and then reusing `/tmp/optimusctx-gocache` and `/tmp/optimusctx-gomodcache`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 5 now has an explicit install wedge that points at the stable MCP serve contract instead of placeholder snippet guidance.
- Phase 6 can build operator-facing install and health ergonomics on top of the same registration and snippet primitives without reworking the MCP contract.

## Self-Check: PASSED

- Found summary file `.planning/phases/05-mcp-serving-and-integration-contracts/05-06-SUMMARY.md`
- Found task commit `1e1a589`
- Found task commit `872682c`
- Found task commit `090de3f`

---
*Phase: 05-mcp-serving-and-integration-contracts*
*Completed: 2026-03-15*
