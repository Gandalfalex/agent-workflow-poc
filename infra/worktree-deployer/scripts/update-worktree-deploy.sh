#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib.sh"

usage() {
  cat <<EOF
Usage: $0 --slug <slug> [--dry-run]
EOF
}

DRY_RUN=0
INPUT_SLUG=""

while [ "$#" -gt 0 ]; do
  case "$1" in
    --slug) INPUT_SLUG="$2"; shift 2 ;;
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) fail "unknown argument: $1" ;;
  esac
done

[ -n "$INPUT_SLUG" ] || fail "--slug is required"

require_cmd docker
require_cmd curl

ensure_dirs
load_root_env
assert_common_env

SLUG=$(slugify "$INPUT_SLUG")
ENV_FILE=$(env_file_for_slug "$SLUG")
source_env_file "$ENV_FILE"

if [ "$DRY_RUN" -eq 1 ]; then
  print_summary
  exit 0
fi

docker_compose_preview "$ENV_FILE" up -d --build

if ! wait_for_local_health "$DEBUG_PORT"; then
  fail "updated deployment failed health check on port $DEBUG_PORT"
fi

write_worktree_route_file "$SLUG" "http://${CONTAINER_NAME}:8080" "$(public_host)" "$PUBLIC_PATH"

if ! wait_for_public_worktree_health "$PUBLIC_PATH"; then
  fail "updated deployment is healthy locally but public route /${PUBLIC_PATH} did not become healthy"
fi

print_summary
