-- Phase 6: subscription billing. Payments so far were always deal escrow;
-- a subscription charge has no deal behind it, so deal_id becomes optional and
-- a kind column distinguishes the two flows (with a check that escrow still
-- always carries its deal).
ALTER TABLE payments ALTER COLUMN deal_id DROP NOT NULL;

ALTER TABLE payments ADD COLUMN IF NOT EXISTS kind VARCHAR(20) NOT NULL DEFAULT 'escrow'
    CHECK (kind IN ('escrow', 'subscription'));

ALTER TABLE payments ADD COLUMN IF NOT EXISTS description TEXT;

ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_escrow_needs_deal;
ALTER TABLE payments ADD CONSTRAINT payments_escrow_needs_deal
    CHECK (kind <> 'escrow' OR deal_id IS NOT NULL);

-- Link a subscription to the payment that last renewed it, for support/audit.
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS last_payment_id UUID;
