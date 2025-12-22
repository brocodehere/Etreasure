-- Fix orders table to add missing user_id column and align with backend expectations
-- This migration addresses the schema mismatch between initial orders table and backend code

-- Add user_id column to orders table (references users table, not customers)
ALTER TABLE orders 
ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id) ON DELETE SET NULL;

-- Update existing orders to migrate customer_id to user_id if user_id is null
-- This assumes there might be a relationship between customers and users tables
-- For now, we'll set user_id to NULL for existing orders and let the application handle it

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

-- Ensure both customer_id and user_id can coexist for backward compatibility
-- The application can choose which one to use based on the context

COMMENT ON COLUMN orders.user_id IS 'Reference to users table for authenticated user orders';
COMMENT ON COLUMN orders.customer_id IS 'Legacy reference to customers table - may be deprecated in favor of user_id';
