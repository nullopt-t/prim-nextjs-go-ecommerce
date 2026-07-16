DROP TABLE IF EXISTS product_tags;
DROP TABLE IF EXISTS product_categories;

DROP TABLE IF EXISTS inventories;

DROP INDEX IF EXISTS product_media_primary_variant_idx;
DROP INDEX IF EXISTS product_media_primary_product_idx;
DROP INDEX IF EXISTS product_media_variant_idx;
DROP INDEX IF EXISTS product_media_product_idx;
DROP TABLE IF EXISTS product_media;

DROP TABLE IF EXISTS product_variant_attributes;

DROP TABLE IF EXISTS attribute_values;
DROP TABLE IF EXISTS attributes;

ALTER TABLE products
DROP CONSTRAINT IF EXISTS products_default_variant_fk;

DROP INDEX IF EXISTS product_variants_currency_idx;
DROP INDEX IF EXISTS product_variants_price_idx;
DROP INDEX IF EXISTS product_variants_product_idx;
DROP INDEX IF EXISTS product_variants_sku_idx;
DROP TABLE IF EXISTS product_variants;

DROP INDEX IF EXISTS products_brand_idx;
DROP INDEX IF EXISTS products_status_idx;
DROP INDEX IF EXISTS products_title_idx;
DROP INDEX IF EXISTS products_slug_idx;
DROP TABLE IF EXISTS products;

DROP INDEX IF EXISTS tags_name_idx;
DROP TABLE IF EXISTS tags;

DROP INDEX IF EXISTS brands_slug_idx;
DROP INDEX IF EXISTS brands_name_idx;
DROP TABLE IF EXISTS brands;

DROP TYPE IF EXISTS tmedia_type;
DROP TYPE IF EXISTS tcurrency;
DROP TYPE IF EXISTS tproduct_status;

