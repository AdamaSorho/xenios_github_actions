-- Drop triggers
DROP TRIGGER IF EXISTS update_insight_cards_updated_at ON insight_cards;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS wearable_summaries;
DROP TABLE IF EXISTS insight_cards;
DROP TABLE IF EXISTS measurements;
-- Note: artifacts table drop moved to migration 000010_create_artifacts_table.down.sql
