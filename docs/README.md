# FeeTable Documentation Entry Point

| Topic | Document |
| --- | --- |
| Data, requests, amounts, and exports | [ARCHITECTURE.md](ARCHITECTURE.md) |
| API and error contracts | [API.md](API.md) |
| Installation, operation, backup, and recovery | [OPERATIONS.md](OPERATIONS.md) |
| Local deployment and rollback | [DEPLOYMENT.md](DEPLOYMENT.md) |

## Confirmed Business Boundaries

The service uses an independent database and permits viewing, modification, and export after shared device authentication. One user can access it from multiple devices. The server stores the authoritative data, and the web page requires a network connection. There is no historical import, browser database, offline synchronization, account authorization model, or recycle bin.

Tables are created by year and month, and multiple tables may have the same month. Each record stores only a day; its year and month belong to the table. Changing the year/month changes dates for the whole table, and the backend rejects every invalid date. Deleting a table cascades to its records and requires page confirmation.

The home page supports ascending or descending sorting by modification time or table year/month, defaulting to the most recently modified first. A merge entry point on each table can select other tables with the same year/month across pages, for 2–100 total tables per operation. After confirmation, records move into the current retained table and source tables are removed. All records, tags, and amounts are preserved in full without automatic deduplication. The merged result may contain at most 5,000 records and cannot be split back directly.

The location-management page in the three-line menu at the upper right of the header supports unified searching, addition, renaming, and deletion of common locations. Renaming and deletion affect only the location library; existing records retain the name used at the time. Saving a later record adds any used name back to the library. On the home page, the exact UI label “全部表格 / 张数” shares a row with clock and calendar icons that sort by modification time and table year/month, respectively. Selecting the current icon toggles ascending/descending order, with the active mode and arrow highlighted.

Quantity supports at most three decimal places. Unit price is nonnegative with at most two decimal places, and zero differs from blank. Blank means a fixed charge, whose amount supports at most two decimal places. Quantity and fixed amount preserve negative and zero values. Quantity-priced amounts round the absolute value to cents and then restore the sign. Each record has at most one optional tag.

The seven export columns are month, day, location 1, location 2, quantity, unit price, and amount. Tags determine only the filter and do not add a column. PNG/PDF preserve the complete long table; XLSX stores numeric cells and merged headers. Exports use the complete filtered set and are not limited by the current browser page.

Each table supports at most 5,000 records and each name at most 80 characters. The absolute value of a quantity in thousandths or an amount in cents must not exceed 10¹²; unit price must not exceed 10¹⁰ yuan. PNG defaults to 3× resolution and uses 2× for long tables, with an actual image limit of 80 million pixels. The service renders only one export file at a time.

The shared server maintenance entry point is [server-operations](https://github.com/Crashmere/agent-config/tree/main/skills/server-operations), with a live copy at `/opt/server-context`. OPERATIONS contains exact deployment facts; do not write plans as if they are deployed.

Follow [local deployment instructions](DEPLOYMENT.md) for routine deployment and validation. GitHub is only a source-code backup.
