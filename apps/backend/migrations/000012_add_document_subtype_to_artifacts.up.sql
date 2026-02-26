-- Add document_subtype column to artifacts table for file type classification
ALTER TABLE artifacts
    ADD COLUMN IF NOT EXISTS document_subtype TEXT;

-- Add check constraint for valid document subtypes
ALTER TABLE artifacts
    ADD CONSTRAINT chk_artifacts_document_subtype
    CHECK (document_subtype IS NULL OR document_subtype IN (
        'inbody_pdf', 'lab_csv', 'lab_pdf',
        'wearable_csv', 'wearable_json',
        'nutrition_csv', 'audio', 'other'
    ));

-- Index for querying artifacts by document subtype
CREATE INDEX IF NOT EXISTS idx_artifacts_document_subtype ON artifacts(document_subtype)
    WHERE document_subtype IS NOT NULL;

-- Add new job types to the job_type enum for extraction workers
ALTER TYPE job_type ADD VALUE IF NOT EXISTS 'extract_inbody';
ALTER TYPE job_type ADD VALUE IF NOT EXISTS 'extract_lab_results';
ALTER TYPE job_type ADD VALUE IF NOT EXISTS 'extract_wearable';
ALTER TYPE job_type ADD VALUE IF NOT EXISTS 'extract_nutrition';
ALTER TYPE job_type ADD VALUE IF NOT EXISTS 'transcribe_audio';
ALTER TYPE job_type ADD VALUE IF NOT EXISTS 'classify_document';
