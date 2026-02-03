# Cloudflare Worker Price Probe: primitives + costs

This document describes the *current* Worker POC in `cloudflare_workers/price_probe/` and outlines the API primitives you can build on top of the Google Flights request shapes already used in this repository.

## What the current POC supports

### `GET /probe` (shopping results for a specific date)

Use when you want a “what’s available on these dates?” query.

Inputs (query string):
- `src`: 1–4 origin airport IATA codes (comma-separated)
- `dst`: 1–4 destination airport IATA codes (comma-separated)
- `date`: departure date `YYYY-MM-DD`
- `return`: return date `YYYY-MM-DD` (required for `trip_type=round_trip`)
- `trip_type`: `one_way` or `round_trip`
- `class`: `economy` | `premium_economy` | `business` | `first`
- `stops`: `any` | `nonstop` | `one_stop` | `two_stops`
- `adults`, `children`, `infants_lap`, `infants_seat`
- Market/locale/currency knobs: `hl`, `gl`, `curr`, `tz_offset_min`
- `carriers`: comma-separated airline codes and/or alliance tokens (best-effort)
- `google_host`: e.g. `www.google.com`, `www.google.de`

Output:
- `prices.sample`: a sorted list of extracted prices (best-effort)
- `prices.price_range`: sometimes available (best-effort)
- `request_cf`: Cloudflare colo/country metadata (debuggable “where did this run?”)

Caching:
- Enabled by default (`cache=0` disables)
- TTL via `cache_ttl_sec` (defaults to 300 seconds)
- Cache stores only the parsed price payload; `request_cf` reflects *the current request*.

### `GET /price-graph` (calendar graph / date range)

Use when you want “cheapest prices across a departure window” rather than a single departure date.

Endpoint:
- `GET /price-graph?src=SFO&dst=CDG&from=2026-04-01&to=2026-05-15&trip_length_days=7&trip_type=round_trip&curr=USD&hl=en-US&gl=US`

Expected output:
- rows keyed by departure date (and return date computed by trip length)
- min price per day and/or top-N offers per day

### `POST /multi-city` (segment list)

Use when the itinerary is: A→B, then C→D, etc.

Endpoint:
- `POST /multi-city` (JSON body)

Body:
- `segments`: array of `{ src, dst, date }` (2–4 segments)
- shared options: passengers, cabin, stops, currency, carriers, etc.

Notes:
- This is implemented using the same “shopping results” endpoint as `/probe` but with `TripType=MultiCity`.

## Cost model (Cloudflare-side)

Each Worker request generally does:
- 0–1 “init cookies” request to Google (cached ~1h in Worker Cache)
- 1 Google Flights “shopping results” request
- JSON parsing to extract prices

For early POCs, your Cloudflare bill will primarily be driven by:
- number of Worker invocations
- CPU time spent parsing and generating responses
- optional extras if you add Durable Objects / KV / D1, etc.

### Rough sizing (check current pricing before committing)

As of early 2026, Cloudflare’s common baseline for Workers-style metering is:
- monthly base fee (often around $5/month on “paid/standard” tiers)
- included requests (often around 10M/month)
- included CPU time (often around 30M CPU-ms/month)
- overages priced per million requests and per million CPU-ms

This Worker does very little CPU work besides parsing a few JSON lines and building a response; the hard part is upstream rate limits / reliability, not Cloudflare CPU.

### “Subrequests”

Workers fetches to upstreams (like `fetch()` to Google) are “subrequests”. They typically do **not** add per-request Worker billing, but they still matter for:
- upstream throttling/blocks (Google)
- Worker timeouts/retries you implement

## Marketplace billing (optional)

If you don’t want to build your own API keys + metering, API marketplaces can handle billing and key issuance, but they take a cut and add platform constraints.

## Operational notes

- This relies on undocumented Google endpoints; stability is not guaranteed.
- If you publish this publicly, you’ll want explicit rate limiting, caching, and abuse controls before increasing traffic.
