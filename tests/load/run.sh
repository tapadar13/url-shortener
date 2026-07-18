#!/usr/bin/env bash

set -Eeuo pipefail

root_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
compose=(
  docker compose
  --file "${root_directory}/deploy/docker-compose.load.yml"
  --profile load
)

scenario="${1:-}"
case "${scenario}" in
  smoke)
    environment_variables=(SMOKE_VUS SMOKE_DURATION SMOKE_P95_MS)
    ;;
  management)
    environment_variables=(
      MANAGEMENT_VUS
      MANAGEMENT_DURATION
      MANAGEMENT_P95_MS
    )
    ;;
  redirects)
    environment_variables=(
      REDIRECT_LINKS
      REDIRECT_RATE
      REDIRECT_DURATION
      REDIRECT_PREALLOCATED_VUS
      REDIRECT_MAX_VUS
      REDIRECT_P95_MS
    )
    ;;
  *)
    printf 'usage: %s {smoke|management|redirects}\n' "$0" >&2
    exit 2
    ;;
esac

cleanup() {
  local exit_code=$?
  trap - EXIT
  "${compose[@]}" down --volumes --remove-orphans || true
  exit "${exit_code}"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

"${compose[@]}" down --volumes --remove-orphans
"${compose[@]}" up --detach --build --wait api

run_options=(--rm)
for variable in "${environment_variables[@]}"; do
  if value="$(printenv "${variable}")"; then
    run_options+=(--env "${variable}=${value}")
  fi
done

"${compose[@]}" run "${run_options[@]}" k6 run "/scripts/${scenario}.js"
