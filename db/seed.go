package db

import (
	"context"
	"fmt"
	"log"
	"strconv"
)

// SeedData seeds the database with initial data
func (p *PostgresDB) SeedData() error {
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
func (p *PostgresDB) seedAirports() error {
	// Common airports data
	airports := [][]string{
		{"ATL", "Hartsfield-Jackson Atlanta International Airport", "Atlanta", "USA", "33.6367", "-84.4281"},
		{"PEK", "Beijing Capital International Airport", "Beijing", "China", "40.0799", "116.6031"},
		{"LHR", "London Heathrow Airport", "London", "United Kingdom", "51.4700", "-0.4543"},
		{"ORD", "O'Hare International Airport", "Chicago", "USA", "41.9742", "-87.9073"},
		{"HND", "Tokyo Haneda Airport", "Tokyo", "Japan", "35.5494", "139.7798"},
		{"LAX", "Los Angeles International Airport", "Los Angeles", "USA", "33.9416", "-118.4085"},
		{"CDG", "Paris Charles de Gaulle Airport", "Paris", "France", "49.0097", "2.5479"},
		{"DFW", "Dallas/Fort Worth International Airport", "Dallas", "USA", "32.8998", "-97.0403"},
		{"FRA", "Frankfurt Airport", "Frankfurt", "Germany", "50.0379", "8.5622"},
		{"IST", "Istanbul Airport", "Istanbul", "Turkey", "41.2608", "28.7418"},
		{"AMS", "Amsterdam Airport Schiphol", "Amsterdam", "Netherlands", "52.3105", "4.7683"},
		{"CAN", "Guangzhou Baiyun International Airport", "Guangzhou", "China", "23.3959", "113.3080"},
		{"ICN", "Incheon International Airport", "Seoul", "South Korea", "37.4602", "126.4407"},
		{"DEL", "Indira Gandhi International Airport", "Delhi", "India", "28.5562", "77.1000"},
		{"SIN", "Singapore Changi Airport", "Singapore", "Singapore", "1.3644", "103.9915"},
		{"DXB", "Dubai International Airport", "Dubai", "United Arab Emirates", "25.2532", "55.3657"},
		{"JFK", "John F. Kennedy International Airport", "New York", "USA", "40.6413", "-73.7781"},
		{"MAD", "Adolfo Suárez Madrid–Barajas Airport", "Madrid", "Spain", "40.4983", "-3.5676"},
		{"LAS", "Harry Reid International Airport", "Las Vegas", "USA", "36.0840", "-115.1537"},
		{"SFO", "San Francisco International Airport", "San Francisco", "USA", "37.6213", "-122.3790"},
		{"BKK", "Suvarnabhumi Airport", "Bangkok", "Thailand", "13.6900", "100.7501"},
		{"YYZ", "Toronto Pearson International Airport", "Toronto", "Canada", "43.6777", "-79.6248"},
		{"MIA", "Miami International Airport", "Miami", "USA", "25.7932", "-80.2906"},
		{"SYD", "Sydney Airport", "Sydney", "Australia", "-33.9399", "151.1753"},
		{"MUC", "Munich Airport", "Munich", "Germany", "48.3538", "11.7861"},
		{"FCO", "Leonardo da Vinci–Fiumicino Airport", "Rome", "Italy", "41.8045", "12.2508"},
		{"BCN", "Josep Tarradellas Barcelona–El Prat Airport", "Barcelona", "Spain", "41.2974", "2.0833"},
		{"LGW", "London Gatwick Airport", "London", "United Kingdom", "51.1481", "-0.1903"},
		{"EWR", "Newark Liberty International Airport", "Newark", "USA", "40.6895", "-74.1745"},
		{"MEX", "Mexico City International Airport", "Mexico City", "Mexico", "19.4363", "-99.0721"},
	}

	// Begin transaction
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statement
	stmt, err := tx.Prepare("INSERT INTO airports (code, name, city, country, latitude, longitude) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert airports
	for _, airport := range airports {
		latitude, _ := strconv.ParseFloat(airport[4], 64)
		longitude, _ := strconv.ParseFloat(airport[5], 64)

		_, err := stmt.Exec(airport[0], airport[1], airport[2], airport[3], latitude, longitude)
		if err != nil {
			return fmt.Errorf("failed to insert airport %s: %w", airport[0], err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// seedAirlines seeds the airlines table with common airlines
func (p *PostgresDB) seedAirlines() error {
	// Common airlines data
	airlines := [][]string{
		{"DL", "Delta Air Lines", "USA"},
		{"AA", "American Airlines", "USA"},
		{"UA", "United Airlines", "USA"},
		{"LH", "Lufthansa", "Germany"},
		{"BA", "British Airways", "United Kingdom"},
		{"AF", "Air France", "France"},
		{"KL", "KLM Royal Dutch Airlines", "Netherlands"},
		{"EK", "Emirates", "United Arab Emirates"},
		{"QR", "Qatar Airways", "Qatar"},
		{"CX", "Cathay Pacific", "Hong Kong"},
		{"SQ", "Singapore Airlines", "Singapore"},
		{"TK", "Turkish Airlines", "Turkey"},
		{"EY", "Etihad Airways", "United Arab Emirates"},
		{"JL", "Japan Airlines", "Japan"},
		{"NH", "All Nippon Airways", "Japan"},
		{"CA", "Air China", "China"},
		{"CZ", "China Southern Airlines", "China"},
		{"MU", "China Eastern Airlines", "China"},
		{"KE", "Korean Air", "South Korea"},
		{"OZ", "Asiana Airlines", "South Korea"},
		{"TG", "Thai Airways", "Thailand"},
		{"SU", "Aeroflot", "Russia"},
		{"LX", "Swiss International Air Lines", "Switzerland"},
		{"OS", "Austrian Airlines", "Austria"},
		{"SK", "Scandinavian Airlines", "Sweden"},
		{"IB", "Iberia", "Spain"},
		{"AY", "Finnair", "Finland"},
		{"WN", "Southwest Airlines", "USA"},
		{"B6", "JetBlue", "USA"},
		{"AS", "Alaska Airlines", "USA"},
	}

	// Begin transaction
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statement
	stmt, err := tx.Prepare("INSERT INTO airlines (code, name, country) VALUES ($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert airlines
	for _, airline := range airlines {
		_, err := stmt.Exec(airline[0], airline[1], airline[2])
		if err != nil {
			return fmt.Errorf("failed to insert airline %s: %w", airline[0], err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SeedNeo4jData seeds the Neo4j database with initial data
func (n *Neo4jDB) SeedNeo4jData(ctx context.Context, postgresDB *PostgresDB) error {
	// Get airports from PostgreSQL
	rows, err := postgresDB.GetDB().Query("SELECT code, name, city, country, latitude, longitude FROM airports")
	if err != nil {
		return fmt.Errorf("failed to get airports from PostgreSQL: %w", err)
	}
	defer rows.Close()

	// Create airports in Neo4j
	for rows.Next() {
		var code, name, city, country string
		var latitude, longitude float64

		if err := rows.Scan(&code, &name, &city, &country, &latitude, &longitude); err != nil {
			return fmt.Errorf("failed to scan airport row: %w", err)
		}

		if err := n.CreateAirport(code, name, city, country, latitude, longitude); err != nil {
			return fmt.Errorf("failed to create airport in Neo4j: %w", err)
		}
	}

	// Get airlines from PostgreSQL
	rows, err = postgresDB.GetDB().Query("SELECT code, name, country FROM airlines")
	if err != nil {
		return fmt.Errorf("failed to get airlines from PostgreSQL: %w", err)
	}
	defer rows.Close()

	// Create airlines in Neo4j
	for rows.Next() {
		var code, name, country string

		if err := rows.Scan(&code, &name, &country); err != nil {
			return fmt.Errorf("failed to scan airline row: %w", err)
		}

		if err := n.CreateAirline(code, name, country); err != nil {
			return fmt.Errorf("failed to create airline in Neo4j: %w", err)
		}
	}

	return nil
}
