# OptimusCtx Operator Release Guide

Use this as the canonical operator flow for `optimusctx release prepare`, tagged release publication, post-release verification, targeted downstream reruns, and rollback.

GitHub Release remains the canonical root and rollback source.

## Prepare

Run the non-mutating review first:

```bash
optimusctx release prepare
```

Confirm the reviewed plan only when the version, tag, and channel readiness look correct:

```bash
optimusctx release prepare --confirm
```

The prepare output should leave you with one canonical semver tag such as `v1.2.3`. Use that exact tag for every later verification, `workflow_dispatch` rerun, and rollback decision.

## Publish

Create and push the reviewed tag:

```bash
git tag v1.2.3
git push origin v1.2.3
```

That tag push starts `.github/workflows/release.yml` and builds the canonical GitHub Release archives plus checksum manifest first. npm, Homebrew, and Scoop publish from that same tagged release contract after the canonical release is available.

## Verify The Canonical GitHub Release

Inspect the release metadata for the exact tag:

```bash
TAG="v1.2.3"
gh release view "$TAG"
```

Download the release assets into a clean verification directory:

```bash
mkdir -p /tmp/optimusctx-release-check
gh release download "$TAG" --dir /tmp/optimusctx-release-check
```

Verify the checksum manifest against one downloaded archive before you inspect downstream channels:

```bash
cd /tmp/optimusctx-release-check
VERSION="${TAG#v}"
sha256sum -c "optimusctx_${VERSION}_checksums.txt" --ignore-missing
```

Unpack one archive and verify the shipped binary surface:

```bash
tar -xzf "optimusctx_${VERSION}_linux_amd64.tar.gz"
./optimusctx version
./optimusctx doctor
./optimusctx snippet
```

If the canonical GitHub Release metadata, archives, or checksums are wrong, stop here. Fix the GitHub Release root before you inspect or rerun any downstream package-manager channel.

## Verify Downstream Channels Against The Same Tag

Every downstream check must trace back to the same GitHub Release tag and checksum manifest you verified first.

### npm

Publication target: `@niccrow/optimusctx`

- Confirm the published package version matches the release tag version.
- Confirm the wrapper still resolves to the same tagged GitHub Release binary.
- Re-run the shipped verification commands after install:

```bash
npm install -g @niccrow/optimusctx
optimusctx version
optimusctx doctor
optimusctx snippet
```

### Homebrew

Publication target: `niccrow/homebrew-tap`

- Confirm the tap update points at the same GitHub Release archive URLs and checksum values for `TAG`.
- Re-run the shipped verification commands after install or upgrade:

```bash
brew install niccrow/tap/optimusctx
optimusctx version
optimusctx doctor
optimusctx snippet
```

### Scoop

Publication target: `niccrow/scoop-bucket`

- Confirm the Scoop manifest points at the same GitHub Release archive URLs and checksum values for `TAG`.
- Re-run the shipped verification commands after install or upgrade:

```powershell
scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git
scoop install niccrow/optimusctx
optimusctx version
optimusctx doctor
optimusctx snippet
```

## Rerun One Downstream Channel

Use this branch only when the canonical GitHub Release is already correct and one downstream publication channel failed or needs to be republished.

Downstream reruns reuse the existing tag via `workflow_dispatch` with `release_tag` and `publication_channel`.

Example with GitHub CLI:

```bash
TAG="v1.2.3"
gh workflow run release.yml -f release_tag="$TAG" -f publication_channel=npm
```

Allowed `publication_channel` values are `all`, `npm`, `homebrew`, and `scoop`. Prefer the narrowest rerun that fixes the failing downstream path. Keep `release_tag` set to the existing canonical GitHub Release tag instead of creating a new tag for a single-channel retry.

After the rerun completes, repeat the matching downstream verification steps and confirm the package-manager output still points back to the same GitHub Release root.

## Roll Back From The Canonical Root

Use rollback when the published output should no longer be the active operator path. Do not treat package-manager-specific recovery as the canonical rollback path.

1. Identify the last known good GitHub Release tag.
2. Download the prior tagged archive from GitHub Releases.
3. Reinstall that archive as the operator-facing fallback.
4. Re-run `optimusctx version`, `optimusctx doctor`, and `optimusctx snippet`.
5. Only after the canonical archive path is stable should you decide whether a downstream rerun or a new fixed release is needed.

GitHub Release remains the canonical root and rollback source even when npm, Homebrew, or Scoop were the failing channels.
