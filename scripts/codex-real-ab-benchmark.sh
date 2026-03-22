#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  scripts/codex-real-ab-benchmark.sh --prompt-file PATH [--attempts N] [--model MODEL] [--reasoning-effort LEVEL] [--timeout-seconds N] [--sandbox-mode MODE] [--output-dir PATH]

Run a real A/B benchmark against Codex CLI with two isolated arms:
  - baseline: no OptimusCtx AGENTS guidance and no repo-local Codex MCP config
  - optimus: repo-local OptimusCtx AGENTS guidance and .codex/config.toml enabled

The script creates fresh repository copies from git HEAD for every attempt, uses a temporary HOME
so your normal Codex config is not inherited, captures Codex JSONL output, wall-clock timing,
provider-reported token usage, and treatment-side OptimusCtx usage evidence.

Required:
  --prompt-file PATH         File containing the exact prompt to send to both arms

Optional:
  --attempts N               Number of paired attempts to run per arm (default: 3)
  --model MODEL              Codex model to use (default: from current Codex config, else gpt-5.4)
  --reasoning-effort LEVEL   Model reasoning effort override (default: from current Codex config)
  --timeout-seconds N        Timeout per Codex run in seconds (default: 180)
  --sandbox-mode MODE        Codex sandbox mode: read-only, workspace-write, danger-full-access (default: read-only)
  --output-dir PATH          Output directory (default: ./tmp/codex-real-ab/<timestamp>)
  --task-name NAME           Label stored in the summary (default: prompt file basename)
  --help                     Show this help
EOF
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

abspath() {
  local target="$1"
  if [[ -d "$target" ]]; then
    (cd "$target" && pwd)
  else
    (cd "$(dirname "$target")" && printf '%s/%s\n' "$(pwd)" "$(basename "$target")")
  fi
}

copy_repo_head() {
  local dest="$1"
  mkdir -p "$dest"
  git -C "$REPO_ROOT" archive --format=tar "$GIT_SHA" | tar -xf - -C "$dest"
}

prepare_home() {
  local home_dir="$1"
  local model_value="$2"
  local reasoning_value="$3"

  mkdir -p "$home_dir/.codex"
  cp "$AUTH_FILE" "$home_dir/.codex/auth.json"
  if [[ -f "$CREDENTIALS_FILE" ]]; then
    cp "$CREDENTIALS_FILE" "$home_dir/.codex/.credentials.json"
  fi

  {
    if [[ -n "$model_value" ]]; then
      printf 'model = %s\n' "$(json_string "$model_value")"
    fi
    if [[ -n "$reasoning_value" ]]; then
      printf 'model_reasoning_effort = %s\n' "$(json_string "$reasoning_value")"
    fi
  } >"$home_dir/.codex/config.toml"
}

prepare_workspace() {
  local arm="$1"
  local workspace="$2"
  local attempt_dir="$3"

  copy_repo_head "$workspace"
  git -C "$workspace" init -q

  case "$arm" in
    baseline)
      rm -f "$workspace/AGENTS.md"
      rm -f "$workspace/AGENTS.override.md"
      rm -rf "$workspace/.codex"
      ;;
    optimus)
      : >"$attempt_dir/init.stdin"
      (cd "$workspace" && optimusctx init <"$attempt_dir/init.stdin" >"$attempt_dir/init.stdout" 2>"$attempt_dir/init.stderr")
      ;;
    *)
      die "unsupported arm: $arm"
      ;;
  esac
}

json_bool() {
  if [[ "$1" == "true" ]]; then
    printf 'true'
  else
    printf 'false'
  fi
}

json_string() {
  printf '%s' "$1" | jq -Rs .
}

jsonl_has_optimus_mcp_calls() {
  local jsonl_file="$1"

  jq -Rrs '
    split("\n")
    | any(
        fromjson?
        | select(.type == "item.completed" or .type == "item.started")
        | .item
        | select(.type == "mcp_tool_call")
        | .server == "optimusctx"
      )
  ' "$jsonl_file"
}

run_attempt() {
  local arm="$1"
  local attempt="$2"

  local attempt_dir="$RUNS_DIR/attempt-$(printf '%02d' "$attempt")/$arm"
  local workspace="$WORKSPACES_ROOT/attempt-$(printf '%02d' "$attempt")/$arm"
  local home_dir="$attempt_dir/home"
  local jsonl_file="$attempt_dir/codex.jsonl"
  local stderr_file="$attempt_dir/codex.stderr"
  local last_message_file="$attempt_dir/last-message.txt"
  local status_file="$attempt_dir/optimus-status.txt"
  local result_file="$attempt_dir/result.json"
  local prompt_sha256
  local start_ms
  local end_ms
  local elapsed_ms
  local exit_code
  local usage_json
  local input_tokens
  local cached_input_tokens
  local output_tokens
  local used_optimus
  local status_summary
  local result_status
  local -a codex_args

  mkdir -p "$attempt_dir"
  prepare_workspace "$arm" "$workspace" "$attempt_dir"
  prepare_home "$home_dir" "$MODEL" "$REASONING_EFFORT"

  codex_args=(
    exec
    --json
    --ephemeral
    --skip-git-repo-check
    --sandbox "$SANDBOX_MODE"
    --cd "$workspace"
    --output-last-message "$last_message_file"
  )
  if [[ "$arm" == "optimus" ]]; then
    codex_args+=(
      -c 'mcp_servers.optimusctx.command="optimusctx"'
      -c 'mcp_servers.optimusctx.args=["run"]'
    )
  fi

  start_ms="$(date +%s%3N)"
  set +e
  HOME="$home_dir" CODEX_HOME="$home_dir/.codex" \
    timeout --foreground "${TIMEOUT_SECONDS}s" \
      codex "${codex_args[@]}" \
      - <"$PROMPT_FILE" >"$jsonl_file" 2>"$stderr_file"
  exit_code="$?"
  set -e
  end_ms="$(date +%s%3N)"
  elapsed_ms="$((end_ms - start_ms))"

  usage_json="$(
    jq -Rrs '
      split("\n")
      | map(fromjson? | select(.type == "turn.completed"))
      | last
      | .usage // {}
    ' "$jsonl_file"
  )"

  input_tokens="$(jq -r '.input_tokens // 0' <<<"$usage_json")"
  cached_input_tokens="$(jq -r '.cached_input_tokens // 0' <<<"$usage_json")"
  output_tokens="$(jq -r '.output_tokens // 0' <<<"$usage_json")"

  used_optimus="false"
  status_summary=""
  if [[ "$arm" == "optimus" ]]; then
    (cd "$workspace" && optimusctx status >"$status_file")
    if [[ "$(jsonl_has_optimus_mcp_calls "$jsonl_file")" == "true" ]]; then
      used_optimus="true"
    fi
    status_summary="$(tr '\n' ' ' <"$status_file" | sed 's/[[:space:]]\+/ /g' | sed 's/^ //; s/ $//')"
  fi

  if [[ "$exit_code" -eq 0 ]]; then
    result_status="ok"
  else
    result_status="failed"
  fi

  prompt_sha256="$(sha256sum "$PROMPT_FILE" | awk '{print $1}')"

  cat >"$result_file" <<EOF
{
  "task_name": $(json_string "$TASK_NAME"),
  "attempt": $attempt,
  "arm": $(json_string "$arm"),
  "repo_root": $(json_string "$REPO_ROOT"),
  "git_sha": $(json_string "$GIT_SHA"),
  "workspace": $(json_string "$workspace"),
  "prompt_file": $(json_string "$PROMPT_FILE"),
  "prompt_sha256": $(json_string "$prompt_sha256"),
  "model": $(json_string "$MODEL"),
  "reasoning_effort": $(json_string "$REASONING_EFFORT"),
  "sandbox_mode": $(json_string "$SANDBOX_MODE"),
  "status": $(json_string "$result_status"),
  "exit_code": $exit_code,
  "elapsed_ms": $elapsed_ms,
  "input_tokens": $input_tokens,
  "cached_input_tokens": $cached_input_tokens,
  "uncached_input_tokens": $((input_tokens - cached_input_tokens)),
  "output_tokens": $output_tokens,
  "jsonl_file": $(json_string "$jsonl_file"),
  "stderr_file": $(json_string "$stderr_file"),
  "last_message_file": $(json_string "$last_message_file"),
  "optimus_status_file": $(json_string "$status_file"),
  "used_optimus_tools": $(json_bool "$used_optimus"),
  "optimus_status_summary": $(json_string "$status_summary")
}
EOF

  jq -c . "$result_file" >>"$RESULTS_JSONL"

  printf 'attempt %d %s: %s elapsed=%sms input=%s cached=%s output=%s optimus_used=%s\n' \
    "$attempt" "$arm" "$result_status" "$elapsed_ms" "$input_tokens" "$cached_input_tokens" "$output_tokens" "$used_optimus"
}

PROMPT_FILE=""
ATTEMPTS=3
MODEL=""
REASONING_EFFORT=""
TIMEOUT_SECONDS=180
SANDBOX_MODE="read-only"
OUTPUT_DIR=""
TASK_NAME=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --prompt-file)
      [[ $# -ge 2 ]] || die "--prompt-file requires a value"
      PROMPT_FILE="$2"
      shift 2
      ;;
    --attempts)
      [[ $# -ge 2 ]] || die "--attempts requires a value"
      ATTEMPTS="$2"
      shift 2
      ;;
    --model)
      [[ $# -ge 2 ]] || die "--model requires a value"
      MODEL="$2"
      shift 2
      ;;
    --reasoning-effort)
      [[ $# -ge 2 ]] || die "--reasoning-effort requires a value"
      REASONING_EFFORT="$2"
      shift 2
      ;;
    --timeout-seconds)
      [[ $# -ge 2 ]] || die "--timeout-seconds requires a value"
      TIMEOUT_SECONDS="$2"
      shift 2
      ;;
    --sandbox-mode)
      [[ $# -ge 2 ]] || die "--sandbox-mode requires a value"
      SANDBOX_MODE="$2"
      shift 2
      ;;
    --output-dir)
      [[ $# -ge 2 ]] || die "--output-dir requires a value"
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --task-name)
      [[ $# -ge 2 ]] || die "--task-name requires a value"
      TASK_NAME="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

[[ -n "$PROMPT_FILE" ]] || die "--prompt-file is required"
[[ "$ATTEMPTS" =~ ^[1-9][0-9]*$ ]] || die "--attempts must be a positive integer"
[[ "$TIMEOUT_SECONDS" =~ ^[1-9][0-9]*$ ]] || die "--timeout-seconds must be a positive integer"
case "$SANDBOX_MODE" in
  read-only|workspace-write|danger-full-access)
    ;;
  *)
    die "--sandbox-mode must be one of: read-only, workspace-write, danger-full-access"
    ;;
esac

require_cmd codex
require_cmd git
require_cmd jq
require_cmd optimusctx
require_cmd python3
require_cmd sha256sum
require_cmd tar
require_cmd rg
require_cmd timeout

PROMPT_FILE="$(abspath "$PROMPT_FILE")"
[[ -f "$PROMPT_FILE" ]] || die "prompt file not found: $PROMPT_FILE"

REPO_ROOT="$(git rev-parse --show-toplevel)"
GIT_SHA="$(git -C "$REPO_ROOT" rev-parse HEAD)"
AUTH_FILE="${HOME}/.codex/auth.json"
CREDENTIALS_FILE="${HOME}/.codex/.credentials.json"
[[ -f "$AUTH_FILE" ]] || die "missing Codex auth file: $AUTH_FILE"

if [[ -z "$MODEL" ]]; then
  MODEL="$(awk -F'"' '/^model = "/ { print $2; exit }' "${HOME}/.codex/config.toml" 2>/dev/null || true)"
  MODEL="${MODEL:-gpt-5.4}"
fi
if [[ -z "$REASONING_EFFORT" ]]; then
  REASONING_EFFORT="$(awk -F'"' '/^model_reasoning_effort = "/ { print $2; exit }' "${HOME}/.codex/config.toml" 2>/dev/null || true)"
fi
if [[ -z "$TASK_NAME" ]]; then
  TASK_NAME="$(basename "$PROMPT_FILE")"
fi

if [[ -z "$OUTPUT_DIR" ]]; then
  OUTPUT_DIR="$REPO_ROOT/tmp/codex-real-ab/$(date +%Y%m%d-%H%M%S)"
fi
RUNS_DIR="$OUTPUT_DIR/runs"
mkdir -p "$RUNS_DIR"
OUTPUT_DIR="$(abspath "$OUTPUT_DIR")"
RUNS_DIR="$OUTPUT_DIR/runs"
WORKSPACES_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/optimusctx-codex-real-ab-XXXXXX")"

RESULTS_JSONL="$OUTPUT_DIR/results.jsonl"
SUMMARY_JSON="$OUTPUT_DIR/summary.json"
SUMMARY_MD="$OUTPUT_DIR/summary.md"

printf 'task_name=%s\nrepo_root=%s\ngit_sha=%s\nprompt_file=%s\nattempts=%s\nmodel=%s\nreasoning_effort=%s\nsandbox_mode=%s\noutput_dir=%s\n' \
  "$TASK_NAME" "$REPO_ROOT" "$GIT_SHA" "$PROMPT_FILE" "$ATTEMPTS" "$MODEL" "$REASONING_EFFORT" "$SANDBOX_MODE" "$OUTPUT_DIR"
printf 'timeout_seconds=%s\n' "$TIMEOUT_SECONDS"
printf 'workspaces_root=%s\n' "$WORKSPACES_ROOT"

for attempt in $(seq 1 "$ATTEMPTS"); do
  run_attempt baseline "$attempt"
  run_attempt optimus "$attempt"
done

python3 - "$RESULTS_JSONL" "$SUMMARY_JSON" "$SUMMARY_MD" "$TASK_NAME" "$PROMPT_FILE" "$GIT_SHA" "$MODEL" "$REASONING_EFFORT" <<'PY'
import json
import pathlib
import statistics
import sys

results_path = pathlib.Path(sys.argv[1])
summary_json_path = pathlib.Path(sys.argv[2])
summary_md_path = pathlib.Path(sys.argv[3])
task_name = sys.argv[4]
prompt_file = sys.argv[5]
git_sha = sys.argv[6]
model = sys.argv[7]
reasoning_effort = sys.argv[8]

rows = []
for line in results_path.read_text(encoding="utf-8").splitlines():
    line = line.strip()
    if line:
        rows.append(json.loads(line))

if not rows:
    raise SystemExit("no benchmark results recorded")

def median(values):
    if not values:
        return None
    return statistics.median(values)

def success_count(items):
    return sum(1 for item in items if item.get("status") == "ok")

arms = {}
for row in rows:
    arms.setdefault(row["arm"], []).append(row)

summary = {
    "task_name": task_name,
    "prompt_file": prompt_file,
    "git_sha": git_sha,
    "model": model,
    "reasoning_effort": reasoning_effort,
    "attempt_pairs": len({row["attempt"] for row in rows}),
    "rows": rows,
    "arms": {},
    "delta_medians": {},
}

for arm, arm_rows in sorted(arms.items()):
    elapsed_values = [row["elapsed_ms"] for row in arm_rows]
    input_values = [row["input_tokens"] for row in arm_rows]
    cached_values = [row["cached_input_tokens"] for row in arm_rows]
    uncached_values = [row["uncached_input_tokens"] for row in arm_rows]
    output_values = [row["output_tokens"] for row in arm_rows]
    summary["arms"][arm] = {
        "count": len(arm_rows),
        "success_count": success_count(arm_rows),
        "used_optimus_tools_count": sum(1 for row in arm_rows if row.get("used_optimus_tools")),
        "median_elapsed_ms": median(elapsed_values),
        "median_input_tokens": median(input_values),
        "median_cached_input_tokens": median(cached_values),
        "median_uncached_input_tokens": median(uncached_values),
        "median_output_tokens": median(output_values),
    }

baseline = summary["arms"].get("baseline")
optimus = summary["arms"].get("optimus")
if baseline and optimus:
    summary["delta_medians"] = {
        "elapsed_ms": baseline["median_elapsed_ms"] - optimus["median_elapsed_ms"],
        "input_tokens": baseline["median_input_tokens"] - optimus["median_input_tokens"],
        "cached_input_tokens": baseline["median_cached_input_tokens"] - optimus["median_cached_input_tokens"],
        "uncached_input_tokens": baseline["median_uncached_input_tokens"] - optimus["median_uncached_input_tokens"],
        "output_tokens": baseline["median_output_tokens"] - optimus["median_output_tokens"],
    }

summary_json_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")

lines = []
lines.append("# Codex Real A/B Benchmark")
lines.append("")
lines.append(f"- task: `{task_name}`")
lines.append(f"- prompt: `{prompt_file}`")
lines.append(f"- git sha: `{git_sha}`")
lines.append(f"- model: `{model}`")
lines.append(f"- reasoning effort: `{reasoning_effort}`" if reasoning_effort else "- reasoning effort: inherited")
lines.append(f"- paired attempts: `{summary['attempt_pairs']}`")
lines.append("")
lines.append("## Median Metrics")
lines.append("")
lines.append("| arm | success | optimus tool evidence | elapsed ms | input | cached input | uncached input | output |")
lines.append("| --- | --- | --- | ---: | ---: | ---: | ---: | ---: |")
for arm in ("baseline", "optimus"):
    arm_summary = summary["arms"].get(arm)
    if not arm_summary:
      continue
    lines.append(
        f"| {arm} | {arm_summary['success_count']}/{arm_summary['count']} | "
        f"{arm_summary['used_optimus_tools_count']}/{arm_summary['count']} | "
        f"{arm_summary['median_elapsed_ms']} | "
        f"{arm_summary['median_input_tokens']} | "
        f"{arm_summary['median_cached_input_tokens']} | "
        f"{arm_summary['median_uncached_input_tokens']} | "
        f"{arm_summary['median_output_tokens']} |"
    )

if summary["delta_medians"]:
    delta = summary["delta_medians"]
    lines.append("")
    lines.append("## Baseline Minus Optimus")
    lines.append("")
    lines.append(f"- elapsed ms: `{delta['elapsed_ms']}`")
    lines.append(f"- input tokens: `{delta['input_tokens']}`")
    lines.append(f"- cached input tokens: `{delta['cached_input_tokens']}`")
    lines.append(f"- uncached input tokens: `{delta['uncached_input_tokens']}`")
    lines.append(f"- output tokens: `{delta['output_tokens']}`")

lines.append("")
lines.append("## Per Attempt")
lines.append("")
lines.append("| attempt | arm | status | used optimus tools | elapsed ms | input | cached input | uncached input | output |")
lines.append("| ---: | --- | --- | --- | ---: | ---: | ---: | ---: | ---: |")
for row in sorted(rows, key=lambda item: (item["attempt"], item["arm"])):
    lines.append(
        f"| {row['attempt']} | {row['arm']} | {row['status']} | {str(row['used_optimus_tools']).lower()} | "
        f"{row['elapsed_ms']} | {row['input_tokens']} | {row['cached_input_tokens']} | "
        f"{row['uncached_input_tokens']} | {row['output_tokens']} |"
    )

summary_md_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
PY

printf 'summary_json=%s\nsummary_md=%s\nresults_jsonl=%s\n' "$SUMMARY_JSON" "$SUMMARY_MD" "$RESULTS_JSONL"
