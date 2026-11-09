BEGIN;

DROP INDEX IF EXISTS idx_order_items_fulfillment_status;

ALTER TABLE order_items
    DROP COLUMN IF EXISTS fulfillment_status,
    DROP COLUMN IF EXISTS tracking_number,
    DROP COLUMN IF EXISTS carrier,
    DROP COLUMN IF EXISTS shipped_at;

DROP TYPE IF EXISTS fulfillment_status;

COMMIT;
