#!/bin/bash

# ==============================================================================
# Simple Go Backend Monitor Daemon Script
# Purpose: Periodically checks if the Backend (default port 8080) is alive.
#          If it's down, this script will restart it via systemctl (Recommended).
# ==============================================================================

# Set your Backend URL
TARGET_URL="http://localhost:8080/"

# Check interval in seconds
CHECK_INTERVAL=10

# Command to restart the backend (Since you are deploying on a Raspberry Pi, 
# it's best to restart the systemd service directly)
RESTART_CMD="sudo systemctl restart house-backend"

echo "Started monitoring $TARGET_URL (Checking every $CHECK_INTERVAL seconds)..."

while true; do
  # Send an HTTP GET request and fetch the HTTP status code.
  # Since the backend serves an SPA (Single Page Application) router, even hitting `/` will return 200 or related info.
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$TARGET_URL")

  # 200 (OK), 404 (Not Found, could map to SPA index route fallback), 405 (Method Not Allowed)
  if [ "$STATUS" = "200" ] || [ "$STATUS" = "404" ] || [ "$STATUS" = "405" ]; then
    # Response received, the Go server is still alive
    # echo "$(date '+%Y-%m-%d %H:%M:%S') - Backend is running (HTTP $STATUS)."
    :
  else
    echo "$(date '+%Y-%m-%d %H:%M:%S') - WARNING: Backend is not responding (HTTP $STATUS)! Attempting to restart via systemd..."
    
    # You can add functionality to send a Telegram / Line Notify alert here
    # curl -X POST -H 'Authorization: Bearer YOUR_TOKEN' -F 'message=Backend is down, restarting!' https://notify-api.line.me/api/notify
    
    # Execute the restart command
    eval "$RESTART_CMD"
    
    # Wait for it to start to avoid continuous restarts
    sleep 5
  fi
  
  sleep $CHECK_INTERVAL
done
