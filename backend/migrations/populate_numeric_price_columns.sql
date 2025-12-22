-- Migration to populate numeric price and total columns from varchar columns
UPDATE order_line_items 
SET 
    price = CASE 
        WHEN unit_price IS NOT NULL AND unit_price != '' THEN
            CAST(unit_price AS NUMERIC(10, 2))
        WHEN price IS NULL THEN 0
        ELSE price
    END,
    total = CASE 
        WHEN total_price IS NOT NULL AND total_price != '' THEN
            CAST(total_price AS NUMERIC(10, 2))
        WHEN total IS NULL THEN 0
        ELSE total
    END
WHERE price IS NULL OR total IS NULL;
