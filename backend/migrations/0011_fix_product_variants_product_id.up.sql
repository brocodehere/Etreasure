-- Migration: Fix product_variants table to use UUID for product_id reference

-- Check if product_variants table exists and has wrong product_id type
-- Drop and recreate with correct UUID reference
DROP TABLE IF EXISTS product_variants CASCADE;

-- Recreate product_variants table with correct UUID product_id reference
CREATE TABLE product_variants (
    id SERIAL PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products(uuid_id) ON DELETE CASCADE,
    sku TEXT UNIQUE NOT NULL,
    title TEXT,
    price_cents INT NOT NULL,
    compare_at_price_cents INT,
    currency TEXT NOT NULL DEFAULT 'INR',
    stock_quantity INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_variants_product ON product_variants(product_id);
