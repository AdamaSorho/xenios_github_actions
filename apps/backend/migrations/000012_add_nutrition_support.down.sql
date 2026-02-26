-- Remove nutrition_averages table
DROP TABLE IF EXISTS nutrition_averages;

-- Remove source_artifact_id column from measurements
ALTER TABLE measurements DROP COLUMN IF EXISTS source_artifact_id;

-- Note: PostgreSQL does not support removing enum values.
-- The extract_nutrition enum value in job_type will remain but is harmless.
