-- Daily nutrition records (individual metric measurements from CSV imports)
CREATE TABLE IF NOT EXISTS nutrition_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    coach_id UUID NOT NULL REFERENCES users(id),
    artifact_id UUID NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    metric_type TEXT NOT NULL CHECK (metric_type IN ('calories', 'protein', 'carbs', 'fat', 'fiber')),
    value NUMERIC(10,3) NOT NULL,
    unit TEXT NOT NULL,
    record_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_nutrition_records_client_id ON nutrition_records(client_id);
CREATE INDEX IF NOT EXISTS idx_nutrition_records_coach_id ON nutrition_records(coach_id);
CREATE INDEX IF NOT EXISTS idx_nutrition_records_artifact_id ON nutrition_records(artifact_id);
CREATE INDEX IF NOT EXISTS idx_nutrition_records_date ON nutrition_records(record_date);
CREATE INDEX IF NOT EXISTS idx_nutrition_records_type_date ON nutrition_records(client_id, metric_type, record_date);

ALTER TABLE nutrition_records ENABLE ROW LEVEL SECURITY;

-- Nutrition summaries (computed rolling averages)
CREATE TABLE IF NOT EXISTS nutrition_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    artifact_id UUID NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    total_days INTEGER NOT NULL DEFAULT 0,
    avg_calories_7d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_protein_7d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_carbs_7d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_fat_7d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_fiber_7d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_calories_14d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_protein_14d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_carbs_14d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_fat_14d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_fiber_14d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_calories_30d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_protein_30d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_carbs_30d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_fat_30d NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_fiber_30d NUMERIC(10,3) NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(client_id, artifact_id)
);

CREATE INDEX IF NOT EXISTS idx_nutrition_summaries_client_id ON nutrition_summaries(client_id);
CREATE INDEX IF NOT EXISTS idx_nutrition_summaries_artifact_id ON nutrition_summaries(artifact_id);

ALTER TABLE nutrition_summaries ENABLE ROW LEVEL SECURITY;
