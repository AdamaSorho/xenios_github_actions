-- 000012_create_insight_cards.up.sql
-- Extends the insight_cards table (created in 000004) with columns needed
-- for the approval queue feature, and adds an evidence link table.

-- Add client_name column for display purposes
DO $$ BEGIN
    ALTER TABLE insight_cards ADD COLUMN client_name TEXT NOT NULL DEFAULT '';
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

-- Add dismissed_at column for tracking dismissal time
DO $$ BEGIN
    ALTER TABLE insight_cards ADD COLUMN dismissed_at TIMESTAMPTZ;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

-- Create evidence table for linking insights to measurements
CREATE TABLE IF NOT EXISTS insight_evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    insight_id UUID NOT NULL REFERENCES insight_cards(id) ON DELETE CASCADE,
    measurement_id UUID,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_insight_evidence_insight_id ON insight_evidence(insight_id);

ALTER TABLE insight_evidence ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS insight_evidence_coach_access ON insight_evidence;
CREATE POLICY insight_evidence_coach_access ON insight_evidence
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM insight_cards ic
            WHERE ic.id = insight_evidence.insight_id
              AND ic.coach_id = current_setting('app.current_user_id', true)::UUID
        )
    );

-- Add composite index for queue query (coach_id + status + priority)
CREATE INDEX IF NOT EXISTS idx_insight_cards_coach_status_priority
    ON insight_cards(coach_id, status, priority);
