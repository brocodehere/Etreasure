-- Create stock notifications table
CREATE TABLE IF NOT EXISTS stock_notifications (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_slug VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    mobile_number VARCHAR(20),
    notification_type VARCHAR(10) NOT NULL CHECK (notification_type IN ('email', 'mobile')),
    is_active BOOLEAN DEFAULT TRUE,
    is_notified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_stock_notifications_product_id ON stock_notifications(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_notifications_product_slug ON stock_notifications(product_slug);
CREATE INDEX IF NOT EXISTS idx_stock_notifications_email ON stock_notifications(email);
CREATE INDEX IF NOT EXISTS idx_stock_notifications_active ON stock_notifications(is_active, is_notified);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_stock_notifications_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER stock_notifications_updated_at_trigger
    BEFORE UPDATE ON stock_notifications
    FOR EACH ROW
    EXECUTE FUNCTION update_stock_notifications_updated_at();
