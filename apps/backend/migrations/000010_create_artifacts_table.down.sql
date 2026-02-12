-- Drop RLS policy
DROP POLICY IF EXISTS artifacts_coach_access ON artifacts;

-- Drop trigger
DROP TRIGGER IF EXISTS update_artifacts_updated_at ON artifacts;

-- Drop table (cascades indexes)
DROP TABLE IF EXISTS artifacts;
