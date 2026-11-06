-- =====================================================================
-- 000006 Coupon seller scope
-- NULL seller_id = platform coupon (admin). A set seller_id scopes the
-- coupon to that merchant's products: it discounts only the seller's
-- eligible subtotal at checkout and is manageable through the seller portal.
-- =====================================================================

BEGIN;

ALTER TABLE coupons
    ADD COLUMN IF NOT EXISTS seller_id uuid NULL
        REFERENCES sellers (id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_coupons_seller_id
    ON coupons (seller_id) WHERE seller_id IS NOT NULL;

COMMIT;
