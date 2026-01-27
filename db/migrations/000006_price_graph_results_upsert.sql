-- Ensure price_graph_results can be safely upserted for on-demand and sweep runs.
-- We treat (route + departure_date + trip_length + search context) as the natural key.
--
-- NOTE: We intentionally exclude return_date from the key so one-way searches (NULL return_date)
-- can still upsert cleanly.
--
-- This migration also deduplicates existing rows that would violate the new unique key.

-- Normalize legacy NULL trip_length values to 0 so the unique key is stable.
UPDATE price_graph_results
SET trip_length = 0
WHERE trip_length IS NULL;

-- Remove duplicates (keep the most recently queried).
WITH ranked AS (
  SELECT
    id,
    ROW_NUMBER() OVER (
      PARTITION BY
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
      ORDER BY
        queried_at DESC,
        id DESC
    ) AS rn
  FROM price_graph_results
)
DELETE FROM price_graph_results
WHERE id IN (SELECT id FROM ranked WHERE rn > 1);

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
