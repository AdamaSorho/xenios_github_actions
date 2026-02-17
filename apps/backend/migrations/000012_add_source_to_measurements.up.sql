-- Add source column to measurements for tracking wearable device origin.
-- This supports the wearable CSV/JSON import feature.

ALTER TABLE measurements ADD COLUMN IF NOT EXISTS source TEXT;

-- Index for efficient queries by source
CREATE INDEX IF NOT EXISTS idx_measurements_source ON measurements(source);

-- Composite unique constraint to prevent duplicate imports:
-- Same client, same source, same measurement type, same date should not have duplicates.
-- Note: We use a partial unique index since source may be NULL for manually entered measurements.
CREATE UNIQUE INDEX IF NOT EXISTS idx_measurements_unique_wearable
    ON measurements(client_id, source, measurement_type, (measured_at::date))
    WHERE source IS NOT NULL;
