# Remote Workers over Tailscale (Komodo main + OCI worker)

This guide shows how to run **remote worker nodes** (example: Oracle OCI) that connect back to your **main Komodo server** over **Tailscale**, without exposing Postgres/Redis/Neo4j to the public internet.

This is a **runtime/deployment** guide. It does not require Docker CLI deploys; redeploy the stack via **Komodo**.

## What You’re Building

- **Main server (Komodo host)** runs:
  - API (optional on worker-only nodes)
  - Postgres
  - Redis
  - Neo4j
- **Remote worker (OCI)** runs:
  - `google-flights-api` binary under `systemd` (no Docker required)
  - Connects to main server over Tailscale to reach:
    - Postgres `5432`
    - Redis `6379`
    - Neo4j Bolt `7687`

## You Do NOT Need an Exit Node

You only need the machines in the same tailnet so the worker can reach the main server’s **Tailscale IP** (`100.x.y.z`).

Use an exit node only if you intentionally want the worker to route **all** internet traffic through the main server (not needed here).

## Step 1 — Install Tailscale on the Main Server (Komodo host VM)

Recommended: install Tailscale on the **Komodo host VM** (the Proxmox guest), not inside an app container.

Why: if Tailscale runs in a normal container network namespace, the `100.x` interface/IP lives inside that container and doesn’t cleanly help you bind Postgres/Redis/Neo4j on the host.

After Tailscale is running on the main server, record its Tailscale IP:

```bash
tailscale ip -4
```

## Step 2 — Bind DB/Redis/Neo4j to the Main Server’s Tailscale IP (Komodo env)

This repo’s `komodo-compose.yml` supports binding Postgres/Redis/Neo4j to a specific host IP via `TAILSCALE_BIND_IP`.

In **Komodo → Environment** for the stack, set:

- `TAILSCALE_BIND_IP=<main-server-tailscale-ip>` (example: `100.x.y.z`)

Then **redeploy via Komodo**.

What this does:
- Publishes the ports on the host **only** on the Tailscale interface (not `0.0.0.0`).
- Remote workers can connect over Tailscale, but the ports are not exposed publicly.

Notes:
- Container-to-container connections on the Komodo host still use Docker DNS names (`postgres`, `redis`, `neo4j`); they do not require host `ports:` at all.
- If you also want host-local tools (like `psql` on the host) to connect via loopback, add additional `127.0.0.1:...` port mappings. The current pattern is “bind to one IP”.

## Step 3 — Verify from the OCI Worker (Network Check)

From the OCI VM, verify it can reach the main server ports over Tailscale:

```bash
nc -vz <main-ts-ip> 5432
nc -vz <main-ts-ip> 6379
nc -vz <main-ts-ip> 7687
```

If any of these fail:
- Confirm Tailscale is up on both hosts.
- Confirm Komodo redeploy applied `TAILSCALE_BIND_IP`.
- Confirm your main server firewall isn’t blocking traffic on the Tailscale interface.

## Step 4 — Install the Worker on OCI (systemd, no Docker)

### 4.1 Build a Linux binary

Build on your dev box (or CI) and copy the binary to the OCI VM:

```bash
# amd64 (most Intel/AMD VMs)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o google-flights-api .

# arm64 (OCI Ampere A1)
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o google-flights-api .
```

Copy to OCI:

- `/opt/google-flights/google-flights-api`

### 4.2 Create `/etc/google-flights/worker.env`

On OCI:

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

Point the worker at the main server’s Tailscale IP:

- `DB_HOST=<main-ts-ip>`
- `REDIS_HOST=<main-ts-ip>`
- `NEO4J_URI=bolt://<main-ts-ip>:7687`

And set passwords to match the main server:

- `DB_PASSWORD=...`
- `REDIS_PASSWORD=...`
- `NEO4J_PASSWORD=...`

### 4.3 Install + start systemd unit

From a repo checkout on OCI:

```bash
sudo ./deploy/systemd/install.sh --instances 1
```

View logs:

```bash
journalctl -u google-flights-worker@1 -f
```

### 4.4 Scale up

To run more workers on the same OCI VM:

```bash
sudo ./deploy/systemd/install.sh --instances 1 2 3
```

Each instance sets a distinct `WORKER_ID=worker-%i`.

## Common Gotchas

- **“password authentication failed”**: your DB user password is stored inside the Postgres data directory volume. Changing env vars does not change the existing user’s password; either ALTER USER or wipe the pgdata volume.
- **Workers can’t connect but API works locally**: the API talks to `postgres` over Docker networking; the worker needs host port publishing + a reachable host IP (Tailscale).
- **Schema errors like “relation does not exist”**: ensure the main stack has schema initialization enabled (default is now safe/idempotent) and that you’re running the latest image.

