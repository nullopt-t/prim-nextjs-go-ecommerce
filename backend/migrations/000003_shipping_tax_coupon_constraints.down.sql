-- =====================================================================
-- REVERT: 000003_shipping_tax_coupon_constraints
-- =====================================================================

DROP INDEX IF EXISTS idx_coupon_redemptions_coupon_user;

ALTER TABLE orders
    DROP COLUMN IF EXISTS tax_amount;

ALTER TABLE orders
    DROP COLUMN IF EXISTS shipping_amount;
