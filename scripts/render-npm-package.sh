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

for required in \
  "${SOURCE_DIR}/package.json" \
  "${SOURCE_DIR}/README.md" \
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

(
  cd "${ROOT_DIR}"
  go run ./cmd/render-npm-package --release-tag "${TAG}" --package-json "${PACKAGE_JSON_PATH}"
)

echo "Rendered @niccrow/optimusctx ${TAG#v} to ${OUTPUT_DIR}"
