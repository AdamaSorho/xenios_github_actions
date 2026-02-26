-- Add document_subtype column to artifacts table for file classification routing.
-- Uses TEXT with a CHECK constraint (not an enum) to avoid ALTER TYPE complexity.
DO $$ BEGIN
    ALTER TABLE artifacts ADD COLUMN document_subtype TEXT
        CHECK (document_subtype IN (
            'inbody_pdf', 'lab_csv', 'lab_pdf',
            'wearable_csv', 'wearable_json',
            'nutrition_csv', 'audio', 'other'
        ));
EXCEPTION
    WHEN duplicate_column THEN NULL;
END $$;

-- Add new job types to the job_type enum for extraction routing.
DO $$ BEGIN ALTER TYPE job_type ADD VALUE 'extract_inbody'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE job_type ADD VALUE 'extract_lab_results'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE job_type ADD VALUE 'extract_wearable'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE job_type ADD VALUE 'extract_nutrition'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE job_type ADD VALUE 'transcribe_audio'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE job_type ADD VALUE 'classify_document'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- Index for querying artifacts by document subtype
CREATE INDEX IF NOT EXISTS idx_artifacts_document_subtype ON artifacts(document_subtype)
    WHERE document_subtype IS NOT NULL;
