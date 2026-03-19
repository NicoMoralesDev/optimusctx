---
phase: 18-multi-channel-publication-fan-out
plan: "03"
subsystem: release
tags: [github-actions, goreleaser, npm, homebrew, scoop]
requires:
  - phase: 18-01
    provides: shared downstream publication contract and rerun semantics
  - phase: 18-02
    provides: package-manager render entrypoints and thin transport wrappers
provides:
  - workflow fan-out for npm, Homebrew, and Scoop from one canonical release workflow
  - selective workflow_dispatch reruns for one downstream publication channel
  - contract tests for manual release reuse, channel gating, and package-manager publication jobs
affects: [phase-18-04, release-operator-docs, downstream-publication]
tech-stack:
  added: []
  patterns: [release job outputs canonical tag metadata for downstream jobs, thin shell wrappers delegate package-manager rendering to Go]
key-files:
  created:
    - .planning/phases/18-multi-channel-publication-fan-out/18-03-SUMMARY.md
  modified:
    - .github/workflows/release.yml
    - internal/release/release_test.go
    - internal/release/package_manager_publication_test.go
    - scripts/render-homebrew-formula.sh
    - scripts/render-scoop-manifest.sh
key-decisions:
  - "The release job now emits canonical ref, tag, version, and checksum manifest URL outputs so every downstream channel reuses one source of truth."
  - "workflow_dispatch reruns verify an existing GitHub Release with gh and skip goreleaser publication instead of rebuilding canonical assets."
  - "Homebrew and Scoop publication remain transport-only wrappers plus workflow git push steps; rendering stays in Go."
patterns-established:
  - "Canonical release reuse: downstream jobs consume release job outputs rather than recomputing tag metadata."
  - "Selective rerun gating: each publication job uses the same publication_channel contract for fanout and exact reruns."
requirements-completed: [PUB-02, PUB-03]
duration: 9m
completed: 2026-03-18
---

# Phase 18 Plan 03: Multi-Channel Publication Fan-Out Summary

**Canonical release workflow fan-out with selective npm, Homebrew, and Scoop reruns over one reused GitHub Release tag**

## Performance

- **Duration:** 9m
- **Started:** 2026-03-18T11:33:05Z
- **Completed:** 2026-03-18T11:42:25Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Added `publication_channel` workflow dispatch support with `all`, `npm`, `homebrew`, and `scoop` options and gated downstream jobs against that contract.
- Reworked the release workflow so manual reruns verify and reuse an existing canonical GitHub Release tag while skipping `goreleaser release --clean`.
- Locked Homebrew and Scoop publication details with workflow contract tests and render-wrapper regression tests.

## Task Commits

Each task was committed atomically:

1. **Task 1: Refactor the workflow for canonical-release reuse and exact downstream channel gating** - `b33d7fd` (`feat`)
2. **Task 2: Add Homebrew and Scoop publication jobs with canonical checksum reuse and per-channel summaries** - `9ca08a0` (`feat`)

## Files Created/Modified

- `.github/workflows/release.yml` - Canonical release reuse, downstream fan-out, selective rerun gating, and per-channel summaries.
- `internal/release/release_test.go` - Workflow contract coverage for rerun gating plus Homebrew and Scoop publication details.
- `internal/release/package_manager_publication_test.go` - Regression coverage for Homebrew and Scoop render helpers and shell wrappers.
- `scripts/render-homebrew-formula.sh` - Thin wrapper that delegates Homebrew formula rendering to Go.
- `scripts/render-scoop-manifest.sh` - Thin wrapper that delegates Scoop manifest rendering to Go.

## Decisions Made

- Downstream publication jobs consume release job outputs instead of repeating ref and checksum URL derivation in each job.
- Manual reruns explicitly validate the canonical GitHub Release exists with `gh release view` before downstream publication proceeds.
- Per-channel `GITHUB_STEP_SUMMARY` output carries channel, target, tag, and rerun guidance so operators can see exactly what ran.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added render-wrapper regression coverage alongside workflow fan-out**
- **Found during:** Task 2 (Add Homebrew and Scoop publication jobs with canonical checksum reuse and per-channel summaries)
- **Issue:** The workflow fan-out depended on Homebrew and Scoop render wrappers remaining executable and byte-stable, but the plan's workflow-only verification would not catch wrapper drift.
- **Fix:** Extended `internal/release/package_manager_publication_test.go` and refreshed the shell wrappers so the workflow references are backed by executable regression tests.
- **Files modified:** `internal/release/package_manager_publication_test.go`, `scripts/render-homebrew-formula.sh`, `scripts/render-scoop-manifest.sh`
- **Verification:** `go test ./internal/release -run 'Test(RenderHomebrewFormulaForTag|RenderScoopManifestForTag|RenderHomebrewFormulaScript|RenderScoopManifestScript)$'`
- **Committed in:** `9ca08a0`

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** The extra coverage was necessary to keep the new workflow runnable against the real publication entrypoints. No architectural scope change.

## Issues Encountered

- Plan 03 depended on the package-manager render surface from Plan 18-02, so verification needed both workflow tests and render-wrapper tests to prove the workflow references were valid.

## User Setup Required

None - no new external service configuration was introduced beyond the existing release tokens already referenced by the workflow.

## Next Phase Readiness

- Phase 18 now has one release workflow entrypoint for canonical publication fan-out and exact downstream reruns.
- Phase 18-04 can build on these workflow contracts for operator verification, recovery guidance, and end-to-end documentation.

## Self-Check: PASSED

- Found `.planning/phases/18-multi-channel-publication-fan-out/18-03-SUMMARY.md`
- Found task commits `b33d7fd` and `9ca08a0`
