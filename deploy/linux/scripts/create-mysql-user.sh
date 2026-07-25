#!/usr/bin/env bash
set -euo pipefail

DB_NAME="${DB_NAME:-dl_nongji_parts}"
DB_USER="${DB_USER:-dlc_app}"
DB_HOST_SCOPE="${DB_HOST_SCOPE:-127.0.0.1}"

[[ "${DB_NAME}" =~ ^[A-Za-z0-9_]+$ ]] || {
  echo "DB_NAME may contain only letters, numbers, and underscores" >&2
  exit 2
}
[[ "${DB_USER}" =~ ^[A-Za-z0-9_]+$ ]] || {
  echo "DB_USER may contain only letters, numbers, and underscores" >&2
  exit 2
}
[[ "${DB_HOST_SCOPE}" == "127.0.0.1" || "${DB_HOST_SCOPE}" == "localhost" ]] || {
  echo "DB_HOST_SCOPE must be 127.0.0.1 or localhost" >&2
  exit 2
}

if [[ -z "${DB_PASSWORD:-}" ]]; then
  read -r -s -p "Password for ${DB_USER}: " DB_PASSWORD
  echo
fi
[[ "${DB_PASSWORD}" != *$'\n'* && "${DB_PASSWORD}" != *$'\r'* ]] || {
  echo "DB_PASSWORD must not contain newlines" >&2
  exit 2
}
ESCAPED_PASSWORD="${DB_PASSWORD//\\/\\\\}"
ESCAPED_PASSWORD="${ESCAPED_PASSWORD//\'/\'\'}"

mysql --protocol=socket --user=root -p <<SQL
CREATE DATABASE IF NOT EXISTS \`${DB_NAME}\`
  DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS '${DB_USER}'@'${DB_HOST_SCOPE}'
  IDENTIFIED BY '${ESCAPED_PASSWORD}';
ALTER USER '${DB_USER}'@'${DB_HOST_SCOPE}'
  IDENTIFIED BY '${ESCAPED_PASSWORD}';
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX, DROP,
  SHOW VIEW, TRIGGER
  ON \`${DB_NAME}\`.* TO '${DB_USER}'@'${DB_HOST_SCOPE}';
FLUSH PRIVILEGES;
SQL

echo "created database ${DB_NAME} and least-scope application account ${DB_USER}@${DB_HOST_SCOPE}"
