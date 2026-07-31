#!/usr/bin/env bash
set -euo pipefail

# Upgrade an existing /opt/dalu-parts installation using the extracted release
# package. Database/media backups and a program snapshot are made before any
# installed application file is replaced.
if [[ "${EUID}" -ne 0 ]]; then
  echo "run as root: sudo ./scripts/upgrade.sh" >&2
  exit 1
fi

PACKAGE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_ROOT="/opt/dalu-parts"
ROLLBACK_ROOT="${ROLLBACK_ROOT:-/var/backups/dalu-parts/releases}"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
SNAPSHOT_DIR="${ROLLBACK_ROOT}/${STAMP}"
WAS_RUNNING=0

for required in "${PACKAGE_ROOT}/server" "${PACKAGE_ROOT}/initdb" "${PACKAGE_ROOT}/public/index.html"; do
  [[ -e "${required}" ]] || { echo "release package is incomplete: ${required}" >&2; exit 2; }
done
[[ -x "${PACKAGE_ROOT}/scripts/install-layout.sh" ]] || { echo "install script is missing" >&2; exit 2; }
[[ -d "${INSTALL_ROOT}" && -x "${INSTALL_ROOT}/server" ]] || {
  echo "no existing installation found; use docs/INSTALL.md instead" >&2
  exit 2
}
[[ -r /etc/dalu-parts/dlc.env ]] || { echo "missing /etc/dalu-parts/dlc.env" >&2; exit 2; }

restore_program() {
  local status="$?"
  trap - ERR
  echo "upgrade failed; restoring the previous program files from ${SNAPSHOT_DIR}" >&2
  systemctl stop dalu-parts.service || true
  [[ -f "${SNAPSHOT_DIR}/server" ]] && install -o root -g root -m 0755 "${SNAPSHOT_DIR}/server" "${INSTALL_ROOT}/server"
  [[ -f "${SNAPSHOT_DIR}/initdb" ]] && install -o root -g root -m 0755 "${SNAPSHOT_DIR}/initdb" "${INSTALL_ROOT}/initdb"
  if [[ -d "${SNAPSHOT_DIR}/public" ]]; then
    rm -rf "${INSTALL_ROOT}/public"
    cp -a "${SNAPSHOT_DIR}/public" "${INSTALL_ROOT}/public"
    chown -R root:root "${INSTALL_ROOT}/public"
  fi
  [[ -f "${SNAPSHOT_DIR}/dalu-parts.service" ]] && install -o root -g root -m 0644 "${SNAPSHOT_DIR}/dalu-parts.service" /etc/systemd/system/dalu-parts.service
  [[ -f "${SNAPSHOT_DIR}/dlc.env" ]] && install -o root -g dlc -m 0640 "${SNAPSHOT_DIR}/dlc.env" /etc/dalu-parts/dlc.env
  systemctl daemon-reload
  if [[ "${WAS_RUNNING}" -eq 1 ]]; then
    systemctl start dalu-parts.service || true
  fi
  echo "program files were restored. If database migration changed data, restore the database/media backup at ${DATABASE_BACKUP} before retrying." >&2
  exit "${status}"
}

echo "[1/6] Backing up database and uploaded media"
DATABASE_BACKUP="$(BACKUP_ROOT="${ROLLBACK_ROOT}/database-media" "${PACKAGE_ROOT}/scripts/backup.sh")"

echo "[2/6] Saving rollback snapshot at ${SNAPSHOT_DIR}"
install -d -o root -g root -m 0700 "${SNAPSHOT_DIR}"
cp -a "${INSTALL_ROOT}/server" "${SNAPSHOT_DIR}/server"
[[ -f "${INSTALL_ROOT}/initdb" ]] && cp -a "${INSTALL_ROOT}/initdb" "${SNAPSHOT_DIR}/initdb"
cp -a "${INSTALL_ROOT}/public" "${SNAPSHOT_DIR}/public"
[[ -f /etc/systemd/system/dalu-parts.service ]] && cp -a /etc/systemd/system/dalu-parts.service "${SNAPSHOT_DIR}/dalu-parts.service"
cp -a /etc/dalu-parts/dlc.env "${SNAPSHOT_DIR}/dlc.env"
printf '%s\n' "database/media backup: ${DATABASE_BACKUP}" >"${SNAPSHOT_DIR}/README.txt"

if systemctl is-active --quiet dalu-parts.service; then
  WAS_RUNNING=1
fi
trap restore_program ERR

echo "[3/6] Stopping the existing service"
systemctl stop dalu-parts.service

echo "[4/6] Installing the new binaries and public assets"
"${PACKAGE_ROOT}/scripts/install-layout.sh"

# Older releases stored static files below /opt, which is intentionally
# read-only to the dlc systemd service. Preserve a custom writable location;
# only replace the old packaged default or add a missing setting.
if grep -q '^STATIC_PAGE_DIR=/opt/dalu-parts/public/static-pages$' /etc/dalu-parts/dlc.env; then
  sed -i 's|^STATIC_PAGE_DIR=/opt/dalu-parts/public/static-pages$|STATIC_PAGE_DIR=/var/lib/dalu-parts/static-pages|' /etc/dalu-parts/dlc.env
  echo "moved STATIC_PAGE_DIR to /var/lib/dalu-parts/static-pages"
elif ! grep -q '^STATIC_PAGE_DIR=' /etc/dalu-parts/dlc.env; then
  printf '\nSTATIC_PAGE_DIR=/var/lib/dalu-parts/static-pages\n' >>/etc/dalu-parts/dlc.env
  echo "added STATIC_PAGE_DIR=/var/lib/dalu-parts/static-pages"
fi

echo "[5/6] Applying database migration"
"${PACKAGE_ROOT}/scripts/migrate.sh"

echo "[6/6] Starting and checking the new service"
systemctl start dalu-parts.service
"${PACKAGE_ROOT}/scripts/health-check.sh"
trap - ERR

echo "upgrade completed"
echo "database/media backup: ${DATABASE_BACKUP}"
echo "program rollback snapshot: ${SNAPSHOT_DIR}"
echo "If this release updates Nginx, follow docs/UPGRADE.md before considering the upgrade complete."
