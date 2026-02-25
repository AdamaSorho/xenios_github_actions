-- Remove document_subtype index
DROP INDEX IF EXISTS idx_artifacts_document_subtype;

-- Remove document_subtype column from artifacts table
ALTER TABLE artifacts DROP COLUMN IF EXISTS document_subtype;

-- Note: PostgreSQL does not support removing values from an enum type.
-- The new job_type enum values (extract_inbody, extract_lab_results, etc.)
-- will remain in the enum but will be unused after rollback.
