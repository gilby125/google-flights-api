-- Add search context fields to price_graph_results so results are reproducible.

-- Traveler counts
ALTER TABLE price_graph_results ADD COLUMN IF NOT EXISTS adults INTEGER NOT NULL DEFAULT 1;
ALTER TABLE price_graph_results ADD COLUMN IF NOT EXISTS children INTEGER NOT NULL DEFAULT 0;
ALTER TABLE price_graph_results ADD COLUMN IF NOT EXISTS infants_lap INTEGER NOT NULL DEFAULT 0;
ALTER TABLE price_graph_results ADD COLUMN IF NOT EXISTS infants_seat INTEGER NOT NULL DEFAULT 0;

-- Search options
ALTER TABLE price_graph_results ADD COLUMN IF NOT EXISTS trip_type VARCHAR(20) NOT NULL DEFAULT 'round_trip';
ALTER TABLE price_graph_results ADD COLUMN IF NOT EXISTS class VARCHAR(20) NOT NULL DEFAULT 'economy';
ALTER TABLE price_graph_results ADD COLUMN IF NOT EXISTS stops VARCHAR(20) NOT NULL DEFAULT 'any';

-- User-facing link for drill-down exploration
ALTER TABLE price_graph_results ADD COLUMN IF NOT EXISTS search_url TEXT;

CREATE INDEX IF NOT EXISTS idx_price_graph_results_trip_type ON price_graph_results(trip_type);
CREATE INDEX IF NOT EXISTS idx_price_graph_results_class ON price_graph_results(class);
CREATE INDEX IF NOT EXISTS idx_price_graph_results_stops ON price_graph_results(stops);

