-- Create wearable_summaries table for rolling averages
CREATE TABLE IF NOT EXISTS wearable_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL,
    source wearable_source NOT NULL,
    metrics JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- One summary per client/source combination
    UNIQUE (client_id, source)
);

CREATE INDEX IF NOT EXISTS idx_wearable_summaries_client
    ON wearable_summaries (client_id);

-- Enable Row Level Security
ALTER TABLE wearable_summaries ENABLE ROW LEVEL SECURITY;

-- Service role bypass for backend operations
DROP POLICY IF EXISTS wearable_summaries_service_role ON wearable_summaries;
CREATE POLICY wearable_summaries_service_role ON wearable_summaries
    FOR ALL
    TO service_role
    USING (true)
    WITH CHECK (true);

-- Add extract_wearable to job_type enum if it doesn't exist
DO $$ BEGIN
    ALTER TYPE job_type ADD VALUE IF NOT EXISTS 'extract_wearable';
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
