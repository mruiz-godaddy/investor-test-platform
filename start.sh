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
DB_PATH="$BACKEND_DIR/biddings.db"

# ── Upstream URLs (from Android appTestEnv build flavor) ──
AUCTION_UPSTREAM="https://auctions.api.int.test-godaddy.com"
FIND_UPSTREAM="https://find.api.test.aws.godaddy.com"
DOMAINS_UPSTREAM="https://api.test-godaddy.com"
APPRAISAL_UPSTREAM="https://api.test-godaddy.com"
GATEWAY_UPSTREAM="https://gateway.api.int.test-godaddy.com/v2"
PAYMENT_UPSTREAM="https://payment.api.test-godaddy.com/v1"
DCC_MGNT_UPSTREAM="https://mgnt.dcc.api.test-godaddy.com/v1"
DCC_UPSTREAM="https://domains.dcc.api.test-godaddy.com/v1"
DNS_UPSTREAM="https://domdns.api.test-godaddy.com/v1"
PROTECTION_UPSTREAM="https://protection.domains.api.test-godaddy.com/v1"
SSO_HOST="sso.test-godaddy.com"
TMUX_SESSION="biddings-dev"

# --- Parse flags ---
SPLIT=true
for arg in "$@"; do
  case "$arg" in
    --no-split) SPLIT=false ;;
  esac
done

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
  # Kill tmux session if split mode
  if "$SPLIT"; then
    tmux kill-session -t "$TMUX_SESSION" 2>/dev/null || true
  fi
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
if "$SPLIT" && ! command -v tmux &>/dev/null; then
  echo "WARN: tmux not found, falling back to standard mode (install with: brew install tmux)"
  SPLIT=false
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

# --- Build backend ---
echo "Building backend..."
(cd "$BACKEND_DIR" && go build -o "$BACKEND_DIR/server" .)

if "$SPLIT"; then
  # ============================================================
  #  TMUX SPLIT-SCREEN MODE
  # ============================================================
  tmux kill-session -t "$TMUX_SESSION" 2>/dev/null || true

  # Export vars so pane commands can use them
  export BACKEND_DIR BACKEND_PORT DB_PATH
  export AUCTION_UPSTREAM FIND_UPSTREAM
  export FRONTEND_DIR FRONTEND_PORT

  # Top pane: backend server
  tmux new-session -d -s "$TMUX_SESSION" -n dev \
    "$BACKEND_DIR/server -port $BACKEND_PORT -db $DB_PATH \
      -auction-upstream $AUCTION_UPSTREAM \
      -find-upstream $FIND_UPSTREAM 2>&1"

  # Bottom pane: frontend dev server
  tmux split-window -v -t "$TMUX_SESSION:dev" \
    "cd $FRONTEND_DIR && npx vite 2>&1"

  # When ANY pane's process exits, kill the entire session
  tmux set-option -t "$TMUX_SESSION" remain-on-exit on
  tmux set-hook -t "$TMUX_SESSION" pane-died "kill-session -t $TMUX_SESSION"

  # Even vertical split — rebalance on every window resize
  tmux select-layout -t "$TMUX_SESSION" even-vertical
  tmux set-hook -t "$TMUX_SESSION" window-resized "select-layout -t $TMUX_SESSION even-vertical"

  # Label each pane with a border title (tmux 3.2+)
  tmux set-option -t "$TMUX_SESSION" pane-border-status top 2>/dev/null || true
  tmux select-pane -t "$TMUX_SESSION:dev.0" -T "BACKEND :$BACKEND_PORT" 2>/dev/null || true
  tmux select-pane -t "$TMUX_SESSION:dev.1" -T "FRONTEND :$FRONTEND_PORT" 2>/dev/null || true

  # Mouse support: click to focus pane, scroll wheel to scroll history
  tmux set-option -t "$TMUX_SESSION" mouse on

  # Copy to system clipboard: select with mouse drag, copies on release
  tmux set-option -t "$TMUX_SESSION" set-clipboard on
  tmux bind-key -T copy-mode MouseDragEnd1Pane send-keys -X copy-pipe-and-cancel "pbcopy"
  tmux bind-key -T copy-mode-vi MouseDragEnd1Pane send-keys -X copy-pipe-and-cancel "pbcopy"

  # Large scroll-back buffer (default is 2000)
  tmux set-option -t "$TMUX_SESSION" history-limit 50000

  # Arrow keys scroll: enters copy mode on first press, -e auto-exits at bottom
  tmux bind-key -T root Up copy-mode -e \; send-keys Up
  tmux bind-key -T root Down copy-mode -e \; send-keys Down

  # Status bar with quit hint
  tmux set-option -t "$TMUX_SESSION" status on
  tmux set-option -t "$TMUX_SESSION" status-style "bg=#1a1a2e,fg=#e0e0e0"
  tmux set-option -t "$TMUX_SESSION" status-left ""
  tmux set-option -t "$TMUX_SESSION" status-right " Scroll: Arrow Keys/MouseWheel │ Ctrl+C quits "
  tmux set-option -t "$TMUX_SESSION" status-right-style "bg=#e74c3c,fg=#ffffff,bold"
  tmux set-option -t "$TMUX_SESSION" status-right-length 60
  tmux set-option -t "$TMUX_SESSION" status-justify centre
  tmux set-option -t "$TMUX_SESSION" message-style "bg=#1a1a2e,fg=#e0e0e0"

  # Select backend pane
  tmux select-pane -t "$TMUX_SESSION:dev.0"

  # Wait for backend health in background, then open browser
  (
    elapsed=0
    while [ $elapsed -lt $MAX_WAIT ]; do
      if curl -sf "$HEALTH_ENDPOINT" >/dev/null 2>&1; then break; fi
      sleep 1; elapsed=$((elapsed + 1))
    done
    # Wait for frontend too, then open browser
    elapsed=0
    while [ $elapsed -lt $MAX_WAIT ]; do
      if curl -sf "http://localhost:$FRONTEND_PORT" >/dev/null 2>&1; then
        open "http://localhost:$FRONTEND_PORT" 2>/dev/null || true
        break
      fi
      sleep 1; elapsed=$((elapsed + 1))
    done
  ) &

  echo "====================================="
  echo "  Backend:  $BACKEND_URL  (top pane)"
  echo "  Frontend: http://localhost:$FRONTEND_PORT  (bottom pane)"
  echo "  Ctrl+C in a pane stops that process"
  echo "  'tmux kill-session -t $TMUX_SESSION' stops all"
  echo "====================================="

  # Attach — blocks until session ends
  tmux attach -t "$TMUX_SESSION"

else
  # ============================================================
  #  STANDARD MODE (original behavior)
  # ============================================================
  echo "Starting backend on :$BACKEND_PORT (db: $DB_PATH)..."
  echo "  Auction upstream: $AUCTION_UPSTREAM"
  echo "  Find upstream:    $FIND_UPSTREAM"
  "$BACKEND_DIR/server" -port "$BACKEND_PORT" -db "$DB_PATH" \
    -auction-upstream "$AUCTION_UPSTREAM" \
    -find-upstream "$FIND_UPSTREAM" &
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
  echo "  Tip: tmux not found, install with: brew install tmux"
  echo "====================================="

  # Monitor both — exit if either dies, signals break sleep
  while kill -0 "$BACKEND_PID" 2>/dev/null && kill -0 "$FRONTEND_PID" 2>/dev/null; do
    sleep 2 || break
  done
fi
