-- Remove artifact_id column and partial_extraction flag from measurements.
ALTER TABLE measurements DROP COLUMN IF EXISTS partial_extraction;
DROP INDEX IF EXISTS idx_measurements_artifact_id;
ALTER TABLE measurements DROP COLUMN IF EXISTS artifact_id;
