#!/usr/bin/env bash
set -euo pipefail

systemctl start dalu-parts.service
systemctl --no-pager --full status dalu-parts.service
