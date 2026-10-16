-- =====================================================================
-- ORDERS: add payment_method, tracking_number, shipping_amount, tax_amount
-- =====================================================================

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS payment_method text NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS tracking_number text NULL,
    ADD COLUMN IF NOT EXISTS shipping_amount bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS tax_amount bigint NOT NULL DEFAULT 0;

COMMIT;
