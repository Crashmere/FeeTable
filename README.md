# 运费明细表

轻量自托管运费记录服务。按年月建立表格，记录运输明细，按整表或标签导出 PNG、PDF 和 Excel。支持手机与电脑，Go 程序内嵌 Vue 页面和中文导出字体，SQLite 保存数据。

## 功能

- 月表创建、修改年月和删除；同年月可建立多张表。
- 明细包含日、两个地点、三位小数数量、两位小数单价、金额和一个可选标签。
- 有单价时自动计算金额并四舍五入到分；单价留空时手填固定费用。
- 地点与标签可直接输入、快速添加和按使用频率选择。
- 按日期排序、分页浏览、按标签筛选与完整导出。
- PNG、长页 PDF、XLSX 保留七列表格和合并表头，数量三位、金额两位。
- 事务保存、编辑冲突检查、一致性备份与独立服务部署。

## 本地运行

需要 go.mod 指定的 Go 工具链和 Node.js 24 或更高版本。依赖版本锁定在 go.sum 和 web/package-lock.json。

```sh
npm --prefix web ci
go mod download
go run ./cmd/feetable init --db var/dev.sqlite
make build
./bin/feetable serve --db var/dev.sqlite
```

访问 `http://127.0.0.1:8081/`。init 只用于创建新库，目标已存在会拒绝覆盖。启动服务要求数据库已经存在。

前端开发可在两个终端分别运行 `make dev` 和 `npm --prefix web run dev`，由 Vite 将 API 请求代理到 8081。

```sh
make test
go vet ./...
make linux BASE_PATH=/feetable/
```

Linux 产物为 `bin/feetable-linux-amd64`，服务器不需要编译器、前端进程或独立数据库服务。若本机版本管理器阻止 Go 自动选择工具链，可用 `make GO=/path/to/go` 指定支持自动工具链的 Go 可执行文件。

## 使用边界

无登录，知道访问地址的人可以查看、修改和导出。需要联网；服务器数据库是正式数据源。浏览器不保存业务数据或离线写入队列。

每张表最多 5,000 条记录；PNG 最多 2,000 万像素，超过限制可按标签分开导出或使用 PDF/Excel。手机通过浏览器下载或打开图片后保存到相册，系统分享取决于浏览器和安全上下文支持。

## 维护文档

从 [docs/README.md](docs/README.md) 阅读架构、API、运维和 CI。中文字体的来源与许可见 [字体说明](internal/export/fonts/README.md)。
