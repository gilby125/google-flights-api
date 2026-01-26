-- Update deal_config baseline_window_days to 0 (all history)
-- NOTE: This must be done via a new migration; do not edit previously applied migrations.

INSERT INTO deal_config (key, value, description)
VALUES ('baseline_window_days', '0', 'Days of history for baseline calculation (0 = all history)')
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    description = EXCLUDED.description,
    updated_at = NOW();
