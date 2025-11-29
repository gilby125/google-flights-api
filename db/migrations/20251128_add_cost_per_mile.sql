-- Add distance and cost-per-mile columns for value analysis

-- Add to bulk_search_offers
ALTER TABLE bulk_search_offers ADD COLUMN IF NOT EXISTS distance_miles DECIMAL(10, 2);
ALTER TABLE bulk_search_offers ADD COLUMN IF NOT EXISTS cost_per_mile DECIMAL(10, 4);

-- Add to price_graph_results
ALTER TABLE price_graph_results ADD COLUMN IF NOT EXISTS distance_miles DECIMAL(10, 2);
ALTER TABLE price_graph_results ADD COLUMN IF NOT EXISTS cost_per_mile DECIMAL(10, 4);

-- Index for sorting by cost-per-mile (most common use case)
CREATE INDEX IF NOT EXISTS idx_bulk_search_offers_cost_per_mile ON bulk_search_offers(cost_per_mile) WHERE cost_per_mile IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_price_graph_results_cost_per_mile ON price_graph_results(cost_per_mile) WHERE cost_per_mile IS NOT NULL;
