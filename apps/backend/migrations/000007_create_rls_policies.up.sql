-- Row Level Security policies for all tables
-- These policies ensure coaches can only access their own clients' data.
-- The current user's ID is set via: SET LOCAL app.current_user_id = '<uuid>';

-- Helper function to get current user ID from session variable
CREATE OR REPLACE FUNCTION current_app_user_id()
RETURNS UUID AS $$
BEGIN
    RETURN NULLIF(current_setting('app.current_user_id', TRUE), '')::UUID;
EXCEPTION
    WHEN OTHERS THEN RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- Helper function to check if current user is a coach of a given client
CREATE OR REPLACE FUNCTION is_coach_of(p_client_id UUID)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM coach_client_relationships
        WHERE coach_id = current_app_user_id()
          AND client_id = p_client_id
          AND status = 'active'
    );
END;
$$ LANGUAGE plpgsql STABLE;

-- ============================================================
-- USERS
-- ============================================================
-- Users can see their own record
DROP POLICY IF EXISTS users_self_access ON users;
CREATE POLICY users_self_access ON users
    FOR ALL
    USING (id = current_app_user_id());

-- Coaches can see their clients' user records
DROP POLICY IF EXISTS users_coach_access ON users;
CREATE POLICY users_coach_access ON users
    FOR SELECT
    USING (is_coach_of(id));

-- ============================================================
-- COACH PROFILES
-- ============================================================
DROP POLICY IF EXISTS coach_profiles_self_access ON coach_profiles;
CREATE POLICY coach_profiles_self_access ON coach_profiles
    FOR ALL
    USING (user_id = current_app_user_id());

-- ============================================================
-- CLIENT PROFILES
-- ============================================================
DROP POLICY IF EXISTS client_profiles_self_access ON client_profiles;
CREATE POLICY client_profiles_self_access ON client_profiles
    FOR ALL
    USING (user_id = current_app_user_id());

DROP POLICY IF EXISTS client_profiles_coach_access ON client_profiles;
CREATE POLICY client_profiles_coach_access ON client_profiles
    FOR SELECT
    USING (is_coach_of(user_id));

-- ============================================================
-- COACH CLIENT RELATIONSHIPS
-- ============================================================
DROP POLICY IF EXISTS ccr_coach_access ON coach_client_relationships;
CREATE POLICY ccr_coach_access ON coach_client_relationships
    FOR ALL
    USING (coach_id = current_app_user_id());

DROP POLICY IF EXISTS ccr_client_access ON coach_client_relationships;
CREATE POLICY ccr_client_access ON coach_client_relationships
    FOR SELECT
    USING (client_id = current_app_user_id());

-- ============================================================
-- SESSIONS
-- ============================================================
DROP POLICY IF EXISTS sessions_coach_access ON sessions;
CREATE POLICY sessions_coach_access ON sessions
    FOR ALL
    USING (coach_id = current_app_user_id());

DROP POLICY IF EXISTS sessions_client_access ON sessions;
CREATE POLICY sessions_client_access ON sessions
    FOR SELECT
    USING (client_id = current_app_user_id());

-- ============================================================
-- TRANSCRIPT SEGMENTS
-- ============================================================
DROP POLICY IF EXISTS transcript_segments_access ON transcript_segments;
CREATE POLICY transcript_segments_access ON transcript_segments
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM sessions s
            WHERE s.id = transcript_segments.session_id
              AND (s.coach_id = current_app_user_id() OR s.client_id = current_app_user_id())
        )
    );

-- ============================================================
-- WORKOUT EXERCISES
-- ============================================================
DROP POLICY IF EXISTS workout_exercises_access ON workout_exercises;
CREATE POLICY workout_exercises_access ON workout_exercises
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM sessions s
            WHERE s.id = workout_exercises.session_id
              AND (s.coach_id = current_app_user_id() OR s.client_id = current_app_user_id())
        )
    );

-- ============================================================
-- FORM CUES TRACKING
-- ============================================================
DROP POLICY IF EXISTS form_cues_tracking_access ON form_cues_tracking;
CREATE POLICY form_cues_tracking_access ON form_cues_tracking
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM workout_exercises we
            JOIN sessions s ON s.id = we.session_id
            WHERE we.id = form_cues_tracking.workout_exercise_id
              AND (s.coach_id = current_app_user_id() OR s.client_id = current_app_user_id())
        )
    );

-- Note: artifacts RLS policies moved to migration 000010_create_artifacts_table.up.sql

-- ============================================================
-- MEASUREMENTS
-- ============================================================
DROP POLICY IF EXISTS measurements_client_access ON measurements;
CREATE POLICY measurements_client_access ON measurements
    FOR ALL
    USING (client_id = current_app_user_id());

DROP POLICY IF EXISTS measurements_coach_access ON measurements;
CREATE POLICY measurements_coach_access ON measurements
    FOR ALL
    USING (is_coach_of(client_id));

-- ============================================================
-- INSIGHT CARDS
-- ============================================================
DROP POLICY IF EXISTS insight_cards_coach_access ON insight_cards;
CREATE POLICY insight_cards_coach_access ON insight_cards
    FOR ALL
    USING (coach_id = current_app_user_id());

DROP POLICY IF EXISTS insight_cards_client_access ON insight_cards;
CREATE POLICY insight_cards_client_access ON insight_cards
    FOR SELECT
    USING (client_id = current_app_user_id() AND status IN ('approved', 'shared'));

-- ============================================================
-- WEARABLE SUMMARIES
-- ============================================================
DROP POLICY IF EXISTS wearable_summaries_client_access ON wearable_summaries;
CREATE POLICY wearable_summaries_client_access ON wearable_summaries
    FOR ALL
    USING (client_id = current_app_user_id());

DROP POLICY IF EXISTS wearable_summaries_coach_access ON wearable_summaries;
CREATE POLICY wearable_summaries_coach_access ON wearable_summaries
    FOR SELECT
    USING (is_coach_of(client_id));

-- ============================================================
-- COACHING ANALYTICS
-- ============================================================
DROP POLICY IF EXISTS coaching_analytics_coach_access ON coaching_analytics;
CREATE POLICY coaching_analytics_coach_access ON coaching_analytics
    FOR ALL
    USING (coach_id = current_app_user_id());

DROP POLICY IF EXISTS coaching_analytics_client_access ON coaching_analytics;
CREATE POLICY coaching_analytics_client_access ON coaching_analytics
    FOR SELECT
    USING (client_id = current_app_user_id());

-- ============================================================
-- CLIENT RISK SCORES
-- ============================================================
DROP POLICY IF EXISTS client_risk_scores_coach_access ON client_risk_scores;
CREATE POLICY client_risk_scores_coach_access ON client_risk_scores
    FOR ALL
    USING (coach_id = current_app_user_id());

-- ============================================================
-- EVENTS AUDIT
-- ============================================================
-- Actors can see their own audit events
DROP POLICY IF EXISTS events_audit_self_access ON events_audit;
CREATE POLICY events_audit_self_access ON events_audit
    FOR SELECT
    USING (actor_id = current_app_user_id());

-- Insert is allowed for any authenticated user
DROP POLICY IF EXISTS events_audit_insert ON events_audit;
CREATE POLICY events_audit_insert ON events_audit
    FOR INSERT
    WITH CHECK (actor_id = current_app_user_id());

-- ============================================================
-- EXERCISE LIBRARY
-- ============================================================
-- Global exercises visible to all authenticated users
DROP POLICY IF EXISTS exercise_library_global_access ON exercise_library;
CREATE POLICY exercise_library_global_access ON exercise_library
    FOR SELECT
    USING (is_global = TRUE);

-- Users can manage their own exercises
DROP POLICY IF EXISTS exercise_library_owner_access ON exercise_library;
CREATE POLICY exercise_library_owner_access ON exercise_library
    FOR ALL
    USING (created_by = current_app_user_id());

-- ============================================================
-- PROGRAMS
-- ============================================================
DROP POLICY IF EXISTS programs_coach_access ON programs;
CREATE POLICY programs_coach_access ON programs
    FOR ALL
    USING (coach_id = current_app_user_id());

DROP POLICY IF EXISTS programs_client_access ON programs;
CREATE POLICY programs_client_access ON programs
    FOR SELECT
    USING (client_id = current_app_user_id());

-- ============================================================
-- PROGRAM VERSIONS
-- ============================================================
DROP POLICY IF EXISTS program_versions_access ON program_versions;
CREATE POLICY program_versions_access ON program_versions
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM programs p
            WHERE p.id = program_versions.program_id
              AND (p.coach_id = current_app_user_id() OR p.client_id = current_app_user_id())
        )
    );

-- ============================================================
-- PHASES
-- ============================================================
DROP POLICY IF EXISTS phases_access ON phases;
CREATE POLICY phases_access ON phases
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM programs p
            WHERE p.id = phases.program_id
              AND (p.coach_id = current_app_user_id() OR p.client_id = current_app_user_id())
        )
    );

-- ============================================================
-- MICROCYCLES
-- ============================================================
DROP POLICY IF EXISTS microcycles_access ON microcycles;
CREATE POLICY microcycles_access ON microcycles
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM phases ph
            JOIN programs p ON p.id = ph.program_id
            WHERE ph.id = microcycles.phase_id
              AND (p.coach_id = current_app_user_id() OR p.client_id = current_app_user_id())
        )
    );

-- ============================================================
-- PROGRAMMED SESSIONS
-- ============================================================
DROP POLICY IF EXISTS programmed_sessions_access ON programmed_sessions;
CREATE POLICY programmed_sessions_access ON programmed_sessions
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM microcycles mc
            JOIN phases ph ON ph.id = mc.phase_id
            JOIN programs p ON p.id = ph.program_id
            WHERE mc.id = programmed_sessions.microcycle_id
              AND (p.coach_id = current_app_user_id() OR p.client_id = current_app_user_id())
        )
    );

-- ============================================================
-- PROGRAMMED EXERCISES
-- ============================================================
DROP POLICY IF EXISTS programmed_exercises_access ON programmed_exercises;
CREATE POLICY programmed_exercises_access ON programmed_exercises
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM programmed_sessions ps
            JOIN microcycles mc ON mc.id = ps.microcycle_id
            JOIN phases ph ON ph.id = mc.phase_id
            JOIN programs p ON p.id = ph.program_id
            WHERE ps.id = programmed_exercises.programmed_session_id
              AND (p.coach_id = current_app_user_id() OR p.client_id = current_app_user_id())
        )
    );

-- ============================================================
-- SESSION COMPLETIONS
-- ============================================================
DROP POLICY IF EXISTS session_completions_client_access ON session_completions;
CREATE POLICY session_completions_client_access ON session_completions
    FOR ALL
    USING (client_id = current_app_user_id());

DROP POLICY IF EXISTS session_completions_coach_access ON session_completions;
CREATE POLICY session_completions_coach_access ON session_completions
    FOR SELECT
    USING (is_coach_of(client_id));

-- ============================================================
-- EXERCISE LOGS
-- ============================================================
DROP POLICY IF EXISTS exercise_logs_access ON exercise_logs;
CREATE POLICY exercise_logs_access ON exercise_logs
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM session_completions sc
            WHERE sc.id = exercise_logs.session_completion_id
              AND (sc.client_id = current_app_user_id() OR is_coach_of(sc.client_id))
        )
    );

-- ============================================================
-- BEHAVIOR GOALS
-- ============================================================
DROP POLICY IF EXISTS behavior_goals_coach_access ON behavior_goals;
CREATE POLICY behavior_goals_coach_access ON behavior_goals
    FOR ALL
    USING (coach_id = current_app_user_id());

DROP POLICY IF EXISTS behavior_goals_client_access ON behavior_goals;
CREATE POLICY behavior_goals_client_access ON behavior_goals
    FOR SELECT
    USING (client_id = current_app_user_id());

-- ============================================================
-- BEHAVIOR CUES
-- ============================================================
DROP POLICY IF EXISTS behavior_cues_access ON behavior_cues;
CREATE POLICY behavior_cues_access ON behavior_cues
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM behavior_goals bg
            WHERE bg.id = behavior_cues.behavior_goal_id
              AND (bg.coach_id = current_app_user_id() OR bg.client_id = current_app_user_id())
        )
    );

-- ============================================================
-- BEHAVIOR CHECK-INS
-- ============================================================
DROP POLICY IF EXISTS behavior_checkins_client_access ON behavior_checkins;
CREATE POLICY behavior_checkins_client_access ON behavior_checkins
    FOR ALL
    USING (client_id = current_app_user_id());

DROP POLICY IF EXISTS behavior_checkins_coach_access ON behavior_checkins;
CREATE POLICY behavior_checkins_coach_access ON behavior_checkins
    FOR SELECT
    USING (is_coach_of(client_id));

-- ============================================================
-- PROGRAM ADJUSTMENTS
-- ============================================================
DROP POLICY IF EXISTS program_adjustments_access ON program_adjustments;
CREATE POLICY program_adjustments_access ON program_adjustments
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM programs p
            WHERE p.id = program_adjustments.program_id
              AND (p.coach_id = current_app_user_id() OR p.client_id = current_app_user_id())
        )
    );
