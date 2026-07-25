#!/usr/bin/env bash
set -euo pipefail

SERVER_NAME="${1:?usage: render-nginx-config.sh <domain> <certificate> <key> [output]}"
CERTIFICATE="${2:?certificate path is required}"
CERTIFICATE_KEY="${3:?certificate key path is required}"
OUTPUT="${4:-/etc/nginx/conf.d/dalu-parts.conf}"
PACKAGE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEMPLATE="${PACKAGE_ROOT}/nginx/dalu-parts.conf.template"

[[ "${SERVER_NAME}" =~ ^[A-Za-z0-9.-]+$ ]] || {
  echo "invalid domain name" >&2
  exit 2
}
for path in "${CERTIFICATE}" "${CERTIFICATE_KEY}" "${OUTPUT}"; do
  [[ "${path}" != *'|'* && "${path}" != *$'\n'* && "${path}" != *$'\r'* ]] || {
    echo "paths must not contain pipes or newlines" >&2
    exit 2
  }
done

sed \
  -e "s|__SERVER_NAME__|${SERVER_NAME}|g" \
  -e "s|__TLS_CERTIFICATE__|${CERTIFICATE}|g" \
  -e "s|__TLS_CERTIFICATE_KEY__|${CERTIFICATE_KEY}|g" \
  "${TEMPLATE}" >"${OUTPUT}"

echo "rendered ${OUTPUT}"
