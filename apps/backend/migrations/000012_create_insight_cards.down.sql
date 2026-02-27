-- 000012_create_insight_cards.down.sql
-- Reverses the approval queue schema additions.

DROP TABLE IF EXISTS insight_evidence;

ALTER TABLE insight_cards DROP COLUMN IF EXISTS client_name;
ALTER TABLE insight_cards DROP COLUMN IF EXISTS dismissed_at;

DROP INDEX IF EXISTS idx_insight_cards_coach_status_priority;
