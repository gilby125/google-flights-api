package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"time"
)

const airportsSeedMigrationPath = "migrations/000003_seed_airports.sql"

func EnsurePostgresAirportsSeeded(ctx context.Context, connString string) error {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return fmt.Errorf("open postgres connection: %w", err)
	}
	defer db.Close()

	seedCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	if err := db.PingContext(seedCtx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	var count int
	if err := db.QueryRowContext(seedCtx, "SELECT COUNT(*) FROM airports").Scan(&count); err != nil {
		return fmt.Errorf("count airports: %w", err)
	}
	if count > 0 {
		return nil
	}

	sqlBytes, err := fs.ReadFile(migrationsFS, airportsSeedMigrationPath)
	if err != nil {
		return fmt.Errorf("read embedded airports seed migration: %w", err)
	}

	if _, err := db.ExecContext(seedCtx, string(sqlBytes)); err != nil {
		return fmt.Errorf("seed airports via embedded migration: %w", err)
	}

	return nil
}
