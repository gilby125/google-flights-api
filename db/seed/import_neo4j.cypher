// Import airports into Neo4j with all fields
// First, copy airports.csv to Neo4j import directory:
//   docker cp airports.csv neo4j:/var/lib/neo4j/import/
// Then run in Neo4j Browser or via cypher-shell

// Load airports from CSV with all fields
LOAD CSV WITH HEADERS FROM 'file:///airports.csv' AS row
MERGE (a:Airport {code: row.iata})
SET a.icao = row.icao,
    a.name = row.name,
    a.city = row.city,
    a.state = row.state,
    a.country = row.country,
    a.latitude = toFloat(row.latitude),
    a.longitude = toFloat(row.longitude),
    a.elevation_ft = toInteger(row.elevation_ft),
    a.timezone = row.timezone;

// Create indexes for faster lookups
CREATE INDEX airport_code_idx IF NOT EXISTS FOR (a:Airport) ON (a.code);
CREATE INDEX airport_country_idx IF NOT EXISTS FOR (a:Airport) ON (a.country);

// Verify import
MATCH (a:Airport) RETURN count(a) as total_airports;
MATCH (a:Airport) WHERE a.country = 'US' RETURN count(a) as us_airports;
MATCH (a:Airport) RETURN a LIMIT 5;
