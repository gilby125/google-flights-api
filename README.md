# Google Flights API Server & Library

[![Go Reference](https://pkg.go.dev/badge/github.com/gilby125/google-flights-api/flights.svg)](https://pkg.go.dev/github.com/gilby125/google-flights-api/flights)

This project provides both a Go client library (`flights` package) for the undocumented Google Flights API and a full web service that uses this library to perform flight searches, schedule jobs, and store results.

The client library produces direct requests to the Google Flights API, which is much faster than using WebDriver. The API doesn't have official documentation, so the project relies on analyzing how the [Google Flights website](https://www.google.com/travel/flights/) communicates with the backend.

The project uses [go-retryablehttp](https://github.com/hashicorp/go-retryablehttp) under the hood. Every request to the Google Flights API is retried five times in case of an error.

## Running the Full Application Service (Docker)

This section describes how to run the complete application stack (API server, databases, queue) using Docker Compose.

### Prerequisites

*   [Docker](https://docs.docker.com/get-docker/)
*   [Docker Compose](https://docs.docker.com/compose/install/)

### Configuration

The application uses environment variables for configuration. You can set these directly in your shell or create a `.env` file in the project root directory.

**Required Environment Variables:**

*   `DB_PASSWORD`: Password for the PostgreSQL database user (`flights`).
*   `NEO4J_PASSWORD`: Password for the Neo4j database user (`neo4j`).

**Optional Environment Variables (Defaults shown):**

*   `PORT=8080`: Port the API server listens on.
*   `ENVIRONMENT=development`: Application environment.
*   `WORKER_ENABLED=true`: Whether the background worker is enabled.
*   `DB_HOST=postgres`: Hostname for the Postgres service.
*   `DB_PORT=5432`: Port for the Postgres service.
*   `DB_USER=flights`: Username for Postgres.
*   `DB_NAME=flights`: Database name for Postgres.
*   `DB_SSLMODE=disable`: SSL mode for Postgres (set to `disable` for local Docker).
*   `NEO4J_URI=bolt://neo4j:7687`: URI for the Neo4j service.
*   `NEO4J_USER=neo4j`: Username for Neo4j.
*   `REDIS_HOST=redis`: Hostname for the Redis service.
*   `REDIS_PORT=6379`: Port for the Redis service.
*   `REDIS_PASSWORD=`: Password for Redis (if any).
*   `ACME_EMAIL=admin@throughfire.net`: Email for Let's Encrypt (used by Traefik).
*   Worker settings (`WORKER_CONCURRENCY`, `WORKER_MAX_RETRIES`, etc.) - see `config/config.go` for defaults.

**Example `.env` file:**

```dotenv
DB_PASSWORD=your_secure_postgres_password
NEO4J_PASSWORD=your_secure_neo4j_password
# Optional: Override other defaults if needed
# REDIS_PASSWORD=your_redis_password
# ACME_EMAIL=your_email@example.com
```

### Database Seeding

After the database containers are running, you may need to seed the database with initial data.

1.  **Wait for Services:** Ensure the `postgres` and `neo4j` services are fully started. You can check their logs:
    ```bash
    docker-compose logs postgres
    docker-compose logs neo4j
    ```
2.  **Run Bootstrap (migrations + airports seed):** The runtime image is distroless (no shell/Go toolchain), so you should use the built-in bootstrap mode:
    ```bash
    docker compose run --rm api -bootstrap
    ```
    This applies embedded PostgreSQL migrations and ensures the `airports` reference table is populated.

### Building and Running

1.  **Set Environment Variables:** Ensure `DB_PASSWORD` and `NEO4J_PASSWORD` are set in your environment or in a `.env` file.
2.  **Start Services:** Run the following command from the project root:
    ```bash
    docker-compose up -d --build
    ```
    This will build the API image and start the `api`, `postgres`, `neo4j`, and `redis` services in detached mode. Traefik will also start as a reverse proxy.

### Verification

1.  **Check Container Status:**
    ```bash
    docker-compose ps
    ```
    All services should show `Up` or `running`.
2.  **Check API Logs:**
    ```bash
    docker-compose logs api
    ```
    Look for messages indicating successful connections to Postgres, Neo4j, and Redis.
3.  **Access API:** If using Traefik locally, you might need to add `127.0.0.1 api.flights.local` to your hosts file (`/etc/hosts` on Linux/macOS, `C:\Windows\System32\drivers\etc\hosts` on Windows). Then you should be able to access the API endpoints (e.g., `http://api.flights.local/airports`). Check `api/routes.go` for available routes.
4.  **Access Neo4j Browser:** Navigate to `http://localhost:7474` in your browser. Log in with user `neo4j` and the password you set in `NEO4J_PASSWORD`.

---

### Price Graph Sweep Jobs

To catalogue the cheapest fares across large origin/destination grids without storing full offer payloads, the API now exposes a dedicated price graph sweep pipeline.

- `POST /api/v1/admin/price-graph-sweeps` enqueues a sweep. Supply origins, destinations, a departure window, optional `trip_lengths` (array of nights), and traveller preferences. The worker reuses a cached price-graph session and throttles requests; you can raise or lower the interval with `rate_limit_millis` (defaults to 750 ms between calls).
- `GET /api/v1/admin/price-graph-sweeps` lists recent sweeps together with status, error counts, and trip-length ranges.
- `GET /api/v1/admin/price-graph-sweeps/:id` returns the stored low-fare grid for a sweep; each row contains route, departure/return dates, trip length, and price snapshot.

The sweep data is stored separately in `price_graph_sweeps` and `price_graph_results`, so it will not interfere with existing bulk-search runs. When worker concurrency is high, adjust `rate_limit_millis` or the global worker concurrency settings to avoid tripping Google Flights rate limits.

---

## Using the Go Client Library

This section details how to use the `flights` package directly in your own Go projects.

### Price Graph Diagnostics (Troubleshooting)

`Session.GetPriceGraph` now returns parsed offers *and* non-fatal parsing diagnostics:

```go
offers, parseErrs, err := session.GetPriceGraph(ctx, args)
```

- `parseErrs` may be non-nil even when `err == nil` (it reports skipped/unparseable entries, date parse failures, and zero-price rows).
- When troubleshooting parsing drift, set `PRICE_GRAPH_DIAGNOSTICS=1` to emit **redacted** diagnostics logs (SHA-256 fingerprints + lengths; no raw payloads).
  - For `$0` fares specifically, diagnostics logs include `raw_price_type` / `raw_price_is_null` to help determine whether Google returned a null/0 price value vs. a parsing mismatch.

### Go protoc plugin used in the project
```
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
```

### Installation

```
go get -u github.com/gilby125/google-flights-api/flights # Note: Update path if necessary
```

### Usage

#### Session
Session is the main object that contains all the API-related functions.

**_NOTE:_** The library relies on the `GOOGLE_ABUSE_EXEMPTION` cookie (the cookie is not always needed), so if you get an unexpected HTTP status code, please go to https://www.google.com/travel/flights, do the captcha, and try once again. (The cookie is gotten from your browser database using https://github.com/browserutils/kooky)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gilby125/google-flights-api/flights" // Update path if necessary
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
)

func main() {
	session, err := flights.New()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	// Example 1: Price Graph
	fmt.Println("--- Example 1: Price Graph ---")
	priceGraphExample(session)

	// Example 2: Serialize URL
	fmt.Println("\n--- Example 2: Serialize URL ---")
	serializeURLExample(session)

	// Example 3: Get Offers
	fmt.Println("\n--- Example 3: Get Offers ---")
	getOffersExample(session)
}

func priceGraphExample(session *flights.Session) {
	offers, _, err := session.GetPriceGraph(
		context.Background(),
		flights.PriceGraphArgs{
			RangeStartDate: time.Now().AddDate(0, 0, 30),
			RangeEndDate:   time.Now().AddDate(0, 0, 60),
			TripLength:     7,
			SrcCities:      []string{"San Francisco"},
			DstCities:      []string{"New York"},
			Options:        flights.OptionsDefault(),
		},
	)
	if err != nil {
		log.Printf("Price Graph Error: %v", err)
		return
	}
	fmt.Println("Price Graph Offers (Date | Return Date | Price):")
	for _, offer := range offers {
		fmt.Printf("{%s %s %.2f} ", offer.StartDate.Format("2006-01-02"), offer.ReturnDate.Format("2006-01-02"), offer.Price)
	}
	fmt.Println()
}

func serializeURLExample(session *flights.Session) {
	url, err := session.SerializeURL(
		context.Background(),
		flights.Args{
			Date:        time.Now().AddDate(0, 0, 30),
			ReturnDate:  time.Now().AddDate(0, 0, 37),
			SrcCities:   []string{"San Diego"},
			SrcAirports: []string{"LAX"},
			DstCities:   []string{"New York", "Philadelphia"},
			Options:     flights.OptionsDefault(),
		},
	)
	if err != nil {
		log.Printf("Serialize URL Error: %v", err)
		return
	}
	fmt.Println("Serialized URL:", url)
}

func getOffersExample(session *flights.Session) {
	offers, priceRange, err := session.GetOffers(
		context.Background(),
		flights.Args{
			Date:       time.Now().AddDate(0, 0, 30),
			ReturnDate: time.Now().AddDate(0, 0, 37),
			SrcCities:  []string{"Madrid"},
			DstCities:  []string{"Estocolmo"},
			Options:    flights.Options{
				Travelers: flights.Travelers{Adults: 2},
				Currency:  currency.EUR,
				Stops:     flights.Stop1,
				Class:     flights.Economy,
				TripType:  flights.RoundTrip,
				Lang:      language.Spanish,
			},
		},
	)
	if err != nil {
		log.Printf("Get Offers Error: %v", err)
		return
	}

	if priceRange != nil {
		fmt.Printf("Price Range: Low %.2f, High %.2f\n", priceRange.Low, priceRange.High)
	}
	fmt.Println("Offers Found:")
	for i, offer := range offers {
		if i > 2 { // Limit output for brevity
			fmt.Println("...")
			break
		}
		fmt.Printf(" Offer %d: Price %.2f, Duration %s\n", i+1, offer.Price, offer.FlightDuration)
		// Print first segment details
		if len(offer.Flight) > 0 && len(offer.Flight[0]) > 0 {
			seg := offer.Flight[0][0]
			fmt.Printf("  -> Segment 1: %s %s (%s -> %s)\n", seg.AirlineName, seg.FlightNumber, seg.DepAirportCode, seg.ArrAirportCode)
		}
	}
}

```
*(Original library usage examples adapted slightly)*

### More advanced examples:
```
go run ./examples/example1/main.go
go run ./examples/example2/main.go
go run ./examples/example3/main.go
```

## Bug / Feature / Suggestion
