-- Seed a built-in admin account so a fresh database always has someone who can
-- moderate. Password is "Admin12345" (bcrypt hash below). Change it in a real
-- deployment. ON CONFLICT keeps it idempotent and promotes an existing row.
INSERT INTO users (email, password_hash, first_name, role, verified)
VALUES ('admin@industrix.local', '$2a$10$GXRszAkjXmrdJEe3rR9A7u1G8/.bXkV2HqfM3i6DXhCx0r76ow1.m', 'Admin', 'admin', true)
ON CONFLICT (email) DO UPDATE SET role = 'admin', verified = true;

-- Listings now pass through moderation before going live. Widen the status
-- check to allow the new states (was: draft, active, archived).
ALTER TABLE listings DROP CONSTRAINT IF EXISTS listings_status_check;
ALTER TABLE listings ADD CONSTRAINT listings_status_check
    CHECK (status IN ('draft', 'moderation', 'active', 'archived', 'rejected'));
