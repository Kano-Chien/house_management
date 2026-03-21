#!/bin/bash
# Run backend and frontend dev servers concurrently
# Uses process group so Ctrl+C kills go run's child process too

set -m  # enable job control

cleanup() {
  pkill -P $$ 2>/dev/null
  kill 0 2>/dev/null
  # Force kill anything still on our ports
  lsof -ti :5173 | xargs kill -9 2>/dev/null
  lsof -ti :8080 | xargs kill -9 2>/dev/null
  wait 2>/dev/null
}
trap cleanup EXIT INT TERM

# Kill any existing processes on ports 5173 and 8080
lsof -ti :5173 | xargs kill -9 2>/dev/null
lsof -ti :8080 | xargs kill -9 2>/dev/null
sleep 1

(cd backend && go run main.go) &
(cd frontend && npm run dev) &

wait
