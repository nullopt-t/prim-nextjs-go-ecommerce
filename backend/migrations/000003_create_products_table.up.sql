CREATE TABLE brands (
  id UUID PRIMARY KEY NOT NULL,
  name TEXT NOT NULL UNIQUE,
  slug TEXT NOT NULL UNIQUE,
  logo_url TEXT NOT NULL,
  logo_label TEXT NOT NULL,
  created_by UUID NOT NULL,
  updated_by UUID NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX brands_name_idx ON brands (name)
WHERE
  deleted_at IS NULL;

CREATE INDEX brands_slug_idx ON brands (slug)
WHERE
  deleted_at IS NULL;

CREATE TABLE tags (
  id UUID PRIMARY KEY NOT NULL,
  name TEXT NOT NULL UNIQUE,
  created_by UUID NOT NULL,
  updated_by UUID NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX tags_name_idx ON brands (name)
WHERE
  deleted_at IS NULL;

DROP TYPE IF EXISTS tproduct_status;

CREATE TYPE tproduct_status AS ENUM('draft', 'active', 'archived');

DROP TYPE IF EXISTS tcurrency;

CREATE TYPE tcurrency AS ENUM('USD', 'EUR', 'EGP');

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
WHERE
  deleted_at IS NULL;

CREATE TABLE product_categories (
  product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
  category_id UUID NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
  PRIMARY KEY (product_id, category_id)
);
