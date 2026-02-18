#!/bin/sh
set -eu

# Create dedicated users and databases for ticketing, Keycloak, and n8n.
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<SQL
DO
\$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${TICKETING_DB_USER}') THEN
    EXECUTE format('CREATE ROLE %I LOGIN PASSWORD %L', '${TICKETING_DB_USER}', '${TICKETING_DB_PASSWORD}');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${KEYCLOAK_DB_USER}') THEN
    EXECUTE format('CREATE ROLE %I LOGIN PASSWORD %L', '${KEYCLOAK_DB_USER}', '${KEYCLOAK_DB_PASSWORD}');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${N8N_DB_USER}') THEN
    EXECUTE format('CREATE ROLE %I LOGIN PASSWORD %L', '${N8N_DB_USER}', '${N8N_DB_PASSWORD}');
  END IF;
END
\$\$;
SQL

for db_owner in \
  "${TICKETING_DB_NAME}:${TICKETING_DB_USER}" \
  "${KEYCLOAK_DB_NAME}:${KEYCLOAK_DB_USER}" \
  "${N8N_DB_NAME}:${N8N_DB_USER}"; do
  db_name="${db_owner%%:*}"
  db_user="${db_owner##*:}"

  if ! psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -tAc "SELECT 1 FROM pg_database WHERE datname='${db_name}'" | grep -q 1; then
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -c "CREATE DATABASE \"${db_name}\" OWNER \"${db_user}\""
  fi
done
