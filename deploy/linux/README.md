# Dalu Parts Linux x86_64 deployment package

This package targets Alibaba Cloud Linux 3.2104 LTS x86_64.

1. Read `docs/INSTALL.md`.
2. Verify/install MySQL Community 8.0.46 and create the application account.
3. Run `sudo ./scripts/install-layout.sh`.
4. Edit `/etc/dalu-parts/dlc.env`, run `scripts/migrate.sh`, configure the
   formal HTTPS site URL, and keep `RUN_MIGRATIONS=false`.
5. Render and enable the Nginx HTTPS configuration.

The package never contains production secrets, database data, uploaded media,
or source-tree user documents. Verify its files with `sha256sum -c SHA256SUMS`.
