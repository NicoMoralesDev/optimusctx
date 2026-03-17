---
phase: 13-distribution-pipeline-and-adoption-plan
research_date: 2026-03-16
objective: "What do I need to know to PLAN this phase well?"
status: complete
---

# Phase 13 Research: Distribution Pipeline and Adoption Plan

## Executive Summary

Phase 13 should package the shipped OptimusCtx binary without changing the product shape. The repository already has the runtime features and operator-facing commands that distribution must expose:

- `optimusctx version` already prints build metadata through `internal/buildinfo`
- `optimusctx install` already previews or writes supported MCP client registration
- `optimusctx snippet` already prints the manual-copy MCP contract
- `optimusctx doctor` already provides the operator-facing health story that first-run verification should use

What is missing is the distribution layer around that existing surface:

1. a tag-driven release pipeline that produces versioned archives and checksums
2. one narrow package-manager path for macOS/Linux and one for Windows
3. a truthful install-and-verify path that exercises the real shipped commands rather than `go run`
4. a concrete rollout and support plan that defines who the first channels are for and how upgrades are expected to work

The main planning risk is pretending the project already has a release system. It does not. There is no `.github/` workflow directory, no GoReleaser config, no packaging metadata, and no release runbook in the repository today. Phase 13 therefore needs to build the release substrate first, then layer package-manager manifests and operator docs on top of that single source of truth.

## Repository Reality Check

### Inputs reviewed

- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `.planning/ROADMAP.md`
- `README.md`
- `cmd/optimusctx/main.go`
- `internal/buildinfo/buildinfo.go`
- `internal/cli/version.go`
- `internal/cli/install.go`
- `internal/app/install.go`
- `internal/app/snippet.go`

### Local guidance files

- `CLAUDE.md`: not present
- `.claude/skills/`: not present
- `.agents/skills/`: not present

### What already exists

- The runtime is still a single Go binary with no hosted dependency or installer daemon.
- `README.md` documents `go install ./cmd/optimusctx` for developers, but that is not yet a release-grade end-user distribution path.
- `internal/buildinfo/buildinfo.go` already exposes `Version`, `Commit`, and `BuildDate` variables for ldflags-backed release metadata.
- `internal/cli/install.go` normalizes ephemeral `go run` binary paths back to `optimusctx`, which is useful for docs and package-manager verification.
- `internal/app/snippet.go` and `internal/app/install.go` already share a canonical MCP serve contract.
- There is no existing CI/release automation in the repo root.

### Distribution constraints already locked by project state

- Distribution must stay narrow: cross-platform archives, first package-manager channels, install verification, and rollout planning.
- The product must remain non-invasive. Distribution cannot introduce automatic edits to repository instruction files.
- The project is still single-binary and local-first, so packaging should preserve that shape instead of adding services or background agents.
- `DIST-05` and `DIST-06` are explicitly v2. Native Linux package formats, signing, and SBOMs should not become hidden blockers for v1.1.

## What Phase 13 Must Decide Before Planning

## 1. Release source of truth

Phase 13 needs one release definition that drives:

- archive naming
- target OS/arch matrix
- checksum generation
- release metadata
- package-manager manifests

Recommended approach: use GoReleaser plus a GitHub Actions workflow as the single release source of truth, with GitHub Releases as the first user-retrievable archive channel. The repo already has build metadata variables and a standard Go module layout, which makes GoReleaser a good fit for:

- tar.gz archives for macOS/Linux
- zip archives for Windows
- checksum generation
- ldflags injection for `version`, `commit`, and `build_date`

Planning implication: do not hand-author per-platform archive scripts if GoReleaser can define the matrix once and expose the same artifact metadata to later package-manager templates.

## 2. Package-manager scope

`DIST-02` requires at least one primary package-manager path on macOS/Linux and one on Windows. The narrowest credible v1.1 channel set is:

- Homebrew for macOS and Linux
- Scoop for Windows

This aligns with the repository's single-binary story and avoids prematurely expanding into:

- `apt` or `yum`
- `.deb` or `.rpm`
- Chocolatey or WinGet
- shell installers that drift from the release archives

Planning implication: phase scope should include a consumable publication path for Homebrew and Scoop, not just local manifest templates. If dedicated tap or bucket repos need credentials or user-owned repository names, that must be surfaced explicitly in the plan as configuration or checkpoint work rather than silently assumed away.

## 3. Install verification must use the real shipped command surface

`DIST-03` explicitly names `doctor` and `snippet`. The install story should therefore prove:

- the binary can be obtained from a release artifact or package-manager path
- `optimusctx version` reports injected release metadata
- `optimusctx doctor` runs and reports a healthy/no-state baseline honestly on a local machine
- `optimusctx snippet` prints the manual-copy MCP integration contract
- `optimusctx install --client ...` remains an explicit opt-in preview or write step, not an automatic side effect of installation

Planning implication: end-user docs should not use `go run` as the main install path. `go run` remains useful for developers, but release docs need to validate the actual shipped binary and package-manager paths.

## 4. Rollout and support assumptions must stay explicit

The roadmap says Phase 13 must define release channels, target users, upgrade path, and support assumptions. That means the phase needs more than a README edit. It needs a concrete plan covering:

- who the initial channels are for
- how users discover the correct install path
- what "upgrade" means for archive users versus package-manager users
- what support promises exist for failed installs, stale configs, or MCP registration problems
- which items are intentionally deferred to v2

Recommended v1.1 audience and support shape:

- first users: local-first coding-agent users comfortable with CLI tools and explicit MCP configuration
- supported channels: release archives, Homebrew, Scoop
- supported verification: `version`, `doctor`, `snippet`, optional `install --client`
- support boundary: best-effort issue-driven support documented in repo docs, not a managed installer or hosted update service

## Recommended Technical Direction

### Release automation

- Add `.goreleaser.yml` as the archive/checksum and metadata source of truth.
- Add `.github/workflows/release.yml` to trigger on version tags and `workflow_dispatch`, then publish archives and checksums to GitHub Releases.
- Inject build metadata through ldflags wired to `internal/buildinfo`.
- Keep release outputs narrow: archives and checksums published through GitHub Releases only.

### Package-manager distribution

- Generate a Homebrew formula from release metadata rather than copying values into docs by hand, and wire it to a consumable tap publication path.
- Generate a Scoop manifest from the same version/checksum source, and wire it to a consumable bucket publication path.
- Keep manifest rendering deterministic and testable in Go.
- If publication targets need credentials or external repositories, make that operator setup explicit instead of hiding it in undocumented workflow assumptions.

### Install-and-verify workflow

- Add one operator-facing install guide that starts from the shipped binary or package-manager install path.
- Use a disposable temp repository for the `doctor` verification path so docs match the existing eval/verification discipline.
- Keep MCP setup explicit by pointing users from `snippet` to `install --client` instead of silently editing configs during install.

### Distribution strategy and rollout docs

- Add a dedicated distribution strategy document instead of burying release policy in README prose.
- Document release channels, upgrade expectations, rollback fallback, and known support assumptions.
- Explicitly call out v2 deferrals: `.deb`, `.rpm`, signing, SBOMs, and broader package-manager expansion.

## Pitfalls To Avoid

- Do not make npm, `npx`, or hosted installers reappear through side doors. They are still out of scope.
- Do not let package-manager manifests diverge from archive names or checksum sources.
- Do not base install docs on `go run` or local build-cache paths.
- Do not claim signed artifacts, SBOMs, WinGet, Chocolatey, `.deb`, or `.rpm` support in v1.1.
- Do not make `install` mutate configs during verification unless the user explicitly passes `--write`.
- Do not introduce a release process that only works on one OS. The phase is specifically about cross-platform distribution.

## Validation Architecture

Phase 13 validation should keep one fast command for planning/execution feedback and one full-suite command for phase gates:

- quick run: `go test ./internal/buildinfo ./internal/cli ./internal/app ./internal/release -run 'Test(BuildInfo|Version|Install|Snippet|Doctor|Release|PackageManager|Distribution)'`
- full run: `go test ./...`

Wave-specific verification should also cover:

- deterministic release metadata and archive naming
- checksum manifest generation
- deterministic Homebrew formula and Scoop manifest rendering
- install-doc examples that stay aligned with the actual command surface

Manual-only validation is still required for final confidence:

- install from a produced archive on a clean machine
- install through Homebrew on macOS/Linux
- install through Scoop on Windows
- run `optimusctx version`, `optimusctx doctor`, and `optimusctx snippet` after installation

## Recommended Plan Split

### Wave 1

- `13-01`: automated release archives and checksums

Why first: every later package-manager path and install document depends on one stable archive/checksum contract.

### Wave 2

- `13-02`: primary package-manager distribution paths

Why second: Homebrew and Scoop manifests should derive from the release metadata defined in `13-01`, not invent their own version/checksum sources.

### Wave 3

- `13-03`: install-and-verify documentation and smoke flow
- `13-04`: distribution strategy, rollout, and support plan

Why parallel: once release artifacts and manifest shapes are defined, the verification docs and adoption strategy can proceed independently while referencing the same release channels.

## Requirement Mapping

- `DIST-01`: release archives, checksums, and automated publication through GitHub Releases
- `DIST-02`: Homebrew plus Scoop as the first consumable package-manager channels
- `DIST-03`: archive/package-manager install docs plus `doctor` and `snippet` verification
- `DIST-04`: explicit rollout, upgrade, and support policy docs

## Planning Recommendation

Plan Phase 13 as four execute plans:

1. create the release artifact source of truth
2. add deterministic package-manager manifests from that source
3. document and smoke-test the real install-and-verify path
4. publish the narrow rollout and support strategy

This keeps the milestone honest. The phase ships packaging and adoption infrastructure around the existing runtime instead of reopening product scope.
