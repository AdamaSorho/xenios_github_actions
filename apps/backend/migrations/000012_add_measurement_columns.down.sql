-- Remove indexes added for client profile API
DROP INDEX IF EXISTS idx_measurements_client_measured_at;
DROP INDEX IF EXISTS idx_measurements_client_type;
DROP INDEX IF EXISTS idx_measurements_artifact_id;

-- Remove columns added for client profile API
DO $$ BEGIN
    ALTER TABLE measurements DROP COLUMN IF EXISTS reference_high;
EXCEPTION WHEN undefined_column THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE measurements DROP COLUMN IF EXISTS reference_low;
EXCEPTION WHEN undefined_column THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE measurements DROP COLUMN IF EXISTS flag;
EXCEPTION WHEN undefined_column THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE measurements DROP COLUMN IF EXISTS artifact_id;
EXCEPTION WHEN undefined_column THEN NULL;
END $$;
