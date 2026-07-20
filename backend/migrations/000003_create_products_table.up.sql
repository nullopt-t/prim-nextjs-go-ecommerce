-- ============================================================
-- ENUMS
-- ============================================================

DROP TYPE IF EXISTS product_status;
CREATE TYPE product_status AS ENUM (
    'draft',
    'published',
    'archived'
);

DROP TYPE IF EXISTS currency;
CREATE TYPE currency AS ENUM (
    'USD',
    'EUR',
    'EGP'
);

DROP TYPE IF EXISTS variant_media_type;
CREATE TYPE variant_media_type AS ENUM (
    'thumbnail',
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
    logo_alt TEXT NOT NULL,

    -- created_by UUID NOT NULL,
    -- updated_by UUID NOT NULL,

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
        REFERENCES brands(id) NULL,

    title TEXT NOT NULL,
    slug TEXT NOT NULL,

    short_description TEXT NOT NULL,
    description TEXT NOT NULL,

    status product_status NOT NULL DEFAULT 'draft',

    default_variant_id UUID,


    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ
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

    sku TEXT NULL,

    price BIGINT NOT NULL,
    currency currency NOT NULL,

    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

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
-- OBJECTS object storage objects table
-- ============================================================
CREATE TABLE IF NOT EXISTS objects (
    id UUID PRIMARY KEY,
    size BIGINT NOT NULL,
    content_type TEXT NOT NULL,
    bucket TEXT NOT NULL,
    key TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ NULL
);

-- ============================================================
-- PRODUCT MEDIA
-- ============================================================

CREATE TABLE IF NOT EXISTS variant_media (
    id UUID PRIMARY KEY,

    variant_id UUID
        REFERENCES product_variants(id)
        ON DELETE CASCADE,

    object_id UUID
        REFERENCES objects(id)
        ON DELETE CASCADE,


    type variant_media_type NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX variant_media_variant_idx
ON variant_media(variant_id);

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
-- CATEGORIES
-- ============================================================
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    parent_id UUID NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    UNIQUE(slug)
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
