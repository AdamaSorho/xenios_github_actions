-- Create artifacts table for file storage metadata
CREATE TABLE IF NOT EXISTS artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES users(id),
    coach_id UUID NOT NULL REFERENCES users(id),
    file_name TEXT NOT NULL,
    file_type TEXT NOT NULL,
    file_size BIGINT NOT NULL CHECK (file_size > 0),
    storage_key TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('document', 'audio', 'image', 'video')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'uploaded', 'failed')),
    content_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for querying artifacts by client
CREATE INDEX IF NOT EXISTS idx_artifacts_client_id ON artifacts(client_id);

-- Index for querying artifacts by coach
CREATE INDEX IF NOT EXISTS idx_artifacts_coach_id ON artifacts(coach_id);

-- Index for querying artifacts by status
CREATE INDEX IF NOT EXISTS idx_artifacts_status ON artifacts(status);

-- Index for looking up by storage key
CREATE UNIQUE INDEX IF NOT EXISTS idx_artifacts_storage_key ON artifacts(storage_key);

-- Trigger for auto-updating updated_at
DROP TRIGGER IF EXISTS update_artifacts_updated_at ON artifacts;
CREATE TRIGGER update_artifacts_updated_at
    BEFORE UPDATE ON artifacts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Enable Row Level Security
ALTER TABLE artifacts ENABLE ROW LEVEL SECURITY;

-- RLS policy: coaches can only access artifacts they own
DROP POLICY IF EXISTS artifacts_coach_access ON artifacts;
CREATE POLICY artifacts_coach_access ON artifacts
    FOR ALL
    USING (coach_id = current_setting('app.current_user_id', true)::UUID);
