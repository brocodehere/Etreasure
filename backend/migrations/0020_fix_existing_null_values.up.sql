-- Migration: Fix existing NULL values in orders table
-- This migration updates any existing NULL values to match the NOT NULL constraints

UPDATE orders 
SET 
    total_price = COALESCE(total_price, 0),
    subtotal = COALESCE(subtotal, 0),
    tax_amount = COALESCE(tax_amount, 0),
    shipping_amount = COALESCE(shipping_amount, 0),
    discount_amount = COALESCE(discount_amount, 0)
WHERE 
    total_price IS NULL OR 
    subtotal IS NULL OR 
    tax_amount IS NULL OR 
    shipping_amount IS NULL OR 
    discount_amount IS NULL;
