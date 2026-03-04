#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"
BACKEND_PORT=8080
FRONTEND_PORT=5173
BACKEND_URL="http://localhost:$BACKEND_PORT"
HEALTH_ENDPOINT="$BACKEND_URL/admin/config"
MAX_WAIT=15

cleanup() {
  echo ""
  echo "Shutting down..."
  # Kill anything listening on our ports — most reliable on macOS
  lsof -ti :"$BACKEND_PORT" -sTCP:LISTEN 2>/dev/null | xargs kill 2>/dev/null || true
  lsof -ti :"$FRONTEND_PORT" -sTCP:LISTEN 2>/dev/null | xargs kill 2>/dev/null || true
  sleep 1
  # Force-kill stragglers
  lsof -ti :"$BACKEND_PORT" -sTCP:LISTEN 2>/dev/null | xargs kill -9 2>/dev/null || true
  lsof -ti :"$FRONTEND_PORT" -sTCP:LISTEN 2>/dev/null | xargs kill -9 2>/dev/null || true
  # Clean up built binary
  rm -f "$BACKEND_DIR/server"
  echo "Done."
}
trap cleanup EXIT

# --- Check prerequisites ---
if ! command -v go &>/dev/null; then
  echo "Error: go is not installed" >&2; exit 1
fi
if ! command -v node &>/dev/null; then
  echo "Error: node is not installed" >&2; exit 1
fi

# --- Check ports ---
if lsof -i :"$BACKEND_PORT" -sTCP:LISTEN &>/dev/null; then
  echo "Error: port $BACKEND_PORT is already in use" >&2; exit 1
fi
if lsof -i :"$FRONTEND_PORT" -sTCP:LISTEN &>/dev/null; then
  echo "Error: port $FRONTEND_PORT is already in use" >&2; exit 1
fi

# --- Install frontend deps if needed ---
if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
  echo "Installing frontend dependencies..."
  (cd "$FRONTEND_DIR" && npm install)
fi

# --- Build and start backend ---
echo "Building backend..."
(cd "$BACKEND_DIR" && go build -o "$BACKEND_DIR/server" .)

DB_PATH="$BACKEND_DIR/biddings.db"

echo "Starting backend on :$BACKEND_PORT (db: $DB_PATH)..."
"$BACKEND_DIR/server" -port "$BACKEND_PORT" -db "$DB_PATH" &
BACKEND_PID=$!

echo -n "Waiting for backend"
elapsed=0
while [ $elapsed -lt $MAX_WAIT ]; do
  if curl -sf "$HEALTH_ENDPOINT" >/dev/null 2>&1; then
    echo " ready!"
    break
  fi
  if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
    echo ""; echo "Error: backend process exited unexpectedly" >&2; exit 1
  fi
  echo -n "."
  sleep 1
  elapsed=$((elapsed + 1))
done
if [ $elapsed -ge $MAX_WAIT ]; then
  echo ""; echo "Error: backend did not respond within ${MAX_WAIT}s" >&2; exit 1
fi

# --- Start frontend ---
echo "Starting frontend..."
(cd "$FRONTEND_DIR" && exec npx vite) &
FRONTEND_PID=$!

FRONTEND_URL="http://localhost:$FRONTEND_PORT"

# Wait for frontend to be ready, then open in browser
(
  elapsed=0
  while [ $elapsed -lt $MAX_WAIT ]; do
    if curl -sf "$FRONTEND_URL" >/dev/null 2>&1; then
      open "$FRONTEND_URL" 2>/dev/null || xdg-open "$FRONTEND_URL" 2>/dev/null || true
      break
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
) &

echo ""
echo "====================================="
echo "  Backend:  $BACKEND_URL"
echo "  Frontend: $FRONTEND_URL"
echo "  Press Ctrl+C to stop both"
echo "====================================="

# Monitor both — exit if either dies, signals break sleep
while kill -0 "$BACKEND_PID" 2>/dev/null && kill -0 "$FRONTEND_PID" 2>/dev/null; do
  sleep 2 || break
done
