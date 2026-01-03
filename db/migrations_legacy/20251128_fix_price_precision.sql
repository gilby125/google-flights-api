-- Increase price column precision from DECIMAL(10,2) to DECIMAL(12,2)
-- This fixes "numeric field overflow" errors for high-value flights

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
