DROP TRIGGER IF EXISTS trg_insight_cards_updated_at ON insight_cards;
DROP POLICY IF EXISTS insight_evidence_access ON insight_evidence;
DROP POLICY IF EXISTS insight_cards_client_access ON insight_cards;
DROP POLICY IF EXISTS insight_cards_coach_access ON insight_cards;
DROP TABLE IF EXISTS insight_evidence;
DROP TABLE IF EXISTS insight_cards;
