---
name: git-workspace-bootstrap
description: Create a new git workspace and branch inside an existing repo using git worktrees. Use when the user asks to bootstrap a workspace, create a worktree, start a new branch for an idea/issue, or set up a clean working directory without switching the current checkout.
---

# Git Workspace Bootstrap

## Overview

Bootstrap a new git worktree and branch inside the current repository. This keeps the main checkout untouched while creating a clean workspace for new work.

## Workflow

1. Confirm the repo root (must run inside a git repo).
2. Decide on branch name and workspace directory.
3. Run the bootstrap script to create the worktree and new branch.

## Defaults

- Branch name: `idea/<issueKey>` if an issue key is provided; otherwise `new-idea`.
- Worktree directory: `<repo-root>/worktrees`.

## Script

Use `scripts/bootstrap_worktree.sh` with optional arguments.

Examples:

```bash
./scripts/bootstrap_worktree.sh idea/TICK-123
./scripts/bootstrap_worktree.sh feature/new-ui /tmp/worktrees
```

Environment:

- `ISSUE_KEY` (optional): Used to form `idea/<ISSUE_KEY>` when no branch is passed.

## Resources

### scripts/

- `scripts/bootstrap_worktree.sh`: Creates a worktree directory and new branch from the current repo.
