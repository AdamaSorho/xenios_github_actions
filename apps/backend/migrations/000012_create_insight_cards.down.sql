-- Reverse migration: drop insight_cards table and related objects
DROP TRIGGER IF EXISTS trg_insight_cards_updated_at ON insight_cards;
DROP FUNCTION IF EXISTS update_insight_cards_updated_at();
DROP TABLE IF EXISTS insight_cards;
