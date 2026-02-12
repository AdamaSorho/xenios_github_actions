-- Coach profiles
CREATE TABLE IF NOT EXISTS coach_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    specializations TEXT[] DEFAULT '{}',
    bio TEXT,
    certifications TEXT[] DEFAULT '{}',
    hourly_rate_cents INTEGER,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    max_clients INTEGER DEFAULT 50,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_coach_profiles_user_id ON coach_profiles(user_id);

ALTER TABLE coach_profiles ENABLE ROW LEVEL SECURITY;

DROP TRIGGER IF EXISTS update_coach_profiles_updated_at ON coach_profiles;
CREATE TRIGGER update_coach_profiles_updated_at
    BEFORE UPDATE ON coach_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Client profiles
CREATE TABLE IF NOT EXISTS client_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    date_of_birth DATE,
    gender TEXT,
    height_cm NUMERIC(5,1),
    weight_kg NUMERIC(5,1),
    goals TEXT[] DEFAULT '{}',
    medical_notes TEXT,
    emergency_contact_name TEXT,
    emergency_contact_phone TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_client_profiles_user_id ON client_profiles(user_id);

ALTER TABLE client_profiles ENABLE ROW LEVEL SECURITY;

DROP TRIGGER IF EXISTS update_client_profiles_updated_at ON client_profiles;
CREATE TRIGGER update_client_profiles_updated_at
    BEFORE UPDATE ON client_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Coach-client relationships
CREATE TABLE IF NOT EXISTS coach_client_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('pending', 'active', 'paused', 'terminated')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(coach_id, client_id)
);

CREATE INDEX IF NOT EXISTS idx_ccr_coach_id ON coach_client_relationships(coach_id);
CREATE INDEX IF NOT EXISTS idx_ccr_client_id ON coach_client_relationships(client_id);
CREATE INDEX IF NOT EXISTS idx_ccr_status ON coach_client_relationships(status);

ALTER TABLE coach_client_relationships ENABLE ROW LEVEL SECURITY;

DROP TRIGGER IF EXISTS update_ccr_updated_at ON coach_client_relationships;
CREATE TRIGGER update_ccr_updated_at
    BEFORE UPDATE ON coach_client_relationships
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
