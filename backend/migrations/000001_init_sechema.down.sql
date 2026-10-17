BEGIN;

-- Drop reviews & social
DROP TABLE IF EXISTS reviews CASCADE;
DROP TABLE IF EXISTS wishlist_items CASCADE;
DROP TABLE IF EXISTS user_notifications CASCADE;

-- Drop orders & payments
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS order_status_history CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;

-- Drop coupons & redemptions
DROP TABLE IF EXISTS coupon_redemptions CASCADE;
DROP TABLE IF EXISTS coupons CASCADE;

-- Drop inventory
DROP TABLE IF EXISTS inventory_reservations CASCADE;
DROP TABLE IF EXISTS inventory_ledgers CASCADE;

-- Drop carts
DROP TABLE IF EXISTS cart_items CASCADE;
DROP TABLE IF EXISTS carts CASCADE;

-- Drop catalog join & child tables
DROP TABLE IF EXISTS variant_media CASCADE;
DROP TABLE IF EXISTS product_tag_assignments CASCADE;
DROP TABLE IF EXISTS category_attributes CASCADE;

-- Drop catalog core tables
DROP TABLE IF EXISTS product_variants CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS product_categories CASCADE;
DROP TABLE IF EXISTS product_tags CASCADE;
DROP TABLE IF EXISTS product_brands CASCADE;

-- Drop users & addresses
DROP TABLE IF EXISTS addresses CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop storage / media
DROP TABLE IF EXISTS storage_objects CASCADE;

-- Drop admin settings
DROP TABLE IF EXISTS admin_settings CASCADE;

-- Drop custom ENUM types if present
DROP TYPE IF EXISTS storage_object_status CASCADE;
DROP TYPE IF EXISTS publication_status CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;
DROP TYPE IF EXISTS order_status CASCADE;
DROP TYPE IF EXISTS payment_status CASCADE;
DROP TYPE IF EXISTS review_status CASCADE;
DROP TYPE IF EXISTS discount_type CASCADE;
DROP TYPE IF EXISTS attribute_value_type CASCADE;
DROP TYPE IF EXISTS coupon_status CASCADE;
DROP TYPE IF EXISTS inventory_reason CASCADE;

COMMIT;
