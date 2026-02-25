DROP INDEX IF EXISTS idx_measurements_client_measured_at;
DROP INDEX IF EXISTS idx_measurements_client_type;
DROP INDEX IF EXISTS idx_measurements_artifact_id;

ALTER TABLE measurements DROP COLUMN IF EXISTS reference_high;
ALTER TABLE measurements DROP COLUMN IF EXISTS reference_low;
ALTER TABLE measurements DROP COLUMN IF EXISTS flag;
ALTER TABLE measurements DROP COLUMN IF EXISTS artifact_id;
