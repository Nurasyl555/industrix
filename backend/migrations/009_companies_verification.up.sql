-- The integrity repository referenced columns that never existed in
-- 003_companies.up.sql (reviewer_note, reputation_score), so company creation
-- failed at runtime. Add the missing columns. The code also used `status` and
-- `user_id`; those are reconciled in code to the real columns
-- verification_status and owner_id (no schema change needed for those).
ALTER TABLE companies ADD COLUMN IF NOT EXISTS reviewer_note TEXT;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS reputation_score DECIMAL(3, 2) DEFAULT 0.0;
