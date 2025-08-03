-- Migration: Drop users table (ROLLBACK)
-- Module: User Management
-- Created: 2024-12-19
-- Description: Rollback the users table creation

-- Drop trigger first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_users_updated_at_column();

-- Drop indexes (they will be dropped automatically with the table, but explicit is better)
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_email;

-- Drop table (this will cascade to dependent tables due to foreign key constraints)
DROP TABLE IF EXISTS users; 