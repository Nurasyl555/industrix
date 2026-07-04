-- Equipment photo. A single primary image URL for the MVP (points at a public
-- object in the equipment-media MinIO bucket). A full gallery would be a
-- separate equipment_images table — deferred.
ALTER TABLE equipment ADD COLUMN IF NOT EXISTS image_url TEXT;
