#!/usr/bin/env bash
set -euo pipefail

branch_arg="${1:-}"
worktrees_arg="${2:-}"

if ! git_root=$(git rev-parse --show-toplevel 2>/dev/null); then
  echo "error: run inside a git repository" >&2
  exit 1
fi

cd "$git_root"

if [[ -n "$branch_arg" ]]; then
  branch_name="$branch_arg"
else
  if [[ -n "${ISSUE_KEY:-}" ]]; then
    branch_name="idea/${ISSUE_KEY}"
  else
    branch_name="new-idea"
  fi
fi

worktrees_dir="${worktrees_arg:-$git_root/worktrees}"
mkdir -p "$worktrees_dir"

worktree_path="$worktrees_dir/${branch_name//\//-}"

if git show-ref --verify --quiet "refs/heads/$branch_name"; then
  echo "error: branch already exists: $branch_name" >&2
  exit 1
fi

if [[ -e "$worktree_path" ]]; then
  echo "error: worktree path already exists: $worktree_path" >&2
  exit 1
fi

git worktree add "$worktree_path" -b "$branch_name"

echo "created worktree: $worktree_path"
