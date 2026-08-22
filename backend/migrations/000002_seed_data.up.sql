-- Seed script for Prim E-Commerce Platform

-- Users (1 admin, 2 customers)
INSERT INTO users (id, email, full_name, role, is_email_verified, created_at, updated_at) VALUES
('10000000-0000-0000-0000-000000000001', 'admin@prim.com', 'Admin User', 'admin', true, now(), now()),
('10000000-0000-0000-0000-000000000002', 'john.doe@example.com', 'John Doe', 'customer', true, now(), now()),
('10000000-0000-0000-0000-000000000003', 'jane.smith@example.com', 'Jane Smith', 'customer', true, now(), now())
ON CONFLICT DO NOTHING;

-- Addresses
INSERT INTO addresses (id, user_id, label, full_name, phone, line1, line2, city, state, postal_code, country, is_default, created_at, updated_at) VALUES
('20000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 'Home', 'John Doe', '+1234567890', '123 Main St', 'Apt 4B', 'New York', 'NY', '10001', 'USA', true, now(), now()),
('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000003', 'Work', 'Jane Smith', '+1987654321', '456 Market St', 'Suite 100', 'San Francisco', 'CA', '94103', 'USA', true, now(), now())
ON CONFLICT DO NOTHING;

-- Brands
INSERT INTO product_brands (id, public_id, name, created_at, updated_at) VALUES
('30000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000002', 'Apple', now(), now()),
('30000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000004', 'Samsung', now(), now()),
('30000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000006', 'Nike', now(), now()),
('30000000-0000-0000-0000-000000000007', '30000000-0000-0000-0000-000000000008', 'Sony', now(), now())
ON CONFLICT DO NOTHING;

-- Categories
INSERT INTO product_categories (id, public_id, parent_id, name, created_at, updated_at) VALUES
('40000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000002', NULL, 'Electronics', now(), now()),
('40000000-0000-0000-0000-000000000003', '40000000-0000-0000-0000-000000000004', NULL, 'Apparel', now(), now()),
('40000000-0000-0000-0000-000000000005', '40000000-0000-0000-0000-000000000006', '40000000-0000-0000-0000-000000000001', 'Laptops', now(), now()),
('40000000-0000-0000-0000-000000000007', '40000000-0000-0000-0000-000000000008', '40000000-0000-0000-0000-000000000001', 'Smartphones', now(), now()),
('40000000-0000-0000-0000-000000000009', '40000000-0000-0000-0000-000000000010', '40000000-0000-0000-0000-000000000003', 'Shoes', now(), now())
ON CONFLICT DO NOTHING;

-- Tags
INSERT INTO product_tags (id, name, created_at, updated_at) VALUES
('50000000-0000-0000-0000-000000000001', 'Featured', now(), now()),
('50000000-0000-0000-0000-000000000003', 'Sale', now(), now()),
('50000000-0000-0000-0000-000000000005', 'New Arrival', now(), now())
ON CONFLICT DO NOTHING;

-- Storage Objects (Thumbnails)
INSERT INTO storage_objects (id, bucket, object_key, content_type, file_size, status, created_at, updated_at) VALUES
('80000000-0000-0000-0000-000000000001', 'products', 'macbook-pro.jpg', 'image/jpeg', 102400, 'uploaded', now(), now()),
('80000000-0000-0000-0000-000000000002', 'products', 's24-ultra.jpg', 'image/jpeg', 153600, 'uploaded', now(), now()),
('80000000-0000-0000-0000-000000000003', 'products', 'af1.jpg', 'image/jpeg', 204800, 'uploaded', now(), now()),
('80000000-0000-0000-0000-000000000004', 'products', 'sony-wh1000xm5.jpg', 'image/jpeg', 102400, 'uploaded', now(), now())
ON CONFLICT DO NOTHING;

-- Products
INSERT INTO products (id, brand_id, category_id, slug, title, description, highlights, status, product_type, thumbnail_object_id, created_at, updated_at) VALUES
('60000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000005', 'macbook-pro-16', 'MacBook Pro 16"', 'Supercharged by M3 Pro or M3 Max.', '["16-inch Liquid Retina XDR display", "Up to 22 hours battery life"]'::jsonb, 'published', 'variable', '80000000-0000-0000-0000-000000000001', now(), now()),
('60000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', '40000000-0000-0000-0000-000000000007', 'galaxy-s24-ultra', 'Galaxy S24 Ultra', 'Welcome to the era of mobile AI.', '["Titanium exterior", "Built-in S Pen", "200MP camera"]'::jsonb, 'published', 'variable', '80000000-0000-0000-0000-000000000002', now(), now()),
('60000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000005', '40000000-0000-0000-0000-000000000009', 'nike-air-force-1', 'Nike Air Force 1', 'Radiance lives on in the Nike Air Force 1.', '["Classic style", "Durable leather", "Air cushioning"]'::jsonb, 'published', 'variable', '80000000-0000-0000-0000-000000000003', now(), now()),
('60000000-0000-0000-0000-000000000007', '30000000-0000-0000-0000-000000000007', '40000000-0000-0000-0000-000000000001', 'sony-wh-1000xm5', 'Sony WH-1000XM5', 'Industry-leading noise canceling headphones.', '["Auto NC Optimizer", "Up to 30-hour battery"]'::jsonb, 'published', 'simple', '80000000-0000-0000-0000-000000000004', now(), now())
ON CONFLICT DO NOTHING;

-- Variants
INSERT INTO product_variants (id, sku, product_id, is_default, title, price, currency, attributes, created_at, updated_at) VALUES
-- MacBook Pro Variants
('70000000-0000-0000-0000-000000000001', 'MAC-PRO-16-BLK-18', '60000000-0000-0000-0000-000000000001', true, 'Space Black, 18GB RAM, 512GB SSD', 249900, 'USD', '{"color": "Space Black", "ram": "18GB", "storage": "512GB"}'::jsonb, now(), now()),
('70000000-0000-0000-0000-000000000003', 'MAC-PRO-16-SLV-36', '60000000-0000-0000-0000-000000000001', false, 'Silver, 36GB RAM, 1TB SSD', 309900, 'USD', '{"color": "Silver", "ram": "36GB", "storage": "1TB"}'::jsonb, now(), now()),
-- Galaxy S24 Variants
('70000000-0000-0000-0000-000000000005', 'S24-ULTRA-GRY-256', '60000000-0000-0000-0000-000000000003', true, 'Titanium Gray, 256GB', 129999, 'USD', '{"color": "Titanium Gray", "storage": "256GB"}'::jsonb, now(), now()),
('70000000-0000-0000-0000-000000000007', 'S24-ULTRA-BLK-512', '60000000-0000-0000-0000-000000000003', false, 'Titanium Black, 512GB', 141999, 'USD', '{"color": "Titanium Black", "storage": "512GB"}'::jsonb, now(), now()),
-- Nike Air Force 1 Variants
('70000000-0000-0000-0000-000000000009', 'AF1-WHT-100', '60000000-0000-0000-0000-000000000005', true, 'White, Size 10', 11500, 'USD', '{"color": "White", "size": "10"}'::jsonb, now(), now()),
('70000000-0000-0000-0000-000000000011', 'AF1-BLK-105', '60000000-0000-0000-0000-000000000005', false, 'Black, Size 10.5', 11500, 'USD', '{"color": "Black", "size": "10.5"}'::jsonb, now(), now()),
-- Sony WH-1000XM5
('70000000-0000-0000-0000-000000000013', 'SONY-WH1000XM5-SLV', '60000000-0000-0000-0000-000000000007', true, 'Silver', 39800, 'USD', '{"color": "Silver"}'::jsonb, now(), now())
ON CONFLICT DO NOTHING;

-- Tag Assignments
INSERT INTO product_tag_assignments (product_id, tag_id, created_at) VALUES
('60000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000001', now()), -- Mac -> Featured
('60000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000005', now()), -- Galaxy -> New Arrival
('60000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000001', now()), -- Nike -> Featured
('60000000-0000-0000-0000-000000000007', '50000000-0000-0000-0000-000000000003', now())  -- Sony -> Sale
ON CONFLICT DO NOTHING;

-- Inventory Ledgers (Add stock to each variant)
INSERT INTO inventory_ledgers (id, variant_id, quantity, reason, created_at) VALUES
('80000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000001', 50, 'restock', now()),
('80000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000003', 20, 'restock', now()),
('80000000-0000-0000-0000-000000000003', '70000000-0000-0000-0000-000000000005', 100, 'restock', now()),
('80000000-0000-0000-0000-000000000004', '70000000-0000-0000-0000-000000000007', 75, 'restock', now()),
('80000000-0000-0000-0000-000000000005', '70000000-0000-0000-0000-000000000009', 150, 'restock', now()),
('80000000-0000-0000-0000-000000000006', '70000000-0000-0000-0000-000000000011', 120, 'restock', now()),
('80000000-0000-0000-0000-000000000007', '70000000-0000-0000-0000-000000000013', 30, 'restock', now())
ON CONFLICT DO NOTHING;

-- Orders (Completed / Delivered for customer accounts)
INSERT INTO orders (id, customer_id, customer_email, shipping_address, billing_address, status, total_amount, currency, created_at, updated_at) VALUES
('90000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 'john.doe@example.com', '{"street": "123 Main St", "city": "New York", "postal_code": "10001", "country": "USA"}'::jsonb, '{"street": "123 Main St", "city": "New York", "postal_code": "10001", "country": "USA"}'::jsonb, 'delivered', 249900, 'USD', now() - INTERVAL '15 days', now() - INTERVAL '10 days'),
('90000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000003', 'jane.smith@example.com', '{"street": "456 Market St", "city": "San Francisco", "postal_code": "94103", "country": "USA"}'::jsonb, '{"street": "456 Market St", "city": "San Francisco", "postal_code": "94103", "country": "USA"}'::jsonb, 'delivered', 129999, 'USD', now() - INTERVAL '12 days', now() - INTERVAL '8 days'),
('90000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002', 'john.doe@example.com', '{"street": "123 Main St", "city": "New York", "postal_code": "10001", "country": "USA"}'::jsonb, '{"street": "123 Main St", "city": "New York", "postal_code": "10001", "country": "USA"}'::jsonb, 'delivered', 11500, 'USD', now() - INTERVAL '7 days', now() - INTERVAL '5 days'),
('90000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000003', 'jane.smith@example.com', '{"street": "456 Market St", "city": "San Francisco", "postal_code": "94103", "country": "USA"}'::jsonb, '{"street": "456 Market St", "city": "San Francisco", "postal_code": "94103", "country": "USA"}'::jsonb, 'delivered', 39800, 'USD', now() - INTERVAL '5 days', now() - INTERVAL '3 days'),
('90000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000002', 'john.doe@example.com', '{"street": "123 Main St", "city": "New York", "postal_code": "10001", "country": "USA"}'::jsonb, '{"street": "123 Main St", "city": "New York", "postal_code": "10001", "country": "USA"}'::jsonb, 'delivered', 39800, 'USD', now() - INTERVAL '4 days', now() - INTERVAL '2 days')
ON CONFLICT DO NOTHING;

-- Order Items
INSERT INTO order_items (id, order_id, variant_id, quantity, price_at_purchase, product_snapshot) VALUES
('91000000-0000-0000-0000-000000000001', '90000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000001', 1, 249900, '{"title": "MacBook Pro 16\""}'::jsonb),
('91000000-0000-0000-0000-000000000002', '90000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000005', 1, 129999, '{"title": "Galaxy S24 Ultra"}'::jsonb),
('91000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000003', '70000000-0000-0000-0000-000000000009', 1, 11500, '{"title": "Nike Air Force 1"}'::jsonb),
('91000000-0000-0000-0000-000000000004', '90000000-0000-0000-0000-000000000004', '70000000-0000-0000-0000-000000000013', 1, 39800, '{"title": "Sony WH-1000XM5"}'::jsonb),
('91000000-0000-0000-0000-000000000005', '90000000-0000-0000-0000-000000000005', '70000000-0000-0000-0000-000000000013', 1, 39800, '{"title": "Sony WH-1000XM5"}'::jsonb)
ON CONFLICT DO NOTHING;

-- Reviews (Approved customer reviews with ratings)
INSERT INTO reviews (id, product_id, user_id, order_item_id, rating, title, body, status, created_at, updated_at) VALUES
('92000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', '91000000-0000-0000-0000-000000000001', 5, 'Unbelievable performance!', 'The M3 Pro chip handles heavy 4K rendering and compiles code effortlessly. Battery lasts all day long.', 'approved', now() - INTERVAL '9 days', now() - INTERVAL '9 days'),
('92000000-0000-0000-0000-000000000002', '60000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000003', '91000000-0000-0000-0000-000000000002', 5, 'Best smartphone screen on the market', 'The anti-reflective titanium display is gorgeous outdoors and the camera zoom is crazy good.', 'approved', now() - INTERVAL '7 days', now() - INTERVAL '7 days'),
('92000000-0000-0000-0000-000000000003', '60000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000002', '91000000-0000-0000-0000-000000000003', 4, 'Classic and timeless sneakers', 'Fits true to size and looks great with almost anything. Takes a couple of days to break in.', 'approved', now() - INTERVAL '4 days', now() - INTERVAL '4 days'),
('92000000-0000-0000-0000-000000000004', '60000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000003', '91000000-0000-0000-0000-000000000004', 5, 'Exceptional noise cancellation', 'The ANC on these Sony headphones is completely next level. Comfortable ear cushions for long flights.', 'approved', now() - INTERVAL '2 days', now() - INTERVAL '2 days'),
('92000000-0000-0000-0000-000000000005', '60000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000002', '91000000-0000-0000-0000-000000000005', 4, 'Great sound, slightly bulky case', 'Audio quality and battery life are fantastic. The only minor gripe is the carrying case size.', 'approved', now() - INTERVAL '1 day', now() - INTERVAL '1 day')
ON CONFLICT DO NOTHING;
