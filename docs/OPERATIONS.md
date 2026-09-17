# 运行与维护

FeeTable 在 SSH 别名 ali 对应的服务器上使用独立目录、运行用户、数据库和发布身份。开始操作前读 server-operations 与远端 /opt/AGENTS.md，核对共享应用清单。访问方式为 HTTP /feetable/，用户已确认无登录、知址可读写和导出。公网地址不写入仓库。

## 运行配置

| 项目 | 配置 |
| --- | --- |
| 程序 | /opt/feetable/bin/feetable，内嵌网页和中文字体 |
| 环境文件 | /opt/feetable/config/feetable.env |
| 监听 | FEETABLE_ADDR=127.0.0.1:18081 |
| 数据库 | FEETABLE_DB=/opt/feetable/data/feetable.sqlite |
| 备份目录 | /opt/feetable/backups |
| 用户 | feetable 运行；feetable-deploy 仅发布 |
| 常驻 | feetable.service，开机自启 |
| 每日备份 | feetable-backup.timer，03:15 Asia/Shanghai，随机延迟最多 5 分钟 |
| Nginx | /etc/nginx/app-locations/feetable.conf 链接到项目 config/nginx-location.conf |
| 项目文档 | /opt/feetable/AGENTS.md、/opt/feetable/docs，来源见 docs/SOURCE |

程序、脚本、配置归 root；data/backups 归 feetable，目录 0700、数据库 0600。服务仅获 data 写权限，备份 unit 单独获 data/backups 写权限。不要把程序或配置交给运行用户修改。内部 18081 不对外开放。

## 从源码首次安装

本地或受信 CI 需要 go.mod 指定的 Go 工具链、Node 24+、npm。服务器运行单个 Linux amd64 程序，无需安装 Go、Node、Docker 或数据库服务。

    npm --prefix web ci
    make test
    go vet ./...
    make linux BASE_PATH=/feetable/
    sha256sum bin/feetable-linux-amd64

macOS 可用 shasum -a 256。仅将校验过的 Linux 产物和已审阅提交中的 deploy 文件导出到独立临时目录，再传到服务器；不传本地 var、凭据或测试库。

先核对 18081、/feetable/、/opt/feetable 和账户没有冲突。从已上传目录运行：

    chmod 0755 feetable-linux-amd64
    ./feetable-linux-amd64 init --db initial.sqlite
    ./feetable-linux-amd64 check --db initial.sqlite
    bash deploy/install.sh ./feetable-linux-amd64 ./initial.sqlite
    ln -s /opt/feetable/config/nginx-location.conf /etc/nginx/app-locations/feetable.conf
    nginx -t
    systemctl reload nginx

这次系统从空库开始。init 只创建不存在的目标；serve 永不自动创建数据库。install.sh 拒绝覆盖已有安装，restore 生成独立一致性副本。共享 Nginx server 由 server-operations 管理，不用本项目替换。若 nginx -t 失败，只撤回本次新增的 location 链接并调查，不重载无效配置。

首次安装后，按 CICD 配置独立发布密钥并完成一次 CI 发布。current-commit 由发布脚本在健康检查成功后写入；不要把文档提交误写成运行版本。

## 验收和排障

    systemctl is-active feetable nginx ledger
    systemctl is-enabled feetable feetable-backup.timer
    curl -fsS http://127.0.0.1:18081/healthz
    curl -fsS http://127.0.0.1/feetable/healthz
    curl -fsS http://127.0.0.1/feetable/tables/1
    curl -fsS http://127.0.0.1/ledger/healthz
    journalctl -u feetable -n 50 --no-pager
    systemctl list-timers feetable-backup.timer
    journalctl -u feetable-backup -n 30 --no-pager

页面验收同时检查裸路径 308、资源 /feetable/assets/、深链接、API 和跨站写入拒绝。生产只读验收；增删改、断网、分页、导出使用隔离合成库。手机检查 375×667，电脑检查桌面宽度。

启动失败先看日志、数据库是否存在、所有者、完整性和端口。不要删除库或执行 init 掩盖故障。HTTP 409 是并发版本冲突；保存超时或 5xx 后先核对，不能盲目重试。导出 429 表示另一份文件正在生成；PNG 超出像素上限时改用 PDF/XLSX 或按标签拆分。

## 一致性备份

backup 使用 VACUUM INTO，包含已提交的 WAL 数据并校验生成文件。禁止只复制活动数据库主文件。每日备份成功后保留最新 14 份 daily；手工、发布前备份另存且不自动轮换。

    systemctl start feetable-backup.service
    systemctl show feetable-backup.service -p Result -p ExecMainStatus
    runuser -u feetable -- /opt/feetable/bin/feetable check --db /opt/feetable/backups/<备份文件>.sqlite

数据库与备份不可进入 Git。当前只配置同服务器同盘备份，没有异机备份。发布历史和发布前备份须定期关注磁盘。

## 恢复与演练

restore 只写一个不存在的新文件，不直接覆盖生产。演练在独立目录执行：

    runuser -u feetable -- /opt/feetable/bin/feetable restore --from /opt/feetable/backups/<备份文件>.sqlite --db /opt/feetable/data/<独立演练文件>.sqlite
    runuser -u feetable -- /opt/feetable/bin/feetable check --db /opt/feetable/data/<独立演练文件>.sqlite

真实恢复需先取得用户对恢复时点及其后数据损失的确认。管理员应暂停发布和备份调度、停止服务，用当前程序备份现状，检查来源并恢复到新文件；保留当前主文件及其 WAL/SHM 作为一组，再原子替换为验证后的新库，设置 feetable 所有者和 0600，重新启动、恢复调度并只读验收。不要把旧 WAL/SHM 留在新库旁边，不在服务运行时覆盖文件。恢复失败保留现场，停止继续写入并调查。

## 发布、配置和文档同步

当前数据库版本为 2，单价以分保存，支持两位小数。已有 v1 库须由管理员完成一次停服升级，普通发布和 serve 不会隐式升级：

1. 从已验证提交构建候选程序，核对哈希。用现有程序备份，并先在恢复出的独立副本上运行候选程序的 migrate、check 和 serve，验证可启动。
2. 持有 /run/lock/feetable-deploy.lock，暂停备份 timer，确认备份 service 已结束，再停止 feetable。用现有程序生成最终升级前备份并校验。
3. 以 feetable 用户运行候选程序的 migrate --db /opt/feetable/data/feetable.sqlite，再运行 check。升级在单一事务中完成；失败回滚事务，不会留下部分新表。原始分值直接保留，不乘除已有单价。
4. 安装候选程序后启动服务，检查本机和代理健康，更新 current-commit，恢复备份 timer，再进行普通 CI 发布。

v1 的程序不支持 v2 数据库，也会截断小数单价，不能在升级后直接回退到 v1 程序。升级后应以 v2 兼容程序修复；若确实需要回退数据，必须先按恢复流程确认时点及数据损失。保留升级前备份，不自动覆盖已经接受新写入的数据库。新程序的 check/backup/restore 可读取 v1 备份；恢复后需显式 migrate 到 v2 才能 serve。

日常程序发布与自动回退见 [CICD.md](CICD.md)。它只替换二进制；配置、unit、发布脚本、文档仍由管理员从已审阅提交部署。配置更新先对比现场与源码，安装后检验语法及权限，按需 daemon-reload 或 nginx -t 后重载。影响共享入口时同时核对 Ledger。

文档同步按 server-operations 的 maintenance 流程：git ls-files 审核 AGENTS.md 和 docs/*.md 白名单，从确认提交 git archive 导出，逐文件安装至 /opt/feetable，最后生成 docs/SOURCE（repository、commit、subdirectory、synced_at）并比对 SHA-256。不要整目录上传工作区。共享清单与入口属于 agent-config，需独立提交、同步 /opt/server-context；未改变的 Ledger 源码和文档无需跟随重写。
