# API Contract Overview

This document formalizes the public contract for the Google Flights API service exposed by this repository. All request bodies and responses are JSON unless otherwise noted, and all endpoints are rooted at `/api/v1` unless flagged as legacy.

## Health & Metadata
- `GET /health`, `/health/ready`, `/health/live` return service status for liveness/readiness automation. Responses include a `status` string (`up` or `down`) and per-component details.

## Airports
- `GET /api/v1/airports`: Returns an array of airports with IATA code, city, country, latitude, and longitude. Supports pagination via `page` and `page_size` query parameters (integers; defaults: 1 and 50). Use optional `search` (substring match on name/city) for filtering.

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
- `GET /api/v1/price-history/:origin/:destination`: Returns price buckets (date, carrier, fare range) computed from stored results. Accepts optional `days` query (default 30) to bound lookback.

## Admin & Operations
- `GET /api/v1/admin/jobs`: Lists scheduled jobs with cron expressions and next run times. Filtering options: `type`, `status`.
- `POST /api/v1/admin/jobs`: Creates a scheduled job. Body includes `name`, `cron`, and job template. Returns `201` with job metadata.
- `POST /api/v1/admin/jobs/:id/run|enable|disable`: Run immediately or toggle job state; success returns updated job record.
- `GET /api/v1/admin/workers` and `GET /api/v1/admin/queue`: Surface worker pool health and queue depth metrics for dashboards.
- Continuous sweep (admin UI support):
  - `GET /api/v1/admin/continuous-sweep/status`: Returns current sweep status, including `trip_lengths` (nights).
  - `PUT /api/v1/admin/continuous-sweep/config`: Updates sweep config. Supported keys include `trip_lengths` (array of ints, 1–30), `class`, `pacing_mode`, `target_duration_hours`, `min_delay_ms`.

## Legacy Endpoints
- `/api/search` (POST) executes an immediate search without queueing; response includes raw flight offers. Reserved for internal tooling—external clients should prefer the queued endpoints.
- `/api/airports` and `/api/price-history` serve cached or mock data for development. They are not part of the supported contract and may be removed without notice.
