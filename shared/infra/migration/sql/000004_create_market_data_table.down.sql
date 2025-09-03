-- Migration: Drop market_data table (ROLLBACK)
-- Module: Market Data
-- Description: Remove market_data table and related indexes

DROP INDEX IF EXISTS idx_market_data_category;
DROP INDEX IF EXISTS idx_market_data_symbol;
DROP TABLE IF EXISTS market_data;
