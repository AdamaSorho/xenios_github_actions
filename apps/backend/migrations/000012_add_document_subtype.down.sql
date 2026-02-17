-- Remove document_subtype index and column from artifacts table
DROP INDEX IF EXISTS idx_artifacts_document_subtype;
ALTER TABLE artifacts DROP COLUMN IF EXISTS document_subtype;

-- Note: PostgreSQL does not support removing enum values directly.
-- The new job_type enum values (extract_inbody, extract_lab_results, etc.)
-- will remain in the enum but won't be used after rollback.
