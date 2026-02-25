-- Migration: 002_companies.sql
-- Description: Creates companies, company_documents, and verification_history tables
-- Created: 2024

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Companies table
CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    bin VARCHAR(12) NOT NULL,
    address TEXT,
    phone VARCHAR(20),
    email VARCHAR(255),
    website VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    verified BOOLEAN NOT NULL DEFAULT false,
    reviewer_note TEXT,
    reputation_score DECIMAL(3,2) DEFAULT 0.00,
    subscription_plan VARCHAR(50) DEFAULT 'FREE',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for companies
CREATE INDEX IF NOT EXISTS idx_companies_user_id ON companies(user_id);
CREATE INDEX IF NOT EXISTS idx_companies_bin ON companies(bin);
CREATE INDEX IF NOT EXISTS idx_companies_status ON companies(status);
CREATE INDEX IF NOT EXISTS idx_companies_verified ON companies(verified);

-- Company documents table (for verification documents)
CREATE TABLE IF NOT EXISTS company_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE,
    verified BOOLEAN NOT NULL DEFAULT false,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for company documents
CREATE INDEX IF NOT EXISTS idx_company_documents_company_id ON company_documents(company_id);
CREATE INDEX IF NOT EXISTS idx_company_documents_type ON company_documents(document_type);

-- Verification history table (audit trail for verification decisions)
CREATE TABLE IF NOT EXISTS verification_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    previous_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    reviewer_id UUID REFERENCES users(id),
    reviewer_note TEXT,
    document_ids UUID[],
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for verification history
CREATE INDEX IF NOT EXISTS idx_verification_history_company_id ON verification_history(company_id);
CREATE INDEX IF NOT EXISTS idx_verification_history_created_at ON verification_history(created_at);

-- Function to update updated_at timestamp for companies
CREATE OR REPLACE FUNCTION update_companies_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for companies
CREATE TRIGGER update_companies_updated_at
    BEFORE UPDATE ON companies
    FOR EACH ROW
    EXECUTE FUNCTION update_companies_updated_at();

-- Add foreign key from users to companies
ALTER TABLE users 
    ADD CONSTRAINT fk_users_company_id 
    FOREIGN KEY (company_id) 
    REFERENCES companies(id) 
    ON DELETE SET NULL;

COMMENT ON TABLE companies IS 'Company profiles with verification status';
COMMENT ON TABLE company_documents IS 'Verification documents (BIN cert, charter, director ID)';
COMMENT ON TABLE verification_history IS 'Audit trail for company verification decisions';
