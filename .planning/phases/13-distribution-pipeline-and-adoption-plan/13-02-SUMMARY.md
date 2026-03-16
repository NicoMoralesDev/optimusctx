---
phase: 13-distribution-pipeline-and-adoption-plan
plan: "02"
subsystem: infra
tags: [goreleaser, homebrew, scoop, distribution, docs]
requires:
  - phase: 13-distribution-pipeline-and-adoption-plan
    provides: canonical release archives, checksums, and GitHub Release publication from plan 13-01
provides:
  - deterministic Homebrew formula rendering from release archive metadata and checksums
  - deterministic Scoop manifest rendering from the same release source of truth
  - truthful v1.1 package-manager docs limited to Homebrew and Scoop with explicit publication targets
affects: [phase-13-plan-03, install-docs, release-operators, package-manager-manifests]
tech-stack:
  added: []
  patterns: [release-derived package-manager manifests, explicit supported-channel boundaries, doc assertions in CLI tests]
key-files:
  created:
    - .planning/phases/13-distribution-pipeline-and-adoption-plan/13-02-SUMMARY.md
  modified:
    - README.md
    - internal/cli/install_test.go
    - internal/release/package_manager.go
    - internal/release/package_manager_test.go
    - packaging/homebrew/optimusctx.rb.tmpl
    - packaging/scoop/optimusctx.json.tmpl
key-decisions:
  - "Homebrew publishes through niccrow/homebrew-tap while user-facing install docs stay on the canonical Homebrew tap name niccrow/tap/optimusctx."
  - "Scoop publishes through niccrow/scoop-bucket and requires explicit bucket registration before install so the Windows path stays truthful about its first-channel boundary."
  - "v1.1 package-manager claims stop at Homebrew and Scoop; native Linux packages, WinGet, Chocolatey, signing, and SBOMs remain deferred."
patterns-established:
  - "Package-manager publication targets are asserted in release tests so repo names and token env vars cannot drift from docs silently."
  - "README package-manager guidance is tested from the CLI package to keep install commands and deferred-scope wording aligned with shipped behavior."
requirements-completed: [DIST-02]
duration: 24min
completed: 2026-03-16
---

# Phase 13 Plan 02: Primary Package-Manager Distribution Paths Summary

**Homebrew and Scoop package-manager channels now derive from the canonical release archives and are documented as the only supported v1.1 install paths beyond GitHub-hosted binaries**

## Performance

- **Duration:** 24 min
- **Started:** 2026-03-16T16:17:33Z
- **Completed:** 2026-03-16T16:41:09Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments

- Added shared Go rendering logic for Homebrew and Scoop manifests that consumes the release archive names and checksum manifest produced by plan `13-01`.
- Defined the first consumable publication targets for package-manager distribution: `niccrow/homebrew-tap` for Homebrew and `niccrow/scoop-bucket` for Scoop.
- Updated the README and test suite so v1.1 now claims exactly two package-manager channels, publishes the real install commands, and explicitly defers broader package-manager and signing scope.

## Task Commits

Each task was committed atomically:

1. **Task 1: Render Homebrew metadata from release artifacts** - `ceb5c6b` (feat)
2. **Task 2: Render Scoop metadata from the same release source** - `5ba1447` (feat)
3. **Task 3: Document the supported package-manager channels and their boundaries** - `032b0b9` (docs)

## Files Created/Modified

- `README.md` - adds the v1.1 package-manager channel contract, install commands, publication targets, and deferred-scope boundaries
- `internal/cli/install_test.go` - asserts the README package-manager guidance stays truthful about supported channels and deferrals
- `internal/release/package_manager.go` - central release-derived package-manager metadata and publication target defaults
- `internal/release/package_manager_test.go` - deterministic rendering coverage plus publication-target assertions for Homebrew and Scoop
- `packaging/homebrew/optimusctx.rb.tmpl` - Homebrew formula template with scoped channel metadata for macOS and Linux
- `packaging/scoop/optimusctx.json.tmpl` - Scoop manifest template with scoped channel metadata for Windows

## Decisions Made

- `niccrow/homebrew-tap` is the publication repository, but user-facing Homebrew install guidance remains `brew install niccrow/tap/optimusctx` because Homebrew tap naming strips the `homebrew-` prefix.
- `niccrow/scoop-bucket` is the publication repository and Windows users are expected to add that bucket explicitly before installing `niccrow/optimusctx`.
- Package-manager documentation now names `HOMEBREW_TAP_GITHUB_TOKEN` and `SCOOP_BUCKET_GITHUB_TOKEN` as the publication credentials while keeping end-user install guidance limited to Homebrew and Scoop only.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Added the Task 3 verification tests that the plan already required**
- **Found during:** Task 3 (Document the supported package-manager channels and their boundaries)
- **Issue:** `TestPackageManagerPublicationConfig` and `TestPackageManagerInstallDocs` were referenced by the plan verification command but did not exist yet.
- **Fix:** Added a release-layer test for the Homebrew and Scoop publication targets plus a CLI-layer test that reads `README.md` and asserts the supported channel text, install commands, token names, and deferred-scope wording.
- **Files modified:** `internal/release/package_manager_test.go`, `internal/cli/install_test.go`
- **Verification:** `go test ./internal/release ./internal/cli -run 'TestPackageManagerPublicationConfig|TestPackageManagerInstallDocs'`
- **Committed in:** `032b0b9` (Task 3 commit)

---

**Total deviations:** 1 auto-fixed (Rule 2)
**Impact on plan:** The added tests were necessary to satisfy the plan's own verification contract without broadening runtime scope.

## Issues Encountered

- Sandboxed Go test runs still could not use the default Go cache location, so verification used `GOCACHE=/tmp/optimusctx-gocache` and `GOMODCACHE=/tmp/optimusctx-gomodcache`.
- The worktree already contained unrelated planning-file changes and untracked planning artifacts; Task 3 staging was limited to the package-manager docs and test files so that user work remained untouched.

## User Setup Required

- None for end users beyond the documented Homebrew or Scoop commands.
- Release operators need GitHub credentials for `niccrow/homebrew-tap` and `niccrow/scoop-bucket`, exposed to automation as `HOMEBREW_TAP_GITHUB_TOKEN` and `SCOOP_BUCKET_GITHUB_TOKEN`.

## Next Phase Readiness

- Plan `13-03` can now build install-and-verify guidance on top of three truthful distribution surfaces: GitHub release archives, Homebrew, and Scoop.
- The package-manager channel contract is narrow and test-backed, so later docs can reference it without claiming `.deb`, `.rpm`, WinGet, Chocolatey, signing, or SBOM work prematurely.

## Self-Check

PASSED
