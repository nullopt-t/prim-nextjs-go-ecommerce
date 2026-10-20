-- Revert the bigint identity keys back to random UUID primary keys.
--
-- Reverting cannot preserve the original identifiers: the bigint values are
-- discarded and each row receives a fresh UUID. This is acceptable for a dev
-- rollback because nothing references these ids.

BEGIN;

ALTER TABLE audit_log DROP CONSTRAINT IF EXISTS audit_log_pkey;
ALTER TABLE audit_log DROP COLUMN id;
ALTER TABLE audit_log ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY;
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log (created_at DESC, id DESC);

ALTER TABLE inventory_ledgers DROP CONSTRAINT IF EXISTS inventory_ledgers_pkey;
ALTER TABLE inventory_ledgers DROP COLUMN id;
ALTER TABLE inventory_ledgers ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY;

ALTER TABLE order_status_history DROP CONSTRAINT IF EXISTS order_status_history_pkey;
ALTER TABLE order_status_history DROP COLUMN id;
ALTER TABLE order_status_history ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY;

ALTER TABLE user_notifications DROP CONSTRAINT IF EXISTS user_notifications_pkey;
ALTER TABLE user_notifications DROP COLUMN id;
ALTER TABLE user_notifications ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY;

ALTER TABLE coupon_redemptions DROP CONSTRAINT IF EXISTS coupon_redemptions_pkey;
ALTER TABLE coupon_redemptions DROP COLUMN id;
ALTER TABLE coupon_redemptions ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY;

ALTER TABLE variant_media DROP CONSTRAINT IF EXISTS variant_media_pkey;
ALTER TABLE variant_media DROP COLUMN id;
ALTER TABLE variant_media ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY;

ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_pkey;
ALTER TABLE cart_items DROP COLUMN id;
ALTER TABLE cart_items ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY;

COMMIT;
