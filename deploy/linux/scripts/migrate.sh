#!/usr/bin/env bash
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "run as root: sudo ./scripts/migrate.sh" >&2
  exit 1
fi

systemd-run \
  --unit=dalu-parts-initdb \
  --wait --collect --pipe \
  --property=User=dlc \
  --property=Group=dlc \
  --property=WorkingDirectory=/opt/dalu-parts \
  --property=EnvironmentFile=/etc/dalu-parts/dlc.env \
  /opt/dalu-parts/initdb
