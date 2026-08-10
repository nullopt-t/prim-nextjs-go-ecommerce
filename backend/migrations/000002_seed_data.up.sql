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
INSERT INTO product_tags (id, public_id, name, created_at, updated_at) VALUES
('50000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000002', 'Featured', now(), now()),
('50000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000004', 'Sale', now(), now()),
('50000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000006', 'New Arrival', now(), now())
ON CONFLICT DO NOTHING;

-- Products
INSERT INTO products (id, brand_id, category_id, public_id, title, description, highlights, status, has_variant_list, created_at, updated_at) VALUES
('60000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000005', '60000000-0000-0000-0000-000000000002', 'MacBook Pro 16"', 'Supercharged by M3 Pro or M3 Max.', '["16-inch Liquid Retina XDR display", "Up to 22 hours battery life"]'::jsonb, 'published', true, now(), now()),
('60000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', '40000000-0000-0000-0000-000000000007', '60000000-0000-0000-0000-000000000004', 'Galaxy S24 Ultra', 'Welcome to the era of mobile AI.', '["Titanium exterior", "Built-in S Pen", "200MP camera"]'::jsonb, 'published', true, now(), now()),
('60000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000005', '40000000-0000-0000-0000-000000000009', '60000000-0000-0000-0000-000000000006', 'Nike Air Force 1', 'The radiance lives on in the Nike Air Force 1.', '["Classic style", "Crisp leather", "All-day comfort"]'::jsonb, 'published', true, now(), now()),
('60000000-0000-0000-0000-000000000007', '30000000-0000-0000-0000-000000000007', '40000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000008', 'Sony WH-1000XM5', 'Industry-leading noise canceling headphones.', '["30-hour battery life", "Multipoint connection"]'::jsonb, 'published', false, now(), now())
ON CONFLICT DO NOTHING;

-- Variants
INSERT INTO product_variants (id, public_id, product_id, is_default, title, price, currency, attributes, created_at, updated_at) VALUES
-- MacBook Pro Variants
('70000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000002', '60000000-0000-0000-0000-000000000001', true, 'Space Black, 18GB RAM, 512GB SSD', 249900, 'USD', '{"color": "Space Black", "ram": "18GB", "storage": "512GB"}'::jsonb, now(), now()),
('70000000-0000-0000-0000-000000000003', '70000000-0000-0000-0000-000000000004', '60000000-0000-0000-0000-000000000001', false, 'Silver, 36GB RAM, 1TB SSD', 309900, 'USD', '{"color": "Silver", "ram": "36GB", "storage": "1TB"}'::jsonb, now(), now()),
-- Galaxy S24 Variants
('70000000-0000-0000-0000-000000000005', '70000000-0000-0000-0000-000000000006', '60000000-0000-0000-0000-000000000003', true, 'Titanium Gray, 256GB', 129999, 'USD', '{"color": "Titanium Gray", "storage": "256GB"}'::jsonb, now(), now()),
('70000000-0000-0000-0000-000000000007', '70000000-0000-0000-0000-000000000008', '60000000-0000-0000-0000-000000000003', false, 'Titanium Black, 512GB', 141999, 'USD', '{"color": "Titanium Black", "storage": "512GB"}'::jsonb, now(), now()),
-- Nike Air Force 1 Variants
('70000000-0000-0000-0000-000000000009', '70000000-0000-0000-0000-000000000010', '60000000-0000-0000-0000-000000000005', true, 'White, Size 10', 11500, 'USD', '{"color": "White", "size": "10"}'::jsonb, now(), now()),
('70000000-0000-0000-0000-000000000011', '70000000-0000-0000-0000-000000000012', '60000000-0000-0000-0000-000000000005', false, 'Black, Size 10.5', 11500, 'USD', '{"color": "Black", "size": "10.5"}'::jsonb, now(), now()),
-- Sony WH-1000XM5
('70000000-0000-0000-0000-000000000013', '70000000-0000-0000-0000-000000000014', '60000000-0000-0000-0000-000000000007', true, 'Silver', 39800, 'USD', '{"color": "Silver"}'::jsonb, now(), now())
ON CONFLICT DO NOTHING;

-- Tag Assignments
INSERT INTO product_tag_assignments (product_id, tag_id, created_at) VALUES
('60000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000001', now()), -- Mac -> Featured
('60000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000005', now()), -- Galaxy -> New Arrival
('60000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000001', now()), -- Nike -> Featured
('60000000-0000-0000-0000-000000000007', '50000000-0000-0000-0000-000000000003', now())  -- Sony -> Sale
ON CONFLICT DO NOTHING;

-- Inventory Ledgers (Add stock to each variant)
INSERT INTO inventory_ledgers (id, variant_id, quantity, reason, reference_id, created_at) VALUES
('80000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000001', 50, 'restock', 'initial-seed', now()),
('80000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000003', 20, 'restock', 'initial-seed', now()),
('80000000-0000-0000-0000-000000000003', '70000000-0000-0000-0000-000000000005', 100, 'restock', 'initial-seed', now()),
('80000000-0000-0000-0000-000000000004', '70000000-0000-0000-0000-000000000007', 75, 'restock', 'initial-seed', now()),
('80000000-0000-0000-0000-000000000005', '70000000-0000-0000-0000-000000000009', 150, 'restock', 'initial-seed', now()),
('80000000-0000-0000-0000-000000000006', '70000000-0000-0000-0000-000000000011', 120, 'restock', 'initial-seed', now()),
('80000000-0000-0000-0000-000000000007', '70000000-0000-0000-0000-000000000013', 30, 'restock', 'initial-seed', now())
ON CONFLICT DO NOTHING;
