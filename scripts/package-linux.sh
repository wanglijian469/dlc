#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C
export TZ=UTC

VERSION="${1:-v2026.07.30-linux.1}"
case "$VERSION" in
  *[!A-Za-z0-9._-]*|'')
    echo "version may contain only letters, numbers, dots, underscores, and hyphens" >&2
    exit 2
    ;;
esac

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PACKAGE_NAME="dlc-deploy-linux-x86_64-${VERSION}"
RELEASE_DIR="${ROOT}/release"
WORK_DIR="$(mktemp -d)"
PACKAGE_ROOT="${WORK_DIR}/${PACKAGE_NAME}"
OUTPUT="${RELEASE_DIR}/${PACKAGE_NAME}.tar.gz"
SOURCE_EPOCH="${SOURCE_DATE_EPOCH:-$(git -C "${ROOT}" log -1 --format=%ct)}"
export SOURCE_DATE_EPOCH

cleanup() {
  rm -rf "${WORK_DIR}"
}
trap cleanup EXIT

mkdir -p "${PACKAGE_ROOT}/public" \
  "${PACKAGE_ROOT}/config" \
  "${PACKAGE_ROOT}/systemd" \
  "${PACKAGE_ROOT}/nginx" \
  "${PACKAGE_ROOT}/scripts" \
  "${PACKAGE_ROOT}/docs" \
  "${RELEASE_DIR}"

(
  cd "${ROOT}/frontend"
  npm run build
)

(
  cd "${ROOT}/backend"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -buildvcs=false -ldflags="-s -w" \
    -o "${PACKAGE_ROOT}/server" ./cmd/server
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -buildvcs=false -ldflags="-s -w" \
    -o "${PACKAGE_ROOT}/initdb" ./cmd/initdb
)

cp -a "${ROOT}/frontend/dist/." "${PACKAGE_ROOT}/public/"
cp "${ROOT}/deploy/linux/config/dlc.env.example" "${PACKAGE_ROOT}/config/dlc.env.example"
cp "${ROOT}/deploy/linux/systemd/dalu-parts.service" "${PACKAGE_ROOT}/systemd/"
cp "${ROOT}/deploy/linux/nginx/dalu-parts.conf.template" "${PACKAGE_ROOT}/nginx/"
cp "${ROOT}/deploy/linux/scripts/"*.sh "${PACKAGE_ROOT}/scripts/"
cp "${ROOT}/deploy/linux/README.md" "${PACKAGE_ROOT}/README.md"
cp "${ROOT}/deploy/linux/INSTALL.md" "${PACKAGE_ROOT}/docs/"
cp "${ROOT}/deploy/linux/UPGRADE.md" "${PACKAGE_ROOT}/docs/"
cp "${ROOT}/deploy/linux/RELEASE_NOTES.md" "${PACKAGE_ROOT}/docs/"

chmod 0755 "${PACKAGE_ROOT}/server" "${PACKAGE_ROOT}/initdb" "${PACKAGE_ROOT}/scripts/"*.sh
chmod 0644 "${PACKAGE_ROOT}/config/dlc.env.example" \
  "${PACKAGE_ROOT}/systemd/dalu-parts.service" \
  "${PACKAGE_ROOT}/nginx/dalu-parts.conf.template" \
  "${PACKAGE_ROOT}/README.md" \
  "${PACKAGE_ROOT}/docs/"*.md

printf '%s\n' "${VERSION}" >"${PACKAGE_ROOT}/VERSION"
(
  cd "${PACKAGE_ROOT}"
  find . -type f ! -name SHA256SUMS -print0 \
    | sort -z \
    | xargs -0 sha256sum >SHA256SUMS
)

find "${PACKAGE_ROOT}" -exec touch -h -d "@${SOURCE_EPOCH}" {} +
rm -f "${OUTPUT}"
tar --sort=name \
  --mtime="@${SOURCE_EPOCH}" \
  --owner=0 --group=0 --numeric-owner \
  -C "${WORK_DIR}" -cf - "${PACKAGE_NAME}" \
  | gzip -n -9 >"${OUTPUT}"

echo "${OUTPUT}"
