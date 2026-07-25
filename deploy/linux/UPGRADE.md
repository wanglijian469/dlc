# Upgrade and rollback

## Before every upgrade

1. Read the release notes and confirm the target package version.
2. Back up the database and uploaded media:

   ```bash
   sudo ./scripts/backup.sh
   sudo sha256sum -c /var/backups/dalu-parts/<timestamp>/SHA256SUMS
   ```

3. Copy the previous `/opt/dalu-parts/server`, `/opt/dalu-parts/initdb`, and
   `/opt/dalu-parts/public` to a versioned, root-only rollback directory.
4. Test restoring the backup on a non-production instance when the release
   contains a schema change.

## Upgrade

```bash
sudo systemctl stop dalu-parts
sudo ./scripts/install-layout.sh
sudo ./scripts/migrate.sh
sudo systemctl start dalu-parts
sudo ./scripts/health-check.sh
```

`install-layout.sh` preserves `/etc/dalu-parts/dlc.env` and uploaded media.
Production must always retain `RUN_MIGRATIONS=false`; `initdb` is the only
approved migration entry point.

## Rollback

If migration fails, do not start the new server. Restore the database and
media backup, restore the previous binaries/public directory, and then start
the previous version. A binary rollback alone is unsafe after a schema change.

```bash
sudo systemctl stop dalu-parts
# Restore the reviewed database and media backup here.
# Restore the previous /opt/dalu-parts files here.
sudo systemctl start dalu-parts
sudo ./scripts/health-check.sh
```

For a MySQL server upgrade, follow MySQL's supported upgrade path and run its
upgrade checks before changing packages. In-place downgrade is not supported.
