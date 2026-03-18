# OptimusCtx Distribution Strategy

## Purpose

OptimusCtx v1.1 needs a narrow distribution story that helps real users adopt the shipped binary without implying a broader installer platform than the repository actually supports today.

This document defines the concrete v1.1 release channels, the users those channels are for, how upgrades and rollbacks work, and what support assumptions apply after installation. For the v1.2 release-operator procedure, use `docs/operator-release-guide.md` as the canonical release, rerun, verification, and rollback flow.

The guiding constraint is unchanged from the product itself: OptimusCtx is a local-first, single-binary tool. Distribution should make that binary easier to obtain, verify, and upgrade, not turn it into a managed service or an invasive installer.

## Product Shape And Guardrails

- OptimusCtx ships as one local-first binary.
- Installation and verification stay on the real shipped command surface.
- The supported post-install commands are `optimusctx version`, `optimusctx doctor`, `optimusctx snippet`, and the explicit opt-in `optimusctx install --client claude-desktop --write`.
- Distribution does not promise a hosted onboarding flow, background agent, or managed update service.
- Configuration writes remain explicit. `optimusctx install` is preview-first, and config files are only written when the operator opts into `--write`.

## Intended First Users

The first public distribution path is for developers and coding-agent operators who are already comfortable with:

- installing CLI tools from release archives or package managers
- managing their own PATH and shell environment
- running local verification commands after installation
- configuring MCP integrations explicitly instead of expecting automatic repository edits

This is also the target audience for early team adoption. The expected rollout path is a user or small team evaluating the existing local-first runtime on their own machines and deciding whether the verification output matches their workflow needs.

## Supported Release Channels

### 1. GitHub Release Archives

GitHub Release archives are the fallback and baseline channel for v1.1.

- Publication target: `github.com/niccrow/optimusctx releases`
- Audience: users who want the raw binary, need a fallback when package-manager metadata lags, or prefer explicit archive installs
- Install path: download the tagged archive from GitHub Releases, unpack it, place `optimusctx` on your PATH, then verify with `optimusctx version` and `optimusctx doctor`
- Strength: this is the least opinionated path and the rollback source for every other channel

### 2. Homebrew

Homebrew is the primary package-manager channel for macOS and Linux users.

- Publication target: `niccrow/homebrew-tap`
- Audience: macOS and Linux users who already rely on Homebrew for CLI distribution
- Install command: `brew install niccrow/tap/optimusctx`
- Upgrade command: `brew upgrade niccrow/tap/optimusctx`
- Verification after install or upgrade: rerun `optimusctx version`, `optimusctx doctor`, and `optimusctx snippet`

### 3. Scoop

Scoop is the primary package-manager channel for Windows users.

- Publication target: `niccrow/scoop-bucket`
- Audience: Windows users who already manage developer tooling through Scoop
- Install commands:
  - `scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git`
  - `scoop install niccrow/optimusctx`
- Upgrade command: `scoop update optimusctx`
- Verification after install or upgrade: rerun `optimusctx version`, `optimusctx doctor`, and `optimusctx snippet`

### 4. npm and npx

npm is the JavaScript ecosystem wrapper channel for users who prefer `npm install -g` or `npx`, while still running the real tagged OptimusCtx binary.

- Publication target: npm registry package `@niccrow/optimusctx`
- Audience: users who expect package installation through npm but still want the real Go release binary underneath the wrapper
- Install commands:
  - `npm install -g @niccrow/optimusctx`
  - `npx @niccrow/optimusctx version`
- Upgrade command: `npm install -g @niccrow/optimusctx@latest`
- Verification after install or upgrade: rerun `optimusctx version`, `optimusctx doctor`, `optimusctx snippet`, and only then the explicit `optimusctx install --client claude-desktop --write` path if desired
- Support boundary: the npm package is a wrapper over the GitHub Release archives, not a JavaScript reimplementation or a silent client-config installer

## How Users Choose A Channel

The intended channel order is:

1. npm if the user wants the simplest install path and is comfortable with npm.
2. Homebrew if the user is on macOS or Linux and already uses Homebrew.
3. Scoop if the user is on Windows and already uses Scoop.
4. GitHub Release archives if the user wants the raw binary, needs a rollback, or does not want to depend on a package manager.

This keeps the product truthful about supported channels while still giving every user one direct binary path that does not depend on an external package-manager state.

## Upgrade Policy

Archive users upgrade by replacing the binary manually and verifying the shipped command surface again.

Package-manager users upgrade through the channel-native command while GitHub Release archives remain the rollback fallback, including the npm wrapper package.

In practice that means:

- GitHub Release archive users download a newer tagged archive, replace the binary manually on their PATH, and rerun the verification commands.
- Homebrew users run `brew upgrade niccrow/tap/optimusctx`, then rerun `optimusctx version` and `optimusctx doctor`.
- Scoop users run `scoop update optimusctx`, then rerun `optimusctx version` and `optimusctx doctor`.
- npm users rerun `npm install -g @niccrow/optimusctx`, or use `npx @niccrow/optimusctx version` for ephemeral execution, then verify with `optimusctx version`, `optimusctx doctor`, and `optimusctx snippet`.

OptimusCtx does not ship an in-product auto-updater. Users are expected to choose when to upgrade and to verify the installed binary explicitly after doing so.

## Rollback Expectations

Recovery and rollback stay simple and explicit:

- If the canonical GitHub Release archives or checksums are wrong, stop and fix GitHub Release first before you rerun any downstream publication channel.
- If exactly one downstream channel failed after the canonical release is correct, rerun only that channel with `gh workflow run release.yml -f release_tag=vX.Y.Z -f publication_channel=npm`, `gh workflow run release.yml -f release_tag=vX.Y.Z -f publication_channel=homebrew`, or `gh workflow run release.yml -f release_tag=vX.Y.Z -f publication_channel=scoop`.
- The primary rollback source is a prior tagged GitHub Release archive.
- If a published release should be abandoned, reinstall a prior tagged GitHub Release archive, verify the fallback binary again, and publish a new fixed version instead of reusing the same version.
- If a Homebrew or Scoop upgrade causes a local problem, the documented recovery path is to reinstall a prior tagged archive from GitHub Releases.
- If npm or `npx` wrapper execution causes a local problem, the documented recovery path is to reinstall or rerun the wrapper against a prior tagged GitHub Release version, or fall back directly to the archive channel.
- Rollback is a binary replacement operation, not a managed state migration system.

Because the tool is single-binary and local-first, rollback does not require uninstalling an agent service or coordinating with a hosted control plane.

## Verification Expectations

Every supported channel should converge on the same verification path:

1. `optimusctx version` to confirm the installed binary reports the expected release metadata.
2. `optimusctx doctor` to confirm the local runtime can report its health honestly.
3. `optimusctx snippet` to show the manual MCP integration contract.
4. `optimusctx install --client claude-desktop --write` only if the operator explicitly wants the config file write path after reviewing the preview output.

These checks are intentionally narrow. They prove the released binary works locally without implying extra installers, hidden services, or automatic configuration edits.

## Support Boundary

Support for v1.1 is best-effort and issue-driven through repository documentation and GitHub issues rather than a managed installer or helpdesk.

Supported help covers:

- obtaining the binary from one of the named channels
- running the documented install commands for Homebrew, Scoop, or npm
- verifying the binary with `version`, `doctor`, and `snippet`
- understanding the explicit `install --client` preview and write flow

Support does not cover:

- repairing unrelated Homebrew or Scoop environment problems outside the documented flow
- operating a hosted update or fleet rollout system
- automatic mutation of repository instruction files
- undocumented package-manager channels

## Deferred Scope

The following distribution work is explicitly deferred to later milestones:

- native Linux packages such as `.deb` and `.rpm`
- WinGet
- Chocolatey
- artifact signing
- SBOM publication

These items can matter later, but they are not required for a truthful v1.1 release. The current milestone is about a credible narrow path, not maximizing channel count.

## Release Messaging Guidance

When communicating v1.1 externally or to early adopters, describe the release as:

- a local-first single binary
- available through GitHub Release archives, Homebrew, Scoop, and the npm wrapper package
- upgraded explicitly by the operator
- verified with `optimusctx version`, `optimusctx doctor`, and `optimusctx snippet`
- supported on a best-effort, issue-driven basis

Avoid language that implies:

- managed rollout infrastructure
- hidden post-install automation
- native Linux packages
- WinGet or Chocolatey support
- signed artifacts or SBOM availability

## Summary

The v1.1 distribution strategy is intentionally small: one direct archive channel, Homebrew for macOS and Linux, Scoop for Windows, one npm wrapper channel for the JavaScript ecosystem, one explicit verification path, and one clear support boundary. That is enough to support adoption without over-promising beyond the current product shape.
