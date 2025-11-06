-- Migration 0003: Add molt notification features
-- This migration adds support for:
-- 1. Post-molt mute windows (suppress feeding notifications after molt)
-- 2. Molt prediction notifications

-- Add post-molt mute tracking to tarantulas table
ALTER TABLE spider_bot.tarantulas
    ADD COLUMN IF NOT EXISTS post_molt_mute_until TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_tarantulas_mute_until ON spider_bot.tarantulas(post_molt_mute_until);

-- Add molt notification settings to user_settings table
ALTER TABLE spider_bot.user_settings
    ADD COLUMN IF NOT EXISTS molt_prediction_enabled BOOLEAN DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS molt_prediction_days INTEGER DEFAULT 5,
    ADD COLUMN IF NOT EXISTS post_molt_mute_days INTEGER DEFAULT 7;

-- Add comments for documentation
COMMENT ON COLUMN spider_bot.tarantulas.post_molt_mute_until IS 'Feeding notifications will be suppressed until this date';
COMMENT ON COLUMN spider_bot.user_settings.molt_prediction_enabled IS 'Enable notifications for upcoming predicted molts';
COMMENT ON COLUMN spider_bot.user_settings.molt_prediction_days IS 'Number of days before predicted molt to send notification';
COMMENT ON COLUMN spider_bot.user_settings.post_molt_mute_days IS 'Number of days after molt to suppress feeding notifications';
