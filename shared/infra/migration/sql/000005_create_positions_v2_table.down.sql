-- Migration Rollback: Drop positions_v2 and recreate legacy positions table
-- Module: Position Management V2 (Domain-Driven Design)
-- Created: 2024-12-19
-- Description: Rollback to legacy positions table structure
-- Schema: yanrodrigues.positions_v2 → positions

-- Drop triggers first (to avoid dependency issues)
DROP TRIGGER IF EXISTS validate_positions_v2_consistency_trigger ON yanrodrigues.positions_v2;
DROP TRIGGER IF EXISTS update_positions_v2_updated_at ON yanrodrigues.positions_v2;

-- Drop trigger functions
DROP FUNCTION IF EXISTS yanrodrigues.validate_positions_v2_consistency();
DROP FUNCTION IF EXISTS yanrodrigues.update_positions_v2_updated_at_column();

-- Drop indexes (they will be automatically dropped with the table, but explicit for clarity)
DROP INDEX IF EXISTS yanrodrigues.idx_positions_v2_symbol_status;
DROP INDEX IF EXISTS yanrodrigues.idx_positions_v2_user_status;
DROP INDEX IF EXISTS yanrodrigues.idx_positions_v2_last_trade_at;
DROP INDEX IF EXISTS yanrodrigues.idx_positions_v2_updated_at;
DROP INDEX IF EXISTS yanrodrigues.idx_positions_v2_created_at;
DROP INDEX IF EXISTS yanrodrigues.idx_positions_v2_user_symbol;
DROP INDEX IF EXISTS yanrodrigues.idx_positions_v2_status;
DROP INDEX IF EXISTS yanrodrigues.idx_positions_v2_symbol;
DROP INDEX IF EXISTS yanrodrigues.idx_positions_v2_user_id;

-- Drop the table
DROP TABLE IF EXISTS yanrodrigues.positions_v2;

-- Optionally drop the schema if it's empty (uncomment if you want full cleanup)
-- Note: This will fail if other objects exist in the schema, which is intentional
-- DROP SCHEMA IF EXISTS yanrodrigues RESTRICT;

-- Recreate basic legacy positions table for rollback compatibility
CREATE TABLE IF NOT EXISTS positions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    instrument_id INTEGER NOT NULL,
    quantity DECIMAL NOT NULL,
    average_price DECIMAL NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create basic indexes for performance
CREATE INDEX IF NOT EXISTS idx_positions_user_id ON positions(user_id);
CREATE INDEX IF NOT EXISTS idx_positions_instrument_id ON positions(instrument_id);

-- Note: We don't drop the UUID extension as it might be used by other tables
