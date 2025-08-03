-- Seed initial balance data
-- This migration adds initial balance records for existing users

-- Insert default balance for user with ID 1 (if exists)
INSERT INTO balances (user_id, available_balance, created_at, updated_at) 
SELECT 1, 10000.00, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
WHERE EXISTS (SELECT 1 FROM users WHERE id = 1)
ON CONFLICT (user_id) DO NOTHING;

-- You can add more initial balance records here as needed
-- Example:
-- INSERT INTO balances (user_id, available_balance) 
-- SELECT 2, 5000.00
-- WHERE EXISTS (SELECT 1 FROM users WHERE id = 2)
-- ON CONFLICT (user_id) DO NOTHING; 