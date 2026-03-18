#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG="${1:-}"
CHECKSUM_PATH="${2:-}"
OUTPUT_PATH="${3:-}"

if [[ -z "${TAG}" || -z "${CHECKSUM_PATH}" || -z "${OUTPUT_PATH}" ]]; then
  echo "usage: scripts/render-scoop-manifest.sh <vX.Y.Z> <checksum-manifest-path> <output-path>" >&2
  exit 1
fi

case "${TAG}" in
  v*) ;;
  *)
    echo "tag must start with v" >&2
    exit 1
    ;;
esac

if [[ ! -f "${CHECKSUM_PATH}" ]]; then
  echo "checksum manifest not found: ${CHECKSUM_PATH}" >&2
  exit 1
fi

(
  cd "${ROOT_DIR}"
  go run ./cmd/render-scoop-manifest --release-tag "${TAG}" --checksum-manifest "${CHECKSUM_PATH}" --output "${OUTPUT_PATH}"
)

echo "Rendered Scoop manifest for ${TAG} to ${OUTPUT_PATH}"
