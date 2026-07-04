-- Rental bookings. The EXCLUDE constraint makes overlapping CONFIRMED bookings
-- for the same listing impossible at the database level — this is race-free,
-- unlike a check-then-insert in application code. Needs btree_gist for the
-- equality part of the exclusion (listing_id WITH =).
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    renter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'cancelled')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT booking_dates_valid CHECK (start_date <= end_date),
    -- No two confirmed bookings of the same listing may overlap. The range is
    -- inclusive of both ends ('[]') since a rental occupies whole days.
    EXCLUDE USING gist (
        listing_id WITH =,
        daterange(start_date, end_date, '[]') WITH &&
    ) WHERE (status = 'confirmed')
);

CREATE INDEX IF NOT EXISTS idx_bookings_listing ON bookings(listing_id);
CREATE INDEX IF NOT EXISTS idx_bookings_renter ON bookings(renter_id);
