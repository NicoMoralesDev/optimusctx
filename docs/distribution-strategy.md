# OptimusCtx Distribution Strategy

## Purpose

OptimusCtx keeps a narrow distribution story that helps real users adopt the shipped binary without implying a broader installer platform than the repository actually supports today.

This document defines the concrete release channels, the users those channels are for, how upgrades and rollbacks work, and what support assumptions apply after installation. For the current release-operator procedure, use `docs/operator-release-guide.md` as the canonical release, rerun, verification, and rollback flow.

The guiding constraint is unchanged from the product itself: OptimusCtx is a local-first single binary. Distribution should make that binary easier to obtain, verify, and upgrade, not turn it into a managed service or an invasive installer.

## Product Shape And Guardrails

- OptimusCtx ships as one local-first binary.
- Installation and verification stay on the real shipped command surface.
- The supported post-install commands are `optimusctx version`, `optimusctx init`, `optimusctx status`, and the agent-facing runtime entrypoint `optimusctx run`; `optimusctx doctor` remains only as a deprecated alias for `status`.
- Distribution does not promise a hosted onboarding flow, background agent, or managed update service.
- Configuration writes remain explicit. `optimusctx init --client ...` reviews the exact change first, and host registration is only written when the operator opts into `--write`.
- When a supported host is registered, it should launch `optimusctx run` automatically; manual `run` remains the direct/debug path.

## Supported Release Channels

### 1. GitHub Release Archives

GitHub Release archives are the fallback and baseline channel.

- Publication target: `github.com/NicoMoralesDev/optimusctx releases`
- Audience: users who want the raw binary, need a fallback when package-manager metadata lags, or prefer explicit archive installs
- Install path: download the tagged archive from GitHub Releases, unpack it, place `optimusctx` on your PATH, then verify with `optimusctx version` and `optimusctx status`
- Strength: this is the least opinionated path and the rollback source for every other channel

### 2. Homebrew

- Publication target: `niccrow/homebrew-tap`
- Install command: `brew install niccrow/tap/optimusctx`
- Upgrade command: `brew upgrade niccrow/tap/optimusctx`
- Verification after install or upgrade: rerun `optimusctx version` and `optimusctx status`

### 3. Scoop

- Publication target: `niccrow/scoop-bucket`
- Install commands:
  - `scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git`
  - `scoop install niccrow/optimusctx`
- Upgrade command: `scoop update optimusctx`
- Verification after install or upgrade: rerun `optimusctx version` and `optimusctx status`

### 4. npm and npx

- Publication target: npm registry package `@niccrow/optimusctx`
- Install commands:
  - `npm install -g @niccrow/optimusctx`
  - `npx @niccrow/optimusctx version`
- Upgrade command: `npm install -g @niccrow/optimusctx@latest`
- Verification after install or upgrade: rerun `optimusctx version` and `optimusctx status`, then use `optimusctx init --client claude-desktop` to review onboarding and `optimusctx init --client claude-desktop --write` only if desired
- Support boundary: the npm package is a wrapper over the canonical tagged GitHub Release binary, not a JavaScript reimplementation or a silent client-config installer

## Upgrade Policy

Archive users upgrade by replacing the binary manually and verifying the shipped command surface again.

Package-manager users upgrade through the channel-native command while GitHub Release archives remain the rollback fallback, including the npm wrapper package.

In practice that means:

- GitHub Release archive users download a newer tagged archive, replace the binary manually on their PATH, and rerun the verification commands.
- Homebrew users run `brew upgrade niccrow/tap/optimusctx`, then rerun `optimusctx version` and `optimusctx status`.
- Scoop users run `scoop update optimusctx`, then rerun `optimusctx version` and `optimusctx status`.
- npm users rerun `npm install -g @niccrow/optimusctx@latest`, or use `npx @niccrow/optimusctx version` for ephemeral execution, then verify with `optimusctx version` and `optimusctx status`.

## Rollback Expectations

- If the canonical GitHub Release archives or checksums are wrong, stop and fix GitHub Release first before you rerun any downstream publication channel.
- If exactly one downstream channel failed after the canonical release is correct, rerun only that channel with `gh workflow run release.yml -f release_tag=vX.Y.Z -f publication_channel=npm`, `gh workflow run release.yml -f release_tag=vX.Y.Z -f publication_channel=homebrew`, or `gh workflow run release.yml -f release_tag=vX.Y.Z -f publication_channel=scoop`.
- If Homebrew or Scoop shows `publication_status=not_published`, that channel did not ship; add the missing repository secret and rerun only that channel.
- The primary rollback source is a prior tagged GitHub Release archive.
- If a published release should be abandoned, reinstall a prior tagged GitHub Release archive, verify the fallback binary again, and publish a new fixed version instead of reusing the same version.
- GitHub Release is the canonical root and rollback source even when downstream automation republishes one package-manager channel.

## Verification Expectations

Every supported channel should converge on the same verification path:

1. `optimusctx version` to confirm the installed binary reports the expected release metadata.
2. `optimusctx status` to confirm the runtime and repository state are ready.
3. `optimusctx doctor` only if you still rely on the deprecated alias; it now delegates to `status`.
4. `optimusctx init --client claude-desktop` to review supported-client onboarding in a real repository.
5. `optimusctx init --client claude-desktop --write` only if the operator explicitly wants the config file write path after reviewing the rendered change.
6. Confirm the MCP host exposes and uses `optimusctx.*` tools if the release claim includes supported-client onboarding.

## Support Boundary

Supported help covers:

- obtaining the binary from one of the named channels
- running the documented install commands for Homebrew, Scoop, or npm
- verifying the binary with `version` and `status`
- understanding the explicit `init --client` review and apply flow
- best-effort and issue-driven support through repository docs and GitHub issues

Support does not cover:

- repairing unrelated Homebrew or Scoop environment problems outside the documented flow
- operating a hosted update or fleet rollout system
- automatic or silent mutation of repository instruction files outside explicit `init ... --write`
- undocumented package-manager channels

## Deferred Scope

The following distribution work is explicitly deferred to later milestones:

- native Linux packages such as `.deb` and `.rpm`
- WinGet
- Chocolatey
- artifact signing
- SBOM publication

## Release Messaging Guidance

When communicating the current OptimusCtx release externally or to early adopters, describe it as:

- a local-first single binary
- available through GitHub Release archives, Homebrew, Scoop, and the npm wrapper package
- upgraded explicitly by the operator
- verified with `optimusctx version` and `optimusctx status`
- supported on a best-effort, issue-driven basis
