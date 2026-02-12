-- Exercise library (shared exercise definitions)
CREATE TABLE IF NOT EXISTS exercise_library (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL DEFAULT 'strength' CHECK (category IN ('strength', 'cardio', 'flexibility', 'balance', 'plyometric', 'other')),
    muscle_groups TEXT[] DEFAULT '{}',
    equipment TEXT[] DEFAULT '{}',
    difficulty TEXT NOT NULL DEFAULT 'intermediate' CHECK (difficulty IN ('beginner', 'intermediate', 'advanced')),
    video_url TEXT,
    instructions TEXT,
    created_by UUID REFERENCES users(id),
    is_global BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_exercise_library_category ON exercise_library(category);
CREATE INDEX IF NOT EXISTS idx_exercise_library_created_by ON exercise_library(created_by);
CREATE INDEX IF NOT EXISTS idx_exercise_library_is_global ON exercise_library(is_global);

ALTER TABLE exercise_library ENABLE ROW LEVEL SECURITY;

CREATE TRIGGER update_exercise_library_updated_at
    BEFORE UPDATE ON exercise_library
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Programs (training programs assigned to clients)
CREATE TABLE IF NOT EXISTS programs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'paused', 'completed', 'archived')),
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_programs_coach_id ON programs(coach_id);
CREATE INDEX IF NOT EXISTS idx_programs_client_id ON programs(client_id);
CREATE INDEX IF NOT EXISTS idx_programs_status ON programs(status);

ALTER TABLE programs ENABLE ROW LEVEL SECURITY;

CREATE TRIGGER update_programs_updated_at
    BEFORE UPDATE ON programs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Program versions (version history for programs)
CREATE TABLE IF NOT EXISTS program_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    program_id UUID NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    change_notes TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(program_id, version_number)
);

CREATE INDEX IF NOT EXISTS idx_program_versions_program_id ON program_versions(program_id);

ALTER TABLE program_versions ENABLE ROW LEVEL SECURITY;

-- Phases (macro-level training phases within a program)
CREATE TABLE IF NOT EXISTS phases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    program_id UUID NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    order_index INTEGER NOT NULL DEFAULT 0,
    duration_weeks INTEGER,
    focus TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_phases_program_id ON phases(program_id);

ALTER TABLE phases ENABLE ROW LEVEL SECURITY;

CREATE TRIGGER update_phases_updated_at
    BEFORE UPDATE ON phases
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Microcycles (weekly training blocks within a phase)
CREATE TABLE IF NOT EXISTS microcycles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phase_id UUID NOT NULL REFERENCES phases(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    week_number INTEGER NOT NULL,
    intensity_level TEXT DEFAULT 'moderate' CHECK (intensity_level IN ('deload', 'low', 'moderate', 'high', 'peak')),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_microcycles_phase_id ON microcycles(phase_id);

ALTER TABLE microcycles ENABLE ROW LEVEL SECURITY;

CREATE TRIGGER update_microcycles_updated_at
    BEFORE UPDATE ON microcycles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Programmed sessions (planned sessions within a microcycle)
CREATE TABLE IF NOT EXISTS programmed_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    microcycle_id UUID NOT NULL REFERENCES microcycles(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    day_of_week INTEGER CHECK (day_of_week BETWEEN 1 AND 7),
    order_index INTEGER NOT NULL DEFAULT 0,
    session_type TEXT NOT NULL DEFAULT 'training' CHECK (session_type IN ('training', 'recovery', 'assessment', 'flexibility')),
    estimated_duration_minutes INTEGER,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_programmed_sessions_microcycle_id ON programmed_sessions(microcycle_id);

ALTER TABLE programmed_sessions ENABLE ROW LEVEL SECURITY;

CREATE TRIGGER update_programmed_sessions_updated_at
    BEFORE UPDATE ON programmed_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Programmed exercises (exercises planned for a programmed session)
CREATE TABLE IF NOT EXISTS programmed_exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    programmed_session_id UUID NOT NULL REFERENCES programmed_sessions(id) ON DELETE CASCADE,
    exercise_id UUID REFERENCES exercise_library(id) ON DELETE SET NULL,
    exercise_name TEXT NOT NULL,
    order_index INTEGER NOT NULL DEFAULT 0,
    sets INTEGER,
    reps TEXT,
    weight_prescription TEXT,
    tempo TEXT,
    rest_seconds INTEGER,
    rpe_target NUMERIC(3,1),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_programmed_exercises_session_id ON programmed_exercises(programmed_session_id);
CREATE INDEX IF NOT EXISTS idx_programmed_exercises_exercise_id ON programmed_exercises(exercise_id);

ALTER TABLE programmed_exercises ENABLE ROW LEVEL SECURITY;

CREATE TRIGGER update_programmed_exercises_updated_at
    BEFORE UPDATE ON programmed_exercises
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Session completions (actual completed sessions linked to programmed ones)
CREATE TABLE IF NOT EXISTS session_completions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    programmed_session_id UUID REFERENCES programmed_sessions(id) ON DELETE SET NULL,
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    adherence_rating INTEGER CHECK (adherence_rating BETWEEN 1 AND 10),
    difficulty_rating INTEGER CHECK (difficulty_rating BETWEEN 1 AND 10),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_session_completions_client_id ON session_completions(client_id);
CREATE INDEX IF NOT EXISTS idx_session_completions_programmed_session_id ON session_completions(programmed_session_id);

ALTER TABLE session_completions ENABLE ROW LEVEL SECURITY;

-- Exercise logs (actual exercise performance data)
CREATE TABLE IF NOT EXISTS exercise_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_completion_id UUID NOT NULL REFERENCES session_completions(id) ON DELETE CASCADE,
    exercise_id UUID REFERENCES exercise_library(id) ON DELETE SET NULL,
    exercise_name TEXT NOT NULL,
    set_number INTEGER NOT NULL,
    reps_completed INTEGER,
    weight_kg NUMERIC(6,2),
    duration_seconds INTEGER,
    rpe_actual NUMERIC(3,1),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_exercise_logs_completion_id ON exercise_logs(session_completion_id);
CREATE INDEX IF NOT EXISTS idx_exercise_logs_exercise_id ON exercise_logs(exercise_id);

ALTER TABLE exercise_logs ENABLE ROW LEVEL SECURITY;

-- Behavior goals (client behavior change goals)
CREATE TABLE IF NOT EXISTS behavior_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL DEFAULT 'general' CHECK (category IN ('nutrition', 'sleep', 'stress', 'movement', 'hydration', 'general')),
    target_frequency TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused', 'completed', 'abandoned')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_behavior_goals_coach_id ON behavior_goals(coach_id);
CREATE INDEX IF NOT EXISTS idx_behavior_goals_client_id ON behavior_goals(client_id);
CREATE INDEX IF NOT EXISTS idx_behavior_goals_status ON behavior_goals(status);

ALTER TABLE behavior_goals ENABLE ROW LEVEL SECURITY;

CREATE TRIGGER update_behavior_goals_updated_at
    BEFORE UPDATE ON behavior_goals
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Behavior cues (contextual cues linked to behavior goals)
CREATE TABLE IF NOT EXISTS behavior_cues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    behavior_goal_id UUID NOT NULL REFERENCES behavior_goals(id) ON DELETE CASCADE,
    cue_text TEXT NOT NULL,
    cue_type TEXT NOT NULL DEFAULT 'reminder' CHECK (cue_type IN ('reminder', 'trigger', 'reward', 'environment')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_behavior_cues_goal_id ON behavior_cues(behavior_goal_id);

ALTER TABLE behavior_cues ENABLE ROW LEVEL SECURITY;

CREATE TRIGGER update_behavior_cues_updated_at
    BEFORE UPDATE ON behavior_cues
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Behavior check-ins (client self-reports on behavior goals)
CREATE TABLE IF NOT EXISTS behavior_checkins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    behavior_goal_id UUID NOT NULL REFERENCES behavior_goals(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    difficulty_rating INTEGER CHECK (difficulty_rating BETWEEN 1 AND 5),
    notes TEXT,
    checked_in_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_behavior_checkins_goal_id ON behavior_checkins(behavior_goal_id);
CREATE INDEX IF NOT EXISTS idx_behavior_checkins_client_id ON behavior_checkins(client_id);
CREATE INDEX IF NOT EXISTS idx_behavior_checkins_checked_in_at ON behavior_checkins(checked_in_at);

ALTER TABLE behavior_checkins ENABLE ROW LEVEL SECURITY;

-- Program adjustments (log of changes made to programs)
CREATE TABLE IF NOT EXISTS program_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    program_id UUID NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    adjusted_by UUID NOT NULL REFERENCES users(id),
    adjustment_type TEXT NOT NULL CHECK (adjustment_type IN ('volume', 'intensity', 'exercise_swap', 'schedule', 'deload', 'progression', 'other')),
    reason TEXT NOT NULL,
    details TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_program_adjustments_program_id ON program_adjustments(program_id);
CREATE INDEX IF NOT EXISTS idx_program_adjustments_adjusted_by ON program_adjustments(adjusted_by);

ALTER TABLE program_adjustments ENABLE ROW LEVEL SECURITY;
