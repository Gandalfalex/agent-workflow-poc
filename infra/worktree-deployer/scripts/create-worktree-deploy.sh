#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib.sh"

usage() {
  cat <<EOF
Usage: $0 --slug <slug> --worktree-path <path> [--branch <branch>] [--dry-run]
EOF
}

DRY_RUN=0
INPUT_SLUG=""
WORKTREE_PATH=""
BRANCH_NAME=""

while [ "$#" -gt 0 ]; do
  case "$1" in
    --slug) INPUT_SLUG="$2"; shift 2 ;;
    --worktree-path) WORKTREE_PATH="$2"; shift 2 ;;
    --branch) BRANCH_NAME="$2"; shift 2 ;;
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) fail "unknown argument: $1" ;;
  esac
done

[ -n "$INPUT_SLUG" ] || fail "--slug is required"
[ -n "$WORKTREE_PATH" ] || fail "--worktree-path is required"

require_cmd docker
require_cmd git
require_cmd curl

ensure_dirs
load_root_env
assert_common_env

WORKTREE_ROOT=$(cd "$WORKTREE_PATH" && pwd)
is_git_worktree "$WORKTREE_ROOT" || fail "not a git worktree: $WORKTREE_ROOT"

SLUG=$(slugify "$INPUT_SLUG")
WORKTREE_TICKETING_ROOT="$WORKTREE_ROOT/ticketing-system"
[ -d "$WORKTREE_TICKETING_ROOT" ] || fail "missing ticketing-system in worktree: $WORKTREE_TICKETING_ROOT"

if [ -z "$BRANCH_NAME" ]; then
  BRANCH_NAME=$(git -C "$WORKTREE_ROOT" rev-parse --abbrev-ref HEAD)
fi

ENV_FILE=$(env_file_for_slug "$SLUG")
META_FILE=$(meta_file_for_slug "$SLUG")

DATABASE_NAME=$(identifier_for_slug "$SLUG" "ticketing_wt_")
DATABASE_USER="$DATABASE_NAME"
DATABASE_PASSWORD="$DATABASE_NAME"
COMPOSE_PROJECT="wt-$SLUG"
CONTAINER_NAME="ticketing-wt-$SLUG"
PUBLIC_PATH="$SLUG"
DEBUG_PORT=$(next_debug_port)
PREVIEW_SHARED_NETWORK="${PREVIEW_SHARED_NETWORK:-coding-agent-workflow_prod-network}"
KEYCLOAK_BASE_URL="http://keycloak:8080"
MINIO_ENDPOINT="minio:9000"
MINIO_ACCESS_KEY="$MINIO_ROOT_USER"
MINIO_SECRET_KEY="$MINIO_ROOT_PASSWORD"
MINIO_USE_SSL="false"
CORS_ALLOWED_ORIGINS="${PUBLIC_BASE_URL}"
DATABASE_URL="postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@postgres:5432/${DATABASE_NAME}?sslmode=disable"

if [ -f "$ENV_FILE" ]; then
  source_env_file "$ENV_FILE"
  log "deployment already exists, reusing saved env for $SLUG"
fi

if [ "$DRY_RUN" -eq 1 ]; then
  print_summary
  exit 0
fi

create_role_if_missing "$DATABASE_USER" "$DATABASE_PASSWORD"
create_db_if_missing "$DATABASE_NAME" "$DATABASE_USER"

render_template \
  "$TEMPLATES_DIR/worktree.env.tpl" \
  "__SLUG__" "$SLUG" \
  "__BRANCH__" "$BRANCH_NAME" \
  "__WORKTREE_ROOT__" "$WORKTREE_ROOT" \
  "__WORKTREE_TICKETING_ROOT__" "$WORKTREE_TICKETING_ROOT" \
  "__COMPOSE_PROJECT__" "$COMPOSE_PROJECT" \
  "__CONTAINER_NAME__" "$CONTAINER_NAME" \
  "__DATABASE_NAME__" "$DATABASE_NAME" \
  "__DATABASE_USER__" "$DATABASE_USER" \
  "__DATABASE_PASSWORD__" "$DATABASE_PASSWORD" \
  "__DATABASE_URL__" "$DATABASE_URL" \
  "__DEBUG_PORT__" "$DEBUG_PORT" \
  "__PREVIEW_SHARED_NETWORK__" "$PREVIEW_SHARED_NETWORK" \
  "__KEYCLOAK_BASE_URL__" "$KEYCLOAK_BASE_URL" \
  "__KEYCLOAK_REALM__" "$KEYCLOAK_REALM" \
  "__KEYCLOAK_CLIENT_ID__" "$KEYCLOAK_CLIENT_ID" \
  "__KEYCLOAK_API_ADMIN_USER__" "$KEYCLOAK_API_ADMIN_USER" \
  "__KEYCLOAK_API_ADMIN_PASSWORD__" "$KEYCLOAK_API_ADMIN_PASSWORD" \
  "__MINIO_ENDPOINT__" "$MINIO_ENDPOINT" \
  "__MINIO_ACCESS_KEY__" "$MINIO_ACCESS_KEY" \
  "__MINIO_SECRET_KEY__" "$MINIO_SECRET_KEY" \
  "__MINIO_BUCKET__" "$MINIO_BUCKET" \
  "__MINIO_USE_SSL__" "$MINIO_USE_SSL" \
  "__CORS_ALLOWED_ORIGINS__" "$CORS_ALLOWED_ORIGINS" \
  "__PUBLIC_PATH__" "$PUBLIC_PATH" >"$ENV_FILE"

source_env_file "$ENV_FILE"
write_metadata_json "$META_FILE"

docker_compose_preview "$ENV_FILE" up -d --build

if ! wait_for_local_health "$DEBUG_PORT"; then
  fail "deployment started but health check failed on port $DEBUG_PORT"
fi

write_worktree_route_file "$SLUG" "http://${CONTAINER_NAME}:8080" "$(public_host)" "$PUBLIC_PATH"

if ! wait_for_public_worktree_health "$PUBLIC_PATH"; then
  fail "deployment is healthy locally but public route /${PUBLIC_PATH} did not become healthy"
fi

print_summary
