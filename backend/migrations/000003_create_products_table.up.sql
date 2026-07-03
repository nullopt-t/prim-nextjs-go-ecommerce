CREATE TABLE brands (
  id UUID PRIMARY KEY NOT NULL,
  name TEXT NOT NULL UNIQUE,
  logo_url TEXT NOT NULL,
  logo_label TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX brands_name_idx ON brands(name)
  WHERE deleted_at IS NULL;


INSERT INTO brands (
    id,
    name,
    logo_url,
    logo_label,
    created_at,
    updated_at
)
VALUES
(gen_random_uuid(), 'Apple',      'https://img.logo.dev/apple.com',        'Apple logo', NOW(), NOW()),
(gen_random_uuid(), 'Samsung',    'https://img.logo.dev/samsung.com',      'Samsung logo', NOW(), NOW()),
(gen_random_uuid(), 'Sony',       'https://img.logo.dev/sony.com',         'Sony logo', NOW(), NOW()),
(gen_random_uuid(), 'LG',         'https://img.logo.dev/lg.com',           'LG logo', NOW(), NOW()),
(gen_random_uuid(), 'Dell',       'https://img.logo.dev/dell.com',         'Dell logo', NOW(), NOW()),
(gen_random_uuid(), 'Lenovo',     'https://img.logo.dev/lenovo.com',       'Lenovo logo', NOW(), NOW()),
(gen_random_uuid(), 'ASUS',       'https://img.logo.dev/asus.com',         'ASUS logo', NOW(), NOW()),
(gen_random_uuid(), 'HP',         'https://img.logo.dev/hp.com',           'HP logo', NOW(), NOW()),
(gen_random_uuid(), 'Logitech',   'https://img.logo.dev/logitech.com',     'Logitech logo', NOW(), NOW()),
(gen_random_uuid(), 'Nike',       'https://img.logo.dev/nike.com',         'Nike logo', NOW(), NOW()),
(gen_random_uuid(), 'Adidas',     'https://img.logo.dev/adidas.com',       'Adidas logo', NOW(), NOW()),
(gen_random_uuid(), 'Puma',       'https://img.logo.dev/puma.com',         'Puma logo', NOW(), NOW()),
(gen_random_uuid(), 'New Balance','https://img.logo.dev/newbalance.com',   'New Balance logo', NOW(), NOW()),
(gen_random_uuid(), 'Canon',      'https://img.logo.dev/canon.com',        'Canon logo', NOW(), NOW()),
(gen_random_uuid(), 'Nikon',      'https://img.logo.dev/nikon.com',        'Nikon logo', NOW(), NOW()),
(gen_random_uuid(), 'Xiaomi',     'https://img.logo.dev/mi.com',           'Xiaomi logo', NOW(), NOW()),
(gen_random_uuid(), 'Huawei',     'https://img.logo.dev/huawei.com',       'Huawei logo', NOW(), NOW()),
(gen_random_uuid(), 'Google',     'https://img.logo.dev/google.com',       'Google logo', NOW(), NOW()),
(gen_random_uuid(), 'Microsoft',  'https://img.logo.dev/microsoft.com',    'Microsoft logo', NOW(), NOW()),
(gen_random_uuid(), 'Amazon',     'https://img.logo.dev/amazon.com',       'Amazon logo', NOW(), NOW());

CREATE TABLE tags (
  id UUID PRIMARY KEY NOT NULL,
  name TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX tags_name_idx ON brands(name)
  WHERE deleted_at IS NULL;

INSERT INTO tags (
    id,
    name,
    created_at,
    updated_at
)
VALUES
(gen_random_uuid(), 'New Arrival', NOW(), NOW()),
(gen_random_uuid(), 'Best Seller', NOW(), NOW()),
(gen_random_uuid(), 'Trending', NOW(), NOW()),
(gen_random_uuid(), 'Featured', NOW(), NOW()),
(gen_random_uuid(), 'Limited Edition', NOW(), NOW()),
(gen_random_uuid(), 'Exclusive', NOW(), NOW()),
(gen_random_uuid(), 'On Sale', NOW(), NOW()),
(gen_random_uuid(), 'Clearance', NOW(), NOW()),
(gen_random_uuid(), 'Flash Sale', NOW(), NOW()),
(gen_random_uuid(), 'Staff Pick', NOW(), NOW()),
(gen_random_uuid(), 'Top Rated', NOW(), NOW()),
(gen_random_uuid(), 'Pre-order', NOW(), NOW()),
(gen_random_uuid(), 'Coming Soon', NOW(), NOW()),
(gen_random_uuid(), 'Free Shipping', NOW(), NOW()),
(gen_random_uuid(), 'Eco Friendly', NOW(), NOW()),
(gen_random_uuid(), 'Refurbished', NOW(), NOW()),
(gen_random_uuid(), 'Open Box', NOW(), NOW()),
(gen_random_uuid(), 'Bundle Deal', NOW(), NOW()),
(gen_random_uuid(), 'Premium', NOW(), NOW()),
(gen_random_uuid(), 'Budget Friendly', NOW(), NOW());

DROP TYPE IF EXISTS tproduct_status;
CREATE TYPE tproduct_status AS ENUM(
  'draft',
  'active',
  'archived'
);

DROP TYPE IF EXISTS tcurrency;
CREATE TYPE tcurrency AS ENUM(
  'USD',
  'EUR',
  'EGP'
);

CREATE TABLE products (
  id UUID PRIMARY KEY NOT NULL,
  title TEXT NOT NULL,
  short_description TEXT NOT NULL,
  description TEXT NOT NULL,
  sku TEXT UNIQUE NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  status tproduct_status NOT NULL,
  price BIGINT NOT NULL,
  currency tcurrency NOT NULL,
  deleted_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX products_title_idx ON products (title);
CREATE INDEX products_sku_idx ON products (sku);
CREATE INDEX products_slug_idx ON products (slug);
CREATE INDEX products_price_idx ON products (price);
CREATE INDEX products_currency_idx ON products (currency);

CREATE INDEX products_active_idx ON products (status)
    WHERE deleted_at IS NULL;


CREATE TABLE product_categories (
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  PRIMARY KEY (product_id, category_id)
);
