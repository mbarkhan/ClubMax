#!/usr/bin/env bash

set -Eeuo pipefail

PROJECT_DIR="${CLUBMAX_PROJECT_DIR:-/opt/clubmax}"
COMPOSE=(docker compose --env-file clubmax.env -f compose.production.yaml)

if [[ ! -d "$PROJECT_DIR/.git" ]]; then
  echo "ClubMax repository was not found at: $PROJECT_DIR" >&2
  exit 1
fi

cd "$PROJECT_DIR"

if [[ ! -f clubmax.env ]]; then
  echo "Missing required production configuration: $PROJECT_DIR/clubmax.env" >&2
  exit 1
fi

echo "Updating source code..."
git fetch origin
git pull --ff-only origin main

echo "Building application images..."
"${COMPOSE[@]}" build --pull

echo "Starting services and applying database migrations..."
# The backend executes its versioned PostgreSQL migrations during startup.
# --wait fails this script if a migration or health check fails.
"${COMPOSE[@]}" up -d --wait --remove-orphans

echo "Applied database migrations:"
"${COMPOSE[@]}" exec -T db sh -c \
  'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT version, name, applied_at FROM schema_migrations ORDER BY version;"'

echo "Deployment completed successfully."
