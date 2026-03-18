---
phase: 18-multi-channel-publication-fan-out
plan: "02"
subsystem: release
tags: [go, homebrew, scoop, release-automation]
requires:
  - phase: 17-canonical-release-orchestration-and-metadata
    provides: canonical release tags, archive URLs, and checksum manifest contracts for downstream publication
provides:
  - exported Homebrew and Scoop render helpers that consume canonical release tags plus checksum manifests
  - Go CLI entrypoints for deterministic package-manager payload rendering
  - transport-only shell wrappers and regression coverage for wrapper versus helper byte stability
affects: [18-03, 18-04, 19-operator-verification-and-end-to-end-guide]
tech-stack:
  added: []
  patterns: [canonical release rendering, transport-only shell wrappers, repo-root template resolution]
key-files:
  created:
    - internal/release/package_manager_publication.go
    - internal/release/package_manager_publication_test.go
    - cmd/render-homebrew-formula/main.go
    - cmd/render-scoop-manifest/main.go
    - scripts/render-homebrew-formula.sh
    - scripts/render-scoop-manifest.sh
  modified: []
key-decisions:
  - "Homebrew and Scoop publication now render from exported Go helpers keyed by the canonical release tag plus checksum manifest content."
  - "Shell wrappers stay transport-only and delegate checksum parsing, template loading, and output rendering to Go entrypoints."
  - "Template loading resolves from the release package source path so go test and wrapper execution share one repo-root contract."
patterns-established:
  - "Package-manager publication helpers accept a release tag and checksum manifest content, normalize the tag, and render from canonical release metadata."
  - "Publication scripts pass through file paths and release tags only; metadata and template logic stays in Go."
requirements-completed: [PUB-02]
duration: 5 min
completed: 2026-03-18
---

# Phase 18 Plan 02: Package-Manager Publication Payloads and Transport Helpers Summary

**Canonical Homebrew and Scoop publication payloads now render from one Go release contract with deterministic CLI and shell entrypoints.**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-18T11:33:17Z
- **Completed:** 2026-03-18T11:38:37Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Added exported Homebrew and Scoop render helpers that normalize release tags, parse checksum manifests, and reuse the existing package-manager template contract.
- Added dedicated Go commands that read checksum manifest files and write deterministic Homebrew formula and Scoop manifest outputs.
- Added thin shell wrappers and regression tests that prove wrapper output stays byte-identical to the direct Go render helpers.

## Task Commits

Each task was committed atomically:

1. **Task 1: Export canonical Homebrew and Scoop render helpers plus Go entrypoints** - `80c1b52` (feat)
2. **Task 2: Add transport-only shell wrappers and deterministic render tests** - `28ab15a` (feat)

## Files Created/Modified

- `internal/release/package_manager_publication.go` - Exported Homebrew and Scoop render helpers plus repo-root template loading.
- `cmd/render-homebrew-formula/main.go` - CLI bridge that reads checksum manifests and writes rendered Homebrew output.
- `cmd/render-scoop-manifest/main.go` - CLI bridge that reads checksum manifests and writes rendered Scoop output.
- `scripts/render-homebrew-formula.sh` - Thin shell wrapper for workflow-driven Homebrew payload rendering.
- `scripts/render-scoop-manifest.sh` - Thin shell wrapper for workflow-driven Scoop payload rendering.
- `internal/release/package_manager_publication_test.go` - Direct-helper and wrapper regression tests for deterministic publication output.

## Decisions Made

- Exported render helpers were added in `internal/release` instead of shell templating so Homebrew and Scoop reuse the same canonical release and checksum logic as other channels.
- The shell entrypoints only handle argument and path transport; all metadata derivation remains in Go.
- Template loading resolves relative to the release package source location, which keeps `go test` and script execution on the same contract.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed template path resolution for release-package execution contexts**
- **Found during:** Task 1 verification
- **Issue:** Initial template loading depended on the process working directory, which would fail under `go test ./internal/release` and any caller not already at the repo root.
- **Fix:** Resolved publication template paths from the `internal/release` source location before reading template files.
- **Files modified:** `internal/release/package_manager_publication.go`
- **Verification:** `go test ./internal/release -run 'Test(RenderHomebrewFormulaForTag|RenderScoopManifestForTag|PackageManagerReleaseContract)$'`
- **Committed in:** `80c1b52`

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** The fix kept the planned design intact and made the new render helpers work in both test and workflow execution contexts.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 18 plan 03 can now wire workflow jobs to render Homebrew and Scoop payloads without duplicating template or checksum logic in YAML.
- The release system now has deterministic Go and shell entrypoints for downstream package-manager publication fan-out.

---
*Phase: 18-multi-channel-publication-fan-out*
*Completed: 2026-03-18*

## Self-Check: PASSED
