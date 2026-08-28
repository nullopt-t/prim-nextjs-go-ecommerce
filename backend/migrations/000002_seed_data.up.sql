-- =====================================================================
-- PRIM E-COMMERCE COMPREHENSIVE SEED DATA
-- PostgreSQL
-- =====================================================================

BEGIN;

-- =====================================================================
-- 1. STORAGE OBJECTS (Media & Images)
-- =====================================================================
INSERT INTO storage_objects (id, bucket, object_key, content_type, file_size, status, created_at, updated_at) VALUES
('80000000-0000-0000-0000-000000000001', 'products', 'macbook-pro-16.jpg', 'image/jpeg', 245000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000002', 'products', 'galaxy-s24-ultra.jpg', 'image/jpeg', 185000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000003', 'products', 'nike-air-force-1.jpg', 'image/jpeg', 210000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000004', 'products', 'sony-wh1000xm5.jpg', 'image/jpeg', 142000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000005', 'brands', 'apple-logo.png', 'image/png', 45000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000006', 'brands', 'samsung-logo.png', 'image/png', 38000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000007', 'brands', 'nike-logo.png', 'image/png', 32000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000008', 'brands', 'sony-logo.png', 'image/png', 29000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000009', 'brands', 'adidas-logo.png', 'image/png', 35000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000010', 'brands', 'dell-logo.png', 'image/png', 41000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000011', 'brands', 'logitech-logo.png', 'image/png', 28000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000012', 'brands', 'bose-logo.png', 'image/png', 31000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000013', 'products', 'dell-xps-15.jpg', 'image/jpeg', 220000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000014', 'products', 'ipad-pro-m4.jpg', 'image/jpeg', 195000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000015', 'products', 'adidas-ultraboost.jpg', 'image/jpeg', 230000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000016', 'products', 'logitech-mx-master-3s.jpg', 'image/jpeg', 125000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000017', 'products', 'bose-qc-ultra.jpg', 'image/jpeg', 160000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000018', 'products', 'nike-tech-fleece.jpg', 'image/jpeg', 280000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000019', 'products', 'apple-watch-ultra-2.jpg', 'image/jpeg', 175000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000020', 'products', 'samsung-galaxy-tab-s9.jpg', 'image/jpeg', 210000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000021', 'products', 'sony-playstation-5.jpg', 'image/jpeg', 310000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000022', 'products', 'adidas-track-jacket.jpg', 'image/jpeg', 190000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000023', 'products', 'logitech-g-pro-x-superlight.jpg', 'image/jpeg', 115000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('80000000-0000-0000-0000-000000000024', 'products', 'macbook-air-m3.jpg', 'image/jpeg', 198000, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 2. USERS (Admins, Support, Customers)
-- =====================================================================
INSERT INTO users (id, email, full_name, role, is_email_verified, status, created_at, updated_at) VALUES
('10000000-0000-0000-0000-000000000001', 'admin@prim.com', 'Admin User', 'admin', true, 'active', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('10000000-0000-0000-0000-000000000002', 'john.doe@example.com', 'John Doe', 'customer', true, 'active', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('10000000-0000-0000-0000-000000000003', 'jane.smith@example.com', 'Jane Smith', 'customer', true, 'active', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('10000000-0000-0000-0000-000000000004', 'support@prim.com', 'Support Specialist', 'support', true, 'active', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('10000000-0000-0000-0000-000000000005', 'alex.turner@example.com', 'Alex Turner', 'customer', true, 'active', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('10000000-0000-0000-0000-000000000006', 'sarah.connor@example.com', 'Sarah Connor', 'customer', true, 'active', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('10000000-0000-0000-0000-000000000007', 'michael.scott@example.com', 'Michael Scott', 'customer', true, 'active', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('10000000-0000-0000-0000-000000000008', 'elena.rostova@example.com', 'Elena Rostova', 'customer', true, 'active', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('10000000-0000-0000-0000-000000000009', 'david.beck@example.com', 'David Beck', 'customer', true, 'active', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('10000000-0000-0000-0000-000000000010', 'emily.watson@example.com', 'Emily Watson', 'customer', true, 'active', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('10000000-0000-0000-0000-000000000011', 'carlos.mendez@example.com', 'Carlos Mendez', 'customer', true, 'active', now() - INTERVAL '60 days', now() - INTERVAL '60 days'),
('10000000-0000-0000-0000-000000000012', 'hannah.abbott@example.com', 'Hannah Abbott', 'customer', false, 'active', now() - INTERVAL '60 days', now() - INTERVAL '60 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 3. ADDRESSES
-- =====================================================================
INSERT INTO addresses (id, user_id, label, full_name, phone, line1, line2, city, state, postal_code, country, is_default, created_at, updated_at) VALUES
('20000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 'Home', 'John Doe', '+1234567890', '123 Main St', 'Apt 4B', 'New York', 'NY', '10001', 'USA', true, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000003', 'Work', 'Jane Smith', '+1987654321', '456 Market St', 'Suite 100', 'San Francisco', 'CA', '94103', 'USA', true, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('20000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000005', 'Home', 'Alex Turner', '+14155551234', '789 Sunset Blvd', 'Apt 12', 'Los Angeles', 'CA', '90028', 'USA', true, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('20000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000006', 'Home', 'Sarah Connor', '+15125556789', '101 Cyberdyne Way', NULL, 'Austin', 'TX', '78701', 'USA', true, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('20000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000007', 'Office', 'Michael Scott', '+15705559876', '1725 Slough Ave', 'Suite 200', 'Scranton', 'PA', '18503', 'USA', true, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('20000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000008', 'Home', 'Elena Rostova', '+442071838750', '221B Baker St', 'Flat 2', 'London', 'Greater London', 'NW1 6XE', 'GBR', true, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('20000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000009', 'Home', 'David Beck', '+4930123456', 'Friedrichstraße 43', NULL, 'Berlin', 'Berlin', '10117', 'DEU', true, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('20000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000010', 'Home', 'Emily Watson', '+12065554321', '500 Pine St', 'Floor 4', 'Seattle', 'WA', '98101', 'USA', true, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('20000000-0000-0000-0000-000000000011', '10000000-0000-0000-0000-000000000011', 'Apartment', 'Carlos Mendez', '+13055557890', '800 Brickell Ave', 'PH 3', 'Miami', 'FL', '33131', 'USA', true, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('20000000-0000-0000-0000-000000000012', '10000000-0000-0000-0000-000000000002', 'Beach House', 'John Doe', '+1234567890', '50 Ocean Drive', NULL, 'Miami Beach', 'FL', '33139', 'USA', false, now() - INTERVAL '50 days', now() - INTERVAL '50 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 4. BRANDS
-- =====================================================================
INSERT INTO product_brands (id, public_id, name, link, logo_object_id, created_at, updated_at) VALUES
('30000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000002', 'Apple', 'https://apple.com', '80000000-0000-0000-0000-000000000005', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('30000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000004', 'Samsung', 'https://samsung.com', '80000000-0000-0000-0000-000000000006', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('30000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000006', 'Nike', 'https://nike.com', '80000000-0000-0000-0000-000000000007', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('30000000-0000-0000-0000-000000000007', '30000000-0000-0000-0000-000000000008', 'Sony', 'https://sony.com', '80000000-0000-0000-0000-000000000008', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('30000000-0000-0000-0000-000000000009', '30000000-0000-0000-0000-000000000010', 'Adidas', 'https://adidas.com', '80000000-0000-0000-0000-000000000009', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('30000000-0000-0000-0000-000000000011', '30000000-0000-0000-0000-000000000012', 'Dell', 'https://dell.com', '80000000-0000-0000-0000-000000000010', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('30000000-0000-0000-0000-000000000013', '30000000-0000-0000-0000-000000000014', 'Logitech', 'https://logitech.com', '80000000-0000-0000-0000-000000000011', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('30000000-0000-0000-0000-000000000015', '30000000-0000-0000-0000-000000000016', 'Bose', 'https://bose.com', '80000000-0000-0000-0000-000000000012', now() - INTERVAL '50 days', now() - INTERVAL '50 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 5. TAGS
-- =====================================================================
INSERT INTO product_tags (id, name, created_at, updated_at) VALUES
('50000000-0000-0000-0000-000000000001', 'Featured', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('50000000-0000-0000-0000-000000000002', 'Sale', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('50000000-0000-0000-0000-000000000003', 'New Arrival', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('50000000-0000-0000-0000-000000000004', 'Best Seller', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('50000000-0000-0000-0000-000000000005', 'Trending', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('50000000-0000-0000-0000-000000000006', 'Limited Edition', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('50000000-0000-0000-0000-000000000007', 'Eco-Friendly', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('50000000-0000-0000-0000-000000000008', 'Editor''s Choice', now() - INTERVAL '50 days', now() - INTERVAL '50 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 6. CATEGORIES
-- =====================================================================
INSERT INTO product_categories (id, public_id, parent_id, name, created_at, updated_at) VALUES
('40000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000002', NULL, 'Electronics', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('40000000-0000-0000-0000-000000000003', '40000000-0000-0000-0000-000000000004', NULL, 'Apparel', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('40000000-0000-0000-0000-000000000005', '40000000-0000-0000-0000-000000000006', '40000000-0000-0000-0000-000000000001', 'Laptops', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('40000000-0000-0000-0000-000000000007', '40000000-0000-0000-0000-000000000008', '40000000-0000-0000-0000-000000000001', 'Smartphones', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('40000000-0000-0000-0000-000000000009', '40000000-0000-0000-0000-000000000010', '40000000-0000-0000-0000-000000000003', 'Shoes', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('40000000-0000-0000-0000-000000000011', '40000000-0000-0000-0000-000000000012', '40000000-0000-0000-0000-000000000001', 'Audio & Headphones', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('40000000-0000-0000-0000-000000000013', '40000000-0000-0000-0000-000000000014', '40000000-0000-0000-0000-000000000001', 'Tablets & Wearables', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('40000000-0000-0000-0000-000000000015', '40000000-0000-0000-0000-000000000016', '40000000-0000-0000-0000-000000000003', 'Outerwear', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('40000000-0000-0000-0000-000000000017', '40000000-0000-0000-0000-000000000018', '40000000-0000-0000-0000-000000000001', 'Computer Accessories', now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('40000000-0000-0000-0000-000000000019', '40000000-0000-0000-0000-000000000020', '40000000-0000-0000-0000-000000000001', 'Gaming Consoles', now() - INTERVAL '50 days', now() - INTERVAL '50 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 7. CATEGORY ATTRIBUTES
-- =====================================================================
INSERT INTO category_attributes (id, category_id, key, label, value_type, allowed_values, is_filterable, sort_order, created_at, updated_at) VALUES
('41000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000005', 'ram', 'RAM Capacity', 'enum', '["8GB", "16GB", "18GB", "32GB", "36GB", "48GB", "64GB"]'::jsonb, true, 1, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('41000000-0000-0000-0000-000000000002', '40000000-0000-0000-0000-000000000005', 'storage', 'Storage Size', 'enum', '["256GB", "512GB", "1TB", "2TB", "4TB"]'::jsonb, true, 2, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('41000000-0000-0000-0000-000000000003', '40000000-0000-0000-0000-000000000007', 'storage', 'Internal Storage', 'enum', '["128GB", "256GB", "512GB", "1TB"]'::jsonb, true, 1, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('41000000-0000-0000-0000-000000000004', '40000000-0000-0000-0000-000000000007', 'color', 'Color Finish', 'enum', '["Titanium Gray", "Titanium Black", "Titanium Violet", "Titanium Yellow"]'::jsonb, true, 2, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('41000000-0000-0000-0000-000000000005', '40000000-0000-0000-0000-000000000009', 'size', 'Shoe Size (US)', 'enum', '["8", "8.5", "9", "9.5", "10", "10.5", "11", "11.5", "12"]'::jsonb, true, 1, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('41000000-0000-0000-0000-000000000006', '40000000-0000-0000-0000-000000000009', 'color', 'Colorway', 'enum', '["White", "Black", "Core Black", "Cloud White", "Blue"]'::jsonb, true, 2, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('41000000-0000-0000-0000-000000000007', '40000000-0000-0000-0000-000000000011', 'wireless', 'Wireless / Bluetooth', 'boolean', NULL, true, 1, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('41000000-0000-0000-0000-000000000008', '40000000-0000-0000-0000-000000000011', 'anc', 'Active Noise Cancelling', 'boolean', NULL, true, 2, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('41000000-0000-0000-0000-000000000009', '40000000-0000-0000-0000-000000000015', 'size', 'Apparel Size', 'enum', '["XS", "S", "M", "L", "XL", "XXL"]'::jsonb, true, 1, now() - INTERVAL '50 days', now() - INTERVAL '50 days'),
('41000000-0000-0000-0000-000000000010', '40000000-0000-0000-0000-000000000017', 'connectivity', 'Connectivity Type', 'enum', '["Bluetooth", "2.4GHz Wireless", "Wired USB-C"]'::jsonb, true, 1, now() - INTERVAL '50 days', now() - INTERVAL '50 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 8. PRODUCTS
-- =====================================================================
INSERT INTO products (id, brand_id, category_id, slug, title, description, highlights, status, product_type, thumbnail_object_id, created_at, updated_at) VALUES
('60000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000005', 'macbook-pro-16', 'MacBook Pro 16"', 'Supercharged by M3 Pro or M3 Max chip. Extreme dynamic range and long battery life.', '["16-inch Liquid Retina XDR display", "Up to 22 hours battery life", "M3 Pro / M3 Max silicon"]'::jsonb, 'published', 'variable', '80000000-0000-0000-0000-000000000001', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000005', 'macbook-air-15-m3', 'MacBook Air 15" M3', 'Lean, mean, M3 machine. Strikingly thin and fast.', '["15.3-inch Liquid Retina display", "MagSafe charging", "Fanless quiet design"]'::jsonb, 'published', 'variable', '80000000-0000-0000-0000-000000000024', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', '40000000-0000-0000-0000-000000000007', 'galaxy-s24-ultra', 'Galaxy S24 Ultra', 'Welcome to the era of mobile AI with Galaxy AI.', '["Titanium frame", "Built-in S Pen", "200MP Quad Telephoto camera"]'::jsonb, 'published', 'variable', '80000000-0000-0000-0000-000000000002', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000011', '40000000-0000-0000-0000-000000000005', 'dell-xps-15-oled', 'Dell XPS 15 OLED', 'High-performance creator laptop with stunning 3.5K OLED touchscreen display.', '["15.6-inch 3.5K OLED", "Intel Core i9 13th Gen", "NVIDIA RTX 4070"]'::jsonb, 'published', 'variable', '80000000-0000-0000-0000-000000000013', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000005', '40000000-0000-0000-0000-000000000009', 'nike-air-force-1', 'Nike Air Force 1 ''07', 'Radiance lives on in the Nike Air Force 1 ''07, the basketball icon.', '["Stitched leather overlays", "Nike Air cushioning", "Padded low-cut collar"]'::jsonb, 'published', 'variable', '80000000-0000-0000-0000-000000000003', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000006', '30000000-0000-0000-0000-000000000009', '40000000-0000-0000-0000-000000000009', 'adidas-ultraboost-light', 'Ultraboost Light Shoes', 'Experience epic energy with the lightest Ultraboost ever made.', '["Light BOOST midsole", "Primeknit+ textile upper", "Continental Rubber outsole"]'::jsonb, 'published', 'variable', '80000000-0000-0000-0000-000000000015', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000007', '30000000-0000-0000-0000-000000000007', '40000000-0000-0000-0000-000000000011', 'sony-wh-1000xm5', 'Sony WH-1000XM5 Wireless Headphones', 'Industry-leading noise canceling with two processors and 8 microphones.', '["Auto NC Optimizer", "Up to 30-hour battery", "Crystal clear hands-free calling"]'::jsonb, 'published', 'simple', '80000000-0000-0000-0000-000000000004', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000008', '30000000-0000-0000-0000-000000000015', '40000000-0000-0000-0000-000000000011', 'bose-quietcomfort-ultra', 'Bose QuietComfort Ultra Headphones', 'Breakthrough spatial audio and world-class quiet tailored just for you.', '["CustomTune technology", "Immersive Audio mode", "24 hours battery life"]'::jsonb, 'published', 'simple', '80000000-0000-0000-0000-000000000017', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000009', '30000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000013', 'apple-watch-ultra-2', 'Apple Watch Ultra 2', 'The most rugged and capable Apple Watch. Designed for outdoor adventure.', '["49mm Titanium case", "Precision dual-frequency GPS", "Up to 72h low power mode"]'::jsonb, 'published', 'variable', '80000000-0000-0000-0000-000000000019', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000010', '30000000-0000-0000-0000-000000000013', '40000000-0000-0000-0000-000000000017', 'logitech-mx-master-3s', 'Logitech MX Master 3S Wireless Mouse', 'An iconic mouse remastered. Feel every moment of your workflow with Quiet Clicks.', '["8K DPI any-surface tracking", "MagSpeed electromagnetic scrolling", "Quiet click buttons"]'::jsonb, 'published', 'variable', '80000000-0000-0000-0000-000000000016', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000011', '30000000-0000-0000-0000-000000000005', '40000000-0000-0000-0000-000000000015', 'nike-tech-fleece-windrunner', 'Nike Tech Fleece Windrunner', 'Premium lightweight fleece that delivers lightweight warmth without bulk.', '["Smooth on both sides fleece", "Full-zip design", "Zippered sleeve pocket"]'::jsonb, 'published', 'variable', '80000000-0000-0000-0000-000000000018', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000012', '30000000-0000-0000-0000-000000000007', '40000000-0000-0000-0000-000000000019', 'playstation-5-slim', 'PlayStation 5 Slim Console', 'Play has no limits with lightning fast loading via ultra-high speed SSD.', '["1TB NVMe SSD", "Tempest 3D AudioTech", "Ray Tracing Support"]'::jsonb, 'published', 'simple', '80000000-0000-0000-0000-000000000021', now() - INTERVAL '40 days', now() - INTERVAL '40 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 9. PRODUCT VARIANTS
-- =====================================================================
INSERT INTO product_variants (id, sku, product_id, is_default, title, price, crossed_out_price, currency, attributes, thumbnail_object_id, created_at, updated_at) VALUES
('70000000-0000-0000-0000-000000000001', 'MAC-PRO-16-BLK-18', '60000000-0000-0000-0000-000000000001', true, 'Space Black, 18GB RAM, 512GB SSD', 249900, 269900, 'USD', '{"color": "Space Black", "ram": "18GB", "storage": "512GB"}'::jsonb, '80000000-0000-0000-0000-000000000001', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000002', 'MAC-PRO-16-SLV-36', '60000000-0000-0000-0000-000000000001', false, 'Silver, 36GB RAM, 1TB SSD', 309900, 329900, 'USD', '{"color": "Silver", "ram": "36GB", "storage": "1TB"}'::jsonb, '80000000-0000-0000-0000-000000000001', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000003', 'MAC-PRO-16-BLK-48', '60000000-0000-0000-0000-000000000001', false, 'Space Black, 48GB RAM, 1TB SSD', 349900, 369900, 'USD', '{"color": "Space Black", "ram": "48GB", "storage": "1TB"}'::jsonb, '80000000-0000-0000-0000-000000000001', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000004', 'MAC-AIR-15-MID-8', '60000000-0000-0000-0000-000000000002', true, 'Midnight, 8GB RAM, 256GB SSD', 129900, 139900, 'USD', '{"color": "Midnight", "ram": "8GB", "storage": "256GB"}'::jsonb, '80000000-0000-0000-0000-000000000024', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000005', 'MAC-AIR-15-STL-16', '60000000-0000-0000-0000-000000000002', false, 'Starlight, 16GB RAM, 512GB SSD', 169900, 179900, 'USD', '{"color": "Starlight", "ram": "16GB", "storage": "512GB"}'::jsonb, '80000000-0000-0000-0000-000000000024', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000006', 'S24-ULTRA-GRY-256', '60000000-0000-0000-0000-000000000003', true, 'Titanium Gray, 256GB', 129999, 139999, 'USD', '{"color": "Titanium Gray", "storage": "256GB"}'::jsonb, '80000000-0000-0000-0000-000000000002', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000007', 'S24-ULTRA-BLK-512', '60000000-0000-0000-0000-000000000003', false, 'Titanium Black, 512GB', 141999, 151999, 'USD', '{"color": "Titanium Black", "storage": "512GB"}'::jsonb, '80000000-0000-0000-0000-000000000002', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000008', 'S24-ULTRA-VLT-1TB', '60000000-0000-0000-0000-000000000003', false, 'Titanium Violet, 1TB', 165999, 175999, 'USD', '{"color": "Titanium Violet", "storage": "1TB"}'::jsonb, '80000000-0000-0000-0000-000000000002', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000009', 'DELL-XPS15-I7-16', '60000000-0000-0000-0000-000000000004', true, 'Intel i7, 16GB RAM, 512GB SSD', 189900, 209900, 'USD', '{"processor": "Intel i7", "ram": "16GB", "storage": "512GB"}'::jsonb, '80000000-0000-0000-0000-000000000013', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000010', 'DELL-XPS15-I9-32', '60000000-0000-0000-0000-000000000004', false, 'Intel i9, 32GB RAM, 1TB SSD', 249900, 269900, 'USD', '{"processor": "Intel i9", "ram": "32GB", "storage": "1TB"}'::jsonb, '80000000-0000-0000-0000-000000000013', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000011', 'AF1-WHT-100', '60000000-0000-0000-0000-000000000005', true, 'White / White, Size 10', 11500, NULL, 'USD', '{"color": "White", "size": "10"}'::jsonb, '80000000-0000-0000-0000-000000000003', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000012', 'AF1-BLK-105', '60000000-0000-0000-0000-000000000005', false, 'Black / Black, Size 10.5', 11500, NULL, 'USD', '{"color": "Black", "size": "10.5"}'::jsonb, '80000000-0000-0000-0000-000000000003', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000013', 'AF1-WHT-095', '60000000-0000-0000-0000-000000000005', false, 'White / White, Size 9.5', 11500, NULL, 'USD', '{"color": "White", "size": "9.5"}'::jsonb, '80000000-0000-0000-0000-000000000003', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000014', 'UB-LGT-CORE-10', '60000000-0000-0000-0000-000000000006', true, 'Core Black / Cloud White, Size 10', 19000, 21000, 'USD', '{"color": "Core Black", "size": "10"}'::jsonb, '80000000-0000-0000-0000-000000000015', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000015', 'UB-LGT-WHT-105', '60000000-0000-0000-0000-000000000006', false, 'Cloud White / Crystal White, Size 10.5', 19000, 21000, 'USD', '{"color": "Cloud White", "size": "10.5"}'::jsonb, '80000000-0000-0000-0000-000000000015', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000016', 'SONY-WH1000XM5-SLV', '60000000-0000-0000-0000-000000000007', true, 'Silver', 39800, 42000, 'USD', '{"color": "Silver"}'::jsonb, '80000000-0000-0000-0000-000000000004', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000017', 'SONY-WH1000XM5-BLK', '60000000-0000-0000-0000-000000000007', false, 'Midnight Black', 39800, 42000, 'USD', '{"color": "Midnight Black"}'::jsonb, '80000000-0000-0000-0000-000000000004', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000018', 'BOSE-QCU-BLK', '60000000-0000-0000-0000-000000000008', true, 'Black', 42900, 44900, 'USD', '{"color": "Black"}'::jsonb, '80000000-0000-0000-0000-000000000017', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000019', 'BOSE-QCU-WHT', '60000000-0000-0000-0000-000000000008', false, 'White Smoke', 42900, 44900, 'USD', '{"color": "White Smoke"}'::jsonb, '80000000-0000-0000-0000-000000000017', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000020', 'AW-ULTRA2-ALP-M', '60000000-0000-0000-0000-000000000009', true, 'Titanium Case with Orange Alpine Loop (Medium)', 79900, NULL, 'USD', '{"band": "Alpine Loop", "size": "Medium"}'::jsonb, '80000000-0000-0000-0000-000000000019', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000021', 'AW-ULTRA2-OCN-L', '60000000-0000-0000-0000-000000000009', false, 'Titanium Case with Blue Ocean Band', 79900, NULL, 'USD', '{"band": "Ocean Band", "size": "One Size"}'::jsonb, '80000000-0000-0000-0000-000000000019', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000022', 'LOGI-MXM3S-GRF', '60000000-0000-0000-0000-000000000010', true, 'Graphite', 9999, 10999, 'USD', '{"color": "Graphite"}'::jsonb, '80000000-0000-0000-0000-000000000016', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000023', 'LOGI-MXM3S-PLG', '60000000-0000-0000-0000-000000000010', false, 'Pale Gray', 9999, 10999, 'USD', '{"color": "Pale Gray"}'::jsonb, '80000000-0000-0000-0000-000000000016', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000024', 'NIKE-TF-BLK-M', '60000000-0000-0000-0000-000000000011', true, 'Black / Medium', 14500, NULL, 'USD', '{"color": "Black", "size": "M"}'::jsonb, '80000000-0000-0000-0000-000000000018', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000025', 'NIKE-TF-GRY-L', '60000000-0000-0000-0000-000000000011', false, 'Dark Heather Grey / Large', 14500, NULL, 'USD', '{"color": "Dark Heather Grey", "size": "L"}'::jsonb, '80000000-0000-0000-0000-000000000018', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('70000000-0000-0000-0000-000000000026', 'SONY-PS5-SLIM-DISC', '60000000-0000-0000-0000-000000000012', true, 'Standard Disc Edition (White)', 49999, NULL, 'USD', '{"edition": "Disc Edition", "color": "White"}'::jsonb, '80000000-0000-0000-0000-000000000021', now() - INTERVAL '40 days', now() - INTERVAL '40 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 10. PRODUCT TAG ASSIGNMENTS
-- =====================================================================
INSERT INTO product_tag_assignments (product_id, tag_id, created_at) VALUES
('60000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000001', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000004', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000003', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000001', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000005', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000004', '50000000-0000-0000-0000-000000000008', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000004', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000001', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000005', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000007', '50000000-0000-0000-0000-000000000002', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000007', '50000000-0000-0000-0000-000000000008', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000008', '50000000-0000-0000-0000-000000000003', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000009', '50000000-0000-0000-0000-000000000006', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000010', '50000000-0000-0000-0000-000000000004', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000011', '50000000-0000-0000-0000-000000000005', now() - INTERVAL '40 days'),
('60000000-0000-0000-0000-000000000012', '50000000-0000-0000-0000-000000000001', now() - INTERVAL '40 days')
ON CONFLICT (product_id, tag_id) DO NOTHING;

-- =====================================================================
-- 11. VARIANT MEDIA
-- =====================================================================
INSERT INTO variant_media (id, public_id, variant_id, object_id, media_type, sort_order) VALUES
('71000000-0000-0000-0000-000000000001', '71000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000001', 'image', 0),
('71000000-0000-0000-0000-000000000003', '71000000-0000-0000-0000-000000000004', '70000000-0000-0000-0000-000000000004', '80000000-0000-0000-0000-000000000024', 'image', 0),
('71000000-0000-0000-0000-000000000005', '71000000-0000-0000-0000-000000000006', '70000000-0000-0000-0000-000000000006', '80000000-0000-0000-0000-000000000002', 'image', 0),
('71000000-0000-0000-0000-000000000007', '71000000-0000-0000-0000-000000000008', '70000000-0000-0000-0000-000000000009', '80000000-0000-0000-0000-000000000013', 'image', 0),
('71000000-0000-0000-0000-000000000009', '71000000-0000-0000-0000-000000000010', '70000000-0000-0000-0000-000000000011', '80000000-0000-0000-0000-000000000003', 'image', 0),
('71000000-0000-0000-0000-000000000011', '71000000-0000-0000-0000-000000000012', '70000000-0000-0000-0000-000000000014', '80000000-0000-0000-0000-000000000015', 'image', 0),
('71000000-0000-0000-0000-000000000013', '71000000-0000-0000-0000-000000000014', '70000000-0000-0000-0000-000000000016', '80000000-0000-0000-0000-000000000004', 'image', 0),
('71000000-0000-0000-0000-000000000015', '71000000-0000-0000-0000-000000000016', '70000000-0000-0000-0000-000000000018', '80000000-0000-0000-0000-000000000017', 'image', 0),
('71000000-0000-0000-0000-000000000017', '71000000-0000-0000-0000-000000000018', '70000000-0000-0000-0000-000000000020', '80000000-0000-0000-0000-000000000019', 'image', 0),
('71000000-0000-0000-0000-000000000019', '71000000-0000-0000-0000-000000000020', '70000000-0000-0000-0000-000000000022', '80000000-0000-0000-0000-000000000016', 'image', 0),
('71000000-0000-0000-0000-000000000021', '71000000-0000-0000-0000-000000000022', '70000000-0000-0000-0000-000000000024', '80000000-0000-0000-0000-000000000018', 'image', 0),
('71000000-0000-0000-0000-000000000023', '71000000-0000-0000-0000-000000000024', '70000000-0000-0000-0000-000000000026', '80000000-0000-0000-0000-000000000021', 'image', 0)
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 12. INVENTORY LEDGERS
-- =====================================================================
INSERT INTO inventory_ledgers (id, variant_id, quantity, reason, created_at) VALUES
('81000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000001', 50, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000002', 30, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000003', '70000000-0000-0000-0000-000000000003', 20, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000004', '70000000-0000-0000-0000-000000000004', 80, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000005', '70000000-0000-0000-0000-000000000005', 60, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000006', '70000000-0000-0000-0000-000000000006', 100, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000007', '70000000-0000-0000-0000-000000000007', 75, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000008', '70000000-0000-0000-0000-000000000008', 40, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000009', '70000000-0000-0000-0000-000000000009', 45, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000010', '70000000-0000-0000-0000-000000000010', 30, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000011', '70000000-0000-0000-0000-000000000011', 150, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000012', '70000000-0000-0000-0000-000000000012', 120, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000013', '70000000-0000-0000-0000-000000000013', 90, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000014', '70000000-0000-0000-0000-000000000014', 110, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000015', '70000000-0000-0000-0000-000000000015', 85, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000016', '70000000-0000-0000-0000-000000000016', 95, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000017', '70000000-0000-0000-0000-000000000017', 60, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000018', '70000000-0000-0000-0000-000000000018', 70, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000019', '70000000-0000-0000-0000-000000000019', 50, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000020', '70000000-0000-0000-0000-000000000020', 40, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000021', '70000000-0000-0000-0000-000000000021', 35, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000022', '70000000-0000-0000-0000-000000000022', 200, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000023', '70000000-0000-0000-0000-000000000023', 150, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000024', '70000000-0000-0000-0000-000000000024', 80, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000025', '70000000-0000-0000-0000-000000000025', 65, 'restock', now() - INTERVAL '35 days'),
('81000000-0000-0000-0000-000000000026', '70000000-0000-0000-0000-000000000026', 45, 'restock', now() - INTERVAL '35 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 13. COUPONS
-- =====================================================================
INSERT INTO coupons (id, code, discount_type, discount_value, currency, min_cart_amount, usage_limit, per_user_limit, starts_at, expires_at, status, created_at, updated_at) VALUES
('c0000000-0000-0000-0000-000000000001', 'WELCOME10', 'percentage', 10, 'USD', 5000, 1000, 1, now() - INTERVAL '40 days', now() + INTERVAL '60 days', 'active', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('c0000000-0000-0000-0000-000000000002', 'SUMMER20', 'percentage', 20, 'USD', 10000, 500, 1, now() - INTERVAL '40 days', now() + INTERVAL '60 days', 'active', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('c0000000-0000-0000-0000-000000000003', 'FLASH50', 'fixed_amount', 5000, 'USD', 20000, 200, 1, now() - INTERVAL '40 days', now() + INTERVAL '60 days', 'active', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('c0000000-0000-0000-0000-000000000004', 'VIP100', 'fixed_amount', 10000, 'USD', 50000, 100, 2, now() - INTERVAL '40 days', now() + INTERVAL '60 days', 'active', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('c0000000-0000-0000-0000-000000000005', 'EXPIRED30', 'percentage', 30, 'USD', 10000, 100, 1, now() - INTERVAL '40 days', now() - INTERVAL '5 days', 'archived', now() - INTERVAL '40 days', now() - INTERVAL '40 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 14. CARTS
-- =====================================================================
INSERT INTO carts (id, user_id, session_id, created_at, updated_at) VALUES
('a0000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', NULL, now() - INTERVAL '5 days', now() - INTERVAL '1 hour'),
('a0000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000003', NULL, now() - INTERVAL '5 days', now() - INTERVAL '1 hour'),
('a0000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000005', NULL, now() - INTERVAL '5 days', now() - INTERVAL '1 hour'),
('a0000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000006', NULL, now() - INTERVAL '5 days', now() - INTERVAL '1 hour'),
('a0000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000007', NULL, now() - INTERVAL '5 days', now() - INTERVAL '1 hour'),
('a0000000-0000-0000-0000-000000000006', NULL, 'sess_anon_987654321_guest_a', now() - INTERVAL '5 days', now() - INTERVAL '1 hour'),
('a0000000-0000-0000-0000-000000000007', NULL, 'sess_anon_123456789_guest_b', now() - INTERVAL '5 days', now() - INTERVAL '1 hour')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 15. CART ITEMS
-- =====================================================================
INSERT INTO cart_items (id, cart_id, variant_id, quantity, price_at_purchase, currency, carted_at) VALUES
('a1000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000004', 1, 129900, 'USD', now() - INTERVAL '2 days'),
('a1000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000022', 1, 9999, 'USD', now() - INTERVAL '2 days'),
('a1000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000018', 1, 42900, 'USD', now() - INTERVAL '2 days'),
('a1000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000003', '70000000-0000-0000-0000-000000000020', 1, 79900, 'USD', now() - INTERVAL '2 days'),
('a1000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000004', '70000000-0000-0000-0000-000000000024', 2, 14500, 'USD', now() - INTERVAL '2 days'),
('a1000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000005', '70000000-0000-0000-0000-000000000026', 1, 49999, 'USD', now() - INTERVAL '2 days'),
('a1000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000006', '70000000-0000-0000-0000-000000000014', 1, 19000, 'USD', now() - INTERVAL '2 days'),
('a1000000-0000-0000-0000-000000000008', 'a0000000-0000-0000-0000-000000000007', '70000000-0000-0000-0000-000000000007', 1, 141999, 'USD', now() - INTERVAL '2 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 16. INVENTORY RESERVATIONS
-- =====================================================================
INSERT INTO inventory_reservations (id, variant_id, cart_id, quantity, expires_at, created_at, released_at) VALUES
('82000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001', 1, now() + INTERVAL '1 hour', now() - INTERVAL '10 minutes', NULL),
('82000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000022', 'a0000000-0000-0000-0000-000000000001', 1, now() + INTERVAL '1 hour', now() - INTERVAL '10 minutes', NULL),
('82000000-0000-0000-0000-000000000003', '70000000-0000-0000-0000-000000000018', 'a0000000-0000-0000-0000-000000000002', 1, now() + INTERVAL '45 minutes', now() - INTERVAL '15 minutes', NULL),
('82000000-0000-0000-0000-000000000004', '70000000-0000-0000-0000-000000000026', 'a0000000-0000-0000-0000-000000000005', 1, now() - INTERVAL '1 day', now() - INTERVAL '2 days', now() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 17. ORDERS
-- =====================================================================
INSERT INTO orders (id, customer_id, customer_email, shipping_address, billing_address, status, coupon_id, discount_amount, total_amount, currency, created_at, updated_at) VALUES
('90000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 'john.doe@example.com', '{"name": "John Doe", "line1": "123 Main St", "city": "New York", "postal_code": "10001", "country": "USA"}'::jsonb, '{"name": "John Doe", "line1": "123 Main St", "city": "New York", "postal_code": "10001", "country": "USA"}'::jsonb, 'delivered', NULL, 0, 249900, 'USD', now() - INTERVAL '20 days', now() - INTERVAL '15 days'),
('90000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000003', 'jane.smith@example.com', '{"name": "Jane Smith", "line1": "456 Market St", "city": "San Francisco", "postal_code": "94103", "country": "USA"}'::jsonb, '{"name": "Jane Smith", "line1": "456 Market St", "city": "San Francisco", "postal_code": "94103", "country": "USA"}'::jsonb, 'delivered', 'c0000000-0000-0000-0000-000000000001', 13000, 116999, 'USD', now() - INTERVAL '18 days', now() - INTERVAL '13 days'),
('90000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000005', 'alex.turner@example.com', '{"name": "Alex Turner", "line1": "789 Sunset Blvd", "city": "Los Angeles", "postal_code": "90028", "country": "USA"}'::jsonb, '{"name": "Alex Turner", "line1": "789 Sunset Blvd", "city": "Los Angeles", "postal_code": "90028", "country": "USA"}'::jsonb, 'delivered', NULL, 0, 39800, 'USD', now() - INTERVAL '15 days', now() - INTERVAL '10 days'),
('90000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000006', 'sarah.connor@example.com', '{"name": "Sarah Connor", "line1": "101 Cyberdyne Way", "city": "Austin", "postal_code": "78701", "country": "USA"}'::jsonb, '{"name": "Sarah Connor", "line1": "101 Cyberdyne Way", "city": "Austin", "postal_code": "78701", "country": "USA"}'::jsonb, 'delivered', 'c0000000-0000-0000-0000-000000000003', 5000, 44999, 'USD', now() - INTERVAL '12 days', now() - INTERVAL '8 days'),
('90000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000007', 'michael.scott@example.com', '{"name": "Michael Scott", "line1": "1725 Slough Ave", "city": "Scranton", "postal_code": "18503", "country": "USA"}'::jsonb, '{"name": "Michael Scott", "line1": "1725 Slough Ave", "city": "Scranton", "postal_code": "18503", "country": "USA"}'::jsonb, 'shipped', NULL, 0, 19000, 'USD', now() - INTERVAL '4 days', now() - INTERVAL '1 day'),
('90000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000008', 'elena.rostova@example.com', '{"name": "Elena Rostova", "line1": "221B Baker St", "city": "London", "postal_code": "NW1 6XE", "country": "GBR"}'::jsonb, '{"name": "Elena Rostova", "line1": "221B Baker St", "city": "London", "postal_code": "NW1 6XE", "country": "GBR"}'::jsonb, 'processing', NULL, 0, 79900, 'USD', now() - INTERVAL '2 days', now() - INTERVAL '1 day'),
('90000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000009', 'david.beck@example.com', '{"name": "David Beck", "line1": "Friedrichstra\u00dfe 43", "city": "Berlin", "postal_code": "10117", "country": "DEU"}'::jsonb, '{"name": "David Beck", "line1": "Friedrichstra\u00dfe 43", "city": "Berlin", "postal_code": "10117", "country": "DEU"}'::jsonb, 'paid', NULL, 0, 189900, 'USD', now() - INTERVAL '1 day', now() - INTERVAL '1 day'),
('90000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000010', 'emily.watson@example.com', '{"name": "Emily Watson", "line1": "500 Pine St", "city": "Seattle", "postal_code": "98101", "country": "USA"}'::jsonb, '{"name": "Emily Watson", "line1": "500 Pine St", "city": "Seattle", "postal_code": "98101", "country": "USA"}'::jsonb, 'pending', NULL, 0, 14500, 'USD', now() - INTERVAL '12 hours', now() - INTERVAL '12 hours'),
('90000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000011', 'carlos.mendez@example.com', '{"name": "Carlos Mendez", "line1": "800 Brickell Ave", "city": "Miami", "postal_code": "33131", "country": "USA"}'::jsonb, '{"name": "Carlos Mendez", "line1": "800 Brickell Ave", "city": "Miami", "postal_code": "33131", "country": "USA"}'::jsonb, 'canceled', NULL, 0, 11500, 'USD', now() - INTERVAL '6 days', now() - INTERVAL '5 days'),
('90000000-0000-0000-0000-000000000010', NULL, 'guest.shopper@example.com', '{"name": "Guest Shopper", "line1": "777 Lucky St", "city": "Las Vegas", "postal_code": "89101", "country": "USA"}'::jsonb, '{"name": "Guest Shopper", "line1": "777 Lucky St", "city": "Las Vegas", "postal_code": "89101", "country": "USA"}'::jsonb, 'delivered', NULL, 0, 11500, 'USD', now() - INTERVAL '8 days', now() - INTERVAL '4 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 18. ORDER ITEMS
-- =====================================================================
INSERT INTO order_items (id, order_id, variant_id, quantity, price_at_purchase, product_snapshot) VALUES
('91000000-0000-0000-0000-000000000001', '90000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000001', 1, 249900, '{"title": "MacBook Pro 16\"", "sku": "MAC-PRO-16-BLK-18", "price": 249900}'::jsonb),
('91000000-0000-0000-0000-000000000002', '90000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000006', 1, 129999, '{"title": "Galaxy S24 Ultra", "sku": "S24-ULTRA-GRY-256", "price": 129999}'::jsonb),
('91000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000003', '70000000-0000-0000-0000-000000000016', 1, 39800, '{"title": "Sony WH-1000XM5", "sku": "SONY-WH1000XM5-SLV", "price": 39800}'::jsonb),
('91000000-0000-0000-0000-000000000004', '90000000-0000-0000-0000-000000000004', '70000000-0000-0000-0000-000000000026', 1, 49999, '{"title": "PlayStation 5 Slim Console", "sku": "SONY-PS5-SLIM-DISC", "price": 49999}'::jsonb),
('91000000-0000-0000-0000-000000000005', '90000000-0000-0000-0000-000000000005', '70000000-0000-0000-0000-000000000014', 1, 19000, '{"title": "Ultraboost Light Shoes", "sku": "UB-LGT-CORE-10", "price": 19000}'::jsonb),
('91000000-0000-0000-0000-000000000006', '90000000-0000-0000-0000-000000000006', '70000000-0000-0000-0000-000000000020', 1, 79900, '{"title": "Apple Watch Ultra 2", "sku": "AW-ULTRA2-ALP-M", "price": 79900}'::jsonb),
('91000000-0000-0000-0000-000000000007', '90000000-0000-0000-0000-000000000007', '70000000-0000-0000-0000-000000000009', 1, 189900, '{"title": "Dell XPS 15 OLED", "sku": "DELL-XPS15-I7-16", "price": 189900}'::jsonb),
('91000000-0000-0000-0000-000000000008', '90000000-0000-0000-0000-000000000008', '70000000-0000-0000-0000-000000000024', 1, 14500, '{"title": "Nike Tech Fleece Windrunner", "sku": "NIKE-TF-BLK-M", "price": 14500}'::jsonb),
('91000000-0000-0000-0000-000000000009', '90000000-0000-0000-0000-000000000009', '70000000-0000-0000-0000-000000000011', 1, 11500, '{"title": "Nike Air Force 1 ''07", "sku": "AF1-WHT-100", "price": 11500}'::jsonb),
('91000000-0000-0000-0000-000000000010', '90000000-0000-0000-0000-000000000010', '70000000-0000-0000-0000-000000000011', 1, 11500, '{"title": "Nike Air Force 1 ''07", "sku": "AF1-WHT-100", "price": 11500}'::jsonb)
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 19. ORDER STATUS HISTORY
-- =====================================================================
INSERT INTO order_status_history (id, order_id, status, notes, created_at) VALUES
('93000000-0000-0000-0000-000000000001', '90000000-0000-0000-0000-000000000001', 'pending', 'Order placed by customer', now() - INTERVAL '20 days'),
('93000000-0000-0000-0000-000000000002', '90000000-0000-0000-0000-000000000001', 'paid', 'Payment authorized and captured via Stripe', now() - INTERVAL '20 days' + INTERVAL '5 minutes'),
('93000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000001', 'processing', 'Fulfillment started at Warehouse East', now() - INTERVAL '19 days'),
('93000000-0000-0000-0000-000000000004', '90000000-0000-0000-0000-000000000001', 'shipped', 'Shipped via FedEx Express #FX982341', now() - INTERVAL '17 days'),
('93000000-0000-0000-0000-000000000005', '90000000-0000-0000-0000-000000000001', 'delivered', 'Package delivered to front door', now() - INTERVAL '15 days'),
('93000000-0000-0000-0000-000000000006', '90000000-0000-0000-0000-000000000002', 'pending', 'Order created with WELCOME10 coupon applied', now() - INTERVAL '18 days'),
('93000000-0000-0000-0000-000000000007', '90000000-0000-0000-0000-000000000002', 'paid', 'Payment successful via Stripe', now() - INTERVAL '18 days' + INTERVAL '3 minutes'),
('93000000-0000-0000-0000-000000000008', '90000000-0000-0000-0000-000000000002', 'processing', 'Packed at San Francisco Fulfillment Center', now() - INTERVAL '16 days'),
('93000000-0000-0000-0000-000000000009', '90000000-0000-0000-0000-000000000002', 'shipped', 'In transit via UPS Ground', now() - INTERVAL '15 days'),
('93000000-0000-0000-0000-000000000010', '90000000-0000-0000-0000-000000000002', 'delivered', 'Delivered and signed by recipient', now() - INTERVAL '13 days'),
('93000000-0000-0000-0000-000000000011', '90000000-0000-0000-0000-000000000005', 'pending', 'Order received', now() - INTERVAL '4 days'),
('93000000-0000-0000-0000-000000000012', '90000000-0000-0000-0000-000000000005', 'paid', 'Card charged successfully', now() - INTERVAL '4 days' + INTERVAL '2 minutes'),
('93000000-0000-0000-0000-000000000013', '90000000-0000-0000-0000-000000000005', 'shipped', 'Shipped via DHL Express', now() - INTERVAL '1 day'),
('93000000-0000-0000-0000-000000000014', '90000000-0000-0000-0000-000000000009', 'pending', 'Order initiated', now() - INTERVAL '6 days'),
('93000000-0000-0000-0000-000000000015', '90000000-0000-0000-0000-000000000009', 'canceled', 'Customer requested cancellation before shipment', now() - INTERVAL '5 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 20. COUPON REDEMPTIONS
-- =====================================================================
INSERT INTO coupon_redemptions (id, coupon_id, user_id, order_id, discount_amount, created_at) VALUES
('c1000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000002', 13000, now() - INTERVAL '18 days'),
('c1000000-0000-0000-0000-000000000002', 'c0000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000006', '90000000-0000-0000-0000-000000000004', 5000, now() - INTERVAL '12 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 21. PAYMENTS
-- =====================================================================
INSERT INTO payments (id, order_id, amount, currency, status, provider, provider_transaction_id, error_message, created_at, updated_at) VALUES
('f0000000-0000-0000-0000-000000000001', '90000000-0000-0000-0000-000000000001', 249900, 'USD', 'captured', 'stripe', 'ch_stripe_live_001_macbook', NULL, now() - INTERVAL '20 days', now() - INTERVAL '20 days'),
('f0000000-0000-0000-0000-000000000002', '90000000-0000-0000-0000-000000000002', 116999, 'USD', 'captured', 'stripe', 'ch_stripe_live_002_galaxy', NULL, now() - INTERVAL '18 days', now() - INTERVAL '18 days'),
('f0000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000003', 39800, 'USD', 'captured', 'stripe', 'ch_stripe_live_003_sony', NULL, now() - INTERVAL '15 days', now() - INTERVAL '15 days'),
('f0000000-0000-0000-0000-000000000004', '90000000-0000-0000-0000-000000000004', 44999, 'USD', 'captured', 'stripe', 'ch_stripe_live_004_ps5', NULL, now() - INTERVAL '12 days', now() - INTERVAL '12 days'),
('f0000000-0000-0000-0000-000000000005', '90000000-0000-0000-0000-000000000005', 19000, 'USD', 'captured', 'stripe', 'ch_stripe_live_005_ub', NULL, now() - INTERVAL '4 days', now() - INTERVAL '4 days'),
('f0000000-0000-0000-0000-000000000006', '90000000-0000-0000-0000-000000000006', 79900, 'USD', 'captured', 'stripe', 'ch_stripe_live_006_applewatch', NULL, now() - INTERVAL '2 days', now() - INTERVAL '2 days'),
('f0000000-0000-0000-0000-000000000007', '90000000-0000-0000-0000-000000000007', 189900, 'USD', 'authorized', 'stripe', 'ch_stripe_live_007_dellxps', NULL, now() - INTERVAL '1 day', now() - INTERVAL '1 day'),
('f0000000-0000-0000-0000-000000000008', '90000000-0000-0000-0000-000000000009', 11500, 'USD', 'refunded', 'stripe', 'ch_stripe_live_008_af1_canceled', NULL, now() - INTERVAL '6 days', now() - INTERVAL '6 days'),
('f0000000-0000-0000-0000-000000000009', '90000000-0000-0000-0000-000000000010', 11500, 'USD', 'captured', 'paypal', 'PAYID-NV54678129-GUEST', NULL, now() - INTERVAL '8 days', now() - INTERVAL '8 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 22. REVIEWS
-- =====================================================================
INSERT INTO reviews (id, product_id, user_id, order_item_id, rating, title, body, status, created_at, updated_at) VALUES
('92000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', '91000000-0000-0000-0000-000000000001', 5, 'Unbelievable performance and battery life!', 'The M3 Pro chip handles heavy video exports and Docker containers without breaking a sweat. Liquid Retina XDR screen is glorious.', 'approved', now() - INTERVAL '14 days', now() - INTERVAL '14 days'),
('92000000-0000-0000-0000-000000000002', '60000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000003', '91000000-0000-0000-0000-000000000002', 5, 'Best screen and zoom camera on any phone', 'The anti-reflective glass makes outdoor reading a breeze. AI photo editing and live transcription features work like magic.', 'approved', now() - INTERVAL '11 days', now() - INTERVAL '11 days'),
('92000000-0000-0000-0000-000000000003', '60000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000005', '91000000-0000-0000-0000-000000000003', 5, 'Superb noise cancellation for travel', 'Wore these on a 10-hour flight and the engine roar completely disappeared. Extremely comfortable memory foam pads.', 'approved', now() - INTERVAL '9 days', now() - INTERVAL '9 days'),
('92000000-0000-0000-0000-000000000004', '60000000-0000-0000-0000-000000000012', '10000000-0000-0000-0000-000000000006', '91000000-0000-0000-0000-000000000004', 5, 'Slimmer form factor, top tier gaming', 'Fast NVMe loading times and the DualSense haptics make games so immersive. Quiet fan operation even under load.', 'approved', now() - INTERVAL '7 days', now() - INTERVAL '7 days')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 23. WISHLIST ITEMS
-- =====================================================================
INSERT INTO wishlist_items (id, user_id, product_id, created_at) VALUES
('e0000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', '60000000-0000-0000-0000-000000000009', now() - INTERVAL '15 days'),
('e0000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', '60000000-0000-0000-0000-000000000012', now() - INTERVAL '15 days'),
('e0000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000003', '60000000-0000-0000-0000-000000000001', now() - INTERVAL '15 days'),
('e0000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000003', '60000000-0000-0000-0000-000000000008', now() - INTERVAL '15 days'),
('e0000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000005', '60000000-0000-0000-0000-000000000004', now() - INTERVAL '15 days'),
('e0000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000006', '60000000-0000-0000-0000-000000000006', now() - INTERVAL '15 days'),
('e0000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000007', '60000000-0000-0000-0000-000000000010', now() - INTERVAL '15 days'),
('e0000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000008', '60000000-0000-0000-0000-000000000002', now() - INTERVAL '15 days'),
('e0000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000009', '60000000-0000-0000-0000-000000000011', now() - INTERVAL '15 days'),
('e0000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000010', '60000000-0000-0000-0000-000000000003', now() - INTERVAL '15 days')
ON CONFLICT (id) DO NOTHING;

COMMIT;
