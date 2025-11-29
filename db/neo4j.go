package db

import (
	"context" // Added context import
	"fmt"
	"strings" // Add strings import

	"github.com/gilby125/google-flights-api/config"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jDatabase defines the interface for Neo4j database operations.
type Neo4jDatabase interface {
	CreateAirport(code, name, city, country string, latitude, longitude float64) error
	CreateRoute(originCode, destCode, airlineCode, flightNumber string, avgPrice float64, avgDuration int) error
	AddPricePoint(originCode, destCode string, date string, price float64, airlineCode string) error
	CreateAirline(code, name, country string) error
	Close() error
	ExecuteReadQuery(ctx context.Context, query string, params map[string]interface{}) (Neo4jResult, error)
	// GetDriver() neo4j.Driver // Remove direct driver access
	InitSchema() error
}

// Neo4jSession defines the interface for a Neo4j session (read operations)
type Neo4jSession interface {
	Run(ctx context.Context, cypher string, params map[string]interface{}) (Neo4jResult, error)
	Close() error // Keep context removed from Close
}

// Neo4jResult defines the interface for Neo4j query results
type Neo4jResult interface {
	Next() bool // Removed context from Next again based on error at line 96
	Record() *neo4j.Record
	Err() error
	Close() error // Close the result and underlying session
}

// Ensure neo4j.Session implements Neo4jSession (adjust if methods differ slightly)
// Note: This might require wrapper types if signatures don't match exactly.
// For simplicity, we assume direct compatibility or will use wrappers later.
// var _ Neo4jSession = (neo4j.Session)(nil) // This won't work directly due to return type mismatch on Run

// Ensure neo4j.Result implements Neo4jResult
// var _ Neo4jResult = (*neo4j.Result)(nil) // Removing this problematic check

// Neo4jDB represents a Neo4j database connection.
// It implicitly implements the Neo4jDatabase interface.
type Neo4jDB struct {
	driver neo4j.Driver
}

// Neo4jResultWithSession wraps a Neo4j result and its session to ensure proper cleanup.
// The session is closed when Close() is called on this wrapper.
type Neo4jResultWithSession struct {
	result  neo4j.Result
	session neo4j.Session
}

// Next advances the result cursor
func (r *Neo4jResultWithSession) Next() bool {
	return r.result.Next()
}

// Record returns the current record
func (r *Neo4jResultWithSession) Record() *neo4j.Record {
	return r.result.Record()
}

// Err returns any error from result iteration
func (r *Neo4jResultWithSession) Err() error {
	return r.result.Err()
}

// Close closes the underlying session (and implicitly the result)
func (r *Neo4jResultWithSession) Close() error {
	if r.session != nil {
		return r.session.Close()
	}
	return nil
}

// Ensure Neo4jResultWithSession implements Neo4jResult
var _ Neo4jResult = (*Neo4jResultWithSession)(nil)

// NewNeo4jDB creates a new Neo4j database connection
func NewNeo4jDB(cfg config.Neo4jConfig) (*Neo4jDB, error) {
	trimmedURI := strings.TrimSpace(cfg.URI) // Explicitly trim whitespace
	driver, err := neo4j.NewDriver(trimmedURI, neo4j.BasicAuth(cfg.User, cfg.Password, ""))
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
	// Driver's Close method takes no arguments based on error message.
	return n.driver.Close()
}

// ExecuteReadQuery runs a read-only query against Neo4j.
// The returned Neo4jResult wraps both the result and session - caller MUST call Close() when done.
func (n *Neo4jDB) ExecuteReadQuery(ctx context.Context, query string, params map[string]interface{}) (Neo4jResult, error) {
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})

	result, err := session.Run(query, params)
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to run Neo4j query: %w", err)
	}

	// Return wrapper that will close session when Close() is called
	return &Neo4jResultWithSession{
		result:  result,
		session: session,
	}, nil
}

// GetDriver returns the underlying database driver - Deprecated for testability
// func (n *Neo4jDB) GetDriver() neo4j.Driver {
// 	return n.driver
// }

// InitSchema initializes the database schema
func (n *Neo4jDB) InitSchema() error {
	// ctx := context.Background() // Removed unused ctx
	// NewSession takes only SessionConfig
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close() // Close takes no arguments

	// Create constraints
	// session.Run takes query, params (nil here) (NO ctx)
	_, err := session.Run(
		"CREATE CONSTRAINT airport_code IF NOT EXISTS FOR (a:Airport) REQUIRE a.code IS UNIQUE",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create airport code constraint: %w", err)
	}

	// session.Run takes query, params (nil here) (NO ctx)
	_, err = session.Run(
		"CREATE CONSTRAINT airline_code IF NOT EXISTS FOR (a:Airline) REQUIRE a.code IS UNIQUE",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create airline code constraint: %w", err)
	}

	return nil
}

// CreateAirport creates or updates an airport node in Neo4j
func (n *Neo4jDB) CreateAirport(code, name, city, country string, latitude, longitude float64) error {
	// ctx := context.Background() // Removed unused ctx
	// NewSession takes only SessionConfig
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close() // Close takes no arguments

	// WriteTransaction takes work function (NO ctx based on errors)
	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		// tx.Run takes query, params (NO ctx based on errors)
		_, err := tx.Run(
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
		return nil, err
	})

	if err != nil {
		return fmt.Errorf("failed to create/update airport %s: %w", code, err)
	}
	return nil
}

// CreateAirline creates or updates an airline node in Neo4j
func (n *Neo4jDB) CreateAirline(code, name, country string) error {
	// ctx := context.Background() // Removed unused ctx
	// NewSession takes only SessionConfig
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close() // Close takes no arguments

	// WriteTransaction takes work function (NO ctx)
	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		// tx.Run takes query, params (NO ctx)
		_, err := tx.Run(
			"MERGE (a:Airline {code: $code}) "+
				"ON CREATE SET a.name = $name, a.country = $country "+
				"ON MATCH SET a.name = $name, a.country = $country",
			map[string]interface{}{
				"code":    code,
				"name":    name,
				"country": country,
			},
		)
		return nil, err
	})

	if err != nil {
		return fmt.Errorf("failed to create/update airline %s: %w", code, err)
	}
	return nil
}

// CreateRoute creates or updates a route relationship between airports
func (n *Neo4jDB) CreateRoute(originCode, destCode, airlineCode, flightNumber string, avgPrice float64, avgDuration int) error {
	// ctx := context.Background() // Removed unused ctx
	// NewSession takes only SessionConfig
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close() // Close takes no arguments

	// WriteTransaction takes work function (NO ctx)
	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		// tx.Run takes query, params (NO ctx)
		_, err := tx.Run(
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
		return nil, err
	})

	if err != nil {
		return fmt.Errorf("failed to create/update route %s->%s (%s %s): %w", originCode, destCode, airlineCode, flightNumber, err)
	}
	return nil
}

// AddPricePoint adds or updates a price point for a route on a specific date
func (n *Neo4jDB) AddPricePoint(originCode, destCode string, date string, price float64, airlineCode string) error {
	// ctx := context.Background() // Removed unused ctx
	// NewSession takes only SessionConfig
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close() // Close takes no arguments

	// WriteTransaction takes work function (NO ctx)
	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		// tx.Run takes query, params (NO ctx)
		_, err := tx.Run(
			"MATCH (origin:Airport {code: $originCode}), (dest:Airport {code: $destCode}) "+
				"MERGE (origin)-[r:PRICE_POINT {date: date($date), airline: $airlineCode}]->(dest) "+ // Use Neo4j date() function
				"SET r.price = $price",
			map[string]interface{}{
				"originCode":  originCode,
				"destCode":    destCode,
				"date":        date,
				"price":       price,
				"airlineCode": airlineCode,
			},
		)
		return nil, err
	})

	if err != nil {
		return fmt.Errorf("failed to add price point for %s->%s on %s (%s): %w", originCode, destCode, date, airlineCode, err)
	}
	return nil
}

// Ensure Neo4jDB implements Neo4jDatabase
var _ Neo4jDatabase = (*Neo4jDB)(nil)
