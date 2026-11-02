-- =====================================================================
-- 000003 Vendor Domain (Design B — vendor-owned products)
-- See docs/adr-vendor-domain.md
-- =====================================================================

BEGIN;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'vendor_status') THEN
        CREATE TYPE vendor_status AS ENUM ('pending', 'active', 'suspended', 'rejected');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payout_status') THEN
        CREATE TYPE payout_status AS ENUM ('requested', 'approved', 'processing', 'paid', 'rejected');
    END IF;
END
$$;

-- Vendor profiles: 1:1 with users. Soft-deleted vendors keep their order
-- history; FKs on products/order_items use ON DELETE SET NULL so deleting a
-- vendor de-attributes instead of destroying money trail.
CREATE TABLE IF NOT EXISTS vendors (
    id                   uuid NOT NULL,
    user_id              uuid NOT NULL,
    store_name           text NOT NULL CHECK (length(store_name) BETWEEN 2 AND 150),
    slug                 text NOT NULL CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    description          text NULL,
    logo_object_id       uuid NULL,
    banner_object_id     uuid NULL,
    commission_rate_bps  int NOT NULL DEFAULT 1000 CHECK (commission_rate_bps BETWEEN 0 AND 10000),
    status               vendor_status NOT NULL DEFAULT 'pending',
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

CREATE INDEX IF NOT EXISTS idx_vendors_status ON vendors (status) WHERE deleted_at IS NULL;

-- Ledger of payout requests. Rejected rows are retained (full audit trail).
-- snapshot_available_balance records what the balance validation computed at
-- request time; the enforceable invariant is re-validated under the vendor
-- row lock in RequestPayout.
CREATE TABLE IF NOT EXISTS vendor_payouts (
    id                          uuid NOT NULL,
    vendor_id                   uuid NOT NULL,
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
    FOREIGN KEY (vendor_id) REFERENCES vendors (id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_vendor_payouts_vendor_id ON vendor_payouts (vendor_id);
CREATE INDEX IF NOT EXISTS idx_vendor_payouts_status ON vendor_payouts (status);

-- Attribution: pre-existing products/orders keep NULL = platform-owned.
ALTER TABLE products ADD COLUMN vendor_id uuid NULL REFERENCES vendors (id) ON DELETE SET NULL;
ALTER TABLE order_items ADD COLUMN vendor_id uuid NULL REFERENCES vendors (id) ON DELETE SET NULL;
ALTER TABLE order_items ADD COLUMN vendor_earning bigint NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_products_vendor_id ON products (vendor_id);
CREATE INDEX IF NOT EXISTS idx_order_items_vendor_id ON order_items (vendor_id);

COMMIT;
