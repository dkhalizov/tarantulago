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
    ADD COLUMN maintenance_reminder_enabled BOOLEAN DEFAULT TRUE,
ADD COLUMN food_water_frequency_days INTEGER DEFAULT 3,
ADD COLUMN cleaning_frequency_days INTEGER DEFAULT 14,
ADD COLUMN adult_removal_frequency_days INTEGER DEFAULT 7;

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
    (6, 'AdultRemoval', 'Remove adult crickets past feeding age', 7);

SELECT setval('spider_bot.colony_maintenance_types_id_seq', (SELECT MAX(id) FROM spider_bot.colony_maintenance_types));
