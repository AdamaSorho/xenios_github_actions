DROP TRIGGER IF EXISTS update_nutrition_summaries_updated_at ON nutrition_summaries;
DROP TABLE IF EXISTS nutrition_summaries;
DROP INDEX IF EXISTS idx_measurements_artifact_id;

DO $$ BEGIN
    ALTER TABLE measurements DROP COLUMN IF EXISTS artifact_id;
EXCEPTION WHEN undefined_column THEN NULL;
END $$;
