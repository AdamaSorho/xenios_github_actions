-- Remove artifact_id column from measurements
DROP INDEX IF EXISTS idx_measurements_artifact_id;

ALTER TABLE measurements DROP COLUMN IF EXISTS artifact_id;

-- Note: PostgreSQL does not support removing enum values.
-- The extract_inbody enum value will remain but become unused.
