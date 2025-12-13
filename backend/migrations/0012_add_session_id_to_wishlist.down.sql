-- Migration: Remove session_id support from wishlist table

-- Drop the session_id index
DROP INDEX IF EXISTS idx_wishlist_session;

-- Remove the constraint
ALTER TABLE wishlist DROP CONSTRAINT IF EXISTS wishlist_user_or_session;

-- Remove the session_id column
ALTER TABLE wishlist DROP COLUMN IF EXISTS session_id;

-- Recreate the original foreign key constraint
ALTER TABLE wishlist ADD CONSTRAINT wishlist_customer_id_fkey 
    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE;
