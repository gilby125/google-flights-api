# Cloudflare Worker POC: Google Flights price probe

Goal: deploy a simple Worker that queries Google Flights from the Worker’s **egress IP** and returns a small price summary plus `request.cf` metadata (colo/country). This lets you test whether prices differ by **region/locale/currency** and/or by where the request originates.

## Setup

```bash
cd cloudflare_workers/price_probe
npm install
```

## Run locally

```bash
npm run dev
```

Example:

```bash
curl "http://127.0.0.1:8787/probe?src=SFO&dst=CDG&date=2026-04-01&return=2026-04-08&trip_type=round_trip&hl=en-US&gl=US&curr=USD"
```

Price graph / calendar (date range):

```bash
curl "http://127.0.0.1:8787/price-graph?src=SFO&dst=CDG&from=2026-04-01&to=2026-04-10&trip_length_days=7&trip_type=round_trip&hl=en-US&gl=US&curr=USD"
```

Multi-city:

```bash
curl -X POST "http://127.0.0.1:8787/multi-city" -H "content-type: application/json" --data '{"segments":[{"src":"SFO","dst":"CDG","date":"2026-04-01"},{"src":"CDG","dst":"SFO","date":"2026-04-08"}],"hl":"en-US","gl":"US","curr":"USD"}'
```

Try a different market/locale/currency:

```bash
curl "http://127.0.0.1:8787/probe?src=SFO&dst=CDG&date=2026-04-01&return=2026-04-08&trip_type=round_trip&hl=de-DE&gl=DE&curr=EUR&tz_offset_min=60"
```

## Deploy

```bash
npm run deploy
```

### “Pick an exit node” (closest approximation)

You can’t select an arbitrary Cloudflare colo/egress IP **per request**. Workers run where Cloudflare decides, based on routing and configuration.

For this POC, the closest approximation is to deploy multiple Worker variants with **placement** pinned near different cloud regions, then call the corresponding deployment:

```bash
# US East-ish egress
npm run deploy -- -e us_east

# EU West-ish egress
npm run deploy -- -e eu_west

# AP Singapore-ish egress
npm run deploy -- -e ap_singapore
```

Each response includes `request.cf.colo` / `request.cf.country` so you can confirm where Cloudflare says the request ran.

## Query params

- `src` (required): IATA airport code(s), comma-separated (e.g. `SFO` or `SFO,LAX`) (max 4)
- `dst` (required): IATA airport code(s), comma-separated (max 4)
- `date` (required): `YYYY-MM-DD` departure date
- `return` (optional): `YYYY-MM-DD` return date (required if `trip_type=round_trip`)
- `trip_type` (optional): `one_way` (default) or `round_trip`
- `stops` (optional): `any` (default), `nonstop`, `one_stop`, `two_stops`
- `adults` (optional): default `1`
- `children` / `infants_lap` / `infants_seat` (optional): default `0`
- `hl` (optional): locale like `en-US` (default `en-US`)
- `gl` (optional): market/country like `US` (default `US`)
- `curr` (optional): ISO currency like `USD` (default `USD`)
- `tz_offset_min` (optional): minutes offset used in a Google header (default `-120`)
- `top` (optional): how many prices to return (default `10`, max `50`)
- `google_host` (optional): e.g. `www.google.com`, `www.google.de` (default `www.google.com`)
- `debug` (optional): `1` to include extra debug info (truncated response)
- `cache` (optional): `0` disables Worker caching (default enabled)
- `cache_ttl_sec` (optional): cache TTL seconds (default `300`, max `3600`)

## Notes / gotchas

- This is a best-effort POC against an **undocumented** Google endpoint and may break as Google changes internals.
- Expect intermittent `429` / `503` rate limits. Keep request volume low.
- If you consistently get non-200 responses, you may need to supply a `GOOGLE_ABUSE_EXEMPTION` cookie as a Worker secret:
  - `wrangler secret put GOOGLE_ABUSE_EXEMPTION`
  - Value can be either `GOOGLE_ABUSE_EXEMPTION=...` or just the cookie value (the Worker will prefix it).
