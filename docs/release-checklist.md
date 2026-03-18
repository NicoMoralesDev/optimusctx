# OptimusCtx Release Checklist

## Goal

Use this checklist when publishing a v1.2 release so the rollout stays aligned with the supported channel contract in `docs/distribution-strategy.md`.

For the full operator path from `optimusctx release prepare` through verification, targeted rerun, and rollback, use [`operator-release-guide.md`](./operator-release-guide.md) as the canonical workflow.

GitHub Release is the canonical root for archives, checksums, and downstream release facts.
After GitHub Release assets are available, npm, Homebrew, and Scoop are published from the same canonical tagged release contract.

## Pre-Tag Checks

- Confirm the release is still limited to the shipped single-binary, local-first product shape.
- Confirm the supported channels are still GitHub Release archives, Homebrew, Scoop, and the npm wrapper package only.
- Confirm GitHub Release remains the canonical tagged root that downstream channels consume rather than a peer channel.
- Confirm the deferred items are still deferred: `.deb`, `.rpm`, WinGet, Chocolatey, artifact signing, and SBOM publication.
- Run the release verification tests before tagging:
  - `go test ./internal/release -run 'TestMultiChannelPublicationDocsStayCanonical|TestChannelPublicationWorkflowSelectiveRerun|TestCanonicalReleaseMatchesGoReleaserContract'`
  - `go test ./internal/release ./internal/cli -run 'TestRolloutPlanExamples|TestUpgradePolicy'`
  - `go test ./...`

## Tag And Publish

- Follow [`operator-release-guide.md`](./operator-release-guide.md) for the canonical end-to-end operator flow before and after the tag push.
- Create the release tag that should drive the canonical GitHub Release and downstream publication flow.
- Use `workflow_dispatch` with `release_tag` and `publication_channel` to rerun `npm`, `homebrew`, or `scoop` for an existing tagged release without rebuilding unrelated channels.
- Verify the GitHub Release contains the canonical versioned archives and checksum manifest for the tag.
- Verify the release metadata lines up with `optimusctx version`.
- Verify the npm package was rendered from `scripts/render-npm-package.sh` and published with `npm publish` against the same tagged release facts.
- Verify Homebrew and Scoop were rendered and published from the same tagged GitHub Release checksum and archive facts.
- Treat GitHub Release archives as the baseline distribution channel, canonical metadata root, and rollback source.

## GitHub Release Archive Checks

- Download one produced archive and confirm it unpacks to the `optimusctx` binary.
- Verify the checksum manifest is present for the tagged release.
- Confirm the archive and checksum names match the canonical tagged GitHub Release contract.
- Confirm the archive instructions stay truthful: unpack, place the binary on PATH, then run `optimusctx version` and `optimusctx doctor`.
- Confirm rollback guidance still points to reinstalling a prior tagged GitHub Release archive.

## Homebrew Checks

- Confirm Homebrew still consumes the existing tagged GitHub Release facts rather than a separate archive source.
- Confirm the Homebrew publication target is `niccrow/homebrew-tap`.
- Confirm the user-facing install command is `brew install niccrow/tap/optimusctx`.
- Confirm the user-facing upgrade command is `brew upgrade niccrow/tap/optimusctx`.
- Confirm the release operator credentials for publication are still `HOMEBREW_TAP_GITHUB_TOKEN`.
- Confirm Homebrew messaging stays scoped to macOS and Linux users who already use Homebrew.
- Confirm Homebrew publication is automated from the same canonical tagged release after GitHub Release assets are available.
- If Homebrew needs recovery, rerun `workflow_dispatch` with `release_tag=<tag>` and `publication_channel=homebrew`.

## Scoop Checks

- Confirm Scoop still consumes the existing tagged GitHub Release facts rather than a separate archive source.
- Confirm the Scoop publication target is `niccrow/scoop-bucket`.
- Confirm the user-facing install commands are `scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git` and `scoop install niccrow/optimusctx`.
- Confirm the user-facing upgrade command is `scoop update optimusctx`.
- Confirm the release operator credentials for publication are still `SCOOP_BUCKET_GITHUB_TOKEN`.
- Confirm Scoop messaging stays scoped to Windows users who already use Scoop.
- Confirm Scoop publication is automated from the same canonical tagged release after GitHub Release assets are available.
- If Scoop needs recovery, rerun `workflow_dispatch` with `release_tag=<tag>` and `publication_channel=scoop`.

## npm Checks

- Confirm the npm publication target is `@niccrow/optimusctx`.
- Confirm the user-facing install command is `npm install -g @niccrow/optimusctx`.
- Confirm the user-facing ephemeral command is `npx @niccrow/optimusctx version`.
- Confirm the release operator credentials for publication are still `NPM_TOKEN`.
- Confirm the npm package remains a wrapper over the canonical tagged GitHub Release binary rather than a separate runtime implementation.
- If npm needs recovery, rerun `workflow_dispatch` with `release_tag=<tag>` and `publication_channel=npm`.

## Verification Commands

- After installation or upgrade, run `optimusctx version`.
- After installation or upgrade, run `optimusctx doctor`.
- After installation or upgrade, run `optimusctx snippet`.
- Use `optimusctx install --client claude-desktop --write` only when the operator explicitly wants the config-write path after reviewing the preview output.

## Rollout Messaging

- Describe the release as a local-first single binary.
- Name the supported channels exactly: GitHub Release archives, Homebrew, Scoop, and the npm wrapper package.
- State that upgrades are explicit operator actions, not automatic background updates.
- State that support is best-effort and issue-driven through repository docs and GitHub issues.
- Avoid claims about native Linux packages, WinGet, Chocolatey, signed artifacts, or SBOM publication.

## Support Follow-Through

- Watch incoming reports for failures in `optimusctx version`, `optimusctx doctor`, `optimusctx snippet`, or the explicit `install --client` flow.
- Ask for the exact channel used by the reporter: GitHub Release archive, Homebrew, Scoop, or npm.
- Treat undocumented channels as unsupported and route users back to the named release channels.
- Prefer GitHub Release archive rollback guidance when package-manager state is unclear.

## Release Complete

- The tagged GitHub Release artifacts are published and retrievable as the canonical root.
- The npm, Homebrew, and Scoop publication paths match the same automated downstream contract rooted in the canonical tagged GitHub Release.
- The npm publication and install path match the structured policy contract.
- The docs still describe the real verification and support flow.
- The release remains narrow, truthful, and aligned with the v1.2 operator plan.
