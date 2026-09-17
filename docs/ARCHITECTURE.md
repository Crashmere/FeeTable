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

地点管理的改名、删除仅操作 locations，保留记录快照及表 revision。改名保留累计使用次数并拒绝重名；以客户端读取的 previousName/name 检查并发改名，冲突返回 409。没有新增数据库列或 schema 迁移。

quantity_milli、unit_price_cents、amount_cents 均为整数。输入十进制字符串直接解析；乘法使用大整数中间值，按量金额由后端计算，不采用客户端传入的金额。合计是逐条已保存金额相加，响应返回固定小数位字符串。浏览器只做 BigInt 预览和格式显示。

月表列表支持 updated_desc（默认）、updated_asc、month_desc、month_asc；修改时间排序以 ID 同向打破平局，年月排序以 updated_at、ID 倒序打破平局。排序在服务器分页前完成，可按年月筛选合并候选。明细按 day、id 升序。月表每页 30 条，明细每页 50 条，合计与计数覆盖完整筛选集。读取头部、明细与汇总使用一致的读事务。

合并请求指定保留表及 1–99 张来源表，包含每张表的 revision。在单个事务中验证存在、版本、年月相同、ID 不重复及合并后记录不超过 5,000 条，再更新来源记录 table_id、删除已空的来源表并增加保留表 revision/updated_at。记录 ID、日期、文本、标签、定点值及建议库计数不变，不去重或重新计算金额。任意失败回滚全部写入；旧预览和旧表单会被 revision 检查拦截。

数据库 application_id 为 1179931714，user_version 为 2；正常启动只打开现有库并验证身份、完整性和当前版本。init 要求目标不存在。migrate 在停止服务并备份后显式运行，在一个事务中将 v1 的明细表约束升级为支持分单位价，保留数据、ID、修订版本、自增序列与索引。check/backup/restore 支持 v1 与 v2，serve 要求 v2；具体步骤和回退边界见 OPERATIONS。

## 界面与保存反馈

Home 是月表列表，提供排序、同月合并及地点管理入口；LocationsPage 管理常用地点；TablePage 管理明细与筛选，ExportPage 展示完整报表。手机用记录卡片和竖向表单，电脑用表格；不把宽导出表直接用作编辑界面。合并先选择来源表，再展示条数、金额和来源表移除确认。

所有 Modal 通过 Teleport 放在 body 下，打开前固定背景 body 并保存滚动位置，弹窗以 overscroll-behavior 阻止边缘滚动传递。卸载后恢复原样式与位置；锁支持重入，避免重叠生命周期提前释放。月表列表和明细刷新期间保留现有内容，避免关闭弹窗时页面短暂变矮而丢失滚动位置。

界面文案只保留业务信息、字段、操作和必要状态反馈；不添加宣传语、装饰性副标题、页脚口号或重复说明。年月并入日期字段，单价留空的含义在输入框提示，校验错误与删除确认仍明确显示。

保存期间阻止重复提交、模态框关闭和站内离开；服务器成功后才清空表单。失败保留输入，网络中断或 5xx 视为结果不明，禁用当前表单重试并要求先核对。读请求提供重试。没有离线写入队列和业务数据缓存。

## 导出

Report 包含表格、完整筛选明细和整数计算后的合计。下载可附 revision，变化时要求刷新预览。文件渲染在 SQL 事务结束后进行；一个并发槽控制生成任务，繁忙返回 429。

PNG 与 PDF 共用布局：七列、合并标题与表头、固定宽度、长文字换行；字体为内嵌 LXGW Neo XiHei。PNG 按逻辑尺寸换行，再以 3 倍尺寸直接绘制字体与网格（宽 2,832 像素）；实际像素超过 80,000,000 时改用 2 倍（宽 1,888 像素），仍超限则拒绝生成。采用 Go image.Gray，每像素 1 字节并保留文字抗锯齿，位图最多 80 MB；逻辑长表容量不变。PDF 使用 gopdf 嵌入字体与 Unicode 映射；XLSX 使用 Excelize，共用金额与数值格式。Excel 的文字字段按字符串写入。

导出文件在内存生成，响应 no-store，按表 ID、年月、标签形成文件名，不持久保存到公共目录。浏览器下载后可另存；原生分享仅在支持时展示。

## 构建与子路径

BASE_PATH 同时决定 Vite 资源、路由 history base 与 API 前缀；Nginx proxy_pass 尾斜线去掉 /feetable/，Go 接收 /api、/assets 等内部路径。未知页面返回 index.html，未知静态文件返回 404。

go:embed 将生产网页和中文字体打包进二进制。正常服务器不运行 Node、Go 编译器或外部数据库。测试位于各 Go 包、money.test.ts 和 deploy-release.test.mjs，使用隔离合成数据。
