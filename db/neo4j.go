package db

import (
	"fmt"

	"github.com/gilby125/google-flights-api/config"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jDB represents a Neo4j database connection
type Neo4jDB struct {
	driver neo4j.Driver
}

// NewNeo4jDB creates a new Neo4j database connection
func NewNeo4jDB(cfg config.Neo4jConfig) (*Neo4jDB, error) {
	driver, err := neo4j.NewDriver(cfg.URI, neo4j.BasicAuth(cfg.User, cfg.Password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	// Test the connection
	if err := driver.VerifyConnectivity(); err != nil {
		return nil, fmt.Errorf("failed to verify Neo4j connectivity: %w", err)
	}

	return &Neo4jDB{driver: driver}, nil
}

// Close closes the database connection
func (n *Neo4jDB) Close() error {
	return n.driver.Close()
}

// GetDriver returns the underlying database driver
func (n *Neo4jDB) GetDriver() neo4j.Driver {
	return n.driver
}

// InitSchema initializes the database schema
func (n *Neo4jDB) InitSchema() error {
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	// Create constraints
	_, err := session.Run(
		"CREATE CONSTRAINT airport_code IF NOT EXISTS FOR (a:Airport) REQUIRE a.code IS UNIQUE",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create airport code constraint: %w", err)
	}

	_, err = session.Run(
		"CREATE CONSTRAINT airline_code IF NOT EXISTS FOR (a:Airline) REQUIRE a.code IS UNIQUE",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create airline code constraint: %w", err)
	}

	return nil
}

// CreateAirport creates an airport node in Neo4j
func (n *Neo4jDB) CreateAirport(code, name, city, country string, latitude, longitude float64) error {
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	_, err := session.Run(
		"MERGE (a:Airport {code: $code}) "+
			"ON CREATE SET a.name = $name, a.city = $city, a.country = $country, a.latitude = $latitude, a.longitude = $longitude "+
			"ON MATCH SET a.name = $name, a.city = $city, a.country = $country, a.latitude = $latitude, a.longitude = $longitude",
		map[string]interface{}{
			"code":      code,
			"name":      name,
			"city":      city,
			"country":   country,
			"latitude":  latitude,
			"longitude": longitude,
		},
	)

	return err
}

// CreateAirline creates an airline node in Neo4j
func (n *Neo4jDB) CreateAirline(code, name, country string) error {
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	_, err := session.Run(
		"MERGE (a:Airline {code: $code}) "+
			"ON CREATE SET a.name = $name, a.country = $country "+
			"ON MATCH SET a.name = $name, a.country = $country",
		map[string]interface{}{
			"code":    code,
			"name":    name,
			"country": country,
		},
	)

	return err
}

// CreateRoute creates a route relationship between airports
func (n *Neo4jDB) CreateRoute(originCode, destCode, airlineCode, flightNumber string, avgPrice float64, avgDuration int) error {
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	_, err := session.Run(
		"MATCH (origin:Airport {code: $originCode}), (dest:Airport {code: $destCode}) "+
			"MERGE (origin)-[r:ROUTE {airline: $airlineCode, flightNumber: $flightNumber}]->(dest) "+
			"ON CREATE SET r.avgPrice = $avgPrice, r.avgDuration = $avgDuration, r.count = 1 "+
			"ON MATCH SET r.avgPrice = (r.avgPrice * r.count + $avgPrice) / (r.count + 1), "+
			"r.avgDuration = (r.avgDuration * r.count + $avgDuration) / (r.count + 1), "+
			"r.count = r.count + 1",
		map[string]interface{}{
			"originCode":   originCode,
			"destCode":     destCode,
			"airlineCode":  airlineCode,
			"flightNumber": flightNumber,
			"avgPrice":     avgPrice,
			"avgDuration":  avgDuration,
		},
	)

	return err
}

// AddPricePoint adds a price point for a route
func (n *Neo4jDB) AddPricePoint(originCode, destCode string, date string, price float64, airlineCode string) error {
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	_, err := session.Run(
		"MATCH (origin:Airport {code: $originCode}), (dest:Airport {code: $destCode}) "+
			"MERGE (origin)-[r:PRICE_POINT {date: $date, airline: $airlineCode}]->(dest) "+
			"SET r.price = $price",
		map[string]interface{}{
			"originCode":  originCode,
			"destCode":    destCode,
			"date":        date,
			"price":       price,
			"airlineCode": airlineCode,
		},
	)

	return err
}
