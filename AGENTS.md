# FeeTable 维护入口

先读 [docs/README.md](docs/README.md)，按需读取架构、API、运维和 CI 文档。

- 保持单服务、单 SQLite、直接 SQL 的实现；金额与数量使用定点表示，API 数值用十进制字符串。
- 业务范围是月表、运输记录、地点/单标签与 PNG/PDF/XLSX 导出。通过统一设备认证后联网读写，服务器是正式数据来源。
- 写入、金额、日期、筛选、导出必须保持一致；测试用合成数据库，生产只读验收。
- 修改后执行 make test、go vet ./...、Linux 子路径构建；界面需实测 375×667 与桌面。
- 部署和服务器操作先读 server-operations 技能，显式执行 ssh ali 'cat /opt/AGENTS.md'；共享服务器文档位于 /opt/server-context。
- FeeTable 只维护自身的 deploy/location、unit、数据和发布身份。共享配置变更必须核对全部应用。
- 源配置与 docs 是维护来源；变更后覆盖过时描述，推送后运行 agent-config 的 `skills/server-operations/scripts/sync-docs.sh FeeTable` 同步服务器副本。
- 哪些事直接做完再告知、哪些先确认，只看 server-operations SKILL.md 的授权表；文档维护与同步不需要事先确认。
- 数据、备份、私钥、服务器公网地址不能进入 Git。main 推送会运行 CI/CD：代码改动只在用户要求部署时推送，纯文档提交加 `[skip ci]`。

## 门户资源同步

- 本项目的 `deploy/portal.json` 是 ServerPortal 资源声明的维护源，记录目录用途、数据库、运行用户、端口、unit、访问路径、API 与备份类型。新增/迁移/删除数据根、接口或运行材料时，必须同步修改声明、对应 docs 与共享应用清单。
- 声明部署在 `/opt/feetable/config/portal.json`，root 管理；CI 共用 server-operations 校验器，发布前预检，发布后通过受限 SSH 自动同步并核对门户加载哈希。`registry.d/feetable.json` 自动登记链接，更新无需重启门户。只修改门户信息时可手动运行 CI 的 `portal_only=true`，保留业务运行版本。
- 统一认证由共享 Nginx 与门户负责，不在本项目另存设备白名单；本机调用和发布健康检查按共享约定保留。生产已启用设备认证；变更后同步 server-operations current-state。
- 门户只读展示不替代本项目原生一致性备份；备份格式或媒体生命周期变化必须同时验证门户全量/增量与离线恢复。真实业务数据、凭据和备份仍不得进入 Git。
