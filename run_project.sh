#!/usr/bin/env bash

set -euo pipefail

PROJECT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RUN_DIR="$PROJECT_DIR/.local/run"
LOG_DIR="$RUN_DIR/logs"

set -a
# shellcheck disable=SC1091
source "$PROJECT_DIR/clubmax.env"
set +a

mkdir -p "$LOG_DIR"

is_running() {
    local pid_file="$1"
    [[ -f "$pid_file" ]] && kill -0 "$(<"$pid_file")" 2>/dev/null
}

listener_pid() {
    local port="$1"
    ss -ltnp "sport = :$port" 2>/dev/null \
        | sed -n 's/.*pid=\([0-9][0-9]*\).*/\1/p' \
        | head -n 1
}

adopt_running_service() {
    local name="$1"
    local port="$2"
    local health_url="$3"
    local pid_file="$RUN_DIR/$name.pid"

    if ! curl --silent --fail --max-time 2 "$health_url" >/dev/null 2>&1; then
        return 1
    fi

    local pid
    pid="$(listener_pid "$port")"
    if [[ -n "$pid" ]]; then
        echo "$pid" >"$pid_file"
        echo "$name is already running (PID $pid); adopted existing process."
    else
        echo "$name is already available at $health_url."
    fi
    return 0
}

start_service() {
    local name="$1"
    local work_dir="$2"
    shift 2

    local pid_file="$RUN_DIR/$name.pid"
    local log_file="$LOG_DIR/$name.log"

    if is_running "$pid_file"; then
        echo "$name is already running (PID $(<"$pid_file"))."
        return
    fi

    rm -f "$pid_file"
    (
        cd "$work_dir"
        exec setsid "$@"
    ) >>"$log_file" 2>&1 &
    local pid=$!
    echo "$pid" >"$pid_file"

    sleep 1
    if ! kill -0 "$pid" 2>/dev/null; then
        rm -f "$pid_file"
        echo "Failed to start $name. See $log_file" >&2
        return 1
    fi

    echo "$name started (PID $pid, log: $log_file)."
}

if command -v docker >/dev/null 2>&1; then
    if docker container inspect clubmax-db >/dev/null 2>&1; then
        docker start clubmax-db >/dev/null
    else
        docker run -d \
            --name clubmax-db \
            -e POSTGRES_DB="$CLUBMAX_DATABASE_NAME" \
            -e POSTGRES_USER="$CLUBMAX_DATABASE_USER" \
            -e POSTGRES_PASSWORD="$CLUBMAX_DATABASE_PASSWORD" \
            -p "$CLUBMAX_SERVER_IP:$CLUBMAX_DATABASE_PORT:5432" \
            -v clubmax_postgres_data:/var/lib/postgresql/data \
            postgres:17-alpine >/dev/null
    fi
    echo "Waiting for the ClubMax database..."
    for _ in {1..30}; do
        if docker exec clubmax-db pg_isready -U "$CLUBMAX_DATABASE_USER" -d "$CLUBMAX_DATABASE_NAME" >/dev/null 2>&1; then
            break
        fi
        sleep 1
    done
    if ! docker exec clubmax-db pg_isready -U "$CLUBMAX_DATABASE_USER" -d "$CLUBMAX_DATABASE_NAME" >/dev/null 2>&1; then
        echo "ClubMax database did not become ready." >&2
        exit 1
    fi
else
    echo "docker was not found; assuming PostgreSQL is already running." >&2
fi

if ! adopt_running_service backend "$CLUBMAX_BACKEND_PORT" "http://$CLUBMAX_SERVER_IP:$CLUBMAX_BACKEND_PORT/healthz"; then
    start_service backend "$PROJECT_DIR/backend" go run main.go
fi

if [[ ! -x "$PROJECT_DIR/frontend/node_modules/.bin/next" ]]; then
    echo "Installing frontend dependencies..."
    (cd "$PROJECT_DIR/frontend" && npm ci)
fi

if ! adopt_running_service frontend "$CLUBMAX_FRONTEND_PORT" "http://$CLUBMAX_SERVER_IP:$CLUBMAX_FRONTEND_PORT/fa"; then
    start_service frontend "$PROJECT_DIR/frontend" npm run dev
fi

echo "ClubMax is available locally at http://127.0.0.1:$CLUBMAX_FRONTEND_PORT."
echo "Local-network address: http://$CLUBMAX_PUBLIC_HOST:$CLUBMAX_FRONTEND_PORT (API: http://$CLUBMAX_PUBLIC_HOST:$CLUBMAX_BACKEND_PORT)."
