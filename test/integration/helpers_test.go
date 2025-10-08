package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"github.com/gilby125/google-flights-api/api"
	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/worker"
	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	if os.Getenv("ENABLE_INTEGRATION_TESTS") != "1" {
		fmt.Println("Skipping integration tests (set ENABLE_INTEGRATION_TESTS=1 to enable)")
		os.Exit(0)
	}

	// Load .env file first
	_ = godotenv.Load() // Load .env first, but we will override DB settings

	// Explicitly set the password for the test run BEFORE loading config
	os.Setenv("DB_PASSWORD", "Lokifish123")

	// Override specific env vars for integration tests to connect to Docker services via exposed ports
	os.Setenv("DB_HOST", "localhost") // Connect to exposed port on host
	os.Setenv("DB_PORT", "5432")      // Default port
	os.Setenv("DB_USER", "flights")   // Match docker-compose
	// os.Setenv("DB_PASSWORD", os.Getenv("DB_PASSWORD")) // No longer needed, set explicitly above
	os.Setenv("DB_NAME_TEST", "flights")            // Match docker-compose and test config
	os.Setenv("DB_SSLMODE", "disable")              // Usually disable SSL for local docker testing
	os.Setenv("NEO4J_URI", "bolt://localhost:7687") // Revert back to bolt:// scheme
	os.Setenv("NEO4J_USER", "neo4j")                // Match docker-compose
	os.Setenv("NEO4J_PASSWORD", "Lokifish123")      // Match .env file
	os.Setenv("REDIS_HOST", "localhost")            // Connect to exposed port on host
	os.Setenv("REDIS_PORT", "6379")                 // Default port

	// Setup test environment using the (potentially overridden) env vars
	cfg := config.LoadTestConfig() // Load config using potentially overridden env vars

	// cfg now holds the config based on overridden env vars

	// Add a small delay to allow DB container to fully initialize after potential recreation
	time.Sleep(3 * time.Second)

	// Initialize test database
	pgPool := db.InitTestPostgres(cfg)

	// Schema creation is now handled by db.InitSchema() called from main.go
	// TestMain should focus on setting up connections and running tests.
	// We might need a way to ensure the DB is clean before TestMain runs,
	// perhaps by dropping/recreating the DB in docker-compose or via a setup script.
	// For now, assume InitSchema will handle table creation correctly on app start.

	defer pgPool.Close()

	// Run migrations
	if err := db.RunMigrations(fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.PostgresConfig.User,
		cfg.PostgresConfig.Password,
		cfg.PostgresConfig.Host,
		cfg.PostgresConfig.Port,
		cfg.PostgresConfig.DBName,
		cfg.PostgresConfig.SSLMode,
	)); err != nil {
		panic(fmt.Sprintf("Failed to run migrations: %v", err))
	}

	// Start test server
	gin.SetMode(gin.TestMode)
	router := gin.New()
	pgDB, err := db.NewPostgresDB(cfg.PostgresConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Postgres DB: %v", err))
	}
	neo4jDB, err := db.NewNeo4jDB(cfg.Neo4jConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Neo4j DB: %v", err))
	}
	q, err := queue.NewRedisQueue(cfg.RedisConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Redis queue: %v", err))
	}
	wm := worker.NewManager(q, pgDB, neo4jDB, cfg.WorkerConfig)
	api.RegisterRoutes(router, pgDB, neo4jDB, q, wm, cfg)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Set environment variable for test server URL
	os.Setenv("TEST_SERVER_URL", testServer.URL)
	defer os.Unsetenv("TEST_SERVER_URL")

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// createTestRequest creates a new HTTP request with context
func createTestRequest(t *testing.T, method, path string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, nil)
	ctx := context.WithValue(req.Context(), "test", t)
	return req.WithContext(ctx), httptest.NewRecorder()
}

// AuthenticatedRequest creates a request with test auth headers
func AuthenticatedRequest(t *testing.T, method, path string) (*http.Request, *httptest.ResponseRecorder) {
	req, rr := createTestRequest(t, method, path)
	req.Header.Set("Authorization", "Bearer test-token")
	return req, rr
}

// ParseJSONResponse helper to decode JSON responses
func ParseJSONResponse(t *testing.T, rr *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(rr.Body).Decode(v); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
}

// CleanupDB resets database state between tests
func CleanupDB(t *testing.T) {
	t.Helper()
	_, err := db.GetPool().Exec(context.Background(), `
		TRUNCATE TABLE flight_segments, flight_offers, search_queries, users, sessions RESTART IDENTITY CASCADE;
	`)
	if err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}
}

// seedFlightOffer inserts a flight offer for testing
func seedFlightOffer(t *testing.T, id int, searchQueryID int, price float64, currency string) {
	t.Helper()
	_, err := db.GetPool().Exec(context.Background(), `
		INSERT INTO flight_offers (id, search_query_id, price, currency, departure_date, return_date, total_duration)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			search_query_id = EXCLUDED.search_query_id,
			price = EXCLUDED.price,
			currency = EXCLUDED.currency,
			departure_date = EXCLUDED.departure_date,
			return_date = EXCLUDED.return_date,
			total_duration = EXCLUDED.total_duration;
	`, id, searchQueryID, price, currency, time.Now().AddDate(0, 0, 7), time.Now().AddDate(0, 0, 14), 36000) // Example dates/duration
	if err != nil {
		t.Fatalf("Failed to seed flight offer %d for search %d: %v", id, searchQueryID, err)
	}
}

// seedFlightSegment inserts a flight segment for testing, including airplane and legroom
func seedFlightSegment(t *testing.T, id int, flightOfferID int, airlineCode, depAirport, arrAirport, airplane, legroom string, isReturn bool) {
	t.Helper()
	_, err := db.GetPool().Exec(context.Background(), `
		INSERT INTO flight_segments (id, flight_offer_id, airline_code, flight_number, departure_airport, arrival_airport, departure_time, arrival_time, duration, airplane, legroom, is_return)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			flight_offer_id = EXCLUDED.flight_offer_id,
			airline_code = EXCLUDED.airline_code,
			flight_number = EXCLUDED.flight_number,
			departure_airport = EXCLUDED.departure_airport,
			arrival_airport = EXCLUDED.arrival_airport,
			departure_time = EXCLUDED.departure_time,
			arrival_time = EXCLUDED.arrival_time,
			duration = EXCLUDED.duration,
			airplane = EXCLUDED.airplane,
			legroom = EXCLUDED.legroom,
			is_return = EXCLUDED.is_return;
	`, id, flightOfferID, airlineCode, fmt.Sprintf("%s%d", airlineCode, id*100), depAirport, arrAirport, time.Now().Add(time.Hour), time.Now().Add(5*time.Hour), 14400, airplane, legroom, isReturn) // Example times/duration
	if err != nil {
		t.Fatalf("Failed to seed flight segment %d for offer %d: %v", id, flightOfferID, err)
	}
}

// seedSearchQueryFixedDates inserts a basic search query with specific dates
func seedSearchQueryFixedDates(t *testing.T, id int, origin, destination string, status string, departureDate, returnDate time.Time) {
	t.Helper()
	_, err := db.GetPool().Exec(context.Background(), `
		INSERT INTO search_queries (id, origin, destination, departure_date, return_date, adults, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			origin = EXCLUDED.origin,
			destination = EXCLUDED.destination,
			departure_date = EXCLUDED.departure_date,
			return_date = EXCLUDED.return_date,
			adults = EXCLUDED.adults,
			status = EXCLUDED.status;
	`, id, origin, destination, departureDate, returnDate, 1, status) // Use provided dates
	if err != nil {
		t.Fatalf("Failed to seed search query %d with fixed dates: %v", id, err)
	}
}

// seedSearchQuery inserts a basic search query for testing GET endpoints
func seedSearchQuery(t *testing.T, id int, origin, destination string, status string) {
	t.Helper()
	_, err := db.GetPool().Exec(context.Background(), `
		INSERT INTO search_queries (id, origin, destination, departure_date, return_date, adults, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			origin = EXCLUDED.origin,
			destination = EXCLUDED.destination,
			departure_date = EXCLUDED.departure_date,
			return_date = EXCLUDED.return_date,
			adults = EXCLUDED.adults,
			status = EXCLUDED.status;
	`, id, origin, destination, time.Now().AddDate(0, 0, 7), time.Now().AddDate(0, 0, 14), 1, status) // Example dates/adults
	if err != nil {
		t.Fatalf("Failed to seed search query %d: %v", id, err)
	}
}
