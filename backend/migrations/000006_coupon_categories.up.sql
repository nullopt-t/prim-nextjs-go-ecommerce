CREATE TABLE IF NOT EXISTS coupon_categories (
    coupon_id   uuid NOT NULL REFERENCES coupons (id) ON DELETE CASCADE,
    category_id uuid NOT NULL REFERENCES product_categories (id) ON DELETE CASCADE,
    PRIMARY KEY (coupon_id, category_id)
);
CREATE INDEX IF NOT EXISTS idx_coupon_categories_category_id ON coupon_categories (category_id);
