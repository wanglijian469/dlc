#!/usr/bin/env bash
set -euo pipefail

LOCAL_URL="${LOCAL_URL:-http://127.0.0.1:8080}"
curl --fail --silent --show-error --max-time 10 "${LOCAL_URL}/api/health"
echo

if [[ -n "${PUBLIC_URL:-}" ]]; then
  curl --fail --silent --show-error --max-time 15 "${PUBLIC_URL%/}/api/health"
  echo
fi
