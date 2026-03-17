---
phase: 15-add-npm-and-npx-distribution-option
research_date: 2026-03-17
objective: "What do I need to know to PLAN this phase well?"
status: complete
---

# Phase 15 Research: Add npm and npx distribution option

## Executive Summary

Phase 15 should add npm and `npx` as a new distribution channel without changing the product shape established in Phase 13. OptimusCtx is still a shipped single-binary Go CLI. The narrowest truthful npm story is therefore:

1. publish an npm package that exposes the `optimusctx` command through the package `bin` field
2. resolve that command to the real tagged GitHub Release binary for the host platform
3. verify the downloaded binary against the canonical checksum manifest
4. keep install verification on the real CLI surface: `version`, `doctor`, `snippet`, and the explicit `install --client` preview/write flow

The repository already has a deterministic release source of truth, archive naming contract, checksum manifest, package-manager renderers, and distribution-policy tests. What it does not have is npm package metadata, a launcher/downloader flow, npm publication automation, or docs/policy that permit npm/`npx`. Phase 15 therefore needs to extend the existing release-and-distribution layer, not reopen runtime scope.

## Repository Reality Check

### Inputs reviewed

- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/REQUIREMENTS.md`
- `README.md`
- `docs/install-and-verify.md`
- `docs/distribution-strategy.md`
- `docs/release-checklist.md`
- `.goreleaser.yml`
- `.github/workflows/release.yml`
- `internal/release/package_manager.go`
- `internal/release/package_manager_test.go`
- `internal/release/release_test.go`
- `internal/release/distribution_plan_test.go`
- `internal/cli/install_test.go`
- `internal/cli/eval_integration_test.go`
- `.planning/phases/13-distribution-pipeline-and-adoption-plan/13-RESEARCH.md`
- `.planning/phases/13-distribution-pipeline-and-adoption-plan/13-VALIDATION.md`
- `.planning/phases/13-distribution-pipeline-and-adoption-plan/13-02-PLAN.md`

### External primary sources consulted

- Official npm docs for `package.json` executable metadata (`bin`)
- Official npm docs for `npm exec` / `npx` command resolution
- Official npm docs for package lifecycle scripts (`postinstall`, `install`)

### What already exists

- `.goreleaser.yml` defines one canonical archive/checksum contract for darwin/linux/windows on amd64/arm64.
- `.github/workflows/release.yml` publishes tagged GitHub Release assets from that GoReleaser contract.
- `internal/release/package_manager.go` and `internal/release/package_manager_test.go` already render Homebrew and Scoop metadata from the same archive/checksum inputs.
- `README.md`, `docs/install-and-verify.md`, and `docs/distribution-strategy.md` define the current operator story and support boundary.
- Multiple tests currently enforce the old scope and explicitly forbid npm/`npx` references.

### What is explicitly missing today

- No npm package metadata or publishable package directory
- No launcher or downloader that resolves npm/`npx` execution to the tagged binary
- No npm publication step or `NPM_TOKEN`-based automation
- No docs claiming npm or `npx` as supported channels

## Official npm Constraints and Planning Implications

## 1. `bin` is the right integration point

Official npm docs define `bin` as the package metadata used to expose executable commands. That matches this phase cleanly: the npm package should expose the command name `optimusctx`, while the actual runtime remains the Go binary.

Planning implication: the package should publish a single bin entry:

- package name: `@niccrow/optimusctx`
- bin command: `optimusctx`
- bin entry file: `bin/optimusctx.js`

This keeps `npm install -g @niccrow/optimusctx` and `npx @niccrow/optimusctx ...` aligned on the same command surface.

## 2. `npx` / `npm exec` resolve the package bin

Official npm docs for `npm exec` / `npx` resolve the package executable from the package `bin` field. For a scoped package whose unscoped name is `optimusctx`, a single matching `bin` entry keeps the user-facing `npx` story simple and predictable.

Planning implication: avoid inventing a second command or wrapper name. The npm package should still run `optimusctx`, not `optimusctx-npm` or a hidden helper.

## 3. Lifecycle scripts can fetch platform artifacts, but should stay narrow

Official npm lifecycle docs confirm that install-time scripts run during package installation. That makes a binary fetch step feasible, but it should be used narrowly:

- fetch only the tagged archive that matches the package version and current OS/arch
- verify the SHA-256 from the canonical release checksum manifest
- unpack into a package-local runtime directory
- never write MCP config files or mutate unrelated user state

Planning implication: if a `postinstall` script is used, it must stay deterministic, idempotent, and scoped to binary acquisition only.

## Recommended Technical Direction

## 1. Treat npm as a distribution wrapper over GitHub Releases

The npm package should not become a second build system or a JavaScript reimplementation. It should wrap the already shipped release assets:

- package version tracks the release version
- download URLs point to the exact GitHub Release archive names already produced by GoReleaser
- checksum verification uses the same `optimusctx_<version>_checksums.txt`
- archive fallback remains GitHub Releases

This preserves the single-binary runtime contract and keeps Phase 15 aligned with the Phase 13 release architecture.

## 2. Keep the installed binary package-local and explicit

The downloader/launcher should place the binary under the npm package directory, for example:

- `packaging/npm/runtime/<os>-<arch>/optimusctx`
- `packaging/npm/runtime/<os>-<arch>/optimusctx.exe` on Windows

The launcher script should resolve the host platform, ensure the package-local binary exists, and `spawn` the real executable with passthrough args/stdin/stdout/stderr. It should not install to `/usr/local/bin`, write shell profiles, or perform MCP client registration.

## 3. Extend the release pipeline with one npm publication step

Phase 13 already publishes archives/checksums. Phase 15 should add one npm publication job after those assets are available:

- set up Node in CI
- render a publishable npm package directory from canonical release metadata
- run `npm publish` with `NPM_TOKEN`
- keep the tagged GitHub Release as the source of truth and fallback channel

The phase should prefer extending the existing release workflow over building a disconnected second publication path.

## 4. Update docs and policy truthfully

The docs must stop saying npm/`npx` are out of scope, but they also must not over-promise:

- add npm global install and `npx` ephemeral usage to `docs/install-and-verify.md`
- update `README.md` to include npm as a supported channel for this milestone follow-up
- update `docs/distribution-strategy.md` and `docs/release-checklist.md` to include npm while keeping archives, Homebrew, and Scoop visible
- keep the support boundary best-effort and issue-driven
- keep `install --client ... --write` explicit and opt-in

## Scope Boundaries After Phase 15

Even after npm/`npx` support lands, these remain out of scope:

- JavaScript or TypeScript reimplementation of OptimusCtx behavior
- silent config-file edits during package installation
- WinGet, Chocolatey, `.deb`, `.rpm`, signing, or SBOM claims
- a hosted installer service or auto-update daemon
- provider-billed token accounting or any benchmark-related scope creep

## Risks and Pitfalls

- Current tests explicitly forbid npm/`npx` in docs and release config. Those guardrails must be updated deliberately rather than worked around.
- npm publication introduces an external namespace and token. The plan should assume `@niccrow/optimusctx` and `NPM_TOKEN`, not an unscoped package name that may be unavailable.
- The npm wrapper must stay version-pinned. `npx` should execute the package version being requested, not download an unrelated "latest" binary behind the user's back.
- Windows execution needs special handling for `optimusctx.exe` and archive extraction, so platform helpers should be explicit and test-covered.
- Package-manager docs must continue to point back to GitHub Release archives as rollback/fallback, not replace them.

## Validation Architecture

Phase 15 validation should stay mostly Go-driven, with one npm packaging smoke check layered on top:

- quick run: `go test ./internal/release ./internal/cli -run 'Test(NPMPackage|PackageManager|Distribution|ReleaseVerificationCommands)'`
- package smoke: `npm pack --dry-run ./packaging/npm`
- full suite: `go test ./...`

Manual-only verification is still required for confidence:

- `npm install -g @niccrow/optimusctx` on a clean macOS or Linux machine
- `npx @niccrow/optimusctx version` on a clean machine
- `optimusctx doctor` and `optimusctx snippet` after npm-based installation
- `optimusctx install --client claude-desktop` preview flow after npm-based installation

## Recommended Plan Split

### Wave 1

- `15-01`: npm package contract, metadata, and release-derived rendering foundation

Why first: every later launcher, docs, and publish step depends on one deterministic npm package definition that derives from the existing archive/checksum contract.

### Wave 2

- `15-02`: launcher/downloader implementation and install/verify guide updates
- `15-03`: npm publication workflow, policy, and supported-channel test updates

Why parallel: once the package contract exists, the runtime wrapper/docs and the CI/policy work can move independently with limited file overlap.

## Requirement Mapping

- `DIST-02`: npm becomes an additional real package-manager channel on top of Homebrew and Scoop
- `DIST-03`: install-and-verify docs gain npm global install and `npx` flows that still exercise the real CLI
- `DIST-04`: the distribution strategy and release checklist expand the supported channel policy while preserving explicit support boundaries

## Planning Recommendation

Plan Phase 15 as three execute plans:

1. define the npm package contract and deterministic release-derived metadata
2. implement the launcher/downloader plus install-guide verification flow
3. extend release automation, policy docs, and guardrail tests for npm support

That keeps the phase honest. npm/`npx` becomes a narrow wrapper channel around the existing shipped binary, not a new runtime or a hidden installer platform.
