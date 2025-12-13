-- Migration 0004 Down: Drop search infrastructure
-- This reverses the full-text search and fuzzy matching setup

-- Drop views first (views depend on tables)
DROP VIEW IF EXISTS products_search_facets;

-- Drop triggers
DROP TRIGGER IF EXISTS products_search_vector_trigger ON products;

-- Drop trigger function
DROP FUNCTION IF EXISTS products_search_vector_update();

-- Drop indexes
DROP INDEX IF EXISTS idx_products_search_vector;
DROP INDEX IF EXISTS idx_products_title_trgm;
DROP INDEX IF EXISTS idx_products_brand_trgm;
DROP INDEX IF EXISTS idx_products_sku_trgm;
DROP INDEX IF EXISTS idx_products_published_status;

-- Drop columns (if they didn't exist before)
ALTER TABLE products
DROP COLUMN IF EXISTS search_vector,
DROP COLUMN IF EXISTS brand,
DROP COLUMN IF EXISTS tags,
DROP COLUMN IF EXISTS primary_sku;

-- Drop extensions (careful: may break other systems if they use these)
-- Uncomment only if you're sure no other tables use these extensions
-- DROP EXTENSION IF EXISTS pg_trgm;
-- DROP EXTENSION IF EXISTS unaccent;
