-- Drop triggers
DROP TRIGGER IF EXISTS update_workout_exercises_updated_at ON workout_exercises;
DROP TRIGGER IF EXISTS update_sessions_updated_at ON sessions;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS form_cues_tracking;
DROP TABLE IF EXISTS workout_exercises;
DROP TABLE IF EXISTS transcript_segments;
DROP TABLE IF EXISTS sessions;
