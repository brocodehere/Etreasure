-- Migration: Add discounted_price column to cart table

-- Add discounted_price column to cart table
ALTER TABLE cart ADD COLUMN discounted_price VARCHAR(255);

-- Create index for better performance
CREATE INDEX idx_cart_discounted_price ON cart(discounted_price);
