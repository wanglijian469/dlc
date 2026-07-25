# Alibaba Cloud Linux 3.2104 LTS x86_64 deployment

## 1. Host and network prerequisites

- Use Alibaba Cloud Linux 3.2104 LTS x86_64 with current security updates.
- Bind the ECS security group to TCP 22 from administrator addresses and TCP
  80/443 from the public Internet. Do not expose TCP 3306 or 8080.
- Point the production domain to the ECS public IP and prepare its full-chain
  certificate and private key.
- Keep SELinux enforcing. Nginx connects only to the loopback Gin listener.

```bash
sudo dnf update -y
sudo dnf install -y nginx curl tar gzip policycoreutils-python-utils
```

## 2. Verify or install MySQL Community 8.0.46

Alibaba Cloud Linux 3 is compatible with the CentOS/RHEL 8 ecosystem, so the
installer uses Oracle's EL8 Yum repository and explicitly enables
`mysql80-community`. Both the client and server are pinned to release
`8.0.46-1.el8`, matching the target host's
`mysql-community-client-8.0.46-1.el8.x86_64`.

Oracle's current repository setup RPM enables the 8.4 LTS channel by default;
the installer disables MySQL server channels first and then enables only
`mysql80-community`, as required by the
[official Yum repository procedure](https://dev.mysql.com/doc/mysql-repo-excerpt/8.0/en/linux-installation-yum-repo.html).

```bash
sudo ./scripts/install-mysql-8.0.46.sh
sudo grep 'temporary password' /var/log/mysqld.log
sudo mysql_secure_installation
rpm -q mysql-community-client mysql-community-server
mysql --version
```

The RPM query must report version `8.0.46-1.el8` for both packages. If the
matching client/server is already installed, the script reuses it. If a
different MySQL or MariaDB server is installed, the script stops without
replacing it. The application remains compatible with MySQL 5.7+, but a server
upgrade requires a separate, tested backup and restore plan.

Create a local-only application account. Its grants include only the schema
DDL/DML needed by the versioned `initdb` migration; it has no global grants.

```bash
sudo DB_NAME=dl_nongji_parts DB_USER=dlc_app \
  ./scripts/create-mysql-user.sh
```

Use the same password in `/etc/dalu-parts/dlc.env`.

## 3. Install the application layout

```bash
sudo ./scripts/install-layout.sh
sudoedit /etc/dalu-parts/dlc.env
sudo grep -n CHANGE_ME /etc/dalu-parts/dlc.env
```

The standard paths are:

- binaries and public assets: `/opt/dalu-parts`
- root-managed environment: `/etc/dalu-parts/dlc.env`
- uploaded media: `/var/lib/dalu-parts/media_storage`
- runtime identity: `dlc`

Replace every `CHANGE_ME`, set the formal HTTPS origin, and generate the
signing secret with `openssl rand -hex 32`. Keep
`HTTP_ADDR=127.0.0.1:8080`, `TRUSTED_PROXY_CIDRS=127.0.0.1,::1`, and
`RUN_MIGRATIONS=false`. The environment-file syntax does not support arbitrary
shell expressions; use literal values without newlines.

## 4. Initialize and migrate the database

Run the migration once as the service identity, then keep automatic
migrations disabled:

```bash
sudo ./scripts/migrate.sh
```

The application refuses to start when the schema is behind. Before production
startup, write the formal HTTPS origin into `site.meta.siteUrl`:

```bash
sudo ./scripts/configure-site-url.sh https://www.example.cn
```

The production server rejects HTTP, empty, and example URLs.

## 5. Configure Nginx HTTPS

Copy the certificate to root-owned paths, then render the template:

```bash
sudo ./scripts/render-nginx-config.sh \
  www.example.cn \
  /etc/pki/tls/certs/www.example.cn.fullchain.pem \
  /etc/pki/tls/private/www.example.cn.key
sudo nginx -t
sudo systemctl enable --now nginx
```

If SELinux denies the loopback proxy connection:

```bash
sudo setsebool -P httpd_can_network_connect 1
```

Do not serve `public/` directly from Nginx: Gin provides dynamic HTML, API,
sitemap, and media authorization from one origin.

## 6. Start and verify

```bash
sudo systemctl enable --now dalu-parts
sudo ./scripts/health-check.sh
sudo PUBLIC_URL=https://www.example.cn ./scripts/health-check.sh
sudo journalctl -u dalu-parts -n 100 --no-pager
```

Verify the HTTPS home page, `/api/health`, admin login, media upload, restart
persistence, HTTP-to-HTTPS redirect, and that ports 3306/8080 are not public.
The systemd sandbox makes the service filesystem read-only except for the
media directory.

## 7. Firewall

If firewalld is enabled in addition to the ECS security group:

```bash
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

Keep MySQL listening locally. Do not add the MySQL service to firewalld.
