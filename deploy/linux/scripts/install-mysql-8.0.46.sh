#!/usr/bin/env bash
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "run as root: sudo ./scripts/install-mysql-8.0.46.sh" >&2
  exit 1
fi

if [[ -f /etc/os-release ]]; then
  # shellcheck disable=SC1091
  source /etc/os-release
fi
if [[ "${ID:-}" != "alinux" || "${VERSION_ID:-}" != 3* ]]; then
  echo "this installer is restricted to Alibaba Cloud Linux 3" >&2
  exit 1
fi
EXPECTED_VERSION="8.0.46"
EXPECTED_RELEASE="1.el8"

if rpm -q mariadb-server >/dev/null 2>&1; then
  echo "an existing MySQL/MariaDB server package was found; use a reviewed migration plan" >&2
  exit 1
fi

verify_package() {
  local package="$1"
  local version release
  version="$(rpm -q --queryformat '%{VERSION}' "${package}")"
  release="$(rpm -q --queryformat '%{RELEASE}' "${package}")"
  [[ "${version}" == "${EXPECTED_VERSION}" && "${release}" == "${EXPECTED_RELEASE}" ]] || {
    echo "${package} is ${version}-${release}; expected ${EXPECTED_VERSION}-${EXPECTED_RELEASE}" >&2
    return 1
  }
}

if rpm -q mysql-community-server >/dev/null 2>&1; then
  verify_package mysql-community-server
else
  if rpm -q mysql-server >/dev/null 2>&1; then
    echo "a non-Community MySQL server package was found; use a reviewed migration plan" >&2
    exit 1
  fi
  dnf install -y dnf-plugins-core
  dnf install -y https://repo.mysql.com/mysql84-community-release-el8.rpm
  dnf config-manager --set-disabled 'mysql*-community' || true
  dnf config-manager --set-enabled mysql80-community
  dnf --disablerepo='mysql*' --enablerepo=mysql80-community install -y \
    "mysql-community-client-${EXPECTED_VERSION}-${EXPECTED_RELEASE}.x86_64" \
    "mysql-community-server-${EXPECTED_VERSION}-${EXPECTED_RELEASE}.x86_64"
fi

verify_package mysql-community-client
verify_package mysql-community-server
systemctl enable --now mysqld
echo "MySQL Community ${EXPECTED_VERSION}-${EXPECTED_RELEASE} is ready."
echo "For a fresh install, read the temporary root password from /var/log/mysqld.log"
echo "and run mysql_secure_installation."
