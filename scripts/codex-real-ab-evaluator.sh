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
  map-final-artifact-failure-path)
    target_file="$BENCH_WORKSPACE/.benchmark/benchmark-failure-triage.md"
    required_terms=(
      "## failure path"
      "internal/app/benchmark_runner.go"
      "internal/app/benchmark_service.go"
      "internal/app/benchmark_service_test.go"
      "final artifact"
      "normalized final artifact is missing"
      "methodology drift"
    )
    ;;
  identify-regression-surface)
    target_file="$BENCH_WORKSPACE/.benchmark/benchmark-failure-triage.md"
    required_terms=(
      "## regression surface"
      "1."
      "2."
      "3."
      "comparebenchmarkevidencebundles"
      "buildbenchmarkhumansummary"
      "final artifact"
      "invalid run"
      "methodology"
    )
    ;;
  propose-fix-and-verification)
    target_file="$BENCH_WORKSPACE/.benchmark/benchmark-failure-triage.md"
    required_terms=(
      "## fix plan"
      "1."
      "2."
      "3."
      ".benchmark/"
      "bug triage"
      "edit verification"
      "go test"
      "final artifact"
    )
    ;;
  map-rerun-contract-surface)
    target_file="$BENCH_WORKSPACE/.benchmark/release-rerun-change-plan.md"
    required_terms=(
      "## contract surface"
      ".github/workflows/release.yml"
      "docs/operator-release-guide.md"
      "docs/release-checklist.md"
      "internal/release/release_test.go"
      "workflow_dispatch"
      "release_tag"
      "publication_channel"
    )
    ;;
  draft-change-plan)
    target_file="$BENCH_WORKSPACE/.benchmark/release-rerun-change-plan.md"
    required_terms=(
      "## change plan"
      "1."
      "2."
      "3."
      "canonical release"
      "downstream rerun"
      "publication_status=already_current"
      "publication_status=not_published"
      "workflow_dispatch"
    )
    ;;
  draft-verification-plan)
    target_file="$BENCH_WORKSPACE/.benchmark/release-rerun-change-plan.md"
    required_terms=(
      "## verification plan"
      "1."
      "2."
      "3."
      ".benchmark/"
      "edit verification"
      "go test ./internal/release"
      "sequential"
      "cross-file"
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
