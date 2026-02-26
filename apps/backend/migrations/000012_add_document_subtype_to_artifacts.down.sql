-- Remove document_subtype index
DROP INDEX IF EXISTS idx_artifacts_document_subtype;

-- Remove check constraint
ALTER TABLE artifacts DROP CONSTRAINT IF EXISTS chk_artifacts_document_subtype;

-- Remove document_subtype column
ALTER TABLE artifacts DROP COLUMN IF EXISTS document_subtype;

-- Note: PostgreSQL does not support removing enum values.
-- The new job_type enum values (extract_inbody, extract_lab_results, etc.)
-- will remain but are harmless if unused.
