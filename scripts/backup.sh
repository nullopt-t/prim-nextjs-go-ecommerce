#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="$ROOT_DIR/backups"
COMPOSE=(docker compose --env-file "$ROOT_DIR/backend/.env" -f "$ROOT_DIR/backend/docker/docker-compose.prod.yml")

mkdir -p "$BACKUP_DIR"
out="$BACKUP_DIR/prim_$(date +%Y%m%d_%H%M%S).sql.gz"
tmp="$out.tmp"

# --clean --if-exists makes the dump restorable into the existing database.
"${COMPOSE[@]}" exec -T postgres pg_dump -U prim -d prim --clean --if-exists | gzip > "$tmp"
mv "$tmp" "$out"
echo "Backup written to $out ($(du -h "$out" | cut -f1))"

find "$BACKUP_DIR" -maxdepth 1 -type f -name 'prim_*.sql.gz' -mtime +7 -exec rm -f {} +
echo "Pruned backups older than 7 days"
