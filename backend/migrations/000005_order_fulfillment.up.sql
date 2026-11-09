BEGIN;

DO $$ BEGIN
    CREATE TYPE fulfillment_status AS ENUM ('pending', 'shipped', 'delivered');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS fulfillment_status fulfillment_status NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS tracking_number text NULL,
    ADD COLUMN IF NOT EXISTS carrier text NULL,
    ADD COLUMN IF NOT EXISTS shipped_at timestamptz NULL;

CREATE INDEX IF NOT EXISTS idx_order_items_fulfillment_status
    ON order_items (order_id, seller_id, fulfillment_status);

COMMIT;
