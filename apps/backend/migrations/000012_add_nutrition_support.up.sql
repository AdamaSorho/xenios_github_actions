-- Add source_artifact_id column to measurements table for linking nutrition imports
DO $$ BEGIN
    ALTER TABLE measurements ADD COLUMN source_artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL;
EXCEPTION
    WHEN duplicate_column THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_measurements_source_artifact_id ON measurements(source_artifact_id);

-- Create nutrition_averages table for computed macro averages
CREATE TABLE IF NOT EXISTS nutrition_averages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL,
    period_days INT NOT NULL CHECK (period_days > 0),
    avg_calories NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_protein NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_carbs NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_fat NUMERIC(10,3) NOT NULL DEFAULT 0,
    avg_fiber NUMERIC(10,3) NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_nutrition_averages_client_id ON nutrition_averages(client_id);
CREATE INDEX IF NOT EXISTS idx_nutrition_averages_artifact_id ON nutrition_averages(source_artifact_id);

ALTER TABLE nutrition_averages ENABLE ROW LEVEL SECURITY;

-- RLS policy: coaches can access averages for their clients
DROP POLICY IF EXISTS nutrition_averages_coach_access ON nutrition_averages;
CREATE POLICY nutrition_averages_coach_access ON nutrition_averages
    FOR ALL
    USING (
        client_id IN (
            SELECT cc.client_id FROM coach_clients cc
            WHERE cc.coach_id = current_setting('app.current_user_id', true)::UUID
        )
    );

-- Add extract_nutrition to job_type enum
DO $$ BEGIN
    ALTER TYPE job_type ADD VALUE 'extract_nutrition';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;
