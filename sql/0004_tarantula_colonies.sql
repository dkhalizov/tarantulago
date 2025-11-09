-- Migration 0004: Add tarantula colony management
-- This migration adds support for managing communal tarantula colonies
-- Allows tracking multiple tarantulas living together (e.g., Monocentropus balfouri)

-- Create tarantula_colonies table for colony groups
CREATE TABLE IF NOT EXISTS spider_bot.tarantula_colonies (
    id SERIAL PRIMARY KEY,
    colony_name VARCHAR(100) NOT NULL,
    species_id INTEGER NOT NULL REFERENCES spider_bot.tarantula_species(id),
    formation_date DATE NOT NULL,
    enclosure_id INTEGER REFERENCES spider_bot.enclosures(id),
    notes TEXT,
    user_id BIGINT NOT NULL REFERENCES spider_bot.telegram_users(telegram_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tarantula_colonies_species ON spider_bot.tarantula_colonies(species_id);
CREATE INDEX IF NOT EXISTS idx_tarantula_colonies_user ON spider_bot.tarantula_colonies(user_id);
CREATE INDEX IF NOT EXISTS idx_tarantula_colonies_enclosure ON spider_bot.tarantula_colonies(enclosure_id);
CREATE INDEX IF NOT EXISTS idx_tarantula_colonies_formation_date ON spider_bot.tarantula_colonies(formation_date);

COMMENT ON TABLE spider_bot.tarantula_colonies IS 'Tracks communal tarantula colonies (e.g., M. balfouri)';
COMMENT ON COLUMN spider_bot.tarantula_colonies.species_id IS 'Must be a communal species';
COMMENT ON COLUMN spider_bot.tarantula_colonies.formation_date IS 'Date when colony was formed';

-- Create tarantula_colony_members table for tracking membership
CREATE TABLE IF NOT EXISTS spider_bot.tarantula_colony_members (
    id SERIAL PRIMARY KEY,
    colony_id INTEGER NOT NULL REFERENCES spider_bot.tarantula_colonies(id) ON DELETE CASCADE,
    tarantula_id INTEGER NOT NULL REFERENCES spider_bot.tarantulas(id) ON DELETE CASCADE,
    joined_date DATE NOT NULL,
    left_date DATE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    notes TEXT,
    user_id BIGINT NOT NULL REFERENCES spider_bot.telegram_users(telegram_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(colony_id, tarantula_id, is_active)
);

CREATE INDEX IF NOT EXISTS idx_colony_members_colony ON spider_bot.tarantula_colony_members(colony_id);
CREATE INDEX IF NOT EXISTS idx_colony_members_tarantula ON spider_bot.tarantula_colony_members(tarantula_id);
CREATE INDEX IF NOT EXISTS idx_colony_members_active ON spider_bot.tarantula_colony_members(is_active);
CREATE INDEX IF NOT EXISTS idx_colony_members_user ON spider_bot.tarantula_colony_members(user_id);

COMMENT ON TABLE spider_bot.tarantula_colony_members IS 'Tracks which tarantulas are in which colonies';
COMMENT ON COLUMN spider_bot.tarantula_colony_members.is_active IS 'False when tarantula has left the colony';
COMMENT ON COLUMN spider_bot.tarantula_colony_members.left_date IS 'Date when tarantula was removed from colony';

-- Create update triggers
CREATE TRIGGER update_tarantula_colonies_updated_at
    BEFORE UPDATE ON spider_bot.tarantula_colonies
    FOR EACH ROW
    EXECUTE FUNCTION spider_bot.update_updated_at_column();

CREATE TRIGGER update_tarantula_colony_members_updated_at
    BEFORE UPDATE ON spider_bot.tarantula_colony_members
    FOR EACH ROW
    EXECUTE FUNCTION spider_bot.update_updated_at_column();

-- Add colony_id to tarantulas table for quick lookups
ALTER TABLE spider_bot.tarantulas
    ADD COLUMN IF NOT EXISTS colony_id INTEGER REFERENCES spider_bot.tarantula_colonies(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_tarantulas_colony ON spider_bot.tarantulas(colony_id);

COMMENT ON COLUMN spider_bot.tarantulas.colony_id IS 'Colony this tarantula belongs to (if any)';

-- Add colony feeding support to feeding_events
ALTER TABLE spider_bot.feeding_events
    ADD COLUMN IF NOT EXISTS tarantula_colony_id INTEGER REFERENCES spider_bot.tarantula_colonies(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_feeding_events_tarantula_colony ON spider_bot.feeding_events(tarantula_colony_id);

COMMENT ON COLUMN spider_bot.feeding_events.tarantula_colony_id IS 'If set, this feeding was for an entire colony rather than individual tarantula';

-- Note: feeding_events will have EITHER tarantula_id OR tarantula_colony_id set, not both
