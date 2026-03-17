---
phase: 13-distribution-pipeline-and-adoption-plan
plan: "01"
subsystem: infra
tags: [goreleaser, github-actions, release, cli, distribution]
requires:
  - phase: 12-token-attribution-and-evidence-reporting
    provides: milestone verification discipline, shipped CLI command surface, and truthful operator-facing documentation patterns
provides:
  - deterministic cross-platform archive and checksum definitions for the shipped optimusctx binary
  - tag-driven and manual GitHub Release publication through one canonical release workflow
  - ldflags-backed build metadata surfaced through `optimusctx version`
affects: [phase-13-plan-02, phase-13-plan-03, package-manager-manifests, install-docs]
tech-stack:
  added: [GoReleaser, GitHub Actions]
  patterns: [single canonical release contract, ldflags-backed runtime metadata, contract tests over release config files]
key-files:
  created: [.planning/phases/13-distribution-pipeline-and-adoption-plan/13-01-SUMMARY.md]
  modified:
    - .goreleaser.yml
    - .github/workflows/release.yml
    - internal/buildinfo/buildinfo.go
    - internal/buildinfo/buildinfo_test.go
    - internal/cli/version.go
    - internal/cli/version_test.go
    - internal/release/release_test.go
    - README.md
key-decisions:
  - "GoReleaser is the only release source of truth, and the GitHub Actions workflow delegates archive, checksum, and ldflags behavior to that file instead of duplicating platform logic."
  - "GitHub Releases is the first retrievable distribution channel, with manual dispatch restricted to existing `v*` tags so republished artifacts stay tied to tagged source."
  - "Release metadata is exposed through `buildinfo.Current()` and verified through `optimusctx version` so shipped binaries report truthful version, commit, and build date values."
patterns-established:
  - "Release workflow verification reads `.goreleaser.yml` and workflow files directly so archive naming and publication assumptions stay test-backed."
  - "Operator-facing docs only claim archive, checksum, and metadata behavior that the automated workflow and CLI tests already enforce."
requirements-completed: [DIST-01]
duration: 24min
completed: 2026-03-16
---

# Phase 13 Plan 01: Automated Release Archives and Checksums Summary

**GoReleaser-driven GitHub release archives now produce deterministic cross-platform asset names, SHA-256 checksums, and ldflags-backed version metadata for the shipped `optimusctx` binary**

## Performance

- **Duration:** 24 min
- **Started:** 2026-03-16T15:43:30Z
- **Completed:** 2026-03-16T16:07:38Z
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments

- Added one canonical `.goreleaser.yml` definition for macOS, Linux, and Windows archives, checksum generation, and build metadata injection.
- Added a tag-driven and manual GitHub Actions workflow that validates the release contract and publishes GitHub Release assets through GoReleaser.
- Documented the real release operator path in the README and locked the archive, checksum, workflow, and version metadata assumptions with targeted tests.

## Task Commits

Each task was committed atomically:

1. **Task 1: Lock the release metadata and archive contract** - `5fb937a` (feat)
2. **Task 2: Add automated release workflow execution around the canonical contract** - `cb9a84c` (feat)
3. **Task 3: Document the release operator path and artifact expectations** - `c87fa35` (docs)

## Files Created/Modified

- `.goreleaser.yml` - canonical archive matrix, checksum manifest, GitHub release target, and ldflags injection contract
- `.github/workflows/release.yml` - tag-driven and manual release workflow that validates then publishes GitHub Release assets
- `internal/buildinfo/buildinfo.go` - structured runtime build metadata surface used by the shipped version command
- `internal/buildinfo/buildinfo_test.go` - build metadata summary coverage
- `internal/cli/version.go` - version command wired through the structured build metadata surface
- `internal/cli/version_test.go` - command coverage for version output and release metadata injection alignment
- `internal/release/release_test.go` - deterministic tests for archive matrix, checksum manifest, and workflow publication contract
- `README.md` - truthful operator guidance for the release entrypoint, GitHub Release artifacts, and metadata expectations

## Decisions Made

- `.goreleaser.yml` is the canonical release contract for archives, checksums, and ldflags so later package-manager work can derive from one source of truth.
- Manual release workflow runs require an existing `v*` tag instead of inventing an alternate non-tag release path.
- Release docs explicitly stop at GitHub-hosted archives and checksums; package managers, signing, and SBOMs remain future work.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Sandboxed `go test` runs could not write to the default Go build cache under `/home/nico/.cache`, so verification was rerun with `GOCACHE` and `GOMODCACHE` redirected into `/tmp`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 13 now has one deterministic archive and checksum contract that package-manager manifests can consume without copying platform logic.
- Plan `13-02` can build Homebrew and Scoop publication paths on top of the existing GoReleaser artifact names and checksum output.

## Self-Check

PASSED
