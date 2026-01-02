# Distributed Worker Deployment Guide
# ===================================
# Deploy workers across multiple servers to spread the scraping load

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│                        MAIN SERVER (Dokploy)                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
│  │ PostgreSQL  │  │    Redis    │  │   Neo4j     │  │   API       │ │
│  │   :5432     │  │    :6379    │  │   :7687     │  │   :8080     │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └─────────────┘ │
│         │                │                │                          │
│         └────────────────┼────────────────┘                          │
│                          │                                           │
│                   ┌──────┴──────┐                                    │
│                   │  Tailscale  │  ← OR Cloudflare Tunnel            │
│                   │  100.x.x.x  │                                    │
│                   └──────┬──────┘                                    │
└──────────────────────────┼───────────────────────────────────────────┘
                           │
            ┌──────────────┼──────────────┬──────────────┐
            │              │              │              │
     ┌──────┴──────┐ ┌─────┴──────┐ ┌─────┴──────┐ ┌─────┴──────┐
     │  Worker 1   │ │  Worker 2  │ │  Worker 3  │ │  Worker 4  │
     │  US-East    │ │  US-West   │ │  EU        │ │  APAC      │
     │  Tailscale  │ │  Tailscale │ │  Tailscale │ │  Tailscale │
     └─────────────┘ └────────────┘ └────────────┘ └────────────┘
```

## Option 1: Tailscale (Recommended)

**Pros:**
- Zero-config mesh VPN
- Works behind NAT/firewalls
- MagicDNS for easy hostname resolution
- Free for up to 100 devices

### Setup on Main Server

```bash
# Install Tailscale
curl -fsSL https://tailscale.com/install.sh | sh

# Authenticate
sudo tailscale up

# Note your Tailscale IP (e.g., 100.64.1.1)
tailscale ip -4
```

### Setup on Worker Servers

```bash
# Install Tailscale
curl -fsSL https://tailscale.com/install.sh | sh

# Authenticate with same account
sudo tailscale up

# Verify connectivity to main server
ping 100.64.1.1  # or use MagicDNS: ping main-server
```

### Docker Compose for Tailscale Workers

See `docker-compose.worker.tailscale.yml`

---

## Option 2: Cloudflare Tunnel

**Pros:**
- No open ports on main server
- Built-in DDoS protection
- Access policies and authentication

**Cons:**
- Requires Cloudflare account
- Higher latency for non-HTTP (Redis/Postgres use TCP tunnels)
- Free tier has limits

### Setup on Main Server

```bash
# Install cloudflared
curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o cloudflared
chmod +x cloudflared
sudo mv cloudflared /usr/local/bin/

# Authenticate
cloudflared tunnel login

# Create tunnel
cloudflared tunnel create flights-infra

# Configure tunnel (save tunnel ID)
cat > ~/.cloudflared/config.yml <<EOF
tunnel: <YOUR_TUNNEL_ID>
credentials-file: /root/.cloudflared/<YOUR_TUNNEL_ID>.json

ingress:
  - hostname: redis.flights.internal
    service: tcp://localhost:6379
  - hostname: postgres.flights.internal  
    service: tcp://localhost:5432
  - service: http_status:404
EOF

# Run tunnel
cloudflared tunnel run flights-infra
```

### Worker Connection via Cloudflare Access

Workers use `cloudflared access tcp` to connect:

```bash
# On worker, connect to Redis through tunnel
cloudflared access tcp --hostname redis.flights.internal --url localhost:6379
```

---

## Recommendation: Tailscale

For your use case, **Tailscale is the better choice** because:

| Factor | Tailscale | Cloudflare Tunnel |
|--------|-----------|-------------------|
| Setup complexity | Very low | Medium |
| Latency for TCP | Low (direct mesh) | Higher (proxied) |
| Redis/Postgres support | Native TCP | Requires TCP tunnel |
| Cost | Free (100 devices) | Free with limits |
| Works behind NAT | Yes | Yes |

---

## Quick Start with Tailscale

### 1. Main Server Setup
```bash
# Already running your stack? Just add Tailscale
curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up
echo "Main server Tailscale IP: $(tailscale ip -4)"
```

### 2. Worker Server Setup
```bash
# On each remote worker server
curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up

# Copy files
scp your-main-server:~/google-flights-api/docker-compose.worker.tailscale.yml .
scp your-main-server:~/google-flights-api/.env.worker.example .env.worker

# Edit .env.worker - use Tailscale IPs!
# DB_HOST=100.64.1.1   (your main server's Tailscale IP)
# REDIS_HOST=100.64.1.1

# Start worker
docker compose -f docker-compose.worker.tailscale.yml up -d
```

### 3. Verify Connectivity
```bash
# From worker, test Redis
redis-cli -h 100.64.1.1 -a your-password PING

# Test PostgreSQL
psql -h 100.64.1.1 -U flights -d flights -c "SELECT 1"
```
