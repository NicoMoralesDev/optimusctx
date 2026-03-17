---
phase: 17-canonical-release-orchestration-and-metadata
plan: "04"
subsystem: infra
tags: [github-actions, goreleaser, release-docs, go-test]
requires:
  - phase: 16-release-versioning-and-preflight-guardrails
    provides: release tag normalization, preflight review, and selected-channel preparation inputs
  - phase: 17-canonical-release-orchestration-and-metadata
    provides: canonical release metadata plus create-versus-reuse orchestration semantics from plans 17-01 through 17-03
provides:
  - Canonical GitHub Release workflow wording and verification coverage for tagged release creation versus reuse
  - Operator documentation that names GitHub Release as the canonical root for archives, checksums, and downstream release facts
  - Regression tests that lock release_tag rerun semantics and Phase 17 scope boundaries
affects: [18-multi-channel-publication-fan-out, 19-operator-verification-recovery-and-end-to-end-guide, release-workflow, operator-docs]
tech-stack:
  added: []
  patterns: [workflow-and-doc contract tests, canonical release root wording, release_tag reuse semantics]
key-files:
  created: []
  modified:
    - .github/workflows/release.yml
    - docs/release-checklist.md
    - docs/install-and-verify.md
    - internal/release/release_test.go
key-decisions:
  - "GitHub Release stays the canonical tagged root in workflow comments, step names, and verification commands instead of being treated as one publish channel among peers."
  - "Operator docs explicitly describe workflow_dispatch release_tag as reuse of an existing tagged release contract and avoid claiming automated Homebrew or Scoop fan-out in Phase 17."
patterns-established:
  - "Workflow contracts are locked with exact-string tests in internal/release/release_test.go."
  - "Release operator docs must mirror the same canonical-root and rerun semantics as the workflow and orchestration code."
requirements-completed: [PUB-01]
duration: 8min
completed: 2026-03-17
---

# Phase 17 Plan 04: Canonical-Root Documentation And Regression Lock Summary

**Canonical GitHub Release workflow wording, release_tag reuse semantics, and operator docs/tests aligned to one tagged release root**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-17T22:55:54Z
- **Completed:** 2026-03-17T23:03:43Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Updated the release workflow to describe GitHub Release as the canonical tagged root and to verify create-versus-reuse orchestration semantics alongside existing release contract tests.
- Updated operator docs to describe GitHub Release as the canonical root for archives, checksums, and downstream release facts, including `workflow_dispatch` plus `release_tag` reuse guidance.
- Added exact-string regression tests that lock workflow wording, doc wording, and the Phase 17 boundary against premature Homebrew or Scoop automation claims.

## Task Commits

Each task was committed atomically:

1. **Task 1: Align the release workflow with the canonical root and reuse contract** - `cde2742` (fix)
2. **Task 2: Update operator docs and tests to describe the canonical release root truthfully** - `190bccd` (docs)

## Files Created/Modified

- `.github/workflows/release.yml` - Canonical-root workflow wording, release_tag reuse comments, and expanded verification coverage for orchestration tests
- `docs/release-checklist.md` - Operator checklist for canonical GitHub Release root, rerun semantics, and truthful Phase 17 channel scope
- `docs/install-and-verify.md` - Install guide wording that anchors package-manager channels to canonical tagged GitHub Release facts
- `internal/release/release_test.go` - Workflow and docs contract tests for canonical-root wording, release_tag reuse, and scope boundaries

## Decisions Made

- Kept the workflow trigger surface and job split stable while expanding comments, step names, and tests to express the canonical-root contract instead of redesigning release automation mid-phase.
- Locked workflow and docs wording with exact-string tests so later channel-fan-out work in Phase 18 cannot silently drift from the canonical tagged-release contract.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Corrected stale v1.1 milestone wording in the install guide**
- **Found during:** Task 2 (Update operator docs and tests to describe the canonical release root truthfully)
- **Issue:** `docs/install-and-verify.md` still described supported channels and scope as `v1.1`, which conflicted with the current v1.2 release-automation milestone.
- **Fix:** Updated the install guide milestone references to `v1.2` while adding the canonical-root wording required by the plan.
- **Files modified:** `docs/install-and-verify.md`
- **Verification:** `go test ./internal/release -run 'Test(ReleaseChecklistPublicationCredentials|GitHubReleaseDocsStayCanonical|GitHubReleaseWorkflowReuseContract)$'`
- **Committed in:** `190bccd` (part of task commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** The fix was directly necessary to keep the operator docs truthful. No scope creep.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 17 now ends with one test-locked canonical GitHub Release contract across workflow, docs, and release code references.
- Phase 18 can automate multi-channel publication on top of a stable release_tag reuse story and canonical tagged-release metadata root.

## Self-Check

PASSED

- Found `.planning/phases/17-canonical-release-orchestration-and-metadata/17-04-SUMMARY.md`
- Found commit `cde2742`
- Found commit `190bccd`

---
*Phase: 17-canonical-release-orchestration-and-metadata*
*Completed: 2026-03-17*
