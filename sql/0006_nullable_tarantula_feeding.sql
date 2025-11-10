-- Migration 0006: Make tarantula_id nullable in feeding_events
-- This allows colony feeding events to have NULL tarantula_id while still
-- referencing a tarantula_colony_id

-- Make tarantula_id nullable to support colony feeding
ALTER TABLE spider_bot.feeding_events
    ALTER COLUMN tarantula_id DROP NOT NULL;

COMMENT ON COLUMN spider_bot.feeding_events.tarantula_id IS 'Individual tarantula being fed (NULL for colony feeding)';
COMMENT ON COLUMN spider_bot.feeding_events.tarantula_colony_id IS 'Colony being fed (NULL for individual feeding)';

-- Add check constraint to ensure either tarantula_id or tarantula_colony_id is set (but not both)
ALTER TABLE spider_bot.feeding_events
    ADD CONSTRAINT feeding_events_target_check
    CHECK (
        (tarantula_id IS NOT NULL AND tarantula_colony_id IS NULL) OR
        (tarantula_id IS NULL AND tarantula_colony_id IS NOT NULL)
    );
