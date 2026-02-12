-- Coaching analytics (aggregated metrics for coach dashboards)
CREATE TABLE IF NOT EXISTS coaching_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_sessions INTEGER NOT NULL DEFAULT 0,
    completed_sessions INTEGER NOT NULL DEFAULT 0,
    cancelled_sessions INTEGER NOT NULL DEFAULT 0,
    avg_session_duration_minutes NUMERIC(6,1),
    adherence_score NUMERIC(5,2),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(coach_id, client_id, period_start, period_end)
);

CREATE INDEX IF NOT EXISTS idx_coaching_analytics_coach_id ON coaching_analytics(coach_id);
CREATE INDEX IF NOT EXISTS idx_coaching_analytics_client_id ON coaching_analytics(client_id);
CREATE INDEX IF NOT EXISTS idx_coaching_analytics_period ON coaching_analytics(period_start, period_end);

ALTER TABLE coaching_analytics ENABLE ROW LEVEL SECURITY;

DROP TRIGGER IF EXISTS update_coaching_analytics_updated_at ON coaching_analytics;
CREATE TRIGGER update_coaching_analytics_updated_at
    BEFORE UPDATE ON coaching_analytics
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Client risk scores (AI-generated risk assessments)
CREATE TABLE IF NOT EXISTS client_risk_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    overall_score NUMERIC(5,2) NOT NULL,
    factors JSONB NOT NULL DEFAULT '{}',
    recommendations TEXT[] DEFAULT '{}',
    assessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_client_risk_scores_coach_id ON client_risk_scores(coach_id);
CREATE INDEX IF NOT EXISTS idx_client_risk_scores_client_id ON client_risk_scores(client_id);
CREATE INDEX IF NOT EXISTS idx_client_risk_scores_assessed_at ON client_risk_scores(assessed_at);

ALTER TABLE client_risk_scores ENABLE ROW LEVEL SECURITY;

-- Events audit log (append-only, immutable)
CREATE TABLE IF NOT EXISTS events_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    metadata JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_audit_actor_id ON events_audit(actor_id);
CREATE INDEX IF NOT EXISTS idx_events_audit_entity ON events_audit(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_events_audit_entity_time ON events_audit(entity_type, entity_id, created_at);
CREATE INDEX IF NOT EXISTS idx_events_audit_action ON events_audit(action);
CREATE INDEX IF NOT EXISTS idx_events_audit_created_at ON events_audit(created_at);

ALTER TABLE events_audit ENABLE ROW LEVEL SECURITY;

-- Append-only rule: prevent UPDATE and DELETE on events_audit
CREATE OR REPLACE RULE events_audit_no_update AS ON UPDATE TO events_audit DO INSTEAD NOTHING;
CREATE OR REPLACE RULE events_audit_no_delete AS ON DELETE TO events_audit DO INSTEAD NOTHING;
