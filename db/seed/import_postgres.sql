-- Import airports from CSV into Postgres
-- Run: psql -d your_database -f import_postgres.sql

-- First, add new columns if they don't exist
ALTER TABLE airports ADD COLUMN IF NOT EXISTS icao VARCHAR(4);
ALTER TABLE airports ADD COLUMN IF NOT EXISTS state VARCHAR(100);
ALTER TABLE airports ADD COLUMN IF NOT EXISTS elevation_ft INTEGER;
ALTER TABLE airports ADD COLUMN IF NOT EXISTS timezone VARCHAR(100);

-- Create temp table for import with all fields
CREATE TEMP TABLE airports_import (
    iata VARCHAR(3),
    icao VARCHAR(4),
    name TEXT,
    city TEXT,
    state VARCHAR(100),
    country VARCHAR(2),
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    elevation_ft INTEGER,
    timezone VARCHAR(100)
);

-- Import from CSV (adjust path as needed)
\copy airports_import FROM 'airports.csv' WITH CSV HEADER;

-- Upsert into airports table with all fields
INSERT INTO airports (code, icao, name, city, state, country, latitude, longitude, elevation_ft, timezone)
SELECT 
    iata AS code,
    icao,
    name,
    city,
    state,
    country,
    latitude,
    longitude,
    elevation_ft,
    timezone
FROM airports_import
ON CONFLICT (code) DO UPDATE SET
    icao = EXCLUDED.icao,
    name = EXCLUDED.name,
    city = EXCLUDED.city,
    state = EXCLUDED.state,
    country = EXCLUDED.country,
    latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    elevation_ft = EXCLUDED.elevation_ft,
    timezone = EXCLUDED.timezone,
    updated_at = NOW();

DROP TABLE airports_import;

SELECT COUNT(*) as imported_airports FROM airports;
SELECT * FROM airports LIMIT 5;
