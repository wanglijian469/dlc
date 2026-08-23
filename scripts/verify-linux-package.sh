#!/usr/bin/env bash
set -euo pipefail

ARCHIVE="${1:?usage: verify-linux-package.sh <archive.tar.gz>}"
case "$ARCHIVE" in
  *.tar.gz) ;;
  *) echo "archive must end in .tar.gz" >&2; exit 2 ;;
esac

WORK_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "${WORK_DIR}"
}
trap cleanup EXIT

tar -xzf "${ARCHIVE}" -C "${WORK_DIR}"
mapfile -t ROOTS < <(find "${WORK_DIR}" -mindepth 1 -maxdepth 1 -type d)
test "${#ROOTS[@]}" -eq 1
ROOT="${ROOTS[0]}"

for required in \
  server initdb VERSION SHA256SUMS README.md \
  public/index.html config/dlc.env.example \
  systemd/dalu-parts.service nginx/dalu-parts.conf.template \
  scripts/install-layout.sh scripts/install-mysql-8.0.46.sh \
  scripts/create-mysql-user.sh scripts/configure-site-url.sh \
  scripts/render-nginx-config.sh \
  scripts/migrate.sh scripts/database-upgrade.sh scripts/upgrade.sh scripts/start.sh scripts/stop.sh \
  scripts/health-check.sh scripts/backup.sh \
  docs/INSTALL.md docs/UPGRADE.md docs/RELEASE_NOTES.md; do
  test -e "${ROOT}/${required}" || {
    echo "missing package entry: ${required}" >&2
    exit 1
  }
done

file "${ROOT}/server" "${ROOT}/initdb" | grep -E 'ELF 64-bit.*x86-64'
readelf -h "${ROOT}/server" | grep -E 'Class:.*ELF64'
readelf -h "${ROOT}/server" | grep -E 'Machine:.*X86-64'
readelf -h "${ROOT}/initdb" | grep -E 'Machine:.*X86-64'

if find "${ROOT}" -type f \( \
  -name '.env' -o -name '*.xls' -o \
  -name '*.sql' -o -name '*.db' -o -name '*.sqlite*' \
  \) -print -quit | grep -q .; then
  echo "package contains a secret, database, or user document file" >&2
  exit 1
fi
if find "${ROOT}" -type f -name '*.xlsx' \
  ! -path "${ROOT}/public/templates/*" -print -quit | grep -q .; then
  echo "package contains an unexpected Excel document" >&2
  exit 1
fi
test ! -e "${ROOT}/media_storage"

find "${ROOT}/scripts" -type f -name '*.sh' -print0 | xargs -0 -n1 bash -n
"${ROOT}/scripts/render-nginx-config.sh" \
  example.cn \
  /etc/pki/tls/certs/example.crt \
  /etc/pki/tls/private/example.key \
  "${WORK_DIR}/nginx.conf"
! grep -q '__[A-Z_]*__' "${WORK_DIR}/nginx.conf"

(
  cd "${ROOT}"
  sha256sum --check SHA256SUMS
)

echo "verified ${ARCHIVE}"
