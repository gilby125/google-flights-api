# Metabase (Docker Compose)

Two ways to run Metabase:

- **Standalone (separate stack):** Metabase runs by itself. On the Komodo host it joins the existing Docker network and connects to Postgres as `postgres`.
- **Overlay (same stack):** Metabase runs alongside `docker-compose.yml` and connects to Postgres via the internal Docker network.

## Start (standalone / separate)

```sh
docker compose -f deploy/metabase/docker-compose.yml up -d
```

Metabase UI: `http://<server-lan-ip>:${METABASE_PORT:-3002}`

Note: on shared hosts, `3000` is often already taken (e.g. Forgejo). Consider setting `METABASE_PORT=3002`.

This standalone setup uses its own Postgres DB for Metabase's internal application data (recommended). Set:

- `METABASE_DB_PASSWORD` (recommended; default is `changeme`)
- Optional: `METABASE_DB_NAME` (default `metabase`), `METABASE_DB_USER` (default `metabase`)

By default this binds to `0.0.0.0` (LAN reachable). You can restrict or choose an interface by setting:

- `METABASE_BIND_IP=<server-lan-ip>` (LAN only) or `METABASE_BIND_IP=127.0.0.1` (host-only)
- Or `TAILSCALE_BIND_IP=100.87.196.33` (Tailscale only)

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
