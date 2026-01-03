CREATE TABLE IF NOT EXISTS bulk_search_offers (
    id SERIAL PRIMARY KEY,
    bulk_search_id INTEGER REFERENCES bulk_searches(id) ON DELETE CASCADE,
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_date DATE NOT NULL,
    return_date DATE,
    price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    airline_codes TEXT[],
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

CREATE INDEX IF NOT EXISTS idx_bulk_search_offers_search_id ON bulk_search_offers(bulk_search_id);
