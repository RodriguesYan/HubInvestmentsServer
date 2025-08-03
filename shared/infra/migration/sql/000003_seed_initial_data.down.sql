-- Migration: Remove initial data (ROLLBACK)
-- Module: Data Seeding
-- Created: 2024-12-19
-- Description: Remove the initial data that was seeded

-- Remove the initial balance (must be done before removing user due to foreign key)
DELETE FROM balances WHERE user_id = 1 AND available_balance = 10000.00;

-- Remove the initial user
DELETE FROM users WHERE id = 1 AND email = 'bla@bla.com'; 