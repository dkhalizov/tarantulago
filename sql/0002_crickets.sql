CREATE TABLE spider_bot.colony_maintenance_types
(
    id              SERIAL PRIMARY KEY,
    type_name       VARCHAR(50) NOT NULL UNIQUE,
    description     TEXT,
    frequency_days  INTEGER NOT NULL DEFAULT 7
);

CREATE TABLE spider_bot.colony_maintenance_records
(
    id                  SERIAL PRIMARY KEY,
    colony_id           INTEGER NOT NULL REFERENCES spider_bot.cricket_colonies(id),
    maintenance_type_id INTEGER NOT NULL REFERENCES spider_bot.colony_maintenance_types(id),
    maintenance_date    DATE NOT NULL,
    notes               TEXT,
    user_id             BIGINT NOT NULL REFERENCES spider_bot.telegram_users(telegram_id),
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_maintenance_records_colony ON spider_bot.colony_maintenance_records(colony_id);
CREATE INDEX idx_maintenance_records_type ON spider_bot.colony_maintenance_records(maintenance_type_id);
CREATE INDEX idx_maintenance_records_date ON spider_bot.colony_maintenance_records(maintenance_date);
CREATE INDEX idx_maintenance_records_user ON spider_bot.colony_maintenance_records(user_id);

CREATE TABLE spider_bot.colony_maintenance_schedules
(
    id                  SERIAL PRIMARY KEY,
    colony_id           INTEGER NOT NULL REFERENCES spider_bot.cricket_colonies(id),
    maintenance_type_id INTEGER NOT NULL REFERENCES spider_bot.colony_maintenance_types(id),
    frequency_days      INTEGER NOT NULL DEFAULT 7,
    enabled             BOOLEAN NOT NULL DEFAULT TRUE,
    last_performed_date DATE,
    user_id             BIGINT NOT NULL REFERENCES spider_bot.telegram_users(telegram_id),
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_maintenance_schedules_colony ON spider_bot.colony_maintenance_schedules(colony_id);
CREATE INDEX idx_maintenance_schedules_type ON spider_bot.colony_maintenance_schedules(maintenance_type_id);
CREATE INDEX idx_maintenance_schedules_user ON spider_bot.colony_maintenance_schedules(user_id);

CREATE TRIGGER update_colony_maintenance_schedules_updated_at
    BEFORE UPDATE
    ON spider_bot.colony_maintenance_schedules
    FOR EACH ROW
    EXECUTE FUNCTION spider_bot.update_updated_at_column();

INSERT INTO spider_bot.colony_maintenance_types (id, type_name, description, frequency_days)
VALUES
    (1, 'Count', 'Count the current number of crickets', 14),
    (2, 'FoodWater', 'Check and replace food and water', 3),
    (3, 'Cleaning', 'Clean the cricket enclosure', 14),
    (4, 'AdultRemoval', 'Remove older crickets to maintain colony quality', 7);

SELECT setval('spider_bot.colony_maintenance_types_id_seq', (SELECT MAX(id) FROM spider_bot.colony_maintenance_types));

ALTER TABLE spider_bot.user_settings
    ADD COLUMN IF NOT EXISTS maintenance_reminder_enabled BOOLEAN DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS food_water_frequency_days INTEGER DEFAULT 3,
ADD COLUMN IF NOT EXISTS cleaning_frequency_days INTEGER DEFAULT 14,
ADD COLUMN IF NOT EXISTS adult_removal_frequency_days INTEGER DEFAULT 7,
ADD COLUMN IF NOT EXISTS notifications_paused BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS pause_start_date TIMESTAMP,
ADD COLUMN IF NOT EXISTS pause_end_date TIMESTAMP,
ADD COLUMN IF NOT EXISTS pause_reason VARCHAR(255);

CREATE TABLE spider_bot.colony_maintenance_types
(
    id              SERIAL PRIMARY KEY,
    type_name       VARCHAR(50) NOT NULL UNIQUE,
    description     TEXT,
    frequency_days  INTEGER NOT NULL DEFAULT 7
);

CREATE TABLE spider_bot.colony_maintenance_records
(
    id                  SERIAL PRIMARY KEY,
    colony_id           INTEGER NOT NULL REFERENCES spider_bot.cricket_colonies(id),
    maintenance_type_id INTEGER NOT NULL REFERENCES spider_bot.colony_maintenance_types(id),
    maintenance_date    DATE NOT NULL,
    notes               TEXT,
    user_id             BIGINT NOT NULL REFERENCES spider_bot.telegram_users(telegram_id),
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_maintenance_records_colony ON spider_bot.colony_maintenance_records(colony_id);
CREATE INDEX idx_maintenance_records_type ON spider_bot.colony_maintenance_records(maintenance_type_id);
CREATE INDEX idx_maintenance_records_date ON spider_bot.colony_maintenance_records(maintenance_date);
CREATE INDEX idx_maintenance_records_user ON spider_bot.colony_maintenance_records(user_id);

CREATE TABLE spider_bot.colony_maintenance_schedules
(
    id                  SERIAL PRIMARY KEY,
    colony_id           INTEGER NOT NULL REFERENCES spider_bot.cricket_colonies(id),
    maintenance_type_id INTEGER NOT NULL REFERENCES spider_bot.colony_maintenance_types(id),
    frequency_days      INTEGER NOT NULL DEFAULT 7,
    enabled             BOOLEAN NOT NULL DEFAULT TRUE,
    last_performed_date DATE,
    next_due_date       DATE,
    user_id             BIGINT NOT NULL REFERENCES spider_bot.telegram_users(telegram_id),
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_maintenance_schedules_colony ON spider_bot.colony_maintenance_schedules(colony_id);
CREATE INDEX idx_maintenance_schedules_type ON spider_bot.colony_maintenance_schedules(maintenance_type_id);
CREATE INDEX idx_maintenance_schedules_user ON spider_bot.colony_maintenance_schedules(user_id);
CREATE INDEX idx_maintenance_schedules_due ON spider_bot.colony_maintenance_schedules(next_due_date);

CREATE TABLE spider_bot.colony_maintenance_settings
(
    id                          SERIAL PRIMARY KEY,
    user_id                     BIGINT NOT NULL UNIQUE REFERENCES spider_bot.telegram_users(telegram_id),
    maintenance_reminder_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    count_frequency_days         INTEGER NOT NULL DEFAULT 7,
    food_water_frequency_days    INTEGER NOT NULL DEFAULT 2,
    cleaning_frequency_days      INTEGER NOT NULL DEFAULT 14,
    environmental_frequency_days INTEGER NOT NULL DEFAULT 3,
    egg_collection_frequency_days INTEGER NOT NULL DEFAULT 5,
    adult_removal_frequency_days INTEGER NOT NULL DEFAULT 7,
    created_at                   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at                   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_maintenance_settings_user ON spider_bot.colony_maintenance_settings(user_id);

CREATE TRIGGER update_colony_maintenance_schedules_updated_at
    BEFORE UPDATE
    ON spider_bot.colony_maintenance_schedules
    FOR EACH ROW
EXECUTE FUNCTION spider_bot.update_updated_at_column();

CREATE TRIGGER update_colony_maintenance_settings_updated_at
    BEFORE UPDATE
    ON spider_bot.colony_maintenance_settings
    FOR EACH ROW
EXECUTE FUNCTION spider_bot.update_updated_at_column();

INSERT INTO spider_bot.colony_maintenance_types (id, type_name, description, frequency_days)
VALUES
    (1, 'Count', 'Count the current number of crickets', 7),
    (2, 'FoodWater', 'Check and replace food and water', 2),
    (3, 'Cleaning', 'Clean the cricket enclosure', 14),
    (6, 'AdultRemoval', 'Remove adult crickets past feeding age', 7)
ON CONFLICT (id) DO NOTHING;

SELECT setval('spider_bot.colony_maintenance_types_id_seq', (SELECT MAX(id) FROM spider_bot.colony_maintenance_types));

-- Add enhanced tracking fields to tarantulas table
ALTER TABLE spider_bot.tarantulas 
    ADD COLUMN IF NOT EXISTS profile_photo_url VARCHAR(255),
    ADD COLUMN IF NOT EXISTS current_weight_grams FLOAT,
    ADD COLUMN IF NOT EXISTS last_weigh_date TIMESTAMP;

-- Create weight_records table for tracking weight history
CREATE TABLE IF NOT EXISTS spider_bot.weight_records (
    id SERIAL PRIMARY KEY,
    tarantula_id INTEGER NOT NULL REFERENCES spider_bot.tarantulas(id) ON DELETE CASCADE,
    weight_grams FLOAT NOT NULL,
    weigh_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    user_id BIGINT NOT NULL REFERENCES spider_bot.telegram_users(telegram_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_weight_records_tarantula ON spider_bot.weight_records(tarantula_id);
CREATE INDEX IF NOT EXISTS idx_weight_records_date ON spider_bot.weight_records(weigh_date);
CREATE INDEX IF NOT EXISTS idx_weight_records_user ON spider_bot.weight_records(user_id);

-- Create tarantula_photos table for photo management
CREATE TABLE IF NOT EXISTS spider_bot.tarantula_photos (
    id SERIAL PRIMARY KEY,
    tarantula_id INTEGER NOT NULL REFERENCES spider_bot.tarantulas(id) ON DELETE CASCADE,
    photo_url VARCHAR(255) NOT NULL,
    photo_type VARCHAR(50) DEFAULT 'general',
    caption TEXT,
    taken_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id BIGINT NOT NULL REFERENCES spider_bot.telegram_users(telegram_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tarantula_photos_tarantula ON spider_bot.tarantula_photos(tarantula_id);
CREATE INDEX IF NOT EXISTS idx_tarantula_photos_date ON spider_bot.tarantula_photos(taken_date);
CREATE INDEX IF NOT EXISTS idx_tarantula_photos_user ON spider_bot.tarantula_photos(user_id);
CREATE INDEX IF NOT EXISTS idx_tarantula_photos_type ON spider_bot.tarantula_photos(photo_type);

-- Insert feeding statuses if they don't exist
INSERT INTO spider_bot.feeding_statuses (id, status_name, description) VALUES
(1, 'Accepted', 'Tarantula accepted and consumed the prey'),
(2, 'Rejected', 'Tarantula refused the prey item'),
(3, 'Partial', 'Tarantula partially consumed the prey')
ON CONFLICT (status_name) DO NOTHING;
