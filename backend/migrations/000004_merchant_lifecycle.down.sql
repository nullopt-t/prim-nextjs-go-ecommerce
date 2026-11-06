-- Rollback of 000004_merchant_lifecycle.

BEGIN;

ALTER TABLE order_items
    DROP COLUMN IF EXISTS commission_amount,
    DROP COLUMN IF EXISTS commission_rate_bps;

DROP INDEX IF EXISTS idx_vendor_payouts_idempotency;
ALTER TABLE vendor_payouts DROP COLUMN IF EXISTS idempotency_key;

ALTER TABLE vendors DROP CONSTRAINT IF EXISTS vendors_slug_check;
ALTER TABLE vendors ADD CONSTRAINT vendors_slug_check
    CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$');

ALTER TABLE vendors
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS status_reason;

COMMIT;
