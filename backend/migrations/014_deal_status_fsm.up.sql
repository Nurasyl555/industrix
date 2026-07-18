-- Phase 3: expand the deal status from the MVP inquiry/closed pair into the
-- full state machine: inquiry -> negotiation -> confirmed -> in_progress ->
-- completed, with cancelled reachable from any live state. Legacy 'closed'
-- rows are migrated to 'cancelled'.

ALTER TABLE deals DROP CONSTRAINT IF EXISTS deals_status_check;

UPDATE deals SET status = 'cancelled' WHERE status = 'closed';

ALTER TABLE deals
    ADD CONSTRAINT deals_status_check
    CHECK (status IN ('inquiry', 'negotiation', 'confirmed', 'in_progress', 'completed', 'cancelled'));
