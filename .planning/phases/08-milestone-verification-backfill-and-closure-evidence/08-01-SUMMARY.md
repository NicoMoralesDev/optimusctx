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
duration: 2min
completed: 2026-03-15
---

# Phase 08 Plan 01: Evidence Inventory and Verification Contract Summary

**Cross-phase evidence matrix for Phase 2, Phase 5, and bounded Phase 6 verification backfill, with explicit Phase 7 ownership guardrails**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-15T22:07:58Z
- **Completed:** 2026-03-15T22:09:45Z
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

## Task 2 Template And Command Contract

### Reusable Verification File Shape

Phase 8 downstream write-ups should reuse the same concise shape already visible in `03-VERIFICATION.md` and `04-VERIFICATION.md`:

1. `Status`
2. `Scope`
3. `Inputs Reviewed`
4. `Verification Summary`
5. `Requirement Verification`
6. `Phase Goal Verification`
7. `Success Criteria Verification`
8. `Test Outcome`
9. `Final Verdict`

This template choice is deliberate:

- it is short enough to stay readable at milestone-audit time
- it anchors every requirement to concrete evidence instead of narrative-only claims
- it already matches the established verification language in completed phases
- it leaves room for residual notes without turning the file into another execution summary

### Required Content Rules For Each Downstream Verification File

- State the phase goal and requirement list near the top of the file.
- Name the planning inputs reviewed, including current `ROADMAP.md`, `REQUIREMENTS.md`, `STATE.md`, the relevant plan summaries, and the phase `VALIDATION.md`.
- Use one subsection per requirement with `Status: satisfied` or equivalent wording plus direct evidence anchors.
- Mention specific implementation files, test files, and command groups instead of vague references to "the suite" or "the codebase".
- End with an explicit verdict for the phase as a whole.

### Current Command Truth

The current verification command truth for Phase 8 is the `/usr/local/go/bin/go` toolchain with Go build caches redirected into `/tmp`:

```bash
env GOCACHE=/tmp/optimusctx-gocache \
    GOMODCACHE=/tmp/optimusctx-gomodcache \
    GOPROXY=off \
    /usr/local/go/bin/go test ./...
```

Downstream plans should treat this as the canonical current command family because:

- `08-RESEARCH.md` says the current full suite is green with `/usr/local/go/bin/go`
- `08-VALIDATION.md` uses `/usr/local/go/bin/go` plus `GOCACHE=/tmp/optimusctx-gocache` and `GOMODCACHE=/tmp/optimusctx-gomodcache` for both quick and full-suite commands
- Phase 05 and Phase 06 later summaries already reflect the same command family

Older summary-era toolchain paths such as `/tmp/optimusctx-go/go/bin/go` are historical evidence only. They can be mentioned as historical notes when needed, but they must not be copied into the new verification files as the present command truth.

### Downstream Command Groups

#### Phase 02

Use the targeted refresh command set from `08-RESEARCH.md` for requirement-level evidence around:

- migration and store contracts
- hash reuse and diffing
- subtree fingerprints
- transactional refresh reconciliation
- CLI refresh freshness and degraded recovery

#### Phase 05

Use the targeted MCP and install command set from `08-RESEARCH.md` for requirement-level evidence around:

- stdio serve entrypoint and readiness behavior
- registry and structured query envelopes
- bounded failures and structured errors
- token tree, pack, and health tool coverage
- install registration, snippet rendering, and consent gating

#### Phase 06

Use the targeted watch and export command set from `08-VALIDATION.md` for requirement-level evidence around:

- watch-triggered canonical refresh reuse
- debounce and overflow fallback semantics
- pack export manifest generation
- budget fit and include/exclude controls

### Template Guardrails

- Do not turn the verification files into chronological execution logs; the summaries already serve that purpose.
- Do not inherit stale environment notes from early phases unless a historical discrepancy matters to the milestone record.
- Do not restate the Phase 7 doctor/watch repair inside Phase 06 verification beyond the explicit scope boundary.
- Prefer requirement-by-requirement proofs over plan-by-plan recaps.

## Task 3 Publication Guidance

### Downstream Writing Contract

The remaining Phase 8 plans should use this summary as the single inventory source before drafting:

- `08-02` should produce `02-VERIFICATION.md` from the Phase 02 matrix rows and explicitly reconcile earlier `02-UAT.md` failures against later hermetic temp-repository coverage.
- `08-03` should produce `05-VERIFICATION.md` from the Phase 05 matrix rows and synthesize eight summaries into one MCP contract argument.
- `08-04` should produce `06-VERIFICATION.md` from the Phase 06 matrix rows and re-check milestone closure only for `OPS-02`, `OPS-03`, and `OPS-04`.

### Required Evidence Standards

- Every verification file must name all in-scope requirement IDs in its `Scope` section.
- Every requirement subsection must identify at least one summary anchor and at least one implementation or test anchor.
- Every file must record the current `/usr/local/go/bin/go` plus `/tmp` cache command truth, not older summary-era paths.
- Phase 06 must explicitly state that `CLI-05`, `OPS-01`, and `OPS-05` were repaired and verified in Phase 7.

### Unresolved Evidence Risks

These are documentation risks for downstream plans, not implementation blockers:

1. Phase 02 has older `02-UAT.md` history that predates hermetic temp-repo fixes; `02-VERIFICATION.md` needs to explain why the later summaries and current tests are the milestone-grade truth.
2. Some older verification examples still mention `/tmp/optimusctx-go/go/bin/go`; downstream files must avoid copying that stale path into present-day test commands.
3. Phase 05 and Phase 06 each span multiple summaries with overlapping requirement coverage, so downstream write-ups must synthesize rather than duplicate evidence.

### Verification Checklist Outcome

- Requirement inventory: complete for all Phase 8 in-scope IDs
- Phase 7 boundary: explicit for `CLI-05`, `OPS-01`, and `OPS-05`
- Template contract: locked to the established `Status` through `Final Verdict` shape
- Command truth: locked to `/usr/local/go/bin/go` with `GOCACHE=/tmp/optimusctx-gocache` and `GOMODCACHE=/tmp/optimusctx-gomodcache`
- Downstream scope: bounded to verification and traceability, not new feature work

## Task Commits

Each task was committed atomically:

1. **Task 1: Build the Phase 8 requirement-to-evidence inventory** - `fa7ffde` (feat)
2. **Task 2: Lock the verification document template and current command truth** - `3b0e2fe` (feat)
3. **Task 3: Publish the inventory summary for downstream plans** - `86e3a68` (feat)

## Files Created/Modified
- `.planning/phases/08-milestone-verification-backfill-and-closure-evidence/08-01-SUMMARY.md` - Phase 8 inventory, verification contract, downstream evidence matrix, and plan execution record.

## Decisions Made
- Phase 8 inventory is keyed to current requirement ownership in `.planning/REQUIREMENTS.md`, not to older summary-era requirement claims.
- Phase 7 remains the sole owner of the doctor/watch regression repair, so Phase 06 verification backfill stays bounded to `OPS-02..04`.
- Downstream Phase 8 verification files will reuse the Phase 03 and Phase 04 verification structure instead of inventing a new artifact format.
- The canonical current verification commands use `/usr/local/go/bin/go` with `GOCACHE` and `GOMODCACHE` rooted in `/tmp`.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The git worktree already contains unrelated untracked planning artifacts for Phases 07 and 08 plus a modified `.planning/config.json`; this plan will stage only its own summary and final metadata updates.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `08-02` can now write `02-VERIFICATION.md` from a stable requirement matrix, a fixed artifact shape, and a current command set.
- `08-03` and `08-04` can reuse the same verification contract without re-deciding scope or template rules.

## Self-Check

PASSED

- Verified the summary names every Phase 8 in-scope requirement: `REFR-01..05`, `CLI-02`, `MCP-01..04`, and `OPS-02..04`.
- Verified the summary explicitly excludes the Phase 7-owned requirements `CLI-05`, `OPS-01`, and `OPS-05` from later Phase 06 verification.
- Verified Task 1 commit `fa7ffde` exists in git history.
- Verified Task 2 commit `3b0e2fe` exists in git history.
- Verified Task 3 commit `86e3a68` exists in git history.

---
*Phase: 08-milestone-verification-backfill-and-closure-evidence*
*Completed: 2026-03-15*
