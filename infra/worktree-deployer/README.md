# Worktree Deployer

This project manages disposable worktree deployments on the same server while keeping the public URL stable at the base domain.

## Model
- Shared singleton services stay in the root production stack: Postgres, Keycloak, MinIO, n8n, Traefik.
- Each worktree deployment gets its own app container, Postgres database, Postgres role, generated env file, and debug port.
- Production stays fixed at `/ticketing`.
- Each worktree gets its own path under the same host, for example `/abc`.

## Layout
- `templates/worktree.compose.yaml` - compose template for one worktree deployment
- `templates/worktree.env.tpl` - generated env file shape
- `templates/traefik-worktree-route.yaml.tpl` - generated Traefik route for one worktree path
- `scripts/` - operator scripts
- `state/deployments/` - generated per-worktree env and metadata
- `state/traefik/00-base.yaml` - tracked base Traefik routes for prod, n8n, and Keycloak
- `state/traefik/<slug>.yaml` - generated per-worktree route files

## Required Environment
Scripts load the repository root `.env` file and expect:
- `PUBLIC_BASE_URL`
- `POSTGRES_ADMIN_USER`
- `POSTGRES_ADMIN_DB`
- `KEYCLOAK_REALM`
- `KEYCLOAK_CLIENT_ID`
- `KEYCLOAK_API_ADMIN_USER`
- `KEYCLOAK_API_ADMIN_PASSWORD`
- `MINIO_ROOT_USER`
- `MINIO_ROOT_PASSWORD`
- `MINIO_BUCKET`

Optional:
- `PREVIEW_SHARED_NETWORK` default: `coding-agent-workflow_prod-network`
- `WORKTREE_DEBUG_PORT_START` default: `18100`
- `WORKTREE_DEBUG_PORT_END` default: `18199`

## Commands
Create a deployment from an existing worktree:
```bash
infra/worktree-deployer/scripts/create-worktree-deploy.sh \
  --slug tkt-033-approval-gates \
  --worktree-path /abs/path/to/worktrees/TKT-033
```

Rebuild and restart an existing deployment:
```bash
infra/worktree-deployer/scripts/update-worktree-deploy.sh --slug tkt-033-approval-gates
```

Destroy a deployment:
```bash
infra/worktree-deployer/scripts/destroy-worktree-deploy.sh --slug tkt-033-approval-gates
```

All scripts support `--dry-run`.

## Operational Notes
- Worktrees must already exist; this project does not create them.
- Direct container debugging uses `http://127.0.0.1:<debug-port>/health`.
- Public verification for a worktree uses `${PUBLIC_BASE_URL}/<slug>/health`.
- Production remains available at `${PUBLIC_BASE_URL}/ticketing`.

## n8n Workflows
Import these workflow files into n8n:
- `n8n/workflows/worktree-create-deploy.json`
- `n8n/workflows/worktree-update-deploy.json`
- `n8n/workflows/worktree-destroy-deploy.json`

They are manual/webhook operator flows that SSH into the deployment host and call the scripts in this directory.
