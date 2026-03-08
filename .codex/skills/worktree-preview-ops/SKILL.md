---
name: worktree-preview-ops
description: Manage path-based worktree deployments on the shared production host. Use when the user wants to create, update, verify, or destroy a disposable worktree deployment under the base domain with production fixed at /ticketing and worktrees exposed under /<slug>.
---

# Worktree Preview Ops

Use this skill for the worktree deployment platform in `infra/worktree-deployer/`.

## When to use it

Trigger this skill when the user wants to:
- deploy a git worktree to the shared server
- rebuild or refresh a worktree deployment
- publish a worktree under `/<slug>` on the base domain
- remove a worktree deployment and drop its isolated database
- inspect or adjust the worktree deployer, its Traefik route generation, or its n8n operator workflows

Do not use this skill for normal app feature work inside `ticketing-system/` unless the request is specifically about the deployment platform.

## Ground truth

The deployment model is:
- production stays at `/ticketing`
- each worktree is exposed at `/<slug>`
- shared singletons remain in place: Postgres, Keycloak, MinIO, n8n, Traefik
- each worktree gets its own app container, Postgres database, Postgres role, and generated route file

Primary files:
- `infra/worktree-deployer/README.md`
- `infra/worktree-deployer/scripts/create-worktree-deploy.sh`
- `infra/worktree-deployer/scripts/update-worktree-deploy.sh`
- `infra/worktree-deployer/scripts/destroy-worktree-deploy.sh`
- `infra/worktree-deployer/scripts/lib.sh`
- `infra/worktree-deployer/templates/worktree.compose.yaml`
- `infra/worktree-deployer/templates/traefik-worktree-route.yaml.tpl`
- `infra/worktree-deployer/state/traefik/00-base.yaml`
- `n8n/workflows/worktree-create-deploy.json`
- `n8n/workflows/worktree-update-deploy.json`
- `n8n/workflows/worktree-destroy-deploy.json`

## Workflow

1. Read the deployer README and the relevant script before changing behavior.
2. Confirm whether the request is about `create`, `update`, `destroy`, or route/config changes.
3. Prefer using the existing deployer scripts rather than inventing new docker or Traefik commands.
4. Keep routing assumptions consistent:
   - prod is `/ticketing`
   - worktrees are `/<slug>`
5. When validating changes, prefer:
   - `bash -n` for scripts
   - `docker compose -f docker-compose.prod.yaml config`
   - JSON parsing for n8n workflow files
6. If you modify the deployer contract, update both:
   - `infra/worktree-deployer/README.md`
   - the affected n8n workflow JSON files

## Operator commands

Examples:

```bash
infra/worktree-deployer/scripts/create-worktree-deploy.sh --slug abc --worktree-path /abs/path/to/worktrees/abc
infra/worktree-deployer/scripts/update-worktree-deploy.sh --slug abc
infra/worktree-deployer/scripts/destroy-worktree-deploy.sh --slug abc
```

Dry-run examples:

```bash
infra/worktree-deployer/scripts/create-worktree-deploy.sh --slug abc --worktree-path /abs/path --dry-run
infra/worktree-deployer/scripts/update-worktree-deploy.sh --slug abc --dry-run
infra/worktree-deployer/scripts/destroy-worktree-deploy.sh --slug abc --dry-run
```

## Guardrails

- Do not reintroduce the old “switch `/ticketing` to a worktree” model unless the user explicitly asks for it.
- Do not use schema-per-worktree; the platform is database-per-worktree.
- Do not add separate Keycloak or MinIO instances for worktrees in v1.
- Keep generated runtime state inside `infra/worktree-deployer/state/`.
