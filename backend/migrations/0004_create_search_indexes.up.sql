-- Migration 0004: Create full-text search indexes and infrastructure
-- Enables fast product search via PostgreSQL full-text search + pg_trgm fuzzy matching

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;

-- Add search-related columns to products table
ALTER TABLE products
ADD COLUMN IF NOT EXISTS brand TEXT,
ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}',
ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Add SKU column reference for easier searching (denormalized from variants)
ALTER TABLE products
ADD COLUMN IF NOT EXISTS primary_sku TEXT;

-- Create a function to generate the search vector
-- Weights: A=title (highest), B=brand+tags, C=description+sku
CREATE OR REPLACE FUNCTION products_search_vector_update()
RETURNS TRIGGER AS $$
BEGIN
  NEW.search_vector := to_tsvector('english', 
    coalesce(NEW.title, '') || ' ' ||
    coalesce(NEW.brand, '') || ' ' ||
    array_to_string(coalesce(NEW.tags, '{}'), ' ') || ' ' ||
    coalesce(NEW.description, '') || ' ' ||
    coalesce(NEW.primary_sku, '')
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to update search vector on insert/update
DROP TRIGGER IF EXISTS products_search_vector_trigger ON products;
CREATE TRIGGER products_search_vector_trigger
BEFORE INSERT OR UPDATE ON products
FOR EACH ROW
EXECUTE FUNCTION products_search_vector_update();

-- Re-generate search vectors for all existing products
UPDATE products SET search_vector = to_tsvector('english',
  coalesce(title, '') || ' ' ||
  coalesce(brand, '') || ' ' ||
  array_to_string(coalesce(tags, '{}'), ' ') || ' ' ||
  coalesce(description, '') || ' ' ||
  coalesce(primary_sku, '')
);

-- Create GIN index on search_vector for full-text search (primary index)
CREATE INDEX idx_products_search_vector ON products USING GIN(search_vector);

-- Create GIN indexes on title and brand for fuzzy/prefix matching (pg_trgm)
CREATE INDEX idx_products_title_trgm ON products USING GIN(title gin_trgm_ops);
CREATE INDEX idx_products_brand_trgm ON products USING GIN(brand gin_trgm_ops);
CREATE INDEX idx_products_sku_trgm ON products USING GIN(primary_sku gin_trgm_ops);

-- Create index on published status + publish_at for visibility filtering
CREATE INDEX idx_products_published_status ON products(published, publish_at, unpublish_at);

-- Create materialized view for fast faceted search (category + price ranges)
-- This will be refreshed periodically (e.g., hourly cron job)
CREATE OR REPLACE VIEW products_search_facets AS
SELECT
  pc.category_id,
  c.name AS category_name,
  COUNT(DISTINCT p.id) AS product_count,
  MIN(pv.price_cents) AS min_price_cents,
  MAX(pv.price_cents) AS max_price_cents,
  AVG(pv.price_cents)::INT AS avg_price_cents
FROM products p
LEFT JOIN product_categories pc ON p.id = pc.product_id
LEFT JOIN categories c ON pc.category_id = c.id
LEFT JOIN product_variants pv ON p.id = pv.product_id
WHERE p.published = TRUE AND (p.publish_at IS NULL OR p.publish_at <= NOW())
  AND (p.unpublish_at IS NULL OR p.unpublish_at > NOW())
GROUP BY pc.category_id, c.name;

-- Grant permissions (if using role-based access)
-- GRANT SELECT ON products_search_facets TO app_user;
