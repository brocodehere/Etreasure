-- Migration: Add price fields to orders table
-- This migration adds the missing price fields that the Go code expects

ALTER TABLE orders 
ADD COLUMN IF NOT EXISTS total_price DECIMAL(10,2),
ADD COLUMN IF NOT EXISTS subtotal DECIMAL(10,2),
ADD COLUMN IF NOT EXISTS tax_amount DECIMAL(10,2),
ADD COLUMN IF NOT EXISTS shipping_amount DECIMAL(10,2),
ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(10,2);

-- Update existing orders to have default values based on total_cents if available
UPDATE orders 
SET 
    total_price = CASE 
        WHEN total_cents IS NOT NULL THEN total_cents::decimal / 100.0
        ELSE 0.0
    END,
    subtotal = CASE 
        WHEN total_cents IS NOT NULL THEN total_cents::decimal / 100.0
        ELSE 0.0
    END
WHERE total_price IS NULL OR subtotal IS NULL;

-- Add comments for documentation
COMMENT ON COLUMN orders.total_price IS 'Total order price in decimal format';
COMMENT ON COLUMN orders.subtotal IS 'Subtotal before tax and shipping';
COMMENT ON COLUMN orders.tax_amount IS 'Tax amount applied to order';
COMMENT ON COLUMN orders.shipping_amount IS 'Shipping charges for order';
COMMENT ON COLUMN orders.discount_amount IS 'Discount applied to order';
