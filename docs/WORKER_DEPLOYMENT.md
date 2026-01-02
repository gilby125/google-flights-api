# Worker Deployment Guide

Deploy workers to any cloud provider to distribute flight search load.

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
hcloud server create --type cax11 --image ubuntu-22.04 --name worker-1
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
