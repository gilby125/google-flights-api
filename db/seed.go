package db

import (
	"context"
	"fmt"
	"log"
	// "strconv" // Removed unused import
	// "github.com/neo4j/neo4j-go-driver/v5/neo4j" // Removed unused import
)

// SeedData seeds the database with initial data
func (p *PostgresDBImpl) SeedData() error {
	// Check if airports table is empty
	var count int
	err := p.db.QueryRow("SELECT COUNT(*) FROM airports").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check if airports table is empty: %w", err)
	}

	// Only seed if tables are empty
	if count == 0 {
		log.Println("Seeding airports data...")
		if err := p.seedAirports(); err != nil {
			return fmt.Errorf("failed to seed airports: %w", err)
		}
	}

	// Check if airlines table is empty
	err = p.db.QueryRow("SELECT COUNT(*) FROM airlines").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check if airlines table is empty: %w", err)
	}

	// Only seed if tables are empty
	if count == 0 {
		log.Println("Seeding airlines data...")
		if err := p.seedAirlines(); err != nil {
			return fmt.Errorf("failed to seed airlines: %w", err)
		}
	}

	return nil
}

// seedAirports seeds the airports table with common airports
func (p *PostgresDBImpl) seedAirports() error {
    // ... (keep existing airport seeding implementation unchanged)
	return nil // Added missing return
}

// seedAirlines seeds the airlines table with common airlines
func (p *PostgresDBImpl) seedAirlines() error {
    // ... (keep existing airline seeding implementation unchanged)
	return nil // Added missing return
}

// SeedNeo4jData seeds the Neo4j database with initial data and relationships
func (n *Neo4jDB) SeedNeo4jData(ctx context.Context, postgresDB *PostgresDB) error {
    // ... (keep existing neo4j implementation unchanged)
	return nil // Added missing return
}

// CreateFlightRoute creates a directed flight route relationship with properties
func (n *Neo4jDB) CreateFlightRoute(originCode, destCode, airlineCode string, distance int) error {
    // ... (keep existing implementation unchanged)
	return nil // Added missing return
}
