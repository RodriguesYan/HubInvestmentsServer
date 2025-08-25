DROP TABLE IF EXISTS orders;

CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    symbol VARCHAR(20) NOT NULL,
    order_type VARCHAR(20) NOT NULL CHECK (order_type IN ('MARKET', 'LIMIT', 'STOP_LOSS', 'STOP_LIMIT')),
    order_side VARCHAR(10) NOT NULL CHECK (order_side IN ('BUY', 'SELL')),
    quantity DECIMAL(18,8) NOT NULL CHECK (quantity > 0),
    price DECIMAL(18,8) CHECK (price > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PROCESSING', 'EXECUTED', 'FAILED', 'CANCELLED')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    executed_at TIMESTAMP,
    execution_price DECIMAL(18,8),
    market_price_at_submission DECIMAL(18,8),
    market_data_timestamp TIMESTAMP,
    failure_reason TEXT,
    retry_count INTEGER DEFAULT 0,
    processing_worker_id VARCHAR(50),
    external_order_id VARCHAR(100)
);

-- Indexes for performance optimization
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_symbol ON orders(symbol);
CREATE INDEX idx_orders_user_status ON orders(user_id, status);
CREATE INDEX idx_orders_symbol_status ON orders(symbol, status);

-- Trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_orders_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_orders_updated_at();

-- Sample test data
INSERT INTO orders (id, user_id, symbol, order_type, order_side, quantity, price, status, market_price_at_submission, market_data_timestamp) 
VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', 1, 'AAPL', 'LIMIT', 'BUY', 100.00000000, 150.50000000, 'PENDING', 150.25000000, CURRENT_TIMESTAMP),
    ('550e8400-e29b-41d4-a716-446655440002', 1, 'GOOGL', 'MARKET', 'SELL', 50.00000000, NULL, 'EXECUTED', 2750.00000000, CURRENT_TIMESTAMP - INTERVAL '1 hour'),
    ('550e8400-e29b-41d4-a716-446655440003', 1, 'MSFT', 'LIMIT', 'BUY', 75.00000000, 300.00000000, 'CANCELLED', 305.50000000, CURRENT_TIMESTAMP - INTERVAL '2 hours');

-- Update executed order with execution details
UPDATE orders 
SET executed_at = CURRENT_TIMESTAMP - INTERVAL '30 minutes',
    execution_price = 2748.75000000,
    updated_at = CURRENT_TIMESTAMP - INTERVAL '30 minutes'
WHERE id = '550e8400-e29b-41d4-a716-446655440002';

-- Query examples for testing
-- SELECT * FROM orders;
-- SELECT * FROM orders WHERE user_id = 1 ORDER BY created_at DESC;
-- SELECT * FROM orders WHERE status = 'PENDING';
-- SELECT symbol, COUNT(*) as order_count FROM orders GROUP BY symbol;
