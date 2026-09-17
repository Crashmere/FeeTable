export interface Table {
  id: number;
  year: number;
  month: number;
  createdAt: number;
  updatedAt: number;
  revision: number;
  recordCount: number;
  total: string;
}
export interface FeeRecord {
  id: number;
  tableId: number;
  day: number;
  location1: string;
  location2: string;
  quantity: string;
  unitPrice: string | null;
  amount: string;
  tag: string | null;
}
export interface RecordInput {
  day: number;
  location1: string;
  location2: string;
  quantity: string;
  unitPrice: string | null;
  amount: string;
  tag: string | null;
  revision: number;
}
export interface Suggestion {
  id: number;
  name: string;
  selectionCount: number;
}
export interface Report {
  table: Table;
  records: FeeRecord[];
  tags: string[];
  filterTag: string | null;
  count: number;
  total: string;
  page: number;
  pageSize: number;
}
export interface TableList {
  items: Table[];
  total: number;
  page: number;
  pageSize: number;
}
export class ApiError extends Error {
  constructor(
    message: string,
    public uncertain = false,
    public code = "",
  ) {
    super(message);
  }
}
export const base = import.meta.env.BASE_URL;
export function query(
  values: Record<string, string | number | null | undefined>,
) {
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(values))
    if (v != null) q.set(k, String(v));
  return q.size ? "?" + q : "";
}
export async function request<T>(
  path: string,
  method = "GET",
  data?: unknown,
): Promise<T> {
  const writing = method !== "GET";
  let response: Response;
  try {
    response = await fetch(base + "api/" + path, {
      method,
      headers:
        data !== undefined ? { "Content-Type": "application/json" } : undefined,
      body: data !== undefined ? JSON.stringify(data) : undefined,
      signal: AbortSignal.timeout(30000),
      cache: "no-store",
    });
  } catch {
    throw new ApiError(
      writing
        ? "连接中断，保存结果尚未确认。请先查看表格，避免重复提交。"
        : "连接失败，请检查网络后重试。",
      writing,
    );
  }
  let body;
  try {
    body = await response.json();
  } catch {
    throw new ApiError(
      writing ? "无法确认保存结果，请先查看表格。" : "服务器响应无效，请重试。",
      writing,
    );
  }
  if (!response.ok)
    throw new ApiError(
      body.error?.message || "操作失败",
      writing && response.status >= 500,
      body.error?.code,
    );
  return body as T;
}
export const api = {
  tables: (page = 1) => request<TableList>("tables" + query({ page })),
  createTable: (year: number, month: number) =>
    request<Table>("tables", "POST", { year, month }),
  updateTable: (t: Table, year: number, month: number) =>
    request<Table>("tables/" + t.id, "PUT", {
      year,
      month,
      revision: t.revision,
    }),
  deleteTable: (t: Table) =>
    request("tables/" + t.id + query({ revision: t.revision }), "DELETE"),
  report: (id: string, tag: string | null, page = 1, all = false) =>
    request<Report>(
      "tables/" + id + (all ? "/report" : "") + query({ tag, page }),
    ),
  saveRecord: (table: Table, id: number | null, input: RecordInput) =>
    request<FeeRecord>(
      "tables/" + table.id + "/records" + (id ? "/" + id : ""),
      id ? "PUT" : "POST",
      input,
    ),
  deleteRecord: (table: Table, id: number) =>
    request(
      "tables/" +
        table.id +
        "/records/" +
        id +
        query({ revision: table.revision }),
      "DELETE",
    ),
  suggestions: (kind: "locations" | "tags") => request<Suggestion[]>(kind),
  addSuggestion: (kind: "locations" | "tags", name: string) =>
    request(kind, "POST", { name }),
};
export async function download(report: Report, format: string, share = false) {
  const response = await fetch(
    base +
      "api/tables/" +
      report.table.id +
      "/export" +
      query({ format, tag: report.filterTag, revision: report.table.revision }),
    { signal: AbortSignal.timeout(60000), cache: "no-store" },
  );
  if (!response.ok) {
    const body = await response.json();
    throw new Error(body.error?.message || "导出失败");
  }
  const blob = await response.blob();
  const filename =
    "运费明细表_" +
    report.table.year +
    "年" +
    report.table.month +
    "月_" +
    report.table.id +
    (report.filterTag ? "_" + report.filterTag.replaceAll("/", "_") : "") +
    "." +
    format;
  const file = new File([blob], filename, { type: blob.type });
  if (share && navigator.canShare?.({ files: [file] })) {
    await navigator.share({ files: [file], title: "运费明细表" });
    return;
  }
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.append(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 60000);
}
