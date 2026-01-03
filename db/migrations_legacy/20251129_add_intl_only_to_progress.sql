-- Migration: Add international_only flag to continuous sweep progress
-- Date: 2024-11-29

-- Add international_only column to track which route set was used
ALTER TABLE continuous_sweep_progress
ADD COLUMN IF NOT EXISTS international_only BOOLEAN DEFAULT TRUE;
