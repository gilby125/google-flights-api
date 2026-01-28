package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/flights" // Added import for flights package
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq" // Blank import for driver registration
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
	DeleteJobDetailsByJobID(ctx context.Context, tx Tx, jobID int) error
	DeleteScheduledJobByID(ctx context.Context, tx Tx, jobID int) (int64, error)
	GetJobDetailsByID(ctx context.Context, jobID int) (*JobDetails, error) // Define JobDetails struct later
	UpdateJobLastRun(ctx context.Context, jobID int) error
	UpdateJobEnabled(ctx context.Context, jobID int, enabled bool) (int64, error)
	GetJobCronExpression(ctx context.Context, jobID int) (string, error)
	ListJobs(ctx context.Context) (Rows, error)
	CreateScheduledJob(ctx context.Context, tx Tx, name, cronExpression string, enabled bool) (int, error)
	CreateJobDetails(ctx context.Context, tx Tx, details JobDetails) error
	UpdateScheduledJob(ctx context.Context, tx Tx, jobID int, name, cronExpression string) error
	UpdateJobDetails(ctx context.Context, tx Tx, jobID int, details JobDetails) error
	GetJobByID(ctx context.Context, jobID int) (*ScheduledJob, error)         // Define ScheduledJob struct later
	GetBulkSearchByID(ctx context.Context, searchID int) (*BulkSearch, error) // Define BulkSearch struct later
	QueryBulkSearchResultsPaginated(ctx context.Context, searchID, limit, offset int) (Rows, error)
	InsertBulkSearchOffer(ctx context.Context, record BulkSearchOfferRecord) error
	QueryBulkSearchOffersPaginated(ctx context.Context, searchID, limit, offset int) (Rows, error)
	ListBulkSearchOffers(ctx context.Context, searchID int) ([]BulkSearchOffer, error)
	CreateBulkSearchRecord(ctx context.Context, jobID sql.NullInt32, totalSearches int, currency, status string) (int, error)
	UpdateBulkSearchStatus(ctx context.Context, bulkSearchID int, status string) error
	UpdateBulkSearchProgress(ctx context.Context, bulkSearchID int, completed, totalOffers, errorCount int) error
	CompleteBulkSearch(ctx context.Context, summary BulkSearchSummary) error
	InsertBulkSearchResult(ctx context.Context, result BulkSearchResultRecord) error
	InsertBulkSearchResultsBatch(ctx context.Context, results []BulkSearchResultRecord) error
	ListBulkSearches(ctx context.Context, limit, offset int) (Rows, error)
	CreatePriceGraphSweep(ctx context.Context, jobID sql.NullInt32, originCount, destinationCount int, tripLengthMin, tripLengthMax sql.NullInt32, currency string) (int, error)
	UpdatePriceGraphSweepStatus(ctx context.Context, sweepID int, status string, startedAt, completedAt sql.NullTime, errorCount int) error
	GetPriceGraphSweepByID(ctx context.Context, sweepID int) (*PriceGraphSweep, error)
	ListPriceGraphSweeps(ctx context.Context, limit, offset int) (Rows, error)
	InsertPriceGraphResult(ctx context.Context, record PriceGraphResultRecord) error
	ListPriceGraphResults(ctx context.Context, sweepID, limit, offset int) (Rows, error)

	// Continuous sweep progress methods
	SaveContinuousSweepProgress(ctx context.Context, progress ContinuousSweepProgress) error
	SetContinuousSweepControlFlags(ctx context.Context, isRunning, isPaused *bool) error
	GetContinuousSweepProgress(ctx context.Context) (*ContinuousSweepProgress, error)
	InsertContinuousSweepStats(ctx context.Context, stats ContinuousSweepStats) error
	ListContinuousSweepStats(ctx context.Context, limit int) ([]ContinuousSweepStats, error)
	ListContinuousSweepResults(ctx context.Context, filters ContinuousSweepResultsFilter) ([]PriceGraphResultRecord, error)

	// Deal detection methods
	GetRouteBaseline(ctx context.Context, origin, dest string, tripLength int, class string) (*RouteBaseline, error)
	UpsertRouteBaseline(ctx context.Context, baseline RouteBaseline) error
	GetPriceHistoryForRoute(ctx context.Context, origin, dest string, tripLength int, class string, windowDays int) ([]float64, error)
	InsertDetectedDeal(ctx context.Context, deal DetectedDeal) (int, error)
	UpsertDetectedDeal(ctx context.Context, deal DetectedDeal) error
	GetDetectedDealByFingerprint(ctx context.Context, fingerprint string) (*DetectedDeal, error)
	ListActiveDeals(ctx context.Context, filter DealFilter) ([]DetectedDeal, error)
	ExpireOldDeals(ctx context.Context) (int64, error)
	InsertDealAlert(ctx context.Context, alert DealAlert) (int, error)
	ListDealAlerts(ctx context.Context, limit, offset int) ([]DealAlert, error)
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
	connStr := BuildPostgresConnString(cfg)

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

// BuildPostgresConnString builds a lib/pq keyword/value connection string from config.
func BuildPostgresConnString(cfg config.PostgresConfig) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
		cfg.SSLCert, cfg.SSLKey, cfg.SSLRootCert,
	)
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
	// First get the UUID search_id from search_queries table, then query flight_offers
	// This joins the integer search_queries.id with the UUID flight_offers.search_id
	return p.db.QueryContext(ctx,
		`SELECT fo.id, fo.price, fo.currency, fo.airline_codes, fo.outbound_duration, fo.outbound_stops, fo.return_duration, fo.return_stops, fo.created_at
		FROM flight_offers fo
		JOIN search_results sr ON fo.search_id = sr.search_id
		WHERE sr.search_query_id = $1`,
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

func (p *PostgresDBImpl) DeleteJobDetailsByJobID(ctx context.Context, tx Tx, jobID int) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM job_details WHERE job_id = $1", jobID)
	if err != nil {
		return fmt.Errorf("error deleting job details for job ID %d: %w", jobID, err)
	}
	return nil
}

func (p *PostgresDBImpl) DeleteScheduledJobByID(ctx context.Context, tx Tx, jobID int) (int64, error) {
	result, err := tx.ExecContext(ctx, "DELETE FROM scheduled_jobs WHERE id = $1", jobID)
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
		return_date_start, return_date_end, trip_length, dynamic_dates, days_from_execution, search_window_days, adults, children,
		infants_lap, infants_seat, trip_type, class, stops, currency
		FROM job_details WHERE job_id = $1`,
		jobID,
	).Scan(
		&details.JobID, &details.Origin, &details.Destination, &details.DepartureDateStart, &details.DepartureDateEnd,
		&details.ReturnDateStart, &details.ReturnDateEnd, &details.TripLength, &details.DynamicDates, &details.DaysFromExecution, &details.SearchWindowDays, &details.Adults, &details.Children,
		&details.InfantsLap, &details.InfantsSeat, &details.TripType, &details.Class, &details.Stops, &details.Currency,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job details for job ID %d not found", jobID)
		}
		return nil, fmt.Errorf("error getting job details for job ID %d: %w", jobID, err)
	}
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
		`SELECT sj.id, sj.name, sj.cron_expression, sj.enabled, sj.last_run, sj.created_at, sj.updated_at,
				jd.origin, jd.destination, jd.dynamic_dates, jd.days_from_execution, jd.search_window_days, jd.trip_length
		 FROM scheduled_jobs sj
		 LEFT JOIN job_details jd ON sj.id = jd.job_id
		 ORDER BY sj.created_at DESC`,
	)
}

func (p *PostgresDBImpl) CreateScheduledJob(ctx context.Context, tx Tx, name, cronExpression string, enabled bool) (int, error) {
	var jobID int
	err := tx.QueryRowContext(ctx,
		`INSERT INTO scheduled_jobs (name, cron_expression, enabled)
		VALUES ($1, $2, $3) RETURNING id`,
		name, cronExpression, enabled,
	).Scan(&jobID)
	if err != nil {
		return 0, fmt.Errorf("error creating scheduled job: %w", err)
	}
	return jobID, nil
}

func (p *PostgresDBImpl) CreateJobDetails(ctx context.Context, tx Tx, details JobDetails) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO job_details
		(job_id, origin, destination, departure_date_start, departure_date_end,
		return_date_start, return_date_end, trip_length, dynamic_dates, days_from_execution, search_window_days, adults, children,
		infants_lap, infants_seat, trip_type, class, stops, currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		details.JobID, details.Origin, details.Destination, details.DepartureDateStart, details.DepartureDateEnd,
		details.ReturnDateStart, details.ReturnDateEnd, details.TripLength, details.DynamicDates, details.DaysFromExecution, details.SearchWindowDays, details.Adults, details.Children,
		details.InfantsLap, details.InfantsSeat, details.TripType, details.Class, details.Stops, details.Currency,
	)
	if err != nil {
		return fmt.Errorf("error creating job details: %w", err)
	}
	return nil
}

func (p *PostgresDBImpl) UpdateScheduledJob(ctx context.Context, tx Tx, jobID int, name, cronExpression string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE scheduled_jobs SET name = $1, cron_expression = $2, updated_at = NOW() WHERE id = $3`,
		name, cronExpression, jobID,
	)
	if err != nil {
		return fmt.Errorf("error updating scheduled job ID %d: %w", jobID, err)
	}
	return nil
}

func (p *PostgresDBImpl) UpdateJobDetails(ctx context.Context, tx Tx, jobID int, details JobDetails) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE job_details SET origin = $1, destination = $2, departure_date_start = $3, departure_date_end = $4,
		return_date_start = $5, return_date_end = $6, trip_length = $7, adults = $8, children = $9,
		infants_lap = $10, infants_seat = $11, trip_type = $12, class = $13, stops = $14, currency = $15, updated_at = NOW() WHERE job_id = $16`,
		details.Origin, details.Destination, details.DepartureDateStart, details.DepartureDateEnd,
		details.ReturnDateStart, details.ReturnDateEnd, details.TripLength, details.Adults, details.Children,
		details.InfantsLap, details.InfantsSeat, details.TripType, details.Class, details.Stops, details.Currency,
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
		`SELECT id, job_id, status, total_searches, completed, total_offers,
            error_count, currency, created_at, updated_at, completed_at,
			min_price, max_price, average_price
		FROM bulk_searches WHERE id = $1`,
		searchID,
	).Scan(&search.ID, &search.JobID, &search.Status, &search.TotalSearches, &search.Completed,
		&search.TotalOffers, &search.ErrorCount, &search.Currency, &search.CreatedAt, &search.UpdatedAt,
		&search.CompletedAt, &search.MinPrice, &search.MaxPrice, &search.AveragePrice)

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
			price, currency, airline_code, duration,
			src_airport_code, dst_airport_code, src_city, dst_city,
			flight_duration, return_flight_duration,
			outbound_flights, return_flights, offer_json
		FROM bulk_search_results
		WHERE bulk_search_id = $1
		ORDER BY price ASC
		LIMIT $2 OFFSET $3`,
		searchID, limit, offset,
	)
}

func (p *PostgresDBImpl) CreateBulkSearchRecord(ctx context.Context, jobID sql.NullInt32, totalSearches int, currency, status string) (int, error) {
	if status == "" {
		status = "queued"
	}
	var bulkSearchID int
	err := p.db.QueryRowContext(ctx,
		`INSERT INTO bulk_searches (job_id, status, total_searches, currency)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		jobID, status, totalSearches, currency,
	).Scan(&bulkSearchID)
	if err != nil {
		return 0, fmt.Errorf("failed to create bulk search record: %w", err)
	}
	return bulkSearchID, nil
}

func (p *PostgresDBImpl) UpdateBulkSearchStatus(ctx context.Context, bulkSearchID int, status string) error {
	_, err := p.db.ExecContext(ctx,
		`UPDATE bulk_searches
         SET status = $1, updated_at = NOW()
         WHERE id = $2`,
		status, bulkSearchID,
	)
	if err != nil {
		return fmt.Errorf("failed to update bulk search %d status to %s: %w", bulkSearchID, status, err)
	}
	return nil
}

func (p *PostgresDBImpl) UpdateBulkSearchProgress(ctx context.Context, bulkSearchID int, completed, totalOffers, errorCount int) error {
	_, err := p.db.ExecContext(ctx,
		`UPDATE bulk_searches
         SET completed = $1,
             total_offers = $2,
             error_count = $3,
             updated_at = NOW()
         WHERE id = $4`,
		completed,
		totalOffers,
		errorCount,
		bulkSearchID,
	)
	if err != nil {
		return fmt.Errorf("failed to update bulk search %d progress: %w", bulkSearchID, err)
	}
	return nil
}

func (p *PostgresDBImpl) CompleteBulkSearch(ctx context.Context, summary BulkSearchSummary) error {
	_, err := p.db.ExecContext(ctx,
		`UPDATE bulk_searches
         SET status = $1,
             completed = $2,
             total_offers = $3,
             error_count = $4,
             min_price = $5,
             max_price = $6,
             average_price = $7,
             completed_at = NOW(),
             updated_at = NOW()
         WHERE id = $8`,
		summary.Status,
		summary.Completed,
		summary.TotalOffers,
		summary.ErrorCount,
		summary.MinPrice,
		summary.MaxPrice,
		summary.AveragePrice,
		summary.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to complete bulk search %d: %w", summary.ID, err)
	}
	return nil
}

func (p *PostgresDBImpl) InsertBulkSearchOffer(ctx context.Context, record BulkSearchOfferRecord) error {
	airlineCodes := pq.Array(record.AirlineCodes)
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO bulk_search_offers
			(bulk_search_id, origin, destination, departure_date, return_date,
			 price, currency, airline_codes, src_airport_code, dst_airport_code,
			 src_city, dst_city, flight_duration, return_flight_duration,
			 distance_miles, cost_per_mile,
			 outbound_flights, return_flights, offer_json)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		record.BulkSearchID,
		record.Origin,
		record.Destination,
		record.DepartureDate,
		record.ReturnDate,
		record.Price,
		record.Currency,
		airlineCodes,
		record.SrcAirportCode,
		record.DstAirportCode,
		record.SrcCity,
		record.DstCity,
		record.FlightDuration,
		record.ReturnFlightDuration,
		record.DistanceMiles,
		record.CostPerMile,
		record.OutboundFlightsJSON,
		record.ReturnFlightsJSON,
		record.OfferJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to insert bulk search offer: %w", err)
	}
	return nil
}

func (p *PostgresDBImpl) QueryBulkSearchOffersPaginated(ctx context.Context, searchID, limit, offset int) (Rows, error) {
	return p.db.QueryContext(ctx,
		`SELECT origin, destination, departure_date, return_date, price, currency,
			airline_codes, src_airport_code, dst_airport_code, src_city, dst_city,
			flight_duration, return_flight_duration, distance_miles, cost_per_mile,
			outbound_flights, return_flights, offer_json, created_at
		 FROM bulk_search_offers
		 WHERE bulk_search_id = $1
		 ORDER BY price ASC
		 LIMIT $2 OFFSET $3`,
		searchID, limit, offset,
	)
}

func (p *PostgresDBImpl) ListBulkSearchOffers(ctx context.Context, searchID int) ([]BulkSearchOffer, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, bulk_search_id, origin, destination, departure_date, return_date, price, currency,
			airline_codes, src_airport_code, dst_airport_code, src_city, dst_city,
			flight_duration, return_flight_duration, distance_miles, cost_per_mile,
			outbound_flights, return_flights, offer_json, created_at
		 FROM bulk_search_offers
		 WHERE bulk_search_id = $1
		 ORDER BY origin, destination, departure_date, return_date NULLS FIRST, price ASC, created_at ASC`,
		searchID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query bulk search offers: %w", err)
	}
	defer rows.Close()

	var offers []BulkSearchOffer
	for rows.Next() {
		var (
			offer        BulkSearchOffer
			airlineCodes pq.StringArray
		)

		if scanErr := rows.Scan(
			&offer.ID,
			&offer.BulkSearchID,
			&offer.Origin,
			&offer.Destination,
			&offer.DepartureDate,
			&offer.ReturnDate,
			&offer.Price,
			&offer.Currency,
			&airlineCodes,
			&offer.SrcAirportCode,
			&offer.DstAirportCode,
			&offer.SrcCity,
			&offer.DstCity,
			&offer.FlightDuration,
			&offer.ReturnFlightDuration,
			&offer.DistanceMiles,
			&offer.CostPerMile,
			&offer.OutboundFlights,
			&offer.ReturnFlights,
			&offer.OfferJSON,
			&offer.CreatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan bulk search offer: %w", scanErr)
		}

		offer.AirlineCodes = append([]string(nil), airlineCodes...)
		offers = append(offers, offer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bulk search offers: %w", err)
	}

	return offers, nil
}

func (p *PostgresDBImpl) InsertBulkSearchResult(ctx context.Context, result BulkSearchResultRecord) error {
	outboundJSON := result.OutboundFlightsJSON
	if len(outboundJSON) == 0 {
		outboundJSON = []byte("[]")
	}
	returnJSON := result.ReturnFlightsJSON
	if len(returnJSON) == 0 {
		returnJSON = []byte("[]")
	}
	offerJSON := result.OfferJSON
	if len(offerJSON) == 0 {
		offerJSON = []byte("{}")
	}

	_, err := p.db.ExecContext(ctx,
		`INSERT INTO bulk_search_results
			(bulk_search_id, origin, destination, departure_date, return_date,
			 price, currency, airline_code, duration,
			 src_airport_code, dst_airport_code, src_city, dst_city,
			 flight_duration, return_flight_duration,
			 outbound_flights, return_flights, offer_json)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,
			 $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		result.BulkSearchID,
		result.Origin,
		result.Destination,
		result.DepartureDate,
		result.ReturnDate,
		result.Price,
		result.Currency,
		result.AirlineCode,
		result.Duration,
		result.SrcAirportCode,
		result.DstAirportCode,
		result.SrcCity,
		result.DstCity,
		result.FlightDuration,
		result.ReturnFlightDuration,
		outboundJSON,
		returnJSON,
		offerJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to insert bulk search result: %w", err)
	}
	return nil
}

// InsertBulkSearchResultsBatch inserts multiple bulk search results in a single transaction
// using multi-row INSERT for better performance. This reduces round-trips to the database.
func (p *PostgresDBImpl) InsertBulkSearchResultsBatch(ctx context.Context, results []BulkSearchResultRecord) error {
	if len(results) == 0 {
		return nil
	}

	// Start a transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for batch insert: %w", err)
	}
	defer tx.Rollback()

	// Build multi-row INSERT statement
	// 18 fields per row
	const numFields = 18
	valueStrings := make([]string, 0, len(results))
	valueArgs := make([]interface{}, 0, len(results)*numFields)

	for i, result := range results {
		base := i * numFields

		outboundJSON := result.OutboundFlightsJSON
		if len(outboundJSON) == 0 {
			outboundJSON = []byte("[]")
		}
		returnJSON := result.ReturnFlightsJSON
		if len(returnJSON) == 0 {
			returnJSON = []byte("[]")
		}
		offerJSON := result.OfferJSON
		if len(offerJSON) == 0 {
			offerJSON = []byte("{}")
		}

		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9,
			base+10, base+11, base+12, base+13, base+14, base+15, base+16, base+17, base+18,
		))
		valueArgs = append(valueArgs,
			result.BulkSearchID,
			result.Origin,
			result.Destination,
			result.DepartureDate,
			result.ReturnDate,
			result.Price,
			result.Currency,
			result.AirlineCode,
			result.Duration,
			result.SrcAirportCode,
			result.DstAirportCode,
			result.SrcCity,
			result.DstCity,
			result.FlightDuration,
			result.ReturnFlightDuration,
			outboundJSON,
			returnJSON,
			offerJSON,
		)
	}

	query := `INSERT INTO bulk_search_results
		(bulk_search_id, origin, destination, departure_date, return_date,
		 price, currency, airline_code, duration,
		 src_airport_code, dst_airport_code, src_city, dst_city,
		 flight_duration, return_flight_duration,
		 outbound_flights, return_flights, offer_json)
	VALUES ` + strings.Join(valueStrings, ",")

	_, err = tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch insert: %w", err)
	}

	return nil
}

func (p *PostgresDBImpl) ListBulkSearches(ctx context.Context, limit, offset int) (Rows, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return p.db.QueryContext(ctx,
		`SELECT id, job_id, status, total_searches, completed, total_offers,
            error_count, currency, created_at, updated_at, completed_at,
            min_price, max_price, average_price
         FROM bulk_searches
         ORDER BY created_at DESC
         LIMIT $1 OFFSET $2`,
		limit, offset,
	)
}

func (p *PostgresDBImpl) CreatePriceGraphSweep(ctx context.Context, jobID sql.NullInt32, originCount, destinationCount int, tripLengthMin, tripLengthMax sql.NullInt32, currency string) (int, error) {
	if currency == "" {
		currency = "USD"
	}
	var sweepID int
	err := p.db.QueryRowContext(ctx,
		`INSERT INTO price_graph_sweeps
			(job_id, origin_count, destination_count, trip_length_min, trip_length_max, currency)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		jobID, originCount, destinationCount, tripLengthMin, tripLengthMax, currency,
	).Scan(&sweepID)
	if err != nil {
		return 0, fmt.Errorf("failed to create price graph sweep: %w", err)
	}
	return sweepID, nil
}

func (p *PostgresDBImpl) UpdatePriceGraphSweepStatus(ctx context.Context, sweepID int, status string, startedAt, completedAt sql.NullTime, errorCount int) error {
	var started interface{}
	if startedAt.Valid {
		started = startedAt.Time
	}
	var completed interface{}
	if completedAt.Valid {
		completed = completedAt.Time
	}

	_, err := p.db.ExecContext(ctx,
		`UPDATE price_graph_sweeps
         SET status = $1,
             error_count = $2,
             started_at = COALESCE($3, started_at),
             completed_at = COALESCE($4, completed_at),
             updated_at = NOW()
         WHERE id = $5`,
		status, errorCount, started, completed, sweepID,
	)
	if err != nil {
		return fmt.Errorf("failed to update price graph sweep %d status: %w", sweepID, err)
	}
	return nil
}

func (p *PostgresDBImpl) GetPriceGraphSweepByID(ctx context.Context, sweepID int) (*PriceGraphSweep, error) {
	var sweep PriceGraphSweep
	err := p.db.QueryRowContext(ctx,
		`SELECT id, job_id, status, origin_count, destination_count,
                trip_length_min, trip_length_max, currency, error_count,
                created_at, updated_at, started_at, completed_at
         FROM price_graph_sweeps
         WHERE id = $1`,
		sweepID,
	).Scan(&sweep.ID, &sweep.JobID, &sweep.Status, &sweep.OriginCount, &sweep.DestinationCount,
		&sweep.TripLengthMin, &sweep.TripLengthMax, &sweep.Currency, &sweep.ErrorCount,
		&sweep.CreatedAt, &sweep.UpdatedAt, &sweep.StartedAt, &sweep.CompletedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("price graph sweep with ID %d not found", sweepID)
		}
		return nil, fmt.Errorf("error getting price graph sweep %d: %w", sweepID, err)
	}
	return &sweep, nil
}

func (p *PostgresDBImpl) ListPriceGraphSweeps(ctx context.Context, limit, offset int) (Rows, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return p.db.QueryContext(ctx,
		`SELECT id, job_id, status, origin_count, destination_count,
                trip_length_min, trip_length_max, currency, error_count,
                created_at, updated_at, started_at, completed_at
         FROM price_graph_sweeps
         ORDER BY created_at DESC
         LIMIT $1 OFFSET $2`,
		limit, offset,
	)
}

func (p *PostgresDBImpl) InsertPriceGraphResult(ctx context.Context, record PriceGraphResultRecord) error {
	if record.QueriedAt.IsZero() {
		record.QueriedAt = time.Now()
	}

	// Ensure TripLength participates in upsert key (avoid NULL breaking ON CONFLICT).
	if !record.TripLength.Valid {
		record.TripLength = sql.NullInt32{Int32: 0, Valid: true}
	}

	_, err := p.db.ExecContext(ctx,
		`INSERT INTO price_graph_results
				(sweep_id, origin, destination, departure_date, return_date,
				 trip_length, price, currency, distance_miles, cost_per_mile,
				 adults, children, infants_lap, infants_seat, trip_type, class, stops, search_url,
				 queried_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
			 ON CONFLICT (sweep_id, origin, destination, departure_date, trip_length, currency, adults, children, infants_lap, infants_seat, trip_type, class, stops)
			 DO UPDATE SET
				return_date = EXCLUDED.return_date,
				price = EXCLUDED.price,
				distance_miles = COALESCE(EXCLUDED.distance_miles, price_graph_results.distance_miles),
				cost_per_mile = COALESCE(EXCLUDED.cost_per_mile, price_graph_results.cost_per_mile),
				search_url = COALESCE(EXCLUDED.search_url, price_graph_results.search_url),
				queried_at = EXCLUDED.queried_at`,
		record.SweepID,
		record.Origin,
		record.Destination,
		record.DepartureDate,
		record.ReturnDate,
		record.TripLength,
		record.Price,
		record.Currency,
		record.DistanceMiles,
		record.CostPerMile,
		record.Adults,
		record.Children,
		record.InfantsLap,
		record.InfantsSeat,
		record.TripType,
		record.Class,
		record.Stops,
		record.SearchURL,
		record.QueriedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert price graph result for %s -> %s: %w", record.Origin, record.Destination, err)
	}
	return nil
}

func (p *PostgresDBImpl) ListPriceGraphResults(ctx context.Context, sweepID, limit, offset int) (Rows, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return p.db.QueryContext(ctx,
		`SELECT id, sweep_id, origin, destination, departure_date,
	                return_date, trip_length, price, currency,
					adults, children, infants_lap, infants_seat, trip_type, class, stops, search_url,
					queried_at, created_at
	         FROM price_graph_results
	         WHERE sweep_id = $1
	         ORDER BY price ASC
	         LIMIT $2 OFFSET $3`,
		sweepID, limit, offset,
	)
}

// SaveContinuousSweepProgress saves or updates the continuous sweep progress (upsert)
func (p *PostgresDBImpl) SaveContinuousSweepProgress(ctx context.Context, progress ContinuousSweepProgress) error {
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO continuous_sweep_progress
			(id, sweep_number, route_index, total_routes, current_origin, current_destination,
			 queries_completed, errors_count, last_error, sweep_started_at, last_updated,
			 trip_lengths, pacing_mode, target_duration_hours, min_delay_ms, is_running, is_paused, international_only)
		 VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), $10, $11, $12, $13, $14, $15, $16)
		 ON CONFLICT (id) DO UPDATE SET
			sweep_number = $1,
			route_index = $2,
			total_routes = $3,
			current_origin = $4,
			current_destination = $5,
			queries_completed = $6,
			errors_count = $7,
			last_error = $8,
			sweep_started_at = $9,
			last_updated = NOW(),
			trip_lengths = $10,
			pacing_mode = $11,
			target_duration_hours = $12,
			min_delay_ms = $13,
			-- Control flags are managed via SetContinuousSweepControlFlags and must not be overwritten
			-- by periodic progress saves (otherwise STOP/PAUSE can be clobbered by a running worker).
			is_running = continuous_sweep_progress.is_running,
			is_paused = continuous_sweep_progress.is_paused,
			international_only = $16`,
		progress.SweepNumber,
		progress.RouteIndex,
		progress.TotalRoutes,
		progress.CurrentOrigin,
		progress.CurrentDestination,
		progress.QueriesCompleted,
		progress.ErrorsCount,
		progress.LastError,
		progress.SweepStartedAt,
		pq.Array(progress.TripLengths),
		progress.PacingMode,
		progress.TargetDurationHours,
		progress.MinDelayMs,
		progress.IsRunning,
		progress.IsPaused,
		progress.InternationalOnly,
	)
	if err != nil {
		return fmt.Errorf("failed to save continuous sweep progress: %w", err)
	}
	return nil
}

func (p *PostgresDBImpl) SetContinuousSweepControlFlags(ctx context.Context, isRunning, isPaused *bool) error {
	var running sql.NullBool
	var paused sql.NullBool
	if isRunning != nil {
		running = sql.NullBool{Bool: *isRunning, Valid: true}
	}
	if isPaused != nil {
		paused = sql.NullBool{Bool: *isPaused, Valid: true}
	}

	_, err := p.db.ExecContext(ctx,
		`UPDATE continuous_sweep_progress
		 SET is_running = COALESCE($1, is_running),
		     is_paused = COALESCE($2, is_paused),
		     last_updated = NOW()
		 WHERE id = 1`,
		running,
		paused,
	)
	if err != nil {
		return fmt.Errorf("failed to set continuous sweep control flags: %w", err)
	}
	return nil
}

// GetContinuousSweepProgress retrieves the current continuous sweep progress
func (p *PostgresDBImpl) GetContinuousSweepProgress(ctx context.Context) (*ContinuousSweepProgress, error) {
	var progress ContinuousSweepProgress
	var tripLengths pq.Int64Array
	err := p.db.QueryRowContext(ctx,
		`SELECT id, sweep_number, route_index, total_routes, current_origin, current_destination,
		        queries_completed, errors_count, last_error, sweep_started_at, last_updated,
		        COALESCE(trip_lengths, '{7,14}'), pacing_mode, target_duration_hours, min_delay_ms, is_running, is_paused,
		        COALESCE(international_only, TRUE)
		 FROM continuous_sweep_progress
		 WHERE id = 1`,
	).Scan(
		&progress.ID,
		&progress.SweepNumber,
		&progress.RouteIndex,
		&progress.TotalRoutes,
		&progress.CurrentOrigin,
		&progress.CurrentDestination,
		&progress.QueriesCompleted,
		&progress.ErrorsCount,
		&progress.LastError,
		&progress.SweepStartedAt,
		&progress.LastUpdated,
		&tripLengths,
		&progress.PacingMode,
		&progress.TargetDurationHours,
		&progress.MinDelayMs,
		&progress.IsRunning,
		&progress.IsPaused,
		&progress.InternationalOnly,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return a default progress if none exists
			return &ContinuousSweepProgress{
				ID:                  1,
				SweepNumber:         1,
				RouteIndex:          0,
				TotalRoutes:         0,
				PacingMode:          "adaptive",
				TargetDurationHours: 24,
				MinDelayMs:          3000,
				IsRunning:           false,
				IsPaused:            false,
				InternationalOnly:   true,
				TripLengths:         []int{7, 14},
			}, nil
		}
		return nil, fmt.Errorf("failed to get continuous sweep progress: %w", err)
	}

	progress.TripLengths = make([]int, len(tripLengths))
	for i, v := range tripLengths {
		progress.TripLengths[i] = int(v)
	}

	return &progress, nil
}

// InsertContinuousSweepStats inserts a completed sweep's statistics
func (p *PostgresDBImpl) InsertContinuousSweepStats(ctx context.Context, stats ContinuousSweepStats) error {
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO continuous_sweep_stats
			(sweep_number, started_at, completed_at, total_routes, successful_queries,
			 failed_queries, total_duration_seconds, avg_delay_ms, min_price_found, max_price_found)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		stats.SweepNumber,
		stats.StartedAt,
		stats.CompletedAt,
		stats.TotalRoutes,
		stats.SuccessfulQueries,
		stats.FailedQueries,
		stats.TotalDurationSeconds,
		stats.AvgDelayMs,
		stats.MinPriceFound,
		stats.MaxPriceFound,
	)
	if err != nil {
		return fmt.Errorf("failed to insert continuous sweep stats: %w", err)
	}
	return nil
}

// ListContinuousSweepStats retrieves the most recent sweep statistics
func (p *PostgresDBImpl) ListContinuousSweepStats(ctx context.Context, limit int) ([]ContinuousSweepStats, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, sweep_number, started_at, completed_at, total_routes, successful_queries,
		        failed_queries, total_duration_seconds, avg_delay_ms, min_price_found, max_price_found, created_at
		 FROM continuous_sweep_stats
		 ORDER BY created_at DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list continuous sweep stats: %w", err)
	}
	defer rows.Close()

	var statsList []ContinuousSweepStats
	for rows.Next() {
		var stats ContinuousSweepStats
		if err := rows.Scan(
			&stats.ID,
			&stats.SweepNumber,
			&stats.StartedAt,
			&stats.CompletedAt,
			&stats.TotalRoutes,
			&stats.SuccessfulQueries,
			&stats.FailedQueries,
			&stats.TotalDurationSeconds,
			&stats.AvgDelayMs,
			&stats.MinPriceFound,
			&stats.MaxPriceFound,
			&stats.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan continuous sweep stats: %w", err)
		}
		statsList = append(statsList, stats)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating continuous sweep stats: %w", err)
	}

	return statsList, nil
}

// ListContinuousSweepResults returns price graph results from continuous sweeps (sweep_id = 0)
func (p *PostgresDBImpl) ListContinuousSweepResults(ctx context.Context, filters ContinuousSweepResultsFilter) ([]PriceGraphResultRecord, error) {
	if filters.Limit <= 0 {
		filters.Limit = 100
	}
	if filters.Limit > 1000 {
		filters.Limit = 1000
	}

	// Build dynamic query with filters
	query := `SELECT sweep_id, origin, destination, departure_date,
	                 return_date, trip_length, price, currency,
					 adults, children, infants_lap, infants_seat, trip_type, class, stops, search_url,
					 queried_at
	          FROM price_graph_results
	          WHERE sweep_id = 0`
	args := []interface{}{}
	argIdx := 1

	if filters.Origin != "" {
		query += fmt.Sprintf(" AND origin = $%d", argIdx)
		args = append(args, filters.Origin)
		argIdx++
	}
	if filters.Destination != "" {
		query += fmt.Sprintf(" AND destination = $%d", argIdx)
		args = append(args, filters.Destination)
		argIdx++
	}
	if !filters.FromDate.IsZero() {
		query += fmt.Sprintf(" AND departure_date >= $%d", argIdx)
		args = append(args, filters.FromDate)
		argIdx++
	}
	if !filters.ToDate.IsZero() {
		query += fmt.Sprintf(" AND departure_date <= $%d", argIdx)
		args = append(args, filters.ToDate)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY queried_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filters.Limit, filters.Offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list continuous sweep results: %w", err)
	}
	defer rows.Close()

	var results []PriceGraphResultRecord
	for rows.Next() {
		var r PriceGraphResultRecord
		if err := rows.Scan(
			&r.SweepID,
			&r.Origin,
			&r.Destination,
			&r.DepartureDate,
			&r.ReturnDate,
			&r.TripLength,
			&r.Price,
			&r.Currency,
			&r.Adults,
			&r.Children,
			&r.InfantsLap,
			&r.InfantsSeat,
			&r.TripType,
			&r.Class,
			&r.Stops,
			&r.SearchURL,
			&r.QueriedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan continuous sweep result: %w", err)
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating continuous sweep results: %w", err)
	}

	return results, nil
}

// --- Deal Detection Methods ---

// GetRouteBaseline retrieves price baseline for a route
func (p *PostgresDBImpl) GetRouteBaseline(ctx context.Context, origin, dest string, tripLength int, class string) (*RouteBaseline, error) {
	var baseline RouteBaseline
	err := p.db.QueryRowContext(ctx,
		`SELECT id, origin, destination, trip_length, class, sample_count,
		        mean_price, median_price, stddev_price, min_price, max_price,
		        p10_price, p25_price, p75_price, p90_price,
		        window_start, window_end, updated_at, created_at
		 FROM route_baselines
		 WHERE origin = $1 AND destination = $2 AND trip_length = $3 AND class = $4`,
		origin, dest, tripLength, class,
	).Scan(
		&baseline.ID, &baseline.Origin, &baseline.Destination, &baseline.TripLength, &baseline.Class,
		&baseline.SampleCount, &baseline.MeanPrice, &baseline.MedianPrice, &baseline.StddevPrice,
		&baseline.MinPrice, &baseline.MaxPrice, &baseline.P10Price, &baseline.P25Price,
		&baseline.P75Price, &baseline.P90Price, &baseline.WindowStart, &baseline.WindowEnd,
		&baseline.UpdatedAt, &baseline.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get route baseline: %w", err)
	}
	return &baseline, nil
}

// UpsertRouteBaseline inserts or updates a route baseline
func (p *PostgresDBImpl) UpsertRouteBaseline(ctx context.Context, baseline RouteBaseline) error {
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO route_baselines (origin, destination, trip_length, class, sample_count,
		                              mean_price, median_price, stddev_price, min_price, max_price,
		                              p10_price, p25_price, p75_price, p90_price,
		                              window_start, window_end, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW())
		 ON CONFLICT (origin, destination, trip_length, class) DO UPDATE SET
		    sample_count = $5, mean_price = $6, median_price = $7, stddev_price = $8,
		    min_price = $9, max_price = $10, p10_price = $11, p25_price = $12,
		    p75_price = $13, p90_price = $14, window_start = $15, window_end = $16, updated_at = NOW()`,
		baseline.Origin, baseline.Destination, baseline.TripLength, baseline.Class,
		baseline.SampleCount, baseline.MeanPrice, baseline.MedianPrice, baseline.StddevPrice,
		baseline.MinPrice, baseline.MaxPrice, baseline.P10Price, baseline.P25Price,
		baseline.P75Price, baseline.P90Price, baseline.WindowStart, baseline.WindowEnd,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert route baseline: %w", err)
	}
	return nil
}

// GetPriceHistoryForRoute retrieves historical prices for baseline calculation
func (p *PostgresDBImpl) GetPriceHistoryForRoute(ctx context.Context, origin, dest string, tripLength int, class string, windowDays int) ([]float64, error) {
	query := `SELECT price FROM price_graph_results
		 WHERE origin = $1 AND destination = $2 
		   AND (trip_length = $3 OR $3 = 0)
		   AND class = $4`
	args := []interface{}{origin, dest, tripLength, class}
	argIdx := 5

	if windowDays > 0 {
		query += fmt.Sprintf(" AND queried_at >= NOW() - INTERVAL '1 day' * $%d", argIdx)
		args = append(args, windowDays)
		argIdx++
	}

	query += " AND price > 0 ORDER BY queried_at DESC"

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}
	defer rows.Close()

	var prices []float64
	for rows.Next() {
		var price float64
		if err := rows.Scan(&price); err != nil {
			return nil, fmt.Errorf("failed to scan price: %w", err)
		}
		prices = append(prices, price)
	}
	return prices, nil
}

// InsertDetectedDeal inserts a new detected deal
func (p *PostgresDBImpl) InsertDetectedDeal(ctx context.Context, deal DetectedDeal) (int, error) {
	var id int
	err := p.db.QueryRowContext(ctx,
		`INSERT INTO detected_deals (origin, destination, departure_date, return_date, trip_length,
		                             price, currency, baseline_mean, baseline_median, discount_percent,
		                             deal_score, deal_classification, distance_miles, cost_per_mile,
		                             cabin_class, source_type, source_id, search_url, deal_fingerprint,
		                             first_seen_at, last_seen_at, times_seen, status, verified,
		                             verified_price, verified_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
		         $20, $21, $22, $23, $24, $25, $26, $27)
		 RETURNING id`,
		deal.Origin, deal.Destination, deal.DepartureDate, deal.ReturnDate, deal.TripLength,
		deal.Price, deal.Currency, deal.BaselineMean, deal.BaselineMedian, deal.DiscountPercent,
		deal.DealScore, deal.DealClassification, deal.DistanceMiles, deal.CostPerMile,
		deal.CabinClass, deal.SourceType, deal.SourceID, deal.SearchURL, deal.DealFingerprint,
		deal.FirstSeenAt, deal.LastSeenAt, deal.TimesSeen, deal.Status, deal.Verified,
		deal.VerifiedPrice, deal.VerifiedAt, deal.ExpiresAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert detected deal: %w", err)
	}
	return id, nil
}

// UpsertDetectedDeal inserts or updates a deal (for deduplication)
func (p *PostgresDBImpl) UpsertDetectedDeal(ctx context.Context, deal DetectedDeal) error {
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO detected_deals (origin, destination, departure_date, return_date, trip_length,
		                             price, currency, baseline_mean, baseline_median, discount_percent,
		                             deal_score, deal_classification, distance_miles, cost_per_mile,
		                             cabin_class, source_type, source_id, search_url, deal_fingerprint,
		                             first_seen_at, last_seen_at, times_seen, status, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
		         NOW(), NOW(), 1, $20, $21)
		 ON CONFLICT (deal_fingerprint) DO UPDATE SET
		    last_seen_at = NOW(),
		    times_seen = detected_deals.times_seen + 1,
		    price = LEAST(detected_deals.price, $6),
		    deal_score = GREATEST(detected_deals.deal_score, $11),
		    updated_at = NOW()`,
		deal.Origin, deal.Destination, deal.DepartureDate, deal.ReturnDate, deal.TripLength,
		deal.Price, deal.Currency, deal.BaselineMean, deal.BaselineMedian, deal.DiscountPercent,
		deal.DealScore, deal.DealClassification, deal.DistanceMiles, deal.CostPerMile,
		deal.CabinClass, deal.SourceType, deal.SourceID, deal.SearchURL, deal.DealFingerprint,
		deal.Status, deal.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert detected deal: %w", err)
	}
	return nil
}

// GetDetectedDealByFingerprint retrieves a deal by its fingerprint
func (p *PostgresDBImpl) GetDetectedDealByFingerprint(ctx context.Context, fingerprint string) (*DetectedDeal, error) {
	var deal DetectedDeal
	err := p.db.QueryRowContext(ctx,
		`SELECT id, origin, destination, departure_date, return_date, trip_length,
		        price, currency, baseline_mean, baseline_median, discount_percent,
		        deal_score, deal_classification, distance_miles, cost_per_mile,
		        cabin_class, source_type, source_id, search_url, deal_fingerprint,
		        first_seen_at, last_seen_at, times_seen, status, verified,
		        verified_price, verified_at, expires_at, created_at, updated_at
		 FROM detected_deals WHERE deal_fingerprint = $1`,
		fingerprint,
	).Scan(
		&deal.ID, &deal.Origin, &deal.Destination, &deal.DepartureDate, &deal.ReturnDate,
		&deal.TripLength, &deal.Price, &deal.Currency, &deal.BaselineMean, &deal.BaselineMedian,
		&deal.DiscountPercent, &deal.DealScore, &deal.DealClassification, &deal.DistanceMiles,
		&deal.CostPerMile, &deal.CabinClass, &deal.SourceType, &deal.SourceID, &deal.SearchURL,
		&deal.DealFingerprint, &deal.FirstSeenAt, &deal.LastSeenAt, &deal.TimesSeen, &deal.Status,
		&deal.Verified, &deal.VerifiedPrice, &deal.VerifiedAt, &deal.ExpiresAt, &deal.CreatedAt, &deal.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get deal by fingerprint: %w", err)
	}
	return &deal, nil
}

// ListActiveDeals retrieves active deals with optional filtering
func (p *PostgresDBImpl) ListActiveDeals(ctx context.Context, filter DealFilter) ([]DetectedDeal, error) {
	query := `SELECT id, origin, destination, departure_date, return_date, trip_length,
	                 price, currency, baseline_mean, baseline_median, discount_percent,
	                 deal_score, deal_classification, distance_miles, cost_per_mile,
	                 cabin_class, source_type, source_id, search_url, deal_fingerprint,
	                 first_seen_at, last_seen_at, times_seen, status, verified,
	                 verified_price, verified_at, expires_at, created_at, updated_at
	          FROM detected_deals WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if !filter.IncludeExpired {
		query += fmt.Sprintf(" AND (status = 'active' AND (expires_at IS NULL OR expires_at > NOW()))")
	}
	if filter.Origin != "" {
		query += fmt.Sprintf(" AND origin = $%d", argIdx)
		args = append(args, filter.Origin)
		argIdx++
	}
	if filter.Destination != "" {
		query += fmt.Sprintf(" AND destination = $%d", argIdx)
		args = append(args, filter.Destination)
		argIdx++
	}
	if filter.MinScore > 0 {
		query += fmt.Sprintf(" AND deal_score >= $%d", argIdx)
		args = append(args, filter.MinScore)
		argIdx++
	}
	if filter.Classification != "" {
		query += fmt.Sprintf(" AND deal_classification = $%d", argIdx)
		args = append(args, filter.Classification)
		argIdx++
	}

	query += " ORDER BY deal_score DESC, first_seen_at DESC"

	if filter.Limit <= 0 {
		filter.Limit = 100
	}
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list active deals: %w", err)
	}
	defer rows.Close()

	var deals []DetectedDeal
	for rows.Next() {
		var deal DetectedDeal
		if err := rows.Scan(
			&deal.ID, &deal.Origin, &deal.Destination, &deal.DepartureDate, &deal.ReturnDate,
			&deal.TripLength, &deal.Price, &deal.Currency, &deal.BaselineMean, &deal.BaselineMedian,
			&deal.DiscountPercent, &deal.DealScore, &deal.DealClassification, &deal.DistanceMiles,
			&deal.CostPerMile, &deal.CabinClass, &deal.SourceType, &deal.SourceID, &deal.SearchURL,
			&deal.DealFingerprint, &deal.FirstSeenAt, &deal.LastSeenAt, &deal.TimesSeen, &deal.Status,
			&deal.Verified, &deal.VerifiedPrice, &deal.VerifiedAt, &deal.ExpiresAt, &deal.CreatedAt, &deal.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan deal: %w", err)
		}
		deals = append(deals, deal)
	}
	return deals, nil
}

// ExpireOldDeals marks expired deals
func (p *PostgresDBImpl) ExpireOldDeals(ctx context.Context) (int64, error) {
	result, err := p.db.ExecContext(ctx,
		`UPDATE detected_deals SET status = 'expired', updated_at = NOW()
		 WHERE status = 'active' AND expires_at IS NOT NULL AND expires_at < NOW()`,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to expire old deals: %w", err)
	}
	return result.RowsAffected()
}

// InsertDealAlert inserts a new deal alert
func (p *PostgresDBImpl) InsertDealAlert(ctx context.Context, alert DealAlert) (int, error) {
	var id int
	err := p.db.QueryRowContext(ctx,
		`INSERT INTO deal_alerts (detected_deal_id, origin, destination, price, currency,
		                          discount_percent, deal_classification, deal_score,
		                          published_at, publish_method, notification_channels)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), $9, $10)
		 RETURNING id`,
		alert.DetectedDealID, alert.Origin, alert.Destination, alert.Price, alert.Currency,
		alert.DiscountPercent, alert.DealClassification, alert.DealScore,
		alert.PublishMethod, pq.Array(alert.NotificationChannels),
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert deal alert: %w", err)
	}
	return id, nil
}

// ListDealAlerts retrieves published deal alerts with pagination
func (p *PostgresDBImpl) ListDealAlerts(ctx context.Context, limit, offset int) ([]DealAlert, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, detected_deal_id, origin, destination, price, currency,
		        discount_percent, deal_classification, deal_score,
		        published_at, publish_method, notification_sent, notification_sent_at,
		        notification_channels, created_at
		 FROM deal_alerts
		 ORDER BY published_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query deal alerts: %w", err)
	}
	defer rows.Close()

	var alerts []DealAlert
	for rows.Next() {
		var alert DealAlert
		var channels []string
		err := rows.Scan(
			&alert.ID, &alert.DetectedDealID, &alert.Origin, &alert.Destination,
			&alert.Price, &alert.Currency, &alert.DiscountPercent, &alert.DealClassification,
			&alert.DealScore, &alert.PublishedAt, &alert.PublishMethod, &alert.NotificationSent,
			&alert.NotificationSentAt, pq.Array(&channels), &alert.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deal alert row: %w", err)
		}
		alert.NotificationChannels = channels
		alerts = append(alerts, alert)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deal alert rows: %w", err)
	}
	return alerts, nil
}

// --- End Implementation ---

// GetDB returns the underlying database connection
// func (p *PostgresDBImpl) GetDB() *sql.DB {
// 	return p.db
// }

// InitSchema initializes the database schema
func (p *PostgresDBImpl) InitSchema() error {
	// Create/ensure tables in the correct order (non-destructive).
	// NOTE: This function used to DROP tables; that was too dangerous for production.
	// If you need a clean slate, wipe the DB/volume explicitly or introduce migrations.
	var err error

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
            price DECIMAL(12, 2) NOT NULL,
            currency VARCHAR(3) NOT NULL,
            cabin_class VARCHAR(20) NOT NULL,
            search_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create flight_prices table: %w", err)
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
	if err != nil {
		return fmt.Errorf("failed to create scheduled_jobs table: %w", err)
	}

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
            dynamic_dates BOOLEAN DEFAULT FALSE,
            days_from_execution INTEGER,
            search_window_days INTEGER,
            adults INTEGER DEFAULT 1,
            children INTEGER DEFAULT 0,
            infants_lap INTEGER DEFAULT 0,
            infants_seat INTEGER DEFAULT 0,
            trip_type VARCHAR(20) DEFAULT 'round_trip',
            class VARCHAR(20) DEFAULT 'economy',
            stops VARCHAR(20) DEFAULT 'any',
            currency VARCHAR(3) DEFAULT 'USD',
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
            min_price DECIMAL(12, 2),
            max_price DECIMAL(12, 2),
            search_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            search_query_id INTEGER REFERENCES search_queries(id)
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
            price DECIMAL(12, 2) NOT NULL,
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

	// Create flight_segments table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS flight_segments (
            id SERIAL PRIMARY KEY,
            flight_offer_id INTEGER NOT NULL REFERENCES flight_offers(id) ON DELETE CASCADE,
            airline_code VARCHAR(3) NOT NULL,
            flight_number VARCHAR(10) NOT NULL,
            departure_airport VARCHAR(3) NOT NULL,
            arrival_airport VARCHAR(3) NOT NULL,
            departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
            arrival_time TIMESTAMP WITH TIME ZONE NOT NULL,
            duration INTEGER NOT NULL,
            airplane VARCHAR(100),
            legroom VARCHAR(50),
            is_return BOOLEAN NOT NULL DEFAULT FALSE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create flight_segments table: %w", err)
	}

	// Create bulk_searches table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS bulk_searches (
            id SERIAL PRIMARY KEY,
            job_id INTEGER REFERENCES scheduled_jobs(id),
            status VARCHAR(30) NOT NULL,
            total_searches INTEGER NOT NULL,
            completed INTEGER NOT NULL DEFAULT 0,
            total_offers INTEGER NOT NULL DEFAULT 0,
            error_count INTEGER NOT NULL DEFAULT 0,
            currency VARCHAR(3) NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            completed_at TIMESTAMP WITH TIME ZONE,
            min_price DECIMAL(12, 2),
			max_price DECIMAL(12, 2),
			average_price DECIMAL(12, 2)
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create bulk_searches table: %w", err)
	}

	// Create bulk_search_results table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS bulk_search_results (
            id SERIAL PRIMARY KEY,
            bulk_search_id INTEGER REFERENCES bulk_searches(id) ON DELETE CASCADE,
            origin VARCHAR(3) NOT NULL,
            destination VARCHAR(3) NOT NULL,
            departure_date DATE NOT NULL,
            return_date DATE,
            price DECIMAL(12, 2) NOT NULL,
            currency VARCHAR(3) NOT NULL,
            airline_code VARCHAR(3),
            duration INTEGER,
            src_airport_code VARCHAR(3),
            dst_airport_code VARCHAR(3),
            src_city TEXT,
            dst_city TEXT,
            flight_duration INTEGER,
            return_flight_duration INTEGER,
            outbound_flights JSONB,
            return_flights JSONB,
            offer_json JSONB,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create bulk_search_results table: %w", err)
	}

	// Create bulk_search_offers table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS bulk_search_offers (
            id SERIAL PRIMARY KEY,
            bulk_search_id INTEGER REFERENCES bulk_searches(id) ON DELETE CASCADE,
            origin VARCHAR(3) NOT NULL,
            destination VARCHAR(3) NOT NULL,
            departure_date DATE NOT NULL,
            return_date DATE,
            price DECIMAL(12, 2) NOT NULL,
            currency VARCHAR(3) NOT NULL,
            airline_codes TEXT[],
            src_airport_code VARCHAR(3),
            dst_airport_code VARCHAR(3),
            src_city TEXT,
            dst_city TEXT,
            flight_duration INTEGER,
            return_flight_duration INTEGER,
            outbound_flights JSONB,
            return_flights JSONB,
            offer_json JSONB,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create bulk_search_offers table: %w", err)
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

	// Create price_graph_sweeps table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS price_graph_sweeps (
            id SERIAL PRIMARY KEY,
            job_id INTEGER REFERENCES scheduled_jobs(id),
            status VARCHAR(32) NOT NULL DEFAULT 'queued',
            origin_count INTEGER NOT NULL DEFAULT 0,
            destination_count INTEGER NOT NULL DEFAULT 0,
            trip_length_min INTEGER,
            trip_length_max INTEGER,
            currency VARCHAR(3) NOT NULL DEFAULT 'USD',
            error_count INTEGER NOT NULL DEFAULT 0,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
            started_at TIMESTAMP WITH TIME ZONE,
            completed_at TIMESTAMP WITH TIME ZONE
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create price_graph_sweeps table: %w", err)
	}

	// Ensure sweep_id=0 exists for continuous sweeps (older code uses sweep_id=0 sentinel).
	// This avoids FK violations when continuous sweeps store results in price_graph_results.
	_, err = p.db.Exec(`
        INSERT INTO price_graph_sweeps (id, job_id, status, origin_count, destination_count, currency, error_count)
        VALUES (0, NULL, 'continuous', 0, 0, 'USD', 0)
        ON CONFLICT (id) DO NOTHING
    `)
	if err != nil {
		return fmt.Errorf("failed to ensure continuous sweep row (id=0): %w", err)
	}

	// Create price_graph_results table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS price_graph_results (
            id SERIAL PRIMARY KEY,
            sweep_id INTEGER NOT NULL REFERENCES price_graph_sweeps(id) ON DELETE CASCADE,
            origin VARCHAR(3) NOT NULL,
            destination VARCHAR(3) NOT NULL,
            departure_date DATE NOT NULL,
            return_date DATE,
            trip_length INTEGER,
            price DECIMAL(12, 2) NOT NULL,
            currency VARCHAR(3) NOT NULL,
            distance_miles DECIMAL(10, 2),
            cost_per_mile DECIMAL(10, 4),
            queried_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create price_graph_results table: %w", err)
	}

	// Create price_graph_sweep_job_details table for scheduled sweep jobs
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS price_graph_sweep_job_details (
            id SERIAL PRIMARY KEY,
            job_id INTEGER REFERENCES scheduled_jobs(id) ON DELETE CASCADE,
            trip_lengths INTEGER[] DEFAULT '{7,14}',
            departure_window_days INTEGER DEFAULT 30,
            dynamic_dates BOOLEAN DEFAULT TRUE,
            trip_type VARCHAR(20) DEFAULT 'round_trip',
            class VARCHAR(20) DEFAULT 'economy',
            stops VARCHAR(20) DEFAULT 'any',
            adults INTEGER DEFAULT 1,
            currency VARCHAR(3) DEFAULT 'USD',
            rate_limit_millis INTEGER DEFAULT 3000,
            international_only BOOLEAN DEFAULT TRUE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create price_graph_sweep_job_details table: %w", err)
	}

	// Create continuous_sweep_progress table for tracking sweep state
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS continuous_sweep_progress (
            id SERIAL PRIMARY KEY,
            sweep_number INTEGER DEFAULT 1,
            route_index INTEGER DEFAULT 0,
            total_routes INTEGER DEFAULT 0,
            current_origin VARCHAR(10),
            current_destination VARCHAR(10),
            queries_completed INTEGER DEFAULT 0,
            errors_count INTEGER DEFAULT 0,
            last_error TEXT,
            sweep_started_at TIMESTAMP WITH TIME ZONE,
            last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            trip_lengths INTEGER[] DEFAULT '{7,14}',
            pacing_mode VARCHAR(20) DEFAULT 'adaptive',
            target_duration_hours INTEGER DEFAULT 24,
            min_delay_ms INTEGER DEFAULT 3000,
            is_running BOOLEAN DEFAULT FALSE,
            is_paused BOOLEAN DEFAULT FALSE,
            international_only BOOLEAN DEFAULT TRUE
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create continuous_sweep_progress table: %w", err)
	}

	_, err = p.db.Exec(`
        ALTER TABLE continuous_sweep_progress
        ADD COLUMN IF NOT EXISTS trip_lengths INTEGER[] DEFAULT '{7,14}'
    `)
	if err != nil {
		return fmt.Errorf("failed to add trip_lengths to continuous_sweep_progress: %w", err)
	}

	// Insert default progress row if not exists
	_, err = p.db.Exec(`
        INSERT INTO continuous_sweep_progress (id, sweep_number, route_index, total_routes)
        VALUES (1, 1, 0, 0)
        ON CONFLICT (id) DO NOTHING
    `)
	if err != nil {
		return fmt.Errorf("failed to insert default continuous_sweep_progress: %w", err)
	}

	// Create continuous_sweep_stats table for historical tracking
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS continuous_sweep_stats (
            id SERIAL PRIMARY KEY,
            sweep_number INTEGER NOT NULL,
            started_at TIMESTAMP WITH TIME ZONE NOT NULL,
            completed_at TIMESTAMP WITH TIME ZONE,
            total_routes INTEGER,
            successful_queries INTEGER DEFAULT 0,
            failed_queries INTEGER DEFAULT 0,
            total_duration_seconds INTEGER,
            avg_delay_ms INTEGER,
            min_price_found DECIMAL(10,2),
            max_price_found DECIMAL(10,2),
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create continuous_sweep_stats table: %w", err)
	}

	// Add job_type column to scheduled_jobs if not exists
	_, err = p.db.Exec(`
        ALTER TABLE scheduled_jobs ADD COLUMN IF NOT EXISTS job_type VARCHAR(50) DEFAULT 'bulk_search'
    `)
	if err != nil {
		return fmt.Errorf("failed to add job_type column: %w", err)
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
        CREATE INDEX IF NOT EXISTS idx_flight_segments_offer_id ON flight_segments(flight_offer_id);
        CREATE INDEX IF NOT EXISTS idx_bulk_searches_status ON bulk_searches(status);
        CREATE INDEX IF NOT EXISTS idx_bulk_search_results_search_id ON bulk_search_results(bulk_search_id);
        CREATE INDEX IF NOT EXISTS idx_bulk_search_offers_search_id ON bulk_search_offers(bulk_search_id);
        CREATE INDEX IF NOT EXISTS idx_price_graph_results_sweep_id ON price_graph_results(sweep_id);
        CREATE INDEX IF NOT EXISTS idx_price_graph_results_route_date ON price_graph_results(origin, destination, departure_date);
        CREATE INDEX IF NOT EXISTS idx_sweep_stats_sweep_number ON continuous_sweep_stats(sweep_number);
        CREATE INDEX IF NOT EXISTS idx_sweep_progress_updated ON continuous_sweep_progress(last_updated);
    `)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}
