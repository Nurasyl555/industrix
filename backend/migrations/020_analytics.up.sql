-- Phase 6: analytics. An append-only event store fed by the Kafka consumer —
-- every domain event the dashboards care about lands here, so aggregates can be
-- computed without reaching into other modules' tables.
CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,        -- the Kafka topic name
    entity_id UUID,                         -- listing / deal / payment id
    seller_id UUID,                         -- attribution for seller dashboards
    amount NUMERIC(14, 2),                  -- money, for payment events (GMV)
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analytics_type ON analytics_events(event_type);
CREATE INDEX IF NOT EXISTS idx_analytics_seller ON analytics_events(seller_id);
CREATE INDEX IF NOT EXISTS idx_analytics_occurred ON analytics_events(occurred_at);
