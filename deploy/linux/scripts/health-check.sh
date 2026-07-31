#!/usr/bin/env bash
set -euo pipefail

LOCAL_URL="${LOCAL_URL:-http://127.0.0.1:8080}"
HEALTH_RETRIES="${HEALTH_RETRIES:-30}"

wait_for_health() {
  local url="$1"
  local retries="$2"
  local attempt

  for ((attempt = 1; attempt <= retries; attempt++)); do
    if curl --fail --silent --max-time 3 "${url}"; then
      echo
      return 0
    fi
    sleep 1
  done

  echo "health check failed after ${retries} attempts: ${url}" >&2
  curl --fail --silent --show-error --max-time 10 "${url}"
}

wait_for_health "${LOCAL_URL}/api/health" "${HEALTH_RETRIES}"

if [[ -n "${PUBLIC_URL:-}" ]]; then
  wait_for_health "${PUBLIC_URL%/}/api/health" "${PUBLIC_HEALTH_RETRIES:-10}"
fi
