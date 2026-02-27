CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL REFERENCES users(id),
    target_entity_id UUID NOT NULL, -- company_id or equipment_id (can be verified against other tables if needed)
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    transaction_id UUID, -- Optional link to a transaction
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reputation_scores (
    entity_id UUID PRIMARY KEY, -- company_id or user_id
    average_rating DECIMAL(3, 2) DEFAULT 0.0,
    review_count INT DEFAULT 0,
    tier VARCHAR(20) DEFAULT 'none', -- gold, silver, bronze, none
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_reviews_target_entity_id ON reviews(target_entity_id);
