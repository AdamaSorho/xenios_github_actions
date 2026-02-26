DROP TABLE IF EXISTS wearable_summaries;
-- Note: Cannot remove enum value from job_type in PostgreSQL without recreating the type

-- Restore the original generic wearable_summaries table from migration 000004
CREATE TABLE IF NOT EXISTS wearable_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source TEXT NOT NULL,
    summary_date DATE NOT NULL,
    metrics JSONB NOT NULL DEFAULT '{}',
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(client_id, source, summary_date)
);

CREATE INDEX IF NOT EXISTS idx_wearable_summaries_client_id ON wearable_summaries(client_id);
CREATE INDEX IF NOT EXISTS idx_wearable_summaries_date ON wearable_summaries(summary_date);

ALTER TABLE wearable_summaries ENABLE ROW LEVEL SECURITY;
