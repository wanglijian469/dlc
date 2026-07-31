#!/usr/bin/env bash
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "run as root: sudo ./scripts/install-layout.sh" >&2
  exit 1
fi

PACKAGE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
test -x "${PACKAGE_ROOT}/server"
test -x "${PACKAGE_ROOT}/initdb"
test -f "${PACKAGE_ROOT}/public/index.html"

getent group dlc >/dev/null || groupadd --system dlc
id dlc >/dev/null 2>&1 || useradd \
  --system --gid dlc --home-dir /var/lib/dalu-parts \
  --shell /sbin/nologin dlc

install -d -o root -g root -m 0755 /opt/dalu-parts
install -d -o root -g dlc -m 0750 /etc/dalu-parts
install -d -o dlc -g dlc -m 0750 /var/lib/dalu-parts/media_storage
install -d -o dlc -g dlc -m 0750 /var/lib/dalu-parts/static-pages
install -d -o root -g root -m 0755 /opt/dalu-parts/public

install -o root -g root -m 0755 "${PACKAGE_ROOT}/server" /opt/dalu-parts/server
install -o root -g root -m 0755 "${PACKAGE_ROOT}/initdb" /opt/dalu-parts/initdb
rm -rf /opt/dalu-parts/public.new
install -d -o root -g root -m 0755 /opt/dalu-parts/public.new
cp -a "${PACKAGE_ROOT}/public/." /opt/dalu-parts/public.new/
chown -R root:root /opt/dalu-parts/public.new
find /opt/dalu-parts/public.new -type d -exec chmod 0755 {} +
find /opt/dalu-parts/public.new -type f -exec chmod 0644 {} +
rm -rf /opt/dalu-parts/public
mv /opt/dalu-parts/public.new /opt/dalu-parts/public

if [[ ! -e /etc/dalu-parts/dlc.env ]]; then
  install -o root -g dlc -m 0640 \
    "${PACKAGE_ROOT}/config/dlc.env.example" /etc/dalu-parts/dlc.env
  echo "created /etc/dalu-parts/dlc.env; replace CHANGE_ME values before starting"
else
  echo "kept existing /etc/dalu-parts/dlc.env"
fi

install -o root -g root -m 0644 \
  "${PACKAGE_ROOT}/systemd/dalu-parts.service" \
  /etc/systemd/system/dalu-parts.service
systemctl daemon-reload

echo "application layout installed; continue with docs/INSTALL.md"
