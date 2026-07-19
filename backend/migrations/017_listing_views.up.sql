-- Phase 2: per-listing view counter (listing stats). Incremented on each
-- public detail view of an active listing.
ALTER TABLE listings ADD COLUMN IF NOT EXISTS view_count INTEGER NOT NULL DEFAULT 0;
