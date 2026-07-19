-- A subscription charge has no counterparty: the platform is paid, so there is
-- no payee row to point at. 021 made deal_id optional but left payee_id NOT
-- NULL, which rejected every subscription charge.
ALTER TABLE payments ALTER COLUMN payee_id DROP NOT NULL;

-- Escrow still requires both: a deal and the seller who gets paid on release.
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_escrow_needs_deal;
ALTER TABLE payments ADD CONSTRAINT payments_escrow_needs_deal
    CHECK (kind <> 'escrow' OR (deal_id IS NOT NULL AND payee_id IS NOT NULL));
