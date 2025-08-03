-- Drop balances table and related objects
-- This migration reverts the creation of the balances table

-- Drop trigger first
DROP TRIGGER IF EXISTS update_balance_updated_at ON balances;

-- Drop function
DROP FUNCTION IF EXISTS update_balance_updated_at_column();

-- Drop indexes (they will be dropped automatically with the table, but explicit is better)
DROP INDEX IF EXISTS idx_balances_updated_at;
DROP INDEX IF EXISTS idx_balances_user_id;

-- Drop table
DROP TABLE IF EXISTS balances; 