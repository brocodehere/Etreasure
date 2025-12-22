-- Fix NULL values in price and total columns
UPDATE order_line_items 
SET 
    price = CASE 
        WHEN price IS NULL AND unit_price IS NOT NULL AND unit_price != '' THEN
            CAST(unit_price AS NUMERIC(10, 2))
        WHEN price IS NULL AND quantity > 0 THEN
            999.00 -- Default price
        ELSE price
    END,
    total = CASE 
        WHEN total IS NULL AND total_price IS NOT NULL AND total_price != '' THEN
            CAST(total_price AS NUMERIC(10, 2))
        WHEN total IS NULL AND price IS NOT NULL AND quantity > 0 THEN
            price * quantity
        WHEN total IS NULL AND quantity > 0 THEN
            999.00 * quantity -- Default total
        ELSE total
    END
WHERE price IS NULL OR total IS NULL;
