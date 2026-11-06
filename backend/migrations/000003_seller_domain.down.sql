-- Rollback of 000003_seller_domain. FK columns are dropped before the
-- referenced tables (products/order_items depend on sellers).

BEGIN;

ALTER TABLE products DROP COLUMN IF EXISTS seller_id;
ALTER TABLE order_items DROP COLUMN IF EXISTS seller_earning;
ALTER TABLE order_items DROP COLUMN IF EXISTS seller_id;

DROP TABLE IF EXISTS seller_payouts;
DROP TABLE IF EXISTS sellers;

DROP TYPE IF EXISTS payout_status;
DROP TYPE IF EXISTS seller_status;

COMMIT;
