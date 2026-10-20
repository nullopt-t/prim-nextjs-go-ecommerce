-- Convert internal, append-only/detail tables from random v4 UUID primary
-- keys to compact, database-owned bigint identity keys.
--
-- These tables have no inbound foreign keys (nothing references their id) and
-- are not exposed as public opaque handles, so the app can let PostgreSQL own
-- key generation and benefit from sequential B-tree inserts.
--
-- Public aggregate tables (users, orders, products, variants, carts, ...) keep
-- UUID keys: their ids are client-generated and/or exposed as opaque handles.

BEGIN;

-- audit_log
ALTER TABLE audit_log DROP CONSTRAINT IF EXISTS audit_log_pkey;
ALTER TABLE audit_log DROP COLUMN id;
ALTER TABLE audit_log ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log (created_at DESC, id DESC);

-- inventory_ledgers
ALTER TABLE inventory_ledgers DROP CONSTRAINT IF EXISTS inventory_ledgers_pkey;
ALTER TABLE inventory_ledgers DROP COLUMN id;
ALTER TABLE inventory_ledgers ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;

-- order_status_history
ALTER TABLE order_status_history DROP CONSTRAINT IF EXISTS order_status_history_pkey;
ALTER TABLE order_status_history DROP COLUMN id;
ALTER TABLE order_status_history ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;

-- user_notifications
ALTER TABLE user_notifications DROP CONSTRAINT IF EXISTS user_notifications_pkey;
ALTER TABLE user_notifications DROP COLUMN id;
ALTER TABLE user_notifications ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;

-- coupon_redemptions
ALTER TABLE coupon_redemptions DROP CONSTRAINT IF EXISTS coupon_redemptions_pkey;
ALTER TABLE coupon_redemptions DROP COLUMN id;
ALTER TABLE coupon_redemptions ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;

-- variant_media
ALTER TABLE variant_media DROP CONSTRAINT IF EXISTS variant_media_pkey;
ALTER TABLE variant_media DROP COLUMN id;
ALTER TABLE variant_media ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;

-- cart_items
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_pkey;
ALTER TABLE cart_items DROP COLUMN id;
ALTER TABLE cart_items ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;

COMMIT;
