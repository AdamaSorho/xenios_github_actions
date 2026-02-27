-- Remove document_subtype index and column from artifacts table
DROP INDEX IF EXISTS idx_artifacts_document_subtype;

ALTER TABLE artifacts DROP COLUMN IF EXISTS document_subtype;

-- Note: PostgreSQL does not support removing values from enums.
-- The added job types (extract_inbody, extract_lab_results, extract_wearable,
-- extract_nutrition, transcribe_audio, classify_document) will remain in the
-- job_type enum but will be unused.
