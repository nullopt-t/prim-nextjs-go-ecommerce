-- Rollback of 000003_vendor_domain. FK columns are dropped before the
-- referenced tables (products/order_items depend on vendors).

BEGIN;

ALTER TABLE products DROP COLUMN IF EXISTS vendor_id;
ALTER TABLE order_items DROP COLUMN IF EXISTS vendor_earning;
ALTER TABLE order_items DROP COLUMN IF EXISTS vendor_id;

DROP TABLE IF EXISTS vendor_payouts;
DROP TABLE IF EXISTS vendors;

DROP TYPE IF EXISTS payout_status;
DROP TYPE IF EXISTS vendor_status;

COMMIT;
