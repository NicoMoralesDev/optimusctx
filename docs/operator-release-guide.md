# OptimusCtx Operator Release Guide

Use this as the canonical operator flow for `optimusctx release prepare`, tagged release publication, post-release verification, targeted downstream reruns, and rollback.

GitHub Release remains the canonical root and rollback source.

## Prepare

```bash
optimusctx release prepare
optimusctx release prepare --confirm
```

Use the exact reviewed tag for every later verification, `workflow_dispatch` rerun, and rollback decision.

## Publish

```bash
git tag v1.2.3
git push origin v1.2.3
```

That tag push starts `.github/workflows/release.yml` and builds the canonical GitHub Release archives plus checksum manifest first. npm, Homebrew, and Scoop publish from that same tagged release contract after the canonical release is available.

## Verify The Canonical GitHub Release

```bash
TAG="v1.2.3"
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
./optimusctx version
./optimusctx status
./optimusctx doctor
./optimusctx status --client claude-desktop
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
TAG="v1.2.3"
gh workflow run release.yml -f release_tag="$TAG" -f publication_channel=npm
```

Allowed `publication_channel` values are `all`, `npm`, `homebrew`, and `scoop`.

## Roll Back From The Canonical Root

1. Identify the last known good GitHub Release tag.
2. Download the prior tagged archive from GitHub Releases.
3. Reinstall that archive as the operator-facing fallback.
4. Re-run `optimusctx version`, `optimusctx status`, and `optimusctx doctor`.
5. Only after the canonical archive path is stable should you decide whether a downstream rerun or a new fixed release is needed.

GitHub Release remains the canonical root and rollback source even when npm, Homebrew, or Scoop were the failing channels.
