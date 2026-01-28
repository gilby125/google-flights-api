# API Contract Overview

This document formalizes the public contract for the Google Flights API service exposed by this repository. All request bodies and responses are JSON unless otherwise noted, and all endpoints are rooted at `/api/v1` unless flagged as legacy.

## Health & Metadata
- `GET /health`, `/health/ready`, `/health/live` return service status for liveness/readiness automation. Responses include a `status` string (`up` or `down`) and per-component details.

## Airports
- `GET /api/v1/airports`: Returns an array of airports with IATA code, city, country, latitude, and longitude. Supports pagination via `page` and `page_size` query parameters (integers; defaults: 1 and 50). Use optional `search` (substring match on name/city) for filtering.
- `GET /api/v1/airports/top`: Returns a curated list of common airport codes (for UI pickers and demos).

## Airlines
- `GET /api/v1/airlines`: Returns airline metadata; structure mirrors the airports endpoint. Same pagination and `search` query semantics apply.

## Flight Search Lifecycle
- `POST /api/v1/search`: Accepts a flight search payload (`origin`, `destination`, dates, pax counts, `trip_type`, `class`, `stops`, `currency`) and enqueues work. Responds with `202 Accepted` and `{ "id": "<search-id>" }`.
- `GET /api/v1/search/:id`: Returns the status (`pending`, `processing`, `completed`, `failed`) and, once available, normalized results for the search ID returned by the create call.
- `GET /api/v1/search`: Lists recent search requests with status and timestamps. Optional `status` filter (one of the status enums) and pagination parameters.

## Bulk Search
- `POST /api/v1/bulk-search`: Accepts expanded payloads (`origins[]`, `destinations[]`, date ranges, pax, class, stops) to schedule many itineraries. Returns `202` with a bulk search ID.
- `GET /api/v1/bulk-search/:id`: Provides run status, queue metrics, and references to completed search jobs for that bulk submission.

## Price History
- `GET /api/v1/price-history/:origin/:destination`: Returns stored price points (date, price, airline) for the route. The response is not time-bounded by default.

## Route Graph (Neo4j)
- `GET /api/v1/graph/path`: Returns up to 10 cheapest multi-hop paths between `origin` and `dest` under `maxPrice`, bounded by `maxHops` (see handler comments for defaults/limits).
- `GET /api/v1/graph/connections`: Returns reachable destinations from `origin` under `maxPrice`, bounded by `maxHops` (see handler comments for defaults/limits).
- `GET /api/v1/graph/route-stats`: Returns aggregated min/max/avg stats for `origin` → `dest` using stored price points.
- `GET /api/v1/graph/explore`: Returns route edges with coordinates for map/globe UIs. Query params: `origin` (single) or `origins` (comma-separated), plus optional `maxHops`, `maxPrice`, `dateFrom`, `dateTo`, `airlines`, `limit`, `source` (`price_point` or `route`). If `dateFrom/dateTo` are omitted, results use the best observed price across all dates.

## Admin & Operations
- `GET /api/v1/admin/jobs`: Lists scheduled jobs with cron expressions and next run times. Filtering options: `type`, `status`.
- `POST /api/v1/admin/jobs`: Creates a scheduled job. Body includes `name`, `cron`, and job template. Returns `201` with job metadata.
- `POST /api/v1/admin/jobs/:id/run|enable|disable`: Run immediately or toggle job state; success returns updated job record.
- `GET /api/v1/admin/workers` and `GET /api/v1/admin/queue`: Surface worker pool health and queue depth metrics for dashboards.
- Price graph sweeps (admin on-demand):
  - `POST /api/v1/admin/price-graph-sweeps`: Enqueues a sweep over `origins[] × destinations[] × trip_lengths[] × classes[]` for the departure date range. Provide either `class` (single) or `classes` (array) to run multiple cabins in one sweep (e.g. economy + business).
  - `GET /api/v1/admin/price-graph-sweeps`: Lists sweep runs.
  - `GET /api/v1/admin/price-graph-sweeps/:id`: Lists results for a sweep.
- Region tokens: some endpoints accept `REGION:*` items inside `origins[]`/`destinations[]` and expand them server-side. `REGION:WORLD_ALL` expands to all airports in the server’s Postgres `airports` table (currently ~3,429 Google Flights-supported airports); routes are still capped per endpoint to prevent accidental explosions.
- Continuous sweep (admin UI support):
  - `GET /api/v1/admin/continuous-sweep/status`: Returns current sweep status, including `trip_lengths` (nights).
  - `PUT /api/v1/admin/continuous-sweep/config`: Updates sweep config. Supported keys include `trip_lengths` (array of ints, 1–30), `class`, `pacing_mode`, `target_duration_hours`, `min_delay_ms`.

## Legacy Endpoints
- `/api/search` (POST) executes an immediate search without queueing; response includes raw flight offers. Reserved for internal tooling—external clients should prefer the queued endpoints.
  - Optional: add `"include_price_graph": true` to request a Google Flights calendar-style price graph alongside the offers.
  - Fixed-date mode (default): provide `"price_graph_window_days"` (2–161, default 30) to fetch a window around `"departure_date"`, using the derived trip length (return - departure).
  - Open-date mode: provide `"price_graph_departure_date_from"` + `"price_graph_departure_date_to"` (YYYY-MM-DD) and optionally `"price_graph_trip_length_days"` to fetch a flexible-date graph for that range.
- `/api/airports` and `/api/price-history` serve cached or mock data for development. They are not part of the supported contract and may be removed without notice.
