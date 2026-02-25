#!/bin/bash

# ==============================================================================
# Build & Deploy Script for House Management
# Builds backend (ARMv6) + frontend, then deploys to Raspberry Pi
# ==============================================================================

set -e

# --- Configuration ---
REMOTE_USER="kano"
REMOTE_HOST="10.147.17.116"
REMOTE_DIR="/home/kano/house_management"
SERVICE_NAME="house-backend"

# Local paths (relative to project root)
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BACKEND_BINARY="${PROJECT_ROOT}/backend/backend_pi"
FRONTEND_DIST="${PROJECT_ROOT}/frontend/dist"
ENV_FILE="${PROJECT_ROOT}/backend/.env"
SYSTEMD_SERVICE="${PROJECT_ROOT}/deploy/systemd/house-backend.service"

# SSH target shorthand
SSH_TARGET="${REMOTE_USER}@${REMOTE_HOST}"

# --- Color helpers ---
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

info()  { echo -e "${GREEN}[INFO]${NC}  $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# --- Step 1: Build backend binary for Raspberry Pi (ARMv6) ---
info "Building backend binary for ARMv6..."
cd "${PROJECT_ROOT}/backend"
GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build -o backend_pi main.go
info "Backend binary built: ${BACKEND_BINARY}"

# --- Step 2: Build frontend ---
info "Building frontend..."
cd "${PROJECT_ROOT}/frontend"
npm run build
info "Frontend built: ${FRONTEND_DIST}"

# --- Start deployment ---
info "Starting deployment to ${SSH_TARGET}:${REMOTE_DIR}"

# --- Step 1: Create remote directory if it doesn't exist ---
info "Ensuring remote directory exists..."
ssh "${SSH_TARGET}" "mkdir -p ${REMOTE_DIR}"

# --- Step 2: Stop the backend service (ignore errors if not running) ---
info "Stopping ${SERVICE_NAME} service on remote..."
ssh "${SSH_TARGET}" "sudo systemctl stop ${SERVICE_NAME} 2>/dev/null || true"

# --- Step 3: Deploy backend binary ---
info "Deploying backend binary..."
rsync -avz --progress "${BACKEND_BINARY}" "${SSH_TARGET}:${REMOTE_DIR}/backend_pi"
ssh "${SSH_TARGET}" "chmod +x ${REMOTE_DIR}/backend_pi"

# --- Step 4: Deploy frontend dist ---
info "Deploying frontend dist..."
rsync -avz --delete --progress "${FRONTEND_DIST}/" "${SSH_TARGET}:${REMOTE_DIR}/dist/"

# --- Step 5: Deploy .env file (only if it doesn't exist on remote) ---
info "Checking .env on remote..."
if ssh "${SSH_TARGET}" "[ ! -f ${REMOTE_DIR}/.env ]"; then
    if [ -f "$ENV_FILE" ]; then
        info "Deploying .env file..."
        rsync -avz "${ENV_FILE}" "${SSH_TARGET}:${REMOTE_DIR}/.env"
    else
        warn ".env file not found locally, skipping."
    fi
else
    info ".env already exists on remote, skipping (use --force-env to overwrite)."
    if echo "$@" | grep -q -- "--force-env"; then
        info "Force overwriting .env..."
        rsync -avz "${ENV_FILE}" "${SSH_TARGET}:${REMOTE_DIR}/.env"
    fi
fi

# --- Step 6: Deploy systemd service file ---
info "Deploying systemd service file..."
scp "${SYSTEMD_SERVICE}" "${SSH_TARGET}:/tmp/${SERVICE_NAME}.service"
ssh "${SSH_TARGET}" "sudo mv /tmp/${SERVICE_NAME}.service /etc/systemd/system/${SERVICE_NAME}.service && sudo systemctl daemon-reload"

# --- Step 7: Ensure uploads directory exists ---
ssh "${SSH_TARGET}" "mkdir -p ${REMOTE_DIR}/uploads"

# --- Step 8: Start the backend service ---
info "Starting ${SERVICE_NAME} service..."
ssh "${SSH_TARGET}" "sudo systemctl enable ${SERVICE_NAME} && sudo systemctl start ${SERVICE_NAME}"

# --- Step 9: Verify ---
sleep 2
info "Checking service status..."
ssh "${SSH_TARGET}" "sudo systemctl status ${SERVICE_NAME} --no-pager" || true

echo ""
info "✅ Deployment complete!"
info "   Backend:  ${SSH_TARGET}:${REMOTE_DIR}/backend_pi"
info "   Frontend: ${SSH_TARGET}:${REMOTE_DIR}/dist/"
info "   Service:  ${SERVICE_NAME}"
echo ""
info "Access the app at: http://${REMOTE_HOST}:8080"
echo ""
warn "⚠️  如果這次有改到資料庫 schema，記得 SSH 進 Pi 手動修改 DB："
warn "   1. ssh ${SSH_TARGET}"
warn "   2. cp ${REMOTE_DIR}/house.db ${REMOTE_DIR}/house.db.bak.\$(date +%Y%m%d%H%M%S)"
warn "   3. sqlite3 ${REMOTE_DIR}/house.db 'ALTER TABLE ...;'"
warn "   4. sudo systemctl restart ${SERVICE_NAME}"
