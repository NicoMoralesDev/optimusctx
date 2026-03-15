---
phase: 07-doctor-health-semantics-and-milestone-state-repair
research_date: 2026-03-15
objective: "What do I need to know to PLAN this phase well?"
status: complete
---

# Phase 7 Research: Doctor Health Semantics and Milestone State Repair

## Executive Summary

Phase 7 is a focused gap-closure phase with two distinct workstreams:

1. Repair `doctor` semantics so optional watch absence does not downgrade overall repository health.
2. Bring milestone planning artifacts back into sync with actual executed Phase 6 work and the post-audit gap-closure state.

The implementation bug is real and localized. In the current code, watch absence is represented correctly at the watch-service layer as `absent`, but the doctor aggregation layer converts that into `missing`, and the summary layer treats any non-degraded missing section as an overall `missing` repository state. That violates the Phase 6 audit finding and the stated optional-watch contract.

The planning-artifact repair is smaller than the audit text suggests because the repository has already partially moved forward since the audit was written: `ROADMAP.md` and `REQUIREMENTS.md` already reflect gap-closure phases 07 and 08, while `STATE.md` still contains stale human-readable Phase 6 completion text. Planning should treat the audit as historical evidence, not as a file to “fix.”

## Repository Reality Check

### Planning files reviewed

- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `.planning/ROADMAP.md`
- `.planning/v1.0-v1.0-MILESTONE-AUDIT.md`

### Code areas reviewed

- `internal/app/doctor.go`
- `internal/app/watch.go`
- `internal/repository/doctor.go`
- `internal/repository/watch.go`
- `internal/app/doctor_test.go`
- `internal/app/watch_test.go`
- `internal/cli/doctor.go`
- `internal/cli/doctor_test.go`

### Local guidance files

- `CLAUDE.md`: not present
- `.claude/skills`: not present
- `.agents/skills`: not present

## Current Behavior and Root Cause

### Watch status semantics today

`WatchService.Status` already models watch as optional operational state:

- Missing `watch-status.json` returns `WatchStatusKindAbsent` with reason `watch status file not found`.
- Missing heartbeat, stale heartbeat, or dead PID return `WatchStatusKindStale`.
- Current heartbeat with live PID returns `WatchStatusKindRunning`.

This means the lower layer already distinguishes:

- no watch process ever started or currently active
- unhealthy/stale watch state
- healthy/running watch state

That separation is good and should likely be preserved.

### Doctor semantics bug today

The regression is introduced in `internal/app/doctor.go`:

- `doctorWatchStatus()` maps `WatchStatusKindRunning` -> `healthy`
- `doctorWatchStatus()` maps `WatchStatusKindAbsent` -> `missing`
- everything else becomes `degraded`

Then `doctorSummary()` adds every non-healthy section as an issue and computes the overall status like this:

- any degraded issue makes overall status `degraded`
- otherwise any missing issue makes overall status `missing`

Result: a healthy repo with no watch process becomes overall `missing`, even though watch is explicitly meant to be optional.

### CLI behavior today

`internal/cli/doctor.go` renders both:

- section status: `report.Watch.Status`
- raw watch state: `report.Watch.Health.Status`

So if the app-layer fix changes `report.Watch.Status` from `missing` to a non-failing state while leaving `Health.Status=absent`, the CLI can already communicate “watch absent” separately from overall health. This is useful because it suggests the CLI may need wording adjustments, not a structural rewrite.

## Existing Test Coverage and Gaps

### What is covered now

- `internal/app/doctor_test.go` covers:
  - fully healthy repo with running watch
  - degraded repo with stale watch and refresh/structural problems
- `internal/app/watch_test.go` covers:
  - absent watch status file -> `WatchStatusKindAbsent`
  - stale heartbeat -> `WatchStatusKindStale`
- `internal/cli/doctor_test.go` covers:
  - healthy rendering with running watch
  - degraded and missing rendering mixes

### Missing coverage relevant to Phase 7

There is no integration-style test for:

- healthy repo
- initialized and refreshed state
- no watch process / no watch status file
- `doctor` still reporting overall healthy

That missing case directly matches the audit’s broken flow:

- `optimusctx init`
- `optimusctx refresh`
- `optimusctx doctor`

Planning should explicitly add this scenario at both app and CLI layers, or at minimum app plus one CLI smoke assertion.

## Planning Artifact Drift: Actual Current State

The audit lists several stale-planning issues. In the current repository state:

- `ROADMAP.md` already shows Phase 6 as fully complete and already contains Phases 7 and 8.
- `REQUIREMENTS.md` already remaps the gap-closure work to Phases 7 and 8 and has an updated footer.
- `STATE.md` is only partially updated:
  - YAML frontmatter says `stopped_at: Completed 06-05-PLAN.md`
  - but the prose section still says `Last Activity Description: Completed 06-04 pack export budget controls`
  - and the footer still says it was last updated after completing Phase 6 plan 04

Implication for planning:

- The work is no longer “repair all planning artifacts from scratch.”
- The likely minimum artifact repair is `STATE.md`, plus any small consistency edits needed after Phase 7 is planned or executed.
- The milestone audit file should remain an audit snapshot unless there is a deliberate policy to regenerate audits after fixes.

## Likely Workstreams

## 1. Doctor Health Semantics Repair

Primary goal: make absent watch state informational/optional instead of a repository-health failure.

Likely tasks:

- Decide the canonical doctor status for absent watch.
- Update doctor watch issue/action wording so it remains actionable without implying the repo is broken.
- Ensure overall summary status ignores optional-watch absence when all required sections are healthy.

The cleanest implementation options are:

- Option A: map `WatchStatusKindAbsent` to `DoctorStatusHealthy`, while preserving `Health.Status=absent` and an explanatory reason.
- Option B: introduce a new informational/non-failing doctor status such as `inactive` or `optional`.

Based on current code shape, Option A is the lower-risk Phase 7 choice because:

- `DoctorStatus` only has `healthy`, `degraded`, and `missing`.
- CLI/tests already understand those three values.
- adding a fourth doctor status would widen scope into rendering, issue triage, and possibly future MCP/JSON stability decisions.

If Option A is chosen, planning should be explicit that:

- watch section status may read `healthy` while watch state reads `absent`
- the “reason” text must make that nuance obvious to operators

## 2. Healthy-Repo No-Watch Verification Backfill

Primary goal: lock the audit regression with tests.

Likely tasks:

- Add an app-level test proving absent watch does not degrade overall status on a healthy initialized repo.
- Add a CLI rendering test proving `overall status: healthy` can coexist with `watch state: absent`.
- Consider one end-to-end command test if the repo already has a temp-repo CLI pattern that can exercise `init`, `refresh`, and `doctor`.

This workstream matters because the bug came from a cross-component interpretation mismatch, not from the watch service itself.

## 3. Milestone State Repair

Primary goal: make planning artifacts match the actual project state after Phase 6 completion and the audit follow-up.

Likely tasks:

- Repair stale prose/footer entries in `STATE.md`.
- Confirm `ROADMAP.md` and `REQUIREMENTS.md` still agree with the intended gap-closure ownership after any Phase 7 planning changes.
- Avoid “fixing” the historical audit narrative unless the project explicitly regenerates milestone audits as a separate step.

## Affected Code Areas

### Most likely files to change

- `internal/app/doctor.go`
- `internal/app/doctor_test.go`
- `internal/cli/doctor_test.go`
- `.planning/STATE.md`

### Possible secondary files

- `internal/cli/doctor.go`
  - only if wording should distinguish optional absence more clearly
- `internal/repository/doctor.go`
  - only if a new doctor status is introduced, which is probably unnecessary scope for Phase 7
- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
  - only if consistency edits are still needed after updating `STATE.md`

### Files that likely should not change in Phase 7

- `internal/app/watch.go`
- `internal/repository/watch.go`

Current evidence suggests the watch-service semantics are already correct enough for this phase. The bug is in doctor interpretation, not watch-state production.

## Architecture and Design Guidance

## Preserve the separation between raw watch state and doctor interpretation

The repository already has a useful layering:

- watch service emits raw operational state: `absent`, `stale`, `running`
- doctor converts that into operator-facing health

That is the right seam. Phase 7 should repair the interpretation layer rather than collapsing raw watch states into doctor semantics earlier.

## Keep “optional” semantics localized

The optionality rule appears specific to watch mode. Planning should avoid creating a broad “missing is okay everywhere” mechanism in the summary layer. The safer change is:

- adjust watch section status or issue generation
- keep state/refresh/budget/structural missing statuses behaving as real problems

This reduces the chance of accidentally masking genuinely broken repository state.

## Favor surgical change over status-model expansion

Adding a new `DoctorStatus` value would create more churn than the goal requires. Unless planning uncovers a real downstream need for informational statuses across multiple sections, Phase 7 should remain a narrow semantic repair.

## Risks

## 1. Masking stale or dead watch processes

If the implementation becomes too permissive, it could accidentally classify a stale watch process as healthy. Planning must preserve:

- `absent` as optional/non-failing
- `stale` as degraded/failing
- `running` as healthy

The distinction between absent and stale is the core safety boundary.

## 2. Confusing operator messaging

If watch section status becomes healthy while the raw state remains absent, CLI output may look contradictory unless the reason text is explicit. Planning should decide whether wording like one of the following is needed:

- `watch not running (optional)`
- `watch state absent; background watch is optional`

## 3. Over-editing milestone documents

The audit file captures a point-in-time gap report. Rewriting it to match repaired state would destroy evidence. Planning should distinguish:

- state-of-the-world documents that must be current: `STATE.md`, `ROADMAP.md`, `REQUIREMENTS.md`
- historical evidence documents that should remain snapshots: milestone audit files

## 4. Incomplete regression coverage

A unit-only change without the exact healthy-no-watch flow will likely let this regress again. The plan should require at least one test that reproduces the audited user journey directly.

## Verification Considerations

The minimum useful verification set for Phase 7 is:

1. Healthy repo with no watch-status file:
   - `doctor` overall status is healthy
   - watch raw state is `absent`
   - output/recommended actions do not imply the repo is broken
2. Healthy repo with running watch:
   - remains healthy
3. Repo with stale/dead watch heartbeat:
   - remains degraded
4. Planning artifact consistency:
   - `STATE.md`, `ROADMAP.md`, and `REQUIREMENTS.md` agree on current milestone/phase state after edits

## Validation Architecture

Validation for this phase should be split into two layers.

### Product behavior validation

- App-layer integration test around `DoctorService.Doctor(...)` on a temp repository.
- Assert the exact healthy no-watch scenario:
  - `init`/baseline refresh completed
  - no watch-status file present
  - `report.Watch.Health.Status == absent`
  - `report.Watch.Status` is non-failing
  - `report.Summary.Status == healthy`
- Preserve existing degraded coverage for stale watch state.

### CLI/rendering validation

- CLI test for `formatDoctorReport(...)` or command execution proving:
  - overall status renders `healthy`
  - watch state still renders `absent`
  - reason text communicates optionality clearly

### Planning-artifact validation

- Manual or scripted consistency pass over:
  - `.planning/STATE.md`
  - `.planning/ROADMAP.md`
  - `.planning/REQUIREMENTS.md`
- Confirm these files all reflect:
  - Phase 6 execution complete
  - Phase 7 as gap-closure work for doctor/watch semantics and state repair
  - Phase 8 as verification backfill

### Suggested evidence outputs

- Targeted Go tests for app and CLI doctor coverage.
- Optional command transcript for the exact flow:
  - `optimusctx init`
  - `optimusctx refresh`
  - `optimusctx doctor`
- Diff review of planning docs showing only current-state artifacts changed, not historical audit evidence.

## Recommended Planning Shape

A good Phase 7 plan likely needs 2 or 3 plans:

1. Doctor semantics repair in `internal/app/doctor.go` with targeted test updates.
2. CLI/operator wording and regression coverage for healthy-no-watch behavior.
3. Planning artifact repair focused on `STATE.md` and consistency checks against `ROADMAP.md` and `REQUIREMENTS.md`.

If the team prefers minimum granularity, this can likely be done in 2 plans by combining app+CLI behavior into one implementation plan and keeping document repair separate.

## Open Questions for Planning

- Should absent watch appear as section status `healthy`, or should the section stay non-healthy but be excluded from overall summary severity?
- Does the CLI need new wording so operators understand that absent watch is optional and stale watch is not?
- Is there already a preferred convention for updating `STATE.md` after a gap-closure phase is introduced, or should this phase define it?
- Should the exact `init -> refresh -> doctor` flow be covered only in Go tests, or also documented in a verification artifact?

## Bottom Line

Plan this phase as a narrow semantic repair plus documentation/state alignment, not as a watch-system redesign. The watch service already distinguishes absent vs stale correctly; the fix belongs in doctor aggregation and regression coverage. The only clearly stale planning artifact in the current repo is `STATE.md`, while `ROADMAP.md` and `REQUIREMENTS.md` already appear to reflect the post-audit gap-closure structure and should only need confirmation-level edits, if any.
