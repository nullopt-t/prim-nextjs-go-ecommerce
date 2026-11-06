-- =====================================================================
-- 000003 Seller Domain (Design B — seller-owned products)
-- See docs/adr-seller-domain.md
-- =====================================================================

BEGIN;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'seller_status') THEN
        CREATE TYPE seller_status AS ENUM ('pending', 'active', 'suspended', 'rejected');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payout_status') THEN
        CREATE TYPE payout_status AS ENUM ('requested', 'approved', 'processing', 'paid', 'rejected');
    END IF;
END
$$;

-- Seller profiles: 1:1 with users. Soft-deleted sellers keep their order
-- history; FKs on products/order_items use ON DELETE SET NULL so deleting a
-- seller de-attributes instead of destroying money trail.
CREATE TABLE IF NOT EXISTS sellers (
    id                   uuid NOT NULL,
    user_id              uuid NOT NULL,
    store_name           text NOT NULL CHECK (length(store_name) BETWEEN 2 AND 150),
    slug                 text NOT NULL CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    description          text NULL,
    logo_object_id       uuid NULL,
    banner_object_id     uuid NULL,
    commission_rate_bps  int NOT NULL DEFAULT 1000 CHECK (commission_rate_bps BETWEEN 0 AND 10000),
    status               seller_status NOT NULL DEFAULT 'pending',
    bank_account_info    jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz NULL,
    PRIMARY KEY (id),
    UNIQUE (user_id),
    UNIQUE (slug),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (logo_object_id) REFERENCES storage_objects (id) ON DELETE SET NULL,
    FOREIGN KEY (banner_object_id) REFERENCES storage_objects (id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_sellers_status ON sellers (status) WHERE deleted_at IS NULL;

-- Ledger of payout requests. Rejected rows are retained (full audit trail).
-- snapshot_available_balance records what the balance validation computed at
-- request time; the enforceable invariant is re-validated under the seller
-- row lock in RequestPayout.
CREATE TABLE IF NOT EXISTS seller_payouts (
    id                          uuid NOT NULL,
    seller_id                   uuid NOT NULL,
    amount                      bigint NOT NULL CHECK (amount > 0),
    currency                    text NOT NULL DEFAULT 'EGP',
    status                      payout_status NOT NULL DEFAULT 'requested',
    reference_number            text NULL,
    notes                       text NULL,
    snapshot_available_balance  bigint NOT NULL DEFAULT 0,
    created_at                  timestamptz NOT NULL DEFAULT now(),
    updated_at                  timestamptz NOT NULL DEFAULT now(),
    processed_at                timestamptz NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (seller_id) REFERENCES sellers (id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_seller_payouts_seller_id ON seller_payouts (seller_id);
CREATE INDEX IF NOT EXISTS idx_seller_payouts_status ON seller_payouts (status);

-- Attribution: pre-existing products/orders keep NULL = platform-owned.
ALTER TABLE products ADD COLUMN seller_id uuid NULL REFERENCES sellers (id) ON DELETE SET NULL;
ALTER TABLE order_items ADD COLUMN seller_id uuid NULL REFERENCES sellers (id) ON DELETE SET NULL;
ALTER TABLE order_items ADD COLUMN seller_earning bigint NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_products_seller_id ON products (seller_id);
CREATE INDEX IF NOT EXISTS idx_order_items_seller_id ON order_items (seller_id);

COMMIT;
