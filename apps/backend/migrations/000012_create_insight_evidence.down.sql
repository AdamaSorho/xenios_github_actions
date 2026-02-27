DROP TABLE IF EXISTS insight_evidence;

DROP INDEX IF EXISTS idx_measurements_artifact_id;
DROP INDEX IF EXISTS idx_measurements_flag;

ALTER TABLE measurements DROP COLUMN IF EXISTS artifact_id;
ALTER TABLE measurements DROP COLUMN IF EXISTS ref_range_high;
ALTER TABLE measurements DROP COLUMN IF EXISTS ref_range_low;
ALTER TABLE measurements DROP COLUMN IF EXISTS flag;
