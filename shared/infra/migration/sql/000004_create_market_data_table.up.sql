-- Migration: Create market_data table
-- Module: Market Data
-- Dependencies: None (independent table)
-- Created: 2025-09-01
-- Description: Create the market_data table for storing financial instrument information

CREATE TABLE IF NOT EXISTS market_data (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    name VARCHAR(50) NOT NULL,
    category INTEGER NOT NULL,
    last_quote DECIMAL NOT NULL
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_market_data_symbol ON market_data(symbol);
CREATE INDEX IF NOT EXISTS idx_market_data_category ON market_data(category);

-- Insert initial market data for testing
INSERT INTO market_data (id, symbol, name, category, last_quote) 
VALUES 	(5, 'VBR', 'Vanguard small caps value', 2, 240.5),
		(2, 'AMZN', 'Amazon prime', 1, 140.5),
 		(3, 'DIS', 'Disneylandia', 1, 244.5),
 		(4, 'VOO', 'Vanguard SP 500', 2, 340.5)
ON CONFLICT (id) DO NOTHING;

-- Reset sequence to ensure proper auto-increment after manual inserts
SELECT setval('market_data_id_seq', (SELECT COALESCE(MAX(id), 1) FROM market_data));
