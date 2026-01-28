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
	InitSchema() error
	// Graph traversal methods
	FindCheapestPath(ctx context.Context, origin, dest string, maxHops int, maxPrice float64) ([]PathResult, error)
	FindConnections(ctx context.Context, origin string, maxHops int, maxPrice float64) ([]Connection, error)
	GetRouteStats(ctx context.Context, origin, dest string) (*RouteStats, error)
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

// PathResult represents a multi-hop route path with total cost
type PathResult struct {
	Stops      []string  `json:"stops"`       // Airport codes in order
	TotalPrice float64   `json:"total_price"` // Sum of leg prices
	Legs       []LegInfo `json:"legs"`        // Details per leg
}

// LegInfo contains pricing info for a single leg
type LegInfo struct {
	Origin      string  `json:"origin"`
	Destination string  `json:"destination"`
	Price       float64 `json:"price"`
	Airline     string  `json:"airline,omitempty"`
}

// Connection represents a reachable destination from an origin
type Connection struct {
	Airport      string  `json:"airport"`       // Destination airport code
	Name         string  `json:"name"`          // Airport name
	Country      string  `json:"country"`       // Country code
	CheapestPath float64 `json:"cheapest_path"` // Cheapest total price to reach
	Hops         int     `json:"hops"`          // Number of stops
}

// RouteStats contains aggregated statistics for a route
type RouteStats struct {
	Origin      string   `json:"origin"`
	Destination string   `json:"destination"`
	MinPrice    float64  `json:"min_price"`
	MaxPrice    float64  `json:"max_price"`
	AvgPrice    float64  `json:"avg_price"`
	PricePoints int      `json:"price_points"` // Number of observations
	Airlines    []string `json:"airlines"`     // Airlines serving the route
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
				"ON MATCH SET "+
				"a.name = CASE WHEN $name <> '' THEN $name ELSE a.name END, "+
				"a.city = CASE WHEN $city <> '' THEN $city ELSE a.city END, "+
				"a.country = CASE WHEN $country <> '' THEN $country ELSE a.country END, "+
				"a.latitude = CASE WHEN $latitude = 0 AND $longitude = 0 THEN a.latitude ELSE $latitude END, "+
				"a.longitude = CASE WHEN $latitude = 0 AND $longitude = 0 THEN a.longitude ELSE $longitude END",
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

// FindCheapestPath finds the cheapest multi-hop paths between two airports
func (n *Neo4jDB) FindCheapestPath(ctx context.Context, origin, dest string, maxHops int, maxPrice float64) ([]PathResult, error) {
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()

	// Build dynamic hop range based on maxHops
	query := fmt.Sprintf(`
		MATCH path = (a:Airport {code: $origin})-[:PRICE_POINT*1..%d]->(b:Airport {code: $dest})
		WITH path, 
		     [n IN nodes(path) | n.code] AS stops,
		     reduce(total = 0.0, r IN relationships(path) | total + r.price) AS totalPrice,
		     [r IN relationships(path) | {origin: startNode(r).code, dest: endNode(r).code, price: r.price, airline: r.airline}] AS legs
		WHERE totalPrice <= $maxPrice
		RETURN stops, totalPrice, legs
		ORDER BY totalPrice
		LIMIT 10
	`, maxHops)

	result, err := session.Run(query, map[string]interface{}{
		"origin":   origin,
		"dest":     dest,
		"maxPrice": maxPrice,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find cheapest path: %w", err)
	}

	var paths []PathResult
	for result.Next() {
		record := result.Record()

		stopsVal, _ := record.Get("stops")
		stops := []string{}
		if stopsArr, ok := stopsVal.([]interface{}); ok {
			for _, s := range stopsArr {
				if str, ok := s.(string); ok {
					stops = append(stops, str)
				}
			}
		}

		totalPrice, _ := record.Get("totalPrice")
		price := 0.0
		if p, ok := totalPrice.(float64); ok {
			price = p
		}

		legsVal, _ := record.Get("legs")
		legs := []LegInfo{}
		if legsArr, ok := legsVal.([]interface{}); ok {
			for _, l := range legsArr {
				if legMap, ok := l.(map[string]interface{}); ok {
					leg := LegInfo{}
					if o, ok := legMap["origin"].(string); ok {
						leg.Origin = o
					}
					if d, ok := legMap["dest"].(string); ok {
						leg.Destination = d
					}
					if p, ok := legMap["price"].(float64); ok {
						leg.Price = p
					}
					if a, ok := legMap["airline"].(string); ok {
						leg.Airline = a
					}
					legs = append(legs, leg)
				}
			}
		}

		paths = append(paths, PathResult{
			Stops:      stops,
			TotalPrice: price,
			Legs:       legs,
		})
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("error iterating path results: %w", err)
	}

	return paths, nil
}

// FindConnections finds all reachable destinations from an origin within budget
func (n *Neo4jDB) FindConnections(ctx context.Context, origin string, maxHops int, maxPrice float64) ([]Connection, error) {
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()

	query := fmt.Sprintf(`
		MATCH path = (a:Airport {code: $origin})-[:PRICE_POINT*1..%d]->(dest:Airport)
		WHERE a <> dest
		WITH dest, path,
		     reduce(total = 0.0, r IN relationships(path) | total + r.price) AS totalPrice,
		     length(path) AS hops
		WHERE totalPrice <= $maxPrice
		RETURN DISTINCT dest.code AS airport, dest.name AS name, dest.country AS country,
		       min(totalPrice) AS cheapestPath, min(hops) AS hops
		ORDER BY cheapestPath
		LIMIT 100
	`, maxHops)

	result, err := session.Run(query, map[string]interface{}{
		"origin":   origin,
		"maxPrice": maxPrice,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find connections: %w", err)
	}

	var connections []Connection
	for result.Next() {
		record := result.Record()

		airport, _ := record.Get("airport")
		name, _ := record.Get("name")
		country, _ := record.Get("country")
		cheapest, _ := record.Get("cheapestPath")
		hops, _ := record.Get("hops")

		conn := Connection{}
		if a, ok := airport.(string); ok {
			conn.Airport = a
		}
		if n, ok := name.(string); ok {
			conn.Name = n
		}
		if c, ok := country.(string); ok {
			conn.Country = c
		}
		if p, ok := cheapest.(float64); ok {
			conn.CheapestPath = p
		}
		if h, ok := hops.(int64); ok {
			conn.Hops = int(h)
		}

		connections = append(connections, conn)
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("error iterating connection results: %w", err)
	}

	return connections, nil
}

// GetRouteStats returns aggregated price statistics for a specific route
func (n *Neo4jDB) GetRouteStats(ctx context.Context, origin, dest string) (*RouteStats, error) {
	session := n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()

	query := `
		MATCH (a:Airport {code: $origin})-[r:PRICE_POINT]->(b:Airport {code: $dest})
		RETURN a.code AS origin, b.code AS destination,
		       min(toFloat(r.price)) AS minPrice,
		       max(toFloat(r.price)) AS maxPrice,
		       avg(toFloat(r.price)) AS avgPrice,
		       count(r) AS pricePointCount,
		       collect(DISTINCT r.airline) AS airlines
	`

	result, err := session.Run(query, map[string]interface{}{
		"origin": origin,
		"dest":   dest,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get route stats: %w", err)
	}

	if !result.Next() {
		return nil, nil // No route found
	}

	record := result.Record()
	stats := &RouteStats{}

	if o, ok := record.Get("origin"); ok {
		if str, ok := o.(string); ok {
			stats.Origin = str
		}
	}
	if d, ok := record.Get("destination"); ok {
		if str, ok := d.(string); ok {
			stats.Destination = str
		}
	}
	if p, ok := record.Get("minPrice"); ok {
		if f, ok := p.(float64); ok {
			stats.MinPrice = f
		}
	}
	if p, ok := record.Get("maxPrice"); ok {
		if f, ok := p.(float64); ok {
			stats.MaxPrice = f
		}
	}
	if p, ok := record.Get("avgPrice"); ok {
		if f, ok := p.(float64); ok {
			stats.AvgPrice = f
		}
	}
	if c, ok := record.Get("pricePointCount"); ok {
		if i, ok := c.(int64); ok {
			stats.PricePoints = int(i)
		}
	}
	if a, ok := record.Get("airlines"); ok {
		if arr, ok := a.([]interface{}); ok {
			seen := make(map[string]bool)
			for _, v := range arr {
				if str, ok := v.(string); ok && str != "" && !seen[str] {
					stats.Airlines = append(stats.Airlines, str)
					seen[str] = true
				}
			}
		}
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("error reading route stats: %w", err)
	}

	return stats, nil
}

// Ensure Neo4jDB implements Neo4jDatabase
var _ Neo4jDatabase = (*Neo4jDB)(nil)
