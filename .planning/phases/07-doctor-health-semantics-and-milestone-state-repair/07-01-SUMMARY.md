---
phase: 07-doctor-health-semantics-and-milestone-state-repair
plan: "01"
subsystem: cli
tags: [doctor, watch, diagnostics, go]
requires:
  - phase: 06-watch-mode-pack-export-and-operational-diagnostics
    provides: doctor aggregation, watch status semantics, and operator diagnostics
provides:
  - doctor aggregation that treats absent watch as optional while stale watch remains degraded
  - typed watch summary metadata for doctor reporting and CLI rendering
  - regression coverage for healthy no-watch and stale-watch doctor flows
affects: [doctor, watch, operator-diagnostics, milestone-state]
tech-stack:
  added: []
  patterns: [doctor section severity stays separate from raw watch-state payloads]
key-files:
  created: []
  modified:
    - internal/app/doctor.go
    - internal/repository/doctor.go
    - internal/cli/doctor.go
    - internal/app/doctor_test.go
    - internal/cli/doctor_test.go
key-decisions:
  - "Doctor now treats absent watch as a healthy optional state while preserving raw watch status as absent for operator visibility."
  - "CLI wording translates absent watch into optional-background-watch guidance instead of implying repository failure."
patterns-established:
  - "Doctor section status can remain healthy even when the underlying subsystem exposes an optional absent state."
  - "Operator-facing CLI text should reinterpret raw diagnostic payloads when the raw reason is technically correct but misleading."
requirements-completed: [CLI-05, OPS-01, OPS-05]
duration: 12min
completed: 2026-03-15
---

# Phase 07 Plan 01: Doctor Health Semantics and Milestone State Repair Summary

**Doctor now keeps initialized repositories healthy when watch is absent, while preserving degraded stale-watch reporting and explicit optional-watch CLI messaging**

## Performance

- **Duration:** 12 min
- **Started:** 2026-03-15T19:32:20Z
- **Completed:** 2026-03-15T19:44:54Z
- **Tasks:** 3
- **Files modified:** 5

## Accomplishments
- Reclassified absent watch as a healthy optional doctor state without changing the underlying watch-service status contract.
- Added typed watch summary metadata so CLI rendering can distinguish optional absence from degraded stale or dead watch behavior.
- Locked the regression with app and CLI coverage plus a temp-repository smoke run for `init -> refresh -> doctor` with no watch process.

## Task Commits

Each task was committed atomically:

1. **Task 1: Reclassify absent watch as non-failing in doctor aggregation** - `a9ff2ef` (fix)
2. **Task 2: Add regression coverage for healthy init-refresh-doctor flow without watch** - `7beff26` (test)
3. **Task 3: Tighten operator wording so absent watch reads as optional rather than broken** - `7ec0d8d` (fix)

## Files Created/Modified
- `internal/app/doctor.go` - Keeps doctor healthy when watch is absent and emits typed watch summary metadata.
- `internal/repository/doctor.go` - Extends the doctor watch section contract with optionality and summary fields.
- `internal/cli/doctor.go` - Renders optional absent-watch wording and operator-facing watch reasons.
- `internal/app/doctor_test.go` - Covers healthy-no-watch and stale-watch doctor semantics.
- `internal/cli/doctor_test.go` - Covers human-facing doctor output for optional absent watch and degraded stale watch.

## Decisions Made
- Treat `WatchStatusKindAbsent` as healthy at the doctor-report boundary because watch mode is optional in v1 and should not downgrade repository health.
- Preserve the raw watch payload as `absent` so the CLI can stay explicit about watch inactivity without rewriting watch-state detection.
- Translate the absent-watch reason at the CLI boundary so operators see optional guidance instead of a misleading missing-file failure.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Default `go test` invocation in this environment misclassified the standard `testing` package; verification succeeded with the established isolated Go cache settings: `GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Doctor semantics and CLI messaging now match the intended optional-watch v1 contract.
- Phase 07 can build on these repaired health semantics without touching watch daemon production logic.

## Verification

- `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./internal/app ./internal/cli`
- `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./... -run 'TestDoctorHealthyWithoutWatch|TestDoctorCommandHealthyWithoutWatch|TestDoctorDetectsStaleWatch|TestDoctorReportSections|TestDoctorCommand'`
- `env GOCACHE=/tmp/optimusctx-gocache GOMODCACHE=/tmp/optimusctx-gomodcache GOPROXY=off /usr/local/go/bin/go test ./...`
- Temp repository smoke flow: `optimusctx init`, `optimusctx refresh`, `optimusctx doctor` reported `overall status: healthy` with `watch state: absent`.

## Self-Check: PASSED

- Found `.planning/phases/07-doctor-health-semantics-and-milestone-state-repair/07-01-SUMMARY.md`.
- Verified task commits `a9ff2ef`, `7beff26`, and `7ec0d8d` in git history.

---
*Phase: 07-doctor-health-semantics-and-milestone-state-repair*
*Completed: 2026-03-15*
