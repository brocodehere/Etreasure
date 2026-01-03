-- Fix stock notifications table to use UUID instead of INTEGER for product_id
-- This migration updates the table to properly reference products(uuid_id)

-- First, create a backup of existing data
CREATE TABLE IF NOT EXISTS stock_notifications_backup AS 
SELECT * FROM stock_notifications;

-- Drop the foreign key constraint (this will be recreated)
ALTER TABLE stock_notifications DROP CONSTRAINT IF EXISTS stock_notifications_product_id_fkey;

-- Add a new UUID column for product reference
ALTER TABLE stock_notifications ADD COLUMN IF NOT EXISTS product_uuid UUID;

-- Update the new column with UUID data from products table using slug
UPDATE stock_notifications 
SET product_uuid = p.uuid_id 
FROM products p 
WHERE stock_notifications.product_slug = p.slug;

-- Make the new column NOT NULL (only if we have data)
-- This might fail if there are notifications without matching products
-- ALTER TABLE stock_notifications ALTER COLUMN product_uuid SET NOT NULL;

-- Drop the old integer column
ALTER TABLE stock_notifications DROP COLUMN IF EXISTS product_id;

-- Rename the UUID column to product_id
ALTER TABLE stock_notifications RENAME COLUMN product_uuid TO product_id;

-- Add the proper foreign key constraint
ALTER TABLE stock_notifications 
ADD CONSTRAINT stock_notifications_product_id_fkey 
FOREIGN KEY (product_id) REFERENCES products(uuid_id) ON DELETE CASCADE;

-- Update the index
DROP INDEX IF EXISTS idx_stock_notifications_product_id;
CREATE INDEX idx_stock_notifications_product_id ON stock_notifications(product_id);

-- Make product_id NOT NULL if all records have valid UUIDs
DO $$
BEGIN
    -- Check if all notifications have valid product_uuid
    IF NOT EXISTS (SELECT 1 FROM stock_notifications WHERE product_id IS NULL) THEN
        ALTER TABLE stock_notifications ALTER COLUMN product_id SET NOT NULL;
    END IF;
END $$;
