---
phase: 13-distribution-pipeline-and-adoption-plan
plan: "03"
subsystem: testing
tags: [distribution, install, docs, smoke-test, doctor, snippet]
requires:
  - phase: 13-distribution-pipeline-and-adoption-plan
    provides: canonical release archives, checksums, and Homebrew or Scoop channel docs from plans 13-01 and 13-02
provides:
  - canonical operator-facing install and verification guidance for release archives, Homebrew, and Scoop
  - shipped-CLI smoke coverage for version, doctor, snippet, and preview-only client registration
  - README guidance that points operators at one release-oriented install and verification path
affects: [phase-13-plan-04, install-docs, release-operators, support-guides]
tech-stack:
  added: []
  patterns: [single canonical install guide, shipped-cli verification flow, preview-first mcp registration]
key-files:
  created:
    - .planning/phases/13-distribution-pipeline-and-adoption-plan/13-03-SUMMARY.md
    - docs/install-and-verify.md
  modified:
    - README.md
    - internal/cli/install_test.go
    - internal/app/snippet_test.go
    - internal/cli/doctor_test.go
    - internal/cli/eval_integration_test.go
key-decisions:
  - "The canonical operator flow lives in docs/install-and-verify.md, while README stays a pointer plus channel boundary summary so install instructions do not drift."
  - "Release verification is anchored on the shipped CLI boundary: version, init, doctor, snippet, and preview-only install registration."
  - "MCP registration remains explicit and opt-in; the smoke flow proves preview mode does not write config files unless --write is supplied."
patterns-established:
  - "Operator-facing release docs should point at package-manager or archive installs, never go run, when claiming supported installation paths."
  - "Distribution verification tests should execute the same CLI commands the docs recommend and compare snippet output against install preview output."
requirements-completed: [DIST-03]
duration: 6min
completed: 2026-03-16
---

# Phase 13 Plan 03: Install and Verify Documentation and Smoke Flow Summary

**Canonical release install guidance now walks operators through archive or package-manager installation, local verification with doctor and snippet, and explicit preview-first MCP registration backed by shipped-CLI smoke tests**

## Performance

- **Duration:** 6 min
- **Started:** 2026-03-16T16:48:00Z
- **Completed:** 2026-03-16T16:54:21Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments

- Added `docs/install-and-verify.md` as the single truthful install guide for GitHub release archives, Homebrew, and Scoop.
- Updated `README.md` to point operators at the canonical guide while preserving the supported channel and scope boundaries from earlier distribution work.
- Added named smoke coverage for `version`, `doctor`, `snippet`, and preview-only `install --client`, then proved the documented command order through the shipped CLI boundary.

## Task Commits

Each task was committed atomically:

1. **Task 1: Write the canonical install and verification guide** - `531b626` (feat)
2. **Task 2: Add archive-install and health-verification smoke coverage** - `d844c23` (test)
3. **Task 3: Prove the end-to-end distribution smoke flow through the shipped CLI boundary** - `5a3863e` (test)

## Files Created/Modified

- `docs/install-and-verify.md` - canonical operator flow for archive installs, Homebrew or Scoop installs, local verification, and explicit MCP registration
- `README.md` - points top-level install guidance at the canonical operator document and keeps release scope boundaries visible
- `internal/cli/install_test.go` - release-oriented version and guide assertions plus preview-only install smoke coverage
- `internal/app/snippet_test.go` - named snippet render verification used by the plan command surface
- `internal/cli/doctor_test.go` - CLI smoke coverage for a healthy initialized repository
- `internal/cli/eval_integration_test.go` - end-to-end release verification flow through the shipped CLI boundary and docs-order guards

## Decisions Made

- Keep the end-user install path in one dedicated document instead of spreading detailed instructions across README and release notes.
- Treat `optimusctx doctor` as the first-run health signal after initializing a disposable repository, while `optimusctx snippet` and `optimusctx install --client` explain MCP setup explicitly.
- Use preview-only config registration as the verification default and require `--write` to cross the persistence boundary.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The branch received concurrent `13-04` commits while this plan was running, so summary generation and metadata staging were scoped to the exact `13-03` files and commit hashes.
- The worktree already contained unrelated planning-file changes and untracked planning artifacts, so each task commit staged only the files owned by this plan.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan `13-04` can reference one stable operator install guide instead of repeating channel-specific setup steps.
- Distribution and rollout docs now have executable evidence for the promised install, health-check, and MCP-registration boundaries.

## Self-Check

PASSED
