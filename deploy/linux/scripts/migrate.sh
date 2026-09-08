#!/usr/bin/env bash
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "run as root: sudo ./scripts/migrate.sh" >&2
  exit 1
fi

MIGRATION_BINARY="${MIGRATION_BINARY:-/opt/dalu-parts/initdb}"
MIGRATION_WORKING_DIRECTORY="${MIGRATION_WORKING_DIRECTORY:-/opt/dalu-parts}"
[[ -x "${MIGRATION_BINARY}" ]] || { echo "migration binary is missing: ${MIGRATION_BINARY}" >&2; exit 2; }
[[ -r /etc/dalu-parts/dlc.env ]] || { echo "missing /etc/dalu-parts/dlc.env" >&2; exit 2; }
MIGRATION_ARGS=()
if [[ "${MIGRATION_ONLY:-false}" == "true" ]]; then
  MIGRATION_ARGS+=(--migrate-only)
fi
UNIT="dalu-parts-initdb-$(date -u +%Y%m%dT%H%M%SZ)"

systemd-run \
	--unit="${UNIT}" \
	--wait --collect --pipe \
	--property=User=dlc \
	--property=Group=dlc \
	--property="WorkingDirectory=${MIGRATION_WORKING_DIRECTORY}" \
	--property=EnvironmentFile=/etc/dalu-parts/dlc.env \
	"${MIGRATION_BINARY}" "${MIGRATION_ARGS[@]}"
