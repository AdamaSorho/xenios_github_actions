-- Drop all RLS policies

-- Program adjustments
DROP POLICY IF EXISTS program_adjustments_access ON program_adjustments;

-- Behavior check-ins
DROP POLICY IF EXISTS behavior_checkins_coach_access ON behavior_checkins;
DROP POLICY IF EXISTS behavior_checkins_client_access ON behavior_checkins;

-- Behavior cues
DROP POLICY IF EXISTS behavior_cues_access ON behavior_cues;

-- Behavior goals
DROP POLICY IF EXISTS behavior_goals_client_access ON behavior_goals;
DROP POLICY IF EXISTS behavior_goals_coach_access ON behavior_goals;

-- Exercise logs
DROP POLICY IF EXISTS exercise_logs_access ON exercise_logs;

-- Session completions
DROP POLICY IF EXISTS session_completions_coach_access ON session_completions;
DROP POLICY IF EXISTS session_completions_client_access ON session_completions;

-- Programmed exercises
DROP POLICY IF EXISTS programmed_exercises_access ON programmed_exercises;

-- Programmed sessions
DROP POLICY IF EXISTS programmed_sessions_access ON programmed_sessions;

-- Microcycles
DROP POLICY IF EXISTS microcycles_access ON microcycles;

-- Phases
DROP POLICY IF EXISTS phases_access ON phases;

-- Program versions
DROP POLICY IF EXISTS program_versions_access ON program_versions;

-- Programs
DROP POLICY IF EXISTS programs_client_access ON programs;
DROP POLICY IF EXISTS programs_coach_access ON programs;

-- Exercise library
DROP POLICY IF EXISTS exercise_library_owner_access ON exercise_library;
DROP POLICY IF EXISTS exercise_library_global_access ON exercise_library;

-- Events audit
DROP POLICY IF EXISTS events_audit_insert ON events_audit;
DROP POLICY IF EXISTS events_audit_self_access ON events_audit;

-- Client risk scores
DROP POLICY IF EXISTS client_risk_scores_coach_access ON client_risk_scores;

-- Coaching analytics
DROP POLICY IF EXISTS coaching_analytics_client_access ON coaching_analytics;
DROP POLICY IF EXISTS coaching_analytics_coach_access ON coaching_analytics;

-- Wearable summaries
DROP POLICY IF EXISTS wearable_summaries_coach_access ON wearable_summaries;
DROP POLICY IF EXISTS wearable_summaries_client_access ON wearable_summaries;

-- Insight cards
DROP POLICY IF EXISTS insight_cards_client_access ON insight_cards;
DROP POLICY IF EXISTS insight_cards_coach_access ON insight_cards;

-- Measurements
DROP POLICY IF EXISTS measurements_coach_access ON measurements;
DROP POLICY IF EXISTS measurements_client_access ON measurements;

-- Note: artifacts RLS policies drop moved to migration 000010_create_artifacts_table.down.sql

-- Form cues tracking
DROP POLICY IF EXISTS form_cues_tracking_access ON form_cues_tracking;

-- Workout exercises
DROP POLICY IF EXISTS workout_exercises_access ON workout_exercises;

-- Transcript segments
DROP POLICY IF EXISTS transcript_segments_access ON transcript_segments;

-- Sessions
DROP POLICY IF EXISTS sessions_client_access ON sessions;
DROP POLICY IF EXISTS sessions_coach_access ON sessions;

-- Coach client relationships
DROP POLICY IF EXISTS ccr_client_access ON coach_client_relationships;
DROP POLICY IF EXISTS ccr_coach_access ON coach_client_relationships;

-- Client profiles
DROP POLICY IF EXISTS client_profiles_coach_access ON client_profiles;
DROP POLICY IF EXISTS client_profiles_self_access ON client_profiles;

-- Coach profiles
DROP POLICY IF EXISTS coach_profiles_self_access ON coach_profiles;

-- Users
DROP POLICY IF EXISTS users_coach_access ON users;
DROP POLICY IF EXISTS users_self_access ON users;

-- Drop helper functions
DROP FUNCTION IF EXISTS is_coach_of(UUID);
DROP FUNCTION IF EXISTS current_app_user_id();
