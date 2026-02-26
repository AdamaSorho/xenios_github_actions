-- Remove lab result reference range and flag columns from measurements
DROP INDEX IF EXISTS idx_measurements_flag;
DROP INDEX IF EXISTS idx_measurements_artifact_id;

ALTER TABLE measurements DROP COLUMN IF EXISTS flag;
ALTER TABLE measurements DROP COLUMN IF EXISTS reference_high;
ALTER TABLE measurements DROP COLUMN IF EXISTS reference_low;
ALTER TABLE measurements DROP COLUMN IF EXISTS artifact_id;
