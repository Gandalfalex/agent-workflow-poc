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

ensure_dirs
load_root_env
assert_common_env

SLUG=$(slugify "$INPUT_SLUG")
ENV_FILE=$(env_file_for_slug "$SLUG")
META_FILE=$(meta_file_for_slug "$SLUG")
source_env_file "$ENV_FILE"

if [ "$DRY_RUN" -eq 1 ]; then
  print_summary
  exit 0
fi

docker_compose_preview "$ENV_FILE" down --remove-orphans
remove_worktree_route_file "$SLUG"
drop_db_and_role "$DATABASE_NAME" "$DATABASE_USER"
rm -f "$ENV_FILE" "$META_FILE"

printf 'destroyed_slug=%s\n' "$SLUG"
