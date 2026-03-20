---
phase: 22-documentation-and-compatibility-verification
plan: "02"
subsystem: docs
tags: [docs, release, policy, tests]
requires:
  - phase: 22-01
    provides: public docs aligned to the corrected init-led onboarding contract
provides:
  - release/operator docs aligned to init-led onboarding
  - distribution policy constants that encode the corrected command surface
  - release doc-lock tests that guard against stale status-led guidance
affects:
  - release-readiness
  - operator-guidance
tech-stack:
  added: []
  patterns:
    - release docs and doc-lock tests evolve together
key-files:
  created: []
  modified:
    - docs/distribution-strategy.md
    - docs/operator-release-guide.md
    - docs/release-checklist.md
    - internal/release/distribution_plan.go
    - internal/release/distribution_plan_test.go
key-decisions:
  - Release/operator docs now distinguish read-only status checks from explicit init-led onboarding writes.
  - Distribution policy support commands now include `optimusctx init` instead of the removed `status --client` onboarding path.
requirements-completed: [DOC-01, TST-01]
duration: 9min
completed: 2026-03-20
---

# Phase 22 Plan 02: Supported-client regression coverage and idempotence verification Summary

**Aligned the release-facing docs and release policy tests with the corrected init-led onboarding contract so the docs and the codebase now fail together if the old status-led story returns**

## Accomplishments

- Updated distribution strategy, release checklist, and operator release guide text to use init-led onboarding and read-only status semantics.
- Adjusted the release verification flow to use a disposable repository before running init-led onboarding commands from an unpacked binary.
- Updated `internal/release/distribution_plan.go` and its tests so the release/support policy encodes the corrected command surface.

## Task Commits

1. **Task 1: Update release/operator docs for init-led onboarding** - `9301257` (`docs`)
2. **Task 2: Lock release policy and docs against stale status-led guidance** - `9301257` (`docs`)

## Verification

- `go test ./internal/release -run 'Test(DistributionChannelPolicy|RolloutPlanExamples|UpgradePolicy|OperatorRecoveryGuideStaysCanonical|DistributionDocsStayWithinSupportedScope)'` passed.

## Self-Check: PASSED

- Release docs no longer advertise `status --client ...` as the onboarding owner.
- Release-policy tests now lock the corrected init-led story.
