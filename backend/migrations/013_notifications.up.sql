-- In-app notification feed. One row per event delivered to a user. A full
-- multi-channel system (push/email via Kafka fanout) is the architecture.md
-- Notification module — this is the in-app slice.
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(40) NOT NULL,   -- inquiry, message, booking, listing_approved, ...
    message TEXT NOT NULL,
    link TEXT,                   -- optional in-app path to open
    read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, read, created_at DESC);
