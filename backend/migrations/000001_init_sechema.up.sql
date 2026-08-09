-- =====================================================================
-- PRIM E-COMMERCE DATABASE SCHEMA (Production Hardened)
-- PostgreSQL
-- =====================================================================

BEGIN;

-- =====================================================================
-- ENUM TYPES
-- =====================================================================

DO $$
BEGIN
    IF to_regtype('storage_object_status') IS NULL THEN
        CREATE TYPE storage_object_status AS ENUM ('uploaded', 'uploading', 'deleting', 'deleted');
    END IF;
    IF to_regtype('publication_status') IS NULL THEN
        CREATE TYPE publication_status AS ENUM ('draft', 'published', 'archived');
    END IF;
    IF to_regtype('user_role') IS NULL THEN
        CREATE TYPE user_role AS ENUM ('customer', 'admin', 'support');
    END IF;
    IF to_regtype('order_status') IS NULL THEN
        CREATE TYPE order_status AS ENUM ('pending', 'paid', 'processing', 'shipped', 'delivered', 'canceled', 'refunded');
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
    email                text NOT NULL,
    full_name            text NULL,
    role                 user_role NOT NULL DEFAULT 'customer',
    is_email_verified    boolean NOT NULL DEFAULT false,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz NULL,
    PRIMARY KEY (id)
);

-- Fix: Allow re-registering with the same email if the old account was soft-deleted
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower
    ON users (lower(email))
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS addresses (
    id             uuid NOT NULL,
    user_id        uuid NOT NULL,
    label          text NULL,
    full_name      text NOT NULL,
    phone          text NULL,
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
    public_id                 uuid NOT NULL,
    name                      text NOT NULL,
    link                      text NULL,
    logo_object_id            uuid NULL,
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz NOT NULL DEFAULT now(),
    deleted_at                timestamptz NULL,
    PRIMARY KEY (id),
    UNIQUE (public_id),
    FOREIGN KEY (logo_object_id) REFERENCES storage_objects (id)
);

CREATE UNIQUE INDEX idx_product_brands_active_name
ON product_brands (name)
WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS product_tags (
    id            uuid NOT NULL,
    public_id     uuid NOT NULL,
    name          text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz NULL,
    PRIMARY KEY (id),
    UNIQUE (public_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_tags_active_name
ON product_tags (name)
WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS product_categories (
    id            uuid NOT NULL,
    public_id     uuid NOT NULL,
    parent_id     uuid NULL,
    name          text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz NULL,
    PRIMARY KEY (id),
    UNIQUE (public_id),
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
    public_id      uuid NOT NULL,
    title          text NOT NULL,
    description    text NOT NULL,
    highlights     jsonb NULL,
    status         publication_status NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    deleted_at     timestamptz NULL,
    PRIMARY KEY (id),
    UNIQUE (public_id),
    FOREIGN KEY (brand_id) REFERENCES product_brands (id),
    FOREIGN KEY (category_id) REFERENCES product_categories (id)
);

CREATE TABLE IF NOT EXISTS product_variants (
    id                   uuid NOT NULL,
    public_id            uuid NOT NULL,
    product_id           uuid NOT NULL,
    is_default           boolean NOT NULL DEFAULT false,
    title                text NOT NULL,
    price                bigint NULL,
    crossed_out_price    bigint NULL,
    currency             text NULL,
    attributes           jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz NULL,
    PRIMARY KEY (id),
    UNIQUE (public_id),
    FOREIGN KEY (product_id) REFERENCES products (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_variants_default
    ON product_variants (product_id)
    WHERE is_default = true AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_product_variants_attributes ON product_variants USING GIN (attributes);
CREATE INDEX IF NOT EXISTS idx_product_variants_product_id ON product_variants (product_id);
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

CREATE TABLE IF NOT EXISTS product_media (
    id                   uuid NOT NULL,
    public_id            uuid NOT NULL,
    product_id           uuid NOT NULL,
    storage_object_id    uuid NOT NULL,
    media_type           text NOT NULL,
    sort_order           int NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE (public_id),
    FOREIGN KEY (product_id) REFERENCES products (id),
    FOREIGN KEY (storage_object_id) REFERENCES storage_objects (id)
);

CREATE TABLE IF NOT EXISTS variant_media (
    id                   uuid NOT NULL,
    public_id            uuid NOT NULL,
    variant_id           uuid NOT NULL,
    storage_object_id    uuid NOT NULL,
    media_type           text NOT NULL,
    sort_order           int NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE (public_id),
    FOREIGN KEY (variant_id) REFERENCES product_variants (id),
    FOREIGN KEY (storage_object_id) REFERENCES storage_objects (id)
);

CREATE INDEX IF NOT EXISTS idx_product_media_product_id ON product_media (product_id);
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
    deleted_at    timestamptz NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    CHECK (user_id IS NOT NULL OR session_id IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS cart_items (
    id                   uuid NOT NULL,
    cart_id              uuid NOT NULL,
    variant_id           uuid NOT NULL,
    quantity             int NOT NULL,
    price_at_purchase    bigint NOT NULL,
    currency             text NOT NULL,
    carted_at            timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz NULL,
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
    id              uuid NOT NULL,
    variant_id      uuid NOT NULL,
    quantity        int NOT NULL,
    reason          inventory_reason NOT NULL,
    reference_id    text NULL,
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


-- =====================================================================
-- COUPONS
-- =====================================================================

CREATE TABLE IF NOT EXISTS coupons (
    id                 uuid NOT NULL,
    code               text NOT NULL,
    discount_type      discount_type NOT NULL,
    discount_value     bigint NOT NULL,
    currency           text NULL,
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
    billing_address     jsonb NOT NULL,
    status              order_status NOT NULL DEFAULT 'pending',
    coupon_id           uuid NULL,
    discount_amount     bigint NOT NULL DEFAULT 0,
    total_amount        bigint NOT NULL,
    currency            text NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (customer_id) REFERENCES users (id),
    FOREIGN KEY (coupon_id) REFERENCES coupons (id)
);

CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders (customer_id);

CREATE TABLE IF NOT EXISTS order_items (
    id                   uuid NOT NULL,
    order_id             uuid NOT NULL,
    variant_id           uuid NOT NULL,
    quantity             int NOT NULL,
    price_at_purchase    bigint NOT NULL,
    product_snapshot     jsonb NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (order_id) REFERENCES orders (id),
    FOREIGN KEY (variant_id) REFERENCES product_variants (id),
    CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_variant_id ON order_items (variant_id);

CREATE TABLE IF NOT EXISTS order_status_history (
    id            uuid NOT NULL,
    order_id      uuid NOT NULL,
    status        order_status NOT NULL,
    notes         text NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (order_id) REFERENCES orders (id)
);

CREATE INDEX IF NOT EXISTS idx_order_status_history_order_id ON order_status_history (order_id);

CREATE TABLE IF NOT EXISTS coupon_redemptions (
    id                 uuid NOT NULL,
    coupon_id          uuid NOT NULL,
    user_id            uuid NULL,
    order_id           uuid NOT NULL,
    discount_amount    bigint NOT NULL,
    created_at         timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (coupon_id) REFERENCES coupons (id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (order_id) REFERENCES orders (id)
);

CREATE INDEX IF NOT EXISTS idx_coupon_redemptions_coupon_id ON coupon_redemptions (coupon_id);


-- =====================================================================
-- PAYMENTS
-- =====================================================================

CREATE TABLE IF NOT EXISTS payments (
    id                         uuid NOT NULL,
    order_id                   uuid NOT NULL,
    amount                     bigint NOT NULL,
    currency                   text NOT NULL,
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
    order_item_id    uuid NOT NULL,
    rating           smallint NOT NULL,
    title            text NULL,
    body             text NULL,
    status           review_status NOT NULL DEFAULT 'pending',
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


-- =====================================================================
-- WISHLIST
-- =====================================================================

CREATE TABLE IF NOT EXISTS wishlist_items (
    id            uuid NOT NULL,
    user_id       uuid NOT NULL,
    product_id    uuid NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (product_id) REFERENCES products (id),
    UNIQUE (user_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_wishlist_items_user_id ON wishlist_items (user_id);

COMMIT;
