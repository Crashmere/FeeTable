# FeeTable Maintenance Entry Point

Write agent instructions, maintenance documentation, technical plans, handoffs, and similar agent-authored operational documents in English by default unless the user explicitly requests another language. User-facing product output follows the language requested by the user.

This project is for personal use: validate the current change locally before release; there is no full-regression gate, and permanent tests are neither added nor retained by default. For UI changes, check the computer and phone scenarios that are actually used. Validate data migrations, bulk writes/deletions, and backup restoration specifically against an isolated copy first.

GitHub backs up only source code and configuration; pushing does not trigger tests or deployment. See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for local entry points and rollback; shared procedures are maintained by server-operations.

Read [docs/README.md](docs/README.md) first, then read the architecture, API, operations, and release documentation as needed.

- Keep the single-service, single-SQLite, direct-SQL implementation. Use fixed-point representations for money and quantities, and decimal strings for API numeric values.
- The business scope is monthly tables, transport records, locations, one tag per record, and PNG/PDF/XLSX exports. Networked reads and writes are available after shared device authentication, and the server is the authoritative data source.
- Writes, amounts, dates, filtering, and exports must remain consistent. Use a synthetic database for tests and read-only acceptance in production.
- Before deployment or server operations, read the server-operations skill and explicitly run `ssh ali 'cat /opt/AGENTS.md'`; shared server documentation is under `/opt/server-context`.
- FeeTable maintains only its own deploy/location configuration, unit, data, and deployment identity. Shared-configuration changes require checking every application.
- Source configuration and `docs` are the maintenance sources. Replace outdated descriptions after changes, then, after pushing, run agent-config's `skills/server-operations/scripts/sync-docs.sh FeeTable` to synchronize the server copies.
- Use only the authorization table in server-operations `SKILL.md` to decide what to finish and report afterward versus what to confirm first. Documentation maintenance and synchronization do not require advance confirmation.
- Data, backups, private keys, and the server's public address must not enter Git.

## Portal Resource Synchronization

- When changing website icons or authentication, also check the iPhone `apple-touch-icon`, its post-build path, and anonymous GET/HEAD access. Only explicit brand icons and a public manifest may be exceptions; pages, APIs, and user media must remain authenticated. See server-operations common issues for shared troubleshooting and acceptance; current-state is authoritative for actual enablement status.

- This project's `deploy/portal.json` is the maintained source for the ServerPortal resource declaration. It records directory purposes, the database, runtime user, port, units, access paths, APIs, and backup types. When adding, migrating, or deleting a data root, interface, or runtime material, update the declaration, the corresponding docs, and the shared application inventory together.
- The declaration is deployed to `/opt/feetable/config/portal.json` and managed by root. Local deployment uses the shared server-operations validator: it performs a preflight before deployment, then automatically synchronizes over restricted SSH and verifies the hash loaded by the portal. `registry.d/feetable.json` registers the link automatically; updates do not require restarting the portal.
- Shared Nginx and the portal provide unified authentication; this project does not maintain a separate device allowlist. Local calls and deployment health checks remain available under the shared conventions. Device authentication is enabled in production; synchronize server-operations current-state after changes.
- The portal's read-only display does not replace this project's native consistent backups. When backup formats or media lifecycles change, validate the affected backup and recovery contracts against an isolated copy. Real business data, credentials, and backups still must not enter Git.

Local deployment automatically preflights and synchronizes the `deploy/portal.json` from the same commit. To update only the declaration, run `make portal`; this preserves the business-program version.

Web pages block ordinary Tab / Shift+Tab control traversal and remove focus highlights by default; retain only exceptions explicitly requested by the user. Maintain and verify this behavior under the [server-operations web keyboard and focus conventions](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/conventions.md#web-keyboard-and-focus).
