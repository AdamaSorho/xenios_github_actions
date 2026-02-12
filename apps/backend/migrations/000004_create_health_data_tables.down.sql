-- Drop triggers
DROP TRIGGER IF EXISTS update_insight_cards_updated_at ON insight_cards;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS wearable_summaries;
DROP TABLE IF EXISTS insight_cards;
DROP TABLE IF EXISTS measurements;
DROP TABLE IF EXISTS artifacts;
