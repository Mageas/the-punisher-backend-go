#!/bin/sh
set -eu

: "${APP_DATABASE_URL:?APP_DATABASE_URL is required}"

MIGRATIONS_DIR="${MIGRATIONS_DIR:-/app/db/migrations}"
MIGRATE_MAX_ATTEMPTS="${MIGRATE_MAX_ATTEMPTS:-30}"
MIGRATE_RETRY_INTERVAL="${MIGRATE_RETRY_INTERVAL:-2}"

echo "Running database migrations from ${MIGRATIONS_DIR}..."

attempt=1
while [ "$attempt" -le "$MIGRATE_MAX_ATTEMPTS" ]; do
  set +e
  output="$(migrate -path "$MIGRATIONS_DIR" -database "$APP_DATABASE_URL" up 2>&1)"
  status=$?
  set -e

  if [ "$status" -eq 0 ]; then
    [ -n "$output" ] && echo "$output"
    echo "Migrations applied."
    break
  fi

  if echo "$output" | grep -qi "no change"; then
    echo "$output"
    echo "No new migrations."
    break
  fi

  if [ "$attempt" -eq "$MIGRATE_MAX_ATTEMPTS" ]; then
    echo "$output"
    echo "Migration failed after ${MIGRATE_MAX_ATTEMPTS} attempts."
    exit "$status"
  fi

  echo "$output"
  echo "Migration attempt ${attempt}/${MIGRATE_MAX_ATTEMPTS} failed. Retrying in ${MIGRATE_RETRY_INTERVAL}s..."
  attempt=$((attempt + 1))
  sleep "$MIGRATE_RETRY_INTERVAL"
done

exec /app/main
