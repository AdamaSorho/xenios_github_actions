-- Remove artifact_id column from measurements table
DROP INDEX IF EXISTS idx_measurements_artifact_id;
ALTER TABLE measurements DROP COLUMN IF EXISTS artifact_id;
