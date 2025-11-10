-- Migration 0005: Add communal species flag
-- This migration adds a flag to identify communal tarantula species
-- that can live together in colonies

-- Add is_communal field to tarantula_species table
ALTER TABLE spider_bot.tarantula_species
    ADD COLUMN IF NOT EXISTS is_communal BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_tarantula_species_communal ON spider_bot.tarantula_species(is_communal);

COMMENT ON COLUMN spider_bot.tarantula_species.is_communal IS 'True if species can live communally (e.g., Monocentropus balfouri)';

-- Mark known communal species
-- Monocentropus balfouri is the most well-known communal tarantula species
UPDATE spider_bot.tarantula_species
SET is_communal = TRUE
WHERE LOWER(scientific_name) LIKE '%monocentropus%balfouri%'
   OR LOWER(scientific_name) LIKE '%balfouri%'
   OR LOWER(common_name) LIKE '%balfouri%';
