-- Drop triggers
DROP TRIGGER IF EXISTS update_behavior_cues_updated_at ON behavior_cues;
DROP TRIGGER IF EXISTS update_behavior_goals_updated_at ON behavior_goals;
DROP TRIGGER IF EXISTS update_programmed_exercises_updated_at ON programmed_exercises;
DROP TRIGGER IF EXISTS update_programmed_sessions_updated_at ON programmed_sessions;
DROP TRIGGER IF EXISTS update_microcycles_updated_at ON microcycles;
DROP TRIGGER IF EXISTS update_phases_updated_at ON phases;
DROP TRIGGER IF EXISTS update_programs_updated_at ON programs;
DROP TRIGGER IF EXISTS update_exercise_library_updated_at ON exercise_library;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS program_adjustments;
DROP TABLE IF EXISTS behavior_checkins;
DROP TABLE IF EXISTS behavior_cues;
DROP TABLE IF EXISTS behavior_goals;
DROP TABLE IF EXISTS exercise_logs;
DROP TABLE IF EXISTS session_completions;
DROP TABLE IF EXISTS programmed_exercises;
DROP TABLE IF EXISTS programmed_sessions;
DROP TABLE IF EXISTS microcycles;
DROP TABLE IF EXISTS phases;
DROP TABLE IF EXISTS program_versions;
DROP TABLE IF EXISTS programs;
DROP TABLE IF EXISTS exercise_library;
