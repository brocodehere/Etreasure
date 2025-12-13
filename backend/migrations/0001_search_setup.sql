-- 0001_search_setup.sql
-- Setup full-text search helpers and indexes for products, categories, offers, banners
-- Enables pg_trgm and unaccent which improve fuzzy matching and accent-insensitive search

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;

-- Add search_vector column to products (weighted tsvector)
ALTER TABLE products
  ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Backfill search_vector using weighted fields:
-- title (A), tags/brand (B), description (C) — adjust weights as needed
UPDATE products
SET search_vector = (
  setweight(to_tsvector('simple', coalesce(unaccent(title),'')), 'A') ||
  setweight(to_tsvector('simple', coalesce(unaccent(array_to_string(tags, ' '),''))), 'B') ||
  setweight(to_tsvector('simple', coalesce(unaccent(brand),'')), 'B') ||
  setweight(to_tsvector('simple', coalesce(unaccent(description),'')), 'C')
)
WHERE search_vector IS NULL;

-- Create trigger function to maintain products.search_vector
CREATE OR REPLACE FUNCTION products_search_vector_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('simple', coalesce(unaccent(NEW.title),'')), 'A') ||
    setweight(to_tsvector('simple', coalesce(unaccent(array_to_string(NEW.tags, ' '),''))), 'B') ||
    setweight(to_tsvector('simple', coalesce(unaccent(NEW.brand),'')), 'B') ||
    setweight(to_tsvector('simple', coalesce(unaccent(NEW.description),'')), 'C');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for insert/update
DROP TRIGGER IF EXISTS trg_products_search_vector ON products;
CREATE TRIGGER trg_products_search_vector
BEFORE INSERT OR UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION products_search_vector_trigger();

-- GIN index on tsvector for fast full-text queries
CREATE INDEX IF NOT EXISTS idx_products_search_vector ON products USING GIN (search_vector);

-- Trigram indexes on lower(title) and lower(slug) to accelerate fuzzy/ILIKE queries
CREATE INDEX IF NOT EXISTS idx_products_title_trgm ON products USING GIN (lower(title) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_products_slug_trgm ON products USING GIN (lower(slug) gin_trgm_ops);

-- Lightweight indexes for categories, offers, banners (trigram on title)
ALTER TABLE categories ADD COLUMN IF NOT EXISTS search_text TEXT;
UPDATE categories SET search_text = coalesce(unaccent(title),'') || ' ' || coalesce(unaccent(coalesce(description,'')), '');
CREATE INDEX IF NOT EXISTS idx_categories_search_trgm ON categories USING GIN (lower(search_text) gin_trgm_ops);

ALTER TABLE offers ADD COLUMN IF NOT EXISTS search_text TEXT;
UPDATE offers SET search_text = coalesce(unaccent(title),'') || ' ' || coalesce(unaccent(coalesce(description,'')), '');
CREATE INDEX IF NOT EXISTS idx_offers_search_trgm ON offers USING GIN (lower(search_text) gin_trgm_ops);

ALTER TABLE banners ADD COLUMN IF NOT EXISTS search_text TEXT;
UPDATE banners SET search_text = coalesce(unaccent(title),'') || ' ' || coalesce(unaccent(coalesce(body,'')), '');
CREATE INDEX IF NOT EXISTS idx_banners_search_trgm ON banners USING GIN (lower(search_text) gin_trgm_ops);

-- Notes:
-- 1) We use 'simple' configuration to keep tokenization predictable and rely on pg_trgm for fuzzy matching.
-- 2) unaccent() normalizes accents; if your managed DB doesn't allow extensions, enable them at the provider console.
