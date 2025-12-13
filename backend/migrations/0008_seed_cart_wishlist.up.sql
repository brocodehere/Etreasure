-- Migration: Seed cart and wishlist tables with sample data

-- Insert sample cart items for demo purposes
INSERT INTO cart (session_id, product_id, variant_id, quantity, created_at, updated_at) VALUES
('demo-session-1', 1, 1, 1, NOW(), NOW()),
('demo-session-1', 2, 2, 2, NOW(), NOW()),
('demo-session-2', 3, 3, 1, NOW(), NOW());

-- Insert sample wishlist items for demo customer (assuming customer with id=1)
INSERT INTO wishlist (customer_id, product_id, created_at) VALUES
(1, 4, NOW()),
(1, 5, NOW()),
(1, 1, NOW());

-- Note: These are sample records for testing.
-- In production, these tables will be populated by user actions through the frontend.
