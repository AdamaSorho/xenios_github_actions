-- Add artifact_id column to measurements table for linking to source files
DO $$ BEGIN
    ALTER TABLE measurements ADD COLUMN artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_measurements_artifact_id ON measurements(artifact_id);

-- Nutrition summaries (computed rolling averages from nutrition CSV imports)
CREATE TABLE IF NOT EXISTS nutrition_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    artifact_id UUID NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    avg_calories_7d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_protein_7d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_carbs_7d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_fat_7d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_fiber_7d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_calories_14d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_protein_14d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_carbs_14d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_fat_14d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_fiber_14d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_calories_30d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_protein_30d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_carbs_30d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_fat_30d NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_fiber_30d NUMERIC(10,2) NOT NULL DEFAULT 0,
    total_days INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(client_id, artifact_id)
);

CREATE INDEX IF NOT EXISTS idx_nutrition_summaries_client_id ON nutrition_summaries(client_id);
CREATE INDEX IF NOT EXISTS idx_nutrition_summaries_artifact_id ON nutrition_summaries(artifact_id);

ALTER TABLE nutrition_summaries ENABLE ROW LEVEL SECURITY;

DROP TRIGGER IF EXISTS update_nutrition_summaries_updated_at ON nutrition_summaries;
CREATE TRIGGER update_nutrition_summaries_updated_at
    BEFORE UPDATE ON nutrition_summaries
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
