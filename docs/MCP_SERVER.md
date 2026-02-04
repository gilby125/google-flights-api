# Google Flights MCP Server

This repo includes an **MCP (Model Context Protocol) server** implemented in Go that exposes a small set of tools over stdio, backed by the `flights` client library.

## Build

```bash
go build -o mcp-server ./cmd/mcp-server
```

## Run

The MCP server communicates over stdio, so you typically run it via an MCP client.

### Claude Desktop example

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "google-flights": {
      "command": "/absolute/path/to/google-flights-api/mcp-server"
    }
  }
}
```

## Tools

### `search_flights`

Searches Google Flights for **one-way**, **round-trip**, or **multi-city** itineraries.

Arguments:
- `trip_type` (string, optional): `one_way`, `round_trip`, or `multi_city`. If omitted, defaults to `round_trip` when `return_date` is provided, else `one_way`.
- `origin` (string): IATA origin airport code (e.g., `SFO`). Required for `one_way` and `round_trip`.
- `destination` (string): IATA destination airport code (e.g., `JFK`). Required for `one_way` and `round_trip`.
- `date` (string): Departure date `YYYY-MM-DD`. Required for `one_way` and `round_trip`.
- `return_date` (string, optional): Return date `YYYY-MM-DD` (round-trip only).
- `segments` (string): JSON array of segments for multi-city. Required for `trip_type=multi_city`.
  - Example: `[{"origin":"SFO","destination":"JFK","date":"2026-06-01"},{"origin":"JFK","destination":"LHR","date":"2026-06-05"}]`
- `adults` (number, optional): Number of adults (default `1`).
- `currency` (string, optional): ISO currency code (default `USD`).
- `carriers` (string, optional): Comma-separated IATA airline codes and/or alliance tokens (best-effort). Example: `UA,DL` or `STAR_ALLIANCE`.

Response:
- `offers`: array of offers (price + flights).
- `price_range`: best-effort price range summary from the scraper.
- `search_url`: a Google Flights URL representing the search (best-effort; may be empty if serialization fails).

### `get_price_graph`

Fetches a **price graph** (calendar fares) across a departure date range for a fixed trip length (round-trip semantics).

Arguments:
- `origin` (string, required): IATA origin airport code.
- `destination` (string, required): IATA destination airport code.
- `range_start_date` (string, required): `YYYY-MM-DD`.
- `range_end_date` (string, required): `YYYY-MM-DD`.
- `trip_length` (number, optional): Trip length in days (default `7`).
- `currency` (string, optional): ISO currency code (default `USD`).
- `carriers` (string, optional): Comma-separated carrier tokens (best-effort).

Response:
- `offers`: array of `{start_date, return_date, price, currency}` rows.
- `best_price`: minimum observed price (0 means “none found”).
- `best_price_airline`: best-effort inferred airline from sampling a single full offer for the cheapest date pair.

### `search_hotels`

Searches Google Hotels for a location + date range.

Arguments:
- `location` (string, required): City/region query (e.g., `Paris`, `San Francisco`).
- `checkin_date` (string, required): `YYYY-MM-DD`.
- `checkout_date` (string, required): `YYYY-MM-DD`.
- `adults` (number, optional): Number of adults (default `1`).
- `children` (number, optional): Number of children (default `0`).
- `currency` (string, optional): ISO currency code (default `USD`).
- `lang` (string, optional): BCP-47 language tag (default `en`).

Response:
- `offers`: array of hotel results (best-effort parsed).
- `count`: number of parsed offers.
- `search_url`: a Google Hotels URL representing the search (best-effort; may be empty if serialization fails).
