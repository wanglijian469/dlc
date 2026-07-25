#!/usr/bin/env bash
set -euo pipefail

SITE_URL="${1:?usage: configure-site-url.sh https://www.example.cn}"
ENV_FILE="${ENV_FILE:-/etc/dalu-parts/dlc.env}"
[[ "${SITE_URL}" =~ ^https://[A-Za-z0-9.-]+(:[0-9]+)?/?$ ]] || {
  echo "site URL must be an HTTPS origin without a path" >&2
  exit 2
}
[[ -r "${ENV_FILE}" ]] || {
  echo "cannot read ${ENV_FILE}" >&2
  exit 1
}

while IFS='=' read -r key value; do
  [[ -z "${key}" || "${key}" == \#* ]] && continue
  case "${key}" in
    DB_HOST|DB_PORT|DB_USER|DB_PASSWORD|DB_NAME)
      printf -v "${key}" '%s' "${value}"
      ;;
  esac
done <"${ENV_FILE}"

: "${DB_HOST:?missing DB_HOST}"
: "${DB_PORT:?missing DB_PORT}"
: "${DB_USER:?missing DB_USER}"
: "${DB_PASSWORD:?missing DB_PASSWORD}"
: "${DB_NAME:?missing DB_NAME}"
CLIENT_CNF="$(mktemp)"
cleanup() {
  rm -f "${CLIENT_CNF}"
}
trap cleanup EXIT
umask 077
ESCAPED_PASSWORD="${DB_PASSWORD//\\/\\\\}"
printf '[client]\nhost=%s\nport=%s\nuser=%s\npassword=%s\n' \
  "${DB_HOST}" "${DB_PORT}" "${DB_USER}" "${ESCAPED_PASSWORD}" >"${CLIENT_CNF}"

mysql --defaults-extra-file="${CLIENT_CNF}" "${DB_NAME}" --execute \
  "UPDATE site_configs SET config_value=JSON_SET(config_value, '$.siteUrl', '${SITE_URL}') WHERE config_key='site.meta'"

echo "configured site.meta.siteUrl=${SITE_URL}"
