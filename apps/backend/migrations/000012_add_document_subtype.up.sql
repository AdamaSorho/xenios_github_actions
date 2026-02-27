-- Add document_subtype column to artifacts table for file classification
DO $$ BEGIN
    ALTER TABLE artifacts ADD COLUMN document_subtype TEXT;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

-- Add new job types to the job_type enum for extraction pipeline
DO $$ BEGIN ALTER TYPE job_type ADD VALUE 'extract_inbody'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE job_type ADD VALUE 'extract_lab_results'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE job_type ADD VALUE 'extract_wearable'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE job_type ADD VALUE 'extract_nutrition'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE job_type ADD VALUE 'transcribe_audio'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE job_type ADD VALUE 'classify_document'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- Index for querying artifacts by document_subtype
CREATE INDEX IF NOT EXISTS idx_artifacts_document_subtype ON artifacts(document_subtype) WHERE document_subtype IS NOT NULL;
