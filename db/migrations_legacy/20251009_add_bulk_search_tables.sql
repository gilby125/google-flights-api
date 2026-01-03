-- Bulk search support tables
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
    min_price DECIMAL(10, 2),
    max_price DECIMAL(10, 2),
    average_price DECIMAL(10, 2)
);

CREATE TABLE IF NOT EXISTS bulk_search_results (
    id SERIAL PRIMARY KEY,
    bulk_search_id INTEGER REFERENCES bulk_searches(id) ON DELETE CASCADE,
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_date DATE NOT NULL,
    return_date DATE,
    price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    airline_code VARCHAR(3),
    duration INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bulk_searches_status ON bulk_searches(status);
CREATE INDEX IF NOT EXISTS idx_bulk_search_results_search_id ON bulk_search_results(bulk_search_id);
