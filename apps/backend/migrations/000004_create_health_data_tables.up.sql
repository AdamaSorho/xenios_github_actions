-- Artifacts (files, images, documents attached to clients)
CREATE TABLE IF NOT EXISTS artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    artifact_type TEXT NOT NULL CHECK (artifact_type IN ('image', 'video', 'document', 'audio')),
    file_url TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_size_bytes BIGINT,
    mime_type TEXT,
    description TEXT,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_artifacts_client_id ON artifacts(client_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_uploaded_by ON artifacts(uploaded_by);
CREATE INDEX IF NOT EXISTS idx_artifacts_expires_at ON artifacts(expires_at);

ALTER TABLE artifacts ENABLE ROW LEVEL SECURITY;

-- Measurements (body measurements, vitals, etc.)
CREATE TABLE IF NOT EXISTS measurements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recorded_by UUID NOT NULL REFERENCES users(id),
    measurement_type TEXT NOT NULL,
    value NUMERIC(10,3) NOT NULL,
    unit TEXT NOT NULL,
    measured_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_measurements_client_id ON measurements(client_id);
CREATE INDEX IF NOT EXISTS idx_measurements_type ON measurements(measurement_type);
CREATE INDEX IF NOT EXISTS idx_measurements_measured_at ON measurements(measured_at);

ALTER TABLE measurements ENABLE ROW LEVEL SECURITY;

-- Insight cards (AI-generated coaching insights, require approval)
CREATE TABLE IF NOT EXISTS insight_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general' CHECK (category IN ('general', 'nutrition', 'recovery', 'performance', 'behavior', 'safety')),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'approved', 'dismissed', 'shared')),
    priority TEXT NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    approved_at TIMESTAMPTZ,
    shared_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_insight_cards_coach_id ON insight_cards(coach_id);
CREATE INDEX IF NOT EXISTS idx_insight_cards_client_id ON insight_cards(client_id);
CREATE INDEX IF NOT EXISTS idx_insight_cards_session_id ON insight_cards(session_id);
CREATE INDEX IF NOT EXISTS idx_insight_cards_status ON insight_cards(status);

ALTER TABLE insight_cards ENABLE ROW LEVEL SECURITY;

CREATE TRIGGER update_insight_cards_updated_at
    BEFORE UPDATE ON insight_cards
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Wearable summaries (aggregated data from wearable devices)
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
