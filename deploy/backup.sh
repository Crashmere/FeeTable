#!/usr/bin/env bash
set -euo pipefail
umask 077
feetable_binary="${FEETABLE_BINARY:-/opt/feetable/bin/feetable}"
feetable_database="${FEETABLE_DB:-/opt/feetable/data/feetable.sqlite}"
feetable_backups="${FEETABLE_BACKUP_DIR:-/opt/feetable/backups}"
mkdir -p "$feetable_backups"
backup_path="$feetable_backups/daily-$(date -u +%Y%m%dT%H%M%S)-$$.sqlite"
"$feetable_binary" backup --db "$feetable_database" --out "$backup_path"

# 只有新备份成功并通过完整性检查后才轮换；手工/升级/恢复前备份不参与。
mapfile -t daily_backups < <(find "$feetable_backups" -maxdepth 1 -type f -name 'daily-*.sqlite' -printf '%f\n' | LC_ALL=C sort -r)
for ((index=14; index<${#daily_backups[@]}; index++)); do
  rm -- "$feetable_backups/${daily_backups[index]}"
done
printf 'Backup created: %s\n' "$backup_path"
