DROP INDEX IF EXISTS idx_audit_logs_object;
DROP TABLE IF EXISTS audit_logs;

DROP TABLE IF EXISTS settings;

DROP TABLE IF EXISTS stock_transfers;
DROP TABLE IF EXISTS inventory_levels;
DROP TABLE IF EXISTS inventory_locations;

DROP INDEX IF EXISTS idx_order_items_order;
DROP TABLE IF EXISTS order_items;

DROP INDEX IF EXISTS idx_orders_customer;
DROP INDEX IF EXISTS idx_orders_status;
DROP TABLE IF EXISTS orders;

DROP TABLE IF EXISTS customer_tag_links;
DROP TABLE IF EXISTS customer_tags;
DROP TABLE IF EXISTS customers;

DROP INDEX IF EXISTS idx_offers_coupon_code;
DROP INDEX IF EXISTS idx_offers_active;
DROP TABLE IF EXISTS offers;

DROP INDEX IF EXISTS idx_banners_slot_priority;
DROP TABLE IF EXISTS banners;

DROP INDEX IF EXISTS idx_product_images_product_sort;
DROP TABLE IF EXISTS product_images;

DROP INDEX IF EXISTS idx_media_created_at;
DROP INDEX IF EXISTS idx_media_path;
DROP TABLE IF EXISTS media;

DROP TABLE IF EXISTS product_categories;

DROP INDEX IF EXISTS idx_product_variants_product;
DROP TABLE IF EXISTS product_variants;

DROP INDEX IF EXISTS idx_products_publish_at;
DROP INDEX IF EXISTS idx_products_slug;
DROP TABLE IF EXISTS products;

DROP INDEX IF EXISTS idx_categories_slug;
DROP INDEX IF EXISTS idx_categories_parent_sort;
DROP TABLE IF EXISTS categories;
