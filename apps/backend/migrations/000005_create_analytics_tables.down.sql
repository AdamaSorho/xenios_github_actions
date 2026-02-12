-- Drop rules on events_audit
DROP RULE IF EXISTS events_audit_no_delete ON events_audit;
DROP RULE IF EXISTS events_audit_no_update ON events_audit;

-- Drop triggers
DROP TRIGGER IF EXISTS update_coaching_analytics_updated_at ON coaching_analytics;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS events_audit;
DROP TABLE IF EXISTS client_risk_scores;
DROP TABLE IF EXISTS coaching_analytics;
