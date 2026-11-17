CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_products_title_trgm ON products USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_product_brands_name_trgm ON product_brands USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_product_categories_name_trgm ON product_categories USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_product_tags_name_trgm ON product_tags USING gin (name gin_trgm_ops);
