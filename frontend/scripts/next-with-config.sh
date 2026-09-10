#!/usr/bin/env bash

set -euo pipefail

FRONTEND_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
PROJECT_DIR="$(cd -- "$FRONTEND_DIR/.." && pwd)"

set -a
# shellcheck disable=SC1091
source "$PROJECT_DIR/clubmax.env"
set +a

# Browser requests stay on the frontend origin. Next proxies them to the local
# backend, so DHCP address changes and HTTPS mixed-content rules cannot break
# API access.
export CLUBMAX_INTERNAL_API_URL="http://127.0.0.1:$CLUBMAX_BACKEND_PORT"

command="${1:?Next.js command is required}"
shift
if [[ "$command" == "dev" || "$command" == "start" ]]; then
    exec next "$command" -H "$CLUBMAX_BIND_IP" -p "$CLUBMAX_FRONTEND_PORT" "$@"
fi
exec next "$command" "$@"
