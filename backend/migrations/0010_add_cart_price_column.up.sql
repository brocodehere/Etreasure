-- Migration: Add price column to cart table for storing discounted prices

-- Add price_cents column to cart table
ALTER TABLE cart ADD COLUMN price_cents INT;

-- Update existing cart items to use current product prices
UPDATE cart 
SET price_cents = (
    SELECT pv.price_cents 
    FROM product_variants pv 
    WHERE pv.id = cart.variant_id
)
WHERE price_cents IS NULL;

-- Create index for better performance
CREATE INDEX idx_cart_price ON cart(price_cents);
