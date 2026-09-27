#!/usr/bin/env bash
set -euo pipefail
# authorized_keys 强制进入这里，不提供 shell、scp 或任意 sudo 命令。
if [[ ${SSH_ORIGINAL_COMMAND:-} =~ ^deploy\ ([0-9a-f]{40})\ ([0-9a-f]{64})$ ]]; then
  exec sudo -n /opt/feetable/bin/deploy-release.sh "${BASH_REMATCH[1]}" "${BASH_REMATCH[2]}"
fi
if [[ ${SSH_ORIGINAL_COMMAND:-} =~ ^(portal|portal-check)\ ([0-9a-f]{40})\ ([0-9a-f]{64})$ ]]; then
  exec sudo -n /opt/feetable/bin/deploy-release.sh "${BASH_REMATCH[1]}" "${BASH_REMATCH[2]}" "${BASH_REMATCH[3]}"
fi
printf 'Only deploy, portal-check or portal with a commit and SHA-256 are allowed.\n' >&2
exit 64
