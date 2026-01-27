-- Ensure price_graph_results can be safely upserted for on-demand and sweep runs.
-- We treat (route + departure_date + trip_length + search context) as the natural key.
--
-- NOTE: We intentionally exclude return_date from the key so one-way searches (NULL return_date)
-- can still upsert cleanly.
CREATE UNIQUE INDEX IF NOT EXISTS idx_price_graph_results_upsert_key
ON price_graph_results (
  sweep_id,
  origin,
  destination,
  departure_date,
  trip_length,
  currency,
  adults,
  children,
  infants_lap,
  infants_seat,
  trip_type,
  class,
  stops
);

