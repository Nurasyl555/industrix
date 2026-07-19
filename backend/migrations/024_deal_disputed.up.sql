-- A deal with an open dispute must not keep moving: advancing it to completed
-- makes the payment consumer release escrow to the seller, and cancelling it
-- refunds the buyer — either way the money settles itself before the
-- arbitrator has decided. The flag lives on the deal so the deal module can
-- enforce the freeze without depending on the dispute module at runtime.
ALTER TABLE deals ADD COLUMN IF NOT EXISTS disputed BOOLEAN NOT NULL DEFAULT FALSE;
