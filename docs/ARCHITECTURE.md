# Architecture and Code Guide

```text
Browser → Nginx /feetable/ → Go HTTP → SQLite
                              ├─ Embedded Vue page
                              └─ PNG / PDF / XLSX
```

The Go entry point is in `cmd/feetable`; `internal/httpapi` decodes requests, checks origins, and emits JSON/files; `internal/feetable` implements domain rules and SQL; and `internal/export` handles exports. `web` uses Vue 3, Vue Router, TypeScript, and Vite.

## Persistence and Data

The four business tables are `fee_tables`, `fee_records`, `locations`, and `tags`. Each monthly table has a `revision`; adding, changing, or deleting a record, or changing the year/month, increments that version. The client submits the version it read, and a conflict returns 409. Every multi-step write uses a transaction. SQLite enables WAL, foreign keys, FULL synchronous mode, and a single connection.

Location and tag text stored in a record is a snapshot taken at that time. Suggestion libraries are globally unique name lists sorted by cumulative count of use in new records. Adding a record inserts missing terms and increments counts; editing inserts missing terms but does not increment counts; deleting does not reduce cumulative counts. Names are trimmed, empty tags become null, and control characters are rejected.

Location-management rename and delete operations affect only `locations`, preserving record snapshots and table `revision`. Renaming preserves cumulative usage count and rejects duplicate names. The `previousName`/`name` read by the client detects concurrent renames, which return 409. There is no new database column or schema migration.

`quantity_milli`, `unit_price_cents`, and `amount_cents` are integers. Decimal input strings are parsed directly. Multiplication uses a large-integer intermediate value, and quantity-based amounts are calculated by the backend rather than using the client-supplied amount. Totals sum each saved amount, and responses use fixed-decimal strings. The browser uses BigInt only for previews and formatting.

The monthly-table list supports `updated_desc` (default), `updated_asc`, `month_desc`, and `month_asc`. Modification-time sorting breaks ties by ID in the same direction; year/month sorting breaks ties by `updated_at` and ID in descending order. Sorting occurs on the server before pagination, and year/month filters can narrow merge candidates. Records are sorted by `day`, then `id`, ascending. Monthly tables use 30 rows per page and records use 50; totals and counts cover the complete filtered set. Headers, records, and summaries are read in one consistent read transaction.

A merge request identifies the retained table and 1–99 source tables and includes each table's `revision`. A single transaction verifies existence, versions, identical year/month, unique IDs, and no more than 5,000 combined records; it then changes each source record's `table_id`, deletes the now-empty source tables, and increments the retained table's `revision`/`updated_at`. Record IDs, dates, text, tags, fixed-point values, and suggestion-library counts remain unchanged; records are neither deduplicated nor have amounts recalculated. Any failure rolls back all writes. Revision checks reject stale previews and forms.

The database `application_id` is 1179931714 and `user_version` is 2. Normal startup only opens an existing database and verifies identity, integrity, and the current version. `init` requires a nonexistent target. After stopping the service and taking a backup, `migrate` must be run explicitly; in one transaction it upgrades the v1 record-table constraint to support fractional-unit prices while preserving data, IDs, revisions, autoincrement sequences, and indexes. `check`/`backup`/`restore` support v1 and v2; `serve` requires v2. See OPERATIONS for exact steps and rollback boundaries.

## UI and Save Feedback

The site-wide header uses `sticky; top: 0` with an opaque background, keeping the brand and menu at the top while the body scrolls. The root scrolling area reserves 104 px of top positioning space so focused controls and anchors do not end up under the header. Mobile viewports use the browser's default safe-area layout and do not enable `viewport-fit=cover`; see the [top safe-area explanation](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/common-issues.md#top-of-an-iphone-home-screen-page-appears-blurry) for the reason and real-device verification boundary.

The website icon combines a transport truck with an itemized bill and retains the dark green, light green, and cream palette. Its source is `web/src/assets/icon.svg`, shared by the header and SVG favicon. `icon-32.png` is a 32×32 browser-compatibility version of the same image, and `apple-touch-icon.png` is the 180×180 home-screen version filling a dark-green background. When changing the source image, rerasterize both PNGs. Vite emits the icons as independent assets with content hashes, and links are built with `BASE_PATH` to avoid stale icon-cache hits after updates. The header image has an empty `alt` so the adjacent site name is not read twice.

Home is the monthly-table list and provides sorting and same-month merging. The location-management entry point is in the three-line menu at the upper right of the site-wide header, and `LocationsPage` manages common locations. The menu closes on outside click, Escape, or page navigation. `TablePage` manages records and filtering; `ExportPage` displays the full report. Phones use record cards and vertical forms, while computers use a table; the wide export table is not reused as the editing interface. Merging first selects source tables, then confirms the record count, amount, and removal of source tables.

The sorting controls share a row with the exact UI label “全部表格 / 张数”. A clock represents modification time, a calendar represents table year/month, and each group includes up/down arrows. Selecting the other group defaults to newest first; selecting the active group toggles ascending/descending order. The active group and direction are highlighted. Buttons use `title` and `aria-label` to describe their meaning and next direction, while `aria-pressed` identifies the current mode.

The site-wide `web/src/keyboard.ts` intercepts ordinary Tab / Shift+Tab during capture, so focus does not move among page, form, location/tag-suggestion, and dialog controls. Ctrl/Command/Alt combinations and IME composition remain available. Controls remove browser focus outlines, while mouse clicks, text input, Escape-to-close, and selected sorting states continue to work, following the [shared web keyboard and focus conventions](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/conventions.md#web-keyboard-and-focus).

Every Modal uses Teleport to render under `body`. Before opening, it fixes the background body and records the scroll position; the dialog uses `overscroll-behavior` to stop edge scrolling from propagating. Unmounting restores the original styles and position. The lock is reentrant so overlapping lifecycles do not release it early. The monthly-table list and record list retain their current content while refreshing, preventing a dialog close from briefly shortening the page and losing its scroll position.

UI copy contains only business information, fields, actions, and necessary status feedback; do not add marketing copy, decorative subtitles, footer slogans, or repeated explanations. Year and month are incorporated into the date field, the input placeholder explains an empty unit price, and validation errors and delete confirmations remain explicit.

While saving, duplicate submission, dialog closing, and in-site navigation are blocked; the form is cleared only after server success. Failures preserve input. A network interruption or 5xx makes the outcome unknown, disables retries from the current form, and requires verification first. Read requests offer retries. There is no offline write queue or business-data cache.

## Exports

A Report contains the table, the complete filtered records, and the total calculated from integers. A download may include `revision`; if it has changed, the user must refresh the preview. File rendering occurs after the SQL transaction ends. One concurrency slot controls generation, returning 429 when busy.

PNG and PDF share a layout: seven columns, merged title and header cells, fixed widths, and wrapped long text; the embedded font is LXGW Neo XiHei. PNG wraps text at logical dimensions and then draws fonts and grid directly at 3× size (2,832 pixels wide). When actual pixels would exceed 80,000,000, it uses 2× size instead (1,888 pixels wide), and rejects generation if the image still exceeds the limit. It uses Go `image.Gray`, one byte per pixel, while retaining text antialiasing; the bitmap uses at most 80 MB, and logical long-table capacity does not change. PDF uses `gopdf` with the embedded font and Unicode mappings. XLSX uses Excelize and shared amount and numeric formats. Excel text fields are written as strings.

Export files are generated in memory. Responses use `no-store`, and filenames incorporate table ID, year/month, and tag. Files are not persisted to a public directory. The browser can save the download elsewhere; native sharing is shown only when supported.

## Build and Subpath

`BASE_PATH` determines Vite assets, the router history base, and the API prefix. The trailing slash in Nginx `proxy_pass` removes `/feetable/`, so Go receives internal paths such as `/api` and `/assets`. Unknown page paths return `index.html`; unknown static files return 404.

`go:embed` packages the production web page and Chinese font into the binary. The normal server does not run Node, a Go compiler, or an external database.
