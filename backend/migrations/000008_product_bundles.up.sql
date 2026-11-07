-- =====================================================================
-- 000008 Product bundles
-- A bundle is a sellable product composed of other seller catalog items.
-- The bundle's own product carries the bundle price; bundle_items records
-- the composition for display and fulfillment.
-- =====================================================================

BEGIN;

CREATE TABLE IF NOT EXISTS product_bundles (
    id           uuid NOT NULL,
    seller_id    uuid NOT NULL REFERENCES sellers (id) ON DELETE CASCADE,
    product_id   uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    bundle_price bigint NOT NULL CHECK (bundle_price > 0),
    currency     text NOT NULL DEFAULT 'EGP',
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    UNIQUE (product_id)
);

CREATE TABLE IF NOT EXISTS bundle_items (
    id         uuid NOT NULL,
    bundle_id  uuid NOT NULL REFERENCES product_bundles (id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    variant_id uuid NULL REFERENCES product_variants (id) ON DELETE CASCADE,
    quantity   int NOT NULL DEFAULT 1 CHECK (quantity > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_bundle_items_bundle ON bundle_items (bundle_id);
CREATE INDEX IF NOT EXISTS idx_product_bundles_seller ON product_bundles (seller_id);

COMMIT;
