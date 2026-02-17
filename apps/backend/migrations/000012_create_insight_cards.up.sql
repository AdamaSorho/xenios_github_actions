-- Create insight_cards table for storing AI-generated draft insights
CREATE TABLE IF NOT EXISTS insight_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL REFERENCES users(id),
    client_id UUID NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('nutrition', 'recovery', 'performance', 'safety')),
    priority TEXT NOT NULL CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'approved', 'rejected')),
    evidence JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Enable Row Level Security
ALTER TABLE insight_cards ENABLE ROW LEVEL SECURITY;

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_insight_cards_client_id ON insight_cards(client_id);
CREATE INDEX IF NOT EXISTS idx_insight_cards_coach_id ON insight_cards(coach_id);
CREATE INDEX IF NOT EXISTS idx_insight_cards_status ON insight_cards(status);
CREATE INDEX IF NOT EXISTS idx_insight_cards_created_at ON insight_cards(created_at);
CREATE INDEX IF NOT EXISTS idx_insight_cards_priority ON insight_cards(priority);

-- GIN index for searching evidence JSONB by measurement_id (duplicate prevention)
CREATE INDEX IF NOT EXISTS idx_insight_cards_evidence ON insight_cards USING gin(evidence);

-- Trigger to auto-update updated_at on row changes
CREATE OR REPLACE FUNCTION update_insight_cards_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_insight_cards_updated_at ON insight_cards;
CREATE TRIGGER trg_insight_cards_updated_at
    BEFORE UPDATE ON insight_cards
    FOR EACH ROW
    EXECUTE FUNCTION update_insight_cards_updated_at();
