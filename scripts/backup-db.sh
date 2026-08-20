#!/usr/bin/env bash
# Copies the live SQLite DB out of the kylistran-api container onto the
# external drive, with a rotation window and a basic integrity check.
set -euo pipefail

# --- config: adjust to match your NUC ---
CONTAINER_NAME="kylistran-api"
DB_PATH_IN_CONTAINER="/data/kylistran.db"
BACKUP_DIR="/mnt/backup/kylistran"   # <-- point this at the external drive's mount path
RETENTION_DAYS=30
# ---

mkdir -p "$BACKUP_DIR"

timestamp="$(date +%Y%m%d-%H%M%S)"
dest="$BACKUP_DIR/kylistran-$timestamp.db"

docker cp "$CONTAINER_NAME:$DB_PATH_IN_CONTAINER" "$dest"

if [ ! -s "$dest" ]; then
    echo "backup failed: $dest is empty or missing" >&2
    rm -f "$dest"
    exit 1
fi

if command -v sqlite3 >/dev/null 2>&1; then
    result="$(sqlite3 "$dest" 'PRAGMA integrity_check;')"
    if [ "$result" != "ok" ]; then
        echo "backup failed integrity check: $dest ($result)" >&2
        exit 1
    fi
fi

find "$BACKUP_DIR" -name 'kylistran-*.db' -type f -mtime "+$RETENTION_DAYS" -delete

echo "backup ok: $dest"
