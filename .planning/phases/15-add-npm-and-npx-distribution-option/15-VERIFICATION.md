---
status: passed
phase: 15
slug: add-npm-and-npx-distribution-option
verified: 2026-03-17
requirements:
  - DIST-02
  - DIST-03
  - DIST-04
---

# Phase 15 Verification: Add npm and npx distribution option

## Status

`passed`

## Scope

- Phase: `15-add-npm-and-npx-distribution-option`
- Goal: Add a truthful npm and `npx` distribution path for the existing single-binary release without changing the runtime contract or making client-configuration writes implicit.
- Requirements: `DIST-02`, `DIST-03`, `DIST-04`
- Verified against: current Phase 15 summaries, release workflow and package wrapper files, current distribution docs/policy files, and the focused Phase 15 Go test matrix

## Inputs Reviewed

- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `.planning/phases/15-add-npm-and-npx-distribution-option/15-01-SUMMARY.md`
- `.planning/phases/15-add-npm-and-npx-distribution-option/15-02-SUMMARY.md`
- `.planning/phases/15-add-npm-and-npx-distribution-option/15-03-SUMMARY.md`
- `.planning/phases/15-add-npm-and-npx-distribution-option/15-VALIDATION.md`
- `packaging/npm/package.json`
- `packaging/npm/bin/optimusctx.js`
- `packaging/npm/lib/install.js`
- `packaging/npm/lib/platform.js`
- `.github/workflows/release.yml`
- `scripts/render-npm-package.sh`
- `docs/install-and-verify.md`
- `README.md`
- `docs/distribution-strategy.md`
- `docs/release-checklist.md`

## Verification Summary

Phase 15 is verified from the implemented repository state. The phase now proves that:

- `@niccrow/optimusctx` is a committed npm wrapper package rooted in the tagged GitHub Release asset and checksum contract rather than a JavaScript reimplementation.
- npm global install and `npx` are documented as supported wrapper channels while preserving the real `optimusctx version -> doctor -> snippet -> install --client` verification path and explicit MCP registration boundary.
- the tagged release workflow now renders and publishes the npm wrapper package after GitHub Release archive publication, using `NPM_TOKEN` and a deterministic render step.
- distribution policy, README guidance, release checklist, and repository tests all agree that npm is supported while WinGet, Chocolatey, apt, dnf, yum, and other broader installer claims remain out of scope.

The focused Phase 15 verification commands passed:

```sh
go test ./internal/release ./internal/cli -run 'Test(NPMPackageMetadata|NPMPackageRendering|NPMPackageArchiveSelection|NPMPackageChecksums|NPMPackageCommands|NPMInstaller|NPMLauncher|NPMSupportedPlatforms|NPMInstallGuide|ReleaseVerificationCommands|PackageManagerInstallDocs|NPMPublishWorkflow|NPMPublishConfig|GitHubReleasePublicationConfig|DistributionChannelPolicy|DistributionDocsStayWithinSupportedScope|ArchiveMatrix|ChecksumManifest)'
bash -n scripts/render-npm-package.sh
node -c packaging/npm/lib/install.js
node -c packaging/npm/bin/optimusctx.js
node -c packaging/npm/lib/platform.js
```

## Requirement Verification

### DIST-02: User can install OptimusCtx through at least one primary package-manager path on macOS/Linux and one on Windows, aligned with the shipped single-binary runtime

Status: satisfied

Why:

- `packaging/npm/package.json`, `packaging/npm/lib/install.js`, and `packaging/npm/bin/optimusctx.js` define a wrapper package that downloads and launches the real tagged release binary for the supported GoReleaser matrix.
- `.github/workflows/release.yml` and `scripts/render-npm-package.sh` add npm publication to the tagged release lifecycle.
- `internal/release/npm_package_test.go` and `internal/release/release_test.go` lock archive naming, checksum coupling, supported platforms, launcher/install contract, and npm publication workflow details.

Evidence:

- `15-01-SUMMARY.md`
- `15-02-SUMMARY.md`
- `15-03-SUMMARY.md`
- `packaging/npm/package.json`
- `packaging/npm/lib/install.js`
- `.github/workflows/release.yml`

### DIST-03: User can follow one documented install-and-verify path that uses the real shipped command surface, including doctor and snippet, to confirm the tool works locally

Status: satisfied

Why:

- `docs/install-and-verify.md` now includes `npm install -g @niccrow/optimusctx` and `npx @niccrow/optimusctx ...` while keeping the real CLI verification order explicit.
- `internal/cli/install_test.go` and `internal/cli/eval_integration_test.go` enforce the npm guide fragments and the required command ordering.
- npm installation is explicitly documented as non-mutating for MCP config, leaving `optimusctx install --client ...` as the explicit write boundary.

Evidence:

- `15-02-SUMMARY.md`
- `docs/install-and-verify.md`
- `internal/cli/install_test.go`
- `internal/cli/eval_integration_test.go`

### DIST-04: User can understand the intended distribution strategy through a concrete plan that defines release channels, target users, upgrade path, and support assumptions for adoption

Status: satisfied

Why:

- `README.md`, `docs/distribution-strategy.md`, and `docs/release-checklist.md` now include npm alongside GitHub Releases, Homebrew, and Scoop.
- `internal/release/distribution_plan.go` and `internal/release/distribution_plan_test.go` encode npm as a supported wrapper channel while continuing to forbid unsupported installer claims.
- The docs preserve GitHub Release archives as the rollback fallback and keep support boundaries best-effort and issue-driven.

Evidence:

- `15-03-SUMMARY.md`
- `README.md`
- `docs/distribution-strategy.md`
- `docs/release-checklist.md`
- `internal/release/distribution_plan.go`
- `internal/release/distribution_plan_test.go`

## Phase Goal Verification

Phase 15 goal: add a truthful npm and `npx` distribution path without changing the runtime contract or making client-configuration writes implicit.

Result: satisfied

Why:

- The npm package is explicitly a wrapper over tagged GitHub Release binaries and checksum manifests.
- The implementation leaves the runtime shape single-binary and keeps MCP config writes opt-in.
- The supported-channel narrative and release automation were extended without broadening into unrelated installer ecosystems.

This conclusion is an inference from the repository implementation and focused tests. A clean-machine `npm install -g` / `npx` smoke remains useful operational validation, but it is not required to conclude that the phase goal is implemented in the codebase.

## Success Criteria Verification

### npm and `npx` are added as truthful wrapper channels over the tagged binary

Satisfied. The package, installer, launcher, docs, and release workflow all point to tagged GitHub Release artifacts rather than a second runtime.

### Verification remains on the real shipped CLI surface

Satisfied. The install guide and CLI tests require `optimusctx version`, `optimusctx doctor`, `optimusctx snippet`, and the explicit `install --client` path.

### Broader installer scope remains constrained

Satisfied. The policy docs and distribution tests continue to reject WinGet, Chocolatey, apt, dnf, yum, and other unsupported package-manager claims.

## Residual Risk

- The repository verification did not execute a real `npm publish`, `npm install -g`, or `npx` smoke against the public registry. Those remain recommended release-operator checks after credentials and a clean-machine environment are available.

## Final Verdict

Phase 15 is verified as `passed`.
