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
ARCHIVE_NAME_TEMPLATE='optimusctx_${versionNoV}_${goos}_${goarch}'

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

PACKAGE_JSON_PATH="${PACKAGE_JSON_PATH}" RELEASE_TAG="${TAG}" VERSION_NO_V="${VERSION_NO_V}" ARCHIVE_NAME_TEMPLATE="${ARCHIVE_NAME_TEMPLATE}" node <<'EOF'
const fs = require('node:fs');

const packageJsonPath = process.env.PACKAGE_JSON_PATH;
const releaseTag = process.env.RELEASE_TAG;
const versionNoV = process.env.VERSION_NO_V;
const archiveNameTemplate = process.env.ARCHIVE_NAME_TEMPLATE;

if (!packageJsonPath || !releaseTag || !versionNoV) {
  throw new Error('package.json rendering requires PACKAGE_JSON_PATH, RELEASE_TAG, and VERSION_NO_V');
}

const pkg = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));

// npm package metadata must stay release-derived and use the same
// optimusctx_${versionNoV}_${goos}_${goarch} naming contract as GoReleaser.
const repositoryURL = `https://github.com/${pkg.optimusctx.repository.owner}/${pkg.optimusctx.repository.name}`;
const checksumFile = `optimusctx_${versionNoV}_checksums.txt`;

pkg.version = versionNoV;
pkg.optimusctx.version = versionNoV;
pkg.optimusctx.releaseTag = releaseTag;
pkg.optimusctx.checksumManifest.file = checksumFile;
pkg.optimusctx.checksumManifest.url = `${repositoryURL}/releases/download/${releaseTag}/${checksumFile}`;

for (const platform of Object.values(pkg.optimusctx.platforms)) {
  const extension = platform.os === 'windows' ? 'zip' : 'tar.gz';
  const archive = `optimusctx_${versionNoV}_${platform.os}_${platform.arch}.${extension}`;
  platform.archive = archive;
  platform.archiveFormat = extension;
  platform.archiveUrl = `${repositoryURL}/releases/download/${releaseTag}/${archive}`;
}

fs.writeFileSync(packageJsonPath, `${JSON.stringify(pkg, null, 2)}\n`);
EOF

echo "Rendered @niccrow/optimusctx ${VERSION_NO_V} to ${OUTPUT_DIR}"
