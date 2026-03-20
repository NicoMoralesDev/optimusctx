# OptimusCtx Operator Release Guide

Use this as the canonical operator flow for `optimusctx release prepare`, tagged release publication, post-release verification, targeted downstream reruns, and rollback.

GitHub Release remains the canonical root and rollback source.

## Prepare

```bash
optimusctx release prepare
optimusctx release prepare --confirm
```

Use the exact reviewed tag for every later verification, `workflow_dispatch` rerun, and rollback decision.

`release prepare` now distinguishes:

- canonical GitHub Release readiness
- downstream publication readiness per channel
- credential verification gaps for Homebrew and Scoop

If Homebrew or Scoop stays `review_required`, `release prepare` could not verify the required GitHub Actions secret yet.
If either channel is `blocked`, the repository secret is known to be missing and that channel will not publish on tag push.

## Publish

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

That tag push starts `.github/workflows/release.yml` and builds the canonical GitHub Release archives plus checksum manifest first. npm, Homebrew, and Scoop publish from that same tagged release contract after the canonical release is available.

## Verify The Canonical GitHub Release

```bash
TAG="vX.Y.Z"
gh release view "$TAG"
mkdir -p /tmp/optimusctx-release-check
gh release download "$TAG" --dir /tmp/optimusctx-release-check
cd /tmp/optimusctx-release-check
VERSION="${TAG#v}"
sha256sum -c "optimusctx_${VERSION}_checksums.txt" --ignore-missing
```

Unpack one archive and verify the shipped binary surface:

```bash
tar -xzf "optimusctx_${VERSION}_linux_amd64.tar.gz"
mkdir -p /tmp/optimusctx-release-check/repo
cd /tmp/optimusctx-release-check/repo
git init
cp ../optimusctx .
./optimusctx version
./optimusctx init
./optimusctx status
./optimusctx doctor
./optimusctx init --client claude-desktop
```

If the canonical GitHub Release metadata, archives, or checksums are wrong, stop here. Fix the GitHub Release root before you inspect or rerun any downstream package-manager channel.

## Verify Downstream Channels Against The Same Tag

Every downstream check must trace back to the same GitHub Release tag and checksum manifest you verified first.

### npm

Publication target: `@niccrow/optimusctx`

```bash
npm install -g @niccrow/optimusctx
optimusctx version
optimusctx status
optimusctx doctor
```

### Homebrew

Publication target: `niccrow/homebrew-tap`

```bash
brew install niccrow/tap/optimusctx
optimusctx version
optimusctx status
optimusctx doctor
```

### Scoop

Publication target: `niccrow/scoop-bucket`

```powershell
scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git
scoop install niccrow/optimusctx
optimusctx version
optimusctx status
optimusctx doctor
```

## Rerun One Downstream Channel

Downstream reruns reuse the existing tag via `workflow_dispatch` with `release_tag` and `publication_channel`.

```bash
TAG="vX.Y.Z"
gh workflow run release.yml -f release_tag="$TAG" -f publication_channel=npm
```

Allowed `publication_channel` values are `all`, `npm`, `homebrew`, and `scoop`.

If the workflow summary shows `publication_status=not_published` for Homebrew or Scoop, that channel did not ship. Add the missing repository secret first, then rerun only that channel.

## Roll Back From The Canonical Root

1. Identify the last known good GitHub Release tag.
2. Download the prior tagged archive from GitHub Releases.
3. Reinstall that archive as the operator-facing fallback.
4. Re-run `optimusctx version`, `optimusctx status`, and `optimusctx doctor`.
5. Re-run `optimusctx init --client claude-desktop` if you need to verify supported-client onboarding again.
6. Only after the canonical archive path is stable should you decide whether a downstream rerun or a new fixed release is needed.

GitHub Release remains the canonical root and rollback source even when npm, Homebrew, or Scoop were the failing channels.
