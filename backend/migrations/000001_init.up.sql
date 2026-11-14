BEGIN;

-- =====================================================================
-- PRIM E-COMMERCE DATABASE SCHEMA (Production Hardened)
-- PostgreSQL
-- Consolidated: schema for all domains (seller, merchant lifecycle,
-- user product views, coupon seller scope, seller attributes, bundles).
-- =====================================================================


-- ---- 000001_init_sechema.up.sql ----
-- =====================================================================
-- PRIM E-COMMERCE DATABASE SCHEMA (Production Hardened)
-- PostgreSQL
-- Consolidated (dev): 000001_init_sechema + 000003_audit_log + 000004_reservation_order_link
-- =====================================================================
-- =====================================================================
-- ENUM TYPES
-- =====================================================================

DO $$
BEGIN
    IF to_regtype('storage_object_status') IS NULL THEN
        CREATE TYPE storage_object_status AS ENUM ('uploaded', 'uploading', 'deleting', 'deleted');
    END IF;
    IF to_regtype('product_status') IS NULL THEN
        CREATE TYPE product_status AS ENUM ('draft', 'published', 'archived');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'product_type') THEN
        CREATE TYPE product_type AS ENUM ('simple', 'variable', 'downloadable');
    END IF;
    IF to_regtype('user_role') IS NULL THEN
        CREATE TYPE user_role AS ENUM ('customer', 'seller', 'support', 'admin', 'superadmin');
    END IF;
    IF to_regtype('order_status') IS NULL THEN
        CREATE TYPE order_status AS ENUM ('pending', 'paid', 'processing', 'shipped', 'delivered', 'cancelled', 'refunded');
    END IF;
    IF to_regtype('payment_status') IS NULL THEN
        CREATE TYPE payment_status AS ENUM ('pending', 'authorized', 'captured', 'failed', 'refunded');
    END IF;
    IF to_regtype('review_status') IS NULL THEN
        CREATE TYPE review_status AS ENUM ('pending', 'approved', 'rejected');
    END IF;
    IF to_regtype('discount_type') IS NULL THEN
        CREATE TYPE discount_type AS ENUM ('percentage', 'fixed_amount');
    END IF;
    IF to_regtype('attribute_value_type') IS NULL THEN
        CREATE TYPE attribute_value_type AS ENUM ('enum', 'text', 'number', 'boolean');
    END IF;
    IF to_regtype('coupon_status') IS NULL THEN
        CREATE TYPE coupon_status AS ENUM ('active', 'disabled', 'archived');
    END IF;
    IF to_regtype('inventory_reason') IS NULL THEN
        CREATE TYPE inventory_reason AS ENUM ('restock', 'sale', 'return', 'adjustment', 'reservation_release');
    END IF;
    IF to_regtype('fulfillment_status') IS NULL THEN
        CREATE TYPE fulfillment_status AS ENUM ('pending', 'shipped', 'delivered');
    END IF;
    IF to_regtype('user_status') IS NULL THEN
        CREATE TYPE user_status AS ENUM ('active', 'suspended', 'deleted');
    END IF;
END
$$;


-- =====================================================================
-- STORAGE / MEDIA
-- =====================================================================

CREATE TABLE IF NOT EXISTS storage_objects (
    id              uuid NOT NULL,
    bucket          text NOT NULL,
    object_key      text NOT NULL,
    content_type    text NOT NULL,
    file_size       bigint NOT NULL,
    status          storage_object_status NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz NULL,
    PRIMARY KEY (id),
    UNIQUE (bucket, object_key)
);


-- =====================================================================
-- USERS & ADDRESSES
-- =====================================================================

CREATE TABLE IF NOT EXISTS users (
    id                   uuid NOT NULL,
    full_name            text NOT NULL,
    phone                text NOT NULL,
    phone_verified        boolean NOT NULL DEFAULT false,
    country_code          text NOT NULL,
    role                 user_role NOT NULL DEFAULT 'customer',
    status               user_status NOT NULL DEFAULT 'active',
    preference           jsonb NOT NULL DEFAULT '{}'::jsonb,
    suspended_until      timestamptz NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz NULL,
    PRIMARY KEY (id)
);

-- Fix: Allow re-registering with the same phone if the old account was soft-deleted
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_lower
    ON users (lower(phone))
    WHERE deleted_at IS NULL;

-- Enforce exactly one active superadmin in the system
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_single_superadmin
    ON users (role)
    WHERE role = 'superadmin' AND deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS addresses (
    id             uuid NOT NULL,
    user_id        uuid NOT NULL,
    line1          text NOT NULL,
    line2          text NULL,
    city           text NOT NULL,
    state          text NULL,
    postal_code    text NULL,
    country        text NOT NULL,
    is_default     boolean NOT NULL DEFAULT false,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    deleted_at     timestamptz NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE UNIQUE INDEX idx_addresses_user_default
    ON addresses (user_id)
    WHERE is_default = true AND deleted_at IS NULL;

-- =====================================================================
-- CATALOG: BRANDS, CATEGORIES, TAGS
-- =====================================================================

CREATE TABLE IF NOT EXISTS product_brands (
    id                        uuid NOT NULL,
    name                      text NOT NULL,
    link                      text NULL,
    logo_object_id            uuid NULL,
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz NOT NULL DEFAULT now(),
    deleted_at                timestamptz NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (logo_object_id) REFERENCES storage_objects (id)
);

CREATE UNIQUE INDEX idx_product_brands_active_name
ON product_brands (name)
WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS product_tags (
    id            uuid NOT NULL,
    name          text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz NULL,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_tags_active_name
ON product_tags (name)
WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS product_categories (
    id            uuid NOT NULL,
    parent_id     uuid NULL,
    name          text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (parent_id) REFERENCES product_categories (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_categories_name_parent
    ON product_categories (name, COALESCE(parent_id, '00000000-0000-0000-0000-000000000000'))
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS category_attributes (
    id                uuid NOT NULL,
    category_id       uuid NOT NULL,
    key               text NOT NULL,
    label             text NOT NULL,
    value_type        attribute_value_type NOT NULL,
    allowed_values    jsonb NULL,
    is_filterable     boolean NOT NULL DEFAULT true,
    sort_order        int NOT NULL DEFAULT 0,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    UNIQUE (category_id, key),
    FOREIGN KEY (category_id) REFERENCES product_categories (id)
);


-- =====================================================================
-- CATALOG: PRODUCTS & VARIANTS
-- =====================================================================

CREATE TABLE IF NOT EXISTS products (
    id             uuid NOT NULL,
    brand_id       uuid NULL,
    category_id    uuid NOT NULL,
    slug           text NOT NULL,
    title          text NOT NULL,
    description    text,
    highlights     jsonb NULL,
    status         product_status NOT NULL,
    product_type   product_type NOT NULL DEFAULT 'simple',
    variation_axes text[] NOT NULL DEFAULT '{}'::text[],
    download_url   text NULL,
    delivery_instructions text NULL,
    thumbnail_object_id uuid NULL,
    default_variant_id uuid NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    deleted_at     timestamptz NULL,
    PRIMARY KEY (id),
    UNIQUE (slug),
    FOREIGN KEY (brand_id) REFERENCES product_brands (id),
    FOREIGN KEY (category_id) REFERENCES product_categories (id),
    FOREIGN KEY (thumbnail_object_id) REFERENCES storage_objects (id)
);

CREATE TABLE IF NOT EXISTS product_variants (
    id                   uuid NOT NULL,
    sku                  text NOT NULL,
    product_id           uuid NOT NULL,
    title                text NOT NULL,
    price                bigint NULL,
    crossed_out_price    bigint NULL,
    currency             text NOT NULL DEFAULT 'EGP',
    attributes           jsonb NOT NULL DEFAULT '{}'::jsonb,
    thumbnail_object_id  uuid NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz NULL,
    PRIMARY KEY (id),
    UNIQUE (sku),
    FOREIGN KEY (product_id) REFERENCES products (id),
    FOREIGN KEY (thumbnail_object_id) REFERENCES storage_objects (id)
);

ALTER TABLE products
    ADD CONSTRAINT products_default_variant_id_fkey
    FOREIGN KEY (default_variant_id) REFERENCES product_variants (id) ON DELETE SET NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_variants_combination
    ON product_variants (product_id, attributes)
    WHERE deleted_at IS NULL AND attributes <> '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_product_variants_attributes ON product_variants USING GIN (attributes);
CREATE INDEX IF NOT EXISTS idx_product_variants_product_id ON product_variants (product_id);
CREATE INDEX IF NOT EXISTS idx_products_default_variant_id ON products (default_variant_id);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products (category_id);
CREATE INDEX IF NOT EXISTS idx_products_brand_id ON products (brand_id);
CREATE INDEX IF NOT EXISTS idx_product_categories_parent_id ON product_categories (parent_id);


-- =====================================================================
-- CATALOG: PRODUCT <-> TAGS (many-to-many)
-- =====================================================================

CREATE TABLE IF NOT EXISTS product_tag_assignments (
    product_id    uuid NOT NULL,
    tag_id        uuid NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (product_id, tag_id),
    FOREIGN KEY (product_id) REFERENCES products (id),
    FOREIGN KEY (tag_id) REFERENCES product_tags (id)
);


-- =====================================================================
-- CATALOG: MEDIA
-- =====================================================================

CREATE TABLE IF NOT EXISTS variant_media (
    id                   bigint GENERATED ALWAYS AS IDENTITY,
    variant_id           uuid NOT NULL,
    object_id            uuid NOT NULL,
    media_type           text NOT NULL,
    sort_order           int NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    FOREIGN KEY (variant_id) REFERENCES product_variants (id),
    FOREIGN KEY (object_id) REFERENCES storage_objects (id)
);

CREATE INDEX IF NOT EXISTS idx_variant_media_variant_id ON variant_media (variant_id);


-- =====================================================================
-- CARTS
-- =====================================================================

CREATE TABLE IF NOT EXISTS carts (
    id            uuid NOT NULL,
    user_id       uuid NULL,
    session_id    text NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    CHECK (user_id IS NOT NULL OR session_id IS NOT NULL)
);

-- One active cart per user at most
CREATE UNIQUE INDEX IF NOT EXISTS idx_carts_active_user
    ON carts (user_id)
    WHERE user_id IS NOT NULL;

-- One active cart per session_id at most
CREATE UNIQUE INDEX IF NOT EXISTS idx_carts_active_session_id
    ON carts (session_id)
    WHERE session_id IS NOT NULL AND user_id IS NULL;

CREATE TABLE IF NOT EXISTS cart_items (
    id                   bigint GENERATED ALWAYS AS IDENTITY,
    cart_id              uuid NOT NULL,
    variant_id           uuid NOT NULL,
    quantity             int NOT NULL,
    price_at_purchase    bigint NOT NULL,
    currency             text NOT NULL DEFAULT 'EGP',
    carted_at            timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    UNIQUE (cart_id, variant_id),
    FOREIGN KEY (cart_id) REFERENCES carts (id) ON DELETE CASCADE,
    FOREIGN KEY (variant_id) REFERENCES product_variants (id),
    CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id ON cart_items (cart_id);


-- =====================================================================
-- INVENTORY
-- =====================================================================

CREATE TABLE IF NOT EXISTS inventory_ledgers (
    id              bigint GENERATED ALWAYS AS IDENTITY,
    variant_id      uuid NOT NULL,
    quantity        int NOT NULL,
    reason          inventory_reason NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (variant_id) REFERENCES product_variants (id),
    CHECK (quantity <> 0)
);

CREATE INDEX IF NOT EXISTS idx_inventory_ledgers_variant_id ON inventory_ledgers (variant_id);

CREATE TABLE IF NOT EXISTS inventory_reservations (
    id             uuid NOT NULL,
    variant_id     uuid NOT NULL,
    cart_id        uuid NULL,
    -- order_id links a reservation to the order created by the checkout that
    -- placed the hold. It is the key used to commit the hold on payment,
    -- release it on cancellation, and expire it from the maintenance reaper.
    -- The FK to orders is added after the orders table is created, further
    -- below (orders is defined after this table, so it cannot be referenced
    -- inline here).
    order_id       uuid NULL,
    quantity       int NOT NULL,
    expires_at     timestamptz NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    released_at    timestamptz NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (variant_id) REFERENCES product_variants (id),
    FOREIGN KEY (cart_id) REFERENCES carts (id) ON DELETE CASCADE,
    CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_inventory_reservations_variant_id ON inventory_reservations (variant_id);
CREATE INDEX IF NOT EXISTS idx_inventory_reservations_expires_at ON inventory_reservations (expires_at) WHERE released_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_inventory_reservations_order_id ON inventory_reservations (order_id);

-- At most one live reservation per (order, variant), so a retried checkout for
-- the same order cannot double-hold the same SKU.
CREATE UNIQUE INDEX IF NOT EXISTS idx_inventory_reservations_order_variant
    ON inventory_reservations (order_id, variant_id)
    WHERE order_id IS NOT NULL AND released_at IS NULL;


-- =====================================================================
-- COUPONS
-- =====================================================================

CREATE TABLE IF NOT EXISTS coupons (
    id                 uuid NOT NULL,
    code               text NOT NULL,
    discount_type      discount_type NOT NULL,
    discount_value     bigint NOT NULL,
    currency           text NOT NULL DEFAULT 'EGP',
    min_cart_amount    bigint NULL,
    usage_limit        int NULL,
    per_user_limit     int NULL DEFAULT 1,
    starts_at          timestamptz NULL,
    expires_at         timestamptz NULL,
    status             coupon_status NOT NULL DEFAULT 'active',
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    UNIQUE (code),
    CHECK (
    (discount_type     = 'percentage' AND discount_value BETWEEN 1 AND 100)
    OR                 (discount_type = 'fixed_amount' AND discount_value > 0)
    )
);


-- =====================================================================
-- ORDERS
-- =====================================================================

CREATE TABLE IF NOT EXISTS orders (
    id                  uuid NOT NULL,
    customer_id         uuid NULL,
    customer_email      text NOT NULL,
    shipping_address    jsonb NOT NULL,
    status              order_status NOT NULL DEFAULT 'pending',
    coupon_id           uuid NULL,
    discount_amount     bigint NOT NULL DEFAULT 0,
    shipping_amount     bigint NOT NULL DEFAULT 0,
    tax_amount          bigint NOT NULL DEFAULT 0,
    cod_fee             bigint NOT NULL DEFAULT 0,
    total_amount        bigint NOT NULL,
    currency            text NOT NULL DEFAULT 'EGP',
    payment_method      text NOT NULL DEFAULT 'pending',
    tracking_number     text NULL,
    locale              text NOT NULL DEFAULT 'en',
    carrier             text NULL,
    tracking_url        text NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (customer_id) REFERENCES users (id),
    FOREIGN KEY (coupon_id) REFERENCES coupons (id)
);

CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders (customer_id);

-- Now that orders exists, link reservations to the order that created the hold.
-- ON DELETE CASCADE keeps reservations consistent with hard order deletion.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'inventory_reservations_order_id_fkey'
    ) THEN
        ALTER TABLE inventory_reservations
            ADD CONSTRAINT inventory_reservations_order_id_fkey
            FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS order_items (
    id                   uuid NOT NULL,
    order_id             uuid NOT NULL,
    variant_id           uuid NOT NULL,
    quantity             int NOT NULL,
    price_at_purchase    bigint NOT NULL,
    product_snapshot     jsonb NOT NULL,
    fulfillment_status   fulfillment_status NOT NULL DEFAULT 'pending',
    tracking_number      text NULL,
    carrier              text NULL,
    shipped_at           timestamptz NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (order_id) REFERENCES orders (id),
    FOREIGN KEY (variant_id) REFERENCES product_variants (id),
    CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_variant_id ON order_items (variant_id);

-- =====================================================================
-- ORDER RETURNS (RMA)
-- =====================================================================
-- Customer return requests for delivered orders. Cumulative returned
-- quantities per line item are enforced across multiple requests, not
-- just within one.
CREATE TABLE IF NOT EXISTS order_returns (
    id          uuid NOT NULL,
    order_id    uuid NOT NULL,
    customer_id uuid NOT NULL,
    reason      text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE,
    FOREIGN KEY (customer_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_order_returns_order_id ON order_returns (order_id);

CREATE TABLE IF NOT EXISTS order_return_items (
    id             uuid NOT NULL,
    return_id      uuid NOT NULL,
    order_item_id  uuid NOT NULL,
    quantity       int NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (return_id, order_item_id),
    CHECK (quantity > 0),
    FOREIGN KEY (return_id) REFERENCES order_returns (id) ON DELETE CASCADE,
    FOREIGN KEY (order_item_id) REFERENCES order_items (id)
);

CREATE INDEX IF NOT EXISTS idx_order_return_items_return_id ON order_return_items (return_id);
CREATE INDEX IF NOT EXISTS idx_order_return_items_order_item_id ON order_return_items (order_item_id);

CREATE TABLE IF NOT EXISTS order_status_history (
    id            bigint GENERATED ALWAYS AS IDENTITY,
    order_id      uuid NOT NULL,
    status        order_status NOT NULL,
    notes         text NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (order_id) REFERENCES orders (id)
);

CREATE INDEX IF NOT EXISTS idx_order_status_history_order_id ON order_status_history (order_id);

CREATE TABLE IF NOT EXISTS coupon_redemptions (
    id                 bigint GENERATED ALWAYS AS IDENTITY,
    coupon_id          uuid NOT NULL,
    user_id            uuid NULL,
    order_id           uuid NOT NULL,
    discount_amount    bigint NOT NULL,
    currency           text NOT NULL DEFAULT 'EGP',
    created_at         timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (coupon_id) REFERENCES coupons (id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (order_id) REFERENCES orders (id)
);

CREATE INDEX IF NOT EXISTS idx_coupon_redemptions_coupon_id ON coupon_redemptions (coupon_id);

-- Enforce at most one redemption per user per coupon at the DB level
CREATE UNIQUE INDEX IF NOT EXISTS idx_coupon_redemptions_coupon_user
    ON coupon_redemptions (coupon_id, user_id);


-- =====================================================================
-- PAYMENTS
-- =====================================================================

CREATE TABLE IF NOT EXISTS payments (
    id                         uuid NOT NULL,
    order_id                   uuid NOT NULL,
    amount                     bigint NOT NULL,
    currency                   text NOT NULL DEFAULT 'EGP',
    status                     payment_status NOT NULL DEFAULT 'pending',
    provider                   text NOT NULL,
    provider_transaction_id    text NULL,
    error_message              text NULL,
    created_at                 timestamptz NOT NULL DEFAULT now(),
    updated_at                 timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    UNIQUE (provider_transaction_id),
    FOREIGN KEY (order_id) REFERENCES orders (id)
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments (order_id);


-- =====================================================================
-- REVIEWS
-- =====================================================================

CREATE TABLE IF NOT EXISTS reviews (
    id               uuid NOT NULL,
    product_id       uuid NOT NULL,
    user_id          uuid NOT NULL,
    user_name        text NOT NULL,
    order_item_id    uuid NOT NULL,
    rating           smallint NOT NULL,
    title            text NULL,
    body             text NULL,
    status           review_status NOT NULL DEFAULT 'pending',
    helpful_count    int NOT NULL DEFAULT 0,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (product_id) REFERENCES products (id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (order_item_id) REFERENCES order_items (id),
    UNIQUE (order_item_id),
    CHECK (rating BETWEEN 1 AND 5)
);

CREATE INDEX IF NOT EXISTS idx_reviews_product_id ON reviews (product_id);

CREATE TABLE IF NOT EXISTS review_votes (
    review_id  uuid NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (review_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_review_votes_review_id ON review_votes (review_id);
CREATE INDEX IF NOT EXISTS idx_review_votes_user_id ON review_votes (user_id);


-- =====================================================================
-- WISHLIST
-- =====================================================================

CREATE TABLE IF NOT EXISTS wishlist_items (
    id            uuid NOT NULL,
    user_id       uuid NOT NULL,
    product_id    uuid NOT NULL,
    variant_id    uuid NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (product_id) REFERENCES products (id),
    FOREIGN KEY (variant_id) REFERENCES product_variants (id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wishlist_items_user_product_variant
    ON wishlist_items (user_id, product_id, COALESCE(variant_id, '00000000-0000-0000-0000-000000000000'::uuid));

CREATE INDEX IF NOT EXISTS idx_wishlist_items_user_id ON wishlist_items (user_id);
CREATE INDEX IF NOT EXISTS idx_wishlist_items_variant_id ON wishlist_items (variant_id);


-- =====================================================================
-- NOTIFICATIONS
-- =====================================================================

CREATE TABLE IF NOT EXISTS user_notifications (
    id          bigint GENERATED ALWAYS AS IDENTITY,
    user_id     uuid NOT NULL,
    title       jsonb NOT NULL,
    message     jsonb NOT NULL,
    category    varchar(50) NOT NULL DEFAULT 'system',
    action_url  text NULL,
    metadata    jsonb NULL,
    read_at     timestamptz NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_notifications_user_unread 
    ON user_notifications (user_id, read_at) 
    WHERE read_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_notifications_user_created 
    ON user_notifications (user_id, created_at DESC);


-- =====================================================================
-- ADMIN SETTINGS: key-value store for store configuration
-- =====================================================================

CREATE TABLE IF NOT EXISTS admin_settings (
    id          uuid NOT NULL DEFAULT gen_random_uuid(),
    key         varchar(100) NOT NULL UNIQUE,
    value       jsonb NOT NULL DEFAULT '{}'::jsonb,
    updated_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);


-- =====================================================================
-- AUDIT LOG: append-only trail of privileged back-office mutations
-- =====================================================================

CREATE TABLE IF NOT EXISTS audit_log (
    id           bigint GENERATED ALWAYS AS IDENTITY,
    actor_id     uuid        NULL,
    actor_role   text        NOT NULL DEFAULT '',
    action       text        NOT NULL,
    entity_type  text        NOT NULL DEFAULT '',
    entity_id    text        NOT NULL DEFAULT '',
    method       text        NOT NULL DEFAULT '',
    path         text        NOT NULL DEFAULT '',
    status_code  integer     NOT NULL DEFAULT 0,
    ip_address   text        NOT NULL DEFAULT '',
    user_agent   text        NOT NULL DEFAULT '',
    before_state jsonb       NULL,
    after_state  jsonb       NULL,
    metadata     jsonb       NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_actor_id   ON audit_log (actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_action     ON audit_log (action);
CREATE INDEX IF NOT EXISTS idx_audit_log_entity     ON audit_log (entity_type, entity_id);


-- ---- 000003_seller_domain.up.sql ----
-- =====================================================================
-- 000003 Seller Domain (Design B — seller-owned products)
-- See docs/adr-seller-domain.md
-- =====================================================================
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

-- Seller fulfillment queue: the seller's own lines and their shipping state.
CREATE INDEX IF NOT EXISTS idx_order_items_fulfillment_status
    ON order_items (order_id, seller_id, fulfillment_status);


-- ---- 000004_merchant_lifecycle.up.sql ----
-- =====================================================================
-- 000004 Merchant lifecycle hardening
-- Status review metadata, slug policy, payout idempotency, and frozen
-- commission components on order lines.
-- =====================================================================
ALTER TABLE sellers
    ADD COLUMN status_reason text NULL,
    ADD COLUMN reviewed_by uuid NULL REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN reviewed_at timestamptz NULL;

-- Slugs are public identifiers: enforce the API's 3-63 character policy at
-- the database boundary as well.
ALTER TABLE sellers DROP CONSTRAINT IF EXISTS sellers_slug_check;
ALTER TABLE sellers ADD CONSTRAINT sellers_slug_check
    CHECK (
        length(slug) BETWEEN 3 AND 63
        AND slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
    );

-- Payout retries must be idempotent per seller. NULL keys preserve the
-- existing manual/admin callers that do not supply one.
ALTER TABLE seller_payouts ADD COLUMN idempotency_key text NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_seller_payouts_idempotency
    ON seller_payouts (seller_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

-- Freeze both sides of the commission split for historical audit. The
-- existing seller_earning stays the authoritative seller share.
ALTER TABLE order_items
    ADD COLUMN commission_rate_bps int NOT NULL DEFAULT 0
        CHECK (commission_rate_bps BETWEEN 0 AND 10000),
    ADD COLUMN commission_amount bigint NOT NULL DEFAULT 0
        CHECK (commission_amount >= 0);


-- ---- 000005_user_product_views.up.sql ----
-- =====================================================================
-- 000005 User browsing history
-- Per-user recently viewed products, deduplicated and capped by the
-- service's rolling window.
-- =====================================================================
CREATE TABLE IF NOT EXISTS user_product_views (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    viewed_at  timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_user_product_views UNIQUE (user_id, product_id)
);

-- Primary read path: the newest N views for one account.
CREATE INDEX IF NOT EXISTS idx_user_product_views_lookup
    ON user_product_views (user_id, viewed_at DESC);

-- Product deletion cascades benefit from an indexed child lookup.
CREATE INDEX IF NOT EXISTS idx_user_product_views_product_id
    ON user_product_views (product_id);


-- ---- 000006_coupon_seller_scope.up.sql ----
-- =====================================================================
-- 000006 Coupon seller scope
-- NULL seller_id = platform coupon (admin). A set seller_id scopes the
-- coupon to that merchant's products: it discounts only the seller's
-- eligible subtotal at checkout and is manageable through the seller portal.
-- =====================================================================
ALTER TABLE coupons
    ADD COLUMN IF NOT EXISTS seller_id uuid NULL
        REFERENCES sellers (id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_coupons_seller_id
    ON coupons (seller_id) WHERE seller_id IS NOT NULL;


-- ---- 000007_seller_attributes.up.sql ----
-- =====================================================================
-- 000007 Seller custom attributes
-- Seller-scoped attribute definitions attached to a category. They extend
-- the admin-defined category attributes for that merchant's products only.
-- =====================================================================
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


-- ---- 000008_product_bundles.up.sql ----
-- =====================================================================
-- 000008 Product bundles
-- A bundle is a sellable product composed of other seller catalog items.
-- The bundle's own product carries the bundle price; bundle_items records
-- the composition for display and fulfillment.
-- =====================================================================
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
