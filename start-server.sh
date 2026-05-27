#!/bin/bash
# CandyPro Server Startup Script
# Run on Linux server after uploading build artifacts

set -e

APP_DIR="/opt/candypro"
FRONTEND_DIR="$APP_DIR/frontend"
BACKEND_BIN="$APP_DIR/candypro-server"
BACKEND_PORT="${PORT:-8080}"
NUXT_PORT="${NUXT_PORT:-3000}"

echo "=== CandyPro Server Startup ==="

# Resolve public site URL (set SITE_URL env var or default to this IP)
export SITE_URL="${SITE_URL:-http://117.50.33.136:3000}"
export BACKEND_URL="${BACKEND_URL:-http://127.0.0.1:8080}"

# 1. Start Go backend
echo "[1/2] Starting Go backend on port $BACKEND_PORT..."
cd "$APP_DIR"
SEED_SUPERADMIN_EMAIL="${SEED_SUPERADMIN_EMAIL:-admin@candypro.com}" \
SEED_SUPERADMIN_PASSWORD="${SEED_SUPERADMIN_PASSWORD:-123456}" \
AUTO_SEED_DATA="${AUTO_SEED_DATA:-true}" \
  nohup "$BACKEND_BIN" > "$APP_DIR/backend.log" 2>&1 &
BACKEND_PID=$!
echo "  Backend PID: $BACKEND_PID  (log: $APP_DIR/backend.log)"

# Wait for backend to be ready
sleep 2

# 2. Start Nuxt frontend (includes API proxy to backend)
echo "[2/2] Starting Nuxt frontend on port $NUXT_PORT..."
cd "$FRONTEND_DIR"
NUXT_PUBLIC_SITE_URL="$SITE_URL" \
  nohup node server/index.mjs > "$APP_DIR/frontend.log" 2>&1 &
FRONTEND_PID=$!
echo "  Frontend PID: $FRONTEND_PID (log: $APP_DIR/frontend.log)"

echo ""
echo "=== Server started ==="
echo "  Backend:  http://127.0.0.1:$BACKEND_PORT"
echo "  Frontend: http://0.0.0.0:$NUXT_PORT"
echo ""
echo "Stop with: kill $BACKEND_PID $FRONTEND_PID"
