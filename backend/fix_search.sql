-- Fix search functionality by setting up search_vector and indexes

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;

-- Add search_vector column if it doesn't exist
ALTER TABLE products
  ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Populate search_vector for all products
UPDATE products
SET search_vector = (
  setweight(to_tsvector('simple', coalesce(unaccent(title),'')), 'A') ||
  setweight(to_tsvector('simple', coalesce(unaccent(description),'')), 'C')
)
WHERE search_vector IS NULL;

-- Create trigger to maintain search_vector
CREATE OR REPLACE FUNCTION products_search_vector_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('simple', coalesce(unaccent(NEW.title),'')), 'A') ||
    setweight(to_tsvector('simple', coalesce(unaccent(NEW.description),'')), 'C');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS products_search_vector_update ON products;
CREATE TRIGGER products_search_vector_update
  BEFORE INSERT OR UPDATE ON products
  FOR EACH ROW EXECUTE FUNCTION products_search_vector_trigger();

-- Create GIN index for fast search
CREATE INDEX IF NOT EXISTS idx_products_search_vector ON products USING GIN(search_vector);

-- Add trigram index for fallback search
CREATE INDEX IF NOT EXISTS idx_products_title_trgm ON products USING GIN (lower(title) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_products_slug_trgm ON products USING GIN (lower(slug) gin_trgm_ops);

-- Verify search_vector is populated
SELECT COUNT(*) as products_with_search_vector FROM products WHERE search_vector IS NOT NULL;
