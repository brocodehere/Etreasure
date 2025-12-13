-- Migration: Remove seeded cart and wishlist data

-- Remove sample cart items
DELETE FROM cart WHERE session_id IN ('demo-session-1', 'demo-session-2');

-- Remove sample wishlist items
DELETE FROM wishlist WHERE customer_id = 1;
