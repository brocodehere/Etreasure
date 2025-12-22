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

	// Migration 011: Create content management tables (content pages and FAQs)
	m.AddMigration("011", "Create content management tables",
		func(ctx context.Context, tx pgx.Tx) error {
			// Enable UUID extension if not already enabled
			if _, err := tx.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`); err != nil {
				return err
			}

			// Create content_pages table for static pages (About, Policies, etc.)
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS content_pages (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					title TEXT NOT NULL,
					slug TEXT NOT NULL UNIQUE,
					content TEXT NOT NULL,
					type TEXT NOT NULL, -- 'about', 'policy', etc.
					is_active BOOLEAN NOT NULL DEFAULT TRUE,
					meta_title TEXT,
					meta_description TEXT,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Create faqs table for frequently asked questions
			if _, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS faqs (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					question TEXT NOT NULL,
					answer TEXT NOT NULL,
					category TEXT NOT NULL DEFAULT 'General',
					is_active BOOLEAN NOT NULL DEFAULT TRUE,
					sort_order INTEGER NOT NULL DEFAULT 0,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return err
			}

			// Create indexes for content_pages
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_content_pages_type ON content_pages(type)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_content_pages_is_active ON content_pages(is_active)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_content_pages_slug ON content_pages(slug)"); err != nil {
				return err
			}

			// Create indexes for faqs
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_faqs_category ON faqs(category)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_faqs_is_active ON faqs(is_active)"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_faqs_sort_order ON faqs(sort_order)"); err != nil {
				return err
			}

			// Insert default content pages
			if _, err := tx.Exec(ctx, `
				INSERT INTO content_pages (title, slug, content, type, meta_title, meta_description) VALUES
				(
					'Our Story',
					'our-story',
					'# Our Story

Ethnic Treasure was founded in 2015 with a simple mission: to preserve and celebrate India''s rich textile heritage while providing sustainable livelihoods to skilled artisans across the country.

## Our Journey

What began as a small passion project has grown into a movement that connects over 500+ artisans with appreciative customers worldwide. We work directly with weaving communities, ensuring fair wages and preserving traditional techniques that have been passed down through generations.

## Our Values

- **Authenticity**: Every piece in our collection is handcrafted using traditional techniques
- **Sustainability**: We use natural dyes and eco-friendly processes wherever possible
- **Fair Trade**: Artisans receive fair compensation for their craftsmanship
- **Preservation**: We actively work to preserve endangered craft forms

## The Artisans Behind Your Treasures

Each piece in our collection tells a story - of the artisan who created it, the tradition it represents, and the future it helps sustain. From Banarasi weavers to Kanchipuram silk experts, our partners are true custodians of India''s textile heritage.

Join us in our mission to keep these traditions alive, one beautiful creation at a time.',
					'about',
					'Our Story - Ethnic Treasure | Handcrafted Traditional Indian Clothing',
					'Learn about Ethnic Treasure''s journey since 2015, our commitment to preserving India''s textile heritage, and the 500+ artisans who create our handcrafted collections.'
				),
				(
					'Our Artisans',
					'artisans',
					'# Our Artisans

At Ethnic Treasure, we believe that behind every beautiful creation is a skilled artisan whose hands bring life to threads. Our artisan community is the heart and soul of everything we do.

## Meet Our Weaving Communities

### Banarasi Weavers
- Location: Varanasi, Uttar Pradesh
- Specialty: intricate brocade work and gold/silver thread weaving
- Generations of expertise: 8+ on average

### Kanchipuram Silk Experts
- Location: Kanchipuram, Tamil Nadu
- Specialty: pure silk sarees with temple borders
- Known for: durability and vibrant colors

### Chikankari Artisans
- Location: Lucknow, Uttar Pradesh
- Specialty: delicate white-on-white embroidery
- Technique: 32 different stitch types

## Empowering Artisans

We work directly with artisan communities, eliminating middlemen and ensuring:

- Fair wages that reflect true craftsmanship
- Safe working conditions
- Skills development programs
- Healthcare and education support
- Preservation of traditional techniques

## Artisan Stories

Every month, we feature stories from our artisan community, sharing their dreams, challenges, and the meaning behind their craft. These aren''t just products - they''re legacies being woven into the fabric of modern India.',
					'about',
					'Our Artisans - Ethnic Treasure | Meet Master Craftsmen',
					'Meet the skilled artisans behind Ethnic Treasure''s handcrafted collections. From Banarasi weavers to Chikankari experts, discover the stories and traditions.'
				),
				(
					'Crafts & Techniques',
					'crafts',
					'# Crafts & Techniques

India''s textile heritage spans thousands of years, with each region developing unique techniques that have been perfected over generations. At Ethnic Treasure, we celebrate and preserve these ancient crafts.

## Weaving Techniques

### Handloom Weaving
- **Process**: Traditional wooden looms operated entirely by hand
- **Time**: 2-4 weeks for a single saree
- **Regions**: Varanasi, Kanchipuram, Mysore, Pochampally

### Block Printing
- **Process**: Hand-carved wooden blocks dipped in natural dyes
- **Regions**: Jaipur, Bagru, Sanganer
- **Specialty**: Geometric patterns and floral motifs

### Tie & Dye (Bandhani)
- **Process**: Tiny dots tied before dyeing to create patterns
- **Regions**: Rajasthan, Gujarat
- **Meaning**: Each dot represents a blessing

## Embroidery Styles

### Chikankari
- **Origin**: Lucknow, 16th century
- **Technique**: 32 different stitch types
- **Character**: Delicate white-on-white work

### Zardozi
- **Technique**: Gold and silver thread embroidery
- **History**: Mughal era royal courts
- **Modern Use**: Bridal and festive wear

### Kantha
- **Origin**: Bengal
- **Technique**: Running stitch storytelling
- **Sustainability**: Upcycled fabric layers

## Dyeing Traditions

### Natural Dyes
- **Sources**: Plants, minerals, and insects
- **Benefits**: Eco-friendly and therapeutic
- **Colors**: Indigo, turmeric, madder, lac

### Bandhani Dyeing
- **Process**: Tie-resist dyeing
- **Patterns**: Dots, waves, and geometric designs
- **Cultural Significance**: Wedding and festival traditions

## Preserving Heritage

We actively work to preserve these techniques through:
- Documentation of traditional methods
- Training programs for younger generations
- Research and development of sustainable practices
- Fair trade partnerships with artisan communities

Each piece in our collection is a testament to these living traditions, carrying forward centuries of craftsmanship into the modern world.',
					'about',
					'Traditional Indian Crafts & Techniques | Ethnic Treasure Heritage',
					'Explore India''s rich textile heritage - handloom weaving, block printing, embroidery styles, and natural dyeing techniques preserved by Ethnic Treasure artisans.'
				)
				ON CONFLICT (slug) DO NOTHING
			`); err != nil {
				return err
			}

			// Insert default FAQs
			if _, err := tx.Exec(ctx, `
				INSERT INTO faqs (question, answer, category, sort_order) VALUES
				('What is your return policy?', 'We offer a 30-day return policy for all unused items in original packaging. Please contact our customer service team to initiate a return. Refunds are processed within 5-7 business days after we receive the returned item.', 'Returns', 1),
				('How long does shipping take?', 'Standard shipping takes 5-7 business days within India. Express shipping takes 2-3 business days. International shipping takes 10-15 business days. You can track your order using the tracking number provided after dispatch.', 'Shipping', 1),
				('Do you ship internationally?', 'Yes, we ship to over 50 countries worldwide. International shipping rates and delivery times vary by destination. Please check our shipping policy for detailed information.', 'Shipping', 2),
				('How do I know my size?', 'We provide detailed size charts for all our products. Each product page includes measurements in both inches and centimeters. If you''re unsure, our customer service team can help you find the perfect fit.', 'Products', 1),
				('Are your products authentic?', 'Absolutely! All our products are handcrafted by skilled artisans using traditional techniques. We provide authenticity certificates with our premium products and work directly with artisan communities.', 'Products', 2),
				('What payment methods do you accept?', 'We accept all major credit/debit cards, UPI, net banking, and cash on delivery (for select locations). All transactions are secured with industry-standard encryption.', 'Payments', 1),
				('How do I care for my ethnic wear?', 'Each product comes with specific care instructions. Generally, we recommend dry cleaning for silk and embroidered items, gentle hand washing for cotton, and avoiding direct sunlight to preserve colors.', 'Products', 3),
				('Do you offer customization?', 'Yes, we offer customization services for select products. Please contact our team at least 4 weeks before your required date to discuss your customization needs.', 'General', 1),
				('How can I track my order?', 'Once your order is dispatched, you''ll receive a tracking number via email. You can use this number to track your order on our website or the courier''s tracking portal.', 'Orders', 1),
				('What if I receive a damaged item?', 'We take utmost care in packaging, but if you receive a damaged item, please contact us within 48 hours with photos. We''ll arrange for a replacement or refund immediately.', 'Returns', 2)
				ON CONFLICT DO NOTHING
			`); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, tx pgx.Tx) error {
			// Drop tables in reverse dependency order
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS faqs CASCADE"); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "DROP TABLE IF EXISTS content_pages CASCADE"); err != nil {
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
			if _, err := tx.Exec(ctx, `ALTER TABLE categories ADD COLUMN IF NOT EXISTS image_id INTEGER REFERENCES media(id) ON DELETE SET NULL`); err != nil {
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
