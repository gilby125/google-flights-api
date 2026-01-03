-- Baseline schema migration (squashed).
-- This file is intentionally written to be idempotent (IF NOT EXISTS / ADD COLUMN IF NOT EXISTS)
-- so existing dev/test DBs can be brought up to date without dropping state.

-- Core reference tables
CREATE TABLE IF NOT EXISTS airports (
    id SERIAL PRIMARY KEY,
    code VARCHAR(3) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS airlines (
    id SERIAL PRIMARY KEY,
    code VARCHAR(3) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    country VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS flights (
    id SERIAL PRIMARY KEY,
    flight_number VARCHAR(10) NOT NULL,
    airline_id INTEGER REFERENCES airlines(id),
    origin_id INTEGER REFERENCES airports(id),
    destination_id INTEGER REFERENCES airports(id),
    departure_time TIMESTAMPTZ NOT NULL,
    arrival_time TIMESTAMPTZ NOT NULL,
    duration INTEGER NOT NULL,
    distance INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS flight_prices (
    id SERIAL PRIMARY KEY,
    flight_id INTEGER REFERENCES flights(id),
    price DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    cabin_class VARCHAR(20) NOT NULL,
    search_date TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Scheduling / search orchestration
CREATE TABLE IF NOT EXISTS scheduled_jobs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    cron_expression VARCHAR(100) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    last_run TIMESTAMPTZ,
    job_type VARCHAR(50) DEFAULT 'bulk_search',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE scheduled_jobs ADD COLUMN IF NOT EXISTS job_type VARCHAR(50) DEFAULT 'bulk_search';

CREATE TABLE IF NOT EXISTS search_queries (
    id SERIAL PRIMARY KEY,
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_date DATE NOT NULL,
    return_date DATE,
    adults INTEGER NOT NULL,
    children INTEGER NOT NULL,
    infants_lap INTEGER NOT NULL,
    infants_seat INTEGER NOT NULL,
    trip_type VARCHAR(20) NOT NULL,
    class VARCHAR(20) NOT NULL,
    stops VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS job_details (
    id SERIAL PRIMARY KEY,
    job_id INTEGER REFERENCES scheduled_jobs(id),
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_date_start DATE NOT NULL,
    departure_date_end DATE NOT NULL,
    return_date_start DATE,
    return_date_end DATE,
    trip_length INTEGER,
    dynamic_dates BOOLEAN DEFAULT FALSE,
    days_from_execution INTEGER,
    search_window_days INTEGER,
    adults INTEGER DEFAULT 1,
    children INTEGER DEFAULT 0,
    infants_lap INTEGER DEFAULT 0,
    infants_seat INTEGER DEFAULT 0,
    trip_type VARCHAR(20) DEFAULT 'round_trip',
    class VARCHAR(20) DEFAULT 'economy',
    stops VARCHAR(20) DEFAULT 'any',
    currency VARCHAR(3) DEFAULT 'USD',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS search_results (
    id SERIAL PRIMARY KEY,
    search_id UUID NOT NULL,
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_date DATE NOT NULL,
    return_date DATE,
    adults INTEGER NOT NULL,
    children INTEGER NOT NULL,
    infants_lap INTEGER NOT NULL,
    infants_seat INTEGER NOT NULL,
    trip_type VARCHAR(20) NOT NULL,
    class VARCHAR(20) NOT NULL,
    stops VARCHAR(20) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    min_price DECIMAL(12, 2),
    max_price DECIMAL(12, 2),
    search_time TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    search_query_id INTEGER REFERENCES search_queries(id)
);

CREATE TABLE IF NOT EXISTS flight_offers (
    id SERIAL PRIMARY KEY,
    search_id UUID NOT NULL,
    price DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    airline_codes TEXT[] NOT NULL,
    outbound_duration INTEGER NOT NULL,
    outbound_stops INTEGER NOT NULL,
    return_duration INTEGER,
    return_stops INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS flight_segments (
    id SERIAL PRIMARY KEY,
    flight_offer_id INTEGER NOT NULL REFERENCES flight_offers(id) ON DELETE CASCADE,
    airline_code VARCHAR(3) NOT NULL,
    flight_number VARCHAR(10) NOT NULL,
    departure_airport VARCHAR(3) NOT NULL,
    arrival_airport VARCHAR(3) NOT NULL,
    departure_time TIMESTAMPTZ NOT NULL,
    arrival_time TIMESTAMPTZ NOT NULL,
    duration INTEGER NOT NULL,
    airplane VARCHAR(100),
    legroom VARCHAR(50),
    is_return BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Bulk search support
CREATE TABLE IF NOT EXISTS bulk_searches (
    id SERIAL PRIMARY KEY,
    job_id INTEGER REFERENCES scheduled_jobs(id),
    status VARCHAR(30) NOT NULL,
    total_searches INTEGER NOT NULL,
    completed INTEGER NOT NULL DEFAULT 0,
    total_offers INTEGER NOT NULL DEFAULT 0,
    error_count INTEGER NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    min_price DECIMAL(12, 2),
    max_price DECIMAL(12, 2),
    average_price DECIMAL(12, 2)
);

CREATE TABLE IF NOT EXISTS bulk_search_results (
    id SERIAL PRIMARY KEY,
    bulk_search_id INTEGER REFERENCES bulk_searches(id) ON DELETE CASCADE,
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_date DATE NOT NULL,
    return_date DATE,
    price DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    airline_code VARCHAR(3),
    duration INTEGER,
    src_airport_code VARCHAR(3),
    dst_airport_code VARCHAR(3),
    src_city TEXT,
    dst_city TEXT,
    flight_duration INTEGER,
    return_flight_duration INTEGER,
    outbound_flights JSONB,
    return_flights JSONB,
    offer_json JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE bulk_search_results ADD COLUMN IF NOT EXISTS src_airport_code VARCHAR(3);
ALTER TABLE bulk_search_results ADD COLUMN IF NOT EXISTS dst_airport_code VARCHAR(3);
ALTER TABLE bulk_search_results ADD COLUMN IF NOT EXISTS src_city TEXT;
ALTER TABLE bulk_search_results ADD COLUMN IF NOT EXISTS dst_city TEXT;
ALTER TABLE bulk_search_results ADD COLUMN IF NOT EXISTS flight_duration INTEGER;
ALTER TABLE bulk_search_results ADD COLUMN IF NOT EXISTS return_flight_duration INTEGER;
ALTER TABLE bulk_search_results ADD COLUMN IF NOT EXISTS outbound_flights JSONB;
ALTER TABLE bulk_search_results ADD COLUMN IF NOT EXISTS return_flights JSONB;
ALTER TABLE bulk_search_results ADD COLUMN IF NOT EXISTS offer_json JSONB;

CREATE TABLE IF NOT EXISTS bulk_search_offers (
    id SERIAL PRIMARY KEY,
    bulk_search_id INTEGER REFERENCES bulk_searches(id) ON DELETE CASCADE,
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_date DATE NOT NULL,
    return_date DATE,
    price DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    airline_codes TEXT[],
    src_airport_code VARCHAR(3),
    dst_airport_code VARCHAR(3),
    src_city TEXT,
    dst_city TEXT,
    flight_duration INTEGER,
    return_flight_duration INTEGER,
    distance_miles DECIMAL(10, 2),
    cost_per_mile DECIMAL(10, 4),
    outbound_flights JSONB,
    return_flights JSONB,
    offer_json JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE bulk_search_offers ADD COLUMN IF NOT EXISTS distance_miles DECIMAL(10, 2);
ALTER TABLE bulk_search_offers ADD COLUMN IF NOT EXISTS cost_per_mile DECIMAL(10, 4);

-- Certificates / secrets
CREATE TABLE IF NOT EXISTS certificate_issuance (
    id SERIAL PRIMARY KEY,
    domain VARCHAR(253) NOT NULL,
    serial_number VARCHAR(255) UNIQUE NOT NULL,
    issued_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    dns_challenge TEXT NOT NULL,
    cert_chain TEXT NOT NULL,
    private_key_enc TEXT NOT NULL,
    cloudflare_validation_id UUID,
    validation_status VARCHAR(50) NOT NULL,
    last_renewal_attempt TIMESTAMPTZ,
    renewal_errors INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS app_secrets (
    id SERIAL PRIMARY KEY,
    secret_name VARCHAR(255) UNIQUE NOT NULL,
    encrypted_value BYTEA NOT NULL,
    key_id VARCHAR(255) NOT NULL,
    rotation_schedule INTERVAL NOT NULL,
    last_rotated TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Price graph (scheduled + continuous sweeps)
CREATE TABLE IF NOT EXISTS price_graph_sweeps (
    id SERIAL PRIMARY KEY,
    job_id INTEGER REFERENCES scheduled_jobs(id),
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    origin_count INTEGER NOT NULL DEFAULT 0,
    destination_count INTEGER NOT NULL DEFAULT 0,
    trip_length_min INTEGER,
    trip_length_max INTEGER,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    error_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

INSERT INTO price_graph_sweeps (id, job_id, status, origin_count, destination_count, currency, error_count)
VALUES (0, NULL, 'continuous', 0, 0, 'USD', 0)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS price_graph_results (
    id SERIAL PRIMARY KEY,
    sweep_id INTEGER NOT NULL REFERENCES price_graph_sweeps(id) ON DELETE CASCADE,
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_date DATE NOT NULL,
    return_date DATE,
    trip_length INTEGER,
    price DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    distance_miles DECIMAL(10, 2),
    cost_per_mile DECIMAL(10, 4),
    queried_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE price_graph_results ADD COLUMN IF NOT EXISTS distance_miles DECIMAL(10, 2);
ALTER TABLE price_graph_results ADD COLUMN IF NOT EXISTS cost_per_mile DECIMAL(10, 4);

CREATE TABLE IF NOT EXISTS price_graph_sweep_job_details (
    id SERIAL PRIMARY KEY,
    job_id INTEGER REFERENCES scheduled_jobs(id) ON DELETE CASCADE,
    trip_lengths INTEGER[] DEFAULT '{7,14}',
    departure_window_days INTEGER DEFAULT 30,
    dynamic_dates BOOLEAN DEFAULT TRUE,
    trip_type VARCHAR(20) DEFAULT 'round_trip',
    class VARCHAR(20) DEFAULT 'economy',
    stops VARCHAR(20) DEFAULT 'any',
    adults INTEGER DEFAULT 1,
    currency VARCHAR(3) DEFAULT 'USD',
    rate_limit_millis INTEGER DEFAULT 3000,
    international_only BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS continuous_sweep_progress (
    id SERIAL PRIMARY KEY,
    sweep_number INTEGER DEFAULT 1,
    route_index INTEGER DEFAULT 0,
    total_routes INTEGER DEFAULT 0,
    current_origin VARCHAR(10),
    current_destination VARCHAR(10),
    queries_completed INTEGER DEFAULT 0,
    errors_count INTEGER DEFAULT 0,
    last_error TEXT,
    sweep_started_at TIMESTAMPTZ,
    last_updated TIMESTAMPTZ DEFAULT NOW(),
    pacing_mode VARCHAR(20) DEFAULT 'adaptive',
    target_duration_hours INTEGER DEFAULT 24,
    min_delay_ms INTEGER DEFAULT 3000,
    is_running BOOLEAN DEFAULT FALSE,
    is_paused BOOLEAN DEFAULT FALSE,
    international_only BOOLEAN DEFAULT TRUE
);

ALTER TABLE continuous_sweep_progress ADD COLUMN IF NOT EXISTS international_only BOOLEAN DEFAULT TRUE;

INSERT INTO continuous_sweep_progress (id, sweep_number, route_index, total_routes)
VALUES (1, 1, 0, 0)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS continuous_sweep_stats (
    id SERIAL PRIMARY KEY,
    sweep_number INTEGER NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    total_routes INTEGER,
    successful_queries INTEGER DEFAULT 0,
    failed_queries INTEGER DEFAULT 0,
    total_duration_seconds INTEGER,
    avg_delay_ms INTEGER,
    min_price_found DECIMAL(10, 2),
    max_price_found DECIMAL(10, 2),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Precision hardening (safe no-op if already correct)
ALTER TABLE flight_prices ALTER COLUMN price TYPE DECIMAL(12, 2);
ALTER TABLE search_results ALTER COLUMN min_price TYPE DECIMAL(12, 2);
ALTER TABLE search_results ALTER COLUMN max_price TYPE DECIMAL(12, 2);
ALTER TABLE flight_offers ALTER COLUMN price TYPE DECIMAL(12, 2);
ALTER TABLE bulk_searches ALTER COLUMN min_price TYPE DECIMAL(12, 2);
ALTER TABLE bulk_searches ALTER COLUMN max_price TYPE DECIMAL(12, 2);
ALTER TABLE bulk_searches ALTER COLUMN average_price TYPE DECIMAL(12, 2);
ALTER TABLE bulk_search_results ALTER COLUMN price TYPE DECIMAL(12, 2);
ALTER TABLE bulk_search_offers ALTER COLUMN price TYPE DECIMAL(12, 2);
ALTER TABLE price_graph_results ALTER COLUMN price TYPE DECIMAL(12, 2);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_airports_code ON airports(code);
CREATE INDEX IF NOT EXISTS idx_airlines_code ON airlines(code);
CREATE INDEX IF NOT EXISTS idx_flights_departure ON flights(departure_time);
CREATE INDEX IF NOT EXISTS idx_flights_arrival ON flights(arrival_time);
CREATE INDEX IF NOT EXISTS idx_flight_prices_search_date ON flight_prices(search_date);
CREATE INDEX IF NOT EXISTS idx_search_results_search_id ON search_results(search_id);
CREATE INDEX IF NOT EXISTS idx_flight_offers_search_id ON flight_offers(search_id);
CREATE INDEX IF NOT EXISTS idx_flight_segments_offer_id ON flight_segments(flight_offer_id);
CREATE INDEX IF NOT EXISTS idx_bulk_searches_status ON bulk_searches(status);
CREATE INDEX IF NOT EXISTS idx_bulk_search_results_search_id ON bulk_search_results(bulk_search_id);
CREATE INDEX IF NOT EXISTS idx_bulk_search_offers_search_id ON bulk_search_offers(bulk_search_id);
CREATE INDEX IF NOT EXISTS idx_price_graph_results_sweep_id ON price_graph_results(sweep_id);
CREATE INDEX IF NOT EXISTS idx_price_graph_results_route_date ON price_graph_results(origin, destination, departure_date);
CREATE INDEX IF NOT EXISTS idx_sweep_stats_sweep_number ON continuous_sweep_stats(sweep_number);
CREATE INDEX IF NOT EXISTS idx_sweep_progress_updated ON continuous_sweep_progress(last_updated);

-- Value-analysis indexes
CREATE INDEX IF NOT EXISTS idx_bulk_search_offers_cost_per_mile ON bulk_search_offers(cost_per_mile) WHERE cost_per_mile IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_price_graph_results_cost_per_mile ON price_graph_results(cost_per_mile) WHERE cost_per_mile IS NOT NULL;

