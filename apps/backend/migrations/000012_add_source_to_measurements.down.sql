-- Remove source column and related indexes from measurements.

DROP INDEX IF EXISTS idx_measurements_unique_wearable;
DROP INDEX IF EXISTS idx_measurements_source;
ALTER TABLE measurements DROP COLUMN IF EXISTS source;
