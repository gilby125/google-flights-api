-- Add trip length configuration to continuous sweep progress tracking.
ALTER TABLE continuous_sweep_progress
ADD COLUMN IF NOT EXISTS trip_lengths INTEGER[] DEFAULT '{7,14}';

UPDATE continuous_sweep_progress
SET trip_lengths = '{7,14}'
WHERE trip_lengths IS NULL;

