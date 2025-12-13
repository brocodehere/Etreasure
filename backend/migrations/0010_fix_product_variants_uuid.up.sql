-- Migration: Fix product_variants table to use UUID for both primary key and product_id

-- Drop existing product_variants table (it has wrong schema)
DROP TABLE IF EXISTS product_variants CASCADE;

-- Recreate product_variants table with UUID primary key and product_id
CREATE TABLE product_variants (
    uuid_id UUID DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
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
CREATE INDEX idx_product_variants_uuid_id ON product_variants(uuid_id);
