-- Migration: Enhance orders table to store complete customer and product information
-- This migration adds customer details and improves order structure for shipping and admin panel

-- Add customer details to orders table
ALTER TABLE orders 
ADD COLUMN IF NOT EXISTS customer_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS customer_email VARCHAR(255),
ADD COLUMN IF NOT EXISTS customer_phone VARCHAR(20);

-- Add payment details
ALTER TABLE orders
ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50) DEFAULT 'razorpay',
ADD COLUMN IF NOT EXISTS payment_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS razorpay_order_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS razorpay_payment_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS razorpay_signature VARCHAR(255);

-- Update shipping_address to be more structured
ALTER TABLE orders 
ADD COLUMN IF NOT EXISTS shipping_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS shipping_email VARCHAR(255),
ADD COLUMN IF NOT EXISTS shipping_phone VARCHAR(20),
ADD COLUMN IF NOT EXISTS shipping_address_line1 TEXT,
ADD COLUMN IF NOT EXISTS shipping_address_line2 TEXT,
ADD COLUMN IF NOT EXISTS shipping_city VARCHAR(100),
ADD COLUMN IF NOT EXISTS shipping_state VARCHAR(100),
ADD COLUMN IF NOT EXISTS shipping_country VARCHAR(100) DEFAULT 'India',
ADD COLUMN IF NOT EXISTS shipping_pin_code VARCHAR(20);

-- Add billing address fields
ALTER TABLE orders
ADD COLUMN IF NOT EXISTS billing_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS billing_email VARCHAR(255),
ADD COLUMN IF NOT EXISTS billing_phone VARCHAR(20),
ADD COLUMN IF NOT EXISTS billing_address_line1 TEXT,
ADD COLUMN IF NOT EXISTS billing_address_line2 TEXT,
ADD COLUMN IF NOT EXISTS billing_city VARCHAR(100),
ADD COLUMN IF NOT EXISTS billing_state VARCHAR(100),
ADD COLUMN IF NOT EXISTS billing_country VARCHAR(100) DEFAULT 'India',
ADD COLUMN IF NOT EXISTS billing_pin_code VARCHAR(20);

-- Add order tracking fields
ALTER TABLE orders
ADD COLUMN IF NOT EXISTS tracking_number VARCHAR(255),
ADD COLUMN IF NOT EXISTS tracking_provider VARCHAR(100),
ADD COLUMN IF NOT EXISTS estimated_delivery DATE;

-- Create order line items table for product details
CREATE TABLE IF NOT EXISTS order_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(uuid_id),
    variant_id INTEGER REFERENCES product_variants(id),
    product_title VARCHAR(255) NOT NULL,
    product_sku VARCHAR(100),
    product_image_url TEXT,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL CHECK (unit_price >= 0),
    total_price DECIMAL(10,2) NOT NULL CHECK (total_price >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_order_line_items_order_id ON order_line_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_line_items_product_id ON order_line_items(product_id);

-- Create order status history table for tracking
CREATE TABLE IF NOT EXISTS order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    old_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    notes TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_status_history_order_id ON order_status_history(order_id);

-- Update existing orders to have default customer info (this will be updated when orders are processed)
UPDATE orders SET 
    customer_name = COALESCE(customer_name, 'Guest Customer'),
    customer_email = COALESCE(customer_email, 'guest@example.com'),
    customer_phone = COALESCE(customer_phone, '0000000000')
WHERE customer_name IS NULL OR customer_email IS NULL OR customer_phone IS NULL;

-- Add comments for documentation
COMMENT ON TABLE orders IS 'Main orders table with customer and shipping information';
COMMENT ON TABLE order_line_items IS 'Individual line items for each order containing product details';
COMMENT ON TABLE order_status_history IS 'History of status changes for orders';
