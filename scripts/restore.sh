#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE=(docker compose --env-file "$ROOT_DIR/backend/.env" -f "$ROOT_DIR/backend/docker/docker-compose.prod.yml")

file="${1:-}"
if [[ -z "$file" ]]; then
  echo "Usage: $0 <backups/prim_YYYYmmdd_HHMMSS.sql.gz>" >&2
  exit 1
fi
if [[ ! -f "$file" ]]; then
  echo "Backup file not found: $file" >&2
  exit 1
fi

read -r -p "Restore $file? This DROPS and recreates all tables. [y/N] " answer
if [[ "$answer" != "y" && "$answer" != "Y" ]]; then
  echo "Aborted"
  exit 1
fi

gunzip -c "$file" | "${COMPOSE[@]}" exec -T postgres psql -U prim -d prim -v ON_ERROR_STOP=1
echo "Restore complete from $file"
