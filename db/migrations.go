package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const (
	migrationsGlobPattern = "migrations/*.sql"
	// Arbitrary, stable advisory lock key to avoid concurrent migrators.
	migrationsAdvisoryLockID int64 = 6905139011623814243
)

// RunMigrations applies embedded SQL migrations in filename order and records them in schema_migrations.
//
// The connString may be either a lib/pq keyword/value DSN or a postgres:// URL.
func RunMigrations(connString string) error {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return fmt.Errorf("open postgres connection for migrations: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping postgres for migrations: %w", err)
	}

	return runMigrations(ctx, db)
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, migrationsAdvisoryLockID); err != nil {
		return fmt.Errorf("acquire migrations advisory lock: %w", err)
	}
	defer func() {
		_, _ = db.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, migrationsAdvisoryLockID)
	}()

	if err := ensureSchemaMigrationsTable(ctx, db); err != nil {
		return err
	}

	migrationPaths, err := fs.Glob(migrationsFS, migrationsGlobPattern)
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(migrationPaths)

	for _, migrationPath := range migrationPaths {
		version := filepath.Base(migrationPath)
		sqlBytes, err := fs.ReadFile(migrationsFS, migrationPath)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", migrationPath, err)
		}

		checksum := sha256.Sum256(sqlBytes)
		checksumHex := hex.EncodeToString(checksum[:])

		appliedChecksum, alreadyApplied, err := getAppliedMigrationChecksum(ctx, db, version)
		if err != nil {
			return err
		}
		if alreadyApplied {
			if !strings.EqualFold(appliedChecksum, checksumHex) {
				return fmt.Errorf("migration %s checksum mismatch (db=%s file=%s)", version, appliedChecksum, checksumHex)
			}
			continue
		}

		if err := applyOneMigration(ctx, db, version, checksumHex, string(sqlBytes)); err != nil {
			return err
		}
	}

	return nil
}

func ensureSchemaMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			checksum TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}
	return nil
}

func getAppliedMigrationChecksum(ctx context.Context, db *sql.DB, version string) (string, bool, error) {
	var checksum string
	err := db.QueryRowContext(ctx, `SELECT checksum FROM schema_migrations WHERE version = $1`, version).Scan(&checksum)
	if err == nil {
		return checksum, true, nil
	}
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	return "", false, fmt.Errorf("check schema_migrations for %s: %w", version, err)
}

func applyOneMigration(ctx context.Context, db *sql.DB, version, checksum, migrationSQL string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration tx for %s: %w", version, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, migrationSQL); err != nil {
		return fmt.Errorf("execute migration %s: %w", version, err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO schema_migrations (version, checksum, applied_at) VALUES ($1, $2, NOW())`,
		version,
		checksum,
	); err != nil {
		return fmt.Errorf("record schema_migrations row for %s: %w", version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}
	return nil
}
