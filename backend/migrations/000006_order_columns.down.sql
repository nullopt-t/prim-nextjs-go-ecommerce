-- =====================================================================
-- ORDERS: drop added columns
-- =====================================================================

ALTER TABLE orders
    DROP COLUMN IF EXISTS payment_method,
    DROP COLUMN IF EXISTS tracking_number,
    DROP COLUMN IF EXISTS shipping_amount,
    DROP COLUMN IF EXISTS tax_amount;

COMMIT;
