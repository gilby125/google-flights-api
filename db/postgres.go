package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/flights" // Added import for flights package
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq" // Blank import for driver registration
)

// PostgresDB interface defines the database operations
type PostgresDB interface {
	GetSearchByID(id string) (*Search, error)
	SaveSearch(search *Search) (string, error)
	Search(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	GetCertificate(domain string) (*Certificate, error)
	StoreCertificate(domain string, cert []byte, key []byte, expires time.Time) error
	Close() error
	// GetDB() *sql.DB // Remove this to enforce mocking via interface methods
	InitSchema() error
	BeginTx(ctx context.Context) (Tx, error) // Add transaction support

	// Add specific query methods needed by handlers
	QueryAirports(ctx context.Context) (Rows, error)
	QueryAirlines(ctx context.Context) (Rows, error)
	GetSearchQueryByID(ctx context.Context, id int) (*SearchQuery, error) // Define SearchQuery struct later
	GetFlightOffersBySearchID(ctx context.Context, searchID int) (Rows, error)
	GetFlightSegmentsByOfferID(ctx context.Context, offerID int) (Rows, error)
	CountSearches(ctx context.Context) (int, error)
	QuerySearchesPaginated(ctx context.Context, limit, offset int) (Rows, error)
	DeleteJobDetailsByJobID(tx Tx, jobID int) error
	DeleteScheduledJobByID(tx Tx, jobID int) (int64, error)
	GetJobDetailsByID(ctx context.Context, jobID int) (*JobDetails, error) // Define JobDetails struct later
	UpdateJobLastRun(ctx context.Context, jobID int) error
	UpdateJobEnabled(ctx context.Context, jobID int, enabled bool) (int64, error)
	GetJobCronExpression(ctx context.Context, jobID int) (string, error)
	ListJobs(ctx context.Context) (Rows, error)
	CreateScheduledJob(tx Tx, name, cronExpression string, enabled bool) (int, error)
	CreateJobDetails(tx Tx, details JobDetails) error // Use JobDetails struct
	UpdateScheduledJob(tx Tx, jobID int, name, cronExpression string) error
	UpdateJobDetails(tx Tx, jobID int, details JobDetails) error              // Use JobDetails struct
	GetJobByID(ctx context.Context, jobID int) (*ScheduledJob, error)         // Define ScheduledJob struct later
	GetBulkSearchByID(ctx context.Context, searchID int) (*BulkSearch, error) // Define BulkSearch struct later
	QueryBulkSearchResultsPaginated(ctx context.Context, searchID, limit, offset int) (Rows, error)
}

// Tx defines the interface for database transactions
type Tx interface {
	Commit() error
	Rollback() error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) RowScanner // Use RowScanner and ...any
	// Add other methods used by handlers if necessary
}

// Rows defines the interface for query results
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
}

// Ensure sql.Rows implements our Rows interface
var _ Rows = (*sql.Rows)(nil)

// PostgresDBImpl is the concrete implementation
type PostgresDBImpl struct {
	db *sql.DB
}

// TxWrapper wraps *sql.Tx to satisfy the db.Tx interface, specifically adapting QueryRowContext.
type TxWrapper struct {
	*sql.Tx
}

// QueryRowContext calls the underlying *sql.Tx.QueryRowContext and returns the *sql.Row,
// which satisfies the db.RowScanner interface.
func (tw *TxWrapper) QueryRowContext(ctx context.Context, query string, args ...any) RowScanner {
	return tw.Tx.QueryRowContext(ctx, query, args...)
}

// Ensure TxWrapper implements db.Tx
var _ Tx = (*TxWrapper)(nil)

// TxImpl wraps sql.Tx to satisfy the db.Tx interface
// No longer strictly needed if we assert compatibility, but can be useful
// type TxImpl struct {
// 	*sql.Tx
// }

// func (t *TxImpl) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
// 	return t.Tx.ExecContext(ctx, query, args...)
// }
// func (t *TxImpl) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
// 	return t.Tx.QueryRowContext(ctx, query, args...)
// }

// ErrInvalidSearchID is returned when the search ID is invalid
type ErrInvalidSearchID struct {
	ID string
}

func (e *ErrInvalidSearchID) Error() string {
	return fmt.Sprintf("invalid search ID: %s", e.ID)
}

// SearchNotFoundError is returned when the search is not found
type SearchNotFoundError struct {
	ID string
}

func (e *SearchNotFoundError) Error() string {
	return fmt.Sprintf("search with ID %s not found", e.ID)
}

// GetCertificate retrieves a certificate from the database
func (p *PostgresDBImpl) GetCertificate(domain string) (*Certificate, error) {
	var cert Certificate
	err := p.db.QueryRow(
		`SELECT domain, cert_chain, private_key_enc, expires 
		FROM certificate_issuance 
		WHERE domain = $1 
		ORDER BY expires DESC 
		LIMIT 1`,
		domain,
	).Scan(&cert.Domain, &cert.CertChain, &cert.PrivateKey, &cert.Expires)

	if err != nil {
		return nil, fmt.Errorf("error getting certificate: %w", err)
	}
	return &cert, nil
}

// StoreCertificate saves a certificate to the database
func (p *PostgresDBImpl) StoreCertificate(domain string, certChain []byte, privateKey []byte, expires time.Time) error {
	_, err := p.db.Exec(
		`INSERT INTO certificate_issuance 
		(domain, cert_chain, private_key_enc, expires) 
		VALUES ($1, $2, $3, $4) 
		ON CONFLICT (domain) 
		DO UPDATE SET 
			cert_chain = EXCLUDED.cert_chain,
			private_key_enc = EXCLUDED.private_key_enc,
			expires = EXCLUDED.expires`,
		domain,
		certChain,
		privateKey,
		expires,
	)
	return err
}

var (
	pgPool *pgxpool.Pool
)

// GetPool returns the PostgreSQL connection pool
func GetPool() *pgxpool.Pool {
	return pgPool
}

// InitTestPostgres initializes a test PostgreSQL pool
func InitTestPostgres(cfg *config.Config) *pgxpool.Pool {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&sslcert=%s&sslkey=%s&sslrootcert=%s",
		cfg.PostgresConfig.User,
		cfg.PostgresConfig.Password,
		cfg.PostgresConfig.Host,
		cfg.PostgresConfig.Port,
		cfg.PostgresConfig.DBName,
		cfg.PostgresConfig.SSLMode,
		cfg.PostgresConfig.SSLCert,
		cfg.PostgresConfig.SSLKey,
		cfg.PostgresConfig.SSLRootCert,
	)

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		panic(fmt.Sprintf("Unable to connect to test database: %v", err))
	}

	pgPool = pool
	return pool
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg config.PostgresConfig) (PostgresDB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
		cfg.SSLCert, cfg.SSLKey, cfg.SSLRootCert)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &PostgresDBImpl{db: db}, nil
}

// GetSearchByID retrieves a search record by its ID
func (p *PostgresDBImpl) GetSearchByID(id string) (*Search, error) {
	var search Search
	// var results pq.StringArray // Commented out as it's declared but not used yet

	// Adjust the query based on the actual schema for search_queries and how results are stored
	// This is a placeholder query assuming a 'results' column exists and stores FlightOffer data somehow (e.g., JSONB)
	// You might need a JOIN or separate query to fetch associated FlightOffers if they are in a different table.
	err := p.db.QueryRow(`
		SELECT 
			id, origin, destination, departure_date, return_date, 
			adults + children + infants_lap + infants_seat, -- Calculate total passengers
			class, status, created_at, updated_at -- Need to fetch results separately or adjust schema
		FROM search_queries 
		WHERE id = $1`, id).Scan(
		&search.ID, &search.Origin, &search.Destination, &search.Departure, &search.Return,
		&search.Passengers, &search.CabinClass, &search.Status, &search.CreatedAt, &search.UpdatedAt,
		// &results, // Placeholder for results scanning - requires schema knowledge
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("search with ID %s not found", id)
		}
		return nil, fmt.Errorf("error getting search by ID %s: %w", id, err)
	}

	// TODO: Process the 'results' data (e.g., unmarshal JSON) into search.Results ([]flights.FullOffer)
	// This part depends heavily on how FullOffer data is stored in the database.
	// For now, leaving search.Results empty.
	search.Results = []flights.FullOffer{} // Initialize empty slice with correct type

	return &search, nil
}

// SaveSearch saves a search record to the database
// TODO: Implement SaveSearch logic
func (p *PostgresDBImpl) SaveSearch(search *Search) (string, error) {
	// Placeholder implementation
	// You'll need an INSERT or UPDATE query here based on your schema
	// Example:
	// err := p.db.QueryRow("INSERT INTO search_queries (...) VALUES (...) RETURNING id", ...).Scan(&search.ID)
	// return search.ID, err
	return "", fmt.Errorf("SaveSearch not implemented")
}

// Search executes a generic search query
// TODO: Implement Search logic if needed, or remove if unused
func (p *PostgresDBImpl) Search(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.QueryContext(ctx, query, args...)
}

// Close closes the database connection
func (p *PostgresDBImpl) Close() error {
	return p.db.Close()
}

// BeginTx starts a new transaction
func (p *PostgresDBImpl) BeginTx(ctx context.Context) (Tx, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Return the wrapped transaction
	return &TxWrapper{Tx: tx}, nil
}

// --- Implementation of Specific Query Methods ---

func (p *PostgresDBImpl) QueryAirports(ctx context.Context) (Rows, error) {
	return p.db.QueryContext(ctx, "SELECT code, name, city, country, latitude, longitude FROM airports")
}

func (p *PostgresDBImpl) QueryAirlines(ctx context.Context) (Rows, error) {
	return p.db.QueryContext(ctx, "SELECT code, name, country FROM airlines")
}

func (p *PostgresDBImpl) GetSearchQueryByID(ctx context.Context, id int) (*SearchQuery, error) {
	var query SearchQuery
	err := p.db.QueryRowContext(ctx,
		`SELECT id, origin, destination, departure_date, return_date, status, created_at
		FROM search_queries WHERE id = $1`,
		id,
	).Scan(&query.ID, &query.Origin, &query.Destination, &query.DepartureDate, &query.ReturnDate, &query.Status, &query.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("search query with ID %d not found", id) // Consider a specific error type
		}
		return nil, fmt.Errorf("error getting search query by ID %d: %w", id, err)
	}
	return &query, nil
}

func (p *PostgresDBImpl) GetFlightOffersBySearchID(ctx context.Context, searchID int) (Rows, error) {
	// Assuming search_query_id is the foreign key in flight_offers
	return p.db.QueryContext(ctx,
		`SELECT id, price, currency, departure_date, return_date, total_duration, created_at
		FROM flight_offers WHERE search_query_id = $1`, // Adjust column name if needed
		searchID,
	)
}

func (p *PostgresDBImpl) GetFlightSegmentsByOfferID(ctx context.Context, offerID int) (Rows, error) {
	// Assuming flight_offer_id is the foreign key in flight_segments
	return p.db.QueryContext(ctx,
		`SELECT airline_code, flight_number, departure_airport, arrival_airport,
		departure_time, arrival_time, duration, airplane, legroom, is_return
		FROM flight_segments WHERE flight_offer_id = $1`, // Adjust column name if needed
		offerID,
	)
}

func (p *PostgresDBImpl) CountSearches(ctx context.Context) (int, error) {
	var total int
	err := p.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM search_queries").Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("error counting searches: %w", err)
	}
	return total, nil
}

func (p *PostgresDBImpl) QuerySearchesPaginated(ctx context.Context, limit, offset int) (Rows, error) {
	return p.db.QueryContext(ctx,
		`SELECT id, origin, destination, departure_date, return_date, status, created_at
		FROM search_queries ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
}

func (p *PostgresDBImpl) DeleteJobDetailsByJobID(tx Tx, jobID int) error {
	_, err := tx.ExecContext(context.Background(), "DELETE FROM job_details WHERE job_id = $1", jobID)
	if err != nil {
		return fmt.Errorf("error deleting job details for job ID %d: %w", jobID, err)
	}
	return nil
}

func (p *PostgresDBImpl) DeleteScheduledJobByID(tx Tx, jobID int) (int64, error) {
	result, err := tx.ExecContext(context.Background(), "DELETE FROM scheduled_jobs WHERE id = $1", jobID)
	if err != nil {
		return 0, fmt.Errorf("error deleting scheduled job ID %d: %w", jobID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error getting rows affected after deleting job ID %d: %w", jobID, err)
	}
	return rowsAffected, nil
}

func (p *PostgresDBImpl) GetJobDetailsByID(ctx context.Context, jobID int) (*JobDetails, error) {
	var details JobDetails
	err := p.db.QueryRowContext(ctx,
		`SELECT job_id, origin, destination, departure_date_start, departure_date_end,
		return_date_start, return_date_end, trip_length, adults, children,
		infants_lap, infants_seat, trip_type, class, stops
		FROM job_details WHERE job_id = $1`,
		jobID,
	).Scan(
		&details.JobID, &details.Origin, &details.Destination, &details.DepartureDateStart, &details.DepartureDateEnd,
		&details.ReturnDateStart, &details.ReturnDateEnd, &details.TripLength, &details.Adults, &details.Children,
		&details.InfantsLap, &details.InfantsSeat, &details.TripType, &details.Class, &details.Stops,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job details for job ID %d not found", jobID)
		}
		return nil, fmt.Errorf("error getting job details for job ID %d: %w", jobID, err)
	}
	// Assuming Currency is not in job_details, set default or fetch separately if needed
	details.Currency = "USD" // Example default
	return &details, nil
}

func (p *PostgresDBImpl) UpdateJobLastRun(ctx context.Context, jobID int) error {
	_, err := p.db.ExecContext(ctx,
		"UPDATE scheduled_jobs SET last_run = NOW() WHERE id = $1",
		jobID,
	)
	if err != nil {
		return fmt.Errorf("error updating last run time for job ID %d: %w", jobID, err)
	}
	return nil
}

func (p *PostgresDBImpl) UpdateJobEnabled(ctx context.Context, jobID int, enabled bool) (int64, error) {
	result, err := p.db.ExecContext(ctx,
		"UPDATE scheduled_jobs SET enabled = $1, updated_at = NOW() WHERE id = $2",
		enabled, jobID,
	)
	if err != nil {
		return 0, fmt.Errorf("error updating enabled status for job ID %d: %w", jobID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error getting rows affected after updating enabled status for job ID %d: %w", jobID, err)
	}
	return rowsAffected, nil
}

func (p *PostgresDBImpl) GetJobCronExpression(ctx context.Context, jobID int) (string, error) {
	var cronExpr string
	err := p.db.QueryRowContext(ctx, "SELECT cron_expression FROM scheduled_jobs WHERE id = $1", jobID).Scan(&cronExpr)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("job with ID %d not found when getting cron expression", jobID)
		}
		return "", fmt.Errorf("error getting cron expression for job ID %d: %w", jobID, err)
	}
	return cronExpr, nil
}

func (p *PostgresDBImpl) ListJobs(ctx context.Context) (Rows, error) {
	return p.db.QueryContext(ctx,
		`SELECT id, name, cron_expression, enabled, last_run, created_at, updated_at
		FROM scheduled_jobs ORDER BY created_at DESC`,
	)
}

func (p *PostgresDBImpl) CreateScheduledJob(tx Tx, name, cronExpression string, enabled bool) (int, error) {
	var jobID int
	err := tx.QueryRowContext(context.Background(), // Use background context within transaction method
		`INSERT INTO scheduled_jobs (name, cron_expression, enabled)
		VALUES ($1, $2, $3) RETURNING id`,
		name, cronExpression, enabled,
	).Scan(&jobID)
	if err != nil {
		return 0, fmt.Errorf("error creating scheduled job: %w", err)
	}
	return jobID, nil
}

func (p *PostgresDBImpl) CreateJobDetails(tx Tx, details JobDetails) error {
	_, err := tx.ExecContext(context.Background(),
		`INSERT INTO job_details
		(job_id, origin, destination, departure_date_start, departure_date_end,
		return_date_start, return_date_end, trip_length, adults, children,
		infants_lap, infants_seat, trip_type, class, stops)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		details.JobID, details.Origin, details.Destination, details.DepartureDateStart, details.DepartureDateEnd,
		details.ReturnDateStart, details.ReturnDateEnd, details.TripLength, details.Adults, details.Children,
		details.InfantsLap, details.InfantsSeat, details.TripType, details.Class, details.Stops,
	)
	if err != nil {
		return fmt.Errorf("error creating job details: %w", err)
	}
	return nil
}

func (p *PostgresDBImpl) UpdateScheduledJob(tx Tx, jobID int, name, cronExpression string) error {
	_, err := tx.ExecContext(context.Background(),
		`UPDATE scheduled_jobs SET name = $1, cron_expression = $2, updated_at = NOW() WHERE id = $3`,
		name, cronExpression, jobID,
	)
	if err != nil {
		return fmt.Errorf("error updating scheduled job ID %d: %w", jobID, err)
	}
	return nil
}

func (p *PostgresDBImpl) UpdateJobDetails(tx Tx, jobID int, details JobDetails) error {
	_, err := tx.ExecContext(context.Background(),
		`UPDATE job_details SET origin = $1, destination = $2, departure_date_start = $3, departure_date_end = $4,
		return_date_start = $5, return_date_end = $6, trip_length = $7, adults = $8, children = $9,
		infants_lap = $10, infants_seat = $11, trip_type = $12, class = $13, stops = $14 WHERE job_id = $15`,
		details.Origin, details.Destination, details.DepartureDateStart, details.DepartureDateEnd,
		details.ReturnDateStart, details.ReturnDateEnd, details.TripLength, details.Adults, details.Children,
		details.InfantsLap, details.InfantsSeat, details.TripType, details.Class, details.Stops,
		jobID,
	)
	if err != nil {
		return fmt.Errorf("error updating job details for job ID %d: %w", jobID, err)
	}
	return nil
}

func (p *PostgresDBImpl) GetJobByID(ctx context.Context, jobID int) (*ScheduledJob, error) {
	var job ScheduledJob
	err := p.db.QueryRowContext(ctx,
		`SELECT id, name, cron_expression, enabled, last_run, created_at, updated_at
		FROM scheduled_jobs WHERE id = $1`,
		jobID,
	).Scan(&job.ID, &job.Name, &job.CronExpression, &job.Enabled, &job.LastRun, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job with ID %d not found", jobID)
		}
		return nil, fmt.Errorf("error getting job by ID %d: %w", jobID, err)
	}
	return &job, nil
}

func (p *PostgresDBImpl) GetBulkSearchByID(ctx context.Context, searchID int) (*BulkSearch, error) {
	var search BulkSearch
	err := p.db.QueryRowContext(ctx,
		`SELECT id, status, total_searches, completed, created_at,
			completed_at, min_price, max_price, average_price
		FROM bulk_searches WHERE id = $1`, // Assuming table name is bulk_searches
		searchID,
	).Scan(&search.ID, &search.Status, &search.TotalSearches, &search.Completed,
		&search.CreatedAt, &search.CompletedAt, &search.MinPrice,
		&search.MaxPrice, &search.AveragePrice)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bulk search with ID %d not found", searchID)
		}
		return nil, fmt.Errorf("error getting bulk search by ID %d: %w", searchID, err)
	}
	return &search, nil
}

func (p *PostgresDBImpl) QueryBulkSearchResultsPaginated(ctx context.Context, searchID, limit, offset int) (Rows, error) {
	return p.db.QueryContext(ctx,
		`SELECT origin, destination, departure_date, return_date,
			price, currency, airline_code, duration
		FROM bulk_search_results
		WHERE bulk_search_id = $1
		ORDER BY price ASC
		LIMIT $2 OFFSET $3`,
		searchID, limit, offset,
	)
}

// --- End Implementation ---

// GetDB returns the underlying database connection
// func (p *PostgresDBImpl) GetDB() *sql.DB {
// 	return p.db
// }

// InitSchema initializes the database schema
func (p *PostgresDBImpl) InitSchema() error {
	// Ensure a clean slate by dropping tables first (in reverse dependency order)
	// Note: This makes InitSchema destructive on existing data if tables exist.
	// In a real application, migrations are preferred.
	_, err := p.db.Exec(`
		DROP TABLE IF EXISTS flight_segments CASCADE;
		DROP TABLE IF EXISTS flight_prices CASCADE;
		DROP TABLE IF EXISTS flights CASCADE;
		DROP TABLE IF EXISTS job_details CASCADE;
		DROP TABLE IF EXISTS scheduled_jobs CASCADE;
		DROP TABLE IF EXISTS search_results CASCADE;
		DROP TABLE IF EXISTS flight_offers CASCADE;
		DROP TABLE IF EXISTS search_queries CASCADE;
		DROP TABLE IF EXISTS certificate_issuance CASCADE;
		DROP TABLE IF EXISTS app_secrets CASCADE;
		DROP TABLE IF EXISTS airports CASCADE;
		DROP TABLE IF EXISTS airlines CASCADE;
		-- Add other tables like users, sessions if they exist and have dependencies
		DROP TABLE IF EXISTS sessions CASCADE;
		DROP TABLE IF EXISTS users CASCADE;
	`)
	if err != nil {
		// Log error during drop, but proceed as tables might not exist initially
		fmt.Printf("Warning: Error dropping tables during InitSchema (might be first run): %v\n", err)
	}

	// Now create tables in the correct order

	// Create airports table first
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS airports (
            id SERIAL PRIMARY KEY,
            code VARCHAR(3) UNIQUE NOT NULL,
            name VARCHAR(255) NOT NULL,
            city VARCHAR(100) NOT NULL,
            country VARCHAR(100) NOT NULL,
            latitude DOUBLE PRECISION NOT NULL,
            longitude DOUBLE PRECISION NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create airports table: %w", err)
	}

	// Then create airlines table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS airlines (
            id SERIAL PRIMARY KEY,
            code VARCHAR(3) UNIQUE NOT NULL,
            name VARCHAR(255) NOT NULL,
            country VARCHAR(100) NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create airlines table: %w", err)
	}

	// Only after both tables above are created, create the flights table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS flights (
            id SERIAL PRIMARY KEY,
            flight_number VARCHAR(10) NOT NULL,
            airline_id INTEGER REFERENCES airlines(id),
            origin_id INTEGER REFERENCES airports(id),
            destination_id INTEGER REFERENCES airports(id),
            departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
            arrival_time TIMESTAMP WITH TIME ZONE NOT NULL,
            duration INTEGER NOT NULL,
            distance INTEGER,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create flights table: %w", err)
	}

	// Create flight_prices table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS flight_prices (
            id SERIAL PRIMARY KEY,
            flight_id INTEGER REFERENCES flights(id),
            price DECIMAL(10, 2) NOT NULL,
            currency VARCHAR(3) NOT NULL,
            cabin_class VARCHAR(20) NOT NULL,
            search_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create flight_prices table: %w", err)
	}

	// Create flight_segments table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS flight_segments (
            id SERIAL PRIMARY KEY,
            flight_id INTEGER REFERENCES flights(id),
            segment_number INTEGER NOT NULL,
            origin_id INTEGER REFERENCES airports(id),
            destination_id INTEGER REFERENCES airports(id),
            departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
            arrival_time TIMESTAMP WITH TIME ZONE NOT NULL,
            duration INTEGER NOT NULL,
            airline_id INTEGER REFERENCES airlines(id),
            flight_number VARCHAR(10) NOT NULL,
            aircraft_type VARCHAR(50),
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create flight_segments table: %w", err)
	}

	// Create scheduled_jobs table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS scheduled_jobs (
            id SERIAL PRIMARY KEY,
            name VARCHAR(100) NOT NULL,
            cron_expression VARCHAR(100) NOT NULL,
            enabled BOOLEAN DEFAULT TRUE,
            last_run TIMESTAMP WITH TIME ZONE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)

	// Create search_queries table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS search_queries (
            id SERIAL PRIMARY KEY,
            origin VARCHAR(3) NOT NULL,
            destination VARCHAR(3) NOT NULL,
            departure_date DATE NOT NULL,
            return_date DATE,
            adults INTEGER NOT NULL,
            children INTEGER NOT NULL,
            infants_lap INTEGER NOT NULL,
            infants_seat INTEGER NOT NULL,
            trip_type VARCHAR(20) NOT NULL,
            class VARCHAR(20) NOT NULL,
            stops VARCHAR(20) NOT NULL,
            status VARCHAR(20) NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create search_queries table: %w", err)
	}

	// Create job_details table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS job_details (
            id SERIAL PRIMARY KEY,
            job_id INTEGER REFERENCES scheduled_jobs(id),
            origin VARCHAR(3) NOT NULL,
            destination VARCHAR(3) NOT NULL,
            departure_date_start DATE NOT NULL,
            departure_date_end DATE NOT NULL,
            return_date_start DATE,
            return_date_end DATE,
            trip_length INTEGER,
            adults INTEGER DEFAULT 1,
            children INTEGER DEFAULT 0,
            infants_lap INTEGER DEFAULT 0,
            infants_seat INTEGER DEFAULT 0,
            trip_type VARCHAR(20) DEFAULT 'round_trip',
            class VARCHAR(20) DEFAULT 'economy',
            stops VARCHAR(20) DEFAULT 'any',
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create job_details table: %w", err)
	}

	// Create search_results table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS search_results (
            id SERIAL PRIMARY KEY,
            search_id UUID NOT NULL,
            origin VARCHAR(3) NOT NULL,
            destination VARCHAR(3) NOT NULL,
            departure_date DATE NOT NULL,
            return_date DATE,
            adults INTEGER NOT NULL,
            children INTEGER NOT NULL,
            infants_lap INTEGER NOT NULL,
            infants_seat INTEGER NOT NULL,
            trip_type VARCHAR(20) NOT NULL,
            class VARCHAR(20) NOT NULL,
            stops VARCHAR(20) NOT NULL,
            currency VARCHAR(3) NOT NULL,
            min_price DECIMAL(10, 2),
            max_price DECIMAL(10, 2),
            search_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create search_results table: %w", err)
	}

	// Create flight_offers table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS flight_offers (
            id SERIAL PRIMARY KEY,
            search_id UUID NOT NULL,
            price DECIMAL(10, 2) NOT NULL,
            currency VARCHAR(3) NOT NULL,
            airline_codes TEXT[] NOT NULL,
            outbound_duration INTEGER NOT NULL,
            outbound_stops INTEGER NOT NULL,
            return_duration INTEGER,
            return_stops INTEGER,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create flight_offers table: %w", err)
	}

	// Create certificate tracking table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS certificate_issuance (
            id SERIAL PRIMARY KEY,
            domain VARCHAR(253) NOT NULL,
            serial_number VARCHAR(255) UNIQUE NOT NULL,
            issued_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
            dns_challenge TEXT NOT NULL,
            cert_chain TEXT NOT NULL,
            private_key_enc TEXT NOT NULL,
            cloudflare_validation_id UUID,
            validation_status VARCHAR(50) NOT NULL,
            last_renewal_attempt TIMESTAMP WITH TIME ZONE,
            renewal_errors INT DEFAULT 0
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create certificate_issuance table: %w", err)
	}

	// Create encrypted secrets storage table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS app_secrets (
            id SERIAL PRIMARY KEY,
            secret_name VARCHAR(255) UNIQUE NOT NULL,
            encrypted_value BYTEA NOT NULL,
            key_id VARCHAR(255) NOT NULL,
            rotation_schedule INTERVAL NOT NULL,
            last_rotated TIMESTAMP WITH TIME ZONE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create app_secrets table: %w", err)
	}

	// Create indexes for better query performance
	_, err = p.db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_airports_code ON airports(code);
        CREATE INDEX IF NOT EXISTS idx_airlines_code ON airlines(code);
        CREATE INDEX IF NOT EXISTS idx_flights_departure ON flights(departure_time);
        CREATE INDEX IF NOT EXISTS idx_flights_arrival ON flights(arrival_time);
        CREATE INDEX IF NOT EXISTS idx_flight_prices_search_date ON flight_prices(search_date);
        CREATE INDEX IF NOT EXISTS idx_search_results_search_id ON search_results(search_id);
        CREATE INDEX IF NOT EXISTS idx_flight_offers_search_id ON flight_offers(search_id);
    `)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// RunMigrations executes database migrations
func RunMigrations(connString string) error {
	// Implementation would use a migration library like golang-migrate
	// For testing purposes, we'll just return nil
	return nil
}
