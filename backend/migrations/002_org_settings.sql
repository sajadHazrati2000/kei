-- Add working-hours and language defaults to the organizations table.
-- Safe to run on existing databases: ADD COLUMN IF NOT EXISTS is idempotent.
ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS default_language TEXT  NOT NULL DEFAULT 'en',
    ADD COLUMN IF NOT EXISTS working_days     INT[] NOT NULL DEFAULT '{1,2,3,4,5}',
    ADD COLUMN IF NOT EXISTS working_start    TIMETZ NOT NULL DEFAULT '09:00+00',
    ADD COLUMN IF NOT EXISTS working_end      TIMETZ NOT NULL DEFAULT '17:00+00';
