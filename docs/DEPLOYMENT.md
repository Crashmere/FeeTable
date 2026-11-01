# Local Deployment and Rollback

This project is for personal use: validate the current change locally before release; there is no full-regression gate, and permanent tests are neither added nor retained by default. For UI changes, check the computer and phone scenarios that are actually used. Validate data migrations, bulk writes/deletions, and backup restoration specifically against an isolated copy first.

GitHub is only a backup of source code and configuration, and Actions is disabled. Committing and pushing do not update production. See [server-operations local deployment](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/release.md) for the shared implementation and environment preparation; this project's parameters come from `deploy/release.json`.

## Routine Use

`make deploy` targets `nas` and its dedicated deployment key. `vps` provides the public HTTPS entrance on port 8443; temporary `ali` access forwards to the same NAS application. Public-check overrides and host-specific keys are maintained in the shared release guide. Explicit legacy ALI deployment must not create a second writer after migration.

```sh
# Exercise the current feature locally, then commit and push the source.
git add <files-changed-in-this-task>
git commit -m "Describe the change"
git push origin main
make release   # Build from the committed snapshot; do not run the full test suite
make deploy    # Reuse the artifact from the same commit and deploy it over SSH
```

Local materials are under `.local/releases/<full-commit>/` and contain `program`, `manifest.json`, and `portal.json` from the same commit. Deployment verifies hashes and confirms that the commit is backed up to `origin/main`; it reuses the artifact when source is unchanged. Untracked files do not enter the build. `make releases` lists local versions. These materials are not committed to Git.

Use `make portal` to update only portal information. It preflights and synchronizes the declaration from the same commit without building or restarting the business program. Failure to synchronize the declaration does not automatically revert a program update that already succeeded; fix the issue and retry the declaration separately.

## Deployment Boundaries

Use the restricted `feetable-deploy` identity; its private key exists only on the maintenance computer. For initial account setup, use `deploy/setup-deploy.sh <local-deployment-public-key>`; do not initialize an existing account again. The forced command accepts only `deploy`, `portal-check`, and `portal`, and sudo permits only this application's fixed deployment script.

Temporary upgrade backups follow the shared [database-only policy](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/retention.md#temporary-upgrade-backups): skip them unless the change can delete or overwrite existing database data, and never include non-database files. The current fixed NAS release flow still uses the broader helper described in the [implementation status](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/retention.md#tool-alignment).

Switching begins only after upload and verification succeed. On `nas`, the current fixed release script cold-copies complete business data and recovery-critical private configuration to root-private `/var/lib/server-upgrades/feetable/<release>/`. Check the candidate on an independent copy and verify runtime plus preserved SQLite rows and files. Successful acceptance immediately deletes only that upgrade's local copy; failure or uncertainty retains it. `previous` remains the program rollback artifact. Normal deployments do not transfer production data. Administrators install units, Nginx, env, authorization, and scripts from source; documentation uses `SYNC_HOST=nas` with the shared `sync-docs.sh`.

The current database is schema 2. A schema 1 program cannot be rolled back by replacing only the binary. Database upgrades and restoration still follow OPERATIONS independently.

## Rollback and Troubleshooting

```sh
make releases
make rollback COMMIT=<explicitly-selected-full-old-commit>
```

Rollback reuses a retained, verified local artifact and uses the current fixed script and health checks; temporary backup selection follows the policy above. First confirm that the old program is compatible with current data. Program rollback does not restore the database, photos, or device state. To restore data, first choose a recovery point and preserve the current state, then follow OPERATIONS/RESTORE.

NAS retains the program, `previous`, `metadata`, `result`, `recovery`, and upgrade-acceptance evidence under `/opt/feetable/releases/`; `current-commit` records the actual running source. `.local/releases/<commit>/last-deployment.json` records the local result. After a failed switch, inspect the result, journal, and retained private snapshot. NAS does not automatically prune routine backups or uncertain upgrade snapshots. [Legacy automatic retention](https://github.com/Crashmere/agent-config/blob/main/skills/server-operations/references/retention.md) applies to `ali`; local builds remain outside it.
