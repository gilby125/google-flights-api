package db

import (
	"database/sql"
	"fmt"

	"github.com/gilby125/google-flights-api/config"
	_ "github.com/lib/pq"
)

// PostgresDB represents a PostgreSQL database connection
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg config.PostgresConfig) (*PostgresDB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &PostgresDB{db: db}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// GetDB returns the underlying database connection
func (p *PostgresDB) GetDB() *sql.DB {
	return p.db
}

// InitSchema initializes the database schema
func (p *PostgresDB) InitSchema() error {
	// Create tables if they don't exist
	_, err := p.db.Exec(`
		-- Airports table
		CREATE TABLE IF NOT EXISTS airports (
			code VARCHAR(3) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			city VARCHAR(255) NOT NULL,
			country VARCHAR(255) NOT NULL,
			latitude DECIMAL(10, 6),
			longitude DECIMAL(10, 6)
		);

		-- Airlines table
		CREATE TABLE IF NOT EXISTS airlines (
			code VARCHAR(3) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			country VARCHAR(255)
		);

		-- Search queries table
		CREATE TABLE IF NOT EXISTS search_queries (
			id SERIAL PRIMARY KEY,
			origin VARCHAR(3) NOT NULL,
			destination VARCHAR(3) NOT NULL,
			departure_date DATE NOT NULL,
			return_date DATE,
			adults INT DEFAULT 1,
			children INT DEFAULT 0,
			infants_lap INT DEFAULT 0,
			infants_seat INT DEFAULT 0,
			trip_type VARCHAR(20) NOT NULL,
			class VARCHAR(20) NOT NULL,
			stops VARCHAR(20) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			status VARCHAR(20) DEFAULT 'pending',
			FOREIGN KEY (origin) REFERENCES airports(code),
			FOREIGN KEY (destination) REFERENCES airports(code)
		);

		-- Flight offers table
		CREATE TABLE IF NOT EXISTS flight_offers (
			id SERIAL PRIMARY KEY,
			search_query_id INT NOT NULL,
			price DECIMAL(10, 2) NOT NULL,
			currency VARCHAR(3) NOT NULL,
			departure_date DATE NOT NULL,
			return_date DATE,
			total_duration INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (search_query_id) REFERENCES search_queries(id)
		);

		-- Flight segments table
		CREATE TABLE IF NOT EXISTS flight_segments (
			id SERIAL PRIMARY KEY,
			flight_offer_id INT NOT NULL,
			airline_code VARCHAR(3) NOT NULL,
			flight_number VARCHAR(10) NOT NULL,
			departure_airport VARCHAR(3) NOT NULL,
			arrival_airport VARCHAR(3) NOT NULL,
			departure_time TIMESTAMP NOT NULL,
			arrival_time TIMESTAMP NOT NULL,
			duration INT NOT NULL,
			airplane VARCHAR(100),
			legroom VARCHAR(50),
			is_return BOOLEAN DEFAULT FALSE,
			FOREIGN KEY (flight_offer_id) REFERENCES flight_offers(id),
			FOREIGN KEY (airline_code) REFERENCES airlines(code),
			FOREIGN KEY (departure_airport) REFERENCES airports(code),
			FOREIGN KEY (arrival_airport) REFERENCES airports(code)
		);

		-- Scheduled jobs table
		CREATE TABLE IF NOT EXISTS scheduled_jobs (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			cron_expression VARCHAR(100) NOT NULL,
			enabled BOOLEAN DEFAULT TRUE,
			last_run TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Job details table
		CREATE TABLE IF NOT EXISTS job_details (
			id SERIAL PRIMARY KEY,
			job_id INT NOT NULL,
			origin VARCHAR(3) NOT NULL,
			destination VARCHAR(3) NOT NULL,
			departure_date_start DATE NOT NULL,
			departure_date_end DATE NOT NULL,
			return_date_start DATE,
			return_date_end DATE,
			trip_length INT,
			adults INT DEFAULT 1,
			children INT DEFAULT 0,
			infants_lap INT DEFAULT 0,
			infants_seat INT DEFAULT 0,
			trip_type VARCHAR(20) NOT NULL,
			class VARCHAR(20) NOT NULL,
			stops VARCHAR(20) NOT NULL,
			FOREIGN KEY (job_id) REFERENCES scheduled_jobs(id),
			FOREIGN KEY (origin) REFERENCES airports(code),
			FOREIGN KEY (destination) REFERENCES airports(code)
		);
	`)

	return err
}
