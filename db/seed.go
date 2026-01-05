package db

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gilby125/google-flights-api/iata"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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
	// Airports are now seeded via migration 000003_seed_airports.sql
	return nil
}

// seedAirlines seeds the airlines table with common airlines
func (p *PostgresDBImpl) seedAirlines() error {
	// Airlines are populated dynamically from flight data
	return nil
}

// SeedNeo4jData seeds the Neo4j database with airports from the iata package.
// This is IDEMPOTENT - uses MERGE to create or update nodes.
func (n *Neo4jDB) SeedNeo4jData(ctx context.Context, postgresDB PostgresDB) error {
	log.Println("Seeding Neo4j airports (Postgres preferred; iata fallback)...")

	// Get airport count first
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	result, err := session.Run("MATCH (a:Airport) RETURN count(a) as count", nil)
	if err != nil {
		session.Close()
		return fmt.Errorf("failed to count airports: %w", err)
	}
	var existingCount int64
	if result.Next() {
		if cnt, ok := result.Record().Get("count"); ok {
			if val, ok := cnt.(int64); ok {
				existingCount = val
			}
		}
	}
	session.Close()

	// Query airports from Postgres (which should have full data after migrations).
	// Some older schemas may not have the optional `icao` column; fall back gracefully.
	queryWithICAO := `
		SELECT code, COALESCE(icao, ''), COALESCE(name, ''), COALESCE(city, ''), 
		       COALESCE(state, ''), COALESCE(country, ''), 
		       COALESCE(latitude, 0), COALESCE(longitude, 0),
		       COALESCE(elevation_ft, 0), COALESCE(timezone, '')
		FROM airports 
		WHERE latitude != 0 AND longitude != 0
	`
	queryWithoutICAO := `
		SELECT code, COALESCE(name, ''), COALESCE(city, ''), 
		       COALESCE(state, ''), COALESCE(country, ''), 
		       COALESCE(latitude, 0), COALESCE(longitude, 0),
		       COALESCE(elevation_ft, 0), COALESCE(timezone, '')
		FROM airports 
		WHERE latitude != 0 AND longitude != 0
	`

	cols, err := postgresAirportColumns(ctx, postgresDB)
	if err != nil {
		log.Printf("Warning: failed to inspect Postgres airports schema (%v); seeding from iata package instead", err)
		return n.seedAirportsFromIATA(ctx)
	}
	hasICAO := containsStringFold(cols, "icao")

	query := queryWithoutICAO
	if hasICAO {
		query = queryWithICAO
	}

	rows, err := postgresDB.Search(ctx, query)
	if err != nil {
		log.Printf("Warning: failed to query airports from Postgres (%v); seeding from iata package instead", err)
		return n.seedAirportsFromIATA(ctx)
	}
	defer rows.Close()

	// Seed airports in batches using UNWIND for efficiency
	var airports []map[string]interface{}
	for rows.Next() {
		var code, icao, name, city, state, country, timezone string
		var lat, lon float64
		var elevation int
		if hasICAO {
			if err := rows.Scan(&code, &icao, &name, &city, &state, &country, &lat, &lon, &elevation, &timezone); err != nil {
				return fmt.Errorf("failed to scan airport row: %w", err)
			}
		} else {
			icao = ""
			if err := rows.Scan(&code, &name, &city, &state, &country, &lat, &lon, &elevation, &timezone); err != nil {
				return fmt.Errorf("failed to scan airport row: %w", err)
			}
		}
		airports = append(airports, map[string]interface{}{
			"code":      code,
			"icao":      icao,
			"name":      name,
			"city":      city,
			"state":     state,
			"country":   country,
			"latitude":  lat,
			"longitude": lon,
			"elevation": elevation,
			"timezone":  timezone,
		})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate airports rows: %w", err)
	}

	if len(airports) == 0 {
		log.Println("No airports found in Postgres, using iata package data...")
		// Fallback: seed from iata package directly
		return n.seedAirportsFromIATA(ctx)
	}

	// Batch insert using UNWIND (idempotent via MERGE)
	session = n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	batchSize := 500
	for i := 0; i < len(airports); i += batchSize {
		end := i + batchSize
		if end > len(airports) {
			end = len(airports)
		}
		batch := airports[i:end]

		_, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			_, err := tx.Run(`
				UNWIND $airports AS a
				MERGE (airport:Airport {code: a.code})
				SET airport.icao = a.icao,
				    airport.name = a.name,
				    airport.city = a.city,
				    airport.state = a.state,
				    airport.country = a.country,
				    airport.latitude = a.latitude,
				    airport.longitude = a.longitude,
				    airport.elevation_ft = a.elevation,
				    airport.timezone = a.timezone
			`, map[string]interface{}{"airports": batch})
			return nil, err
		})
		if err != nil {
			return fmt.Errorf("failed to seed airport batch %d-%d: %w", i, end, err)
		}
	}

	log.Printf("Neo4j seeding complete: %d airports (was %d)", len(airports), existingCount)
	return nil
}

// seedAirportsFromIATA seeds airports directly from the iata package (fallback)
func (n *Neo4jDB) seedAirportsFromIATA(ctx context.Context) error {
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	// Get all known IATA codes and seed them
	count := 0
	for _, code := range getKnownIATACodes() {
		loc := iata.IATATimeZone(code)
		if loc.Tz == "Not supported IATA Code" {
			continue
		}

		_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			_, err := tx.Run(`
				MERGE (a:Airport {code: $code})
				SET a.city = $city,
				    a.timezone = $timezone,
				    a.latitude = $lat,
				    a.longitude = $lon
			`, map[string]interface{}{
				"code":     code,
				"city":     loc.City,
				"timezone": loc.Tz,
				"lat":      loc.Lat,
				"lon":      loc.Lon,
			})
			return nil, err
		})
		if err != nil {
			log.Printf("Warning: failed to seed airport %s: %v", code, err)
			continue
		}
		count++
	}

	log.Printf("Neo4j seeding from iata package complete: %d airports", count)
	return nil
}

func postgresAirportColumns(ctx context.Context, postgresDB PostgresDB) ([]string, error) {
	rows, err := postgresDB.Search(ctx, "SELECT * FROM airports LIMIT 0")
	if err != nil {
		return nil, fmt.Errorf("query airports columns: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("read airports columns: %w", err)
	}
	return cols, nil
}

func containsStringFold(haystack []string, needle string) bool {
	for _, s := range haystack {
		if strings.EqualFold(s, needle) {
			return true
		}
	}
	return false
}

// getKnownIATACodes returns a list of common IATA codes to seed
// This is a subset - full list comes from Postgres after migration runs
func getKnownIATACodes() []string {
	return Top100AirportCodes()
}

// Top100AirportCodes returns IATA codes from the Top100Airports list
func Top100AirportCodes() []string {
	codes := make([]string, len(Top100Airports))
	for i, a := range Top100Airports {
		codes[i] = a.Code
	}
	return codes
}

// CreateFlightRoute creates a directed flight route relationship with properties
func (n *Neo4jDB) CreateFlightRoute(originCode, destCode, airlineCode string, distance int) error {
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		_, err := tx.Run(`
			MATCH (origin:Airport {code: $originCode}), (dest:Airport {code: $destCode})
			MERGE (origin)-[r:ROUTE {airline: $airlineCode}]->(dest)
			SET r.distance = $distance
		`, map[string]interface{}{
			"originCode":  originCode,
			"destCode":    destCode,
			"airlineCode": airlineCode,
			"distance":    distance,
		})
		return nil, err
	})

	if err != nil {
		return fmt.Errorf("failed to create flight route %s->%s: %w", originCode, destCode, err)
	}
	return nil
}
