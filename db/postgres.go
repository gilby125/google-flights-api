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
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

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
	fmt.Println("InitSchema called")
	// Create tables in the correct order to respect foreign key constraints

	// First create airports table
	_, err := p.db.Exec(`
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

	if err != nil {
		return fmt.Errorf("failed to create scheduled_jobs table: %w", err)
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

	// Drop flight_offers table if it exists - must be dropped before search_queries due to foreign key constraint
	_, err = p.db.Exec(`DROP TABLE IF EXISTS flight_offers;`)
	if err != nil {
		return fmt.Errorf("failed to drop flight_offers table: %w", err)
	}

	// Drop search_results table if it exists - must be dropped before search_queries due to foreign key constraint
	_, err = p.db.Exec(`DROP TABLE IF EXISTS search_results;`)
	if err != nil {
		return fmt.Errorf("failed to drop search_results table: %w", err)
	}

	// Drop search_queries table if it exists
	_, err = p.db.Exec(`DROP TABLE IF EXISTS search_queries;`)
	if err != nil {
		return fmt.Errorf("failed to drop search_queries table: %w", err)
	}

	// Create search_queries table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS search_queries (
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
            status VARCHAR(20) NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create search_queries table: %w", err)
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

	// Create search_queries table
	_, err = p.db.Exec(`
        CREATE TABLE IF NOT EXISTS search_queries (
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
            status VARCHAR(20) NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create search_queries table: %w", err)
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
            search_query_id INTEGER REFERENCES search_queries(id),
            search_id UUID NOT NULL,
            departure_date DATE NOT NULL,
            return_date DATE,
            price DECIMAL(10, 2) NOT NULL,
            currency VARCHAR(3) NOT NULL,
			airline_codes TEXT[] NOT NULL,
			outbound_duration INTEGER NOT NULL,
			outbound_stops INTEGER,
			return_duration INTEGER,
			return_stops INTEGER,
			total_duration INTEGER,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create flight_offers table: %w", err)
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
