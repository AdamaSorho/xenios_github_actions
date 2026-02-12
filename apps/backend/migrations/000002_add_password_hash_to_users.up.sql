-- Add password_hash column for authentication
-- Stores bcrypt hashes (~60 chars); TEXT is safer than VARCHAR for future algorithm changes
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '';
