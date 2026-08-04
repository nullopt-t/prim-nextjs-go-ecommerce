BEGIN;

-- Drop dependent / join / child tables first
DROP TABLE IF EXISTS variant_media CASCADE;
DROP TABLE IF EXISTS product_media CASCADE;
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
