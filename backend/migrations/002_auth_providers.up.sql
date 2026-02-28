ALTER TABLE users 
    ALTER COLUMN email DROP NOT NULL, 
    ALTER COLUMN phone DROP NOT NULL, 
    ALTER COLUMN password_hash DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS google_id VARCHAR(255) UNIQUE,
    DROP CONSTRAINT IF EXISTS require_one_auth_method;

ALTER TABLE users     
    ADD CONSTRAINT require_one_auth_method CHECK (email IS NOT NULL OR phone IS NOT NULL OR google_id IS NOT NULL);
