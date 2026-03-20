---
phase: 21-real-write-paths-and-operator-surface-integration
plan: "03"
subsystem: infra
tags: [mcp, onboarding, cli, doctor, snippet, operator-guidance]
requires:
  - phase: 21-02
    provides: real write-backed Claude and Codex client registration paths that onboarding and status can now reference truthfully
provides:
  - supported-client status output that enumerates the four supported native hosts
  - onboarding copy that points operators at the canonical `status --client <client> [--write]` flow
  - doctor and snippet guidance aligned with the supported-client registration contract
affects:
  - 22-documentation-and-compatibility-verification
  - operator-guidance
  - supported-client-onboarding
tech-stack:
  added: []
  patterns:
    - shared supported-client operator copy across runtime status, onboarding, doctor, and snippet surfaces
    - `optimusctx run` remains the canonical runtime handoff while host registration varies by client
key-files:
  created: []
  modified:
    - internal/cli/status.go
    - internal/cli/status_test.go
    - internal/cli/init.go
    - internal/cli/init_onboarding_test.go
    - internal/cli/init_integration_test.go
    - internal/app/snippet.go
    - internal/app/snippet_test.go
    - internal/app/doctor.go
    - internal/cli/doctor_test.go
key-decisions:
  - Supported-client guidance now explicitly enumerates `claude-desktop`, `claude-cli`, `codex-app`, and `codex-cli` instead of implying a Claude Desktop-only path.
  - Operator-facing surfaces keep `optimusctx run` as the canonical runtime handoff and direct registration through `optimusctx status --client <client> [--write]`.
patterns-established:
  - "Truthful operator guidance: every in-product operator surface should reference the same supported-client list and the same `status --client <client> [--write]` discovery path."
  - "Runtime handoff stability: host-specific registration guidance changes must not replace `optimusctx run` as the canonical MCP serve command."
requirements-completed: [OPS-01]
duration: 5min
completed: 2026-03-20
---

# Phase 21 Plan 03: Operator Surface Truthfulness Summary

**Supported-client onboarding, status, doctor, and deprecated snippet guidance now all point at the real Claude and Codex registration paths while keeping `optimusctx run` canonical**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-20T01:38:33Z
- **Completed:** 2026-03-20T01:43:50Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments

- Added a supported-client line to `optimusctx status` and changed the healthy next action to the canonical `status --client <client> [--write]` registration flow.
- Replaced `optimusctx init`'s Claude Desktop-only onboarding text with guidance that truthfully names all four supported hosts.
- Updated the deprecated snippet and doctor MCP remediation text so every operator-facing surface now reinforces the same supported-client story and still hands off to `optimusctx run`.

## Task Commits

Each task was committed atomically:

1. **Task 1: Make status and init onboarding enumerate the supported clients** - `a4c910c` (`feat`)
2. **Task 2: Align doctor and snippet guidance with the supported-client story** - `59e7dc8` (`feat`)

## Files Created/Modified

- `internal/cli/status.go` - adds the supported-client status line and the new healthy next-action copy.
- `internal/cli/status_test.go` - locks the exact supported-client list and canonical status next step.
- `internal/cli/init.go` - replaces the Claude Desktop-only onboarding next step with supported-client guidance.
- `internal/cli/init_onboarding_test.go` - keeps named-client onboarding preview output covered through the updated generic copy.
- `internal/cli/init_integration_test.go` - verifies plain `init` now points operators at the supported-client preview and write path.
- `internal/app/snippet.go` - rewrites deprecated snippet comments to enumerate supported native clients and current status usage examples.
- `internal/app/snippet_test.go` - locks the new snippet examples and guards against a Claude Desktop-only fallback.
- `internal/app/doctor.go` - updates the MCP remediation action to the supported-client validation and registration path.
- `internal/cli/doctor_test.go` - verifies doctor output keeps `optimusctx run` visible and renders the supported-client MCP action text.

## Decisions Made

- The supported-client list is spelled out verbatim in operator copy instead of indirectly referring operators to one example client.
- `optimusctx status --client <client> [--write]` is now the canonical operator discovery path for preview and registration across supported hosts.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Manually corrected stale planning metadata after the GSD update commands**
- **Found during:** Summary and state update
- **Issue:** `state`/`roadmap` automation updated counters but left `ROADMAP.md` Phase 21 checklist entries and `STATE.md` last-activity text stale at plan `21-02`.
- **Fix:** Patched `STATE.md`, `ROADMAP.md`, and `REQUIREMENTS.md` so the planning artifacts reflect completed plan `21-03` and Phase 22 readiness.
- **Files modified:** `.planning/STATE.md`, `.planning/ROADMAP.md`, `.planning/REQUIREMENTS.md`
- **Verification:** Re-read the updated planning files and confirmed Phase 21 plan completion, `OPS-01` completion, and the Phase 22 next step.
- **Committed in:** docs metadata commit

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Planning metadata now matches the completed implementation and summary. No product-scope creep.

## Issues Encountered

- The `roadmap update-plan-progress` helper did not rewrite the visible Phase 21 checklist or next-step text, so the planning docs were corrected manually after the automated state update.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 21 now has truthful in-product guidance for all supported clients, closing `OPS-01`.
- Phase 22 can update public docs and broader verification evidence against one consistent supported-client operator contract.

## Self-Check: PASSED

- Found `.planning/phases/21-real-write-paths-and-operator-surface-integration/21-03-SUMMARY.md`
- Found task commit `a4c910c`
- Found task commit `59e7dc8`
