-- =====================================================================
-- 000006 Coupon vendor scope
-- NULL vendor_id = platform coupon (admin). A set vendor_id scopes the
-- coupon to that merchant's products: it discounts only the vendor's
-- eligible subtotal at checkout and is manageable through the vendor portal.
-- =====================================================================

BEGIN;

ALTER TABLE coupons
    ADD COLUMN IF NOT EXISTS vendor_id uuid NULL
        REFERENCES vendors (id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_coupons_vendor_id
    ON coupons (vendor_id) WHERE vendor_id IS NOT NULL;

COMMIT;
