---
phase: 17-canonical-release-orchestration-and-metadata
plan: "07"
subsystem: infra
tags: [go-release, npm, github-release, shell-wrapper]
requires:
  - phase: 17-canonical-release-orchestration-and-metadata
    provides: canonical release metadata and downstream consumer wiring from plans 17-01 through 17-06
provides:
  - Canonical npm package manifest rendering for a requested release tag from Go release helpers
  - Transport-only npm render wrapper that copies package assets and delegates manifest generation to the Go bridge
  - Contract tests that compare the rendered npm package manifest byte-for-byte against canonical release metadata
affects: [18-multi-channel-publication-fan-out, 19-operator-verification-recovery-and-end-to-end-guide, npm-publication, release-testing]
tech-stack:
  added: []
  patterns: [go-backed manifest rendering, transport-only shell wrappers, canonical manifest byte-for-byte regression checks]
key-files:
  created:
    - .planning/phases/17-canonical-release-orchestration-and-metadata/17-07-SUMMARY.md
  modified:
    - internal/release/canonical_release.go
    - internal/release/npm_package.go
    - internal/release/release_test.go
    - scripts/render-npm-package.sh
key-decisions:
  - "The npm wrapper must render package.json from Go canonical release helpers instead of retagging committed JSON in shell or Node."
  - "Downstream regression coverage should compare the rendered package manifest byte-for-byte against RenderNPMPackageManifestForTag for the canonical tag."
patterns-established:
  - "Shell release wrappers should stay transport-only: copy package assets, preserve executable bits, and delegate metadata assembly to Go helpers."
  - "Release-channel contract tests should lock exact rendered artifacts instead of reconstructing canonical URLs or archive names locally."
requirements-completed: [PUB-01]
duration: 8min
completed: 2026-03-17
---

# Phase 17 Plan 07: Npm Render Canonical-Source Closure Summary

**Canonical npm manifest rendering now comes directly from Go release helpers, with the shell wrapper reduced to transport-only copy plus delegation and byte-for-byte downstream regression coverage**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-17T20:45:45-03:00
- **Completed:** 2026-03-17T23:54:59Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Exported `RenderNPMPackageManifestForTag` from `internal/release` so npm package metadata is rendered from normalized canonical release tags instead of retagged committed JSON.
- Simplified `scripts/render-npm-package.sh` into a transport-only wrapper that copies `packaging/npm`, preserves the launcher executable bit, and delegates manifest generation to `go run ./cmd/render-npm-package`.
- Locked the downstream contract with tests that assert the rendered `package.json` matches `RenderNPMPackageManifestForTag(canonicalRelease.Tag)` byte-for-byte and that script-local retag helpers are absent.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add a Go-backed canonical npm manifest renderer for release tags** - `81282ce`, `5424afc` (feat, fix)
2. **Task 2: Replace script-local retag logic with the canonical Go renderer and lock the new contract** - `954ae7b` (feat)

## Files Created/Modified

- `.planning/phases/17-canonical-release-orchestration-and-metadata/17-07-SUMMARY.md` - Execution summary, decisions, deviations, and self-check evidence for plan 17-07
- `internal/release/canonical_release.go` - Shared canonical target and synthetic release construction used by tagged and development manifest rendering
- `internal/release/npm_package.go` - Exported release-tag manifest renderer that normalizes tags and renders from canonical Go release metadata
- `internal/release/release_test.go` - Transport-only wrapper assertions and byte-for-byte downstream manifest regression coverage
- `scripts/render-npm-package.sh` - Copy-and-delegate wrapper that invokes the Go manifest renderer directly

## Decisions Made

- Kept canonical npm manifest facts in Go so release tag normalization, checksum URLs, and archive URLs are assembled once from `CanonicalRelease`.
- Switched the end-to-end regression to exact manifest equality so future channel work cannot silently reintroduce local URL rewriting or archive reconstruction in the shell layer.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Align synthetic canonical release construction with the shared target inventory**
- **Found during:** Task 1 (Add a Go-backed canonical npm manifest renderer for release tags)
- **Issue:** `syntheticCanonicalRelease` still built assets through an older inline path, which conflicted with the newer canonical target inventory helpers already present in `internal/release/canonical_release.go`.
- **Fix:** Routed synthetic release construction through `newCanonicalRelease(version)` so tagged and development npm manifest rendering use the same asset and checksum contract.
- **Files modified:** `internal/release/npm_package.go`, `internal/release/canonical_release.go`
- **Verification:** `go test ./internal/release -run 'Test(NPMPackageReleaseContract|RenderCommittedNPMPackageManifest|RenderNPMPackageManifestForTag)$'`
- **Committed in:** `5424afc` (follow-up correction after the initial Task 1 feature commit `81282ce`)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Required for correctness. No scope creep; it unified an already-evolved canonical release helper surface with the new npm manifest entrypoint.

## Issues Encountered

- Task 1 overlap already existed in `HEAD`: `cmd/render-npm-package/main.go`, the exported renderer entrypoint, and its focused manifest test had landed earlier in commit `8e54faa` while plan 17-06 was executed. This run kept the plan atomic by committing the remaining shared-helper fix in Task 1 and completing the shell-wrapper and regression-lock work in Task 2.
- The end-to-end `bash scripts/render-npm-package.sh v1.2.3 /tmp/optimusctx-npm-render-gap-check` verification needed one escalated rerun because `go run` inside the wrapper could not write to the sandboxed default Go build cache.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The npm publication path now consumes the same canonical release metadata contract as GitHub Release, Homebrew, and Scoop.
- Phase 17 can return to verification with the npm render truth closed; remaining planning metadata still needs the missing 17-06 summary reconciled separately.

## Self-Check

PASSED

- Found `.planning/phases/17-canonical-release-orchestration-and-metadata/17-07-SUMMARY.md`
- Found commit `81282ce`
- Found commit `5424afc`
- Found commit `954ae7b`

---
*Phase: 17-canonical-release-orchestration-and-metadata*
*Completed: 2026-03-17*
