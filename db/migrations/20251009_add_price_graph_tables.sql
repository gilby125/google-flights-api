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
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS price_graph_results (
    id SERIAL PRIMARY KEY,
    sweep_id INTEGER NOT NULL REFERENCES price_graph_sweeps(id) ON DELETE CASCADE,
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_date DATE NOT NULL,
    return_date DATE,
    trip_length INTEGER,
    price NUMERIC(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    queried_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_price_graph_results_sweep_id ON price_graph_results (sweep_id);
CREATE INDEX IF NOT EXISTS idx_price_graph_results_route_date ON price_graph_results (origin, destination, departure_date);
