-- =====================================================================
-- 000007 Seller custom attributes
-- Seller-scoped attribute definitions attached to a category. They extend
-- the admin-defined category attributes for that merchant's products only.
-- =====================================================================

BEGIN;

CREATE TABLE IF NOT EXISTS seller_attributes (
    id             uuid NOT NULL,
    seller_id      uuid NOT NULL REFERENCES sellers (id) ON DELETE CASCADE,
    category_id    uuid NOT NULL REFERENCES product_categories (id) ON DELETE CASCADE,
    key            text NOT NULL,
    label          text NOT NULL,
    value_type     attribute_value_type NOT NULL,
    allowed_values jsonb NULL,
    is_filterable  boolean NOT NULL DEFAULT false,
    sort_order     int NOT NULL DEFAULT 0,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    UNIQUE (seller_id, category_id, key)
);

CREATE INDEX IF NOT EXISTS idx_seller_attributes_lookup
    ON seller_attributes (seller_id, category_id);

COMMIT;
