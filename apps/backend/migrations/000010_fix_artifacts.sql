-- Fix script for artifacts table migration
-- Run this manually in staging before retrying the migration

-- Drop the table if it exists in an incomplete state
DROP TABLE IF EXISTS artifacts CASCADE;

-- Now the 000010_create_artifacts_table.up.sql migration can run cleanly
