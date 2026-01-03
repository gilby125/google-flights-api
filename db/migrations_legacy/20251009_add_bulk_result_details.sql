ALTER TABLE bulk_search_results
    ADD COLUMN IF NOT EXISTS src_airport_code VARCHAR(3),
    ADD COLUMN IF NOT EXISTS dst_airport_code VARCHAR(3),
    ADD COLUMN IF NOT EXISTS src_city TEXT,
    ADD COLUMN IF NOT EXISTS dst_city TEXT,
    ADD COLUMN IF NOT EXISTS flight_duration INTEGER,
    ADD COLUMN IF NOT EXISTS return_flight_duration INTEGER,
    ADD COLUMN IF NOT EXISTS outbound_flights JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS return_flights JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS offer_json JSONB DEFAULT '{}'::jsonb;

-- ensure defaults are removed for future inserts if not desired
ALTER TABLE bulk_search_results
    ALTER COLUMN outbound_flights DROP DEFAULT,
    ALTER COLUMN return_flights DROP DEFAULT,
    ALTER COLUMN offer_json DROP DEFAULT;
