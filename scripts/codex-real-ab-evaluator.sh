#!/usr/bin/env bash
set -euo pipefail

[[ -n "${BENCH_TASK_ID:-}" ]] || { printf 'missing BENCH_TASK_ID\n' >&2; exit 1; }
[[ -n "${BENCH_WORKSPACE:-}" ]] || { printf 'missing BENCH_WORKSPACE\n' >&2; exit 1; }
[[ -d "${BENCH_WORKSPACE}" ]] || { printf 'workspace not found: %s\n' "$BENCH_WORKSPACE" >&2; exit 1; }

lower_file() {
  tr '[:upper:]' '[:lower:]' <"$1"
}

target_file=""
required_terms=()

case "$BENCH_TASK_ID" in
  inspect-suite-loading)
    target_file="$BENCH_WORKSPACE/.benchmark/benchmark-contract-audit.md"
    required_terms=(
      "## suite loading"
      "internal/repository/benchmark.go"
      "internal/app/benchmark_runner.go"
      "loadbenchmarksuite"
      "loadbenchmarksuites"
      "validatebenchmarkfixturereferences"
      "duplicate benchmark suite id"
    )
    ;;
  trace-evidence-bundle)
    target_file="$BENCH_WORKSPACE/.benchmark/benchmark-contract-audit.md"
    required_terms=(
      "## evidence bundle"
      "internal/app/benchmark_service.go"
      "buildbenchmarkevidencebundle"
      "internal/app/benchmark_service_test.go"
      "buildbenchmarkhumansummary"
      "methodology drift"
    )
    ;;
  propose-benchmark-workflows)
    target_file="$BENCH_WORKSPACE/.benchmark/benchmark-contract-audit.md"
    required_terms=(
      "## benchmark workflow candidates"
      "1."
      "2."
      "3."
      ".benchmark/"
      "navigation"
      "bug triage"
      "edit verification"
      "persisted runs"
      "final artifact"
    )
    ;;
  map-canonical-release-contract)
    target_file="$BENCH_WORKSPACE/.benchmark/release-contract-audit.md"
    required_terms=(
      "## canonical release contract"
      ".github/workflows/release.yml"
      "docs/operator-release-guide.md"
      "internal/release/release_test.go"
      "github release remains the canonical root and rollback source"
      "workflow_dispatch"
      "release_tag"
      "publication_channel"
    )
    ;;
  trace-downstream-rerun-rules)
    target_file="$BENCH_WORKSPACE/.benchmark/release-contract-audit.md"
    required_terms=(
      "## downstream rerun rules"
      "npm"
      "homebrew"
      "scoop"
      "publication_status=already_current"
      "publication_status=not_published"
      "workflow_dispatch"
    )
    ;;
  propose-release-benchmark-flow)
    target_file="$BENCH_WORKSPACE/.benchmark/release-contract-audit.md"
    required_terms=(
      "## release benchmark flow"
      "1."
      "2."
      "3."
      ".benchmark/"
      "sequential"
      "cross-file"
      "canonical release"
      "downstream rerun"
    )
    ;;
  *)
    printf 'unsupported BENCH_TASK_ID: %s\n' "$BENCH_TASK_ID" >&2
    exit 1
    ;;
esac

[[ -n "$target_file" ]] || { printf 'no target file for %s\n' "$BENCH_TASK_ID" >&2; exit 1; }
if [[ ! -f "$target_file" ]]; then
  cat <<EOF
{
  "pass": false,
  "score": 0,
  "note": "expected artifact file was not created: $target_file",
  "metrics": {
    "artifact_exists": false,
    "matched_required_terms": 0,
    "required_terms": ${#required_terms[@]},
    "missing_terms": $(printf '%s\n' "${required_terms[@]}" | jq -Rsc 'split("\n") | map(select(length > 0))')
  }
}
EOF
  exit 0
fi

content="$(lower_file "$target_file")"
matched_count=0
missing_terms=()
for term in "${required_terms[@]}"; do
  normalized_term="$(printf '%s' "$term" | tr '[:upper:]' '[:lower:]')"
  if grep -Fqi -- "$normalized_term" <<<"$content"; then
    matched_count=$((matched_count + 1))
  else
    missing_terms+=("$term")
  fi
done

required_count="${#required_terms[@]}"
score=0
passed="false"
if [[ "$matched_count" -eq "$required_count" ]]; then
  score=2
  passed="true"
elif [[ "$matched_count" -gt 0 ]]; then
  score=1
fi

note="matched ${matched_count}/${required_count} required terms in $(basename "$target_file")"
if [[ "${#missing_terms[@]}" -gt 0 ]]; then
  note="${note}; missing: $(IFS=', '; printf '%s' "${missing_terms[*]}")"
fi

missing_json="$(printf '%s\n' "${missing_terms[@]}" | jq -Rsc 'split("\n") | map(select(length > 0))')"

cat <<EOF
{
  "pass": ${passed},
  "score": ${score},
  "note": $(printf '%s' "$note" | jq -Rs .),
  "metrics": {
    "artifact_exists": true,
    "artifact_file": $(printf '%s' "$target_file" | jq -Rs .),
    "matched_required_terms": ${matched_count},
    "required_terms": ${required_count},
    "missing_terms": ${missing_json}
  }
}
EOF
