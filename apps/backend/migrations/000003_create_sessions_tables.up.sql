-- Sessions table with soft delete support
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_type TEXT NOT NULL DEFAULT 'in_person' CHECK (session_type IN ('in_person', 'virtual', 'async_review')),
    status TEXT NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'in_progress', 'completed', 'cancelled')),
    scheduled_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    duration_minutes INTEGER,
    summary JSONB,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sessions_coach_id ON sessions(coach_id);
CREATE INDEX IF NOT EXISTS idx_sessions_client_id ON sessions(client_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_scheduled_at ON sessions(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_sessions_deleted_at ON sessions(deleted_at);

ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;

DROP TRIGGER IF EXISTS update_sessions_updated_at ON sessions;
CREATE TRIGGER update_sessions_updated_at
    BEFORE UPDATE ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Transcript segments (from session audio)
CREATE TABLE IF NOT EXISTS transcript_segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    speaker TEXT NOT NULL CHECK (speaker IN ('coach', 'client')),
    content TEXT NOT NULL,
    start_time_ms INTEGER NOT NULL,
    end_time_ms INTEGER NOT NULL,
    confidence NUMERIC(4,3),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transcript_segments_session_id ON transcript_segments(session_id);

ALTER TABLE transcript_segments ENABLE ROW LEVEL SECURITY;

-- Workout exercises performed during a session
CREATE TABLE IF NOT EXISTS workout_exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    exercise_name TEXT NOT NULL,
    sets INTEGER,
    reps INTEGER,
    weight_kg NUMERIC(6,2),
    duration_seconds INTEGER,
    rest_seconds INTEGER,
    order_index INTEGER NOT NULL DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workout_exercises_session_id ON workout_exercises(session_id);

ALTER TABLE workout_exercises ENABLE ROW LEVEL SECURITY;

DROP TRIGGER IF EXISTS update_workout_exercises_updated_at ON workout_exercises;
CREATE TRIGGER update_workout_exercises_updated_at
    BEFORE UPDATE ON workout_exercises
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Form cues tracking (coaching feedback on exercise form)
CREATE TABLE IF NOT EXISTS form_cues_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_exercise_id UUID NOT NULL REFERENCES workout_exercises(id) ON DELETE CASCADE,
    cue TEXT NOT NULL,
    cue_type TEXT NOT NULL DEFAULT 'corrective' CHECK (cue_type IN ('corrective', 'positive', 'safety')),
    timestamp_ms INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_form_cues_workout_exercise_id ON form_cues_tracking(workout_exercise_id);

ALTER TABLE form_cues_tracking ENABLE ROW LEVEL SECURITY;
