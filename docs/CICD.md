# 自动检查与部署

GitHub Actions 的 CI and deploy 工作流负责日常发布：

- Pull Request：运行 Go/前端/发布脚本测试、类型检查、go vet 和 /feetable/ 生产构建；不使用生产密钥。
- 推送 main：同样检查通过后上传 Linux 程序，通过 SSH 更新服务器。
- Actions 页面也可手动运行；只有 main 允许部署。过期提交在发布前会被跳过。同一分支工作流串行，服务器再用 flock 防止两次发布重叠。

FeeTable 不需要服务器安装 Go、Node、Docker 或 GitHub Runner；主机已有其他用途的工具不属于本项目依赖，也不要因此删除。Actions 使用 GitHub 官方托管 runner，官方 actions 固定到 commit SHA。

## 发布做什么

1. CI 构建 make linux BASE_PATH=/feetable/，产物只含程序与内嵌网页，保留 7 天。
2. SSH 将程序从标准输入传给受限入口，同时传入代码 commit 和文件 SHA-256。
3. 服务器限制上传大小 64 MiB、超时 90 秒，校验哈希后才开始切换。
4. 保存旧程序、停止 FeeTable，用旧程序生成升级前一致性备份。
5. 候选程序校验备份，原子替换程序并启动，检查本机 /healthz 与 Nginx 下 /feetable/tables/1。
6. 成功后记录 /opt/feetable/current-commit；失败则恢复旧程序并重启，返回失败使 Actions 标红。

通常只有备份与重启期间短暂停服。不会上传或覆盖数据库，不更新 Nginx、systemd、环境配置或发布脚本本身。没有数据库自动回退：候选程序若改变 schema，需要单独设计迁移和恢复流程，不能指望旧程序一定兼容。

当前 schema 为 v2。首次从 v1 升级须先按 OPERATIONS 完成管理员停服迁移和兼容程序切换，再恢复普通发布；自动回退只适用于兼容当前 schema 的程序。禁止直接部署仅支持 v1 的提交来回退。

文档也不随二进制自动上传。当前任何 main 推送（包括只改 Markdown）都会执行检查和发布；只更新文档且不需要发布时可使用 GitHub 支持的 `[skip ci]` 提交标记，先确认本次确实没有运行代码或部署逻辑改动。共享/项目文档需管理员按 [OPERATIONS.md](OPERATIONS.md) 的同步流程更新服务器副本，不能只推 GitHub 就认为服务器文档已更新。

## 权限与密钥

production 环境应仅允许 main，保存四个 secrets：SSH_HOST、SSH_USER、SSH_PRIVATE_KEY、SSH_KNOWN_HOSTS。主机公钥经现有受信 SSH 连接核对，部署时开启 StrictHostKeyChecking，不临时盲信 ssh-keyscan。

feetable-deploy 用户的 SSH 公钥设置 restrict 和强制命令，只接受 deploy <commit-sha> <binary-sha256>，不允许 shell、scp、端口转发。home、authorized_keys、两个发布脚本均由 root 管理；sudo 只允许固定 deploy-release.sh。

上传的程序以及备份命令始终以 feetable 用户执行，不以 root 执行。该密钥有发布任意 FeeTable 程序的能力，因此等同于 FeeTable 应用/数据权限，不等同于整台服务器管理员权限。勿把不可信代码合入 main。

首次配置由管理员运行 deploy/setup-ci.sh <public-key.pub>，随后设置 GitHub secrets。它拒绝覆盖已有账户；轮换密钥时显式替换 authorized_keys 和 SSH_PRIVATE_KEY。变更发布脚本需管理员审阅后安装，不能由一次普通发布自行替换。

## 查看状态和回退

GitHub 仓库 Actions 页面查看测试、构建和 SSH 日志。服务器保存：

```text
/opt/feetable/current-commit
/opt/feetable/releases/<commit>.<随机后缀>/
  feetable        本次程序
  previous      部署前程序
  metadata      commit 和校验和
  result        success / failed
/opt/feetable/backups/before-deploy-<发布目录名>.sqlite
```

发布历史和升级前备份暂不自动清理，不受每日 14 份备份轮换影响。定期检查磁盘，确认无需回退后再按明确目录清理。

失败先看 Actions 和 journalctl -u feetable。若日志显示 Previous program is healthy，旧程序已恢复，账本未回退。若显示 ROLLBACK FAILED，按 OPERATIONS.md 排查。手工回退程序也应先停服和备份；数据库恢复需人工确认恢复时点。

deploy 的 SSH 步骤以 exit code 124 结束、最新发布目录为 failed 且没有 metadata，是 GitHub runner 到服务器的上传超时。不要反复重跑，直接按共享的 [GitHub 上传过慢时的备用发布](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/common-issues.md#github-上传过慢时的备用发布)处理（服务器副本 `/opt/server-context/references/common-issues.md`），其中也包括残留清理。FeeTable 的参数：产物取自失败的同一次 CI and deploy run，artifact `feetable-linux`，文件 `feetable-linux-amd64`，验收 `/feetable/healthz` 与 `/feetable/tables/1`，不写测试数据。

测试命令 node --test deploy/deploy-release.test.mjs 使用隔离目录和合成命令，覆盖成功、上传哈希错误、候选校验失败、健康检查失败；不会连接生产或读取真实数据，已纳入 make test。
