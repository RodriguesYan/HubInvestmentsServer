-- Migration: Seed initial data
-- Module: Data Seeding
-- Dependencies: 000001_create_users_table, 000002_create_balances_table
-- Created: 2024-12-19
-- Description: Insert initial data for development and testing

-- Insert initial user (if not exists)
INSERT INTO users (id, email, name, password, created_at, updated_at) 
VALUES (1, 'bla@bla.com', 'John Doe', '12345678', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- Insert initial balance for the user (if not exists)
INSERT INTO balances (user_id, available_balance, created_at, updated_at) 
VALUES (1, 10000.00, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (user_id) DO NOTHING;

-- Reset sequence to ensure proper auto-increment after manual inserts
SELECT setval('users_id_seq', (SELECT COALESCE(MAX(id), 1) FROM users));
SELECT setval('balances_id_seq', (SELECT COALESCE(MAX(id), 1) FROM balances)); 