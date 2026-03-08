#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
DEPLOYER_DIR="$ROOT_DIR/infra/worktree-deployer"
STATE_DIR="$DEPLOYER_DIR/state"
DEPLOYMENTS_DIR="$STATE_DIR/deployments"
TRAEFIK_DIR="$STATE_DIR/traefik"
TEMPLATES_DIR="$DEPLOYER_DIR/templates"
PROD_COMPOSE_FILE="$ROOT_DIR/docker-compose.prod.yaml"
WORKTREE_COMPOSE_FILE="$TEMPLATES_DIR/worktree.compose.yaml"

log() {
  printf '[worktree-deployer] %s\n' "$*" >&2
}

fail() {
  printf '[worktree-deployer] ERROR: %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

ensure_dirs() {
  mkdir -p "$DEPLOYMENTS_DIR" "$TRAEFIK_DIR"
}

load_root_env() {
  if [ -f "$ROOT_DIR/.env" ]; then
    set -a
    # shellcheck disable=SC1091
    . "$ROOT_DIR/.env"
    set +a
  fi
}

assert_common_env() {
  : "${PUBLIC_BASE_URL:?PUBLIC_BASE_URL is required}"
  : "${POSTGRES_ADMIN_USER:?POSTGRES_ADMIN_USER is required}"
  : "${POSTGRES_ADMIN_DB:?POSTGRES_ADMIN_DB is required}"
  : "${KEYCLOAK_REALM:?KEYCLOAK_REALM is required}"
  : "${KEYCLOAK_CLIENT_ID:?KEYCLOAK_CLIENT_ID is required}"
  : "${KEYCLOAK_API_ADMIN_USER:?KEYCLOAK_API_ADMIN_USER is required}"
  : "${KEYCLOAK_API_ADMIN_PASSWORD:?KEYCLOAK_API_ADMIN_PASSWORD is required}"
  : "${MINIO_ROOT_USER:?MINIO_ROOT_USER is required}"
  : "${MINIO_ROOT_PASSWORD:?MINIO_ROOT_PASSWORD is required}"
  : "${MINIO_BUCKET:?MINIO_BUCKET is required}"
}

public_host() {
  local value="${PUBLIC_BASE_URL#*://}"
  printf '%s\n' "${value%%/*}"
}

slugify() {
  local value
  value=$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')
  value=$(printf '%s' "$value" | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/-+/-/g')
  [ -n "$value" ] || fail "could not derive slug"
  printf '%.40s\n' "$value"
}

identifier_for_slug() {
  local slug="$1"
  local prefix="$2"
  local ident
  ident=$(printf '%s' "$slug" | tr '-' '_' | tr -cd 'a-zA-Z0-9_')
  printf '%.63s\n' "${prefix}${ident}"
}

env_file_for_slug() {
  printf '%s/%s.env\n' "$DEPLOYMENTS_DIR" "$1"
}

meta_file_for_slug() {
  printf '%s/%s.json\n' "$DEPLOYMENTS_DIR" "$1"
}

route_file_for_slug() {
  printf '%s/%s.yaml\n' "$TRAEFIK_DIR" "$1"
}

source_env_file() {
  local file="$1"
  [ -f "$file" ] || fail "missing env file: $file"
  set -a
  # shellcheck disable=SC1090
  . "$file"
  set +a
}

write_metadata_json() {
  local file="$1"
  cat >"$file" <<EOF
{
  "slug": "${SLUG}",
  "branch": "${BRANCH}",
  "worktreePath": "${WORKTREE_ROOT}",
  "ticketingRoot": "${WORKTREE_TICKETING_ROOT}",
  "composeProject": "${COMPOSE_PROJECT}",
  "containerName": "${CONTAINER_NAME}",
  "databaseName": "${DATABASE_NAME}",
  "databaseUser": "${DATABASE_USER}",
  "debugPort": ${DEBUG_PORT}
}
EOF
}

next_debug_port() {
  local start="${WORKTREE_DEBUG_PORT_START:-18100}"
  local end="${WORKTREE_DEBUG_PORT_END:-18199}"
  local port used

  for port in $(seq "$start" "$end"); do
    used=0
    while IFS= read -r file; do
      [ -f "$file" ] || continue
      if grep -q "^DEBUG_PORT=${port}\$" "$file"; then
        used=1
        break
      fi
    done < <(find "$DEPLOYMENTS_DIR" -maxdepth 1 -name '*.env' -type f | sort)

    if [ "$used" -eq 0 ]; then
      printf '%s\n' "$port"
      return
    fi
  done

  fail "no free debug port in range ${start}-${end}"
}

is_git_worktree() {
  git -C "$1" rev-parse --is-inside-work-tree >/dev/null 2>&1
}

psql_admin() {
  docker compose -f "$PROD_COMPOSE_FILE" exec -T postgres \
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_ADMIN_USER" --dbname "$POSTGRES_ADMIN_DB" "$@"
}

database_exists() {
  psql_admin -tAc "SELECT 1 FROM pg_database WHERE datname='${1}'" | grep -q 1
}

role_exists() {
  psql_admin -tAc "SELECT 1 FROM pg_roles WHERE rolname='${1}'" | grep -q 1
}

create_role_if_missing() {
  local role="$1"
  local password="$2"
  if role_exists "$role"; then
    log "role already exists: $role"
    return
  fi
  psql_admin <<EOF
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${role}') THEN
    EXECUTE format('CREATE ROLE %I LOGIN PASSWORD %L', '${role}', '${password}');
  END IF;
END
\$\$;
EOF
}

create_db_if_missing() {
  local db_name="$1"
  local role="$2"
  if database_exists "$db_name"; then
    log "database already exists: $db_name"
    return
  fi
  psql_admin -c "CREATE DATABASE \"${db_name}\" OWNER \"${role}\""
}

drop_db_and_role() {
  local db_name="$1"
  local role="$2"

  if database_exists "$db_name"; then
    psql_admin <<EOF
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = '${db_name}' AND pid <> pg_backend_pid();
DROP DATABASE "${db_name}";
EOF
  fi

  if role_exists "$role"; then
    psql_admin -c "DROP ROLE \"${role}\""
  fi
}

docker_compose_preview() {
  local env_file="$1"
  shift
  docker compose --env-file "$env_file" -p "$COMPOSE_PROJECT" -f "$WORKTREE_COMPOSE_FILE" "$@"
}

wait_for_local_health() {
  local port="$1"
  local attempts=30
  local url="http://127.0.0.1:${port}/health"
  local i

  for i in $(seq 1 "$attempts"); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done

  return 1
}

wait_for_public_ticketing_health() {
  local attempts=20
  local url="${PUBLIC_BASE_URL}/ticketing/health"
  local i

  for i in $(seq 1 "$attempts"); do
    if curl -kfsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done

  return 1
}

wait_for_public_worktree_health() {
  local slug="$1"
  local attempts=20
  local url="${PUBLIC_BASE_URL}/${slug}/health"
  local i

  for i in $(seq 1 "$attempts"); do
    if curl -kfsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done

  return 1
}

render_template() {
  local template="$1"
  shift
  local rendered
  rendered=$(cat "$template")
  while [ "$#" -gt 1 ]; do
    rendered=$(printf '%s' "$rendered" | sed "s|$1|$2|g")
    shift 2
  done
  printf '%s\n' "$rendered"
}

write_worktree_route_file() {
  local slug="$1"
  local target_url="$2"
  local host="$3"
  local public_path="$4"
  local route_file

  route_file=$(route_file_for_slug "$slug")

  render_template \
    "$TEMPLATES_DIR/traefik-worktree-route.yaml.tpl" \
    "__SLUG__" "$slug" \
    "__TARGET_URL__" "$target_url" \
    "__PUBLIC_HOST__" "$host" \
    "__PUBLIC_PATH__" "$public_path" >"$route_file"
}

remove_worktree_route_file() {
  rm -f "$(route_file_for_slug "$1")"
}

print_summary() {
  cat <<EOF
slug=${SLUG}
compose_project=${COMPOSE_PROJECT}
container_name=${CONTAINER_NAME}
database_name=${DATABASE_NAME}
database_user=${DATABASE_USER}
debug_port=${DEBUG_PORT}
worktree_root=${WORKTREE_ROOT}
public_path=/${PUBLIC_PATH}
EOF
}
