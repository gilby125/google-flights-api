-- Migration: Add scheduled price graph sweep support
-- Date: 2024-11-29

-- Add job_type to scheduled_jobs table
ALTER TABLE scheduled_jobs ADD COLUMN IF NOT EXISTS job_type VARCHAR(50) DEFAULT 'bulk_search';

-- Create price graph sweep job details table
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

-- Create continuous sweep progress tracking table
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
    is_paused BOOLEAN DEFAULT FALSE
);

-- Create sweep statistics table for historical tracking
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
    min_price_found DECIMAL(10,2),
    max_price_found DECIMAL(10,2),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_sweep_stats_sweep_number ON continuous_sweep_stats(sweep_number);
CREATE INDEX IF NOT EXISTS idx_sweep_progress_updated ON continuous_sweep_progress(last_updated);

-- Insert default progress row if not exists
INSERT INTO continuous_sweep_progress (id, sweep_number, route_index, total_routes)
VALUES (1, 1, 0, 0)
ON CONFLICT (id) DO NOTHING;
