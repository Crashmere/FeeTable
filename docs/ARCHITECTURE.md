# 架构与代码导读

```text
浏览器 → Nginx /feetable/ → Go HTTP → SQLite
                              ├─ 内嵌 Vue 页面
                              └─ PNG / PDF / XLSX
```

Go 入口在 cmd/feetable；internal/httpapi 解码请求、检查来源并输出 JSON/文件；internal/feetable 承担领域规则和 SQL；internal/export 处理导出。web 使用 Vue 3、Vue Router、TypeScript 和 Vite。

## 保存与数据

四张业务表为 fee_tables、fee_records、locations、tags。每张月表有 revision，新增/修改/删除明细或更改年月都会增加版本。客户端提交读取时的版本，冲突返回 409；所有多步写入使用事务。SQLite 启用 WAL、外键、FULL 同步及单连接。

地点与标签保存在记录中的文本是当次快照；建议库是全局唯一名称列表，按累计新增使用次数排序。新增记录会补词条并加计数，编辑补词条但不加计数，删除不减少累计次数。名称 trim，空标签转 null，拒绝控制字符。

quantity_milli、unit_price_cents、amount_cents 均为整数。输入十进制字符串直接解析；乘法使用大整数中间值，按量金额由后端计算，不采用客户端传入的金额。合计是逐条已保存金额相加，响应返回固定小数位字符串。浏览器只做 BigInt 预览和格式显示。

月表列表按 updated_at、id 倒序；明细按 day、id 升序。月表每页 30 条，明细每页 50 条，合计与计数覆盖完整筛选集。读取头部、明细与汇总使用一致的读事务。

数据库 application_id 为 1179931714，user_version 为 1；正常启动只打开现有库并验证身份和完整性。init 要求目标不存在。升级 schema 必须有明确的迁移与回退策略。

## 界面与保存反馈

Home 是月表列表，TablePage 管理明细与筛选，ExportPage 展示完整报表。手机用记录卡片和竖向表单，电脑用表格；不把宽导出表直接用作编辑界面。

保存期间阻止重复提交、模态框关闭和站内离开；服务器成功后才清空表单。失败保留输入，网络中断或 5xx 视为结果不明，禁用当前表单重试并要求先核对。读请求提供重试。没有离线写入队列和业务数据缓存。

## 导出

Report 包含表格、完整筛选明细和整数计算后的合计。下载可附 revision，变化时要求刷新预览。文件渲染在 SQL 事务结束后进行；一个并发槽控制生成任务，繁忙返回 429。

PNG 与 PDF 共用布局：七列、合并标题与表头、固定宽度、长文字换行；字体为内嵌 LXGW Neo XiHei。PNG 使用 Go image，超过 20,000,000 像素拒绝生成。PDF 使用 gopdf 嵌入字体与 Unicode 映射；XLSX 使用 Excelize，共用金额与数值格式。Excel 的文字字段按字符串写入。

导出文件在内存生成，响应 no-store，按表 ID、年月、标签形成文件名，不持久保存到公共目录。浏览器下载后可另存；原生分享仅在支持时展示。

## 构建与子路径

BASE_PATH 同时决定 Vite 资源、路由 history base 与 API 前缀；Nginx proxy_pass 尾斜线去掉 /feetable/，Go 接收 /api、/assets 等内部路径。未知页面返回 index.html，未知静态文件返回 404。

go:embed 将生产网页和中文字体打包进二进制。正常服务器不运行 Node、Go 编译器或外部数据库。测试位于各 Go 包、money.test.ts 和 deploy-release.test.mjs，使用隔离合成数据。
