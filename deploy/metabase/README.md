# Metabase (Docker Compose)

Two ways to run Metabase:

- **Standalone (separate stack):** Metabase runs by itself. On the Komodo host it joins the existing Docker network and connects to Postgres as `postgres`.
- **Overlay (same stack):** Metabase runs alongside `docker-compose.yml` and connects to Postgres via the internal Docker network.

## Start (standalone / separate)

```sh
docker compose -f deploy/metabase/docker-compose.yml up -d
```

Metabase UI: `http://<host>:${METABASE_PORT:-3000}`

By default this binds to `127.0.0.1`. To expose it over Tailscale, set:

- `TAILSCALE_BIND_IP=100.87.196.33` (or whatever the host's Tailscale IP is)

## Start (overlay / same stack)

```sh
docker compose -f docker-compose.yml -f docker-compose.metabase.yml up -d
```

## Connect Metabase to Postgres (Flight API)

In Metabase → **Add your data** → **PostgreSQL**:

- Host:
  - Standalone (same host): `postgres`
  - Standalone (different host): `100.87.196.33` (your Komodo host Tailscale IP, if Postgres is published on it)
  - Overlay: `postgres`
- Port: `5432`
- Database name: `flights`
- Database username: `flights`
- Database password: `${DB_PASSWORD:-changeme}`
- SSL: off (if you're using `DB_SSLMODE=disable`)
