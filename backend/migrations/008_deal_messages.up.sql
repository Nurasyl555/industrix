-- Two-way message thread inside a deal. The deal's own `message` column stays
-- as the original inquiry, but is ALSO copied here as the first row so the
-- thread reads as one continuous conversation.
CREATE TABLE IF NOT EXISTS deal_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deal_messages_deal ON deal_messages(deal_id, created_at);
