# FeeTable 维护入口

先读 [docs/README.md](docs/README.md)，按需读取架构、API、运维和 CI 文档。

- 保持单服务、单 SQLite、直接 SQL 的实现；金额与数量使用定点表示，API 数值用十进制字符串。
- 业务范围是月表、运输记录、地点/单标签与 PNG/PDF/XLSX 导出。无登录，联网读写，服务器是正式数据来源。
- 写入、金额、日期、筛选、导出必须保持一致；测试用合成数据库，生产只读验收。
- 修改后执行 make test、go vet ./...、Linux 子路径构建；界面需实测 375×667 与桌面。
- 部署和服务器操作先读 server-operations 技能，显式执行 ssh ali 'cat /opt/AGENTS.md'；共享服务器文档位于 /opt/server-context。
- FeeTable 只维护自身的 deploy/location、unit、数据和发布身份。共享配置变更必须核对全部应用。
- 源配置与 docs 是维护来源；变更后覆盖过时描述，并同步服务器 /opt/feetable/docs、AGENTS.md 和 SOURCE。
- 数据、备份、私钥、服务器公网地址不能进入 Git。main 推送会运行 CI/CD；提交推送依本次授权执行。
