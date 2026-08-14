#!/usr/bin/env bash

set -euo pipefail

PROJECT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RUN_DIR="$PROJECT_DIR/.local/run"

set -a
# shellcheck disable=SC1091
source "$PROJECT_DIR/gymax.env"
set +a

stop_service() {
    local name="$1"
    local pid_file="$RUN_DIR/$name.pid"

    if [[ ! -f "$pid_file" ]]; then
        echo "$name is not running (no PID file)."
        return
    fi

    local pid
    pid="$(<"$pid_file")"
    if kill -0 "$pid" 2>/dev/null; then
        kill -TERM -- "-$pid" 2>/dev/null || kill -TERM "$pid" 2>/dev/null || true
        for _ in {1..20}; do
            if ! kill -0 "$pid" 2>/dev/null; then
                break
            fi
            sleep 0.25
        done
        if kill -0 "$pid" 2>/dev/null; then
            kill -KILL -- "-$pid" 2>/dev/null || kill -KILL "$pid" 2>/dev/null || true
        fi
        echo "$name stopped."
    else
        echo "$name was not running; removed stale PID file."
    fi

    rm -f "$pid_file"
}

stop_port_listener() {
    local name="$1"
    local port="$2"
    local pid
    pid="$(ss -ltnp "sport = :$port" 2>/dev/null | sed -n 's/.*pid=\([0-9][0-9]*\).*/\1/p' | head -n 1)"
    if [[ -n "$pid" ]]; then
        kill -TERM "$pid" 2>/dev/null || true
        echo "$name listener on port $port stopped (PID $pid)."
    fi
}

stop_service frontend
stop_service backend
stop_port_listener frontend "$GYMAX_FRONTEND_PORT"
stop_port_listener backend "$GYMAX_BACKEND_PORT"

if command -v docker >/dev/null 2>&1 && docker container inspect gymax-db >/dev/null 2>&1; then
    docker stop gymax-db >/dev/null
else
    echo "GyMax PostgreSQL container is not running."
fi

echo "GyMax stopped."
