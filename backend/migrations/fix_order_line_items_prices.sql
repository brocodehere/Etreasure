-- Migration to fix null price and total values in order_line_items table
-- This script calculates price and total based on order subtotal and quantity distribution

UPDATE order_line_items 
SET 
    price = CASE 
        WHEN price IS NULL THEN 
            (SELECT COALESCE(subtotal, 0) FROM orders WHERE id = order_id) / 
            (SELECT COUNT(*) FROM order_line_items WHERE order_id = order_line_items.order_id)
        ELSE price 
    END,
    total = CASE 
        WHEN total IS NULL THEN 
            quantity * CASE 
                WHEN price IS NULL THEN 
                    (SELECT COALESCE(subtotal, 0) FROM orders WHERE id = order_id) / 
                    (SELECT COUNT(*) FROM order_line_items WHERE order_id = order_line_items.order_id)
                ELSE price 
            END
        ELSE total 
    END
WHERE price IS NULL OR total IS NULL;

-- Verify the update
SELECT 
    oli.id,
    oli.order_id,
    oli.title,
    oli.sku,
    oli.quantity,
    oli.price,
    oli.total,
    o.subtotal as order_subtotal
FROM order_line_items oli
JOIN orders o ON oli.order_id = o.id
WHERE oli.order_id = '213a63b0-d097-4fd3-8f00-b6ffedf65ee8';
