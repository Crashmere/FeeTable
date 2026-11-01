# Operations and Maintenance

The public entry point uses HTTPS with a trusted IP certificate, preserving the existing `/feetable/` path. Public HTTP returns 308; API clients use HTTPS directly. Nginx overwrites `X-Forwarded-Proto`; write-origin validation trusts proxy headers only from loopback addresses and still rejects cross-site origins. See [shared HTTPS operations](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/https.md) (server copy: `/opt/server-context/references/https.md`) for certificates, renewal, rollback, and whole-server acceptance. This project's backend and deployment checks retain local HTTP. The proxy-inspection entry point at `127.0.0.1:80` is not publicly reachable. Public access uses ServerPortal's shared device authentication: enter the passphrase at `/portal/login` to authorize a device, then use the same-origin Secure/HttpOnly Cookie. Unauthorized APIs return 401. Local deployment checks and service-to-service calls remain available.

FeeTable's application target is `nas`, using an independent directory, runtime user, database, and deployment identity. `vps` exposes HTTPS port 8443; temporary `ali` access forwards to the same NAS instance. NAS also provides local-CA HTTPS for home devices. Read server-operations and the target's remote `/opt/AGENTS.md`; current-state records activation. Both entrances retain shared device authentication. Keep addresses outside this repository.

Request sizes: Nginx permits up to 2 MiB, and business JSON up to 64 KiB. The shared authentication layer's former default 1 MiB problem did not affect normal business input. See [shared authentication incorrectly blocks large requests](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/common-issues.md#shared-authentication-incorrectly-blocks-large-requests) for authentication-configuration maintenance and large-request checks.

FeeTable receives bounded JSON requests. Shared connection-level transfer tuning and cross-project checks are maintained in [Upload throughput and request limits](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/conventions.md#upload-throughput-and-request-limits), also at `/opt/server-context/references/conventions.md` on NAS.

## Runtime Configuration

| Item | Configuration |
| --- | --- |
| Program | `/opt/feetable/bin/feetable`, with embedded web pages and Chinese font |
| Environment file | `/opt/feetable/config/feetable.env` |
| Listener | `FEETABLE_ADDR=127.0.0.1:18081` |
| Database | `FEETABLE_DB=/opt/feetable/data/feetable.sqlite` |
| Legacy native backup directory | `/opt/feetable/backups`; no routine NAS snapshots |
| Users | `feetable` for runtime; `feetable-deploy` only for deployment |
| Persistent service | `feetable.service`, enabled at boot |
| Legacy native timer | `feetable-backup.timer`; disabled on NAS |
| Nginx | `/etc/nginx/app-locations/feetable.conf` links to the project's `config/nginx-location.conf` |
| Project documentation | `/opt/feetable/AGENTS.md` and `/opt/feetable/docs`; source recorded in `docs/SOURCE` |

Programs, scripts, and configuration are owned by root. `data`/`backups` are owned by `feetable`; directories use 0700 and the database uses 0600. The service has write permission only to `data`; the backup unit separately receives write permission to `data`/`backups`. Do not let the runtime user modify programs or configuration. Internal port 18081 is not exposed publicly.

## Initial Installation from Source

The local machine requires the Go toolchain specified by `go.mod`, Node 24+, and npm. The server runs one Linux amd64 program and does not need Go, Node, Docker, or a database service installed.

    make linux BASE_PATH=/feetable/
    sha256sum bin/feetable-linux-amd64

On macOS, use `shasum -a 256`. Export only the verified Linux artifact and deploy files from the reviewed commit into a separate temporary directory, then transfer that directory to the server. Do not transfer local `var`, credentials, or test databases.

First verify that port 18081, `/feetable/`, `/opt/feetable`, and the accounts do not conflict. From the uploaded directory, run:

    chmod 0755 feetable-linux-amd64
    ./feetable-linux-amd64 init --db initial.sqlite
    ./feetable-linux-amd64 check --db initial.sqlite
    bash deploy/install.sh ./feetable-linux-amd64 ./initial.sqlite
    ln -s /opt/feetable/config/nginx-location.conf /etc/nginx/app-locations/feetable.conf
    nginx -t
    systemctl reload nginx

This installation starts from an empty database. `init` creates only a nonexistent target; `serve` never creates a database automatically. `install.sh` refuses to overwrite an existing installation, and `restore` creates an independent consistent copy. server-operations manages the shared Nginx server; do not replace it from this project. If `nginx -t` fails, remove only the new location link from this installation and investigate; do not reload invalid configuration.

After initial installation, configure the independent deployment key under DEPLOYMENT and complete one local deployment. The deployment script writes `current-commit` only after health checks succeed; do not record a documentation commit as the running version.

## Acceptance and Troubleshooting

    systemctl is-active feetable nginx ledger
    systemctl is-enabled feetable
    curl -fsS http://127.0.0.1:18081/healthz
    curl -fsS http://127.0.0.1/feetable/healthz
    curl -fsS http://127.0.0.1/feetable/tables/1
    curl -fsS http://127.0.0.1/ledger/healthz
    journalctl -u feetable -n 50 --no-pager

Page acceptance also checks the bare-path 308 redirect, assets under `/feetable/assets/`, deep links including `/feetable/locations`, APIs, and rejection of cross-site writes. Production acceptance is read-only; use an isolated synthetic database for create/update/delete, network-loss, pagination, and export checks. Check a phone at 375×667 and a computer at desktop width. Also validate location management, all four sorting modes, and cross-page same-month merging against the synthetic database. Merge checks cover version conflicts, different months, the record limit, and rollback after a mid-transaction failure. Dialog acceptance covers internal scrolling, swipes at the top/bottom edges and on the backdrop, and restoration of background position after closing.

For startup failure, first inspect logs, database existence, ownership, integrity, and the port. Do not delete the database or run `init` to conceal the failure. HTTP 409 is a concurrent version conflict. After a save timeout or 5xx, verify the outcome before retrying. Export status 429 means another file is being generated. When a PNG exceeds the pixel limit, use PDF/XLSX or split it by tag.

## Consistent Backups

Use the agent-owned [backup and recovery runbook](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/backup-recovery.md), also at `/opt/server-context/references/backup-recovery.md` on NAS. Portal provides no backup or recovery operation.

Manual backup follows the [agent backup runbook](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/backup-recovery.md): complete NAS business data and recovery-critical private configuration and VPS recovery material are staged/compressed into one package on NAS, then copied/read-back verified externally. Required writers resume before compression; only successful task staging is removed. Native daily timers and automatic rotation remain disabled. The native `backup` command remains a consistent `VACUUM INTO` database tool for legacy/isolated checks; committed WAL data is included, but this is not the full-state job. Apply the shared [data-protection scope](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/retention.md#data-protection-scope) and tool-alignment requirement before capture.

    runuser -u feetable -- /opt/feetable/bin/feetable check --db /opt/feetable/backups/<backup-file>.sqlite

Keep real data/archives outside Git. The verified 2026-09-27 maintenance-computer archive and independent ALI Cloud Backup remain preserved in shared current-state; neither is a NAS schedule. Upgrades use the complete temporary local copy in [DEPLOYMENT.md](DEPLOYMENT.md), removed only after runtime and data preservation pass.

## Recovery and Drills

`restore` writes only a new, nonexistent file and never overwrites production directly. Run drills in an independent directory:

    runuser -u feetable -- /opt/feetable/bin/feetable restore --from /opt/feetable/backups/<backup-file>.sqlite --db /opt/feetable/data/<independent-drill-file>.sqlite
    runuser -u feetable -- /opt/feetable/bin/feetable check --db /opt/feetable/data/<independent-drill-file>.sqlite

Recovery remains agent-operated after confirming the recovery point and later data loss. For native packages, stop writes, preserve current full state, verify the selected source, and restore into a new file. Preserve the current main file and WAL/SHM together before atomic replacement; set owner `feetable` and mode 0600, restart, and inspect read-only. Keep NAS native scheduling disabled. Full cold archives are instead extracted and verified in an isolated directory before the agent restores the complete application subtree. Do not overwrite an active WAL database or repeatedly restore after unclear failures.

## Release, Configuration, and Documentation Synchronization

The current database version is 2. Unit prices are stored in cents and support two decimal places. An administrator must perform a one-time stopped-service upgrade for an existing v1 database; ordinary deployment and `serve` do not upgrade implicitly:

1. Build the candidate program from a verified commit and check its hash. Rehearse the candidate's `migrate`, `check`, and `serve` against an isolated complete data copy and verify startup and preservation of unit-price cent values.
2. Hold the shared maintenance lock and `/run/lock/feetable-deploy.lock`, confirm that any legacy backup job has ended, and stop `feetable`. Use the shared NAS helper to create and verify complete cold business-data and recovery-critical private-configuration protection under `/var/lib/server-upgrades/feetable/<release>/`; keep native scheduling disabled. Apply the shared [data-protection scope](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/retention.md#data-protection-scope) and tool-alignment requirement before capture.
3. As user `feetable`, run the candidate program's `migrate --db /opt/feetable/data/feetable.sqlite`, then run `check`. The upgrade completes in one transaction; failure rolls back the transaction and leaves no partial new table. Existing unit-price cent values are preserved directly, without multiplication or division.
4. Install the candidate program, start it, verify local/proxied health and affected data, and record `current-commit`. Keep NAS native scheduling disabled; accept and remove only the validated upgrade copy before returning to normal local releases.

A v1 program does not support a v2 database and truncates fractional unit prices, so it cannot be rolled back directly after the upgrade. Repair with a v2-compatible program. If data truly must be rolled back, first use the recovery procedure to confirm the recovery point and data loss. Retain the pre-upgrade backup and do not automatically overwrite a database that has accepted new writes. The new program's `check`/`backup`/`restore` can read v1 backups; a restored v1 database must be explicitly migrated to v2 before `serve`.

See [DEPLOYMENT.md](DEPLOYMENT.md) for routine program deployment and automatic rollback. It replaces only the binary. An administrator installs configuration, units, and deployment scripts from a pushed commit; documentation uses the synchronization command below. Compare live and source configuration before updates. After installation, verify syntax and permissions, then run `daemon-reload` or `nginx -t` and reload as needed. When changing the shared entry point, check Ledger as well.

After pushing managed documentation, run `SYNC_HOST=nas ~/agent-config/skills/server-operations/scripts/sync-docs.sh FeeTable`. It checks drift and installs verified documentation with `docs/SOURCE`. Shared inventory and gateway conventions belong to agent-config and must synchronize to their respective hosts.

## ServerPortal Integration Materials

`deploy/portal.json` is the maintained source for this application's resource description. Local deployment validates it with the server-operations validator, then stores the same declaration and binary in the same local version directory. It runs `portal-check` before deployment and `portal` afterward, using the existing restricted SSH identity to install `/opt/feetable/config/portal.json` and verify the SHA-256 actually loaded by the collector. `config/portal-source.json` records the declaration's source commit; it and the program's `current-commit` identify versions of different materials.

The portal discovers this application through the controlled link `/opt/serverportal/registry.d/feetable.json`. It automatically reloads a successfully updated declaration without a restart. The first normal local deployment also creates the link, so there is no separate central application list to edit. An ordinary deployment may update this application's declaration; an administrator still installs other units, env files, Nginx configuration, and deployment scripts.

Source-code or database-behavior changes cannot use this option as a substitute for program deployment.

When data roots, media, backup formats, units, ports, or access paths change, update the declaration and corresponding documentation in the same commit, update the shared inventory, and check resource coverage. The portal reads files, media, database tables, and systemd state automatically; a maintenance agent must explicitly update directory purposes, API descriptions, and permission boundaries. See [portal maintenance](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/portal.md) for shared protocols, failure handling, and new-application onboarding.

Portal protects the NAS, VPS, and compatibility ALI HTTPS entrances; loopback release checks remain private and public unauthorized APIs return 401. Device grants persist until revocation; preserve shared cookie-renewal headers. Agent backups include complete NAS state in the combined NAS/VPS package, not only the native database snapshot. Recovery remains agent-operated after confirming the recovery point.

## Phone Home-Screen Icon

The current 180×180 icon is built from `web/src/assets/apple-touch-icon.png` into content-hashed `/feetable/assets/apple-touch-icon-<hash>.png`. Nginx rules match hashes for the three brand-icon names rather than pinning a one-build filename, and do not expose the whole assets directory. Pages, APIs, and user media remain protected by device authentication. See the [shared troubleshooting record](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/common-issues.md#missing-iphone-home-screen-icons-after-shared-authentication) for the common cause, deployment status, and phone acceptance.

## Current Release Entry Point

This project is for personal use: validate the current change locally before release; there is no full-regression gate, and permanent tests are neither added nor retained by default. For UI changes, check the computer and phone scenarios that are actually used. Validate data migrations, bulk writes/deletions, and backup restoration specifically against an isolated copy first.

See [Local Deployment and Rollback](DEPLOYMENT.md) for the complete procedure. GitHub stores only source code. Local `make release` builds, `make deploy` updates production, and documentation is synchronized separately.

## Deployment-Material Cleanup

NAS has no daily backup/release pruning. Successful upgrade acceptance removes only its private `/var/lib/server-upgrades/feetable/<release>/` and candidate work copy; failed or uncertain copies remain for the agent. Existing ALI and maintenance-computer backups are untouched. [Legacy retention](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/retention.md) is not enabled for NAS upgrades.
