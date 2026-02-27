-- Remove artifact_id column from measurements
ALTER TABLE measurements DROP COLUMN IF EXISTS artifact_id;

-- Revert artifact status constraint to original values
ALTER TABLE artifacts DROP CONSTRAINT IF EXISTS artifacts_status_check;

DO $$ BEGIN
    ALTER TABLE artifacts ADD CONSTRAINT artifacts_status_check
        CHECK (status IN ('pending', 'uploaded', 'failed'));
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Note: PostgreSQL does not support removing values from an existing enum type.
-- The extract_inbody value will remain in the job_type enum.
-- To fully remove it, the enum would need to be recreated, which is beyond
-- the scope of a safe down migration.
