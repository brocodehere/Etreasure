-- Rollback migration: Convert UUID categories back to SERIAL

-- Drop the UUID index
DROP INDEX IF EXISTS idx_categories_uuid_id;

-- Add back the integer category_id column to product_categories
ALTER TABLE product_categories ADD COLUMN category_id_int INT;

-- Update the new column with integer from categories
UPDATE product_categories pc
SET category_id_int = c.id
FROM categories c
WHERE pc.category_id = c.uuid_id;

-- Drop UUID foreign key and column
ALTER TABLE product_categories DROP CONSTRAINT product_categories_category_id_fkey;
ALTER TABLE product_categories DROP COLUMN category_id;

-- Rename the integer column to category_id
ALTER TABLE product_categories RENAME COLUMN category_id_int TO category_id;

-- Add back the foreign key constraint
ALTER TABLE product_categories 
ADD CONSTRAINT product_categories_category_id_fkey 
FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE;

-- Recreate the primary key
ALTER TABLE product_categories DROP CONSTRAINT product_categories_pkey;
ALTER TABLE product_categories ADD PRIMARY KEY (product_id, category_id);

-- Drop the UUID column from categories
ALTER TABLE categories DROP COLUMN uuid_id;
