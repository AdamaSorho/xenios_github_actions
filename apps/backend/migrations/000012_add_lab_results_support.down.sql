-- Drop indexes
DROP INDEX IF EXISTS idx_measurements_flag;
DROP INDEX IF EXISTS idx_measurements_artifact_id;

-- Remove columns from measurements table
ALTER TABLE measurements DROP COLUMN IF EXISTS flag;
ALTER TABLE measurements DROP COLUMN IF EXISTS reference_high;
ALTER TABLE measurements DROP COLUMN IF EXISTS reference_low;
ALTER TABLE measurements DROP COLUMN IF EXISTS artifact_id;

-- Note: PostgreSQL does not support removing values from enums.
-- The 'extract_lab_results' value will remain in the job_type enum.
