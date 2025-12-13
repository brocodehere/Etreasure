-- Migration: Convert categories to UUID and update product_categories relationship

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Add UUID column to categories table
ALTER TABLE categories ADD COLUMN uuid_id UUID DEFAULT uuid_generate_v4() NOT NULL;

-- Create a temporary mapping table for old_id to new_uuid
CREATE TEMP TABLE category_id_mapping AS
SELECT id, uuid_id FROM categories;

-- Update product_categories to use UUID instead of integer
-- First, add a new UUID column to product_categories
ALTER TABLE product_categories ADD COLUMN category_uuid UUID;

-- Update the new column with UUID from categories
UPDATE product_categories pc
SET category_uuid = c.uuid_id
FROM categories c
WHERE pc.category_id = c.id;

-- Drop old foreign key and column
ALTER TABLE product_categories DROP CONSTRAINT product_categories_category_id_fkey;
ALTER TABLE product_categories DROP COLUMN category_id;

-- Rename the UUID column to category_id
ALTER TABLE product_categories RENAME COLUMN category_uuid TO category_id;

-- Add new foreign key constraint
ALTER TABLE product_categories 
ADD CONSTRAINT product_categories_category_id_fkey 
FOREIGN KEY (category_id) REFERENCES categories(uuid_id) ON DELETE CASCADE;

-- Drop the old primary key and create new one
ALTER TABLE product_categories DROP CONSTRAINT product_categories_pkey;
ALTER TABLE product_categories ADD PRIMARY KEY (product_id, category_id);

-- Update any other references if needed (check for tables that reference categories.id)
-- For now, we'll keep the old id column for backward compatibility during transition

-- Create index on the new UUID column for performance
CREATE INDEX idx_categories_uuid_id ON categories(uuid_id);
