# Deploying to Raspberry Pi

This guide explains how to deploy the House Management backend onto a Raspberry Pi using `systemd` to keep it running permanently and optionally using a monitoring script.

## 1. Build the Go Binary for Raspberry Pi

Instead of running `go run main.go` on the Pi (which is slow and uses more memory), you should compile the backend into a standalone binary file on your PC (or on the Pi itself).

Run this on your PC (Assuming your Pi is a 64-bit OS `arm64`, or use `arm` for older 32-bit OS):
```bash
cd backend
GOOS=linux GOARCH=arm64 go build -o house-backend main.go
```

Then, copy the `house-backend` file to your Raspberry Pi:
```bash
scp house-backend pi@<raspberry-pi-ip>:/home/pi/projects/house_management/backend/
```

## 2. Setup Systemd Service (Recommended)

`systemd` is built into the Raspberry Pi OS. Using it will automatically start your backend when the Pi boots up and automatically restart it if it crashes.

1. Ensure the paths in `deploy/systemd/house-backend.service` match your Raspberry Pi setup. (Default is set for user `pi` and `/home/pi/projects/...`).
2. Copy the service file to the systemd directory on your Raspberry Pi:
   ```bash
   sudo cp deploy/systemd/house-backend.service /etc/systemd/system/
   ```
3. Reload systemd, enable the service (starts on boot), and start it:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable house-backend
   sudo systemctl start house-backend
   ```
4. Verify it's running:
   ```bash
   sudo systemctl status house-backend
   ```

## 3. (Optional) Run the Monitoring Daemon

If you want an extra layer of protection (e.g., the backend freezes but doesn't crash, making systemd unaware that it's stuck), you can run the `monitor_backend.sh` daemon. 

This script checks the application over HTTP every 10 seconds. If it stops responding, the script will execute `sudo systemctl restart house-backend`.

1. Make it executable:
   ```bash
   chmod +x deploy/scripts/monitor_backend.sh
   ```
2. You can run it manually in the background, or set it up in `crontab`, or even create a *second* systemd service just for the monitor.
   ```bash
   # Run in background with nohup
   nohup ./deploy/scripts/monitor_backend.sh &
   ```
