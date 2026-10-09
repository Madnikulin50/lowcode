#!/usr/bin/env bash
# Copy PostgreSQL database localhost/test9 → 161.104.90.77/db_ru3
# Replaces public schema on the destination (destructive).
#
# Usage:
#   LOCAL_PGPASSWORD=... REMOTE_PGPASSWORD=... ./scripts/migrate-test9-to-db_ru3.sh
#
# After restore, restart the VPS Corteza container:
#   docker-compose restart server
#
# This copies SQL data only. Object storage (attachments under data/server)
# must be copied separately if files are needed.

set -euo pipefail

LOCAL_PGPASSWORD="Zse45rdx"
LOCAL_HOST="${LOCAL_HOST:-127.0.0.1}"
LOCAL_PORT="${LOCAL_PORT:-5432}"
LOCAL_USER="${LOCAL_USER:-postgres}"
LOCAL_DB="${LOCAL_DB:-test9}"
LOCAL_PGPASSWORD="${LOCAL_PGPASSWORD:?set LOCAL_PGPASSWORD}"

REMOTE_PGPASSWORD="lowcode2026"
REMOTE_HOST="${REMOTE_HOST:-161.104.90.77}"
REMOTE_PORT="${REMOTE_PORT:-5432}"
REMOTE_USER="${REMOTE_USER:-postgres}"
REMOTE_DB="${REMOTE_DB:-db_ru3}"
REMOTE_PGPASSWORD="${REMOTE_PGPASSWORD:?set REMOTE_PGPASSWORD}"

DUMP="${DUMP:-/tmp/${LOCAL_DB}.dump}"

export PGSSLMODE=disable

echo "==> dump ${LOCAL_USER}@${LOCAL_HOST}:${LOCAL_PORT}/${LOCAL_DB} → ${DUMP}"
PGPASSWORD="${LOCAL_PGPASSWORD}" pg_dump \
  -h "${LOCAL_HOST}" -p "${LOCAL_PORT}" -U "${LOCAL_USER}" -d "${LOCAL_DB}" \
  -Fc --no-owner --no-acl \
  -f "${DUMP}"
ls -lh "${DUMP}"

REMOTE_URL="postgres://${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PORT}/${REMOTE_DB}?sslmode=disable"

echo "==> terminate sessions on ${REMOTE_HOST}/${REMOTE_DB}"
PGPASSWORD="${REMOTE_PGPASSWORD}" psql "${REMOTE_URL}" -v ON_ERROR_STOP=1 -c \
  "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${REMOTE_DB}' AND pid <> pg_backend_pid();" \
  || true

echo "==> replace public schema on ${REMOTE_HOST}/${REMOTE_DB}"
PGPASSWORD="${REMOTE_PGPASSWORD}" psql "${REMOTE_URL}" -v ON_ERROR_STOP=1 -c \
  "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO ${REMOTE_USER}; GRANT ALL ON SCHEMA public TO public;"

echo "==> restore ${DUMP} → ${REMOTE_HOST}/${REMOTE_DB}"
PGPASSWORD="${REMOTE_PGPASSWORD}" pg_restore \
  -h "${REMOTE_HOST}" -p "${REMOTE_PORT}" -U "${REMOTE_USER}" -d "${REMOTE_DB}" \
  --no-owner --no-acl \
  "${DUMP}"

echo "==> verify"
PGPASSWORD="${REMOTE_PGPASSWORD}" psql "${REMOTE_URL}" -c \
  "SELECT current_database(), pg_size_pretty(pg_database_size(current_database())) AS size;
   SELECT COUNT(*) AS roles FROM roles;
   SELECT COUNT(*) AS users FROM users;
   SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid='roles'::regclass AND contype='p';"

echo "done. restart Corteza: docker-compose restart server"
