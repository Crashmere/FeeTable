# HTTP API

以下是 Go 服务内部路径，`{id}`、`{record}` 为路径参数。部署后外部路径统一增加 `/feetable`，Nginx 去前缀后转发。公网 API 由 ServerPortal 统一设备认证保护，未授权返回 401；本机回环调用保留。JSON 响应禁用缓存，浏览器写请求另检查 Origin 的协议、Host 与请求是否一致。

| 方法与路径 | 用途 |
| --- | --- |
| GET /healthz | 数据库连接检查 |
| GET /api/tables | 月表列表、条数、金额，30 条/页；查询参数 page（默认 1）、sort（默认 updated_desc），可附 year/month 筛选 |
| POST /api/tables | 创建，year/month 必填 |
| GET /api/tables/{id} | 表格、明细、全部标签、筛选条数/合计，50 条/页；查询参数 page（默认 1）、tag（可选） |
| PUT /api/tables/{id} | 修改 year/month，revision 必须匹配 |
| DELETE /api/tables/{id} | 删除表格与明细，查询参数 revision 必须匹配当前表格版本 |
| POST /api/tables/{id}/merge | 将同年月的来源表合并到 id 保留表 |
| POST /api/tables/{id}/records | 新增明细，body 包含当前表格 revision |
| PUT /api/tables/{id}/records/{record} | 完整更新明细，body 包含当前表格 revision |
| DELETE /api/tables/{id}/records/{record} | 删除明细，查询参数 revision 必须匹配当前表格版本 |
| GET /api/locations | 常用地点，按累计使用次数排序 |
| GET /api/tags | 常用标签，按累计使用次数排序 |
| POST /api/locations | 添加常用地点，body 为 {name}，已存在时不重复创建 |
| POST /api/tags | 添加常用标签，body 为 {name}，已存在时不重复创建 |
| PUT /api/locations/{id} | 地点改名，body 为 name、previousName |
| DELETE /api/locations/{id} | 删除常用地点，查询参数 name 为原名称，用于检查并发修改 |
| GET /api/tables/{id}/report | 完整导出数据与 revision；可选查询参数 tag，不分页 |
| GET /api/tables/{id}/export | 下载；查询参数 format 为 png/pdf/xlsx，可选 revision、tag，inline=1 可内联打开 |

本表覆盖当前业务与健康接口；门户的 `deploy/portal.json.apis` 与本表保持一致。每个方法和路径单独登记，查询参数写在用途和下文契约中。 核对与同步流程见[共享排障说明](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/common-issues.md#门户-api-列表与服务不一致)。

## 月表排序、合并与地点管理

sort 可为 updated_desc（默认）、updated_asc、month_desc、month_asc，分别表示修改时间或表格年月倒序/正序。year/month 必须同时传入且有效，筛选后的 total 与分页一致；不支持的排序返回 400。

合并请求包含 revision 和 sources 数组；每个来源项包含 id、revision。路径 id 是保留表，顶层 revision 是它的版本；sources 包含 1–99 个不同来源表及各自版本，不能含保留表。成功返回保留表的新 revision、记录数与总额，来源表随后返回 404。年月必须全部相同，合计记录不超过 5,000 条；任意版本冲突返回 409，缺失表返回 404，全部操作原子回滚。记录包括重复内容均完整迁入。

地点改名、删除保留已有记录快照；改名保留使用次数，重复名称返回 400。previousName/name 必须匹配服务器当前名称，否则 409；不存在的地点返回 404。删除的名称以后再次用于保存记录时会重新加入地点库。

## 记录输入

所有字段必传，unitPrice/tag 可以为 null，其他字段不能为 null。日期必须属于表格年月，名称 1–80 字符。数值使用普通十进制字符串，不接受指数、NaN、Infinity 或分组逗号。

```json
{
  "day": 3,
  "location1": "测试起点",
  "location2": "测试终点",
  "quantity": "0.145",
  "unitPrice": "1",
  "amount": "0.15",
  "tag": "测试标签",
  "revision": 1
}
```

unitPrice 非 null 时按 quantity 计算，amount 仅作完整表单字段，不决定保存金额；unitPrice=null 时使用 amount。金额与数量响应固定两位/三位小数，单价响应固定两位小数字符串。单价输入允许非负整数或最多两位小数（例如 "2.88"）；数量和固定金额允许负数及零。revision 是所属表格的版本，不是明细自己的版本。

写成功返回已保存实体或 `{"ok":true}`。客户端随后重读表格取得新 revision。tag 以 URLSearchParams 编码，省略表示全部记录，传入则精确匹配；标签中的斜线等字符不是路由路径。

## 错误

应用错误结构为 `{"error":{"code":"VALIDATION","message":"说明"}}`。400 为字段/日期/格式错误，403 为写请求来源不符，404 为不存在，409 为版本冲突，422 为导出尺寸限制，429 为渲染繁忙，500 为内部失败；公网设备认证失败由共享入口返回 401。写入网络失败或 500 不能自动重试，应先查看服务器是否已保存。

每次 JSON 输入上限 64 KiB。导出响应带明确 MIME、Content-Disposition 和 no-store；提交非零 revision 且与当前版本不同时返回 409，省略或为 0 时不做该检查。客户端从预览取得 revision 后应原样提交。空筛选集拒绝导出，页面关闭相应下载按钮。
