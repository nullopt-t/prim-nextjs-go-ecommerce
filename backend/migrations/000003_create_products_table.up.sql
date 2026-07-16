-- ============================================================
-- ENUMS
-- ============================================================

DROP TYPE IF EXISTS tproduct_status;
CREATE TYPE tproduct_status AS ENUM (
    'draft',
    'active',
    'archived'
);

DROP TYPE IF EXISTS tcurrency;
CREATE TYPE tcurrency AS ENUM (
    'USD',
    'EUR',
    'EGP'
);

DROP TYPE IF EXISTS tmedia_type;
CREATE TYPE tmedia_type AS ENUM (
    'image',
    'video',
    'document'
);

-- ============================================================
-- BRANDS
-- ============================================================

CREATE TABLE IF NOT EXISTS brands (
    id UUID PRIMARY KEY,

    name TEXT NOT NULL,
    slug TEXT NOT NULL,

    logo_url TEXT NOT NULL,
    logo_label TEXT NOT NULL,

    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX brands_name_idx
ON brands(name)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX brands_slug_idx
ON brands(slug)
WHERE deleted_at IS NULL;

-- ============================================================
-- TAGS
-- ============================================================

CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY,

    name TEXT NOT NULL,

    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX tags_name_idx
ON tags(name)
WHERE deleted_at IS NULL;

-- ============================================================
-- PRODUCTS
-- ============================================================

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,

    brand_id UUID
        REFERENCES brands(id),

    title TEXT NOT NULL,
    slug TEXT NOT NULL,

    short_description TEXT NOT NULL,
    description TEXT NOT NULL,

    status tproduct_status NOT NULL DEFAULT 'draft',

    default_variant_id UUID,

    deleted_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX products_slug_idx
ON products(slug)
WHERE deleted_at IS NULL;

CREATE INDEX products_title_idx
ON products(title);

CREATE INDEX products_status_idx
ON products(status)
WHERE deleted_at IS NULL;

-- ============================================================
-- PRODUCT VARIANTS
-- ============================================================

CREATE TABLE IF NOT EXISTS product_variants (
    id UUID PRIMARY KEY,

    product_id UUID NOT NULL
        REFERENCES products(id)
        ON DELETE CASCADE,

    title TEXT,

    sku TEXT NOT NULL,

    price BIGINT NOT NULL,
    currency tcurrency NOT NULL,

    compare_at_price BIGINT,

    deleted_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX product_variants_sku_idx
ON product_variants(sku)
WHERE deleted_at IS NULL;

CREATE INDEX product_variants_product_idx
ON product_variants(product_id)
WHERE deleted_at IS NULL;

ALTER TABLE products
ADD CONSTRAINT products_default_variant_fk
FOREIGN KEY (default_variant_id)
REFERENCES product_variants(id);

-- ============================================================
-- ATTRIBUTES
-- ============================================================

CREATE TABLE IF NOT EXISTS attributes (
    id UUID PRIMARY KEY,

    product_id UUID NOT NULL
        REFERENCES products(id)
        ON DELETE CASCADE,

    name TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (product_id, name)
);

-- Examples:
-- Color
-- Storage
-- Size

-- ============================================================
-- ATTRIBUTE VALUES
-- ============================================================

CREATE TABLE IF NOT EXISTS attribute_values (
    id UUID PRIMARY KEY,

    attribute_id UUID NOT NULL
        REFERENCES attributes(id)
        ON DELETE CASCADE,

    value TEXT NOT NULL,

    UNIQUE(attribute_id, value)
);

-- Examples:
-- Color -> Black
-- Color -> Blue
-- Storage -> 128GB
-- Storage -> 256GB

-- ============================================================
-- VARIANT ATTRIBUTE VALUES
-- ============================================================

CREATE TABLE IF NOT EXISTS product_variant_attributes (
    variant_id UUID NOT NULL
        REFERENCES product_variants(id)
        ON DELETE CASCADE,

    attribute_value_id UUID NOT NULL
        REFERENCES attribute_values(id)
        ON DELETE CASCADE,

    PRIMARY KEY (
        variant_id,
        attribute_value_id
    )
);

-- ============================================================
-- PRODUCT MEDIA
-- ============================================================

CREATE TABLE IF NOT EXISTS product_media (
    id UUID PRIMARY KEY,

    product_id UUID NOT NULL
        REFERENCES products(id)
        ON DELETE CASCADE,

    variant_id UUID
        REFERENCES product_variants(id)
        ON DELETE CASCADE,

    type tmedia_type NOT NULL,

    url TEXT NOT NULL,
    storage_key TEXT NOT NULL,

    alt TEXT,
    mime_type TEXT,

    width INTEGER,
    height INTEGER,

    file_size BIGINT,

    sort_order INTEGER NOT NULL DEFAULT 0,

    is_primary BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX product_media_product_idx
ON product_media(product_id);

CREATE INDEX product_media_variant_idx
ON product_media(variant_id);

CREATE UNIQUE INDEX product_media_primary_product_idx
ON product_media(product_id)
WHERE is_primary = TRUE
AND variant_id IS NULL;

CREATE UNIQUE INDEX product_media_primary_variant_idx
ON product_media(variant_id)
WHERE is_primary = TRUE;

-- ============================================================
-- INVENTORIES
-- ============================================================

CREATE TABLE IF NOT EXISTS inventories (
    variant_id UUID PRIMARY KEY
        REFERENCES product_variants(id)
        ON DELETE CASCADE,

    quantity BIGINT NOT NULL DEFAULT 0,

    reserved_quantity BIGINT NOT NULL DEFAULT 0,

    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (quantity >= 0),
    CHECK (reserved_quantity >= 0),
    CHECK (reserved_quantity <= quantity)
);

-- ============================================================
-- PRODUCT CATEGORIES
-- ============================================================

CREATE TABLE IF NOT EXISTS product_categories (
    product_id UUID NOT NULL
        REFERENCES products(id)
        ON DELETE CASCADE,

    category_id UUID NOT NULL
        REFERENCES categories(id)
        ON DELETE CASCADE,

    PRIMARY KEY (
        product_id,
        category_id
    )
);

-- ============================================================
-- PRODUCT TAGS
-- ============================================================

CREATE TABLE IF NOT EXISTS product_tags (
    product_id UUID NOT NULL
        REFERENCES products(id)
        ON DELETE CASCADE,

    tag_id UUID NOT NULL
        REFERENCES tags(id)
        ON DELETE CASCADE,

    PRIMARY KEY (
        product_id,
        tag_id
    )
);

-- ============================================================
-- USEFUL INDEXES
-- ============================================================

CREATE INDEX products_brand_idx
ON products(brand_id)
WHERE deleted_at IS NULL;

CREATE INDEX product_variants_price_idx
ON product_variants(price)
WHERE deleted_at IS NULL;

CREATE INDEX product_variants_currency_idx
ON product_variants(currency)
WHERE deleted_at IS NULL;

CREATE INDEX products_active_idx ON products (status)
WHERE
  deleted_at IS NULL;



  -- ============================================================
  -- PRODUCT DATABASE SEEDING SCRIPT (100 Products)
  -- ============================================================
  -- This migration script seeds the database with 100 realistic products
  -- including brands, categories, tags, variants, attributes, media, and inventory

  -- Enable UUID extension if not already enabled
  CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

  -- ============================================================
  -- SEED BRANDS (5 brands)
  -- ============================================================

  INSERT INTO brands (id, name, slug, logo_url, logo_label, created_by, updated_by, created_at, updated_at)
  SELECT
      gen_random_uuid(),
      name,
      slug,
      logo_url,
      logo_label,
      gen_random_uuid(),
      gen_random_uuid(),
      now(),
      now()
  FROM (VALUES
      ('TechPro', 'techpro', 'https://via.placeholder.com/200x100?text=TechPro', 'TechPro Logo'),
      ('StyleMax', 'stylemax', 'https://via.placeholder.com/200x100?text=StyleMax', 'StyleMax Logo'),
      ('HomeComfort', 'homecomfort', 'https://via.placeholder.com/200x100?text=HomeComfort', 'HomeComfort Logo'),
      ('SportZone', 'sportzone', 'https://via.placeholder.com/200x100?text=SportZone', 'SportZone Logo'),
      ('FoodDelite', 'fooddelite', 'https://via.placeholder.com/200x100?text=FoodDelite', 'FoodDelite Logo')
  ) AS v(name, slug, logo_url, logo_label)
  ON CONFLICT DO NOTHING;

  -- ============================================================
  -- SEED CATEGORIES (5 categories)
  -- ============================================================

  CREATE TABLE IF NOT EXISTS categories (
      id UUID PRIMARY KEY,
      name TEXT NOT NULL,
      slug TEXT NOT NULL,
      created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      deleted_at TIMESTAMPTZ,
      UNIQUE(slug)
  );

  INSERT INTO categories (id, name, slug, created_at, updated_at)
  SELECT
      gen_random_uuid(),
      name,
      slug,
      now(),
      now()
  FROM (VALUES
      ('Electronics', 'electronics'),
      ('Fashion', 'fashion'),
      ('Home & Garden', 'home-garden'),
      ('Sports', 'sports'),
      ('Food & Beverage', 'food-beverage')
  ) AS v(name, slug)
  ON CONFLICT DO NOTHING;

  -- ============================================================
  -- SEED TAGS (6 tags)
  -- ============================================================

  INSERT INTO tags (id, name, created_by, updated_by, created_at, updated_at)
  SELECT
      gen_random_uuid(),
      name,
      gen_random_uuid(),
      gen_random_uuid(),
      now(),
      now()
  FROM (VALUES
      ('bestseller'),
      ('new-arrival'),
      ('limited-edition'),
      ('eco-friendly'),
      ('premium'),
      ('budget-friendly')
  ) AS v(name)
  ON CONFLICT DO NOTHING;

  -- ============================================================
  -- SEED 100 PRODUCTS WITH VARIANTS, ATTRIBUTES, MEDIA & INVENTORY
  -- ============================================================

  INSERT INTO products (id, brand_id, title, slug, short_description, description, status, created_at, updated_at)
  SELECT
      gen_random_uuid(),
      (ARRAY(SELECT id FROM brands ORDER BY random()))[1],
      name || ' v' || row_num,
      lower(replace(name, ' ', '-')) || '-' || row_num,
      'Short description for ' || name || ' v' || row_num,
      'This is a detailed description for ' || name || ' v' || row_num ||
      '. It includes features, benefits, and usage information.',
      CASE WHEN random() > 0.25 THEN 'active'::tproduct_status ELSE 'draft'::tproduct_status END,
      now(),
      now()
  FROM (
      SELECT
          unnest(ARRAY[
              'Wireless Headphones Pro', 'Smart Watch Ultra', 'USB-C Hub', 'Laptop Stand',
              'Mechanical Keyboard', 'Portable Monitor', 'Webcam 4K', 'USB Flash Drive',
              'Cable Organizer', 'Phone Mount', 'Cooling Pad', 'Desk Lamp LED',
              'Premium T-Shirt', 'Athletic Shorts', 'Running Shoes', 'Yoga Mat',
              'Sport Socks Pack', 'Windbreaker Jacket', 'Training Hat', 'Fitness Belt',
              'Cotton Bedsheet Set', 'Pillow Premium', 'Comforter Queen', 'Bath Towel Set',
              'Kitchen Knife Set', 'Mixing Bowls', 'Coffee Maker', 'Blender Pro',
              'Measuring Cups', 'Cutting Board', 'Baking Pan Set', 'Utensil Holder',
              'Organic Tea Collection', 'Dark Chocolate Bars', 'Granola Mix', 'Almond Butter',
              'Protein Powder', 'Energy Bars', 'Greek Yogurt', 'Honey Raw',
              'Camera DSLR', 'Tripod Stand', 'Lens Filter', 'Memory Card',
              'Phone Case', 'Screen Protector', 'Charger Cable', 'Power Bank',
              'Wireless Mouse', 'USB Keyboard', 'Monitor Stand', 'Desk Organizer',
              'Notebook Pad', 'Pen Set', 'Desk Chair', 'Standing Desk',
              'Coffee Mug', 'Water Bottle', 'Lunch Box', 'Food Container',
              'Backpack', 'Shoulder Bag', 'Travel Luggage', 'Duffel Bag',
              'Running Belt', 'Water Bottle Carrier', 'Gym Towel', 'Resistance Bands',
              'Dumbbells Set', 'Kettlebell', 'Weight Bench', 'Yoga Block',
              'Pillow Case', 'Mattress Pad', 'Bed Frame', 'Storage Box',
              'Shelving Unit', 'Wall Hooks', 'Door Mat', 'Shoe Rack',
              'Desk Lamp', 'Ceiling Light', 'String Lights', 'Lamp Base',
              'Candle Set', 'Air Purifier', 'Humidifier', 'Dehumidifier',
              'Vacuum Cleaner', 'Mop Bucket', 'Broom Set', 'Dustpan',
              'Dish Soap', 'Laundry Detergent', 'Fabric Softener', 'Bleach'
          ]) as name,
          row_number() over () as row_num
  ) product_list
  LIMIT 100;

  -- ============================================================
  -- SEED PRODUCT VARIANTS (2-3 per product)
  -- ============================================================

  INSERT INTO product_variants (id, product_id, title, sku, price, currency, compare_at_price, created_at, updated_at)
  SELECT
      gen_random_uuid(),
      p.id,
      'Variant ' || v_num,
      'SKU-' || to_char(now(), 'YYYYMMDDHH24MISS') || '-' || substr(p.id::text, 1, 8) || '-' || v_num,
      floor(random() * 50000 + 1000)::bigint,
      CASE floor(random() * 3)
          WHEN 0 THEN 'USD'::tcurrency
          WHEN 1 THEN 'EUR'::tcurrency
          ELSE 'EGP'::tcurrency
      END,
      CASE WHEN random() > 0.5 THEN floor(random() * 5000 + 5000)::bigint ELSE NULL END,
      now(),
      now()
  FROM products p
  CROSS JOIN (SELECT generate_series(1, 2 + floor(random() * 2)::int) as v_num) variants;

  -- ============================================================
  -- SET DEFAULT VARIANT FOR EACH PRODUCT
  -- ============================================================

  UPDATE products p
  SET default_variant_id = (
      SELECT id FROM product_variants
      WHERE product_id = p.id
      ORDER BY id
      LIMIT 1
  )
  WHERE default_variant_id IS NULL;

  -- ============================================================
  -- INSERT PRODUCT MEDIA (2 images per variant)
  -- ============================================================

  INSERT INTO product_media (
      id, product_id, variant_id, type, url, storage_key, alt, mime_type, width, height, file_size, sort_order, is_primary, created_at
  )
  SELECT
      gen_random_uuid(),
      pv.product_id,
      pv.id,
      'image',
      'https://via.placeholder.com/800x800?text=Product-' || substr(pv.id::text, 1, 8),
      'media/' || pv.product_id || '/' || pv.id || '/image-' || img_num || '.jpg',
      'Product image ' || img_num,
      'image/jpeg',
      800,
      800,
      floor(random() * 5000000 + 500000)::bigint,
      img_num - 1,
      img_num = 1,
      now()
  FROM product_variants pv
  CROSS JOIN (SELECT generate_series(1, 2) as img_num) images;

  -- ============================================================
  -- INSERT INVENTORY FOR VARIANTS
  -- ============================================================

  WITH inventory_data AS (
      SELECT
          id,
          (floor(random() * 900) + 100)::bigint as qty
      FROM product_variants
  )
  SELECT
      id,
      qty,
      (floor(random() * qty * 0.2))::bigint as reserved,  -- 0-20% of quantity
      now()
  FROM inventory_data;

  -- ============================================================
  -- INSERT ATTRIBUTES (Color, Size, Storage - rotating)
  -- ============================================================

  INSERT INTO attributes (id, product_id, name, created_at)
  SELECT
      gen_random_uuid(),
      p.id,
      CASE (row_number() over (order by p.id))::int % 3
          WHEN 1 THEN 'Color'
          WHEN 2 THEN 'Size'
          ELSE 'Storage'
      END,
      now()
  FROM products p;

  -- ============================================================
  -- INSERT ATTRIBUTE VALUES
  -- ============================================================

  INSERT INTO attribute_values (id, attribute_id, value)
  SELECT gen_random_uuid(), a.id, val
  FROM attributes a
  CROSS JOIN unnest(
      CASE a.name
          WHEN 'Color' THEN ARRAY['Black', 'White', 'Silver', 'Gold', 'Blue', 'Red', 'Green']
          WHEN 'Size' THEN ARRAY['XS', 'S', 'M', 'L', 'XL', 'XXL']
          WHEN 'Storage' THEN ARRAY['128GB', '256GB', '512GB', '1TB']
          ELSE ARRAY[]::text[]
      END
  ) AS vals(val)
  ON CONFLICT DO NOTHING;

  -- ============================================================
  -- LINK VARIANTS TO ATTRIBUTE VALUES
  -- ============================================================

  INSERT INTO product_variant_attributes (variant_id, attribute_value_id)
  SELECT DISTINCT ON (pv.id, av.id)
      pv.id,
      av.id
  FROM product_variants pv
  JOIN attributes a ON a.product_id = pv.product_id
  JOIN attribute_values av ON av.attribute_id = a.id
  ON CONFLICT DO NOTHING;

  -- ============================================================
  -- LINK PRODUCTS TO CATEGORIES (2 random per product)
  -- ============================================================

  INSERT INTO product_categories (product_id, category_id)
  SELECT
      p.id,
      c.id
  FROM products p
  CROSS JOIN (SELECT id FROM categories ORDER BY random() LIMIT 2) c
  ON CONFLICT DO NOTHING;

  -- ============================================================
  -- LINK PRODUCTS TO TAGS (3 random per product)
  -- ============================================================

  INSERT INTO product_tags (product_id, tag_id)
  SELECT
      p.id,
      t.id
  FROM products p
  CROSS JOIN (SELECT id FROM tags ORDER BY random() LIMIT 3) t
  ON CONFLICT DO NOTHING;

  -- ============================================================
  -- VERIFY SEEDING
  -- ============================================================

  SELECT
      (SELECT COUNT(*) FROM brands) as total_brands,
      (SELECT COUNT(*) FROM categories) as total_categories,
      (SELECT COUNT(*) FROM tags) as total_tags,
      (SELECT COUNT(*) FROM products) as total_products,
      (SELECT COUNT(*) FROM product_variants) as total_variants,
      (SELECT COUNT(*) FROM product_media) as total_media,
      (SELECT COUNT(*) FROM inventories) as total_inventories,
      (SELECT COUNT(*) FROM attributes) as total_attributes,
      (SELECT COUNT(*) FROM attribute_values) as total_attribute_values,
      (SELECT COUNT(*) FROM product_categories) as total_product_categories,
      (SELECT COUNT(*) FROM product_tags) as total_product_tags;
