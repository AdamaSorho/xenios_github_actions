-- Remove password_hash column (preserves all other user data)
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
