-- =====================================================================
-- ORDERS: explicit shipping and tax amounts
-- =====================================================================

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS shipping_amount bigint NOT NULL DEFAULT 0;

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS tax_amount bigint NOT NULL DEFAULT 0;

-- =====================================================================
-- COUPONS: enforce at most one redemption per user per coupon at the DB level
-- (backstop for the per-user limit enforced inside checkout transactions)
-- =====================================================================

CREATE UNIQUE INDEX IF NOT EXISTS idx_coupon_redemptions_coupon_user
    ON coupon_redemptions (coupon_id, user_id);
