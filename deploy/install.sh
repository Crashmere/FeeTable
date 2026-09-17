#!/usr/bin/env bash
set -euo pipefail
if [[ "$EUID" -ne 0 || $# -ne 2 ]]; then
  printf 'Usage: sudo bash deploy/install.sh <linux-binary> <verified-initial-database>\n' >&2
  exit 1
fi
feetable_binary="$(realpath "$1")"
feetable_initial="$(realpath "$2")"
feetable_deploy="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
if [[ -e /opt/feetable || -e /etc/systemd/system/feetable.service ]]; then
  printf 'Existing installation found. Follow the documented upgrade procedure.\n' >&2
  exit 1
fi
"$feetable_binary" check --db "$feetable_initial"
if ss -H -ltn 'sport = :18081' | grep -q .; then
  printf 'Port 18081 is already in use. Choose a separate application port first.\n' >&2
  exit 1
fi
if ! id feetable >/dev/null 2>&1; then
  useradd --system --home-dir /opt/feetable --shell /usr/sbin/nologin feetable
fi
install -d -m 0755 /opt/feetable /opt/feetable/bin /opt/feetable/config
install -d -m 0700 -o feetable -g feetable /opt/feetable/data /opt/feetable/backups
install -m 0755 "$feetable_binary" /opt/feetable/bin/feetable
install -m 0755 "$feetable_deploy/backup.sh" /opt/feetable/bin/backup.sh
install -m 0644 "$feetable_deploy/feetable.env" "$feetable_deploy/nginx-location.conf" /opt/feetable/config/
install -m 0644 "$feetable_deploy/feetable.service" "$feetable_deploy/feetable-backup.service" "$feetable_deploy/feetable-backup.timer" /opt/feetable/config/
# restore 创建独立一致性副本，不直接复制可能带 WAL 的源主文件。
/opt/feetable/bin/feetable restore --from "$feetable_initial" --db /opt/feetable/data/feetable.sqlite
chown feetable:feetable /opt/feetable/data/feetable.sqlite
systemctl link /opt/feetable/config/feetable.service /opt/feetable/config/feetable-backup.service /opt/feetable/config/feetable-backup.timer
systemctl daemon-reload
systemctl enable --now feetable.service feetable-backup.timer
systemctl start feetable-backup.service
curl --fail --silent --retry 10 --retry-delay 1 --retry-connrefused http://127.0.0.1:18081/healthz
printf '\nFeeTable installed. Enable /opt/feetable/config/nginx-location.conf in the shared Nginx server after verification.\n'
