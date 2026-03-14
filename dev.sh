#!/bin/bash
# Run backend and frontend dev servers concurrently
# Uses process group so Ctrl+C kills go run's child process too

set -m  # enable job control

cleanup() {
  pkill -P $$ 2>/dev/null
  kill 0 2>/dev/null
  wait 2>/dev/null
}
trap cleanup EXIT INT TERM

(cd backend && go run main.go) &
(cd frontend && npm run dev) &

wait
