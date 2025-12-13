-- Domain schema for products, categories, media, orders, etc.

-- WARNING: Running this file will drop existing domain tables and ALL data.
-- This is intended for development convenience only.

-- Drop in reverse dependency order to avoid FK issues
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS settings CASCADE;
DROP TABLE IF EXISTS stock_transfers CASCADE;
DROP TABLE IF EXISTS inventory_levels CASCADE;
DROP TABLE IF EXISTS inventory_locations CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS customer_tag_links CASCADE;
DROP TABLE IF EXISTS customer_tags CASCADE;
DROP TABLE IF EXISTS customers CASCADE;
DROP TABLE IF EXISTS offers CASCADE;
DROP TABLE IF EXISTS banners CASCADE;
DROP TABLE IF EXISTS product_images CASCADE;
DROP TABLE IF EXISTS media CASCADE;
DROP TABLE IF EXISTS product_categories CASCADE;
DROP TABLE IF EXISTS product_variants CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS categories CASCADE;

CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    parent_id INT REFERENCES categories(id) ON DELETE SET NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent_sort ON categories(parent_id, sort_order);
CREATE INDEX idx_categories_slug ON categories(slug);

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    slug TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    subtitle TEXT,
    description TEXT,
    seo_title TEXT,
    seo_description TEXT,
    published BOOLEAN NOT NULL DEFAULT FALSE,
    publish_at TIMESTAMPTZ,
    unpublish_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_slug ON products(slug);
CREATE INDEX idx_products_publish_at ON products(publish_at);

CREATE TABLE product_variants (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
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

CREATE TABLE product_categories (
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, category_id)
);

CREATE TABLE media (
    id SERIAL PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,
    original_filename TEXT,
    mime_type TEXT,
    file_size_bytes BIGINT,
    width INT,
    height INT,
    dominant_color TEXT,
    lqip_base64 TEXT,
    variants JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_path ON media(path);
CREATE INDEX idx_media_created_at ON media(created_at);

CREATE TABLE product_images (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    media_id INT NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_product_images_product_sort ON product_images(product_id, sort_order);

CREATE TABLE banners (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    slot TEXT NOT NULL,
    target_url TEXT,
    priority INT NOT NULL DEFAULT 0,
    start_at TIMESTAMPTZ,
    end_at TIMESTAMPTZ,
    desktop_media_id INT REFERENCES media(id) ON DELETE SET NULL,
    mobile_media_id INT REFERENCES media(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_banners_slot_priority ON banners(slot, priority);

CREATE TABLE offers (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL, -- percent,fixed,bogo
    value_cents INT,
    percent_off NUMERIC(5,2),
    bogo_buy_qty INT,
    bogo_get_qty INT,
    coupon_code TEXT,
    max_uses INT,
    used_count INT NOT NULL DEFAULT 0,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_offers_active ON offers(is_active, starts_at, ends_at);
CREATE INDEX idx_offers_coupon_code ON offers(coupon_code);

CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    first_name TEXT,
    last_name TEXT,
    phone TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE customer_tags (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE customer_tag_links (
    customer_id INT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    tag_id INT NOT NULL REFERENCES customer_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (customer_id, tag_id)
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    order_number TEXT UNIQUE NOT NULL,
    customer_id INT REFERENCES customers(id) ON DELETE SET NULL,
    status TEXT NOT NULL,
    total_cents INT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'INR',
    placed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    fulfilled_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    refund_cents INT NOT NULL DEFAULT 0,
    shipping_address JSONB,
    billing_address JSONB,
    tracking_info JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_customer ON orders(customer_id);

CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INT REFERENCES products(id) ON DELETE SET NULL,
    variant_id INT REFERENCES product_variants(id) ON DELETE SET NULL,
    quantity INT NOT NULL,
    unit_price_cents INT NOT NULL,
    total_price_cents INT NOT NULL
);

CREATE INDEX idx_order_items_order ON order_items(order_id);

CREATE TABLE inventory_locations (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    code TEXT UNIQUE NOT NULL
);

CREATE TABLE inventory_levels (
    id SERIAL PRIMARY KEY,
    location_id INT NOT NULL REFERENCES inventory_locations(id) ON DELETE CASCADE,
    variant_id INT NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 0,
    UNIQUE(location_id, variant_id)
);

CREATE TABLE stock_transfers (
    id SERIAL PRIMARY KEY,
    from_location_id INT REFERENCES inventory_locations(id) ON DELETE SET NULL,
    to_location_id INT REFERENCES inventory_locations(id) ON DELETE SET NULL,
    variant_id INT NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value JSONB NOT NULL
);

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id INT REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    object_type TEXT NOT NULL,
    object_id TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_object ON audit_logs(object_type, object_id, created_at DESC);
