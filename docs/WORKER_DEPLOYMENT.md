# Worker Deployment Guide

Deploy workers to any cloud provider to distribute flight search load.

This repo supports two worker deployment patterns:
- Recommended: **remote workers over Tailscale + CI/CD binaries** (no Docker, automated builds)
- Alternative: **Docker worker with Tailscale sidecar**
- Alternative: **Docker worker managed by systemd**

If you already have Tailscale installed on the **worker VM itself**, you can use plain Docker Compose on the worker (Dokploy/Compose/etc.) and point `DB_HOST`/`REDIS_HOST`/`NEO4J_URI` at the main server’s Tailscale IP. You do not need a Tailscale sidecar container in that case.

If you run your “main” stack under **Komodo**, redeploy via **Komodo** (do not use Docker CLI deploys).

## Komodo: build-on-server (no waiting for GHCR)

If you deploy from a git repo in Komodo, you can build the image on the Komodo server during deploy.

In `komodo-compose.yml`, the `api` service is configured with `build:` and `pull_policy: build` so Komodo will build from source.

Recommended Komodo env:
- `IMAGE_TAG=local` (tags the locally-built image; no registry push required)
- `INIT_SCHEMA=true` (safe/idempotent; ensures tables exist after volume wipes)

## Remote Workers over Tailscale (Komodo main + OCI worker)

Goal: run remote worker nodes (example: Oracle OCI) that connect back to your main Komodo server over Tailscale, without exposing Postgres/Redis/Neo4j to the public internet.

### You do NOT need an exit node

You only need the machines in the same tailnet so the worker can reach the main server’s Tailscale IP (`100.x.y.z`).

### Step 1 — Install Tailscale on the main server

Install Tailscale on the **Komodo host VM** (the Proxmox guest), not inside an app container.

After Tailscale is running, record the main server’s Tailscale IP:

```bash
tailscale ip -4
```

### Step 2 — Bind Postgres/Redis/Neo4j to the Tailscale IP (Komodo env)

Remote workers must be able to reach your main server’s:
- Postgres `5432`
- Redis `6379`
- Neo4j Bolt `7687`

If your main stack runs under Komodo, publish these ports **only** on the Tailscale interface by setting, in **Komodo → Environment**:

- `TAILSCALE_BIND_IP=<main-server-tailscale-ip>` (example: `100.x.y.z`)

Then **redeploy via Komodo**.

Notes:
- Container-to-container connections on the Komodo host still use Docker DNS names (`postgres`, `redis`, `neo4j`). They do not require host `ports:` at all.
- The host `ports:` bindings are only for *remote machines* (OCI workers) to connect over Tailscale.

### Step 3 — Verify from the OCI worker (network check)

From the OCI VM:

```bash
nc -vz <main-ts-ip> 5432
nc -vz <main-ts-ip> 6379
nc -vz <main-ts-ip> 7687
```

### Step 4 — Install the worker on Remote VM (systemd, no Docker)

#### Option 1: CI/CD Download (Recommended)
Our GitHub Actions pipeline automatically builds and releases binaries for every push to `main`.

```bash
# On the VM (amd64 for GCP/Azure, arm64 for OCI)
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
  URL="https://github.com/gilby125/google-flights-api/releases/latest/download/flight-api-linux-arm64"
else
  URL="https://github.com/gilby125/google-flights-api/releases/latest/download/flight-api-linux-amd64"
fi

curl -L -o /tmp/google-flights-api "$URL"
chmod +x /tmp/google-flights-api
```

#### Option 2: Manual local build
1) Build a Linux binary and copy to Remote VM as `/opt/google-flights/google-flights-api`:

```bash
# amd64 (most Intel/AMD VMs)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o google-flights-api .

# arm64 (OCI Ampere A1)
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o google-flights-api .
```

2) Create `/etc/google-flights/worker.env`:

```bash
sudo mkdir -p /etc/google-flights /opt/google-flights
sudo cp deploy/systemd/worker.env.example /etc/google-flights/worker.env
sudo nano /etc/google-flights/worker.env
```

Set at minimum:
- `ENVIRONMENT=production`
- `API_ENABLED=false`
- `WORKER_ENABLED=true`
- `INIT_SCHEMA=false`
- `SEED_NEO4J=false`
- `DB_HOST=<main-ts-ip>`
- `REDIS_HOST=<main-ts-ip>`
- `NEO4J_URI=bolt://<main-ts-ip>:7687`
- `DB_PASSWORD=...` (same as main server)
- `REDIS_PASSWORD=...` (same as main server)
- `NEO4J_PASSWORD=...` (same as main server, if you use Neo4j features)

3) Install + start systemd units:

```bash
sudo ./deploy/systemd/install.sh --instances 1
journalctl -u google-flights-worker@1 -f
```

## Option B: Docker worker managed by systemd (no scripts / no binary copy)

Use this if you want the worker VM to just run a container image (pulled from GHCR) and have `systemd` keep it running.

Recommendation:
- Use the **built image** directly (pull from GHCR) rather than trying to “run the binary from Docker”.
- If you truly want a native binary on the VM, use **Option A** (copy a binary + systemd) instead of extracting it from an image.

On the worker VM:

1) Create env file:

```bash
sudo mkdir -p /etc/google-flights
sudo cp deploy/systemd/worker.docker.env.example /etc/google-flights/worker.docker.env
sudo nano /etc/google-flights/worker.docker.env
```

Set:
- `WORKER_IMAGE=ghcr.io/gilby125/flight-api:latest` (or a pinned SHA tag)
- `DB_HOST`, `REDIS_HOST`, `NEO4J_URI` to your main server’s Tailscale IP
- `DB_PASSWORD`, `REDIS_PASSWORD`, `NEO4J_PASSWORD`

2) Install + start unit:

```bash
sudo cp deploy/systemd/google-flights-worker-docker@.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now google-flights-worker-docker@1
journalctl -u google-flights-worker-docker@1 -f
```

## Option A: systemd (no Docker)

Use this if you want the simplest reliable setup on a VPS (compiled Go binary + `systemd`).

Important:
- Remote workers must be able to reach **Postgres (5432)**, **Redis (6379)**, and **Neo4j Bolt (7687)** on your main server.
- If your main stack runs in Docker (Komodo), you must publish those ports on the main server **to the Tailscale interface only** (not public internet).
  - In `komodo-compose.yml`, set `TAILSCALE_BIND_IP` to your main server's Tailscale IP (e.g. `100.x.y.z`) and redeploy via Komodo.

1. Build the binary (recommended: in CI) and copy it to the VPS as `/opt/google-flights/google-flights-api`
2. Copy `deploy/systemd/worker.env.example` to `/etc/google-flights/worker.env` and fill in required values
3. Copy `deploy/systemd/google-flights-worker.service` (or `google-flights-worker@.service`) to `/etc/systemd/system/`
4. Enable + start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now google-flights-worker
journalctl -u google-flights-worker -f
```

Recommended worker env settings:
- `API_ENABLED=false` (avoid port conflicts; run many workers per host)
- `INIT_SCHEMA=false` and `SEED_NEO4J=false` (run schema/seed on your main server only)

---

## Quick Start (3 Steps)

```bash
# 1. Copy files to your VPS
scp docker-compose.worker.tailscale.yml .env.worker.tailscale.example user@your-server:~/

# 2. Configure (on the VPS)
ssh user@your-server
cp .env.worker.tailscale.example .env
nano .env  # Set: TS_AUTHKEY, TS_MAIN_SERVER_IP, DB_PASSWORD, REDIS_PASSWORD

# 3. Start worker
docker compose -f docker-compose.worker.tailscale.yml up -d
```

## Dokploy / Compose on OCI (Tailscale already installed on the VM)

If the OCI VM already has Tailscale installed/running, you can deploy a single `worker` service using `docker-compose.worker.yml` and set:
- The “core flags” alone are not enough. The worker must be able to connect to your **main server Postgres + Redis**, so you must provide DB/Redis credentials too.

Required env vars:
- `DB_HOST=<main-ts-ip>`
- `DB_PASSWORD=<same as main>`
- `REDIS_HOST=<main-ts-ip>`
- `REDIS_PASSWORD=<same as main>`

Recommended env vars:
- `WORKER_ID=worker-oci-1`
- `ENVIRONMENT=production`
- `API_ENABLED=false`
- `WORKER_ENABLED=true`
- `INIT_SCHEMA=false`
- `SEED_NEO4J=false`
- `DB_SSLMODE=disable` and `DB_REQUIRE_SSL=false` (Tailscale encrypts the transport)

Optional (Neo4j):
- If you need Neo4j-backed features on the worker, set `NEO4J_ENABLED=true` and `NEO4J_URI=bolt://<main-ts-ip>:7687` and `NEO4J_PASSWORD=<same as main>`.

Then start it as your platform expects (Dokploy UI or `docker compose -f docker-compose.worker.yml up -d`).

If your worker VM only has 1 vCPU and you see an error like “range of CPUs is from 0.01 to 1.00”, set:
- `WORKER_CPU_LIMIT=1.0` (or lower, e.g. `0.9`)

---

## Required Values

| Variable | How to Get |
|----------|------------|
| `TS_AUTHKEY` | [Tailscale Admin](https://login.tailscale.com/admin/settings/keys) → Generate reusable key |
| `TS_MAIN_SERVER_IP` | On main server: `tailscale ip -4` |
| `DB_PASSWORD` | Same as main server's `DB_PASSWORD` |
| `REDIS_PASSWORD` | Same as main server's `REDIS_PASSWORD` |

---

## Provider Guides

### Oracle Cloud Free Tier (ARM64)

1. Create Always Free Ampere A1 instance (4 OCPU / 24GB RAM)
2. Open Security List: allow outbound TCP (Tailscale handles the rest)
3. SSH in and follow Quick Start above

### AWS EC2 / Graviton

```bash
# Use Graviton for ARM64 (cheaper)
aws ec2 run-instances --instance-type t4g.micro --image-id ami-ubuntu-arm64
```

### Hetzner Cloud

```bash
```bash
hcloud server create --type cax11 --image ubuntu-22.04 --name worker-1
```

### Google Cloud Platform (Free Tier)

See the dedicated guide: [GCP Free Tier Worker Setup](GCP_FREE_TIER_WORKER.md)


---

## Deployment Automation

Use the provided script to deploy binaries and configuration updates to all workers:

### Basic Usage

```bash
# Deploy latest binary and restart all workers
./scripts/deploy-workers.sh

# Only update binary (don't restart)
./scripts/deploy-workers.sh --binary-only

# Only restart workers (e.g., after manual config change)
./scripts/deploy-workers.sh --restart-only

# Update environment variable and restart
./scripts/deploy-workers.sh --env LOG_LEVEL=debug
./scripts/deploy-workers.sh --env LOG_LEVEL=info --env WORKER_CONCURRENCY=8
```

### How It Works

The script automatically:
1. **Tries Tailscale SSH first** - Direct private network access
2. **Falls back to cloud CLIs** if Tailscale unavailable:
   - **GCP**: Uses IAP tunnel (`gcloud compute ssh --tunnel-through-iap`)
   - **Azure**: Uses run-command API (`az vm run-command invoke`)
3. Downloads the correct binary (amd64/arm64) from GitHub Releases
4. Stops the worker, installs the binary, and restarts

### Requirements

- `gcloud` CLI (for GCP workers)
- `az` CLI (for Azure workers)
- Tailscale installed (optional, but recommended)

### Adding More Workers

Edit `scripts/deploy-workers.sh` and add to the `WORKERS` array:

```bash
WORKERS[my-worker-name]="gcp:vm-name:zone"
WORKERS[my-worker-name]="azure:vm-name"
```

---

## Security Hardening

To protect your workers, we recommend using **Tailscale SSH** and disabling all public inbound ports (including 22).

### 1. Enable Tailscale SSH
On the worker VM:
```bash
sudo tailscale up --ssh --accept-routes
```

### 2. Lock Down Firewalls
Once Tailscale SSH is verified, delete the public SSH access rules.

**GCP:**
```bash
gcloud compute firewall-rules delete default-allow-ssh
gcloud compute firewall-rules delete default-allow-rdp # Highly recommended
```

**Azure:**
```bash
az network nsg rule delete --nsg-name flights-workerNSG --resource-group YOUR_RG --name default-allow-ssh
```

### 3. SSH into Workers
You no longer need public IPs. SSH directly via Tailscale:
```bash
ssh flights-worker
ssh azure-worker
```

---

## Verify Worker is Connected

```bash
# Check Tailscale status
docker compose -f docker-compose.worker.tailscale.yml exec tailscale tailscale status

# Check worker logs
docker compose -f docker-compose.worker.tailscale.yml logs worker

# Should see: "Tailscale connected! Starting worker..."
```

---

## Scaling Workers

Spin up more VMs with unique `WORKER_ID` values:

```bash
WORKER_ID=worker-oracle-2 docker compose -f docker-compose.worker.tailscale.yml up -d
```

Workers auto-register via Redis and participate in leader election for scheduling.
