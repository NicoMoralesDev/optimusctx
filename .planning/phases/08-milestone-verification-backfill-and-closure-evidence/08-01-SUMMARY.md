---
phase: 08-milestone-verification-backfill-and-closure-evidence
plan: "01"
subsystem: docs
tags: [verification, traceability, milestone-audit, requirements, evidence]
requires:
  - phase: 07-doctor-health-semantics-and-milestone-state-repair
    provides: explicit ownership of CLI-05, OPS-01, and OPS-05 plus repaired doctor/watch semantics
provides:
  - requirement-to-evidence inventory for Phase 8 verification backfill
  - explicit Phase 7 ownership boundary for doctor-related requirements
  - reusable verification contract for downstream Phase 8 plans
affects: [phase-08-02, phase-08-03, phase-08-04, milestone-audit]
tech-stack:
  added: []
  patterns: [requirement-to-evidence matrices, phase-bounded verification backfill, verification-template reuse]
key-files:
  created: [.planning/phases/08-milestone-verification-backfill-and-closure-evidence/08-01-SUMMARY.md]
  modified: []
key-decisions:
  - "Phase 8 verification backfill uses current planning sources and current code truth while keeping historical audit artifacts immutable."
  - "Phase 06 backfill must exclude CLI-05, OPS-01, and OPS-05 because those doctor/watch requirements moved to Phase 7."
patterns-established:
  - "Every downstream verification file should map each in-scope requirement to concrete summaries, implementation anchors, test names, and command groups."
  - "Requirement ownership boundaries must be restated explicitly before phase-level verification backfill begins."
requirements-completed: [REFR-01, REFR-02, REFR-03, REFR-04, REFR-05, CLI-02, MCP-01, MCP-02, MCP-03, MCP-04, OPS-02, OPS-03, OPS-04]
duration: in_progress
completed: 2026-03-15
---

# Phase 08 Plan 01: Evidence Inventory and Verification Contract Summary

**Cross-phase evidence matrix for Phase 2, Phase 5, and bounded Phase 6 verification backfill, with explicit Phase 7 ownership guardrails**

## Performance

- **Duration:** in progress
- **Started:** 2026-03-15T22:07:58Z
- **Completed:** in progress
- **Tasks:** 3
- **Files modified:** 1

## Accomplishments
- Built one requirement-to-evidence inventory for all Phase 8 in-scope IDs across Phase 02, Phase 05, and Phase 06.
- Locked the Phase 7 ownership boundary so later Phase 06 verification cannot reclaim `CLI-05`, `OPS-01`, or `OPS-05`.
- Established a single downstream contract for evidence anchors, test names, and verification command groups.

## Task 1 Inventory

### In-Scope Requirement Set

Phase 8 backfill is responsible only for:

- `REFR-01`, `REFR-02`, `REFR-03`, `REFR-04`, `REFR-05`
- `CLI-02`
- `MCP-01`, `MCP-02`, `MCP-03`, `MCP-04`
- `OPS-02`, `OPS-03`, `OPS-04`

The requirement list matches the current blockers in `.planning/v1.0-v1.0-MILESTONE-AUDIT.md`, the pending ownership rows in `.planning/REQUIREMENTS.md`, and the scoped Phase 8 research notes in `.planning/phases/08-milestone-verification-backfill-and-closure-evidence/08-RESEARCH.md`.

### Excluded Phase 7 Ownership Boundary

Phase 8 must not claim or re-verify the doctor/watch repair requirements:

- `CLI-05`
- `OPS-01`
- `OPS-05`

Those requirements are now owned by Phase 7 in current planning truth:

- `.planning/REQUIREMENTS.md` maps all three to Phase 7 with status `Complete`
- `.planning/ROADMAP.md` defines Phase 7 as the doctor/watch regression closure phase
- `.planning/phases/08-milestone-verification-backfill-and-closure-evidence/08-RESEARCH.md` explicitly limits Phase 06 backfill to `OPS-02`, `OPS-03`, and `OPS-04`

Downstream implication: `06-VERIFICATION.md` must describe watch-refresh reuse, portable export, and budget fitting only. It must not treat doctor health aggregation as Phase 06 evidence.

### Requirement-to-Evidence Inventory

| Requirement | Phase 8 Target File | Summary Evidence | Implementation/Test Anchors | Command Group |
| --- | --- | --- | --- | --- |
| `REFR-01` | `02-VERIFICATION.md` | `02-02-SUMMARY.md`, `02-05-SUMMARY.md` | `internal/repository/discovery_test.go`, `internal/refresh/diff_test.go` | Phase 02 targeted refresh suite |
| `REFR-02` | `02-VERIFICATION.md` | `02-01-SUMMARY.md`, `02-02-SUMMARY.md` | `internal/store/sqlite/store_test.go`, `internal/refresh/fingerprint_test.go` | Phase 02 targeted refresh suite |
| `REFR-03` | `02-VERIFICATION.md` | `02-02-SUMMARY.md`, `02-03-SUMMARY.md`, `02-05-SUMMARY.md` | `internal/refresh/diff_test.go`, `internal/store/sqlite/refresh_test.go` | Phase 02 targeted refresh suite |
| `REFR-04` | `02-VERIFICATION.md` | `02-03-SUMMARY.md`, `02-04-SUMMARY.md`, `02-05-SUMMARY.md`, `02-06-SUMMARY.md` | `internal/app/refresh_test.go`, `internal/cli/refresh_test.go`, `internal/cli/refresh_integration_test.go` | Phase 02 targeted refresh suite |
| `REFR-05` | `02-VERIFICATION.md` | `02-01-SUMMARY.md`, `02-03-SUMMARY.md`, `02-04-SUMMARY.md`, `02-06-SUMMARY.md` | `internal/store/sqlite/store_test.go`, `internal/app/refresh_test.go`, `internal/cli/refresh_integration_test.go`, `README.md` | Phase 02 targeted refresh suite plus README smoke path |
| `CLI-02` | `05-VERIFICATION.md` | `05-06-SUMMARY.md`, `05-07-SUMMARY.md`, `05-08-SUMMARY.md` | `internal/cli/install_test.go`, `internal/app/snippet_test.go`, `internal/repository/client_config.go` | Phase 05 targeted MCP/install suite |
| `MCP-01` | `05-VERIFICATION.md` | `05-01-SUMMARY.md`, `05-05-SUMMARY.md`, `05-07-SUMMARY.md`, `05-08-SUMMARY.md` | `internal/mcp/server.go`, `internal/mcp/server_test.go`, `internal/mcp/integration_test.go`, `internal/cli/mcp_test.go` | Phase 05 targeted MCP/install suite |
| `MCP-02` | `05-VERIFICATION.md` | `05-02-SUMMARY.md`, `05-05-SUMMARY.md` | `internal/mcp/query_tools.go`, `internal/mcp/query_tools_test.go` | Phase 05 targeted MCP/install suite |
| `MCP-03` | `05-VERIFICATION.md` | `05-02-SUMMARY.md`, `05-03-SUMMARY.md`, `05-04-SUMMARY.md`, `05-05-SUMMARY.md` | `internal/mcp/query_tools_test.go`, `internal/app/token_tree_test.go`, `internal/app/health_pack_test.go` | Phase 05 targeted MCP/install suite |
| `MCP-04` | `05-VERIFICATION.md` | `05-02-SUMMARY.md`, `05-03-SUMMARY.md`, `05-04-SUMMARY.md`, `05-05-SUMMARY.md` | `internal/mcp/query_tools_test.go`, `internal/mcp/server_test.go` | Phase 05 targeted MCP/install suite |
| `OPS-02` | `06-VERIFICATION.md` | `06-02-SUMMARY.md` | `internal/app/watch_test.go`, `internal/app/refresh_test.go`, `internal/cli/watch_test.go` | Phase 06 targeted watch/export suite |
| `OPS-03` | `06-VERIFICATION.md` | `06-03-SUMMARY.md`, `06-04-SUMMARY.md` | `internal/app/pack_export_test.go`, `internal/cli/pack_test.go` | Phase 06 targeted watch/export suite |
| `OPS-04` | `06-VERIFICATION.md` | `06-04-SUMMARY.md` | `internal/app/pack_export_test.go`, `internal/cli/pack_test.go` | Phase 06 targeted watch/export suite |

### Evidence Group Notes

- Phase 02 evidence should treat `02-01` through `02-06` as one cumulative refresh story rather than six isolated documents.
- Phase 05 evidence should collapse eight summaries into one requirement argument centered on stdio serving, structured query envelopes, bounded failures, and install registration.
- Phase 06 evidence should separate valid watch/export work from the later doctor semantics repair so the milestone trace remains clean.

## Task Commits

Each task was committed atomically:

1. **Task 1: Build the Phase 8 requirement-to-evidence inventory** - pending
2. **Task 2: Lock the verification document template and current command truth** - pending
3. **Task 3: Publish the inventory summary for downstream plans** - pending

## Files Created/Modified
- `.planning/phases/08-milestone-verification-backfill-and-closure-evidence/08-01-SUMMARY.md` - Phase 8 inventory, verification contract, downstream evidence matrix, and plan execution record.

## Decisions Made
- Phase 8 inventory is keyed to current requirement ownership in `.planning/REQUIREMENTS.md`, not to older summary-era requirement claims.
- Phase 7 remains the sole owner of the doctor/watch regression repair, so Phase 06 verification backfill stays bounded to `OPS-02..04`.

## Deviations from Plan

None so far.

## Issues Encountered

- The git worktree already contains unrelated untracked planning artifacts for Phases 07 and 08 plus a modified `.planning/config.json`; this plan will stage only its own summary and final metadata updates.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Task 2 will convert this inventory into a fixed verification template and one current command set for all downstream write-ups.
- Task 3 will finalize downstream risks, self-checks, and publication details once the contract text is complete.
