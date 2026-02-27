-- Add flag and reference range columns to measurements for out-of-range detection
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS flag TEXT NOT NULL DEFAULT 'normal'
    CHECK (flag IN ('normal', 'low', 'high', 'critical_low', 'critical_high'));
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS ref_range_low NUMERIC(10,3);
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS ref_range_high NUMERIC(10,3);
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_measurements_flag ON measurements(flag);
CREATE INDEX IF NOT EXISTS idx_measurements_artifact_id ON measurements(artifact_id);

-- Insight evidence: links insight cards to the measurements/artifacts they reference
CREATE TABLE IF NOT EXISTS insight_evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    insight_card_id UUID NOT NULL REFERENCES insight_cards(id) ON DELETE CASCADE,
    measurement_id UUID REFERENCES measurements(id) ON DELETE SET NULL,
    artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_insight_evidence_card_id ON insight_evidence(insight_card_id);
CREATE INDEX IF NOT EXISTS idx_insight_evidence_measurement_id ON insight_evidence(measurement_id);

ALTER TABLE insight_evidence ENABLE ROW LEVEL SECURITY;
