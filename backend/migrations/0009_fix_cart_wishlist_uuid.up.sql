-- Migration: Fix cart and wishlist tables to use UUID for product_id

-- Drop existing tables (they have wrong schema)
DROP TABLE IF EXISTS cart CASCADE;
DROP TABLE IF EXISTS wishlist CASCADE;

-- Recreate cart table with UUID customer_id, UUID product_id and integer variant_id
CREATE TABLE cart (
    id SERIAL PRIMARY KEY,
    session_id TEXT, -- For guest users
    customer_id UUID REFERENCES customers(id) ON DELETE CASCADE, -- For authenticated users
    product_id UUID REFERENCES products(uuid_id) ON DELETE CASCADE,
    variant_id INT REFERENCES product_variants(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Ensure either session_id or customer_id is present
    CONSTRAINT cart_user_or_session CHECK (
        (session_id IS NOT NULL AND customer_id IS NULL) OR 
        (session_id IS NULL AND customer_id IS NOT NULL) OR
        (session_id IS NULL AND customer_id IS NULL) -- Allow empty for admin operations
    )
);

CREATE INDEX idx_cart_session ON cart(session_id);
CREATE INDEX idx_cart_customer ON cart(customer_id);
CREATE INDEX idx_cart_product ON cart(product_id);

-- Recreate wishlist table with UUID customer_id and UUID product_id
CREATE TABLE wishlist (
    id SERIAL PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(uuid_id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(customer_id, product_id)
);

CREATE INDEX idx_wishlist_customer ON wishlist(customer_id);
CREATE INDEX idx_wishlist_product ON wishlist(product_id);
