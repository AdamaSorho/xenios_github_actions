-- Remove extraction_status column
ALTER TABLE measurements DROP COLUMN IF EXISTS extraction_status;

-- Remove artifact_id column and index
DROP INDEX IF EXISTS idx_measurements_artifact_id;
ALTER TABLE measurements DROP COLUMN IF EXISTS artifact_id;
