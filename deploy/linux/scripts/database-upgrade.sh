#!/usr/bin/env bash
set -euo pipefail

# Upgrade only the database using the initdb binary in this extracted release.
# The script is idempotent: an already-current schema exits successfully.
if [[ "${EUID}" -ne 0 ]]; then
  echo "run as root: sudo ./scripts/database-upgrade.sh" >&2
  exit 1
fi

PACKAGE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="/etc/dalu-parts/dlc.env"
MIGRATION_BINARY="${PACKAGE_ROOT}/initdb"
WAS_RUNNING=0

[[ -x "${MIGRATION_BINARY}" ]] || { echo "release migration binary is missing: ${MIGRATION_BINARY}" >&2; exit 2; }
[[ -r "${ENV_FILE}" ]] || { echo "missing ${ENV_FILE}" >&2; exit 2; }
id dlc >/dev/null 2>&1 || { echo "system user dlc does not exist; install the application layout first" >&2; exit 2; }

if systemctl is-active --quiet dalu-parts.service; then
  WAS_RUNNING=1
fi

echo "[1/4] Backing up database and uploaded media"
BACKUP_PATH="$(ENV_FILE="${ENV_FILE}" BACKUP_ROOT="${BACKUP_ROOT:-/var/backups/dalu-parts/database-media}" "${PACKAGE_ROOT}/scripts/backup.sh")"

echo "[2/4] Stopping application service"
if [[ "${WAS_RUNNING}" -eq 1 ]]; then
  systemctl stop dalu-parts.service
fi

echo "[3/4] Applying database schema migration"
if ! MIGRATION_BINARY="${MIGRATION_BINARY}" \
  MIGRATION_WORKING_DIRECTORY="${PACKAGE_ROOT}" \
  MIGRATION_ONLY=true "${PACKAGE_ROOT}/scripts/migrate.sh"; then
  echo "database migration failed; the application remains stopped" >&2
  echo "restore backup before retrying: ${BACKUP_PATH}" >&2
  exit 1
fi

echo "[4/4] Restoring application service"
if [[ "${WAS_RUNNING}" -eq 1 ]]; then
  systemctl start dalu-parts.service
  "${PACKAGE_ROOT}/scripts/health-check.sh"
fi

echo "database upgrade completed"
echo "backup: ${BACKUP_PATH}"
