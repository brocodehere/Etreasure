-- Migration: Add session_id support to wishlist table for guest users

-- Add session_id column to wishlist table
ALTER TABLE wishlist ADD COLUMN session_id TEXT;

-- Update constraint to allow either session_id or customer_id
ALTER TABLE wishlist DROP CONSTRAINT IF EXISTS wishlist_customer_id_fkey;
ALTER TABLE wishlist ADD CONSTRAINT wishlist_user_or_session CHECK (
    (session_id IS NOT NULL AND customer_id IS NULL) OR 
    (session_id IS NULL AND customer_id IS NOT NULL)
);

-- Create index for session_id
CREATE INDEX idx_wishlist_session ON wishlist(session_id);

-- Update existing records to use session_id for demo purposes
-- This converts existing customer_id records to session_id format for local development
UPDATE wishlist 
SET session_id = 'demo-session-127.0.0.1' 
WHERE customer_id = (SELECT id FROM customers WHERE uuid_id = '53545e0e-8a69-4bb5-be76-7f8176ac906a' LIMIT 1);
