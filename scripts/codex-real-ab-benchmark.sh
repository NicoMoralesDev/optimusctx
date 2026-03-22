#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  scripts/codex-real-ab-benchmark.sh (--prompt-file PATH | --suite-file PATH) [options]

Run a real Codex A/B benchmark with paired baseline vs optimus executions.

Arms:
  - baseline: no OptimusCtx guidance and no repo-local Codex MCP config
  - optimus: project prepared with `optimusctx init --client codex-cli --write`

The benchmark treats project setup as out-of-band preparation. Timed execution starts only when
Codex begins handling the task. Results capture Codex JSONL output, wall-clock timing, token
usage, workspace diffs, and optional task-quality evaluations.

Inputs:
  --prompt-file PATH         Single prompt file to benchmark
  --suite-file PATH          JSON suite file describing multiple benchmark tasks

Optional:
  --attempts N               Paired attempts per task (default: 5)
  --model MODEL              Codex model to use (default: from current Codex config, else gpt-5.4)
  --reasoning-effort LEVEL   Model reasoning effort override (default: from current Codex config)
  --timeout-seconds N        Timeout per Codex run in seconds (default: 180)
  --sandbox-mode MODE        Codex sandbox mode: read-only, workspace-write, danger-full-access (default: read-only)
  --output-dir PATH          Output directory (default: ./tmp/codex-real-ab/<timestamp>)
  --task-name NAME           Label for a single prompt benchmark (default: prompt filename)
  --task-category NAME       Category for a single prompt benchmark (default: uncategorized)
  --evaluator-command CMD    Shell command that emits JSON quality results to stdout
  --order-mode MODE          baseline-first, optimus-first, alternating, paired-random (default: paired-random)
  --random-seed N            Seed used for execution ordering (default: current epoch seconds)
  --help                     Show this help

Suite file schema:
  {
    "schema_version": 1,
    "task_defaults": {
      "category": "analysis",
      "evaluator_command": "./scripts/evaluate-task.sh"
    },
    "tasks": [
      {
        "id": "locate-symbol",
        "name": "Locate repository symbol",
        "prompt_file": "bench/prompts/locate-symbol.txt",
        "category": "navigation",
        "evaluator_command": "./scripts/evaluate-locate-symbol.sh"
      }
    ]
  }

Evaluator contract:
  The evaluator command runs with benchmark metadata exposed as BENCH_* environment variables and
  must print a JSON object. Supported fields:
    {"pass": true, "score": 2, "note": "concise explanation", "metrics": {"matched_files": 3}}
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

json_string() {
  printf '%s' "$1" | jq -Rs .
}

json_bool() {
  if [[ "$1" == "true" ]]; then
    printf 'true'
  else
    printf 'false'
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
  local extra_config_file="${4:-}"
  local home_config_file="$home_dir/.codex/config.toml"

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
  } >"$home_config_file"

  # `codex exec` reads the active config from CODEX_HOME/HOME. When the benchmark
  # stages a repo-local `.codex/config.toml` for the optimus arm, merge that MCP
  # block into the effective home config so the run actually sees the server.
  if [[ -n "$extra_config_file" && -f "$extra_config_file" ]]; then
    printf '\n' >>"$home_config_file"
    cat "$extra_config_file" >>"$home_config_file"
  fi
}

prepare_workspace() {
  local arm="$1"
  local workspace="$2"
  local attempt_dir="$3"

  copy_repo_head "$workspace"
  git -C "$workspace" init -q
  git -C "$workspace" config user.email 'benchmark@optimusctx.local'
  git -C "$workspace" config user.name 'OptimusCtx Benchmark'

  case "$arm" in
    baseline)
      rm -f "$workspace/AGENTS.md"
      rm -f "$workspace/AGENTS.override.md"
      rm -rf "$workspace/.codex"
      ;;
    optimus)
      mkdir -p "$workspace/.codex"
      (
        cd "$workspace" && \
          optimusctx init \
            --client codex-cli \
            --config "$workspace/.codex/config.toml" \
            --write \
            >"$attempt_dir/init.stdout" \
            2>"$attempt_dir/init.stderr"
      )
      ;;
    *)
      die "unsupported arm: $arm"
      ;;
  esac

  git -C "$workspace" add -A
  git -C "$workspace" commit -q --allow-empty -m 'benchmark snapshot'
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

build_plan() {
  PLAN_JSONL="$OUTPUT_DIR/plan.jsonl"
  TASKS_JSON="$OUTPUT_DIR/tasks.json"

  PROMPT_FILE="$PROMPT_FILE" \
  SUITE_FILE="$SUITE_FILE" \
  TASK_NAME="$TASK_NAME" \
  TASK_CATEGORY="$TASK_CATEGORY" \
  EVALUATOR_COMMAND="$EVALUATOR_COMMAND" \
  ATTEMPTS="$ATTEMPTS" \
  ORDER_MODE="$ORDER_MODE" \
  RANDOM_SEED="$RANDOM_SEED" \
  PLAN_JSONL="$PLAN_JSONL" \
  TASKS_JSON="$TASKS_JSON" \
  python3 - <<'PY'
import json
import os
import pathlib
import random
import re
def slugify(value):
    slug = re.sub(r"[^a-z0-9]+", "-", value.lower()).strip("-")
    return slug or "task"


def unique_slug(base, seen):
    candidate = slugify(base)
    if candidate not in seen:
        seen.add(candidate)
        return candidate
    suffix = 2
    while f"{candidate}-{suffix}" in seen:
        suffix += 1
    candidate = f"{candidate}-{suffix}"
    seen.add(candidate)
    return candidate


prompt_file = os.environ.get("PROMPT_FILE", "")
suite_file = os.environ.get("SUITE_FILE", "")
task_name = os.environ.get("TASK_NAME", "")
task_category = os.environ.get("TASK_CATEGORY", "")
evaluator_command = os.environ.get("EVALUATOR_COMMAND", "")
attempts = int(os.environ["ATTEMPTS"])
order_mode = os.environ["ORDER_MODE"]
seed = int(os.environ["RANDOM_SEED"])
plan_path = pathlib.Path(os.environ["PLAN_JSONL"])
tasks_path = pathlib.Path(os.environ["TASKS_JSON"])

if bool(prompt_file) == bool(suite_file):
    raise SystemExit("exactly one of PROMPT_FILE or SUITE_FILE must be set")

tasks = []
seen_slugs = set()

if prompt_file:
    prompt_path = pathlib.Path(prompt_file).resolve()
    if not prompt_path.is_file():
        raise SystemExit(f"prompt file not found: {prompt_path}")
    resolved_task_name = task_name or prompt_path.name
    tasks.append(
        {
            "id": slugify(resolved_task_name),
            "name": resolved_task_name,
            "slug": unique_slug(resolved_task_name, seen_slugs),
            "prompt_file": str(prompt_path),
            "category": task_category or "uncategorized",
            "evaluator_command": evaluator_command,
            "source": "prompt-file",
        }
    )
else:
    suite_path = pathlib.Path(suite_file).resolve()
    if not suite_path.is_file():
        raise SystemExit(f"suite file not found: {suite_path}")
    payload = json.loads(suite_path.read_text(encoding="utf-8"))
    if payload.get("schema_version", 1) != 1:
        raise SystemExit("suite schema_version must be 1")
    raw_tasks = payload.get("tasks")
    if not isinstance(raw_tasks, list) or not raw_tasks:
        raise SystemExit("suite must define a non-empty tasks array")
    defaults = payload.get("task_defaults") or {}
    if defaults and not isinstance(defaults, dict):
        raise SystemExit("suite task_defaults must be an object")

    for idx, raw_task in enumerate(raw_tasks, start=1):
        if not isinstance(raw_task, dict):
            raise SystemExit(f"tasks[{idx}] must be an object")
        prompt_value = raw_task.get("prompt_file")
        if not prompt_value:
          raise SystemExit(f"tasks[{idx}] missing prompt_file")
        prompt_path = pathlib.Path(prompt_value)
        if not prompt_path.is_absolute():
            prompt_path = (suite_path.parent / prompt_path).resolve()
        if not prompt_path.is_file():
            raise SystemExit(f"tasks[{idx}] prompt_file not found: {prompt_path}")
        resolved_name = str(raw_task.get("name") or raw_task.get("id") or prompt_path.stem)
        resolved_id = str(raw_task.get("id") or slugify(resolved_name))
        tasks.append(
            {
                "id": resolved_id,
                "name": resolved_name,
                "slug": unique_slug(str(raw_task.get("id") or resolved_name), seen_slugs),
                "prompt_file": str(prompt_path),
                "category": str(raw_task.get("category") or defaults.get("category") or "uncategorized"),
                "evaluator_command": str(raw_task.get("evaluator_command") or defaults.get("evaluator_command") or evaluator_command),
                "source": "suite-file",
            }
        )

rng = random.Random(seed)
plan_rows = []
for task_index, task in enumerate(tasks, start=1):
    for attempt in range(1, attempts + 1):
        arm_order = ["baseline", "optimus"]
        if order_mode == "baseline-first":
            arm_order = ["baseline", "optimus"]
        elif order_mode == "optimus-first":
            arm_order = ["optimus", "baseline"]
        elif order_mode == "alternating":
            if (task_index + attempt) % 2 == 0:
                arm_order = ["optimus", "baseline"]
        elif order_mode == "paired-random":
            arm_order = ["baseline", "optimus"]
            rng.shuffle(arm_order)
        else:
            raise SystemExit(f"unsupported order mode: {order_mode}")

        pair_key = f"{task['id']}::attempt-{attempt:02d}"
        for order_index, arm in enumerate(arm_order, start=1):
            plan_rows.append(
                {
                    "task_id": task["id"],
                    "task_name": task["name"],
                    "task_slug": task["slug"],
                    "task_category": task["category"],
                    "task_prompt_file": task["prompt_file"],
                    "evaluator_command": task["evaluator_command"],
                    "attempt": attempt,
                    "pair_key": pair_key,
                    "task_index": task_index,
                    "arm": arm,
                    "arm_order": order_index,
                    "pair_execution_order": arm_order,
                }
            )

tasks_payload = {
    "schema_version": 1,
    "task_count": len(tasks),
    "attempts_per_task": attempts,
    "order_mode": order_mode,
    "random_seed": seed,
    "tasks": tasks,
}
tasks_path.write_text(json.dumps(tasks_payload, indent=2) + "\n", encoding="utf-8")
plan_path.write_text(
    "".join(json.dumps(row) + "\n" for row in plan_rows),
    encoding="utf-8",
)
PY
}

evaluate_attempt() {
  local attempt_dir="$1"
  local evaluator_command="$2"
  local task_id="$3"
  local task_name="$4"
  local task_category="$5"
  local prompt_file="$6"
  local arm="$7"
  local attempt="$8"
  local pair_key="$9"
  local arm_order="${10}"
  local workspace="${11}"
  local jsonl_file="${12}"
  local stderr_file="${13}"
  local last_message_file="${14}"
  local diff_file="${15}"
  local workspace_status_file="${16}"
  local run_status="${17}"
  local run_exit_code="${18}"

  QUALITY_STATUS="not_evaluated"
  QUALITY_PASS_JSON="null"
  QUALITY_SCORE_JSON="null"
  QUALITY_NOTE=""
  QUALITY_METRICS_JSON="null"
  QUALITY_EVALUATOR_EXIT_CODE="null"
  QUALITY_STDOUT_FILE="$attempt_dir/evaluator.stdout.json"
  QUALITY_STDERR_FILE="$attempt_dir/evaluator.stderr"

  if [[ -z "$evaluator_command" ]]; then
    return 0
  fi

  local evaluator_exit_code
  set +e
  BENCH_REPO_ROOT="$REPO_ROOT" \
  BENCH_TASK_ID="$task_id" \
  BENCH_TASK_NAME="$task_name" \
  BENCH_TASK_CATEGORY="$task_category" \
  BENCH_PROMPT_FILE="$prompt_file" \
  BENCH_ARM="$arm" \
  BENCH_ATTEMPT="$attempt" \
  BENCH_PAIR_KEY="$pair_key" \
  BENCH_ARM_ORDER="$arm_order" \
  BENCH_WORKSPACE="$workspace" \
  BENCH_OUTPUT_DIR="$attempt_dir" \
  BENCH_JSONL_FILE="$jsonl_file" \
  BENCH_STDERR_FILE="$stderr_file" \
  BENCH_LAST_MESSAGE_FILE="$last_message_file" \
  BENCH_DIFF_FILE="$diff_file" \
  BENCH_WORKSPACE_STATUS_FILE="$workspace_status_file" \
  BENCH_RUN_STATUS="$run_status" \
  BENCH_RUN_EXIT_CODE="$run_exit_code" \
    bash -lc "$evaluator_command" >"$QUALITY_STDOUT_FILE" 2>"$QUALITY_STDERR_FILE"
  evaluator_exit_code="$?"
  set -e

  QUALITY_EVALUATOR_EXIT_CODE="$evaluator_exit_code"
  if [[ "$evaluator_exit_code" -ne 0 ]]; then
    QUALITY_STATUS="error"
    QUALITY_NOTE="$(tr '\n' ' ' <"$QUALITY_STDERR_FILE" | sed 's/[[:space:]]\+/ /g' | sed 's/^ //; s/ $//')"
    return 0
  fi

  if ! jq -e 'type == "object"' "$QUALITY_STDOUT_FILE" >/dev/null 2>&1; then
    QUALITY_STATUS="error"
    QUALITY_NOTE="evaluator output was not a JSON object"
    return 0
  fi

  QUALITY_STATUS="ok"
  QUALITY_PASS_JSON="$(jq -r 'if has("pass") then (.pass | if . == null then "null" elif . then "true" else "false" end) else "null" end' "$QUALITY_STDOUT_FILE")"
  QUALITY_SCORE_JSON="$(jq -r 'if has("score") and .score != null then (.score | tostring) else "null" end' "$QUALITY_STDOUT_FILE")"
  QUALITY_NOTE="$(jq -r '.note // ""' "$QUALITY_STDOUT_FILE")"
  QUALITY_METRICS_JSON="$(jq -c 'if has("metrics") then .metrics else null end' "$QUALITY_STDOUT_FILE")"
}

run_plan_row() {
  local row_json="$1"

  local task_id
  local task_name
  local task_slug
  local task_category
  local prompt_file
  local evaluator_command
  local attempt
  local pair_key
  local task_index
  local arm
  local arm_order
  local pair_order_json
  task_id="$(jq -r '.task_id' <<<"$row_json")"
  task_name="$(jq -r '.task_name' <<<"$row_json")"
  task_slug="$(jq -r '.task_slug' <<<"$row_json")"
  task_category="$(jq -r '.task_category' <<<"$row_json")"
  prompt_file="$(jq -r '.task_prompt_file' <<<"$row_json")"
  evaluator_command="$(jq -r '.evaluator_command // ""' <<<"$row_json")"
  attempt="$(jq -r '.attempt' <<<"$row_json")"
  pair_key="$(jq -r '.pair_key' <<<"$row_json")"
  task_index="$(jq -r '.task_index' <<<"$row_json")"
  arm="$(jq -r '.arm' <<<"$row_json")"
  arm_order="$(jq -r '.arm_order' <<<"$row_json")"
  pair_order_json="$(jq -c '.pair_execution_order' <<<"$row_json")"

  local pair_dir="$RUNS_DIR/$task_slug/attempt-$(printf '%02d' "$attempt")"
  local attempt_dir="$pair_dir/$arm"
  local workspace="$WORKSPACES_ROOT/$task_slug/attempt-$(printf '%02d' "$attempt")/$arm"
  local home_dir="$attempt_dir/home"
  local jsonl_file="$attempt_dir/codex.jsonl"
  local stderr_file="$attempt_dir/codex.stderr"
  local last_message_file="$attempt_dir/last-message.txt"
  local status_file="$attempt_dir/optimus-status.txt"
  local diff_file="$attempt_dir/workspace.diff"
  local workspace_status_file="$attempt_dir/workspace.status"
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
  local uncached_input_tokens
  local changed_files
  local completion_status
  local effective_config_file
  local workspace_config_file
  local -a codex_args

  mkdir -p "$attempt_dir"
  printf '%s\n' "$row_json" >"$attempt_dir/plan-row.json"
  prepare_workspace "$arm" "$workspace" "$attempt_dir"
  workspace_config_file="$workspace/.codex/config.toml"
  effective_config_file="$home_dir/.codex/config.toml"
  if [[ "$arm" == "optimus" ]]; then
    prepare_home "$home_dir" "$MODEL" "$REASONING_EFFORT" "$workspace_config_file"
    grep -q '^\[mcp_servers\.optimusctx\]$' "$effective_config_file" || die "optimus config was not merged into effective Codex config: $effective_config_file"
  else
    prepare_home "$home_dir" "$MODEL" "$REASONING_EFFORT"
  fi

  codex_args=(
    exec
    --json
    --ephemeral
    --skip-git-repo-check
    --sandbox "$SANDBOX_MODE"
    --cd "$workspace"
    --output-last-message "$last_message_file"
  )

  start_ms="$(date +%s%3N)"
  set +e
  HOME="$home_dir" CODEX_HOME="$home_dir/.codex" \
    timeout --foreground "${TIMEOUT_SECONDS}s" \
      codex "${codex_args[@]}" \
      - <"$prompt_file" >"$jsonl_file" 2>"$stderr_file"
  exit_code="$?"
  set -e
  end_ms="$(date +%s%3N)"
  elapsed_ms="$((end_ms - start_ms))"

  git -C "$workspace" diff --binary HEAD -- >"$diff_file"
  git -C "$workspace" status --short >"$workspace_status_file"
  changed_files="$(git -C "$workspace" status --short | wc -l | awk '{print $1}')"

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
  uncached_input_tokens="$((input_tokens - cached_input_tokens))"

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
  completion_status="$result_status"

  prompt_sha256="$(sha256sum "$prompt_file" | awk '{print $1}')"

  evaluate_attempt \
    "$attempt_dir" \
    "$evaluator_command" \
    "$task_id" \
    "$task_name" \
    "$task_category" \
    "$prompt_file" \
    "$arm" \
    "$attempt" \
    "$pair_key" \
    "$arm_order" \
    "$workspace" \
    "$jsonl_file" \
    "$stderr_file" \
    "$last_message_file" \
    "$diff_file" \
    "$workspace_status_file" \
    "$completion_status" \
    "$exit_code"

  cat >"$result_file" <<EOF
{
  "schema_version": 2,
  "task_id": $(json_string "$task_id"),
  "task_name": $(json_string "$task_name"),
  "task_category": $(json_string "$task_category"),
  "task_index": $task_index,
  "attempt": $attempt,
  "pair_key": $(json_string "$pair_key"),
  "arm": $(json_string "$arm"),
  "arm_order": $arm_order,
  "pair_execution_order": $pair_order_json,
  "repo_root": $(json_string "$REPO_ROOT"),
  "git_sha": $(json_string "$GIT_SHA"),
  "workspace": $(json_string "$workspace"),
  "effective_codex_config_file": $(json_string "$effective_config_file"),
  "prompt_file": $(json_string "$prompt_file"),
  "prompt_sha256": $(json_string "$prompt_sha256"),
  "model": $(json_string "$MODEL"),
  "reasoning_effort": $(json_string "$REASONING_EFFORT"),
  "sandbox_mode": $(json_string "$SANDBOX_MODE"),
  "status": $(json_string "$result_status"),
  "exit_code": $exit_code,
  "elapsed_ms": $elapsed_ms,
  "input_tokens": $input_tokens,
  "cached_input_tokens": $cached_input_tokens,
  "uncached_input_tokens": $uncached_input_tokens,
  "output_tokens": $output_tokens,
  "changed_files": $changed_files,
  "jsonl_file": $(json_string "$jsonl_file"),
  "stderr_file": $(json_string "$stderr_file"),
  "last_message_file": $(json_string "$last_message_file"),
  "workspace_diff_file": $(json_string "$diff_file"),
  "workspace_status_file": $(json_string "$workspace_status_file"),
  "optimus_status_file": $(json_string "$status_file"),
  "used_optimus_tools": $(json_bool "$used_optimus"),
  "optimus_status_summary": $(json_string "$status_summary"),
  "quality_evaluation_status": $(json_string "$QUALITY_STATUS"),
  "quality_pass": $QUALITY_PASS_JSON,
  "quality_score": $QUALITY_SCORE_JSON,
  "quality_note": $(json_string "$QUALITY_NOTE"),
  "quality_metrics": $QUALITY_METRICS_JSON,
  "evaluator_command": $(json_string "$evaluator_command"),
  "evaluator_exit_code": $QUALITY_EVALUATOR_EXIT_CODE,
  "evaluator_stdout_file": $(json_string "$QUALITY_STDOUT_FILE"),
  "evaluator_stderr_file": $(json_string "$QUALITY_STDERR_FILE")
}
EOF

  jq -c . "$result_file" >>"$RESULTS_JSONL"

  printf 'task=%s attempt=%d arm=%s order=%d status=%s elapsed=%sms uncached_input=%s output=%s pass=%s score=%s optimus_used=%s\n' \
    "$task_id" "$attempt" "$arm" "$arm_order" "$result_status" "$elapsed_ms" "$uncached_input_tokens" "$output_tokens" "$QUALITY_PASS_JSON" "$QUALITY_SCORE_JSON" "$used_optimus"
}

PROMPT_FILE=""
SUITE_FILE=""
ATTEMPTS=5
MODEL=""
REASONING_EFFORT=""
TIMEOUT_SECONDS=180
SANDBOX_MODE="read-only"
OUTPUT_DIR=""
TASK_NAME=""
TASK_CATEGORY="uncategorized"
EVALUATOR_COMMAND=""
ORDER_MODE="paired-random"
RANDOM_SEED="$(date +%s)"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --prompt-file)
      [[ $# -ge 2 ]] || die "--prompt-file requires a value"
      PROMPT_FILE="$2"
      shift 2
      ;;
    --suite-file)
      [[ $# -ge 2 ]] || die "--suite-file requires a value"
      SUITE_FILE="$2"
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
    --task-category)
      [[ $# -ge 2 ]] || die "--task-category requires a value"
      TASK_CATEGORY="$2"
      shift 2
      ;;
    --evaluator-command)
      [[ $# -ge 2 ]] || die "--evaluator-command requires a value"
      EVALUATOR_COMMAND="$2"
      shift 2
      ;;
    --order-mode)
      [[ $# -ge 2 ]] || die "--order-mode requires a value"
      ORDER_MODE="$2"
      shift 2
      ;;
    --random-seed)
      [[ $# -ge 2 ]] || die "--random-seed requires a value"
      RANDOM_SEED="$2"
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

if [[ -n "$PROMPT_FILE" && -n "$SUITE_FILE" ]]; then
  die "use either --prompt-file or --suite-file, not both"
fi
if [[ -z "$PROMPT_FILE" && -z "$SUITE_FILE" ]]; then
  die "one of --prompt-file or --suite-file is required"
fi

[[ "$ATTEMPTS" =~ ^[1-9][0-9]*$ ]] || die "--attempts must be a positive integer"
[[ "$TIMEOUT_SECONDS" =~ ^[1-9][0-9]*$ ]] || die "--timeout-seconds must be a positive integer"
[[ "$RANDOM_SEED" =~ ^[0-9]+$ ]] || die "--random-seed must be a non-negative integer"
case "$SANDBOX_MODE" in
  read-only|workspace-write|danger-full-access)
    ;;
  *)
    die "--sandbox-mode must be one of: read-only, workspace-write, danger-full-access"
    ;;
esac
case "$ORDER_MODE" in
  baseline-first|optimus-first|alternating|paired-random)
    ;;
  *)
    die "--order-mode must be one of: baseline-first, optimus-first, alternating, paired-random"
    ;;
esac

require_cmd codex
require_cmd git
require_cmd jq
require_cmd optimusctx
require_cmd python3
require_cmd sha256sum
require_cmd tar
require_cmd timeout

if [[ -n "$PROMPT_FILE" ]]; then
  PROMPT_FILE="$(abspath "$PROMPT_FILE")"
  [[ -f "$PROMPT_FILE" ]] || die "prompt file not found: $PROMPT_FILE"
fi
if [[ -n "$SUITE_FILE" ]]; then
  SUITE_FILE="$(abspath "$SUITE_FILE")"
  [[ -f "$SUITE_FILE" ]] || die "suite file not found: $SUITE_FILE"
fi

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
if [[ -z "$TASK_NAME" && -n "$PROMPT_FILE" ]]; then
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

build_plan

TASK_COUNT="$(jq -r '.task_count' "$TASKS_JSON")"
printf 'repo_root=%s\ngit_sha=%s\nprompt_file=%s\nsuite_file=%s\ntask_count=%s\nattempts=%s\nmodel=%s\nreasoning_effort=%s\nsandbox_mode=%s\norder_mode=%s\nrandom_seed=%s\noutput_dir=%s\n' \
  "$REPO_ROOT" "$GIT_SHA" "${PROMPT_FILE:-}" "${SUITE_FILE:-}" "$TASK_COUNT" "$ATTEMPTS" "$MODEL" "$REASONING_EFFORT" "$SANDBOX_MODE" "$ORDER_MODE" "$RANDOM_SEED" "$OUTPUT_DIR"
printf 'timeout_seconds=%s\n' "$TIMEOUT_SECONDS"
printf 'workspaces_root=%s\n' "$WORKSPACES_ROOT"
printf 'tasks_manifest=%s\nplan_jsonl=%s\n' "$TASKS_JSON" "$PLAN_JSONL"

while IFS= read -r plan_row; do
  [[ -n "$plan_row" ]] || continue
  run_plan_row "$plan_row"
done <"$PLAN_JSONL"

python3 - "$RESULTS_JSONL" "$SUMMARY_JSON" "$SUMMARY_MD" "$TASKS_JSON" "$PROMPT_FILE" "$SUITE_FILE" "$GIT_SHA" "$MODEL" "$REASONING_EFFORT" "$ORDER_MODE" "$RANDOM_SEED" <<'PY'
import json
import pathlib
import statistics
import sys
from collections import defaultdict


results_path = pathlib.Path(sys.argv[1])
summary_json_path = pathlib.Path(sys.argv[2])
summary_md_path = pathlib.Path(sys.argv[3])
tasks_path = pathlib.Path(sys.argv[4])
prompt_file = sys.argv[5]
suite_file = sys.argv[6]
git_sha = sys.argv[7]
model = sys.argv[8]
reasoning_effort = sys.argv[9]
order_mode = sys.argv[10]
random_seed = int(sys.argv[11])


def load_jsonl(path):
    rows = []
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if line:
            rows.append(json.loads(line))
    return rows


def median(values):
    return statistics.median(values) if values else None


def mean(values):
    return statistics.mean(values) if values else None


def summarize_rows(items):
    elapsed_values = [item["elapsed_ms"] for item in items]
    uncached_values = [item["uncached_input_tokens"] for item in items]
    output_values = [item["output_tokens"] for item in items]
    score_values = [item["quality_score"] for item in items if item.get("quality_score") is not None]
    pass_values = [item for item in items if item.get("quality_pass") is not None]
    completion_count = sum(1 for item in items if item.get("status") == "ok")
    pass_count = sum(1 for item in pass_values if item.get("quality_pass") is True)
    return {
        "count": len(items),
        "completion_count": completion_count,
        "completion_rate": completion_count / len(items) if items else None,
        "evaluation_count": len(pass_values),
        "pass_count": pass_count,
        "pass_rate": pass_count / len(pass_values) if pass_values else None,
        "median_quality_score": median(score_values),
        "mean_quality_score": mean(score_values),
        "median_elapsed_ms": median(elapsed_values),
        "median_uncached_input_tokens": median(uncached_values),
        "median_output_tokens": median(output_values),
        "used_optimus_tools_count": sum(1 for item in items if item.get("used_optimus_tools")),
    }


def win_loss_tie(deltas, smaller_is_better):
    optimus_better = 0
    baseline_better = 0
    ties = 0
    for delta in deltas:
        if delta == 0:
            ties += 1
        elif (delta > 0 and smaller_is_better) or (delta < 0 and not smaller_is_better):
            optimus_better += 1
        else:
            baseline_better += 1
    return {
        "optimus_better": optimus_better,
        "baseline_better": baseline_better,
        "ties": ties,
    }


rows = load_jsonl(results_path)
if not rows:
    raise SystemExit("no benchmark results recorded")

tasks_payload = json.loads(tasks_path.read_text(encoding="utf-8"))
tasks = tasks_payload["tasks"]

summary = {
    "schema_version": 2,
    "prompt_file": prompt_file or None,
    "suite_file": suite_file or None,
    "git_sha": git_sha,
    "model": model,
    "reasoning_effort": reasoning_effort or None,
    "order_mode": order_mode,
    "random_seed": random_seed,
    "task_count": tasks_payload["task_count"],
    "attempts_per_task": tasks_payload["attempts_per_task"],
    "row_count": len(rows),
    "tasks": tasks,
    "rows": rows,
    "arms": {},
    "categories": {},
    "tasks_summary": {},
    "paired": {
        "count": 0,
        "median_deltas": {},
        "wins": {},
    },
}

for arm in ("baseline", "optimus"):
    arm_rows = [row for row in rows if row["arm"] == arm]
    if arm_rows:
        summary["arms"][arm] = summarize_rows(arm_rows)

category_groups = defaultdict(list)
task_groups = defaultdict(list)
for row in rows:
    category_groups[row["task_category"]].append(row)
    task_groups[row["task_id"]].append(row)

for category, items in sorted(category_groups.items()):
    summary["categories"][category] = {
        arm: summarize_rows([item for item in items if item["arm"] == arm])
        for arm in ("baseline", "optimus")
        if any(item["arm"] == arm for item in items)
    }

for task in tasks:
    items = task_groups.get(task["id"], [])
    summary["tasks_summary"][task["id"]] = {
        "task_name": task["name"],
        "task_category": task["category"],
        "arms": {
            arm: summarize_rows([item for item in items if item["arm"] == arm])
            for arm in ("baseline", "optimus")
            if any(item["arm"] == arm for item in items)
        },
    }

paired_by_key = defaultdict(dict)
for row in rows:
    paired_by_key[row["pair_key"]][row["arm"]] = row

elapsed_deltas = []
uncached_input_deltas = []
output_deltas = []
score_deltas = []
pass_deltas = []
completion_deltas = []

for pair_key, pair in sorted(paired_by_key.items()):
    baseline = pair.get("baseline")
    optimus = pair.get("optimus")
    if not baseline or not optimus:
        continue
    summary["paired"]["count"] += 1
    elapsed_deltas.append(baseline["elapsed_ms"] - optimus["elapsed_ms"])
    uncached_input_deltas.append(baseline["uncached_input_tokens"] - optimus["uncached_input_tokens"])
    output_deltas.append(baseline["output_tokens"] - optimus["output_tokens"])
    completion_deltas.append((1 if baseline["status"] == "ok" else 0) - (1 if optimus["status"] == "ok" else 0))
    if baseline.get("quality_score") is not None and optimus.get("quality_score") is not None:
        score_deltas.append(baseline["quality_score"] - optimus["quality_score"])
    if baseline.get("quality_pass") is not None and optimus.get("quality_pass") is not None:
        pass_deltas.append((1 if baseline["quality_pass"] else 0) - (1 if optimus["quality_pass"] else 0))

summary["paired"]["median_deltas"] = {
    "elapsed_ms": median(elapsed_deltas),
    "uncached_input_tokens": median(uncached_input_deltas),
    "output_tokens": median(output_deltas),
    "completion_rate_points": median(completion_deltas),
    "quality_score": median(score_deltas),
    "quality_pass_points": median(pass_deltas),
}
summary["paired"]["wins"] = {
    "elapsed_ms": win_loss_tie(elapsed_deltas, smaller_is_better=True),
    "uncached_input_tokens": win_loss_tie(uncached_input_deltas, smaller_is_better=True),
    "output_tokens": win_loss_tie(output_deltas, smaller_is_better=True),
    "completion": win_loss_tie(completion_deltas, smaller_is_better=False),
    "quality_score": win_loss_tie(score_deltas, smaller_is_better=False),
    "quality_pass": win_loss_tie(pass_deltas, smaller_is_better=False),
}

summary_json_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")

lines = []
lines.append("# Codex Real A/B Benchmark")
lines.append("")
lines.append(f"- prompt file: `{prompt_file}`" if prompt_file else f"- suite file: `{suite_file}`")
lines.append(f"- git sha: `{git_sha}`")
lines.append(f"- model: `{model}`")
lines.append(f"- reasoning effort: `{reasoning_effort}`" if reasoning_effort else "- reasoning effort: inherited")
lines.append(f"- tasks: `{summary['task_count']}`")
lines.append(f"- attempts per task: `{summary['attempts_per_task']}`")
lines.append(f"- order mode: `{order_mode}`")
lines.append(f"- random seed: `{random_seed}`")
lines.append("")
lines.append("## Arm Summary")
lines.append("")
lines.append("| arm | completion | evaluated | pass | median score | median elapsed ms | median uncached input | median output | optimus tool evidence |")
lines.append("| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- |")
for arm in ("baseline", "optimus"):
    arm_summary = summary["arms"].get(arm)
    if not arm_summary:
        continue
    completion = f"{arm_summary['completion_count']}/{arm_summary['count']}"
    evaluated = f"{arm_summary['evaluation_count']}/{arm_summary['count']}"
    passed = (
        f"{arm_summary['pass_count']}/{arm_summary['evaluation_count']}"
        if arm_summary["evaluation_count"]
        else "n/a"
    )
    median_score = arm_summary["median_quality_score"]
    median_score_rendered = "n/a" if median_score is None else str(median_score)
    lines.append(
        f"| {arm} | {completion} | {evaluated} | {passed} | {median_score_rendered} | "
        f"{arm_summary['median_elapsed_ms']} | {arm_summary['median_uncached_input_tokens']} | "
        f"{arm_summary['median_output_tokens']} | {arm_summary['used_optimus_tools_count']}/{arm_summary['count']} |"
    )

lines.append("")
lines.append("## Paired Deltas")
lines.append("")
lines.append(f"- paired comparisons: `{summary['paired']['count']}`")
for metric, value in summary["paired"]["median_deltas"].items():
    lines.append(f"- median baseline minus optimus {metric}: `{value}`")
for metric, counts in summary["paired"]["wins"].items():
    lines.append(
        f"- {metric} wins: optimus `{counts['optimus_better']}`, baseline `{counts['baseline_better']}`, ties `{counts['ties']}`"
    )

if summary["categories"]:
    lines.append("")
    lines.append("## Category Summary")
    lines.append("")
    lines.append("| category | arm | completion | pass | median score | median elapsed ms | median uncached input |")
    lines.append("| --- | --- | --- | --- | ---: | ---: | ---: |")
    for category, arms in summary["categories"].items():
        for arm in ("baseline", "optimus"):
            arm_summary = arms.get(arm)
            if not arm_summary:
                continue
            passed = (
                f"{arm_summary['pass_count']}/{arm_summary['evaluation_count']}"
                if arm_summary["evaluation_count"]
                else "n/a"
            )
            median_score = arm_summary["median_quality_score"]
            median_score_rendered = "n/a" if median_score is None else str(median_score)
            lines.append(
                f"| {category} | {arm} | {arm_summary['completion_count']}/{arm_summary['count']} | "
                f"{passed} | {median_score_rendered} | {arm_summary['median_elapsed_ms']} | "
                f"{arm_summary['median_uncached_input_tokens']} |"
            )

lines.append("")
lines.append("## Task Summary")
lines.append("")
lines.append("| task | category | arm | completion | pass | median score | median elapsed ms |")
lines.append("| --- | --- | --- | --- | --- | ---: | ---: |")
for task in tasks:
    task_summary = summary["tasks_summary"].get(task["id"], {})
    arms = task_summary.get("arms", {})
    for arm in ("baseline", "optimus"):
        arm_summary = arms.get(arm)
        if not arm_summary:
            continue
        passed = (
            f"{arm_summary['pass_count']}/{arm_summary['evaluation_count']}"
            if arm_summary["evaluation_count"]
            else "n/a"
        )
        median_score = arm_summary["median_quality_score"]
        median_score_rendered = "n/a" if median_score is None else str(median_score)
        lines.append(
            f"| {task['name']} | {task['category']} | {arm} | "
            f"{arm_summary['completion_count']}/{arm_summary['count']} | {passed} | "
            f"{median_score_rendered} | {arm_summary['median_elapsed_ms']} |"
        )

summary_md_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
PY

printf 'summary_json=%s\nsummary_md=%s\nresults_jsonl=%s\ntasks_json=%s\nplan_jsonl=%s\n' \
  "$SUMMARY_JSON" "$SUMMARY_MD" "$RESULTS_JSONL" "$TASKS_JSON" "$PLAN_JSONL"
