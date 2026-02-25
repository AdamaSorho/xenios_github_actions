-- Insight card evidence references: links insight cards to source measurements/artifacts.
CREATE TABLE IF NOT EXISTS insight_card_evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    insight_card_id UUID NOT NULL REFERENCES insight_cards(id) ON DELETE CASCADE,
    measurement_id UUID NOT NULL,
    artifact_id UUID,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_insight_card_evidence_card_id ON insight_card_evidence(insight_card_id);
CREATE INDEX IF NOT EXISTS idx_insight_card_evidence_measurement_id ON insight_card_evidence(measurement_id);
CREATE INDEX IF NOT EXISTS idx_insight_card_evidence_artifact_id ON insight_card_evidence(artifact_id) WHERE artifact_id IS NOT NULL;

-- Unique constraint to prevent duplicate evidence for the same insight card and measurement.
CREATE UNIQUE INDEX IF NOT EXISTS idx_insight_card_evidence_unique
    ON insight_card_evidence(insight_card_id, measurement_id);

ALTER TABLE insight_card_evidence ENABLE ROW LEVEL SECURITY;
