-- Migration: Replace legacy positions table with new Position domain model
-- Module: Position Management V2 (Domain-Driven Design)
-- Dependencies: 000001_create_users_table
-- Created: 2024-12-19
-- Description: Drop legacy positions table and create the new Position domain model table
-- Schema: yanrodrigues.positions_v2

-- Drop legacy positions table if it exists
DROP TABLE IF EXISTS positions CASCADE;

-- Create yanrodrigues schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS yanrodrigues;

-- Create UUID extension if not exists (required for UUID generation)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create positions_v2 table in yanrodrigues schema
CREATE TABLE IF NOT EXISTS yanrodrigues.positions_v2 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    quantity DECIMAL(20,8) NOT NULL,
    average_price DECIMAL(20,8) NOT NULL,
    total_investment DECIMAL(20,8) NOT NULL,
    current_price DECIMAL(20,8) DEFAULT 0,
    market_value DECIMAL(20,8) DEFAULT 0,
    unrealized_pnl DECIMAL(20,8) DEFAULT 0,
    unrealized_pnl_pct DECIMAL(10,4) DEFAULT 0,
    position_type VARCHAR(10) NOT NULL CHECK (position_type IN ('LONG', 'SHORT')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('ACTIVE', 'PARTIAL', 'CLOSED')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_trade_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT positive_quantity CHECK (quantity >= 0),
    CONSTRAINT positive_average_price CHECK (average_price >= 0),
    CONSTRAINT positive_total_investment CHECK (total_investment >= 0),
    CONSTRAINT non_empty_symbol CHECK (LENGTH(TRIM(symbol)) > 0),
    CONSTRAINT valid_position_type CHECK (position_type IN ('LONG', 'SHORT')),
    CONSTRAINT valid_status CHECK (status IN ('ACTIVE', 'PARTIAL', 'CLOSED')),
    CONSTRAINT closed_position_zero_quantity CHECK (
        (status = 'CLOSED' AND quantity = 0) OR 
        (status != 'CLOSED')
    ),
    CONSTRAINT active_position_positive_quantity CHECK (
        (status = 'ACTIVE' AND quantity > 0) OR 
        (status != 'ACTIVE')
    ),
    
    -- Prevent duplicate positions per user/symbol
    CONSTRAINT unique_user_symbol UNIQUE (user_id, symbol)
);

-- Create indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_positions_v2_user_id ON yanrodrigues.positions_v2(user_id);
CREATE INDEX IF NOT EXISTS idx_positions_v2_symbol ON yanrodrigues.positions_v2(symbol);
CREATE INDEX IF NOT EXISTS idx_positions_v2_status ON yanrodrigues.positions_v2(status);
CREATE INDEX IF NOT EXISTS idx_positions_v2_user_symbol ON yanrodrigues.positions_v2(user_id, symbol);
CREATE INDEX IF NOT EXISTS idx_positions_v2_created_at ON yanrodrigues.positions_v2(created_at);
CREATE INDEX IF NOT EXISTS idx_positions_v2_updated_at ON yanrodrigues.positions_v2(updated_at);
CREATE INDEX IF NOT EXISTS idx_positions_v2_last_trade_at ON yanrodrigues.positions_v2(last_trade_at);

-- Create composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_positions_v2_user_status ON yanrodrigues.positions_v2(user_id, status);
CREATE INDEX IF NOT EXISTS idx_positions_v2_symbol_status ON yanrodrigues.positions_v2(symbol, status);

-- Create trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION yanrodrigues.update_positions_v2_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at timestamp
CREATE TRIGGER update_positions_v2_updated_at 
    BEFORE UPDATE ON yanrodrigues.positions_v2
    FOR EACH ROW 
    EXECUTE FUNCTION yanrodrigues.update_positions_v2_updated_at_column();

-- Create trigger function to maintain data consistency
CREATE OR REPLACE FUNCTION yanrodrigues.validate_positions_v2_consistency()
RETURNS TRIGGER AS $$
BEGIN
    -- Ensure total_investment = quantity * average_price (with precision tolerance)
    IF ABS(NEW.total_investment - (NEW.quantity * NEW.average_price)) > 0.001 THEN
        RAISE EXCEPTION 'Total investment must equal quantity × average price. Got: % vs %', 
            NEW.total_investment, (NEW.quantity * NEW.average_price);
    END IF;
    
    -- Ensure market_value = quantity * current_price (when current_price > 0)
    IF NEW.current_price > 0 AND ABS(NEW.market_value - (NEW.quantity * NEW.current_price)) > 0.001 THEN
        RAISE EXCEPTION 'Market value must equal quantity × current price. Got: % vs %', 
            NEW.market_value, (NEW.quantity * NEW.current_price);
    END IF;
    
    -- Ensure unrealized_pnl = market_value - total_investment (when current_price > 0)
    IF NEW.current_price > 0 AND ABS(NEW.unrealized_pnl - (NEW.market_value - NEW.total_investment)) > 0.001 THEN
        RAISE EXCEPTION 'Unrealized P&L must equal market value - total investment. Got: % vs %', 
            NEW.unrealized_pnl, (NEW.market_value - NEW.total_investment);
    END IF;
    
    -- Ensure unrealized_pnl_pct calculation is correct (when total_investment > 0)
    IF NEW.total_investment > 0 AND NEW.current_price > 0 THEN
        DECLARE
            expected_pnl_pct DECIMAL(10,4);
        BEGIN
            expected_pnl_pct := (NEW.unrealized_pnl / NEW.total_investment) * 100;
            IF ABS(NEW.unrealized_pnl_pct - expected_pnl_pct) > 0.01 THEN
                RAISE EXCEPTION 'Unrealized P&L percentage calculation incorrect. Got: % vs %', 
                    NEW.unrealized_pnl_pct, expected_pnl_pct;
            END IF;
        END;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for data consistency validation
CREATE TRIGGER validate_positions_v2_consistency_trigger
    BEFORE INSERT OR UPDATE ON yanrodrigues.positions_v2
    FOR EACH ROW 
    EXECUTE FUNCTION yanrodrigues.validate_positions_v2_consistency();

-- Grant permissions (adjust as needed for your security requirements)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON yanrodrigues.positions_v2 TO your_application_user;
-- GRANT USAGE ON SCHEMA yanrodrigues TO your_application_user;

-- Add helpful comments
COMMENT ON TABLE yanrodrigues.positions_v2 IS 'Enhanced positions table for Position domain model with comprehensive business logic validation';
COMMENT ON COLUMN yanrodrigues.positions_v2.id IS 'UUID primary key for position';
COMMENT ON COLUMN yanrodrigues.positions_v2.user_id IS 'UUID reference to user who owns this position';
COMMENT ON COLUMN yanrodrigues.positions_v2.symbol IS 'Stock/Asset symbol (e.g., AAPL, GOOGL)';
COMMENT ON COLUMN yanrodrigues.positions_v2.quantity IS 'Number of shares/units held (supports up to 8 decimal places)';
COMMENT ON COLUMN yanrodrigues.positions_v2.average_price IS 'Weighted average price per unit (supports up to 8 decimal places)';
COMMENT ON COLUMN yanrodrigues.positions_v2.total_investment IS 'Total amount invested (quantity × average_price)';
COMMENT ON COLUMN yanrodrigues.positions_v2.current_price IS 'Current market price per unit';
COMMENT ON COLUMN yanrodrigues.positions_v2.market_value IS 'Current market value (quantity × current_price)';
COMMENT ON COLUMN yanrodrigues.positions_v2.unrealized_pnl IS 'Unrealized profit/loss (market_value - total_investment)';
COMMENT ON COLUMN yanrodrigues.positions_v2.unrealized_pnl_pct IS 'Unrealized profit/loss percentage';
COMMENT ON COLUMN yanrodrigues.positions_v2.position_type IS 'Position type: LONG or SHORT';
COMMENT ON COLUMN yanrodrigues.positions_v2.status IS 'Position status: ACTIVE, PARTIAL, or CLOSED';
COMMENT ON COLUMN yanrodrigues.positions_v2.last_trade_at IS 'Timestamp of last trade affecting this position';
