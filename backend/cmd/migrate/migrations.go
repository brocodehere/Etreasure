package main

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// registerMigrations registers all database migrations
func registerMigrations(m *Migrator) {

	// Migration 001: Create core tables
	m.AddMigration("001", "Create core authentication and user tables",
		func(ctx context.Context, tx pgx.Tx) error {
			// Users table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS users (
					id SERIAL PRIMARY KEY,
					email TEXT UNIQUE NOT NULL,
					password_hash TEXT NOT NULL,
					full_name TEXT,
					is_active BOOLEAN NOT NULL DEFAULT TRUE,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Roles table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS roles (
					id SERIAL PRIMARY KEY,
					name TEXT UNIQUE NOT NULL,
					description TEXT,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// User roles junction table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS user_roles (
					user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
					PRIMARY KEY (user_id, role_id)
				)
			`); err != nil {
				return err
			}

			// Insert default roles
			if _, err := tx.Exec(ctx, `
				INSERT INTO roles (name, description) VALUES 
				('SuperAdmin', 'Full system access'),
				('Admin', 'Administrative access'),
				('Manager', 'Managerial access'),
				('Staff', 'Staff access')
				ON CONFLICT (name) DO NOTHING
			`); err != nil {
				return err
			}

			// Create indexes
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active)"); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			// Drop tables in reverse dependency order
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS user_roles CASCADE"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS roles CASCADE"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS users CASCADE"); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 002: Create categories table
	m.AddMigration("002", "Create categories table",
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS categories (
					id SERIAL PRIMARY KEY,
					slug TEXT UNIQUE NOT NULL,
					name TEXT NOT NULL,
					description TEXT,
					parent_id INT REFERENCES categories(id) ON DELETE SET NULL,
					sort_order INT NOT NULL DEFAULT 0,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Create indexes
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_categories_parent_sort ON categories(parent_id, sort_order)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug)"); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS categories CASCADE"); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 003: Create products and related tables
	m.AddMigration("003", "Create products and related tables",
		func(ctx context.Context, tx pgx.Tx) error {
			// Products table (without category_id initially)
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS products (
					uuid_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
				)
			`); err != nil {
				return err
			}

			// Product variants table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS product_variants (
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
				)
			`); err != nil {
				return err
			}

			// Create indexes
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_products_slug ON products(slug)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_products_publish_at ON products(publish_at)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_product_variants_product ON product_variants(product_id)"); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS product_variants CASCADE"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS products CASCADE"); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 004: Create media and product images tables
	m.AddMigration("004", "Create media and product images tables",
		func(ctx context.Context, tx pgx.Tx) error {
			// Media table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS media (
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
				)
			`); err != nil {
				return err
			}

			// Product images table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS product_images (
					id SERIAL PRIMARY KEY,
					product_id UUID NOT NULL REFERENCES products(uuid_id) ON DELETE CASCADE,
					media_id INT NOT NULL REFERENCES media(id) ON DELETE CASCADE,
					sort_order INT NOT NULL DEFAULT 0
				)
			`); err != nil {
				return err
			}

			// Create indexes
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_media_path ON media(path)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_media_created_at ON media(created_at)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_product_images_product_sort ON product_images(product_id, sort_order)"); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS product_images CASCADE"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS media CASCADE"); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 005: Create banners table (matching handler schema)
	m.AddMigration("005", "Create banners table",
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS banners (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					title TEXT NOT NULL,
					image_url TEXT NOT NULL,
					link_url TEXT,
					is_active BOOLEAN NOT NULL DEFAULT TRUE,
					sort_order INT NOT NULL DEFAULT 0,
					starts_at TIMESTAMPTZ,
					ends_at TIMESTAMPTZ,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Create indexes
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_banners_active ON banners(is_active)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_banners_sort ON banners(sort_order)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_banners_dates ON banners(starts_at, ends_at)"); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS banners CASCADE"); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 006: Create offers table (matching handler schema)
	m.AddMigration("006", "Create offers table",
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS offers (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					title TEXT NOT NULL,
					description TEXT,
					discount_type TEXT NOT NULL, -- percentage | fixed
					discount_value NUMERIC(10,2) NOT NULL,
					applies_to TEXT NOT NULL, -- all | products | categories | collections
					applies_to_ids TEXT, -- comma-separated IDs
					min_order_amount NUMERIC(10,2) DEFAULT 0,
					usage_limit INT,
					usage_count INT NOT NULL DEFAULT 0,
					is_active BOOLEAN NOT NULL DEFAULT TRUE,
					starts_at TIMESTAMPTZ,
					ends_at TIMESTAMPTZ,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Create indexes
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_offers_active ON offers(is_active, starts_at, ends_at)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_offers_type ON offers(discount_type)"); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS offers CASCADE"); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 007: Create customers and addresses tables (matching handler schema)
	m.AddMigration("007", "Create customers and addresses tables",
		func(ctx context.Context, tx pgx.Tx) error {
			// Customers table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS customers (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					email TEXT UNIQUE NOT NULL,
					first_name TEXT,
					last_name TEXT,
					phone TEXT,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Addresses table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS addresses (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
					type TEXT NOT NULL, -- shipping | billing
					first_name TEXT,
					last_name TEXT,
					company TEXT,
					address1 TEXT,
					address2 TEXT,
					city TEXT,
					province TEXT,
					country TEXT,
					postal_code TEXT,
					phone TEXT,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Create indexes
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_addresses_customer ON addresses(customer_id)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_addresses_type ON addresses(type)"); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS addresses CASCADE"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS customers CASCADE"); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 008: Create orders and order items tables (matching handler schema)
	m.AddMigration("008", "Create orders and order items tables",
		func(ctx context.Context, tx pgx.Tx) error {
			// Orders table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS orders (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					order_number TEXT UNIQUE NOT NULL,
					customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
					status TEXT NOT NULL DEFAULT 'pending', -- pending | processing | shipped | delivered | cancelled
					currency TEXT NOT NULL DEFAULT 'INR',
					total_price NUMERIC(10,2) NOT NULL,
					subtotal NUMERIC(10,2) NOT NULL,
					tax_amount NUMERIC(10,2) DEFAULT 0,
					shipping_amount NUMERIC(10,2) DEFAULT 0,
					discount_amount NUMERIC(10,2) DEFAULT 0,
					shipping_address JSONB,
					billing_address JSONB,
					notes TEXT,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Order line items table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS order_line_items (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
					product_id UUID REFERENCES products(uuid_id) ON DELETE SET NULL,
					variant_id INT REFERENCES product_variants(id) ON DELETE SET NULL,
					title TEXT,
					sku TEXT,
					quantity INT NOT NULL,
					price NUMERIC(10,2) NOT NULL,
					total NUMERIC(10,2) NOT NULL
				)
			`); err != nil {
				return err
			}

			// Create indexes
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_orders_number ON orders(order_number)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_line_items(order_id)"); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS order_line_items CASCADE"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS orders CASCADE"); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 009: Create inventory tables (matching handler schema)
	m.AddMigration("009", "Create inventory tables",
		func(ctx context.Context, tx pgx.Tx) error {
			// Inventory items table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS inventory_items (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					product_id UUID NOT NULL REFERENCES products(uuid_id) ON DELETE CASCADE,
					variant_id INT REFERENCES product_variants(id) ON DELETE SET NULL,
					sku TEXT NOT NULL,
					quantity INT NOT NULL DEFAULT 0,
					reserved INT NOT NULL DEFAULT 0,
					available INT NOT NULL DEFAULT 0,
					location TEXT,
					cost_price NUMERIC(10,2),
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Inventory adjustments table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS inventory_adjustments (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					item_id UUID NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
					quantity INT NOT NULL, -- positive for addition, negative for subtraction
					reason TEXT NOT NULL,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Create indexes
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_inventory_items_product ON inventory_items(product_id)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_inventory_items_variant ON inventory_items(variant_id)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_inventory_items_sku ON inventory_items(sku)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_inventory_adjustments_item ON inventory_adjustments(item_id)"); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS inventory_adjustments CASCADE"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS inventory_items CASCADE"); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 010: Create settings table (matching handler schema)
	m.AddMigration("010", "Create settings table",
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS settings (
					key TEXT PRIMARY KEY,
					value TEXT NOT NULL,
					type TEXT NOT NULL DEFAULT 'string', -- string | number | boolean | json
					description TEXT,
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Insert default settings
			if _, err := tx.Exec(ctx, `
				INSERT INTO settings (key, value, type, description) VALUES 
				('store_name', '"Etreasure Store"', 'json', 'Store name'),
				('store_description', '"Your premium online store"', 'json', 'Store description'),
				('currency', '"INR"', 'json', 'Default currency'),
				('contact_email', '"contact@etreasure.com"', 'json', 'Contact email'),
				('tax_rate', '18', 'number', 'Default tax rate in percent')
				ON CONFLICT (key) DO NOTHING
			`); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS settings CASCADE"); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 012: Convert categories table to use UUID primary key
	m.AddMigration("012", "Convert categories table to use UUID primary key",
		func(ctx context.Context, tx pgx.Tx) error {
			// Enable UUID extension if not already enabled
			if _, err := tx.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`); err != nil {
				return err
			}

			// Create new categories table with UUID primary key
			if _, err := tx.Exec(ctx, `
				CREATE TABLE categories_new (
					uuid_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
					slug TEXT UNIQUE NOT NULL,
					name TEXT NOT NULL,
					description TEXT,
					parent_id UUID REFERENCES categories_new(uuid_id) ON DELETE SET NULL,
					sort_order INT NOT NULL DEFAULT 0,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Create indexes
			if _, err := tx.Exec(ctx, `CREATE INDEX idx_categories_new_parent_sort ON categories_new(parent_id, sort_order)`); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `CREATE INDEX idx_categories_new_slug ON categories_new(slug)`); err != nil {
				return err
			}

			// Drop old table if it exists and rename new one
			if _, err := tx.Exec(ctx, `DROP TABLE IF EXISTS categories CASCADE`); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `ALTER TABLE categories_new RENAME TO categories`); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			// Rollback: Drop categories table
			if _, err := tx.Exec(ctx, `DROP TABLE IF EXISTS categories CASCADE`); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 013: Add category_id to products table
	m.AddMigration("013", "Add category_id to products table",
		func(ctx context.Context, tx pgx.Tx) error {
			// Add category_id column to products table
			if _, err := tx.Exec(ctx, `ALTER TABLE products ADD COLUMN IF NOT EXISTS category_id UUID REFERENCES categories(uuid_id) ON DELETE SET NULL`); err != nil {
				return err
			}

			// Create index for performance
			if _, err := tx.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id)`); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			// Rollback: Remove category_id column and index
			if _, err := tx.Exec(ctx, `DROP INDEX IF EXISTS idx_products_category_id`); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `ALTER TABLE products DROP COLUMN IF EXISTS category_id`); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 014: Add image_id to categories table
	m.AddMigration("014", "Add image_id to categories table",
		func(ctx context.Context, tx pgx.Tx) error {
			// Add image_id column to categories table
			if _, err := tx.Exec(ctx, `ALTER TABLE categories ADD COLUMN IF NOT EXISTS image_id UUID REFERENCES media(uuid_id) ON DELETE SET NULL`); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			// Rollback: Remove image_id column
			if _, err := tx.Exec(ctx, `ALTER TABLE categories DROP COLUMN IF EXISTS image_id`); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 015: Create newsletter subscribers table
	m.AddMigration("015", "Create newsletter subscribers table",
		func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS newsletter_subscribers (
					id SERIAL PRIMARY KEY,
					email TEXT UNIQUE NOT NULL,
					subscribed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					is_active BOOLEAN NOT NULL DEFAULT TRUE,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Create indexes for performance
			if _, err := tx.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_newsletter_subscribers_email ON newsletter_subscribers(email)`); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_newsletter_subscribers_active ON newsletter_subscribers(is_active)`); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			// Rollback: Drop newsletter subscribers table
			if _, err := tx.Exec(ctx, `DROP TABLE IF EXISTS newsletter_subscribers CASCADE`); err != nil {
				return err
			}
			return nil
		},
	)

	// Migration 016: Create cart and wishlist tables
	m.AddMigration("016", "Create cart and wishlist tables",
		func(ctx context.Context, tx pgx.Tx) error {
			// Create cart table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS cart (
					id SERIAL PRIMARY KEY,
					session_id TEXT, -- For guest users
					user_id INTEGER REFERENCES users(id) ON DELETE CASCADE, -- For authenticated users
					product_id UUID REFERENCES products(uuid_id) ON DELETE CASCADE,
					variant_id INT REFERENCES product_variants(id) ON DELETE CASCADE,
					quantity INT NOT NULL DEFAULT 1,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					-- Ensure either session_id or user_id is present
					CONSTRAINT cart_user_or_session CHECK (
						(session_id IS NOT NULL AND user_id IS NULL) OR 
						(session_id IS NULL AND user_id IS NOT NULL) OR
						(session_id IS NULL AND user_id IS NULL) -- Allow empty for admin operations
					)
				)
			`); err != nil {
				return err
			}

			// Create wishlist table
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS wishlist (
					id SERIAL PRIMARY KEY,
					user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					product_id UUID NOT NULL REFERENCES products(uuid_id) ON DELETE CASCADE,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					UNIQUE(user_id, product_id)
				)
			`); err != nil {
				return err
			}

			// Create indexes for cart
			if _, err := tx.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_cart_session ON cart(session_id)`); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_cart_user ON cart(user_id)`); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_cart_product ON cart(product_id)`); err != nil {
				return err
			}

			// Create indexes for wishlist
			if _, err := tx.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_wishlist_user ON wishlist(user_id)`); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_wishlist_product ON wishlist(product_id)`); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			// Rollback: Drop cart and wishlist tables
			if _, err := tx.Exec(ctx, `DROP TABLE IF EXISTS cart CASCADE`); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `DROP TABLE IF EXISTS wishlist CASCADE`); err != nil {
				return err
			}
			return nil
		},
	)
}
