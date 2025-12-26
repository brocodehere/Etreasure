-- Migration: Add session_id to wishlist table for guest users

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

-- Allow NULL customer_id for guest users
ALTER TABLE wishlist ALTER COLUMN customer_id DROP NOT NULL;
