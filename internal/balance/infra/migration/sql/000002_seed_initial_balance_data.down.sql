-- Remove seeded balance data
-- This migration removes the initial balance data that was inserted

-- Remove the default balance for user with ID 1
DELETE FROM balances WHERE user_id = 1 AND available_balance = 10000.00;

-- Add more deletions here if you added more initial records in the up migration
-- Example:
-- DELETE FROM balances WHERE user_id = 2 AND available_balance = 5000.00; 