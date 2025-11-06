-- Migration 0004: Normalize database schema to meet 3NF requirements
-- This migration addresses:
-- 1. Removes redundant feeding_frequency column from feeding_schedules (3NF violation)
-- 2. The feeding frequency name can be derived via the frequency_id foreign key

-- Remove the redundant feeding_frequency TEXT column from feeding_schedules
-- This column creates a transitive dependency since frequency information
-- is already available through the frequency_id foreign key to feeding_frequencies table
ALTER TABLE spider_bot.feeding_schedules
    DROP COLUMN IF EXISTS feeding_frequency;

-- Add comment to document the normalization
COMMENT ON TABLE spider_bot.feeding_schedules IS 'Feeding schedules by species and size. Uses frequency_id to reference feeding_frequencies table for frequency information (3NF compliant).';
COMMENT ON COLUMN spider_bot.feeding_schedules.frequency_id IS 'References feeding_frequencies table. Use JOIN to get frequency_name instead of storing redundantly.';
