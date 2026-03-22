#!/usr/bin/env bash
set -euo pipefail

[[ -n "${BENCH_TASK_ID:-}" ]] || { printf 'missing BENCH_TASK_ID\n' >&2; exit 1; }
[[ -n "${BENCH_LAST_MESSAGE_FILE:-}" ]] || { printf 'missing BENCH_LAST_MESSAGE_FILE\n' >&2; exit 1; }
[[ -f "${BENCH_LAST_MESSAGE_FILE}" ]] || { printf 'last message file not found: %s\n' "$BENCH_LAST_MESSAGE_FILE" >&2; exit 1; }

message="$(tr '[:upper:]' '[:lower:]' <"$BENCH_LAST_MESSAGE_FILE")"
required_terms=()

case "$BENCH_TASK_ID" in
  list-benchmark-suite-files)
    required_terms=(
      "go-benchmark-discovery-v1.json"
      "go-benchmark-refresh-v1.json"
    )
    ;;
  load-benchmark-suites-validation)
    required_terms=(
      "loadbenchmarksuites"
      "duplicate benchmark suite id"
    )
    ;;
  parse-benchmark-args-common-constraint)
    required_terms=(
      "parsebenchmarkargscommon"
      "exactly one of --suite or --suite-file"
    )
    ;;
  verify-output-flag-restriction)
    required_terms=(
      "verify"
      "--output"
      "does not accept"
    )
    ;;
  *)
    printf 'unsupported BENCH_TASK_ID: %s\n' "$BENCH_TASK_ID" >&2
    exit 1
    ;;
esac

matched_count=0
missing_terms=()
for term in "${required_terms[@]}"; do
  if grep -Fqi -- "$term" <<<"$message"; then
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

if [[ "$required_count" -eq 0 ]]; then
  note="no required terms configured"
else
  note="matched ${matched_count}/${required_count} required terms"
fi
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
    "matched_required_terms": ${matched_count},
    "required_terms": ${required_count},
    "missing_terms": ${missing_json}
  }
}
EOF
