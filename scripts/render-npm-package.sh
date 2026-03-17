#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_DIR="${ROOT_DIR}/packaging/npm"
TAG="${1:-}"
OUTPUT_DIR="${2:-${ROOT_DIR}/dist/npm-package}"

if [[ -z "${TAG}" ]]; then
  echo "usage: scripts/render-npm-package.sh <vX.Y.Z> [output-dir]" >&2
  exit 1
fi

case "${TAG}" in
  v*) ;;
  *)
    echo "tag must start with v" >&2
    exit 1
    ;;
esac

VERSION_NO_V="${TAG#v}"

for required in \
  "${SOURCE_DIR}/package.json" \
  "${SOURCE_DIR}/bin/optimusctx.js" \
  "${SOURCE_DIR}/lib/install.js" \
  "${SOURCE_DIR}/lib/platform.js"; do
  if [[ ! -f "${required}" ]]; then
    echo "missing required npm package file: ${required}" >&2
    exit 1
  fi
done

rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"
cp -R "${SOURCE_DIR}/." "${OUTPUT_DIR}/"
chmod +x "${OUTPUT_DIR}/bin/optimusctx.js"

PACKAGE_JSON_PATH="${OUTPUT_DIR}/package.json"

PACKAGE_JSON_PATH="${PACKAGE_JSON_PATH}" RELEASE_TAG="${TAG}" VERSION_NO_V="${VERSION_NO_V}" node <<'EOF'
const fs = require('node:fs');

const packageJsonPath = process.env.PACKAGE_JSON_PATH;
const releaseTag = process.env.RELEASE_TAG;
const versionNoV = process.env.VERSION_NO_V;

if (!packageJsonPath || !releaseTag || !versionNoV) {
  throw new Error('package.json rendering requires PACKAGE_JSON_PATH, RELEASE_TAG, and VERSION_NO_V');
}

const pkg = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
const currentVersion = pkg.optimusctx.version;
const currentTag = pkg.optimusctx.releaseTag;
const releasePath = `/releases/download/${currentTag}/`;
const nextReleasePath = `/releases/download/${releaseTag}/`;

function retagCanonicalURL(url) {
  if (!url.includes(releasePath)) {
    throw new Error(`expected canonical release URL containing ${releasePath}: ${url}`);
  }
  return url.replace(releasePath, nextReleasePath).replaceAll(currentVersion, versionNoV);
}

pkg.version = versionNoV;
pkg.optimusctx.version = versionNoV;
pkg.optimusctx.releaseTag = releaseTag;

// Preserve the canonical release contract emitted by the Go helper and retag it.
const checksumManifest = pkg.optimusctx.checksumManifest;
checksumManifest.file = checksumManifest.file.replace(currentVersion, versionNoV);
checksumManifest.url = retagCanonicalURL(checksumManifest.url);

for (const platform of Object.values(pkg.optimusctx.platforms)) {
  // The archive contract stays canonical: optimusctx_${versionNoV}_${platform.os}_${platform.arch}
  const extension = platform.os === 'windows' ? 'zip' : 'tar.gz';
  const archive = platform.archive.replace(currentVersion, versionNoV);
  const expectedArchive = `optimusctx_${versionNoV}_${platform.os}_${platform.arch}.${extension}`;
  if (archive !== expectedArchive) {
    throw new Error(`canonical archive contract drifted: expected ${expectedArchive}, got ${archive}`);
  }
  platform.archive = archive;
  platform.archiveFormat = extension;
  platform.archiveUrl = retagCanonicalURL(platform.archiveUrl);
}

fs.writeFileSync(packageJsonPath, `${JSON.stringify(pkg, null, 2)}\n`);
EOF

echo "Rendered @niccrow/optimusctx ${VERSION_NO_V} to ${OUTPUT_DIR}"
