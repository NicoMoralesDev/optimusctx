---
phase: 15-add-npm-and-npx-distribution-option
plan: "01"
subsystem: distribution
tags: [go, npm, npx, release, packaging, node]
requires:
  - phase: 13-distribution-pipeline-and-adoption-plan
    provides: canonical GitHub Release archive, checksum, Homebrew, and Scoop publication contract reused by the npm wrapper metadata
provides:
  - deterministic npm package metadata rooted in the canonical GitHub Release asset matrix
  - committed npm package manifest and platform mapping for darwin, linux, and windows on amd64 and arm64
  - release-side tests that pin package naming, bin command, archive filenames, and checksum-manifest coupling
affects: [15-02 launcher downloader flow, 15-03 npm publication workflow, distribution docs]
tech-stack:
  added: [none]
  patterns: [release-derived npm wrapper metadata, explicit host-platform mapping, committed package manifest parity tests]
key-files:
  created: [internal/release/npm_package.go, internal/release/npm_package_test.go, packaging/npm/package.json, packaging/npm/lib/platform.js, packaging/npm/bin/optimusctx.js]
  modified: [internal/release/release_test.go]
key-decisions:
  - "The committed npm package manifest stays on a development placeholder version while retaining the full tagged-release metadata shape so later publication can inject the exact tag without inventing a second contract."
  - "npm package metadata derives archive names, checksum manifest names, runtime directories, and binary names from the existing GoReleaser contract instead of maintaining a separate npm-only asset matrix."
  - "Host-platform detection is explicit and limited to darwin, linux, and windows on amd64 and arm64 so unsupported npm hosts fail loudly instead of silently drifting outside the shipped Go release set."
patterns-established:
  - "Package manifest parity: the committed packaging/npm/package.json must match the release-side renderer output exactly."
  - "Wrapper truthfulness: npm metadata can expose launcher and postinstall hooks, but the canonical release source remains the tagged GitHub Release archives and checksums."
requirements-completed: [DIST-02]
duration: 2min
completed: 2026-03-17
---

# Phase 15 Plan 01: npm Package Foundation Summary

**Deterministic npm package metadata, platform mapping, and release-contract tests for the `@niccrow/optimusctx` wrapper channel**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-17T13:02:24Z
- **Completed:** 2026-03-17T13:04:06Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Added release-side npm package metadata helpers that derive package name, launcher contract, archive filenames, checksum manifest URLs, and platform runtime paths from the existing GoReleaser release contract.
- Committed the initial `packaging/npm/` foundation with `package.json`, explicit host-platform mapping, and a Node launcher stub at the future `optimusctx` bin path.
- Locked the package contract with focused tests covering manifest rendering parity, archive selection, checksum-manifest coupling, exact package/bin naming, and `.goreleaser.yml` scope boundaries.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add the deterministic npm package metadata contract** - `13aa2ee` (feat)
2. **Task 2: Lock the checksum and command invariants with release tests** - `e13d540` (test)

## Files Created/Modified

- `internal/release/npm_package.go` - npm package release metadata model, manifest renderer, checksum manifest naming, and per-platform asset selection helpers
- `internal/release/npm_package_test.go` - focused rendering, metadata, archive-selection, checksum, and command invariants for the npm wrapper contract
- `packaging/npm/package.json` - committed npm package manifest for `@niccrow/optimusctx` with the canonical `optimusctx` bin and postinstall hook
- `packaging/npm/lib/platform.js` - explicit mapping from Node host platforms to the supported GoReleaser OS and architecture matrix
- `packaging/npm/bin/optimusctx.js` - launcher stub with Node shebang and runtime-path resolution for the real binary
- `internal/release/release_test.go` - release-scope guardrails that keep npm launcher logic out of `.goreleaser.yml`

## Decisions Made

- Kept the committed package manifest on `0.0.0-development` while preserving the full release-derived metadata shape so CI can later inject a tagged version deterministically.
- Reused the canonical archive/checksum naming helpers instead of introducing npm-specific filename rules.
- Limited supported npm hosts to the shipped Go matrix and surfaced unsupported platforms as explicit errors.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Go test inside the sandbox could not write to the default build cache, so focused verification was rerun with escalated permissions to allow the normal Go build-cache path.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan `15-02` can implement the real package-local downloader and launcher against the committed `optimusctx` metadata contract and explicit platform map.
- Plan `15-03` can render and publish the npm package in CI using the same release-derived package structure without redefining package identity or asset naming.

## Self-Check

PASSED

---
*Phase: 15-add-npm-and-npx-distribution-option*
*Completed: 2026-03-17*
