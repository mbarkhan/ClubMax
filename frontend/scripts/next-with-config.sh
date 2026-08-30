#!/usr/bin/env bash

set -euo pipefail

FRONTEND_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
PROJECT_DIR="$(cd -- "$FRONTEND_DIR/.." && pwd)"

set -a
# shellcheck disable=SC1091
source "$PROJECT_DIR/clubmax.env"
set +a

export NEXT_PUBLIC_API_URL="http://$CLUBMAX_PUBLIC_HOST:$CLUBMAX_BACKEND_PORT"

command="${1:?Next.js command is required}"
shift
if [[ "$command" == "dev" || "$command" == "start" ]]; then
    exec next "$command" -H "$CLUBMAX_BIND_IP" -p "$CLUBMAX_FRONTEND_PORT" "$@"
fi
exec next "$command" "$@"
