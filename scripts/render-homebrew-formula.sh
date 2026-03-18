#!/usr/bin/env bash
set -euo pipefail

if [[ "$#" -ne 3 ]]; then
  echo "usage: scripts/render-homebrew-formula.sh <vX.Y.Z> <checksum-manifest-path> <output-path>" >&2
  exit 1
fi

TAG="$1"
CHECKSUM_PATH="$2"
OUTPUT_PATH="$3"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CALLER_DIR="$(pwd)"

if [[ -z "${TAG}" || -z "${CHECKSUM_PATH}" || -z "${OUTPUT_PATH}" ]]; then
  echo "usage: scripts/render-homebrew-formula.sh <vX.Y.Z> <checksum-manifest-path> <output-path>" >&2
  exit 1
fi

case "${CHECKSUM_PATH}" in
  /*) ;;
  *) CHECKSUM_PATH="${CALLER_DIR}/${CHECKSUM_PATH}" ;;
esac

case "${OUTPUT_PATH}" in
  /*) ;;
  *) OUTPUT_PATH="${CALLER_DIR}/${OUTPUT_PATH}" ;;
esac

(
  cd "${ROOT_DIR}"
  go run ./cmd/render-homebrew-formula --release-tag "${TAG}" --checksum-manifest "${CHECKSUM_PATH}" --output "${OUTPUT_PATH}"
)

echo "Rendered Homebrew formula for ${TAG} to ${OUTPUT_PATH}"
