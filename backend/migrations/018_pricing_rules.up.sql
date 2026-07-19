-- Phase 2: pricing rules. Listings gain a pricing model — 'fixed' (price is
-- final) or 'negotiable' (price is a starting point buyers can negotiate via a
-- deal). Bookings record the total price computed from the rental period rate.
ALTER TABLE listings ADD COLUMN IF NOT EXISTS pricing_type VARCHAR(20) NOT NULL DEFAULT 'fixed'
    CHECK (pricing_type IN ('fixed', 'negotiable'));

ALTER TABLE bookings ADD COLUMN IF NOT EXISTS total_price NUMERIC(14, 2) NOT NULL DEFAULT 0;
