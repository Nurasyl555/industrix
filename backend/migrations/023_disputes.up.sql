-- Phase 6: disputes. When a deal goes wrong while money sits in escrow, a
-- participant files a dispute with evidence and an admin arbitrates: refund the
-- buyer or release to the seller.
CREATE TABLE IF NOT EXISTS disputes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    filed_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    -- Image URLs from the media module (photos of the equipment, documents).
    evidence_urls TEXT[] NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'resolved_refund', 'resolved_release', 'rejected')),
    resolution_note TEXT,
    resolved_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- At most one dispute may be open per deal; resolved ones stay for history.
CREATE UNIQUE INDEX IF NOT EXISTS idx_disputes_one_open_per_deal
    ON disputes(deal_id) WHERE status = 'open';

CREATE INDEX IF NOT EXISTS idx_disputes_status ON disputes(status);
CREATE INDEX IF NOT EXISTS idx_disputes_filed_by ON disputes(filed_by);
