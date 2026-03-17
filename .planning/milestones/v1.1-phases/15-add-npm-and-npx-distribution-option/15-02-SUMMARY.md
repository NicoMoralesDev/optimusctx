---
phase: 15-add-npm-and-npx-distribution-option
plan: "02"
subsystem: distribution
tags: [npm, npx, node, cli, docs, release]
requires:
  - phase: 15-add-npm-and-npx-distribution-option
    provides: committed npm package identity, launcher path, and release-derived platform metadata from plan 15-01
provides:
  - package-local downloader and launcher for the tagged release binary
  - npm and npx install guidance rooted in the real shipped CLI verification flow
  - release-side and CLI-side guardrails for npm install commands and wrapper behavior
affects: [15-03 release publication, install guide, npm verification flow]
tech-stack:
  added: [none]
  patterns: [package-local runtime installation, checksum-verified wrapper downloads, npm/npx install-guide verification]
key-files:
  created: [packaging/npm/lib/install.js]
  modified: [packaging/npm/bin/optimusctx.js, internal/release/npm_package_test.go, docs/install-and-verify.md, internal/cli/install_test.go, internal/cli/eval_integration_test.go]
key-decisions:
  - "The npm wrapper installs binaries only inside the package-local runtime directory and never mutates MCP client configuration during package installation."
  - "Host-platform detection stays locked to the GoReleaser matrix, and the installer verifies the tagged archive against the canonical SHA-256 checksum manifest before extraction."
  - "The canonical install guide now treats npm and npx as wrapper channels while preserving the real CLI verification order of version, doctor, snippet, then explicit install preview/write."
patterns-established:
  - "Wrapper runtime acquisition: npm installs fetch the tagged archive plus checksum manifest, verify SHA-256, and unpack only the one matching host target."
  - "Guide truthfulness: npm and npx commands are documented as release wrappers, not as a source-build path or implicit MCP registration flow."
requirements-completed: [DIST-02, DIST-03]
duration: 9min
completed: 2026-03-17
---

# Phase 15 Plan 02: npm Launcher and Install Guide Summary

**Checksum-verified npm wrapper installer, package-local runtime launcher, and canonical npm/npx install-and-verify guidance**

## Performance

- **Duration:** 9 min
- **Started:** 2026-03-17T13:04:06Z
- **Completed:** 2026-03-17T13:13:28Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Added `packaging/npm/lib/install.js` to resolve the host platform, download the exact tagged release archive and checksum manifest, verify SHA-256, and unpack the binary under `runtime/<goos>-<goarch>/`.
- Upgraded the committed launcher stub so it resolves the package-local binary path and forwards CLI arguments to the real executable with inherited stdio.
- Updated the install guide and CLI-facing doc tests to include `npm install -g @niccrow/optimusctx` and `npx @niccrow/optimusctx ...` while preserving the verification order on the real `optimusctx` command surface.

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement package-local binary acquisition and launcher passthrough** - `a730511` (feat)
2. **Task 2: Add npm and npx to the canonical install-and-verify guide** - `c46b7b6` (docs)

## Files Created/Modified

- `packaging/npm/lib/install.js` - package-local downloader, checksum verification, archive extraction, and runtime target resolution
- `packaging/npm/bin/optimusctx.js` - launcher that executes the unpacked package-local binary with argument passthrough
- `internal/release/npm_package_test.go` - release-side guardrails for installer, launcher, and supported platform coverage
- `docs/install-and-verify.md` - canonical operator guide with npm global install and `npx` usage
- `internal/cli/install_test.go` - guide assertions for npm and `npx` install commands
- `internal/cli/eval_integration_test.go` - release verification guidance checks that now include npm and `npx`

## Decisions Made

- The installer remains package-local and does not write outside the npm package directory.
- The wrapper continues to treat the tagged GitHub Release binary as the runtime source of truth.
- npm and `npx` were documented without weakening the explicit MCP registration boundary.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The installer guardrail test initially failed on a literal string assertion for the `runtime/` path contract, so the installer error messaging was tightened to keep the required contract wording explicit.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan `15-03` can publish the npm package from the tagged release workflow using the committed package shape and runtime wrapper files.
- The supported install guide now has the operator-facing npm and `npx` commands that the distribution policy and README need to reference.

## Self-Check

PASSED

---
*Phase: 15-add-npm-and-npx-distribution-option*
*Completed: 2026-03-17*
