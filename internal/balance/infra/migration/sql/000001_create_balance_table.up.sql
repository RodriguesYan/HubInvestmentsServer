-- Create balances table
-- This migration creates the main balances table for user account balances

CREATE TABLE IF NOT EXISTS balances (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    available_balance DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT positive_balance CHECK (available_balance >= 0),
    CONSTRAINT unique_user_balance UNIQUE (user_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_balances_user_id ON balances(user_id);
CREATE INDEX IF NOT EXISTS idx_balances_updated_at ON balances(updated_at);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_balance_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_balance_updated_at 
    BEFORE UPDATE ON balances 
    FOR EACH ROW 
    EXECUTE FUNCTION update_balance_updated_at_column(); 