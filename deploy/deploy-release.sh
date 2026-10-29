#!/usr/bin/env bash
set -euo pipefail
export PATH=/usr/sbin:/usr/bin:/sbin:/bin
umask 077

# The deployment identity may publish only this application's bounded declaration.
# All validation and atomic publication logic belongs to ServerPortal/server-operations.
if [[ $EUID -eq 0 && $# -eq 3 && ( $1 == portal || $1 == portal-check ) && $2 =~ ^[0-9a-f]{40}$ && $3 =~ ^[0-9a-f]{64}$ ]]; then
  portal_action=register
  if [[ $1 == portal-check ]]; then portal_action=check-registration; fi
  exec timeout 45 /opt/serverportal/bin/portal "$portal_action" --app feetable --commit "$2" --sha256 "$3"
fi


# 此脚本由 root 安装和维护；本机可发送程序和受校验的本应用声明，不能更新本脚本。
if [[ $EUID -ne 0 || $# -ne 2 || ! $1 =~ ^[0-9a-f]{40}$ || ! $2 =~ ^[0-9a-f]{64}$ ]]; then
  printf 'Usage: deploy-release.sh <commit-sha> <binary-sha256> < binary\n' >&2
  exit 64
fi
commit=$1
expected=$2
app=/opt/feetable
database=$app/data/feetable.sqlite
nas=false
upgrade=/opt/server-context/scripts/upgrade-state.py
if [[ -r /etc/server-maintenance-host && $(< /etc/server-maintenance-host) == nas ]]; then
  nas=true
  test -f "$upgrade"
  exec 8>/run/lock/server-maintenance.lock
  flock -s -n 8 || { printf 'A full backup or maintenance operation is running.\n' >&2; exit 75; }
fi
exec 9>/run/lock/feetable-deploy.lock
flock -n 9 || { printf 'Another deployment is running.\n' >&2; exit 75; }
test -f "$database"
test -x "$app/bin/feetable"
install -d -m 0755 "$app/releases"
release=$(mktemp -d "$app/releases/$commit.XXXXXX")
chmod 0755 "$release"
stopped=false
replaced=false

healthy() {
  systemctl is-active --quiet feetable &&
    curl --fail --silent --show-error --max-time 3 http://127.0.0.1:18081/healthz | grep -q '"status":"ok"' &&
    curl --fail --silent --show-error --max-time 3 http://127.0.0.1/feetable/tables/1 | grep -q '/feetable/assets/'
}
wait_healthy() {
  for ((attempt=0; attempt<20; attempt++)); do
    if healthy; then return 0; fi
    sleep 1
  done
  return 1
}
finish() {
  result=$?
  trap - EXIT HUP INT TERM
  recovery_status=not-needed
  if [[ $result -ne 0 && $stopped == true ]]; then
    recovery_status=failed
    printf 'Deployment failed; restarting the previous program. Database is not restored.\n' >&2
    systemctl stop feetable || true
    if [[ $replaced == true ]]; then
      install -m 0755 "$release/previous" "$app/bin/feetable.rollback"
      mv -f "$app/bin/feetable.rollback" "$app/bin/feetable"
    fi
    if systemctl start feetable && wait_healthy; then
      recovery_status=healthy
      printf 'Previous program is healthy.\n' >&2
    else
      printf 'ROLLBACK FAILED: inspect journalctl -u feetable.\n' >&2
    fi
  fi
  printf '%s\n' "$recovery_status" > "$release/recovery"
  if [[ $result -eq 0 ]]; then printf 'success\n' > "$release/result";
  else printf 'failed\n' > "$release/result"; fi
  exit "$result"
}
trap finish EXIT
trap 'exit 130' INT
trap 'exit 143' HUP TERM

# 限制上传时间和大小；校验完成前不触碰运行中的程序。
timeout 90 head -c 67108865 > "$release/feetable"
size=$(stat -c %s "$release/feetable")
if ((size == 0 || size > 67108864)); then
  printf 'Binary must be between 1 byte and 64 MiB.\n' >&2; exit 65
fi
actual=$(sha256sum "$release/feetable")
if [[ ${actual%% *} != "$expected" ]]; then
  printf 'Binary checksum mismatch.\n' >&2; exit 65
fi
chmod 0755 "$release/feetable"
cp "$app/bin/feetable" "$release/previous"
chmod 0755 "$release/previous"
printf 'commit=%s\nsha256=%s\n' "$commit" "$expected" > "$release/metadata"

# 旧程序备份，候选程序校验；两者都以 feetable 身份运行，绝不以 root 执行上传内容。
backup="$app/backups/before-deploy-$(basename "$release").sqlite"
stopped=true
systemctl stop feetable
if [[ $nas == true ]]; then
  candidate_data=$(python3 "$upgrade" begin --app feetable --release "$(basename "$release")")
  backup="/var/lib/server-upgrades/feetable/$(basename "$release")"
  runuser -u feetable -- timeout 30 "$release/feetable" check --db "$candidate_data/feetable.sqlite"
else
  runuser -u feetable -- timeout 60 "$app/bin/feetable" backup --db "$database" --out "$backup"
  runuser -u feetable -- timeout 30 "$release/feetable" check --db "$backup"
fi
install -m 0755 "$release/feetable" "$app/bin/feetable.next"
replaced=true
mv -f "$app/bin/feetable.next" "$app/bin/feetable"
systemctl start feetable
wait_healthy
if [[ $nas == true ]]; then
  systemctl stop feetable
  python3 "$upgrade" verify --app feetable --release "$(basename "$release")"
  systemctl start feetable
  wait_healthy
fi
printf '%s\n' "$commit" > "$app/current-commit"
chmod 0644 "$app/current-commit"
if [[ $nas == true ]]; then
  python3 "$upgrade" accept --app feetable --release "$(basename "$release")"
  printf 'Deployed %s.\n' "$commit"
else printf 'Deployed %s; backup: %s\n' "$commit" "$backup"; fi
