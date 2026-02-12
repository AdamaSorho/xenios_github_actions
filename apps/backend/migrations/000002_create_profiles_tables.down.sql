-- Drop triggers
DROP TRIGGER IF EXISTS update_ccr_updated_at ON coach_client_relationships;
DROP TRIGGER IF EXISTS update_client_profiles_updated_at ON client_profiles;
DROP TRIGGER IF EXISTS update_coach_profiles_updated_at ON coach_profiles;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS coach_client_relationships;
DROP TABLE IF EXISTS client_profiles;
DROP TABLE IF EXISTS coach_profiles;
