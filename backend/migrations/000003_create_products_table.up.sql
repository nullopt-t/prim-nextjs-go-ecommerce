CREATE TABLE products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title TEXT NOT NULL,
  short_description TEXT NOT NULL,
  description TEXT NOT NULL,
  sku TEXT UNIQUE NOT NULL,
  slug TEXT UNIQUE NOT NULL CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
  status TEXT NOT NULL DEFAULT 'draft' CHECK(
        status IN ('draft', 'active', 'archived')
  ),
  price BIGINT NOT NULL,
  currency TEXT NOT NULL DEFAULT 'USD' CHECK(
    currency IN ('USD', 'EUR', 'EGP')
  ),
  deleted_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX products_title_idx ON products (title);
CREATE INDEX products_sku_idx ON products (sku);
CREATE INDEX products_slug_idx ON products (slug);
CREATE INDEX products_price_idx ON products (price);
CREATE INDEX products_currency_idx ON products (currency);
CREATE INDEX products_status_idx ON products (status);
CREATE INDEX products_active_idx ON products (status)
    WHERE deleted_at IS NULL;


CREATE TABLE product_categories (
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  PRIMARY KEY (product_id, category_id)
);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER products_updated_at BEFORE
UPDATE ON products FOR EACH ROW EXECUTE FUNCTION set_updated_at();
