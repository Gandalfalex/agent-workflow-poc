#!/bin/bash

# Create Git Worktree for Ticket Implementation
# Usage: ./create-worktree.sh <ticket-key> <repo-path> [worktree-root]
# Example: ./create-worktree.sh PROJ-001 /path/to/repo /path/to/worktrees

set -e

TICKET_KEY="${1:-}"
REPO_PATH="${2:-.}"
WORKTREE_ROOT="${3:-.}/worktrees"

# Validate inputs
if [ -z "$TICKET_KEY" ]; then
  echo "Error: Ticket key is required" >&2
  exit 1
fi

# Validate repo path exists and is a git repository
if [ ! -d "$REPO_PATH" ]; then
  echo "Error: Repository path does not exist: $REPO_PATH" >&2
  exit 1
fi

if ! cd "$REPO_PATH" && git rev-parse --git-dir > /dev/null 2>&1; then
  echo "Error: Not a valid git repository: $REPO_PATH" >&2
  exit 1
fi

# Get absolute paths
REPO_ROOT=$(cd "$REPO_PATH" && git rev-parse --show-toplevel)
WORKTREE_ROOT=$(cd "$REPO_PATH" && mkdir -p "$WORKTREE_ROOT" && cd "$WORKTREE_ROOT" && pwd)

# Branch and worktree paths
BRANCH_NAME="feature/$TICKET_KEY"
WORKTREE_PATH="$WORKTREE_ROOT/$TICKET_KEY"

# Check if worktree already exists
if [ -d "$WORKTREE_PATH" ]; then
  echo "Cleaning up existing worktree at $WORKTREE_PATH" >&2
  cd "$REPO_ROOT"
  git worktree prune
  if [ -d "$WORKTREE_PATH" ]; then
    rm -rf "$WORKTREE_PATH"
  fi
fi

# Check if branch already exists
cd "$REPO_ROOT"
if git rev-parse --verify "$BRANCH_NAME" > /dev/null 2>&1; then
  echo "Branch $BRANCH_NAME already exists, creating worktree from existing branch" >&2
  git worktree add "$WORKTREE_PATH" "$BRANCH_NAME" 2>/dev/null || {
    # If worktree already linked, just use it
    if [ -d "$WORKTREE_PATH" ]; then
      echo "Worktree already exists at $WORKTREE_PATH" >&2
    else
      echo "Error: Failed to create worktree" >&2
      exit 1
    fi
  }
else
  echo "Creating new branch and worktree: $BRANCH_NAME" >&2
  # Ensure main/master branch exists for checkout
  git fetch origin main:main 2>/dev/null || git fetch origin master:master 2>/dev/null || true
  git worktree add -b "$BRANCH_NAME" "$WORKTREE_PATH" 2>/dev/null || {
    echo "Error: Failed to create worktree with new branch" >&2
    exit 1
  }
fi

# Verify worktree was created
if [ ! -d "$WORKTREE_PATH" ]; then
  echo "Error: Worktree directory was not created" >&2
  exit 1
fi

# Output JSON result
cat << JSON
{
  "success": true,
  "worktreePath": "$WORKTREE_PATH",
  "branch": "$BRANCH_NAME",
  "repoRoot": "$REPO_ROOT",
  "ticketKey": "$TICKET_KEY"
}
JSON
