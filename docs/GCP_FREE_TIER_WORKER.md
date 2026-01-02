# GCP Free Tier Worker Deployment

This guide documents how to run a **free** worker node on Google Cloud Platform's `e2-micro` instance, using a native Go binary and Tailscale side-loading (no Docker on the worker VM) to maximize performance on limited resources (2 vCPU, 1GB RAM).

## Prerequisites

- GCP Account with billing enabled
- `gcloud` CLI installed
- Tailscale Auth Key ([Generate here](https://login.tailscale.com/admin/settings/keys))
- Dokploy/Main Server Tailscale IP

---

## 1. Build the Binary (Native Linux AMD64)

Cross-compile the binary locally. This same binary (`flight-api`) serves as both the API server and the worker, depending on environment variables.

```bash
# Disable CGO for static linking
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o flight-api ./
```

## 2. Create the Free Tier VM

GCP offers one free `e2-micro` instance per month in specific US regions (`us-west1`, `us-central1`, `us-east1`).

```bash
gcloud compute instances create flights-worker \
  --machine-type=e2-micro \
  --zone=us-west1-a \
  --image-family=ubuntu-2204-lts \
  --image-project=ubuntu-os-cloud \
  --boot-disk-size=30GB \
  --tags=http-server
```

## 3. Upload Binary & Systemd Configs

Transfer the binary and the systemd unit templates to the VM.

```bash
# Upload binary
gcloud compute scp flight-api flights-worker:/tmp/ --zone=us-west1-a

# Upload systemd unit file (from this repo)
gcloud compute scp deploy/systemd/google-flights-worker@.service flights-worker:/tmp/ --zone=us-west1-a
```

## 4. Install Tailscale on the VM

SSH into the VM and install Tailscale to connect to your private Dokploy network.

```bash
gcloud compute ssh flights-worker --zone=us-west1-a
```

Inside the VM:
```bash
# Install Tailscale
curl -fsSL https://tailscale.com/install.sh | sh

# Connect to your tailnet
# Option A: Interactive (copy/paste URL)
sudo tailscale up

# Option B: Auth Key (one-liner)
sudo tailscale up --authkey=tskey-auth-YOUR_KEY
```

## 5. Configure the Worker

Set up the user, directories, and environment variables.

```bash
# Create dedicated user
sudo useradd --system --home /opt/google-flights --shell /usr/sbin/nologin flights
sudo mkdir -p /opt/google-flights /etc/google-flights

# Install binary
sudo mv /tmp/flight-api /opt/google-flights/google-flights-api
sudo chmod +x /opt/google-flights/google-flights-api
sudo chown flights:flights /opt/google-flights/google-flights-api

# Create Environment File
sudo nano /etc/google-flights/worker.env
```

**Content of `/etc/google-flights/worker.env`:**
```ini
ENVIRONMENT=production

# Worker Mode Configuration
WORKER_ENABLED=true
API_ENABLED=false
INIT_SCHEMA=false
SEED_NEO4J=false
NEO4J_ENABLED=false

# Resource Tuning (Important for e2-micro)
WORKER_CONCURRENCY=2  # 2 concurrent browser contexts
WORKER_JOB_TIMEOUT=10m

# Connection to Main Server (use Tailscale IPs)
# Run `tailscale ip -4` on your Komodo/Dokploy server to get this IP
DB_HOST=100.x.y.z
REDIS_HOST=100.x.y.z
NEO4J_URI=bolt://100.x.y.z:7687

# Credentials (same as main server)
DB_USER=flights
DB_PASSWORD=your_db_password
DB_NAME=flights
DB_SSLMODE=disable
DB_REQUIRE_SSL=false

REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password

# Log Settings
LOG_LEVEL=info
LOG_FORMAT=json
```

## 6. Enable & Start Service

Install the systemd unit and start the worker.

```bash
# Install unit file
sudo cp /tmp/google-flights-worker@.service /etc/systemd/system/
sudo systemctl daemon-reload

# Enable and start instance 1
sudo systemctl enable --now google-flights-worker@1

# Check status
systemctl status google-flights-worker@1
```

## Management Commands

```bash
# View Logs
journalctl -u google-flights-worker@1 -f

# Restart Worker
sudo systemctl restart google-flights-worker@1

# Stop Worker
sudo systemctl stop google-flights-worker@1
```

## Cost Summary

| Item | Cost |
|------|------|
| **e2-micro instance** | **Free** (744 hours/month) |
| **30GB HDD** | **Free** |
| **Outbound Data** | **Free** (first 1GB/month) |
| **Tailscale** | **Free** (Personal plan) |
