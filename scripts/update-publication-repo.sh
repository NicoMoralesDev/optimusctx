#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 4 ]; then
  echo "usage: $0 <repo-dir> <rendered-file> <target-path> <commit-message>" >&2
  exit 1
fi

repo_dir="$1"
rendered_file="$2"
target_path="$3"
commit_message="$4"

write_output() {
  local key="$1"
  local value="$2"
  if [ -n "${GITHUB_OUTPUT:-}" ]; then
    echo "${key}=${value}" >> "$GITHUB_OUTPUT"
  fi
  printf '%s=%s\n' "$key" "$value"
}

install -m 0644 -D "$rendered_file" "${repo_dir}/${target_path}"

cd "$repo_dir"
status=$(git status --porcelain --untracked-files=all -- "$target_path")
if [ -z "$status" ]; then
  write_output "changed" "false"
  write_output "write_result" "already_current"
  exit 0
fi

git config user.name "github-actions[bot]"
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"
git add "$target_path"
git commit -m "$commit_message"
branch=$(git rev-parse --abbrev-ref HEAD)
git push origin HEAD:"$branch"

write_output "changed" "true"
write_output "write_result" "published"
