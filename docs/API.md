# HTTP API

The following are internal paths of the Go service; `{id}` and `{record}` are path parameters. After deployment, external paths all receive the `/feetable` prefix, which Nginx removes before forwarding. Public APIs are protected by ServerPortal's shared device authentication and return 401 when unauthorized; local loopback calls remain available. JSON responses disable caching. Browser write requests additionally verify that Origin has the same scheme and Host as the request.

| Method and path | Purpose |
| --- | --- |
| GET /healthz | Check the database connection |
| GET /api/tables | List monthly tables, record counts, and amounts, 30 per page; query parameters: `page` (default 1), `sort` (default `updated_desc`), with optional `year`/`month` filtering |
| POST /api/tables | Create a table; `year` and `month` are required |
| GET /api/tables/{id} | Return the table, records, all tags, and filtered count/total, 50 records per page; query parameters: `page` (default 1), `tag` (optional) |
| PUT /api/tables/{id} | Change `year`/`month`; `revision` must match |
| DELETE /api/tables/{id} | Delete the table and its records; query parameter `revision` must match the current table version |
| POST /api/tables/{id}/merge | Merge source tables for the same year and month into the retained table identified by `id` |
| POST /api/tables/{id}/records | Add a record; the body includes the current table `revision` |
| PUT /api/tables/{id}/records/{record} | Fully update a record; the body includes the current table `revision` |
| DELETE /api/tables/{id}/records/{record} | Delete a record; query parameter `revision` must match the current table version |
| GET /api/locations | List common locations, sorted by cumulative usage count |
| GET /api/tags | List common tags, sorted by cumulative usage count |
| POST /api/locations | Add a common location; body is `{name}`; do not create a duplicate if it already exists |
| POST /api/tags | Add a common tag; body is `{name}`; do not create a duplicate if it already exists |
| PUT /api/locations/{id} | Rename a location; body contains `name` and `previousName` |
| DELETE /api/locations/{id} | Delete a common location; query parameter `name` is the original name and is used to detect concurrent modification |
| GET /api/tables/{id}/report | Return complete export data and `revision`; optional query parameter `tag`; not paginated |
| GET /api/tables/{id}/export | Download; query parameter `format` is `png`, `pdf`, or `xlsx`; optional `revision` and `tag`; `inline=1` permits inline opening |

This table covers the current business and health interfaces; keep `deploy/portal.json.apis` consistent with it. Register each method and path separately. Query parameters are documented in the purpose column and the contracts below. See the [shared troubleshooting guidance](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/common-issues.md#portal-api-list-does-not-match-the-service) for checking and synchronization procedures.

## Monthly Table Sorting, Merging, and Location Management

`sort` may be `updated_desc` (default), `updated_asc`, `month_desc`, or `month_asc`, representing descending/ascending modification time or table year and month. `year` and `month` must be supplied together and be valid. The filtered `total` must agree with pagination; unsupported sorting returns 400.

A merge request contains `revision` and a `sources` array; each source entry contains `id` and `revision`. The path `id` identifies the retained table, and the top-level `revision` is its version. `sources` contains 1–99 distinct source tables and their versions and cannot contain the retained table. Success returns the retained table's new `revision`, record count, and total amount; source tables subsequently return 404. All years and months must match, and the combined records must not exceed 5,000. Any version conflict returns 409, a missing table returns 404, and every operation rolls back atomically. All records, including duplicate content, are moved in full.

Renaming or deleting a location preserves snapshots in existing records. Renaming preserves the usage count and returns 400 for a duplicate name. `previousName`/`name` must match the server's current name or the request returns 409; a missing location returns 404. If a deleted name is later used to save a record, it is added to the location library again.

## Record Input

Every field is required. `unitPrice` and `tag` may be null; other fields may not. The date must belong to the table's year and month, and names must be 1–80 characters. Numeric values use ordinary decimal strings; exponent notation, NaN, Infinity, and grouping commas are not accepted.

```json
{
  "day": 3,
  "location1": "Test origin",
  "location2": "Test destination",
  "quantity": "0.145",
  "unitPrice": "1",
  "amount": "0.15",
  "tag": "Test tag",
  "revision": 1
}
```

When `unitPrice` is non-null, the amount is calculated from `quantity`; `amount` is required only as part of the complete form and does not determine the saved amount. When `unitPrice=null`, `amount` is used. Amount and quantity responses use fixed two- and three-decimal strings, respectively; unit-price responses use fixed two-decimal strings. Unit-price input allows a nonnegative integer or at most two decimal places (for example, `"2.88"`); quantity and fixed amounts allow negative values and zero. `revision` is the containing table's version, not the record's own version.

A successful write returns the saved entity or `{"ok":true}`. The client then rereads the table to obtain the new `revision`. `tag` is encoded with `URLSearchParams`; omitting it means all records, while supplying it performs an exact match. Slashes and similar characters in tags are not route paths.

## Errors

Application errors have the shape `{"error":{"code":"VALIDATION","message":"description"}}`. Status 400 means a field/date/format error; 403, a mismatched write-request origin; 404, not found; 409, a version conflict; 422, an export-size limit; 429, renderer busy; and 500, an internal failure. The shared public entry point returns 401 for device-authentication failure. A network failure or 500 during a write must not be retried automatically; first check whether the server saved it.

Each JSON input is limited to 64 KiB. Export responses include an explicit MIME type, `Content-Disposition`, and `no-store`. A submitted nonzero `revision` that differs from the current version returns 409; omitting it or supplying 0 skips that check. The client should submit the `revision` from the preview unchanged. An empty filtered set cannot be exported, and the page disables the corresponding download buttons.
