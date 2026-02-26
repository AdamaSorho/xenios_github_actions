-- Create insight cards table for AI-generated insight approval queue
CREATE TABLE IF NOT EXISTS insight_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES users(id),
    coach_id UUID NOT NULL REFERENCES users(id),
    client_name TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('nutrition', 'exercise', 'sleep', 'stress', 'general')),
    priority TEXT NOT NULL DEFAULT 'medium' CHECK (priority IN ('urgent', 'high', 'medium', 'low')),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'approved', 'dismissed', 'shared')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    dismissed_at TIMESTAMPTZ,
    shared_at TIMESTAMPTZ
);

-- Create insight evidence table for linked measurements
CREATE TABLE IF NOT EXISTS insight_evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    insight_id UUID NOT NULL REFERENCES insight_cards(id) ON DELETE CASCADE,
    measurement_id UUID NOT NULL,
    description TEXT NOT NULL DEFAULT ''
);

-- Indexes for insight cards
CREATE INDEX IF NOT EXISTS idx_insight_cards_coach_id_status ON insight_cards(coach_id, status);
CREATE INDEX IF NOT EXISTS idx_insight_cards_client_id_status ON insight_cards(client_id, status);
CREATE INDEX IF NOT EXISTS idx_insight_cards_priority ON insight_cards(priority);
CREATE INDEX IF NOT EXISTS idx_insight_cards_created_at ON insight_cards(created_at DESC);

-- Index for insight evidence
CREATE INDEX IF NOT EXISTS idx_insight_evidence_insight_id ON insight_evidence(insight_id);

-- Trigger for auto-updating updated_at
DROP TRIGGER IF EXISTS trg_insight_cards_updated_at ON insight_cards;
CREATE TRIGGER trg_insight_cards_updated_at
    BEFORE UPDATE ON insight_cards
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Enable Row Level Security
ALTER TABLE insight_cards ENABLE ROW LEVEL SECURITY;
ALTER TABLE insight_evidence ENABLE ROW LEVEL SECURITY;

-- RLS policy: coaches can access insight cards they own
DROP POLICY IF EXISTS insight_cards_coach_access ON insight_cards;
CREATE POLICY insight_cards_coach_access ON insight_cards
    FOR ALL
    USING (coach_id = current_setting('app.current_user_id', true)::UUID);

-- RLS policy: clients can only see approved/shared insight cards
DROP POLICY IF EXISTS insight_cards_client_access ON insight_cards;
CREATE POLICY insight_cards_client_access ON insight_cards
    FOR SELECT
    USING (client_id = current_setting('app.current_user_id', true)::UUID AND status IN ('approved', 'shared'));

-- RLS policy: evidence follows insight card access
DROP POLICY IF EXISTS insight_evidence_access ON insight_evidence;
CREATE POLICY insight_evidence_access ON insight_evidence
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM insight_cards ic
            WHERE ic.id = insight_evidence.insight_id
              AND (ic.coach_id = current_setting('app.current_user_id', true)::UUID
                OR (ic.client_id = current_setting('app.current_user_id', true)::UUID AND ic.status IN ('approved', 'shared')))
        )
    );
