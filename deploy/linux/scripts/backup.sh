#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${ENV_FILE:-/etc/dalu-parts/dlc.env}"
BACKUP_ROOT="${BACKUP_ROOT:-/var/backups/dalu-parts}"
[[ -r "${ENV_FILE}" ]] || {
  echo "cannot read ${ENV_FILE}" >&2
  exit 1
}

while IFS='=' read -r key value; do
  [[ -z "${key}" || "${key}" == \#* ]] && continue
  case "${key}" in
    DB_HOST|DB_PORT|DB_USER|DB_PASSWORD|DB_NAME|MEDIA_DIR)
      printf -v "${key}" '%s' "${value}"
      ;;
  esac
done <"${ENV_FILE}"

: "${DB_HOST:?missing DB_HOST}"
: "${DB_PORT:?missing DB_PORT}"
: "${DB_USER:?missing DB_USER}"
: "${DB_PASSWORD:?missing DB_PASSWORD}"
: "${DB_NAME:?missing DB_NAME}"
: "${MEDIA_DIR:?missing MEDIA_DIR}"
[[ "${DB_PASSWORD}" != *$'\n'* && "${DB_PASSWORD}" != *$'\r'* ]] || {
  echo "DB_PASSWORD must not contain newlines" >&2
  exit 2
}

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
DEST="${BACKUP_ROOT}/${STAMP}"
CLIENT_CNF="$(mktemp)"
cleanup() {
  rm -f "${CLIENT_CNF}"
}
trap cleanup EXIT
umask 077
mkdir -p "${DEST}"

ESCAPED_PASSWORD="${DB_PASSWORD//\\/\\\\}"
printf '[client]\nhost=%s\nport=%s\nuser=%s\npassword=%s\n' \
  "${DB_HOST}" "${DB_PORT}" "${DB_USER}" "${ESCAPED_PASSWORD}" >"${CLIENT_CNF}"

mysqldump --defaults-extra-file="${CLIENT_CNF}" \
  --single-transaction --triggers --no-tablespaces \
  --set-gtid-purged=OFF "${DB_NAME}" | gzip -9 >"${DEST}/database.sql.gz"
tar -C "$(dirname "${MEDIA_DIR}")" -czf "${DEST}/media_storage.tar.gz" \
  "$(basename "${MEDIA_DIR}")"
(
  cd "${DEST}"
  sha256sum database.sql.gz media_storage.tar.gz >SHA256SUMS
)

echo "${DEST}"
