-- Migration: Drop balances table (ROLLBACK)
-- Module: Balance Management
-- Created: 2024-12-19
-- Description: Rollback the balances table creation

-- Drop trigger first
DROP TRIGGER IF EXISTS update_balance_updated_at ON balances;

-- Drop function
DROP FUNCTION IF EXISTS update_balance_updated_at_column();

-- Drop indexes (they will be dropped automatically with the table, but explicit is better)
DROP INDEX IF EXISTS idx_balances_updated_at;
DROP INDEX IF EXISTS idx_balances_user_id;

-- Drop table
DROP TABLE IF EXISTS balances; 